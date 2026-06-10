package handlers

import (
	"context"
	"slices"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/rout"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/fga/generate/modelparse"
	fgamodel "github.com/theopenlane/core/fga/model"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
)

// RolesHandler lists available roles that can be assigned to users in addition to the base organization role
func (h *Handler) RolesHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	return ProcessAuthenticatedRequest(ctx, h, openapi, models.ExampleRolesRequest, models.ExampleRolesReply,
		func(reqCtx context.Context, _ *models.RolesRequest, _ *auth.Caller) (*models.RolesReply, error) {
			roles, err := fgamodel.OrganizationRoles()
			if err != nil {
				logx.FromContext(reqCtx).Error().Err(err).Msg("error retrieving api roles")
				return nil, ErrProcessingRequest
			}

			resp := &models.RolesReply{
				Reply: rout.Reply{Success: true},
				Roles: convertOrgRolesToOpenAPI(roles),
			}

			return resp, nil
		})
}

// AssignOrganizationRolesHandler assigns a role to the provided user or group
func (h *Handler) AssignOrganizationRolesHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	return h.handleRoleMutation(ctx, openapi, false)
}

// DeleteOrganizationRolesHandler removes a role from the provided user or group
func (h *Handler) DeleteOrganizationRolesHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	return h.handleRoleMutation(ctx, openapi, true)
}

// AccountRolesMeHandler returns the roles the user has access to
func (h *Handler) AccountRolesMeHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	if _, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleAccountRolesMeRequest, models.ExampleAccountRolesMeReply, openapi.Registry); err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()
	caller, ok := auth.CallerFromContext(reqCtx)
	if !ok || caller == nil {
		logx.FromContext(reqCtx).Error().Msg("error getting caller from context")
		return h.InternalServerError(ctx, auth.ErrNoAuthUser, openapi)
	}

	orgID, err := h.getOrganizationID("", caller)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}
	if orgID == "" {
		return h.BadRequest(ctx, ErrInvalidInput, openapi)
	}

	roles, err := fgamodel.OrganizationRoles()
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error retrieving organization roles")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	ids := make([]string, 0, len(roles))
	for _, role := range roles {
		ids = append(ids, role.ID)
	}

	req := fgax.ListAccess{
		SubjectType: caller.SubjectType(),
		SubjectID:   caller.SubjectID,
		ObjectID:    orgID,
		ObjectType:  fgax.Kind(generated.TypeOrganization),
		Relations:   ids,
		Context:     utils.NewOrganizationContextKey(caller.SubjectEmail),
	}

	assignedRoles, err := h.DBClient.Authz.ListRelations(reqCtx, req)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Interface("access_request", req).Msg("error checking organization role access")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	return h.Success(ctx, models.AccountRolesMeReply{
		Reply:          rout.Reply{Success: true},
		Roles:          convertOrgRolesToOpenAPI(filterOrganizationRoles(roles, assignedRoles)),
		OrganizationID: orgID,
	}, openapi)
}

func (h *Handler) handleRoleMutation(ctx echo.Context, openapi *OpenAPIContext, isDeleteOp bool) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleOrganizationRolesRequest, models.ExampleOrganizationRolesReply, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(reqCtx)
	if !ok || caller == nil {
		logx.FromContext(reqCtx).Error().Msg("error getting caller from context")
		return h.InternalServerError(ctx, auth.ErrNoAuthUser, openapi)
	}

	orgID, err := h.getOrganizationID(in.OrganizationID, caller)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	if orgID == "" {
		return h.BadRequest(ctx, ErrInvalidInput, openapi)
	}

	allowed, err := h.DBClient.Authz.CheckAccess(reqCtx, fgax.AccessCheck{
		SubjectType: caller.SubjectType(),
		SubjectID:   caller.SubjectID,
		Relation:    fgax.CanEdit,
		ObjectID:    orgID,
		ObjectType:  fgax.Kind(generated.TypeOrganization),
		Context:     utils.NewOrganizationContextKey(caller.SubjectEmail),
	})
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("organization_id", orgID).Msg("error checking organization role management access")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	if !allowed {
		return h.BadRequest(ctx, ErrInvalidInput, openapi)
	}

	tuples := convertOrgRolesToTuples(orgID, in.Role, in.UserIDs, in.GroupIDs)
	writes := tuples
	var deletes []fgax.TupleKey
	if isDeleteOp {
		writes = nil
		deletes = tuples
	}

	if _, err := h.DBClient.Authz.WriteTupleKeys(reqCtx, writes, deletes); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("organization_id", orgID).Str("role", in.Role).Msg("error updating organization roles")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	return h.Success(ctx, models.OrganizationRolesReply{
		Reply:          rout.Reply{Success: true},
		OrganizationID: orgID,
		Role:           in.Role,
	}, openapi)
}

func convertOrgRolesToOpenAPI(roles []modelparse.OrganizationRole) []models.OrganizationRole {
	resp := make([]models.OrganizationRole, 0, len(roles))
	for _, role := range roles {
		resp = append(resp, models.OrganizationRole{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
		})
	}

	return resp
}

func filterOrganizationRoles(roles []modelparse.OrganizationRole, assigned []string) []modelparse.OrganizationRole {
	filtered := make([]modelparse.OrganizationRole, 0, len(assigned))
	for _, role := range roles {
		if slices.Contains(assigned, role.ID) {
			filtered = append(filtered, role)
		}
	}

	return filtered
}

func convertOrgRolesToTuples(orgID, role string, userIDs, groupIDs []string) []fgax.TupleKey {
	tuples := make([]fgax.TupleKey, 0, len(userIDs)+len(groupIDs))
	object := fgax.Entity{
		Kind:       fgax.Kind(generated.TypeOrganization),
		Identifier: orgID,
	}

	for _, userID := range userIDs {
		tuples = append(tuples, fgax.TupleKey{
			Subject: fgax.Entity{
				Kind:       fgax.Kind(generated.TypeUser),
				Identifier: userID,
			},
			Object:   object,
			Relation: fgax.Relation(role),
		})
	}

	for _, groupID := range groupIDs {
		tuples = append(tuples, fgax.TupleKey{
			Subject: fgax.Entity{
				Kind:       fgax.Kind(generated.TypeGroup),
				Identifier: groupID,
				Relation:   fgax.Relation(fgax.MemberRelation),
			},
			Object:   object,
			Relation: fgax.Relation(role),
		})
	}

	return tuples
}
