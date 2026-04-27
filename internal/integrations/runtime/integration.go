package runtime

import (
	"context"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/entity"
	"github.com/theopenlane/core/internal/ent/generated/entitytype"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/subprocessor"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// IntegrationLookup holds the query constraints for resolving an integration
type IntegrationLookup struct {
	// IntegrationID is the unique identifier of the integration installation and required
	IntegrationID string
	// OwnerID scopes the integration to a specific owner, if provided
	OwnerID string
	// DefinitionID validates the integration belongs to a specific definition, if provided
	DefinitionID string
}

// ResolveIntegration resolves one integration by explicit ID with optional owner and definition cross-checks
func (r *Runtime) ResolveIntegration(ctx context.Context, lookup IntegrationLookup) (*ent.Integration, error) {
	if lookup.IntegrationID == "" {
		return nil, ErrIntegrationIDRequired
	}

	query := r.DB().Integration.Query().Where(integration.IDEQ(lookup.IntegrationID))
	if lookup.OwnerID != "" {
		query = query.Where(integration.OwnerIDEQ(lookup.OwnerID))
	}

	record, err := query.Only(ctx)
	if err != nil {
		return nil, err
	}

	if lookup.DefinitionID != "" && record.DefinitionID != lookup.DefinitionID {
		return nil, ErrInstallationDefinitionMismatch
	}

	return record, nil
}

// EnsureInstallation returns an existing installation when integrationID is provided, or creates a new one
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
		SetDefinitionID(def.ID).
		SetFamily(def.Family).
		SetStatus(enums.IntegrationStatusPending).
		Save(ctx)
	if err != nil {
		return nil, false, err
	}

	// attempt to create vendor record
	r.createVendor(ctx, ownerID, def, record.ID)

	return record, true, nil
}

// createVendor will to a best-effort create of the integration family as a vendor in the organization
// if it already exists, it will link the integration id
// if it doesn't exist, it will create the record, add data from the system-owned subprocessors, and link the integration
func (r *Runtime) createVendor(ctx context.Context, ownerID string, def types.Definition, integrationID string) {
	vendorIDs, err := r.DB().Entity.Query().Where(
		entity.NameEqualFold(def.Family),
		entity.OwnerID(ownerID),
	).IDs(ctx)
	if err != nil {
		logx.FromContext(ctx).Info().Err(err).Msg("error looking for existing vendor, skipping creation")
		return
	}

	if len(vendorIDs) > 0 {
		// update the integration edges
		ctxAllow := privacy.DecisionContext(ctx, privacy.Allow)
		if err := r.DB().Entity.Update().Where(entity.IDIn(vendorIDs...)).AddIntegrationIDs(
			integrationID).Exec(ctxAllow); err != nil {
			logx.FromContext(ctx).Info().Err(err).Msg("error update vendor edges to integration")
		}
	}

	vendorInput := generated.CreateEntityInput{
		Name:           &def.Family,
		Tags:           []string{"integration"},
		ApprovedForUse: lo.ToPtr(true),
		IntegrationIDs: []string{integrationID},
	}

	// lookup subprocessor for existing data
	subprocessors, err := r.DB().Subprocessor.Query().Where(
		subprocessor.NameEqualFold(def.Family),
	).All(ctx)
	if err == nil && len(subprocessors) > 0 {
		vendorInput.Description = &subprocessors[0].Description
	}

	existingEntityType, err := r.DB().EntityType.Query().
		Where(
			entitytype.NameEqualFold("vendor"),
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

	logx.FromContext(ctx).Debug().Str("vendor", def.Family).Msg("successfully created vendor from integration setup")
}
