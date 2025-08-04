package handlers

import (
	"errors"

	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/models"
)

// ResendEmail will resend an email verification email if the provided email exists
func (h *Handler) ResendEmail(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleResendEmailSuccessRequest, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	// set viewer context
	ctxWithToken := token.NewContextWithSignUpToken(ctx.Request().Context(), in.Email)

	out := &models.ResendReply{
		Reply:   rout.Reply{Success: true},
		Message: "We've received your request to be resent an email to complete verification. Please check your email.",
	}

	// email verifications only come to users that were created with username/password logins
	entUser, err := h.getUserByEmail(ctxWithToken, in.Email)
	if err != nil {
		if ent.IsNotFound(err) {
			// return a 200 response even if user is not found to avoid
			// exposing confidential information
			return h.Success(ctx, out, openapi)
		}

		log.Error().Err(err).Msg("error retrieving user email")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// check to see if user is already confirmed
	if entUser.Edges.Setting.EmailConfirmed {
		out.Message = "email is already confirmed"

		return h.Success(ctx, out, openapi)
	}

	// setup user context
	userCtx := setAuthenticatedContext(ctxWithToken, entUser)

	// create email verification token
	user := &User{
		FirstName: entUser.FirstName,
		LastName:  entUser.LastName,
		Email:     entUser.Email,
		ID:        entUser.ID,
	}

	if _, err = h.storeAndSendEmailVerificationToken(userCtx, user); err != nil {
		log.Error().Err(err).Msg("error storing email verification token")

		if errors.Is(err, ErrMaxAttempts) {
			return h.TooManyRequests(ctx, err)
		}

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	return h.Success(ctx, out, openapi)
}
