package handlers

import (
	"errors"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

// ResendEmail will resend an email verification email if the provided email exists
func (h *Handler) ResendEmail(ctx echo.Context) error {
	var in models.ResendRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}

	if err := in.Validate(); err != nil {
		return h.InvalidInput(ctx, err)
	}

	// set viewer context
	ctxWithToken := token.NewContextWithSignUpToken(ctx.Request().Context(), in.Email)

	out := &models.ResendReply{
		Reply:   rout.Reply{Success: true},
		Message: "We've received your request to be resent an email to complete verification. Please check your email.",
	}

	// email verifications only come to users that were created with username/password logins
	entUser, err := h.getUserByEmail(ctxWithToken, in.Email, enums.AuthProviderCredentials)
	if err != nil {
		if ent.IsNotFound(err) {
			// return a 200 response even if user is not found to avoid
			// exposing confidential information
			return h.Success(ctx, out)
		}

		log.Error().Err(err).Msg("error retrieving user email")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	// check to see if user is already confirmed
	if entUser.Edges.Setting.EmailConfirmed {
		out.Message = "email is already confirmed"

		return h.Success(ctx, out)
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

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	return h.Success(ctx, out)
}

// BindResendEmailHandler binds the resend email verification endpoint to the OpenAPI schema
func (h *Handler) BindResendEmailHandler() *openapi3.Operation {
	resendEmail := openapi3.NewOperation()
	resendEmail.Description = "ResendEmail accepts an email address via a POST request and always returns a 200 Status OK response, no matter the input or result of the processing. This is to ensure that no secure information is leaked from this unauthenticated endpoint. If the email address belongs to a user who has not been verified, another verification email is sent. If the post request contains an orgID and the user is invited to that organization but hasn't accepted the invite, then the invite is resent."
	resendEmail.Tags = []string{"accountRegistration"}
	resendEmail.OperationID = "ResendEmail"
	resendEmail.Security = &openapi3.SecurityRequirements{}

	h.AddRequestBody("ResendEmailRequest", models.ExampleResendEmailSuccessRequest, resendEmail)
	h.AddResponse("ResendEmailReply", "success", models.ExampleResendEmailSuccessResponse, resendEmail, http.StatusOK)
	resendEmail.AddResponse(http.StatusInternalServerError, internalServerError())
	resendEmail.AddResponse(http.StatusBadRequest, badRequest())
	resendEmail.AddResponse(http.StatusBadRequest, invalidInput())
	resendEmail.AddResponse(http.StatusTooManyRequests, tooManyRequests())

	return resendEmail
}
