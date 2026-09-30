package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/samber/lo"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationrun"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationwebhook"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// ensureCurrentVersion upgrades an installation if versions dont match
func (r *Runtime) ensureCurrentVersion(ctx context.Context, installation *ent.Integration, skip ...types.CredentialSlotID) error {
	if installation.DefinitionVersion == r.Registry().Version(installation.DefinitionID) {
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

	records, err := r.keystore().LoadAllCredentials(ctx, installation)
	if err != nil {
		return fmt.Errorf("load credentials: %w", err)
	}

	providerState, err := def.ProviderState(installation.ProviderState)
	if err != nil {
		return fmt.Errorf("resolve provider state: %w", err)
	}

	connection, err := r.resolvePersistedConnection(def, installation)
	if err != nil && !errors.Is(err, ErrConnectionRequired) && !errors.Is(err, ErrConnectionNotFound) {
		return fmt.Errorf("resolve persisted connection: %w", err)
	}

	req := types.InstallationRequest{
		Integration: installation,
		Connection:  connection,
		Credentials: lo.MapToSlice(records, func(slot types.CredentialSlotID, credential types.CredentialSet) types.CredentialBinding {
			return types.CredentialBinding{Ref: slot, Credential: credential}
		}),
		Config: installation.Config,
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

		credential := records[slot]

		registration, err := def.CredentialRegistration(slot)
		if err == nil {
			payload, err := conformPayload(ctx, req, registration.StoredSchema, registration.Backfill, credential.Data, ErrCredentialInvalid)
			if err != nil {
				return fmt.Errorf("%w: slot %s", err, slot)
			}

			next[slot] = types.CredentialSet{Data: payload}

			continue
		}

		registration, found := def.CredentialReplacing(slot)
		if !found {
			logx.FromContext(ctx).Warn().Str("slot", slot.String()).Msg("stored credential slot is not declared by the definition and was left untouched")

			continue
		}

		converted, err := registration.Convert(slot, credential.Data)
		if err != nil {
			return fmt.Errorf("%w: slot %s", err, slot)
		}

		if converted, err = conformPayload(ctx, req, registration.StoredSchema, registration.Backfill, converted, ErrCredentialInvalid); err != nil {
			return fmt.Errorf("%w: slot %s", err, slot)
		}

		if _, stored := records[registration.Ref]; !stored {
			next[registration.Ref] = types.CredentialSet{Data: converted}
		}

		delete(next, slot)
	}

	if err := r.keystore().ReplaceCredentials(ctx, installation, records, next); err != nil {
		return fmt.Errorf("replace credentials: %w", err)
	}

	credentialRef := providerState.CredentialRef

	if _, err := def.CredentialRegistration(credentialRef); err != nil {
		if registration, found := def.CredentialReplacing(credentialRef); found {
			credentialRef = registration.Ref
		}
	}

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	if credentialRef != providerState.CredentialRef {
		if err := r.persistConnectionState(systemCtx, installation, def, credentialRef); err != nil {
			return fmt.Errorf("persist connection state: %w", err)
		}
	}

	if err := r.upgradeUserInput(systemCtx, req, installation, def); err != nil {
		return fmt.Errorf("upgrade user input: %w", err)
	}

	if err := r.upgradeOperations(systemCtx, installation, def); err != nil {
		return fmt.Errorf("upgrade operations: %w", err)
	}

	if err := r.upgradeWebhooks(systemCtx, installation, def); err != nil {
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

// upgradeUserInput conforms stored user input to the current definition, converting retired layouts
func (r *Runtime) upgradeUserInput(ctx context.Context, req types.InstallationRequest, installation *ent.Integration, def types.Definition) error {
	if def.UserInput == nil {
		return nil
	}

	conformed, err := conformUserInput(ctx, req, *def.UserInput, installation.Config.ClientConfig)
	if err != nil {
		return err
	}

	config := installation.Config
	config.ClientConfig = conformed

	if err := r.DB().Integration.UpdateOneID(installation.ID).SetConfig(config).Exec(ctx); err != nil {
		return err
	}

	installation.Config = config

	r.keystore().InvalidateClients(installation.ID)

	return nil
}

// conformUserInput converts a retired layout when needed, then backfills and validates the result
func conformUserInput(ctx context.Context, req types.InstallationRequest, input types.UserInputRegistration, stored json.RawMessage) (json.RawMessage, error) {
	document := stored

	if input.Convert != nil {
		if result, err := jsonx.ValidateSchema(input.Schema, stored); err != nil || !result.Valid() {
			if converted, convertErr := input.Convert(stored); convertErr == nil {
				document = converted
			}
		}
	}

	return conformPayload(ctx, req, input.Schema, input.Backfill, document, ErrUserInputInvalid)
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
			fragment, err := reconcileLoopFragment(installation.ID, old)
			if err != nil {
				return err
			}

			purged, err := r.Gala().PurgeActiveJobsWithMetadata(ctx, fragment)
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

// upgradeWebhooks renames persisted webhook rows off retired contract names
func (r *Runtime) upgradeWebhooks(ctx context.Context, installation *ent.Integration, def types.Definition) error {
	db := r.DB()

	for _, webhook := range def.Webhooks {
		retired := webhook.Replaces

		rows, err := db.IntegrationWebhook.Query().
			Where(
				integrationwebhook.IntegrationIDEQ(installation.ID),
				integrationwebhook.NameIn(append(slices.Clone(retired), webhook.Name)...),
				integrationwebhook.ExternalEventIDIsNil(),
			).
			Order(integrationwebhook.ByCreatedAt()).
			All(ctx)
		if err != nil {
			return err
		}

		if len(rows) == 0 {
			continue
		}

		keep, renamed := lo.Find(rows, func(row *ent.IntegrationWebhook) bool {
			return row.Name != webhook.Name
		})
		if !renamed {
			keep = rows[0]
		}

		stale := lo.FilterMap(rows, func(row *ent.IntegrationWebhook, _ int) (string, bool) {
			return row.ID, row.ID != keep.ID
		})

		if len(stale) > 0 {
			if _, err := db.IntegrationWebhook.Delete().Where(integrationwebhook.IDIn(stale...)).Exec(ctx); err != nil {
				return err
			}
		}

		allowedEvents := lo.Map(webhook.Events, func(event types.WebhookEventRegistration, _ int) string {
			return event.Name
		})

		if err := db.IntegrationWebhook.UpdateOneID(keep.ID).SetName(webhook.Name).SetAllowedEvents(allowedEvents).Exec(ctx); err != nil {
			return err
		}

		if len(retired) == 0 {
			continue
		}

		if err := db.IntegrationWebhook.Update().
			Where(
				integrationwebhook.IntegrationIDEQ(installation.ID),
				integrationwebhook.NameIn(retired...),
				integrationwebhook.ExternalEventIDNotNil(),
			).
			SetName(webhook.Name).
			Exec(ctx); err != nil {
			return err
		}
	}

	return nil
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

// conformPayload strips and defaults the payload to the schema, applies backfill, then validates
func conformPayload(ctx context.Context, req types.InstallationRequest, schema json.RawMessage, backfill types.BackfillFunc, payload json.RawMessage, sentinel error) (json.RawMessage, error) {
	conformed, err := jsonx.ConformToSchema(schema, payload)
	if err != nil {
		return nil, fmt.Errorf("conform to schema: %w", err)
	}

	if backfill != nil {
		if conformed, err = backfill(ctx, req, conformed); err != nil {
			return nil, fmt.Errorf("backfill: %w", err)
		}
	}

	if err := validatePayload(ctx, schema, conformed, sentinel); err != nil {
		return nil, err
	}

	return conformed, nil
}
