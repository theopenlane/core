package handlers

import (
	"errors"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/tokens"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/logx"
)

// VerifyEmail is the handler for the email verification endpoint
func (h *Handler) VerifyEmail(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleVerifySuccessRequest, models.ExampleVerifySuccessResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	// setup viewer context
	ctxWithToken := token.NewContextWithVerifyToken(reqCtx, in.Token)

	entUser, verificationTokenRecord, err := h.getUserByEVToken(ctxWithToken, in.Token)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.BadRequest(ctx, err, openapi)
		}

		logx.FromContext(reqCtx).Error().Err(err).Msg("error retrieving user token")

		return h.InternalServerError(ctx, ErrUnableToVerifyEmail, openapi)
	}

	userCtx := setAuthenticatedContext(ctxWithToken, entUser)

	// check to see if user is already confirmed
	if !entUser.Edges.Setting.EmailConfirmed {
		if verificationTokenRecord.TTL == nil || verificationTokenRecord.Secret == nil {
			logx.FromContext(reqCtx).Error().Msg("verification token missing required fields")

			return h.InternalServerError(ctx, ErrUnableToVerifyEmail, openapi)
		}

		// Construct the user token from the database fields
		t := &tokens.VerificationToken{
			Email: entUser.Email,
		}
		t.ExpiresAt = *verificationTokenRecord.TTL

		// Verify the token with the stored secret
		if err = t.Verify(verificationTokenRecord.Token, *verificationTokenRecord.Secret); err != nil {
			if errors.Is(err, tokens.ErrTokenExpired) {
				userCtx = token.NewContextWithSignUpToken(userCtx, entUser.Email)

				meowtoken, err := h.storeAndSendEmailVerificationToken(userCtx, &User{
					ID:        entUser.ID,
					Email:     entUser.Email,
					FirstName: entUser.FirstName,
					LastName:  entUser.LastName,
				})
				if err != nil {
					logx.FromContext(reqCtx).Error().Err(err).Msg("unable to resend verification token")

					return h.InternalServerError(ctx, ErrUnableToVerifyEmail, openapi)
				}

				out := &models.VerifyReply{
					Reply:   rout.Reply{Success: false},
					ID:      meowtoken.ID,
					Email:   entUser.Email,
					Message: "Token expired, a new token has been issued. Please check your email and try again.",
				}

				return h.Created(ctx, out)
			}

			return h.BadRequest(ctx, err, openapi)
		}

		if err := h.setEmailConfirmed(userCtx, entUser); err != nil {
			return h.BadRequest(ctx, err, openapi)
		}
	}

	// create new claims for the user
	auth, err := h.AuthManager.GenerateUserAuthSession(userCtx, ctx.Response().Writer, entUser)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to create new auth session")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	out := &models.VerifyReply{
		ID:       entUser.ID,
		Email:    entUser.Email,
		Reply:    rout.Reply{Success: true},
		Message:  "success",
		AuthData: *auth,
	}

	return h.Success(ctx, out, openapi)
}
