package runtime

import (
	"context"
	"time"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/entity"
	"github.com/theopenlane/core/v2/internal/ent/generated/entitytype"
	"github.com/theopenlane/core/v2/internal/ent/generated/subprocessor"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/metrics"
)

// PendingInstallationTTL is how long a pending installation may wait for its auth flow
const PendingInstallationTTL = 168 * time.Hour

// IntegrationLookup holds the query constraints for resolving an integration
type IntegrationLookup struct {
	// IntegrationID is the unique identifier of the integration installation and required
	IntegrationID string
	// OwnerID scopes the integration to a specific owner, if provided
	OwnerID string
	// DefinitionID validates the integration belongs to a specific definition, if provided
	DefinitionID string
}

// ResolveIntegration resolves an integration by ID with optional owner/definition checks
func (r *Runtime) ResolveIntegration(ctx context.Context, lookup IntegrationLookup) (*ent.Integration, error) {
	return operations.ResolveIntegration(ctx, r.DB(), lookup.IntegrationID, lookup.OwnerID, lookup.DefinitionID)
}

// ResolveOwnerIntegration finds a connected integration for the given definition and owner
func (r *Runtime) ResolveOwnerIntegration(ctx context.Context, definitionID, ownerID string, prefer ...func(*ent.Integration) bool) (string, error) {
	return operations.ResolveOwnerIntegration(ctx, r.DB(), definitionID, ownerID, prefer...)
}

// EnsureInstallation returns an existing installation, or creates a new one
func (r *Runtime) EnsureInstallation(ctx context.Context, ownerID, integrationID string, def types.Definition) (*ent.Integration, bool, error) {
	if integrationID != "" {
		record, err := r.ResolveIntegration(ctx, IntegrationLookup{
			IntegrationID: integrationID,
			OwnerID:       ownerID,
			DefinitionID:  def.ID,
		})
		if err != nil {
			return nil, false, err
		}

		return record, false, nil
	}

	record, err := r.DB().Integration.Create().
		SetOwnerID(ownerID).
		SetName(def.DisplayName).
		SetDescription(def.Description).
		SetKind(def.Family).
		SetIntegrationType(def.Category).
		SetDefinitionID(def.ID).
		SetDefinitionVersion(r.Registry().Version(def.ID)).
		SetDefinitionSlug(def.ID).
		SetFamily(def.Family).
		SetStatus(enums.IntegrationStatusPending).
		SetExpiresAt(time.Now().Add(PendingInstallationTTL)).
		Save(ctx)
	if err != nil {
		return nil, false, err
	}

	metrics.RecordIntegrationInstalled(def.ID)

	r.createVendor(ctx, ownerID, def, record.ID)

	return record, true, nil
}

// createVendor best-effort links or creates the integration family as a vendor in the org
func (r *Runtime) createVendor(ctx context.Context, ownerID string, def types.Definition, integrationID string) {
	ctx = logx.WithFields(ctx, map[string]any{"vendor": def.Family, "org_id": ownerID})

	vendorIDs, err := r.DB().Entity.Query().Where(
		entity.Or(
			entity.NameEqualFold(def.Family),
			entity.DisplayNameEqualFold(def.Family),
		),
		entity.OwnerID(ownerID),
	).IDs(ctx)
	if err != nil {
		logx.FromContext(ctx).Info().Err(err).Msg("error looking for existing vendor, skipping creation")
		return
	}

	if len(vendorIDs) > 0 {
		// update the integration edges
		if err := r.DB().Entity.Update().Where(entity.IDIn(vendorIDs...)).AddIntegrationIDs(
			integrationID).Exec(ctx); err != nil {
			logx.FromContext(ctx).Info().Err(err).Msg("error update vendor edges to integration")
		}

		logx.FromContext(ctx).Debug().Msg("successfully updated vendor from integration setup")

		return
	}

	vendorInput := ent.CreateEntityInput{
		Name:           &def.Family,
		Tags:           []string{"integration"},
		ApprovedForUse: lo.ToPtr(true),
		IntegrationIDs: []string{integrationID},
	}

	subprocessors, err := r.DB().Subprocessor.Query().Where(
		subprocessor.NameEqualFold(def.Family),
	).All(ctx)
	if err == nil && len(subprocessors) > 0 {
		vendorInput.Description = &subprocessors[0].Description
	}

	existingEntityType, err := r.DB().EntityType.Query().
		Where(
			entitytype.NameEqualFold("vendor"),
			entitytype.OwnerID(ownerID),
		).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Info().Err(err).Msg("error looking up vendor entity type, skipping creation")
		return
	}

	if err := r.DB().Entity.Create().SetInput(vendorInput).SetEntityTypeID(existingEntityType.ID).Exec(ctx); err != nil {
		logx.FromContext(ctx).Info().Err(err).Msg("error creating vendor")
		return
	}

	logx.FromContext(ctx).Debug().Msg("successfully created vendor from integration setup")
}
