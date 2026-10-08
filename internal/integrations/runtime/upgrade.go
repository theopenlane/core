package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/samber/lo"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationrun"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationwebhook"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/workflows"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/mapx"
)

// documentKind projects one stored document kind onto name-keyed documents
type documentKind struct {
	// label is the noun used in logs and errors for one document of this kind
	label string
	// sentinel is the error a failed conformance is wrapped with
	sentinel error
	// resolve maps a stored name onto its current layout, replaced when the stored name is retired, ok false when the definition does not declare it
	resolve func(name string) (layout types.InputRegistration, replaced bool, ok bool)
}

// outdated reports whether a stored definition version predates the current one; versions are ULIDs, so string order is mint order
func outdated(stored, current string) bool {
	return stored < current
}

// ensureCurrentVersion upgrades an installation whose stored definition version predates the registry version
func (r *Runtime) ensureCurrentVersion(ctx context.Context, installation *ent.Integration) error {
	if !outdated(installation.DefinitionVersion, r.Registry().Version(installation.DefinitionID)) {
		return nil
	}

	if err := r.upgradeInstallation(ctx, installation); err != nil {
		failed := fmt.Errorf("%w: %w", ErrInstallationUpgradeFailed, err)

		if markErr := r.MarkIntegrationUnhealthy(ctx, installation, failed.Error()); markErr != nil {
			failed = errors.Join(failed, markErr)
		}

		return types.Unhealthy(failed, failed.Error())
	}

	return nil
}

