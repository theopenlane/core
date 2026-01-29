package graphapi

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/campaign"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/utils/rout"
)

// createCampaignWithTargets creates a campaign and its targets in a single transaction.
func (r *mutationResolver) createCampaignWithTargets(ctx context.Context, campaignInput *generated.CreateCampaignInput, targets []*generated.CreateCampaignTargetInput) (*model.CampaignCreateWithTargetsPayload, error) {
	if campaignInput == nil {
		return nil, rout.NewMissingRequiredFieldError("campaign")
	}

	// set the organization in the auth context if its not done for us
	if err := common.SetOrganizationInAuthContext(ctx, campaignInput.OwnerID); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to set organization in auth context")

		if campaignInput.OwnerID == nil || *campaignInput.OwnerID == "" {
			return nil, rout.NewMissingRequiredFieldError("owner_id")
		}

		return nil, rout.ErrPermissionDenied
	}

	targetCount := 0
	for _, target := range targets {
		if target != nil {
			targetCount++
		}
	}

	if targetCount == 0 {
		return nil, rout.NewMissingRequiredFieldError("targets")
	}

	if campaignInput.RecipientCount == nil {
		campaignInput.RecipientCount = &targetCount
	}

	res, err := withTransactionalMutation(ctx).Campaign.Create().SetInput(*campaignInput).Save(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "campaign"})
	}

	var createdTargets []*generated.CampaignTarget
	if len(targets) > 0 {
		inputTargets := make([]*generated.CreateCampaignTargetInput, 0, len(targets))
		for _, target := range targets {
			if target == nil {
				continue
			}

			target.CampaignID = res.ID
			inputTargets = append(inputTargets, target)
		}

		if len(inputTargets) > 0 {
			payload, err := r.bulkCreateCampaignTarget(ctx, inputTargets)
			if err != nil {
				return nil, err
			}

			createdTargets = payload.CampaignTargets
		}
	}

	query, err := withTransactionalMutation(ctx).Campaign.
		Query().
		Where(campaign.ID(res.ID)).
		CollectFields(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "campaign"})
	}

	finalResult, err := query.Only(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "campaign"})
	}

	return &model.CampaignCreateWithTargetsPayload{
		Campaign:        finalResult,
		CampaignTargets: createdTargets,
	}, nil
}
