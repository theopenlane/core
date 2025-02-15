package handlers

import (
	"errors"
	"net/http"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/models"
)

// VerifyEmail is the handler for the email verification endpoint
func (h *Handler) VerifyEmail(ctx echo.Context) error {
	var in models.VerifyRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}

	if err := in.Validate(); err != nil {
		return h.InvalidInput(ctx, err)
	}

	// setup viewer context
	ctxWithToken := token.NewContextWithVerifyToken(ctx.Request().Context(), in.Token)

	entUser, err := h.getUserByEVToken(ctxWithToken, in.Token)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.BadRequest(ctx, err)
		}

		log.Error().Err(err).Msg("error retrieving user token")

		return h.InternalServerError(ctx, ErrUnableToVerifyEmail)
	}

	// create email verification
	user := &User{
		ID:    entUser.ID,
		Email: entUser.Email,
	}

	userCtx := auth.AddAuthenticatedUserContext(ctx, &auth.AuthenticatedUser{
		SubjectID: entUser.ID,
	})

	// check to see if user is already confirmed
	if !entUser.Edges.Setting.EmailConfirmed {
		// set tokens for request
		if err := user.setUserTokens(entUser, in.Token); err != nil {
			log.Error().Err(err).Msg("unable to set user tokens for request")

			return h.BadRequest(ctx, err)
		}

		// Construct the user token from the database fields
		t := &tokens.VerificationToken{
			Email: entUser.Email,
		}

		if t.ExpiresAt, err = user.GetVerificationExpires(); err != nil {
			log.Error().Err(err).Msg("unable to parse expiration")

			return h.InternalServerError(ctx, ErrUnableToVerifyEmail)
		}

		// Verify the token with the stored secret
		if err = t.Verify(user.GetVerificationToken(), user.EmailVerificationSecret); err != nil {
			if errors.Is(err, tokens.ErrTokenExpired) {
				userCtx = token.NewContextWithSignUpToken(userCtx, user.Email)

				meowtoken, err := h.storeAndSendEmailVerificationToken(userCtx, user)
				if err != nil {
					log.Error().Err(err).Msg("unable to resend verification token")

					return h.InternalServerError(ctx, ErrUnableToVerifyEmail)
				}

				out := &models.VerifyReply{
					Reply:   rout.Reply{Success: false},
					ID:      meowtoken.ID,
					Email:   user.Email,
					Message: "Token expired, a new token has been issued. Please check your email and try again.",
				}

				return h.Created(ctx, out)
			}

			return h.BadRequest(ctx, err)
		}

		if err := h.setEmailConfirmed(userCtx, entUser); err != nil {
			return h.BadRequest(ctx, err)
		}
	}

	if err := h.addDefaultOrgToUserQuery(userCtx, entUser); err != nil {
		return h.InternalServerError(ctx, err)
	}

	if err := h.validateAllowedDomains(ctxWithToken, entUser); err != nil {
		return h.BadRequest(ctx, err)
	}

	// create new claims for the user
	auth, err := h.AuthManager.GenerateUserAuthSession(ctx, entUser)
	if err != nil {
		log.Error().Err(err).Msg("unable to create new auth session")

		return h.InternalServerError(ctx, err)
	}

	out := &models.VerifyReply{
		ID:       entUser.ID,
		Email:    entUser.Email,
		Reply:    rout.Reply{Success: true},
		Message:  "success",
		AuthData: *auth,
	}

	return h.Success(ctx, out)
}

// setUserTokens sets the fields to verify the email
func (u *User) setUserTokens(user *generated.User, reqToken string) error {
	tokens := user.Edges.EmailVerificationTokens
	for _, t := range tokens {
		if t.Token == reqToken {
			u.EmailVerificationToken = sql.NullString{String: t.Token, Valid: true}
			u.EmailVerificationSecret = *t.Secret
			u.EmailVerificationExpires = sql.NullString{String: t.TTL.Format(time.RFC3339Nano), Valid: true}

			return nil
		}
	}

	return ErrNotFound
}

// BindVerifyEmailHandler binds the verify email verification endpoint to the OpenAPI schema
func (h *Handler) BindVerifyEmailHandler() *openapi3.Operation {
	verify := openapi3.NewOperation()
	verify.Description = "VerifyEmail verifies a user's email address by validating the token in the request and setting the user's validated field in the database to true. This endpoint is intended to be called by frontend applications after the user has followed the link in the verification email"
	verify.Tags = []string{"accountRegistration"}
	verify.OperationID = "VerifyEmail"
	verify.Security = &openapi3.SecurityRequirements{}

	h.AddQueryParameter("VerifyRequest", "token", models.ExampleVerifySuccessRequest, verify)
	h.AddResponse("VerifyReply", "success", models.ExampleVerifySuccessResponse, verify, http.StatusOK)
	verify.AddResponse(http.StatusInternalServerError, internalServerError())
	verify.AddResponse(http.StatusBadRequest, badRequest())
	verify.AddResponse(http.StatusCreated, created())
	verify.AddResponse(http.StatusBadRequest, invalidInput())

	return verify
}