// upgradeInstallation conforms every stored document onto its current name, failing the whole upgrade when any document does not conform,
// cancels loops queued under retired operation names, persists the documents, renamed runs, and webhook rows with the definition version in
// one transaction, and resets the reconcile loops when an operation was renamed; the transaction first claims the version so a concurrent
// upgrade that already stamped it leaves the installation to that upgrade and reloads it
func (r *Runtime) upgradeInstallation(ctx context.Context, installation *ent.Integration) error {
	def, err := r.resolveDefinitionForInstallation(installation)
	if err != nil {
		return fmt.Errorf("resolve definition: %w", err)
	}

	providerState, err := def.ProviderState(installation.ProviderState)
	if err != nil {
		return fmt.Errorf("resolve provider state: %w", err)
	}

	req, records, err := r.installationRequest(ctx, installation)
	if err != nil {
		return fmt.Errorf("load credentials: %w", err)
	}

	credentials, err := conformCredentials(ctx, req, def, records)
	if err != nil {
		return fmt.Errorf("upgrade credentials: %w", err)
	}

	// TODO: replace with installation.UserInput, installation.OperationConfig once every installation has been upgraded off main's client config
	userInput, operationConfig := legacyDocuments(installation, def)
	req.UserInput = userInput.Data

	userInput, err = conformUserInput(ctx, req, def, userInput)
	if err != nil {
		return fmt.Errorf("upgrade user input: %w", err)
	}

	req.UserInput = userInput.Data

	operationDocuments, err := conformDocuments(ctx, req, operationInputKind(def), operationConfig.Operations)
	if err != nil {
		return fmt.Errorf("upgrade operation input: %w", err)
	}

	operationConfig = types.IntegrationOperationConfig{Operations: operationDocuments}

	providerStateNext := installation.ProviderState
	nextState := providerState

	connection, replaced, resolved := def.ResolveConnection(providerState.CredentialRef)

	switch {
	case resolved && replaced:
		nextState.CredentialRef = connection.Credential.Name
	case providerState.CredentialRef == "" && len(def.ConnectionList()) > 0:
		named := lo.FilterMap(lo.Keys(records), func(slot string, _ int) (string, bool) {
			held, _, ok := def.ResolveConnection(slot)

			return held.Credential.Name, ok
		})

		if unique := lo.Uniq(named); len(unique) == 1 {
			nextState.CredentialRef = unique[0]
		}
	}

	if nextState != providerState {
		providerStateNext, err = def.WithProviderState(installation.ProviderState, nextState)
		if err != nil {
			return fmt.Errorf("resolve provider state: %w", err)
		}
	}

	installationMetadataNext, metadataNext := installation.InstallationMetadata, installation.Metadata

	health := installation.Health
	health.UnhealthyOperations = retiredHealth(installation.Health.UnhealthyOperations, def)
	version := r.Registry().Version(def.ID)

	for _, operation := range def.Operations {
		for _, old := range operation.Replaces {
			purged, err := r.purgeReconcileLoop(ctx, installation.ID, old)
			if err != nil {
				return fmt.Errorf("purge retired loops: %w", err)
			}

			if purged > 0 {
				logx.FromContext(ctx).Info().Str("retired_operation", old).Str("operation", operation.Name).Int("purged", purged).Msg("cancelled reconcile loops queued under a retired operation name")
			}
		}
	}

	claimed, err := workflows.WithTx(ctx, r.DB(), nil, func(ctx context.Context, tx *ent.Tx) (bool, error) {
		stamped, err := tx.Integration.Update().
			Where(integration.ID(installation.ID), integration.Or(integration.DefinitionVersionIsNil(), integration.DefinitionVersionLT(version))).
			SetDefinitionVersion(version).
			Save(ctx)
		if err != nil || stamped == 0 {
			return false, err
		}

		if err := r.keystore().ReplaceCredentials(ctx, installation, records, credentials); err != nil {
			return false, fmt.Errorf("replace credentials: %w", err)
		}

		if def.Installation != nil && len(def.ConnectionList()) > 0 {
			current, err := tx.Integration.Get(ctx, installation.ID)
			if err != nil {
				return false, err
			}

			stored := map[string]json.RawMessage{}

			if current.InstallationMetadata.Layout != "" || !jsonx.IsEmptyRawMessage(current.InstallationMetadata.Attributes) {
				stored[current.InstallationMetadata.Layout] = current.InstallationMetadata.Attributes
			}

			conformed, err := conformDocuments(ctx, req, installationKind(def), stored)
			if err != nil {
				return false, fmt.Errorf("upgrade installation metadata: %w", err)
			}

			if len(conformed) > 0 {
				attributes := conformed[def.Installation.Name]
				display := current.InstallationMetadata.Display
				display.CredentialRef = nextState.CredentialRef

				switch {
				case def.Installation.Identifiable:
					identity, err := def.Installation.Identify(attributes)
					if err != nil {
						return false, fmt.Errorf("upgrade installation metadata: %w", err)
					}

					display.ExternalID, display.ExternalName = identity.ExternalID, identity.ExternalName
				default:
					display.ExternalID = installation.ID
				}

				displayMap, _ := jsonx.ToMap(display)

				installationMetadataNext = types.IntegrationInstallationMetadata{Layout: def.Installation.Name, Attributes: attributes, Display: display}
				metadataNext = mapx.DeepMergeMapAny(current.Metadata, mapx.PruneMapZeroAny(displayMap))

				if err := tx.Integration.UpdateOneID(installation.ID).
					SetInstallationMetadata(installationMetadataNext).
					SetMetadata(metadataNext).
					Exec(ctx); err != nil {
					return false, err
				}
			}
		}

		if err := tx.Integration.UpdateOneID(installation.ID).
			SetUserInput(userInput).
			SetOperationConfig(operationConfig).
			SetProviderState(providerStateNext).
			SetHealth(health).
			Exec(ctx); err != nil {
			return false, err
		}

		for _, operation := range def.Operations {
			if len(operation.Replaces) == 0 {
				continue
			}

			if err := tx.IntegrationRun.Update().
				Where(integrationrun.IntegrationIDEQ(installation.ID), integrationrun.OperationNameIn(operation.Replaces...)).
				SetOperationName(operation.Name).
				Exec(ctx); err != nil {
				return false, err
			}
		}

		for _, webhook := range def.Webhooks {
			if len(webhook.Replaces) == 0 {
				continue
			}

			if err := tx.IntegrationWebhook.Update().
				Where(
					integrationwebhook.IntegrationIDEQ(installation.ID),
					integrationwebhook.NameIn(webhook.Replaces...),
					integrationwebhook.ExternalEventIDNotNil(),
				).
				SetName(webhook.Name).
				Exec(ctx); err != nil {
				return false, err
			}
		}

		if err := r.reconcileInstallationWebhooks(ctx, tx.Client(), installation, ""); err != nil {
			return false, fmt.Errorf("upgrade webhooks: %w", err)
		}

		return true, nil
	})
	if err != nil {
		return err
	}

	if !claimed {
		current, err := r.DB().Integration.Get(ctx, installation.ID)
		if err != nil {
			return err
		}

		*installation = *current

		return nil
	}

	installation.UserInput = userInput
	installation.OperationConfig = operationConfig
	installation.ProviderState = providerStateNext
	installation.InstallationMetadata = installationMetadataNext
	installation.Metadata = metadataNext
	installation.Health = health
	installation.DefinitionVersion = version

	r.keystore().InvalidateClients(installation.ID)

	if lo.SomeBy(def.Operations, func(operation types.OperationRegistration) bool { return len(operation.Replaces) > 0 }) {
		if err := r.ResetReconcileLoops(ctx, installation); err != nil {
			return fmt.Errorf("reset reconcile loops: %w", err)
		}
	}

	return nil
}

