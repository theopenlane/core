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

// ScopesHandler lists available scopes that can be used for api tokens
func (h *Handler) ScopesHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	return ProcessAuthenticatedRequest(ctx, h, openapi, models.ExampleScopesRequest, models.ExampleScopesReply,
		func(reqCtx context.Context, _ *models.ScopesRequest, _ *auth.Caller) (*models.ScopesReply, error) {
			scopes, err := fgamodel.ScopeOptions()
			if err != nil {
				logx.FromContext(reqCtx).Error().Err(err).Msg("error retrieving api scopes")
				return nil, ErrProcessingRequest
			}

			resp := &models.ScopesReply{
				Reply:  rout.Reply{Success: true},
				Scopes: scopes,
			}

			return resp, nil
		})
}
