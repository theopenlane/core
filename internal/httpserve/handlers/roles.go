package handlers

import (
	"context"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	models "github.com/theopenlane/core/common/openapi"
	fgamodel "github.com/theopenlane/core/fga/model"
	"github.com/theopenlane/core/pkg/logx"
)

// RolesHandler lists available roles that can be assigned to users in addition to the base organization role
func (h *Handler) RolesHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	return ProcessAuthenticatedRequest(ctx, h, openapi, models.ExampleRolesRequest, models.ExampleRolesReply,
		func(reqCtx context.Context, _ *models.RolesRequest, _ *auth.Caller) (*models.RolesReply, error) {
			roles, err := fgamodel.RoleOptions()
			if err != nil {
				logx.FromContext(reqCtx).Error().Err(err).Msg("error retrieving api roles")
				return nil, ErrProcessingRequest
			}

			resp := &models.RolesReply{
				Reply: rout.Reply{Success: true},
				Roles: roles,
			}

			return resp, nil
		})
}
