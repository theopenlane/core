package handlers

import (
	"context"
	"errors"
	"strings"
	"time"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/privacy/token"
	entval "github.com/theopenlane/ent/validator"

	"github.com/theopenlane/shared/enums"
	"github.com/theopenlane/shared/logx"
	models "github.com/theopenlane/shared/openapi"
)

const (
	maxEmailAttempts = 5
)

// RegisterHandler handles the registration of a new user, creating the user, personal organization
// and sending an email verification to the email address in the request
// the user will not be able to authenticate until the email is verified
func (h *Handler) RegisterHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleRegisterSuccessRequest, models.ExampleRegisterSuccessResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	// create user
	input := generated.CreateUserInput{
		FirstName:         &req.FirstName,
		LastName:          &req.LastName,
		Email:             req.Email,
		Password:          &req.Password,
		LastLoginProvider: &enums.AuthProviderCredentials,
	}

	reqCtx := ctx.Request().Context()

	// set viewer context
	ctxWithToken := token.NewContextWithSignUpToken(reqCtx, req.Email)

	if req.Token != nil {
		ctxWithToken = token.NewContextWithOrgInviteToken(ctxWithToken, *req.Token)

		invitedUser, err := h.getUserByInviteToken(ctxWithToken, *req.Token)
		if err != nil {
			logx.FromContext(reqCtx).Error().Err(err).Msg("error retrieving invite token")

			return h.BadRequest(ctx, err, openapi)
		}

		if !strings.EqualFold(invitedUser.Recipient, input.Email) {
			return h.BadRequest(ctx, ErrUnableToVerifyEmail, openapi)
		}

		if invitedUser.Expires.Before(time.Now()) {
			return h.BadRequest(ctx, ErrExpiredToken, openapi)
		}
	}

	meowuser, err := h.createUser(ctxWithToken, input)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error creating new user")

		if IsUniqueConstraintError(err) {
			return h.Conflict(ctx, "user already exists", UserExistsErrCode, openapi)
		}

		if generated.IsValidationError(err) {
			return h.InvalidInput(ctx, invalidInputError(err), openapi)
		}

		if errors.Is(err, entval.ErrEmailNotAllowed) {
			logx.FromContext(reqCtx).Error().Err(err).Str("email", input.Email).Msg("email not allowed")

			return h.InvalidInput(ctx, err, openapi)
		}

		return err
	}

	// setup user context
	userCtx := setAuthenticatedContext(ctxWithToken, meowuser)

	if req.Token != nil {
		ctx.SetRequest(ctx.Request().WithContext(userCtx))

		_, _, _, err := h.processInvitation(ctx, *req.Token, meowuser.Email)
		if err != nil {
			logx.FromContext(reqCtx).Error().Err(err).Msg("error processing invitation")

			return h.BadRequest(ctx, ErrUnableToVerifyEmail, openapi)
		}

		if err := h.setEmailConfirmed(userCtx, meowuser); err != nil {
			logx.FromContext(reqCtx).Error().Err(err).Msg("unable to set email as verified")

			return h.BadRequest(ctx, ErrUnableToVerifyEmail, openapi)
		}
	}

	out := &models.RegisterReply{
		Reply:   rout.Reply{Success: true},
		ID:      meowuser.ID,
		Email:   meowuser.Email,
		Message: "Welcome to Openlane!",
	}

	if req.Token == nil {
		// create email verification token
		user := &User{
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
			ID:        meowuser.ID,
		}

		meowtoken, err := h.storeAndSendEmailVerificationToken(userCtx, user)
		if err != nil {
			logx.FromContext(reqCtx).Error().Err(err).Msg("error storing email verification token")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
		}

		// only return the token in development
		if h.IsDev {
			out.Token = meowtoken.Token
		}
	}

	return h.Created(ctx, out, openapi)
}

func (h *Handler) storeAndSendEmailVerificationToken(ctx context.Context, user *User) (*generated.EmailVerificationToken, error) {
	// expire all existing tokens
	if err := h.expireAllVerificationTokensUserByEmail(ctx, user.Email); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error expiring existing tokens")

		return nil, err
	}

	// check if the user has attempted to verify their email too many times
	attempts, err := h.countVerificationTokensUserByEmail(ctx, user.Email)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error getting existing tokens")

		return nil, err
	}

	if attempts >= maxEmailAttempts {
		return nil, ErrMaxAttempts
	}

	// create a new token and store it in the database
	if err := user.CreateVerificationToken(); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating verification token")

		return nil, err
	}

	meowtoken, err := h.createEmailVerificationToken(ctx, user)
	if err != nil {
		return nil, err
	}

	return meowtoken, h.sendVerificationEmail(ctx, user, meowtoken.Token)
}
