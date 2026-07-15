package handlers

import (
	"context"
	"errors"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/pkg/logx"
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

	// scope the caller to the subscriber's owning org so the update passes the org-ownership pre-policy
	// (DenyIfNotInOrganization); the verify token set above is preserved and authorizes the mutation
	ctxWithToken = auth.WithCaller(ctxWithToken, &auth.Caller{
		OrganizationID:  entSubscriber.OwnerID,
		OrganizationIDs: []string{entSubscriber.OwnerID},
	})

	// single-use: token stays on the row for the unsubscribe link, so gate on the verified flag
	if entSubscriber.VerifiedEmail {
		logx.FromContext(reqCtx).Error().Err(ErrSubscriptionTokenAlreadyUsed).Msg("subscription verify token replayed")

		return h.BadRequest(ctx, ErrSubscriptionTokenAlreadyUsed, openapi)
	}

	// never resurrect an unsubscribed contact; re-subscription goes through createSubscriber
	if entSubscriber.Unsubscribed {
		logx.FromContext(reqCtx).Warn().Str("subscriber_id", entSubscriber.ID).Msg("verify link replayed for unsubscribed contact")

		return h.BadRequest(ctx, ErrSubscriberUnsubscribed, openapi)
	}

	if err := h.verifySubscriberToken(ctxWithToken, entSubscriber); err != nil {
		switch {
		case errors.Is(err, ErrExpiredToken):
			// a fresh link was already emailed; confirmation is still pending on that link
			return h.Success(ctx, &models.VerifySubscribeReply{
				Reply:   rout.Reply{Success: true},
				Message: "The verification link has expired, a new one has been sent - check your inbox to confirm.",
			}, openapi)
		case errors.Is(err, ErrMaxAttempts):
			logx.FromContext(reqCtx).Error().Err(err).Str("subscriber_id", entSubscriber.ID).Msg("subscriber exceeded verification attempts")

			return h.BadRequest(ctx, ErrMaxAttempts, openapi)
		default:
			logx.FromContext(reqCtx).Error().Err(err).Msg("error verifying subscriber token")

			return h.InternalServerError(ctx, ErrUnableToVerifyEmail, openapi)
		}
	}

	if err := h.updateSubscriberVerifiedEmail(ctxWithToken, entSubscriber.ID, generated.UpdateSubscriberInput{Email: &entSubscriber.Email}); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error updating subscriber")

		return h.InternalServerError(ctx, ErrUnableToVerifyEmail, openapi)
	}

	// the confirmation UX lives on the trust center's own domain (the page that called this endpoint), so
	// reply inline; the caller lands the subscriber on the trust center
	out := &models.VerifySubscribeReply{
		Reply:   rout.Reply{Success: true},
		Message: "Subscription confirmed, looking forward to sending you updates!",
	}

	return h.Success(ctx, out, openapi)
}

// verifySubscriberToken checks the token provided by the user and verifies it against the database
func (h *Handler) verifySubscriberToken(ctx context.Context, entSubscriber *generated.Subscriber) error {
	if entSubscriber.TTL == nil || entSubscriber.Secret == nil {
		logx.FromContext(ctx).Error().Msg("subscriber token missing required fields")

		return ErrUnableToVerifyEmail
	}

	// setup token to be validated
	t := &tokens.VerificationToken{
		Email: entSubscriber.Email,
	}
	t.ExpiresAt = *entSubscriber.TTL

	// verify token is valid, otherwise reset and send new token
	if err := t.Verify(entSubscriber.Token, *entSubscriber.Secret); err != nil {
		// a non-expired verification failure (tampered/mismatched secret) is not recoverable by a resend;
		// reject it rather than masquerading as an expired link that "sent a new email"
		if !errors.Is(err, tokens.ErrTokenExpired) {
			logx.FromContext(ctx).Error().Err(err).Msg("subscriber verification token failed validation")

			return ErrUnableToVerifyEmail
		}

		// cap the auto-resend at the same attempt budget the subscribe flow enforces so repeated
		// expired-link hits cannot send unbounded verification emails
		if entSubscriber.SendAttempts >= maxEmailAttempts {
			return ErrMaxAttempts
		}

		// token is expired, create a new token and send email
		verify, err := tokens.NewVerificationToken(entSubscriber.Email)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error creating verification token")

			return err
		}

		tokenValue, secret, err := verify.Sign()
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error signing verification token")

			return err
		}

		// update token settings in the database
		if err := h.updateSubscriberVerificationToken(ctx, entSubscriber.ID, tokenValue, verify.ExpiresAt, secret); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error updating subscriber verification token")

			return err
		}

		// set viewer context
		ctxWithToken := token.NewContextWithSignUpToken(ctx, entSubscriber.Email)

		// resend email with new token to the subscriber, naming the trust center they subscribed to
		orgName, err := h.subscriberNotificationOrgName(ctxWithToken, entSubscriber)
		if err != nil {
			return err
		}

		verifyURL, unsubscribeURL, branding := h.subscriberTrustCenterLinks(ctxWithToken, entSubscriber, tokenValue)

		if err := h.sendEmail(ctxWithToken, email.SubscribeOp.Name(), email.SubscribeRequest{
			RecipientInfo:       email.RecipientInfo{Email: entSubscriber.Email},
			TrustCenterBranding: branding,
			OrgName:             orgName,
			Token:               tokenValue,
			VerifyURL:           verifyURL,
			UnsubscribeURL:      unsubscribeURL,
		}); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error sending subscriber email")

			return err
		}

		return ErrExpiredToken
	}

	return nil
}
