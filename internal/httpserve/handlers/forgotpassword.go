package handlers

import (
	"context"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	models "github.com/theopenlane/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/logx"
)

// ForgotPassword will send an forgot password email if the provided email exists
func (h *Handler) ForgotPassword(ctx echo.Context, openapi *OpenAPIContext) error {
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleForgotPasswordSuccessRequest, models.ExampleForgotPasswordSuccessResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	out := &models.ForgotPasswordReply{
		Reply: rout.Reply{
			Success: true,
		},
		Message: "We've received your request to have the password associated with this email reset. Please check your email.",
	}

	reqCtx := ctx.Request().Context()

	entUser, err := h.getUserByEmail(reqCtx, req.Email)
	if err != nil {
		if ent.IsNotFound(err) {
			// return a 200 response even if user is not found to avoid
			// exposing confidential information
			return h.Success(ctx, out, openapi)
		}

		logx.FromContext(reqCtx).Error().Err(err).Msg("error retrieving user email")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// create password reset email token
	user := &User{
		FirstName: entUser.FirstName,
		LastName:  entUser.LastName,
		Email:     entUser.Email,
		ID:        entUser.ID,
	}

	authCtx := setAuthenticatedContext(reqCtx, entUser)

	if _, err = h.storeAndSendPasswordResetToken(authCtx, user); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error storing and sending password reset token")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	return h.Success(ctx, out, openapi)
}

// storeAndSendPasswordResetToken creates a password reset token for the user and sends an email with the token
func (h *Handler) storeAndSendPasswordResetToken(ctx context.Context, user *User) (*ent.PasswordResetToken, error) {
	if err := h.expireAllResetTokensUserByEmail(ctx, user.Email); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error expiring existing tokens")

		return nil, err
	}

	if err := user.CreatePasswordResetToken(); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating password reset token")
		return nil, err
	}

	meowtoken, err := h.createPasswordResetToken(ctx, user)
	if err != nil {
		return nil, err
	}

	// add email send to the job queue
	if err := h.sendPasswordResetRequestEmail(ctx, user); err != nil {
		return nil, err
	}

	return meowtoken, nil
}
