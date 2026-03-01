package handlers

import (
	"context"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
)

// AccountAccessHandler list roles a subject has access to in relation an object
func (h *Handler) AccountRolesHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	return ProcessAuthenticatedRequest(ctx, h, openapi, models.ExampleAccountRolesRequest, models.ExampleAccountRolesReply,
		func(reqCtx context.Context, in *models.AccountRolesRequest, caller *auth.Caller) (*models.AccountRolesReply, error) {
			ids := in.ObjectIDs
			if len(ids) == 0 {
				ids = []string{in.ObjectID}
			}

			objectRoles := make(map[string][]string)

			for _, id := range ids {
				req := fgax.ListAccess{
					SubjectType: in.SubjectType,
					SubjectID:   caller.SubjectID,
					ObjectID:    id,
					ObjectType:  fgax.Kind(in.ObjectType),
					Relations:   in.Relations,
					Context:     utils.NewOrganizationContextKey(caller.SubjectEmail),
				}

				roles, err := h.DBClient.Authz.ListRelations(reqCtx, req)
				if err != nil {
					logx.FromContext(reqCtx).Error().Err(err).Interface("access_request", req).Msg("error checking access")
					return nil, ErrInvalidInput
				}

				objectRoles[id] = roles
			}

			resp := &models.AccountRolesReply{
				Reply:       rout.Reply{Success: true},
				ObjectRoles: objectRoles,
			}

			// kept for backward compatibility if only one object ID is requested
			if len(ids) == 1 {
				resp.Roles = objectRoles[ids[0]]
			}

			return resp, nil
		})
}
