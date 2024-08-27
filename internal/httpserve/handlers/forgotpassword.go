package handlers

import (
	"context"
	"net/http"

	"github.com/cenkalti/backoff/v4"
	"github.com/getkin/kin-openapi/openapi3"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/marionette"
	"github.com/theopenlane/utils/rout"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/auth"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

// ForgotPassword will send an forgot password email if the provided email exists
func (h *Handler) ForgotPassword(ctx echo.Context) error {
	var in models.ForgotPasswordRequest
	if err := ctx.Bind(&in); err != nil {
		return h.BadRequest(ctx, err)
	}

	if err := in.Validate(); err != nil {
		return h.InvalidInput(ctx, err)
	}

	out := &models.ForgotPasswordReply{
		Reply: rout.Reply{
			Success: true,
		},
		Message: "We've received your request to have the password associated with this email reset. Please check your email.",
	}

	entUser, err := h.getUserByEmail(ctx.Request().Context(), in.Email, enums.AuthProviderCredentials)
	if err != nil {
		if ent.IsNotFound(err) {
			// return a 200 response even if user is not found to avoid
			// exposing confidential information
			return h.Success(ctx, out)
		}

		h.Logger.Errorf("error retrieving user email", "error", err)

		return h.InternalServerError(ctx, err)
	}

	// create password reset email token
	user := &User{
		FirstName: entUser.FirstName,
		LastName:  entUser.LastName,
		Email:     entUser.Email,
		ID:        entUser.ID,
	}

	authCtx := auth.AddAuthenticatedUserContext(ctx, &auth.AuthenticatedUser{
		SubjectID: entUser.ID,
	})

	if _, err = h.storeAndSendPasswordResetToken(authCtx, user); err != nil {
		return h.InternalServerError(ctx, err)
	}

	return h.Success(ctx, out)
}

// storeAndSendPasswordResetToken creates a password reset token for the user and sends an email with the token
func (h *Handler) storeAndSendPasswordResetToken(ctx context.Context, user *User) (*ent.PasswordResetToken, error) {
	if err := h.expireAllResetTokensUserByEmail(ctx, user.Email); err != nil {
		h.Logger.Errorw("error expiring existing tokens", "error", err)

		return nil, err
	}

	if err := user.CreatePasswordResetToken(); err != nil {
		h.Logger.Errorw("unable to create password reset token", "error", err)
		return nil, err
	}

	meowtoken, err := h.createPasswordResetToken(ctx, user)
	if err != nil {
		return nil, err
	}

	// send emails via TaskMan as to not create blocking operations in the server
	if err := h.TaskMan.Queue(marionette.TaskFunc(func(ctx context.Context) error {
		return h.SendPasswordResetRequestEmail(user)
	}), marionette.WithRetries(3), //nolint:mnd
		marionette.WithBackoff(backoff.NewExponentialBackOff()),
		marionette.WithErrorf("could not send password reset email to user %s", user.ID),
	); err != nil {
		return nil, err
	}

	return meowtoken, nil
}

// BindForgotPassword is used to bind the forgot password endpoint to the OpenAPI schema
func (h *Handler) BindForgotPassword() *openapi3.Operation {
	forgotPassword := openapi3.NewOperation()
	forgotPassword.Description = "ForgotPassword is a service for users to request a password reset email. The email address must be provided in the POST request and the user must exist in the database. This endpoint always returns 200 regardless of whether the user exists or not to avoid leaking information about users in the database"
	forgotPassword.OperationID = "ForgotPassword"
	forgotPassword.Security = &openapi3.SecurityRequirements{}

	h.AddRequestBody("ForgotPasswordRequest", models.ExampleForgotPasswordSuccessRequest, forgotPassword)
	h.AddResponse("ForgotPasswordReply", "success", models.ExampleForgotPasswordSuccessResponse, forgotPassword, http.StatusOK)
	forgotPassword.AddResponse(http.StatusInternalServerError, internalServerError())
	forgotPassword.AddResponse(http.StatusBadRequest, badRequest())

	return forgotPassword
}
