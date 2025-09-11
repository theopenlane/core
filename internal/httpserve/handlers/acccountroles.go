package handlers

import (
	"context"

	"github.com/rs/zerolog"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/privacy/utils"
	models "github.com/theopenlane/core/pkg/openapi"
)

// AccountAccessHandler list roles a subject has access to in relation an object
func (h *Handler) AccountRolesHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	return ProcessAuthenticatedRequest(ctx, h, openapi, models.ExampleAccountRolesRequest, models.ExampleAccountRolesReply,
		func(reqCtx context.Context, in *models.AccountRolesRequest, au *auth.AuthenticatedUser) (*models.AccountRolesReply, error) {
			req := fgax.ListAccess{
				SubjectType: in.SubjectType,
				SubjectID:   au.SubjectID,
				ObjectID:    in.ObjectID,
				ObjectType:  fgax.Kind(in.ObjectType),
				Relations:   in.Relations,
				Context:     utils.NewOrganizationContextKey(au.SubjectEmail),
			}

			roles, err := h.DBClient.Authz.ListRelations(reqCtx, req)
			if err != nil {
				zerolog.Ctx(reqCtx).Error().Err(err).Interface("access_request", req).Msg("error checking access")
				return nil, ErrInvalidInput
			}

			return &models.AccountRolesReply{
				Reply: rout.Reply{Success: true},
				Roles: roles,
			}, nil
		})
}
