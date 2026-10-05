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
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/workflows"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// documentLayout is the current layout a stored document name resolves to
type documentLayout struct {
	// name is the current name the document is stored under
	name string
	// schema is the reflected JSON schema of the current layout
	schema json.RawMessage
	// upgrade reshapes a document stored under an earlier layout, nil when none is declared
	upgrade types.UpgradeFunc
	// validate checks a schema-valid document for constraints the schema cannot express, nil when none is declared
	validate types.ValidateFunc
	// replaced reports that the stored name is retired and name is its replacement
	replaced bool
}

// documentKind projects one stored document kind onto name-keyed documents
type documentKind struct {
	// label is the noun used in logs and errors for one document of this kind
	label string
	// sentinel is the error a failed conformance is wrapped with
	sentinel error
	// resolve maps a stored name onto its current layout, false when the definition does not declare it
	resolve func(name string) (documentLayout, bool)
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

// upgradeInstallation conforms every stored document onto its current name and persists them with the definition version in one transaction;
// the transaction first claims the version so a concurrent upgrade that already stamped it leaves the installation to that upgrade and reloads it,
// and an operation input that fails to conform keeps its stored value and marks only that operation unhealthy
func (r *Runtime) upgradeInstallation(ctx context.Context, installation *ent.Integration) error {
	def, err := r.resolveDefinitionForInstallation(installation)
	if err != nil {
		return fmt.Errorf("resolve definition: %w", err)
	}

	providerState, err := def.ProviderState(installation.ProviderState)
	if err != nil {
		return fmt.Errorf("resolve provider state: %w", err)
	}

	req, records, err := r.installationRequest(ctx, installation, def)
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

	operationDocuments, failed := conformDocuments(ctx, req, operationInputKind(def), operationConfig.Operations)
	operationConfig = types.IntegrationOperationConfig{Operations: operationDocuments}

	providerStateNext := installation.ProviderState

	if registration, replaced, ok := def.ResolveCredential(providerState.CredentialRef); ok && replaced {
		providerStateNext, err = def.WithProviderState(installation.ProviderState, types.DefinitionProviderState{CredentialRef: registration.Ref})
		if err != nil {
			return fmt.Errorf("resolve provider state: %w", err)
		}
	}

	health := installation.Health
	health.UnhealthyOperations = retiredHealth(installation.Health.UnhealthyOperations, def)
	version := r.Registry().Version(def.ID)

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	claimed, err := workflows.WithTx(systemCtx, r.DB(), nil, func(ctx context.Context, tx *ent.Tx) (bool, error) {
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

		return true, nil
	})
	if err != nil {
		return err
	}

	if !claimed {
		current, err := r.DB().Integration.Get(systemCtx, installation.ID)
		if err != nil {
			return err
		}

		*installation = *current

		return nil
	}

	installation.UserInput = userInput
	installation.OperationConfig = operationConfig
	installation.ProviderState = providerStateNext
	installation.Health = health
	installation.DefinitionVersion = version

	r.keystore().InvalidateClients(installation.ID)

	return r.finishUpgrade(ctx, installation, def, failed)
}

// finishUpgrade runs the post-commit steps of an upgrade: purging loops queued under retired operation names, reconciling webhooks,
// marking operations whose input failed to conform, and resetting loops when an operation was renamed
func (r *Runtime) finishUpgrade(ctx context.Context, installation *ent.Integration, def types.Definition, failed map[string]error) error {
	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	for _, operation := range def.Operations {
		for _, old := range operation.Replaces {
			purged, err := r.purgeReconcileLoop(systemCtx, installation.ID, old)
			if err != nil {
				return err
			}

			if purged > 0 {
				logx.FromContext(ctx).Info().Str("retired_operation", old).Str("operation", operation.Name).Int("purged", purged).Msg("cancelled reconcile loops queued under a retired operation name")
			}
		}
	}

	if err := r.reconcileInstallationWebhooks(systemCtx, installation, ""); err != nil {
		return fmt.Errorf("upgrade webhooks: %w", err)
	}

	for _, name := range slices.Sorted(maps.Keys(failed)) {
		operation, _, _ := def.ResolveOperation(name)

		if err := r.MarkOperationUnhealthy(ctx, installation, operation.Name, failed[name].Error()); err != nil {
			return err
		}
	}

	if lo.SomeBy(def.Operations, func(operation types.OperationRegistration) bool { return len(operation.Replaces) > 0 }) {
		if err := r.ResetReconcileLoops(ctx, installation); err != nil {
			return fmt.Errorf("reset reconcile loops: %w", err)
		}
	}

	return nil
}

// legacyDocuments returns the installation's stored user input and operation input, seeded from main's client config when neither has been written:
// user input from the whole document, each stored-input operation from the section under its camelCase name, else the whole document
//
// TODO: remove with the integration config column once every installation has been upgraded off main's client config
func legacyDocuments(installation *ent.Integration, def types.Definition) (types.IntegrationUserInput, types.IntegrationOperationConfig) {
	legacy := installation.Config.ClientConfig

	if installation.UserInput.Layout != "" || len(installation.OperationConfig.Operations) > 0 || jsonx.IsEmptyRawMessage(legacy) {
		return installation.UserInput, installation.OperationConfig
	}

	operationConfig := installation.OperationConfig

	for _, operation := range def.Operations {
		if operation.Input == nil {
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

	userInput := installation.UserInput

	if def.UserInput != nil {
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
func conformCredentials(ctx context.Context, req types.InstallationRequest, def types.Definition, records map[types.CredentialSlotID]types.CredentialSet) (map[types.CredentialSlotID]types.CredentialSet, error) {
	stored := make(map[string]json.RawMessage, len(records))

	for slot, credential := range records {
		stored[slot.String()] = credential.Data
	}

	conformed, failed := conformDocuments(ctx, req, credentialKind(def), stored)
	if err := joinFailures(failed); err != nil {
		return nil, err
	}

	next := make(map[types.CredentialSlotID]types.CredentialSet, len(conformed))

	for name, document := range conformed {
		next[types.NewCredentialSlotID(name)] = types.CredentialSet{Data: document}
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

	conformed, failed := conformDocuments(ctx, req, userInputKind(def), stored)
	if err := joinFailures(failed); err != nil {
		return userInput, err
	}

	if len(conformed) == 0 {
		return userInput, nil
	}

	return types.IntegrationUserInput{Layout: def.UserInput.Name, Data: conformed[def.UserInput.Name]}, nil
}

// joinFailures joins the failed documents' errors in name order, nil when none failed
func joinFailures(failed map[string]error) error {
	return errors.Join(lo.Map(slices.Sorted(maps.Keys(failed)), func(name string, _ int) error { return failed[name] })...)
}

// credentialKind projects stored credential slots onto the definition's credential registrations
func credentialKind(def types.Definition) documentKind {
	return documentKind{
		label:    "slot",
		sentinel: ErrCredentialInvalid,
		resolve: func(name string) (documentLayout, bool) {
			registration, replaced, ok := def.ResolveCredential(types.NewCredentialSlotID(name))
			if !ok {
				return documentLayout{}, false
			}

			return documentLayout{name: registration.Ref.String(), schema: registration.StoredSchema, upgrade: registration.Upgrade, validate: registration.Validate, replaced: replaced}, true
		},
	}
}

// userInputKind projects the stored user input document onto the definition's user input layout
func userInputKind(def types.Definition) documentKind {
	return documentKind{
		label:    "layout",
		sentinel: ErrUserInputInvalid,
		resolve: func(name string) (documentLayout, bool) {
			if def.UserInput == nil {
				return documentLayout{}, false
			}

			return documentLayout{name: def.UserInput.Name, schema: def.UserInput.Schema, upgrade: def.UserInput.Upgrade, validate: def.UserInput.Validate, replaced: name != def.UserInput.Name}, true
		},
	}
}

// operationInputKind projects stored operation input documents onto the definition's operation registrations
func operationInputKind(def types.Definition) documentKind {
	return documentKind{
		label:    "operation",
		sentinel: types.ErrOperationConfigInvalid,
		resolve: func(name string) (documentLayout, bool) {
			operation, replaced, ok := def.ResolveOperation(name)
			if !ok || operation.Input == nil {
				return documentLayout{}, false
			}

			return documentLayout{name: operation.Name, schema: operation.Input.Schema, upgrade: operation.Input.Upgrade, validate: operation.Input.Validate, replaced: replaced}, true
		},
	}
}

// conformDocuments conforms every stored document onto its current name, dropping a retired document whose replacement is already stored;
// a document that fails to conform keeps its stored value and is reported in failed under its stored name
func conformDocuments(ctx context.Context, req types.InstallationRequest, kind documentKind, stored map[string]json.RawMessage) (map[string]json.RawMessage, map[string]error) {
	next := maps.Clone(stored)

	if next == nil {
		next = map[string]json.RawMessage{}
	}

	failed := map[string]error{}

	for _, name := range slices.Sorted(maps.Keys(stored)) {
		layout, ok := kind.resolve(name)
		if !ok {
			logx.FromContext(ctx).Warn().Str(kind.label, name).Msg("stored document is not declared by the definition and was left untouched")

			continue
		}

		if layout.replaced && lo.HasKey(next, layout.name) {
			delete(next, name)

			continue
		}

		document, err := conformStored(ctx, req, layout.schema, layout.upgrade, layout.validate, name, stored[name], kind.sentinel)
		if err != nil {
			failed[name] = fmt.Errorf("%w: %s %s", err, kind.label, name)

			continue
		}

		next[layout.name] = document

		if layout.replaced {
			delete(next, name)
		}
	}

	return next, failed
}

// conformStored upgrades a stored document from the name it was persisted under, strips and defaults it against schema, then validates it
func conformStored(ctx context.Context, req types.InstallationRequest, schema json.RawMessage, upgrade types.UpgradeFunc, validate types.ValidateFunc, from string, stored json.RawMessage, sentinel error) (json.RawMessage, error) {
	document := stored

	if upgrade != nil {
		upgraded, err := upgrade(ctx, req, from, stored)
		if err != nil {
			return nil, fmt.Errorf("%w: upgrade: %w", sentinel, err)
		}

		document = upgraded
	}

	conformed, err := jsonx.ConformToSchema(schema, document)
	if err != nil {
		return nil, fmt.Errorf("%w: conform to schema: %w", sentinel, err)
	}

	if err := operations.ValidateInput(ctx, req, schema, validate, conformed, sentinel); err != nil {
		return nil, err
	}

	return conformed, nil
}
