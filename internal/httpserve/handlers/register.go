package handlers

import (
	"context"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/jobs"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/emailtemplates"
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
		log.Error().Err(err).Msg("error creating new user")

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
		log.Error().Err(err).Msg("error storing email verification token")

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
		log.Error().Err(err).Msg("error expiring existing tokens")

		return nil, err
	}

	// check if the user has attempted to verify their email too many times
	attempts, err := h.countVerificationTokensUserByEmail(ctx, user.Email)
	if err != nil {
		log.Error().Err(err).Msg("error getting existing tokens")

		return nil, err
	}

	if attempts >= maxEmailAttempts {
		return nil, ErrMaxAttempts
	}

	// create a new token and store it in the database
	if err := user.CreateVerificationToken(); err != nil {
		log.Error().Err(err).Msg("error creating verification token")

		return nil, err
	}

	meowtoken, err := h.createEmailVerificationToken(ctx, user)
	if err != nil {
		return nil, err
	}

	email, err := h.Email.NewVerifyEmail(emailtemplates.Recipient{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, meowtoken.Token)
	if err != nil {
		log.Error().Err(err).Msg("error rendering email")

		return nil, err
	}

	// send the email
	_, err = h.JobQueue.Insert(ctx, jobs.EmailArgs{
		Message: *email,
	}, nil)
	if err != nil {
		log.Error().Err(err).Msg("error queueing email verification")

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
