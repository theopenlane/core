package graphapi

import (
	"context"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi/common"
	"github.com/theopenlane/core/v2/internal/graphapi/model"
)

func leaveOrganization(ctx context.Context, organizationID string) (*model.OrgMembershipDeletePayload, error) {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller.SubjectID == "" {
		return nil, rout.ErrPermissionDenied
	}

	isMember, err := common.CheckCallerOrgMembership(ctx, withTransactionalMutation(ctx).Authz, organizationID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, rout.ErrPermissionDenied
	}

	// act as the caller in the org being left, internal operation allows deleting their own membership
	leaveCtx := auth.WithInternalOperationContext(auth.WithCallerScopedToOrg(ctx, organizationID))

	res, err := withTransactionalMutation(ctx).OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(organizationID),
			orgmembership.UserID(caller.SubjectID),
		).
		Only(leaveCtx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionDelete, Object: "orgmembership"})
	}

	if res.Role == enums.RoleOwner {
		return nil, hooks.ErrOrgOwnerCannotBeDeleted
	}

	// group and program memberships scoped to the organization are removed by the
	// org membership delete hook
	if err := withTransactionalMutation(ctx).OrgMembership.DeleteOneID(res.ID).Exec(leaveCtx); err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionDelete, Object: "orgmembership"})
	}

	return &model.OrgMembershipDeletePayload{
		DeletedID: res.ID,
	}, nil
}
