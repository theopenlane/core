package graphapi

import (
	"context"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
)

func leaveOrganization(ctx context.Context, organizationID string) (*model.OrgMembershipDeletePayload, error) {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil || caller.SubjectID == "" {
		return nil, rout.ErrPermissionDenied
	}

	allowCtx := auth.WithCaller(privacy.DecisionContext(ctx, privacy.Allow), caller)

	res, err := withTransactionalMutation(ctx).OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(organizationID),
			orgmembership.UserID(caller.SubjectID),
		).
		Only(allowCtx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionDelete, Object: "orgmembership"})
	}

	if res.Role == enums.RoleOwner {
		return nil, hooks.ErrOrgOwnerCannotBeDeleted
	}

	if err := withTransactionalMutation(ctx).OrgMembership.DeleteOneID(res.ID).Exec(allowCtx); err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionDelete, Object: "orgmembership"})
	}

	if err := generated.OrgMembershipEdgeCleanup(ctx, res.UserID); err != nil {
		return nil, common.NewCascadeDeleteError(ctx, err)
	}

	return &model.OrgMembershipDeletePayload{
		DeletedID: res.ID,
	}, nil
}
