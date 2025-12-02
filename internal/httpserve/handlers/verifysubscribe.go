package handlers

import (
	"context"
	"errors"
	"time"

	"entgo.io/ent/dialect/sql"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/pkg/logx"
	models "github.com/theopenlane/core/pkg/openapi"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/privacy/token"
)

// VerifySubscriptionHandler is the handler for the subscription verification endpoint
func (h *Handler) VerifySubscriptionHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleVerifySubscriptionSuccessRequest, models.ExampleVerifySubscriptionResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	// setup viewer context
	ctxWithToken := token.NewContextWithVerifyToken(reqCtx, in.Token)

	entSubscriber, err := h.getSubscriberByToken(ctxWithToken, in.Token)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.BadRequest(ctx, err, openapi)
		}

		logx.FromContext(reqCtx).Error().Err(err).Msg("error retrieving subscriber")

		return h.InternalServerError(ctx, ErrUnableToVerifyEmail, openapi)
	}

	// add org to the authenticated context
	reqCtx = auth.WithAuthenticatedUser(ctxWithToken, &auth.AuthenticatedUser{
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

			logx.FromContext(reqCtx).Error().Err(err).Msg("error verifying subscriber token")

			return h.InternalServerError(ctx, ErrUnableToVerifyEmail, openapi)
		}

		input := generated.UpdateSubscriberInput{
			Email: &entSubscriber.Email,
		}

		if err := h.updateSubscriberVerifiedEmail(ctxWithToken, entSubscriber.ID, input); err != nil {
			logx.FromContext(reqCtx).Error().Err(err).Msg("error updating subscriber")

			return h.InternalServerError(ctx, ErrUnableToVerifyEmail, openapi)
		}
	}

	out := &models.VerifySubscribeReply{
		Reply:   rout.Reply{Success: true},
		Message: "Subscription confirmed, looking forward to sending you updates!",
	}

	return h.Success(ctx, out, openapi)
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
		logx.FromContext(ctx).Error().Err(err).Msg("unable to parse expiration")

		return ErrUnableToVerifyEmail
	}

	// verify token is valid, otherwise reset and send new token
	if err := t.Verify(user.GetVerificationToken(), user.EmailVerificationSecret); err != nil {
		// if token is expired, create new token and send email
		if errors.Is(err, tokens.ErrTokenExpired) {
			if err := user.CreateVerificationToken(); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("error creating verification token")

				return err
			}

			// update token settings in the database
			if err := h.updateSubscriberVerificationToken(ctx, user); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("error updating subscriber verification token")

				return err
			}

			// set viewer context
			ctxWithToken := token.NewContextWithSignUpToken(ctx, entSubscriber.Email)

			// resend email with new token to the subscriber
			if err := h.sendSubscriberEmail(ctxWithToken, user, entSubscriber.OwnerID); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("error sending subscriber email")

				return err
			}
		}

		return ErrExpiredToken
	}

	return nil
}
