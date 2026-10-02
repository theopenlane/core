package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/samber/lo"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationrun"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationwebhook"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/types"
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

// ensureCurrentVersion upgrades an installation if versions dont match
func (r *Runtime) ensureCurrentVersion(ctx context.Context, installation *ent.Integration) error {
	if installation.DefinitionVersion == r.Registry().Version(installation.DefinitionID) {
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

// upgradeInstallation conforms every stored document onto its current name and stamps the installation with the definition version
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

	if err := r.upgradeCredentials(ctx, req, installation, def, records); err != nil {
		return fmt.Errorf("upgrade credentials: %w", err)
	}

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	if registration, replaced, ok := def.ResolveCredential(providerState.CredentialRef); ok && replaced {
		if err := r.persistConnectionState(systemCtx, installation, def, registration.Ref); err != nil {
			return fmt.Errorf("persist connection state: %w", err)
		}
	}

	if err := r.liftLegacyConfig(systemCtx, installation, def); err != nil {
		return fmt.Errorf("lift legacy config: %w", err)
	}

	req.UserInput = installation.UserInput.Data

	if err := r.upgradeUserInput(systemCtx, req, installation, def); err != nil {
		return fmt.Errorf("upgrade user input: %w", err)
	}

	req.UserInput = installation.UserInput.Data

	if err := r.upgradeOperationConfig(systemCtx, req, installation, def); err != nil {
		return fmt.Errorf("upgrade operation config: %w", err)
	}

	if err := r.upgradeOperations(systemCtx, installation, def); err != nil {
		return fmt.Errorf("upgrade operations: %w", err)
	}

	if err := r.upgradeWebhookDeliveries(systemCtx, installation, def); err != nil {
		return fmt.Errorf("upgrade webhooks: %w", err)
	}

	if err := r.RefreshInstallationMetadata(ctx, installation); err != nil {
		return fmt.Errorf("refresh installation metadata: %w", err)
	}

	renamed := lo.SomeBy(def.Operations, func(operation types.OperationRegistration) bool {
		return len(operation.Replaces) > 0
	})

	if renamed {
		if err := r.ResetReconcileLoops(ctx, installation); err != nil {
			return fmt.Errorf("reset reconcile loops: %w", err)
		}
	}

	if err := r.persistDefinitionVersion(systemCtx, installation, def); err != nil {
		return fmt.Errorf("set definition version: %w", err)
	}

	return nil
}

// persistDefinitionVersion stamps the installation with the current definition version and mirrors it on the record
func (r *Runtime) persistDefinitionVersion(ctx context.Context, installation *ent.Integration, def types.Definition) error {
	version := r.Registry().Version(def.ID)

	if err := r.DB().Integration.UpdateOneID(installation.ID).SetDefinitionVersion(version).Exec(ctx); err != nil {
		return err
	}

	installation.DefinitionVersion = version

	return nil
}

// liftLegacyConfig splits a pre-split flat client config document into user input and per-operation input
func (r *Runtime) liftLegacyConfig(ctx context.Context, installation *ent.Integration, def types.Definition) error {
	if installation.UserInput.Layout != "" || len(installation.OperationConfig.Operations) > 0 || jsonx.IsEmptyRawMessage(installation.Config.ClientConfig) {
		return nil
	}

	document, err := jsonx.ToRawMap(installation.Config.ClientConfig)
	if err != nil {
		return err
	}

	globals, err := schemaPropertyKeys(lo.FromPtr(def.UserInput).Schema)
	if err != nil {
		return err
	}

	settings, err := schemaPropertyKeys(types.OperationSettingsSchema())
	if err != nil {
		return err
	}

	claimed := map[string]string{}
	moved := map[string]struct{}{}
	lifted := installation.OperationConfig

	for _, operation := range def.Operations {
		if operation.Input == nil {
			continue
		}

		keys, err := schemaPropertyKeys(operation.Input.Schema)
		if err != nil {
			return fmt.Errorf("operation %s: %w", operation.Name, err)
		}

		section := map[string]json.RawMessage{}

		for _, key := range keys {
			value, stored := document[key]
			if !stored || lo.Contains(globals, key) {
				continue
			}

			if holder, taken := claimed[key]; taken && !lo.Contains(settings, key) {
				return fmt.Errorf("%w: key %s claimed by %s and %s", ErrLegacyConfigAmbiguous, key, holder, operation.Name)
			}

			claimed[key] = operation.Name
			moved[key] = struct{}{}
			section[key] = value
		}

		if len(section) == 0 {
			continue
		}

		raw, err := jsonx.ToRawMessage(section)
		if err != nil {
			return err
		}

		lifted = lifted.With(operation.Name, raw)
	}

	maps.DeleteFunc(document, func(key string, _ json.RawMessage) bool {
		_, lifted := moved[key]

		return lifted
	})

	if err := r.persistOperationConfig(ctx, installation, lifted); err != nil {
		return err
	}

	if def.UserInput != nil {
		raw, err := jsonx.ToRawMessage(document)
		if err != nil {
			return err
		}

		if err := r.persistUserInput(ctx, installation, types.IntegrationUserInput{Layout: def.UserInput.Name, Data: raw}, r.DB().Integration.UpdateOneID(installation.ID)); err != nil {
			return err
		}
	}

	logx.FromContext(ctx).Info().Int("operations", len(lifted.Operations)).Msg("lifted legacy client config into user input and operation config")

	return nil
}

// schemaPropertyKeys lists the root property keys of a reflected schema, none for an empty schema
func schemaPropertyKeys(schema json.RawMessage) ([]string, error) {
	if len(schema) == 0 {
		return nil, nil
	}

	root, _, err := jsonx.SchemaRoot(schema)
	if err != nil {
		return nil, err
	}

	if root.Properties == nil {
		return nil, nil
	}

	keys := make([]string, 0, root.Properties.Len())

	for pair := root.Properties.Oldest(); pair != nil; pair = pair.Next() {
		keys = append(keys, pair.Key)
	}

	return keys, nil
}

// upgradeCredentials conforms every stored credential onto its current slot and replaces the stored rows
func (r *Runtime) upgradeCredentials(ctx context.Context, req types.InstallationRequest, installation *ent.Integration, def types.Definition, records map[types.CredentialSlotID]types.CredentialSet) error {
	stored := make(map[string]json.RawMessage, len(records))

	for slot, credential := range records {
		stored[slot.String()] = credential.Data
	}

	conformed, err := conformDocuments(ctx, req, credentialKind(def), stored)
	if err != nil {
		return err
	}

	next := make(map[types.CredentialSlotID]types.CredentialSet, len(conformed))

	for name, document := range conformed {
		next[types.NewCredentialSlotID(name)] = types.CredentialSet{Data: document}
	}

	if err := r.keystore().ReplaceCredentials(ctx, installation, records, next); err != nil {
		return fmt.Errorf("replace credentials: %w", err)
	}

	return nil
}

// upgradeUserInput conforms the stored user input document onto the current layout name
func (r *Runtime) upgradeUserInput(ctx context.Context, req types.InstallationRequest, installation *ent.Integration, def types.Definition) error {
	if def.UserInput == nil {
		return nil
	}

	stored := map[string]json.RawMessage{}

	if installation.UserInput.Layout != "" || !jsonx.IsEmptyRawMessage(installation.UserInput.Data) {
		stored[installation.UserInput.Layout] = installation.UserInput.Data
	}

	conformed, err := conformDocuments(ctx, req, userInputKind(def), stored)
	if err != nil {
		return err
	}

	if len(conformed) == 0 {
		return nil
	}

	next := types.IntegrationUserInput{Layout: def.UserInput.Name, Data: conformed[def.UserInput.Name]}

	if next.Layout == installation.UserInput.Layout && bytes.Equal(next.Data, installation.UserInput.Data) {
		return nil
	}

	return r.persistUserInput(ctx, installation, next, r.DB().Integration.UpdateOneID(installation.ID))
}

// upgradeOperationConfig conforms every stored operation document onto its current operation name
func (r *Runtime) upgradeOperationConfig(ctx context.Context, req types.InstallationRequest, installation *ent.Integration, def types.Definition) error {
	stored := installation.OperationConfig.Operations

	next, err := conformDocuments(ctx, req, operationInputKind(def), stored)
	if err != nil {
		return err
	}

	if maps.EqualFunc(next, stored, func(a, b json.RawMessage) bool { return bytes.Equal(a, b) }) {
		return nil
	}

	return r.persistOperationConfig(ctx, installation, types.IntegrationOperationConfig{Operations: next})
}

// upgradeOperations moves health and run history off retired operation names, cancels their loops
func (r *Runtime) upgradeOperations(ctx context.Context, installation *ent.Integration, def types.Definition) error {
	unhealthy := lo.Assign(installation.Health.UnhealthyOperations)

	for _, operation := range def.Operations {
		retired := operation.Replaces
		if len(retired) == 0 {
			continue
		}

		for _, old := range retired {
			purged, err := r.purgeReconcileLoop(ctx, installation.ID, old)
			if err != nil {
				return err
			}

			if purged > 0 {
				logx.FromContext(ctx).Info().Str("retired_operation", old).Str("operation", operation.Name).Int("purged", purged).Msg("cancelled reconcile loops queued under a retired operation name")
			}

			reason, recorded := unhealthy[old]
			if !recorded {
				continue
			}

			if _, kept := unhealthy[operation.Name]; !kept {
				unhealthy[operation.Name] = reason
			}

			delete(unhealthy, old)
		}

		if err := r.DB().IntegrationRun.Update().
			Where(integrationrun.IntegrationIDEQ(installation.ID), integrationrun.OperationNameIn(retired...)).
			SetOperationName(operation.Name).
			Exec(ctx); err != nil {
			return err
		}
	}

	declared := lo.SliceToMap(def.Operations, func(operation types.OperationRegistration) (string, struct{}) {
		return operation.Name, struct{}{}
	})

	maps.DeleteFunc(unhealthy, func(name string, _ string) bool {
		_, ok := declared[name]

		return !ok
	})

	if maps.Equal(unhealthy, installation.Health.UnhealthyOperations) {
		return nil
	}

	health := installation.Health
	health.UnhealthyOperations = unhealthy

	if len(unhealthy) == 0 {
		health.UnhealthyOperations = nil
	}

	if err := r.DB().Integration.UpdateOneID(installation.ID).SetHealth(health).Exec(ctx); err != nil {
		return err
	}

	installation.Health = health

	return nil
}

// upgradeWebhookDeliveries moves persisted delivery rows off retired contract names; endpoint rows are reconciled by ensureWebhook
func (r *Runtime) upgradeWebhookDeliveries(ctx context.Context, installation *ent.Integration, def types.Definition) error {
	for _, webhook := range def.Webhooks {
		if len(webhook.Replaces) == 0 {
			continue
		}

		if err := r.DB().IntegrationWebhook.Update().
			Where(
				integrationwebhook.IntegrationIDEQ(installation.ID),
				integrationwebhook.NameIn(webhook.Replaces...),
				integrationwebhook.ExternalEventIDNotNil(),
			).
			SetName(webhook.Name).
			Exec(ctx); err != nil {
			return err
		}
	}

	return r.reconcileInstallationWebhooks(ctx, installation, "")
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

// conformDocuments conforms every stored document onto its current name, dropping a retired document whose replacement is already stored
func conformDocuments(ctx context.Context, req types.InstallationRequest, kind documentKind, stored map[string]json.RawMessage) (map[string]json.RawMessage, error) {
	next := maps.Clone(stored)

	if next == nil {
		next = map[string]json.RawMessage{}
	}

	names := lo.Keys(stored)

	slices.Sort(names)

	for _, name := range names {
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
			return nil, fmt.Errorf("%w: %s %s", err, kind.label, name)
		}

		next[layout.name] = document

		if layout.replaced {
			delete(next, name)
		}
	}

	return next, nil
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
