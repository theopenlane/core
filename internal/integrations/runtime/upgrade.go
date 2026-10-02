package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/samber/lo"
	"github.com/theopenlane/utils/contextx"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationrun"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationwebhook"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// upgradedInstallationKey carries the id of the installation whose upgrade already ran in this request
var upgradedInstallationKey = contextx.NewKey[string]()

// ensureCurrentVersion upgrades an installation if versions dont match
func (r *Runtime) ensureCurrentVersion(ctx context.Context, installation *ent.Integration, skip ...types.CredentialSlotID) error {
	if installation.DefinitionVersion == r.Registry().Version(installation.DefinitionID) || upgradedInstallationKey.GetOr(ctx, "") == installation.ID {
		return nil
	}

	if err := r.upgradeInstallation(ctx, installation, skip); err != nil {
		failed := fmt.Errorf("%w: %w", ErrInstallationUpgradeFailed, err)

		if markErr := r.MarkIntegrationUnhealthy(ctx, installation, failed.Error()); markErr != nil {
			failed = errors.Join(failed, markErr)
		}

		return types.Unhealthy(failed, failed.Error())
	}

	return nil
}

// upgradeInstallation conforms stored credentials and user input, moving data off retired names
func (r *Runtime) upgradeInstallation(ctx context.Context, installation *ent.Integration, skip []types.CredentialSlotID) error {
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

	next := maps.Clone(records)
	excluded := upgradeExclusions(def, skip)

	slots := lo.Keys(records)

	slices.SortFunc(slots, func(a, b types.CredentialSlotID) int {
		return strings.Compare(a.String(), b.String())
	})

	for _, slot := range slots {
		if lo.Contains(excluded, slot) {
			continue
		}

		registration, replaced, ok := def.ResolveCredential(slot)
		if !ok {
			logx.FromContext(ctx).Warn().Str("slot", slot.String()).Msg("stored credential slot is not declared by the definition and was left untouched")

			continue
		}

		payload, err := conformStored(ctx, req, registration.StoredSchema, registration.Upgrade, registration.Validate, slot.String(), records[slot].Data, ErrCredentialInvalid)
		if err != nil {
			return fmt.Errorf("%w: slot %s", err, slot)
		}

		if !replaced {
			next[slot] = types.CredentialSet{Data: payload}

			continue
		}

		if _, stored := records[registration.Ref]; !stored {
			next[registration.Ref] = types.CredentialSet{Data: payload}
		}

		delete(next, slot)
	}

	if err := r.keystore().ReplaceCredentials(ctx, installation, records, next); err != nil {
		return fmt.Errorf("replace credentials: %w", err)
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

	if len(skip) > 0 {
		return nil
	}

	version := r.Registry().Version(def.ID)

	if err := r.DB().Integration.UpdateOneID(installation.ID).SetDefinitionVersion(version).Exec(systemCtx); err != nil {
		return fmt.Errorf("set definition version: %w", err)
	}

	installation.DefinitionVersion = version

	if err := r.RefreshInstallationMetadata(ctx, installation); err != nil {
		return fmt.Errorf("refresh installation metadata: %w", err)
	}

	renamed := lo.SomeBy(def.Operations, func(operation types.OperationRegistration) bool {
		return len(operation.Replaces) > 0
	})

	if !renamed {
		return nil
	}

	if err := r.ResetReconcileLoops(ctx, installation); err != nil {
		return fmt.Errorf("reset reconcile loops: %w", err)
	}

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

// upgradeUserInput upgrades stored user input from the layout it was persisted under and conforms it to the current layout
func (r *Runtime) upgradeUserInput(ctx context.Context, req types.InstallationRequest, installation *ent.Integration, def types.Definition) error {
	if def.UserInput == nil {
		return nil
	}

	conformed, err := conformStored(ctx, req, def.UserInput.Schema, def.UserInput.Upgrade, def.UserInput.Validate, installation.UserInput.Layout, installation.UserInput.Data, ErrUserInputInvalid)
	if err != nil {
		return fmt.Errorf("%w: layout %s", err, installation.UserInput.Layout)
	}

	next := types.IntegrationUserInput{Layout: def.UserInput.Name, Data: conformed}

	if next.Layout == installation.UserInput.Layout && bytes.Equal(next.Data, installation.UserInput.Data) {
		return nil
	}

	return r.persistUserInput(ctx, installation, next, r.DB().Integration.UpdateOneID(installation.ID))
}

// upgradeOperationConfig upgrades each stored operation document, moving retired names onto their replacements
func (r *Runtime) upgradeOperationConfig(ctx context.Context, req types.InstallationRequest, installation *ent.Integration, def types.Definition) error {
	stored := installation.OperationConfig.Operations

	next, err := upgradeOperationDocuments(ctx, req, def, stored)
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

// upgradeExclusions expands skipped slots to include every retired slot they replace
func upgradeExclusions(def types.Definition, skip []types.CredentialSlotID) []types.CredentialSlotID {
	return lo.FlatMap(skip, func(slot types.CredentialSlotID, _ int) []types.CredentialSlotID {
		registration, err := def.CredentialRegistration(slot)
		if err != nil {
			return []types.CredentialSlotID{slot}
		}

		return append(slices.Clone(registration.Replaces), slot)
	})
}

// upgradeOperationDocuments upgrades each stored operation document by the name it was persisted under, moving retired names onto their replacements
func upgradeOperationDocuments(ctx context.Context, req types.InstallationRequest, def types.Definition, stored map[string]json.RawMessage) (map[string]json.RawMessage, error) {
	next := maps.Clone(stored)

	if next == nil {
		next = map[string]json.RawMessage{}
	}

	names := lo.Keys(stored)

	slices.Sort(names)

	for _, name := range names {
		operation, replaced, ok := def.ResolveOperation(name)
		if !ok || operation.Input == nil {
			logx.FromContext(ctx).Warn().Str("operation", name).Msg("stored operation input is not declared by the definition and was left untouched")

			continue
		}

		conformed, err := conformStored(ctx, req, operation.Input.Schema, operation.Input.Upgrade, operation.Input.Validate, name, stored[name], types.ErrOperationConfigInvalid)
		if err != nil {
			return nil, fmt.Errorf("%w: operation %s", err, name)
		}

		if !replaced {
			next[name] = conformed

			continue
		}

		if _, kept := stored[operation.Name]; !kept {
			next[operation.Name] = conformed
		}

		delete(next, name)
	}

	return next, nil
}

// conformStored upgrades a stored document from the name it was persisted under, strips and defaults it against schema, then validates it
func conformStored(ctx context.Context, req types.InstallationRequest, schema json.RawMessage, upgrade types.UpgradeFunc, validate types.ValidateFunc, from string, stored json.RawMessage, sentinel error) (json.RawMessage, error) {
	document := stored

	if upgrade != nil {
		upgraded, err := upgrade(ctx, req, from, stored)
		if err != nil {
			return nil, fmt.Errorf("upgrade: %w", err)
		}

		document = upgraded
	}

	conformed, err := jsonx.ConformToSchema(schema, document)
	if err != nil {
		return nil, fmt.Errorf("conform to schema: %w", err)
	}

	if err := operations.ValidateInput(ctx, req, schema, validate, conformed, sentinel); err != nil {
		return nil, err
	}

	return conformed, nil
}
