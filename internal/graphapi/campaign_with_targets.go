package graphapi

import (
	"context"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/utils/rout"
)

// maxCampaignTargets is the upper bound on recipients per campaign creation
const maxCampaignTargets = 500

// createCampaignWithTargets creates a campaign and its targets in a single transaction.
func (r *mutationResolver) createCampaignWithTargets(ctx context.Context, campaignInput *generated.CreateCampaignInput, targets []*generated.CreateCampaignTargetInput) (*model.CampaignCreateWithTargetsPayload, error) {
	var err error

	ctx, err = validateCampaignWithTargetsInput(ctx, campaignInput, targets)
	if err != nil {
		return nil, err
	}

	validTargets := lo.Compact(targets)
	setRecipientCountIfNeeded(campaignInput, len(validTargets))

	createdCampaign, err := r.createCampaign(ctx, campaignInput)
	if err != nil {
		return nil, err
	}

	createdTargets, err := r.createTargetsForCampaign(ctx, createdCampaign.ID, validTargets)
	if err != nil {
		return nil, err
	}

	finalCampaign, err := r.reloadCampaign(ctx, createdCampaign.ID)
	if err != nil {
		return nil, err
	}

	return &model.CampaignCreateWithTargetsPayload{
		Campaign:        finalCampaign,
		CampaignTargets: createdTargets,
	}, nil
}

// validateCampaignWithTargetsInput validates the campaign input and targets,
// returning the enriched context with the organization set in the auth context
func validateCampaignWithTargetsInput(ctx context.Context, campaignInput *generated.CreateCampaignInput, targets []*generated.CreateCampaignTargetInput) (context.Context, error) {
	if campaignInput == nil {
		return ctx, rout.NewMissingRequiredFieldError("campaign")
	}

	ctx, err := common.SetOrganizationInAuthContext(ctx, campaignInput.OwnerID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to set organization in auth context")

		if lo.FromPtrOr(campaignInput.OwnerID, "") == "" {
			return ctx, rout.NewMissingRequiredFieldError("owner_id")
		}

		return ctx, rout.ErrPermissionDenied
	}

	compacted := lo.Compact(targets)

	if len(compacted) == 0 {
		return ctx, rout.NewMissingRequiredFieldError("targets")
	}

	if len(compacted) > maxCampaignTargets {
		return ctx, ErrCampaignTargetLimitExceeded
	}

	return ctx, nil
}

// setRecipientCountIfNeeded sets the recipient count if not already set.
func setRecipientCountIfNeeded(input *generated.CreateCampaignInput, count int) {
	if input.RecipientCount == nil {
		input.RecipientCount = &count
	}
}

// createCampaign creates the campaign entity.
func (r *mutationResolver) createCampaign(ctx context.Context, input *generated.CreateCampaignInput) (*generated.Campaign, error) {
	res, err := withTransactionalMutation(ctx).Campaign.Create().SetInput(*input).Save(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "campaign"})
	}

	return res, nil
}

// createTargetsForCampaign creates campaign targets linked to the campaign.
func (r *mutationResolver) createTargetsForCampaign(ctx context.Context, campaignID string, targets []*generated.CreateCampaignTargetInput) ([]*generated.CampaignTarget, error) {
	if len(targets) == 0 {
		return nil, nil
	}

	for _, target := range targets {
		target.CampaignID = campaignID
	}

	payload, err := r.bulkCreateCampaignTarget(ctx, targets)
	if err != nil {
		return nil, err
	}

	return payload.CampaignTargets, nil
}

// reloadCampaign fetches the campaign after creation to ensure any hooks or computed fields are reflected.
func (r *mutationResolver) reloadCampaign(ctx context.Context, campaignID string) (*generated.Campaign, error) {
	result, err := withTransactionalMutation(ctx).Campaign.Get(ctx, campaignID)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "campaign"})
	}

	return result, nil
}
