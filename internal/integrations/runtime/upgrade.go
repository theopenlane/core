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

// upgradeInstallation conforms stored credentials and user input to the current definition and moves data stored under retired slot, operation and webhook names onto their replacements
func (r *Runtime) upgradeInstallation(ctx context.Context, installation *ent.Integration, skip []types.CredentialSlotID) error {
	def, err := r.resolveDefinitionForInstallation(installation)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to resolve definition for installation")

		return err
	}

	records, err := r.keystore().LoadAllCredentials(ctx, installation)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to load all credentials for installation")

		return err
	}

	providerState, err := def.ProviderState(installation.ProviderState)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to resolve provider state for installation")

		return err
	}

	connection, err := r.resolvePersistedConnection(def, installation)
	if err != nil && !errors.Is(err, ErrConnectionRequired) && !errors.Is(err, ErrConnectionNotFound) {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to resolve persisted connection for installation")

		return err
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
			payload, err := conformPayload(ctx, req, def.CredentialSchema(slot), registration.Backfill, credential.Data, ErrCredentialInvalid)
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

		if converted, err = conformPayload(ctx, req, def.CredentialSchema(registration.Ref), registration.Backfill, converted, ErrCredentialInvalid); err != nil {
			return fmt.Errorf("%w: slot %s", err, slot)
		}

		if _, stored := records[registration.Ref]; !stored {
			next[registration.Ref] = types.CredentialSet{Data: converted}
		}

		delete(next, slot)
	}

	if err := r.keystore().ReplaceCredentials(ctx, installation, records, next); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to replace credentials in keystore")

		return err
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
			logx.FromContext(ctx).Error().Err(err).Msg("failed to persist connection state after credential slot replacement")

			return err
		}
	}

	if err := r.upgradeUserInput(systemCtx, req, installation, def); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to upgrade installation user input")

		return err
	}

	if err := r.upgradeOperations(systemCtx, installation, def); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to upgrade installation operation records")

		return err
	}

	if err := r.upgradeWebhooks(systemCtx, installation, def); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to upgrade installation webhook rows")

		return err
	}

	if len(skip) > 0 {
		return nil
	}

	version := r.Registry().Version(def.ID)

	if err := r.DB().Integration.UpdateOneID(installation.ID).SetDefinitionVersion(version).Exec(systemCtx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to update installation to current definition version")

		return err
	}

	installation.DefinitionVersion = version

	if err := r.RefreshInstallationMetadata(ctx, installation); err != nil {
		logx.FromContext(ctx).Debug().Err(err).Msg("upgrade: installation metadata refresh failed")

		return err
	}

	renamed := lo.SomeBy(def.Operations, func(operation types.OperationRegistration) bool {
		return len(operation.Replaces) > 0
	})

	if !renamed {
		return nil
	}

	if err := r.ResetReconcileLoops(ctx, installation); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to reseed reconcile loops after operation rename")

		return err
	}

	return nil
}

// upgradeUserInput conforms the stored user input to the current definition, converting it from a retired layout when it no longer validates
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

// conformUserInput conforms stored user input to the registered schema, converting it from a retired layout when it no longer validates, then backfilling and validating the result through conformPayload
func conformUserInput(ctx context.Context, req types.InstallationRequest, input types.UserInputRegistration, stored json.RawMessage) (json.RawMessage, error) {
	document := stored

	if validatePayload(ctx, input.Schema, stored, ErrUserInputInvalid) != nil && input.Convert != nil {
		if converted, convertErr := input.Convert(stored); convertErr == nil {
			document = converted
		}
	}

	return conformPayload(ctx, req, input.Schema, input.Backfill, document, ErrUserInputInvalid)
}

// upgradeOperations moves recorded operation health and run history stored under retired operation names onto the operations that replace them, cancels reconcile loops still queued under retired names, and drops health records of operations the definition no longer declares
func (r *Runtime) upgradeOperations(ctx context.Context, installation *ent.Integration, def types.Definition) error {
	unhealthy := map[string]string{}

	maps.Copy(unhealthy, installation.Health.UnhealthyOperations)

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

// upgradeWebhooks renames persisted webhook rows stored under retired contract names onto the contracts that replace them, keeping their endpoint and secret, and refreshes each endpoint row's allowed events
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

// upgradeExclusions expands the skipped slots to include every retired slot a skipped slot's ref replaces, so neither the skipped payload nor its source is processed or deleted
func upgradeExclusions(def types.Definition, skip []types.CredentialSlotID) []types.CredentialSlotID {
	return lo.FlatMap(skip, func(slot types.CredentialSlotID, _ int) []types.CredentialSlotID {
		registration, err := def.CredentialRegistration(slot)
		if err != nil {
			return []types.CredentialSlotID{slot}
		}

		return append(slices.Clone(registration.Replaces), slot)
	})
}

// conformPayload strips and defaults the payload to the schema, applies the declared backfill when one is set, then validates the result once
func conformPayload(ctx context.Context, req types.InstallationRequest, schema json.RawMessage, backfill types.BackfillFunc, payload json.RawMessage, sentinel error) (json.RawMessage, error) {
	conformed, err := jsonx.ConformToSchema(schema, payload)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to conform payload to schema")

		return nil, err
	}

	if backfill != nil {
		if conformed, err = backfill(ctx, req, conformed); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to backfill payload")

			return nil, err
		}
	}

	if err := validatePayload(ctx, schema, conformed, sentinel); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to validate backfilled payload")

		return nil, err
	}

	return conformed, nil
}
