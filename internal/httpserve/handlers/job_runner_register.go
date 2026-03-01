package handlers

import (
	"time"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/logx"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"
)

func (h *Handler) RegisterJobRunner(ctx echo.Context, openapi *OpenAPIContext) error {
	r, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleJobRunnerRegistrationRequest, models.ExampleJobRunnerRegistrationResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	reqCtx := ctx.Request().Context()

	ctxWithToken := token.NewContextWithJobRunnerRegistrationToken(reqCtx, r.Token)

	registrationToken, err := h.getOrgByJobRunnerVerificationToken(ctxWithToken, r.Token)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.Unauthorized(ctx, err)
		}

		logx.FromContext(reqCtx).Error().Err(err).Msg("error retrieving job runner registration token")

		return h.InternalServerError(ctx, ErrUnableToRegisterJobRunner, openapi)
	}

	if registrationToken.ExpiresAt.Before(time.Now()) {
		return h.Unauthorized(ctx, ErrJobRunnerRegistrationTokenExpired)
	}

	ctxWithToken = auth.WithCaller(ctxWithToken, &auth.Caller{
		SubjectID:       registrationToken.ID,
		OrganizationID:  registrationToken.OwnerID,
		OrganizationIDs: []string{registrationToken.OwnerID},
	})

	if err := h.createJobRunner(ctxWithToken, registrationToken, *r); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("could not create a new runner with your token")

		if generated.IsConstraintError(err) {
			return h.BadRequest(ctx, ErrJobRunnerAlreadyRegistered, openapi)
		}

		return h.InternalServerError(ctx, ErrUnableToRegisterJobRunner, openapi)
	}

	out := &models.JobRunnerRegistrationReply{
		Reply: rout.Reply{
			Success: true,
		},
		Message: "Job runner node registered",
	}

	return h.Created(ctx, out)
}
