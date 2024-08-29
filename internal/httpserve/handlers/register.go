package handlers

import (
	"context"
	"net/http"

	"github.com/cenkalti/backoff/v4"
	"github.com/getkin/kin-openapi/openapi3"
	ph "github.com/posthog/posthog-go"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/marionette"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/models"
)

const (
	maxEmailAttempts = 5
)

// RegisterHandler handles the registration of a new user, creating the user, personal organization
// and sending an email verification to the email address in the request
// the user will not be able to authenticate until the email is verified
// [MermaidChart: 5a357443-f959-4f16-a07f-ec504f67f0eb]
func (h *Handler) RegisterHandler(ctx echo.Context) error {
	var in models.RegisterRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}

	if err := in.Validate(); err != nil {
		return h.InvalidInput(ctx, err)
	}

	// create user
	input := generated.CreateUserInput{
		FirstName: &in.FirstName,
		LastName:  &in.LastName,
		Email:     in.Email,
		Password:  &in.Password,
	}

	// set viewer context
	ctxWithToken := token.NewContextWithSignUpToken(ctx.Request().Context(), in.Email)

	meowuser, err := h.createUser(ctxWithToken, input)
	if err != nil {
		h.Logger.Errorw("error creating new user", "error", err)

		if IsUniqueConstraintError(err) {
			return h.Conflict(ctx, "user already exists", UserExistsErrCode)
		}

		if generated.IsValidationError(err) {
			return h.InvalidInput(ctx, invalidInputError(err))
		}

		return err
	}

	// setup user context
	userCtx := auth.AddAuthenticatedUserContext(ctx, &auth.AuthenticatedUser{
		SubjectID: meowuser.ID,
	})

	// create email verification token
	user := &User{
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Email:     in.Email,
		ID:        meowuser.ID,
	}

	meowtoken, err := h.storeAndSendEmailVerificationToken(userCtx, user)
	if err != nil {
		h.Logger.Errorw("error storing token", "error", err)

		return h.InternalServerError(ctx, err)
	}

	out := &models.RegisterReply{
		Reply:   rout.Reply{Success: true},
		ID:      meowuser.ID,
		Email:   meowuser.Email,
		Message: "Welcome to Openlane!",
		Token:   meowtoken.Token,
	}

	return h.Created(ctx, out)
}

func (h *Handler) storeAndSendEmailVerificationToken(ctx context.Context, user *User) (*generated.EmailVerificationToken, error) {
	// expire all existing tokens
	if err := h.expireAllVerificationTokensUserByEmail(ctx, user.Email); err != nil {
		h.Logger.Errorw("error expiring existing tokens", "error", err)

		return nil, err
	}

	// check if the user has attempted to verify their email too many times
	attempts, err := h.countVerificationTokensUserByEmail(ctx, user.Email)
	if err != nil {
		h.Logger.Errorw("error getting existing tokens", "error", err)

		return nil, err
	}

	if attempts >= maxEmailAttempts {
		return nil, ErrMaxAttempts
	}

	// create a new token and store it in the database
	if err := user.CreateVerificationToken(); err != nil {
		h.Logger.Errorw("unable to create verification token", "error", err)

		return nil, err
	}

	meowtoken, err := h.createEmailVerificationToken(ctx, user)
	if err != nil {
		return nil, err
	}

	props := ph.NewProperties().
		Set("user_id", user.ID).
		Set("email", user.Email).
		Set("first_name", user.FirstName).
		Set("last_name", user.LastName)

	h.AnalyticsClient.Event("email_verification_sent", props)

	// send emails via TaskMan as to not create blocking operations in the server
	if err := h.TaskMan.Queue(marionette.TaskFunc(func(ctx context.Context) error {
		return h.SendVerificationEmail(user)
	}), marionette.WithRetries(3), //nolint:mnd
		marionette.WithBackoff(backoff.NewExponentialBackOff()),
		marionette.WithErrorf("could not send verification email to user %s", user.ID),
	); err != nil {
		return nil, err
	}

	return meowtoken, nil
}

// BindRegisterHandler is used to bind the register endpoint to the OpenAPI schema
func (h *Handler) BindRegisterHandler() *openapi3.Operation {
	register := openapi3.NewOperation()
	register.Description = "Register creates a new user in the database with the specified password, allowing the user to login to Openlane. This endpoint requires a 'strong' password and a valid register request, otherwise a 400 reply is returned. The password is stored in the database as an argon2 derived key so it is impossible for a hacker to get access to raw passwords. A personal organization is created for the user registering based on the organization data in the register request and the user is assigned the Owner role"
	register.OperationID = "RegisterHandler"
	register.Security = &openapi3.SecurityRequirements{}

	h.AddRequestBody("RegisterRequest", models.ExampleRegisterSuccessRequest, register)
	h.AddResponse("RegisterReply", "success", models.ExampleRegisterSuccessResponse, register, http.StatusCreated)
	register.AddResponse(http.StatusInternalServerError, internalServerError())
	register.AddResponse(http.StatusBadRequest, badRequest())
	register.AddResponse(http.StatusConflict, conflict())

	return register
}
