package handlers

import (
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/models"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"
)

func (h *Handler) BindRegisterRunnerNode() *openapi3.Operation {
	registerJobRunner := openapi3.NewOperation()
	registerJobRunner.Description = "Register a new job runner node"
	registerJobRunner.Tags = []string{"job-runners", "agents"}
	registerJobRunner.OperationID = "RegisterRunnerNode"
	registerJobRunner.Security = &openapi3.SecurityRequirements{}

	h.AddRequestBody("JobRunnerRegistrationRequest", models.ExampleJobRunnerRegistrationRequest, registerJobRunner)
	h.AddResponse("JobRunnerRegistrationReply", "success", models.ExampleJobRunnerRegistrationResponse, registerJobRunner, http.StatusOK)
	registerJobRunner.AddResponse(http.StatusBadRequest, badRequest())
	registerJobRunner.AddResponse(http.StatusBadRequest, invalidInput())

	return registerJobRunner
}

func (h *Handler) RegisterJobRunner(ctx echo.Context) error {
	var r models.JobRunnerRegistrationRequest
	if err := ctx.Bind(&r); err != nil {
		return h.InvalidInput(ctx, err)
	}

	if err := r.Validate(); err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	ctxWithToken := token.NewContextWithJobRunnerRegistrationToken(reqCtx, r.Token)

	registrationToken, err := h.getOrgByJobRunnerVerificationToken(ctxWithToken, r.Token)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.Unauthorized(ctx, err)
		}

		log.Error().Err(err).Msg("error retrieving job runner registration token")

		return h.InternalServerError(ctx, ErrUnableToRegisterJobRunner)
	}

	if registrationToken.ExpiresAt.Before(time.Now()) {
		return h.Unauthorized(ctx, ErrJobRunnerRegistrationTokenExpired)
	}

	ctxWithToken = auth.WithAuthenticatedUser(ctxWithToken, &auth.AuthenticatedUser{
		SubjectID:       registrationToken.ID,
		OrganizationID:  registrationToken.OwnerID,
		OrganizationIDs: []string{registrationToken.OwnerID},
	})

	if err := h.createJobRunner(ctxWithToken, registrationToken, r); err != nil {
		log.Error().Err(err).Msg("could not create a new runner with your token")

		if generated.IsConstraintError(err) {
			return h.BadRequest(ctx, ErrJobRunnerAlreadyRegistered)
		}

		return h.InternalServerError(ctx, ErrUnableToRegisterJobRunner)
	}

	out := &models.JobRunnerRegistrationReply{
		Reply: rout.Reply{
			Success: true,
		},
		Message: "Job runner node registered",
	}

	return h.Created(ctx, out)
}
