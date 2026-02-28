package graphapi

import (
	"context"

	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/utils/rout"
)

func validateCustomDomain(ctx context.Context, id string, r *mutationResolver) (*model.CustomDomainValidatePayload, error) {
	res, err := withTransactionalMutation(ctx).CustomDomain.Get(ctx, id)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "customdomain"})
	}

	ctx, err = common.SetOrganizationInAuthContext(ctx, &res.OwnerID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to set organization in auth context")

		return nil, rout.ErrPermissionDenied
	}

	if _, err := r.db.Job.Insert(ctx, jobspec.ValidateCustomDomainArgs{
		CustomDomainID: id,
	}, nil); err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionUpdate, Object: "customdomain"})
	}

	return &model.CustomDomainValidatePayload{
		CustomDomain: res,
	}, nil
}
