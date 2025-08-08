package handlers

import (
	"context"

	"github.com/rs/zerolog"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/models"
)

// AccountAccessHandler checks if a subject has access to an object
func (h *Handler) AccountAccessHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	return ProcessAuthenticatedRequest(ctx, h, openapi, models.ExampleAccountAccessRequest, models.ExampleAccountAccessReply,
		func(reqCtx context.Context, in *models.AccountAccessRequest, subject *auth.AuthenticatedUser) (*models.AccountAccessReply, error) {
			req := fgax.AccessCheck{
				SubjectType: in.SubjectType,
				Relation:    in.Relation,
				ObjectID:    in.ObjectID,
				ObjectType:  fgax.Kind(in.ObjectType),
				SubjectID:   subject.SubjectID,
				Context:     utils.NewOrganizationContextKey(subject.SubjectEmail),
			}

			allow, err := h.DBClient.Authz.CheckAccess(reqCtx, req)
			if err != nil {
				zerolog.Ctx(reqCtx).Error().Err(err).Interface("access_request", req).Msg("error checking access")
				return nil, ErrInvalidInput
			}

			return &models.AccountAccessReply{
				Reply:   rout.Reply{Success: true},
				Allowed: allow,
			}, nil
		})
}
