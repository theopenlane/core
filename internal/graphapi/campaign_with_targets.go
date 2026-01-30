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

// createCampaignWithTargets creates a campaign and its targets in a single transaction.
func (r *mutationResolver) createCampaignWithTargets(ctx context.Context, campaignInput *generated.CreateCampaignInput, targets []*generated.CreateCampaignTargetInput) (*model.CampaignCreateWithTargetsPayload, error) {
	if err := validateCampaignWithTargetsInput(ctx, campaignInput, targets); err != nil {
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

// validateCampaignWithTargetsInput validates the campaign input and targets.
func validateCampaignWithTargetsInput(ctx context.Context, campaignInput *generated.CreateCampaignInput, targets []*generated.CreateCampaignTargetInput) error {
	if campaignInput == nil {
		return rout.NewMissingRequiredFieldError("campaign")
	}

	if err := common.SetOrganizationInAuthContext(ctx, campaignInput.OwnerID); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to set organization in auth context")

		if lo.FromPtrOr(campaignInput.OwnerID, "") == "" {
			return rout.NewMissingRequiredFieldError("owner_id")
		}

		return rout.ErrPermissionDenied
	}

	if len(lo.Compact(targets)) == 0 {
		return rout.NewMissingRequiredFieldError("targets")
	}

	return nil
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
