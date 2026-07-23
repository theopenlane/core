package handlers

import (
	"context"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/pkg/logx"
)

// AccountAccessHandler checks if a subject has access to an object
func (h *Handler) AccountAccessHandler(ctx echo.Context) error {
	return ProcessAuthenticatedRequest(ctx, h,
		func(reqCtx context.Context, in *models.AccountAccessRequest, caller *auth.Caller) (*models.AccountAccessResponse, error) {
			req := fgax.AccessCheck{
				SubjectType: in.SubjectType,
				Relation:    in.Relation,
				ObjectID:    in.ObjectID,
				ObjectType:  fgax.Kind(in.ObjectType),
				SubjectID:   caller.SubjectID,
			}

			allow, err := h.DBClient.Authz.CheckAccess(reqCtx, req)
			if err != nil {
				logx.FromContext(reqCtx).Error().Err(err).Interface("access_request", req).Msg("error checking access")
				return nil, ErrInvalidInput
			}

			return &models.AccountAccessResponse{
				Reply:   rout.Reply{Success: true},
				Allowed: allow,
			}, nil
		})
}