// legacyDocuments returns the installation's user input and operation input, seeding each document not yet stored from main's client config
// on an installation main created, which carries no definition version: user input from the whole document, each stored-input operation from
// the section under its camelCase name, else the whole document
//
// TODO: remove with the integration config column once every installation has been upgraded off main's client config
func legacyDocuments(installation *ent.Integration, def types.Definition) (types.IntegrationUserInput, types.IntegrationOperationConfig) {
	legacy := installation.Config.ClientConfig
	userInput, operationConfig := installation.UserInput, installation.OperationConfig

	if installation.DefinitionVersion != "" || jsonx.IsEmptyRawMessage(legacy) {
		return userInput, operationConfig
	}

	for _, operation := range def.Operations {
		if !operation.Stored || lo.HasKey(operationConfig.Operations, operation.Name) {
			continue
		}

		section, ok := jsonx.DecodeObjectKey[map[string]json.RawMessage](legacy, lo.CamelCase(operation.Name))
		if !ok {
			operationConfig = operationConfig.With(operation.Name, legacy)

			continue
		}

		raw, err := jsonx.ToRawMessage(section)
		if err != nil {
			operationConfig = operationConfig.With(operation.Name, legacy)

			continue
		}

		operationConfig = operationConfig.With(operation.Name, raw)
	}

	if def.UserInput != nil && userInput.Layout == "" {
		userInput = types.IntegrationUserInput{Layout: def.UserInput.Name, Data: legacy}
	}

	return userInput, operationConfig
}

// retiredHealth moves unhealthy marks off retired operation names onto their replacements and drops marks for undeclared operations
func retiredHealth(unhealthy map[string]string, def types.Definition) map[string]string {
	next := lo.Assign(unhealthy)

	for _, operation := range def.Operations {
		for _, old := range operation.Replaces {
			reason, recorded := next[old]
			if !recorded {
				continue
			}

			if _, kept := next[operation.Name]; !kept {
				next[operation.Name] = reason
			}

			delete(next, old)
		}
	}

	maps.DeleteFunc(next, func(name string, _ string) bool {
		_, declared := def.Operation(name)

		return !declared
	})

	if len(next) == 0 {
		return nil
	}

	return next
}

// conformCredentials conforms every stored credential onto its current slot, failing when any credential does not conform
func conformCredentials(ctx context.Context, req types.InstallationRequest, def types.Definition, records map[string]types.CredentialSet) (map[string]types.CredentialSet, error) {
	stored := make(map[string]json.RawMessage, len(records))

	for slot, credential := range records {
		stored[slot] = credential.Data
	}

	conformed, err := conformDocuments(ctx, req, credentialKind(def), stored)
	if err != nil {
		return nil, err
	}

	next := make(map[string]types.CredentialSet, len(conformed))

	for name, document := range conformed {
		next[name] = types.CredentialSet{Data: document}
	}

	return next, nil
}

// conformUserInput conforms the stored user input document onto the current layout name, failing when it does not conform
func conformUserInput(ctx context.Context, req types.InstallationRequest, def types.Definition, userInput types.IntegrationUserInput) (types.IntegrationUserInput, error) {
	if def.UserInput == nil {
		return userInput, nil
	}

	stored := map[string]json.RawMessage{}

	if userInput.Layout != "" || !jsonx.IsEmptyRawMessage(userInput.Data) {
		stored[userInput.Layout] = userInput.Data
	}

	conformed, err := conformDocuments(ctx, req, userInputKind(def), stored)
	if err != nil {
		return userInput, err
	}

	if len(conformed) == 0 {
		return userInput, nil
	}

	return types.IntegrationUserInput{Layout: def.UserInput.Name, Data: conformed[def.UserInput.Name]}, nil
}

