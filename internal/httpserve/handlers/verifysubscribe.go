package handlers

import (
	"context"
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

// VerifySubscriptionHandler is the handler for the subscription verification endpoint
func (h *Handler) VerifySubscriptionHandler(ctx echo.Context) error {
	var in models.VerifySubscribeRequest
	if err := ctx.Bind(&in); err != nil {
		return h.BadRequest(ctx, err)
	}

	if err := in.Validate(); err != nil {
		return h.InvalidInput(ctx, err)
	}

	// setup viewer context
	ctxWithToken := token.NewContextWithVerifyToken(ctx.Request().Context(), in.Token)

	entSubscriber, err := h.getSubscriberByToken(ctxWithToken, in.Token)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.BadRequest(ctx, err)
		}

		log.Error().Err(err).Msg("error retrieving subscriber")

		return h.InternalServerError(ctx, ErrUnableToVerifyEmail)
	}

	// add org to the authenticated context
	reqCtx := auth.AddAuthenticatedUserContext(ctx, &auth.AuthenticatedUser{
		OrganizationID:  entSubscriber.OwnerID,
		OrganizationIDs: []string{entSubscriber.OwnerID},
	})

	ctxWithToken = token.NewContextWithVerifyToken(reqCtx, in.Token)

	if !entSubscriber.VerifiedEmail {
		if err := h.verifySubscriberToken(ctxWithToken, entSubscriber); err != nil {
			if errors.Is(err, ErrExpiredToken) {
				out := &models.VerifySubscribeReply{
					Reply:   rout.Reply{Success: false},
					Message: "The verification link has expired, a new one has been sent to your email.",
				}

				return h.Created(ctx, out)
			}

			log.Error().Err(err).Msg("error verifying subscriber token")

			return h.InternalServerError(ctx, ErrUnableToVerifyEmail)
		}

		input := generated.UpdateSubscriberInput{
			Email: &entSubscriber.Email,
		}

		if err := h.updateSubscriberVerifiedEmail(ctxWithToken, entSubscriber.ID, input); err != nil {
			log.Error().Err(err).Msg("error updating subscriber")

			return h.InternalServerError(ctx, ErrUnableToVerifyEmail)
		}
	}

	out := &models.VerifySubscribeReply{
		Reply:   rout.Reply{Success: true},
		Message: "Subscription confirmed, looking forward to sending you updates!",
	}

	return h.Success(ctx, out)
}

// verifySubscriberToken checks the token provided by the user and verifies it against the database
func (h *Handler) verifySubscriberToken(ctx context.Context, entSubscriber *generated.Subscriber) error {
	// create User struct from entSubscriber
	user := &User{
		ID:                       entSubscriber.ID,
		Email:                    entSubscriber.Email,
		EmailVerificationSecret:  *entSubscriber.Secret,
		EmailVerificationToken:   sql.NullString{String: entSubscriber.Token, Valid: true},
		EmailVerificationExpires: sql.NullString{String: entSubscriber.TTL.Format(time.RFC3339Nano), Valid: true},
	}

	// setup token to be validated
	t := &tokens.VerificationToken{
		Email: entSubscriber.Email,
	}

	var err error
	t.ExpiresAt, err = user.GetVerificationExpires()

	if err != nil {
		log.Error().Err(err).Msg("unable to parse expiration")

		return ErrUnableToVerifyEmail
	}

	// verify token is valid, otherwise reset and send new token
	if err := t.Verify(user.GetVerificationToken(), user.EmailVerificationSecret); err != nil {
		// if token is expired, create new token and send email
		if errors.Is(err, tokens.ErrTokenExpired) {
			if err := user.CreateVerificationToken(); err != nil {
				log.Error().Err(err).Msg("error creating verification token")

				return err
			}

			// update token settings in the database
			if err := h.updateSubscriberVerificationToken(ctx, user); err != nil {
				log.Error().Err(err).Msg("error updating subscriber verification token")

				return err
			}

			// set viewer context
			ctxWithToken := token.NewContextWithSignUpToken(ctx, entSubscriber.Email)

			// resend email with new token to the subscriber
			if err := h.sendSubscriberEmail(ctxWithToken, user, entSubscriber.OwnerID); err != nil {
				log.Error().Err(err).Msg("error sending subscriber email")

				return err
			}
		}

		return ErrExpiredToken
	}

	return nil
}

// BindVerifySubscriberHandler creates the openapi operation for the subscription verification endpoint
func (h *Handler) BindVerifySubscriberHandler() *openapi3.Operation {
	verify := openapi3.NewOperation()
	verify.Description = "Verify an email address for a subscription"
	verify.Tags = []string{"subscribe"}
	verify.OperationID = "VerifySubscriberEmail"
	verify.Security = &openapi3.SecurityRequirements{}

	h.AddQueryParameter("VerifySubscriptionRequest", "token", models.ExampleVerifySubscriptionSuccessRequest, verify)
	h.AddResponse("VerifySubscriptionReply", "success", models.ExampleVerifySubscriptionResponse, verify, http.StatusOK)
	verify.AddResponse(http.StatusInternalServerError, internalServerError())
	verify.AddResponse(http.StatusBadRequest, badRequest())
	verify.AddResponse(http.StatusBadRequest, invalidInput())
	verify.AddResponse(http.StatusCreated, created())

	return verify
}
