package graphapi

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessment"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/logx"
)

func createAssessmentTemplate(ctx context.Context, input model.CreateAssessmentTemplateInput) (*model.AssessmentTemplateCreatePayload, error) {
	txnCtx := withTransactionalMutation(ctx)

	obj, err := txnCtx.Assessment.Query().
		Where(assessment.ID(input.AssessmentID)).
		Only(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "assessment"})
	}

	ctx, err = common.SetOrganizationInAuthContext(ctx, &obj.OwnerID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to set organization in auth context")

		return nil, rout.ErrPermissionDenied
	}

	if obj.Jsonconfig == nil {
		return nil, parseRequestError(ctx, fmt.Errorf("%w: assessment jsonconfig is required", common.ErrInvalidInput), common.Action{Action: common.ActionCreate, Object: "template"})
	}

	name := obj.Name
	if input.Name != nil && *input.Name != "" {
		name = *input.Name
	}

	creationInput := generated.CreateTemplateInput{
		Name:         name,
		TemplateType: lo.ToPtr(enums.Document),
		Kind:         lo.ToPtr(enums.TemplateKindQuestionnaire),
		Jsonconfig:   obj.Jsonconfig,
		Uischema:     obj.Uischema,
		OwnerID:      lo.ToPtr(obj.OwnerID),
	}

	if input.Description != nil {
		creationInput.Description = input.Description
	}

	if len(input.Tags) > 0 {
		creationInput.Tags = input.Tags
	}

	t, err := txnCtx.Template.Create().SetInput(creationInput).Save(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "template"})
	}

	return &model.AssessmentTemplateCreatePayload{
		Template: t,
	}, nil
}