// credentialKind projects stored credential slots onto the definition's connections
func credentialKind(def types.Definition) documentKind {
	return documentKind{
		label:    "slot",
		sentinel: ErrCredentialInvalid,
		resolve: func(name string) (types.InputRegistration, bool, bool) {
			connection, replaced, ok := def.ResolveConnection(name)

			return connection.Credential, replaced, ok
		},
	}
}

// installationKind projects the stored installation metadata document onto the definition's installation layout
func installationKind(def types.Definition) documentKind {
	return documentKind{
		label:    "installation",
		sentinel: ErrInstallationMetadataInvalid,
		resolve: func(name string) (types.InputRegistration, bool, bool) {
			if def.Installation == nil {
				return types.InputRegistration{}, false, false
			}

			return def.Installation.InputRegistration, name != def.Installation.Name, true
		},
	}
}

// userInputKind projects the stored user input document onto the definition's user input layout
func userInputKind(def types.Definition) documentKind {
	return documentKind{
		label:    "layout",
		sentinel: ErrUserInputInvalid,
		resolve: func(name string) (types.InputRegistration, bool, bool) {
			if def.UserInput == nil {
				return types.InputRegistration{}, false, false
			}

			return *def.UserInput, name != def.UserInput.Name, true
		},
	}
}

// operationInputKind projects stored operation input documents onto the definition's operation registrations
func operationInputKind(def types.Definition) documentKind {
	return documentKind{
		label:    "operation",
		sentinel: types.ErrOperationConfigInvalid,
		resolve: func(name string) (types.InputRegistration, bool, bool) {
			operation, replaced, ok := def.ResolveOperation(name)
			if !ok || !operation.Stored {
				return types.InputRegistration{}, false, false
			}

			return operation.Input, replaced, true
		},
	}
}

// conformDocuments conforms every stored document onto its current name, dropping a retired document whose replacement is already stored,
// and returns every document's conformance failure joined in name order
func conformDocuments(ctx context.Context, req types.InstallationRequest, kind documentKind, stored map[string]json.RawMessage) (map[string]json.RawMessage, error) {
	next := maps.Clone(stored)

	if next == nil {
		next = map[string]json.RawMessage{}
	}

	var failed []error

	for _, name := range slices.Sorted(maps.Keys(stored)) {
		layout, replaced, ok := kind.resolve(name)
		if !ok {
			logx.FromContext(ctx).Warn().Str(kind.label, name).Msg("stored document is not declared by the definition and was left untouched")

			continue
		}

		if replaced && lo.HasKey(next, layout.Name) {
			delete(next, name)

			continue
		}

		document, err := conformStored(ctx, req, layout, name, stored[name], kind.sentinel)
		if err != nil {
			failed = append(failed, fmt.Errorf("%w: %s %s", err, kind.label, name))

			continue
		}

		next[layout.Name] = document

		if replaced {
			delete(next, name)
		}
	}

	return next, errors.Join(failed...)
}

// conformStored upgrades a stored document from the name it was persisted under, strips and defaults it against the layout schema, then validates it
func conformStored(ctx context.Context, req types.InstallationRequest, layout types.InputRegistration, from string, stored json.RawMessage, sentinel error) (json.RawMessage, error) {
	document := stored

	if layout.Upgrade != nil {
		upgraded, err := layout.Upgrade(ctx, req, from, stored)
		if err != nil {
			return nil, fmt.Errorf("%w: upgrade: %w", sentinel, err)
		}

		document = upgraded
	}

	conformed, err := jsonx.ConformToSchema(layout.Schema, document)
	if err != nil {
		return nil, fmt.Errorf("%w: conform to schema: %w", sentinel, err)
	}

	if err := operations.ValidateInput(ctx, req, layout.Schema, layout.Validate, conformed, sentinel); err != nil {
		return nil, err
	}

	return conformed, nil
}
