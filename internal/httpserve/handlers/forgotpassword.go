package handlers

import (
	"context"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	ent "github.com/theopenlane/core/internal/ent/generated"
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

	reqCtx := ctx.Request().Context()

	entUser, err := h.getUserByEmail(reqCtx, in.Email)
	if err != nil {
		if ent.IsNotFound(err) {
			// return a 200 response even if user is not found to avoid
			// exposing confidential information
			return h.Success(ctx, out)
		}

		log.Error().Err(err).Msg("error retrieving user email")

		return h.InternalServerError(ctx, err)
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
		return h.InternalServerError(ctx, err)
	}

	return h.Success(ctx, out)
}

// storeAndSendPasswordResetToken creates a password reset token for the user and sends an email with the token
func (h *Handler) storeAndSendPasswordResetToken(ctx context.Context, user *User) (*ent.PasswordResetToken, error) {
	if err := h.expireAllResetTokensUserByEmail(ctx, user.Email); err != nil {
		log.Error().Err(err).Msg("error expiring existing tokens")

		return nil, err
	}

	if err := user.CreatePasswordResetToken(); err != nil {
		log.Error().Err(err).Msg("error creating password reset token")
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

// BindForgotPassword is used to bind the forgot password endpoint to the OpenAPI schema
func (h *Handler) BindForgotPassword() *openapi3.Operation {
	forgotPassword := openapi3.NewOperation()
	forgotPassword.Description = "ForgotPassword is a service for users to request a password reset email. The email address must be provided in the POST request and the user must exist in the database. This endpoint always returns 200 regardless of whether the user exists or not to avoid leaking information about users in the database"
	forgotPassword.Tags = []string{"forgotpassword"}
	forgotPassword.OperationID = "ForgotPassword"
	forgotPassword.Security = &openapi3.SecurityRequirements{}

	h.AddRequestBody("ForgotPasswordRequest", models.ExampleForgotPasswordSuccessRequest, forgotPassword)
	h.AddResponse("ForgotPasswordReply", "success", models.ExampleForgotPasswordSuccessResponse, forgotPassword, http.StatusOK)
	forgotPassword.AddResponse(http.StatusInternalServerError, internalServerError())
	forgotPassword.AddResponse(http.StatusBadRequest, badRequest())
	forgotPassword.AddResponse(http.StatusBadRequest, invalidInput())

	return forgotPassword
}
