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

// upgradeInstallation conforms every stored credential to the current definition
// it moves payloads stored under retired slots onto the slots that replace them
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
			payload, err := conformCredential(ctx, registration.Ref, req, credential.Data)
			if err != nil {
				return fmt.Errorf("%w: slot %s", err, slot)
			}

			next[slot] = types.CredentialSet{Data: payload}

			continue
		}

		registration, found := replacingRegistration(def, slot)
		if !found {
			logx.FromContext(ctx).Warn().Str("slot", slot.String()).Msg("stored credential slot is not declared by the definition and was left untouched")

			continue
		}

		converted, err := registration.Ref.Convert(slot, credential.Data)
		if err != nil {
			return fmt.Errorf("%w: slot %s", err, slot)
		}

		if converted, err = conformCredential(ctx, registration.Ref, req, converted); err != nil {
			return fmt.Errorf("%w: slot %s", err, slot)
		}

		if _, stored := records[registration.Ref.ID()]; !stored {
			next[registration.Ref.ID()] = types.CredentialSet{Data: converted}
		}

		delete(next, slot)
	}

	if err := r.keystore().ReplaceCredentials(ctx, installation, records, next); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to replace credentials in keystore")

		return err
	}

	credentialRef := providerState.CredentialRef

	if _, err := def.CredentialRegistration(credentialRef); err != nil {
		if registration, found := replacingRegistration(def, credentialRef); found {
			credentialRef = registration.Ref.ID()
		}
	}

	// ensure we can make the necessary updates to the connection state
	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	if credentialRef != providerState.CredentialRef {
		if err := r.persistConnectionState(systemCtx, installation, def, credentialRef); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to persist connection state after credential slot replacement")

			return err
		}
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

	return nil
}

// upgradeExclusions expands the skipped slots to include every retired slot a skipped slot's ref replaces, so neither the skipped payload nor its source is processed or deleted
func upgradeExclusions(def types.Definition, skip []types.CredentialSlotID) []types.CredentialSlotID {
	return lo.FlatMap(skip, func(slot types.CredentialSlotID, _ int) []types.CredentialSlotID {
		registration, err := def.CredentialRegistration(slot)
		if err != nil {
			return []types.CredentialSlotID{slot}
		}

		return append(registration.Ref.Replaces(), slot)
	})
}

// replacingRegistration finds the declared registration whose slot takes over payloads stored under the retired slot
func replacingRegistration(def types.Definition, retired types.CredentialSlotID) (types.CredentialRegistration, bool) {
	return lo.Find(def.CredentialRegistrations, func(reg types.CredentialRegistration) bool {
		return lo.Contains(reg.Ref.Replaces(), retired)
	})
}

// conformCredential strips and defaults the payload to the slot schema
func conformCredential(ctx context.Context, slot types.CredentialSlot, req types.InstallationRequest, payload json.RawMessage) (json.RawMessage, error) {
	schema := slot.Schema()

	conformed, err := jsonx.ConformToSchema(schema, payload)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to conform credential to slot schema")

		return nil, err
	}

	invalid := validatePayload(ctx, schema, conformed, ErrCredentialInvalid)
	if invalid == nil {
		return conformed, nil
	}

	if !slot.Backfills() {
		return nil, invalid
	}

	backfilled, err := slot.Backfill(ctx, req, conformed)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to backfill credential")

		return nil, err
	}

	if err := validatePayload(ctx, schema, backfilled, ErrCredentialInvalid); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to validate backfilled credential")

		return nil, err
	}

	return backfilled, nil
}
