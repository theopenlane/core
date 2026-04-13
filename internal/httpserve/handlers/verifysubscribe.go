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
	"github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/gala"
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

	// add org to the authenticated context
	reqCtx = auth.WithCaller(ctxWithToken, &auth.Caller{
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
		// if token is expired, create new token and send email
		if errors.Is(err, tokens.ErrTokenExpired) {
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

			// resend email with new token to the subscriber
			org, err := h.getOrgByID(ctxWithToken, entSubscriber.OwnerID)
			if err != nil {
				return err
			}

			input := email.SubscribeRequest{
				RecipientInfo: email.RecipientInfo{Email: entSubscriber.Email},
				OrgName:       org.DisplayName,
				Token:         tokenValue,
			}

			if receipt := h.Gala.EmitWithHeaders(ctxWithToken, email.SubscribeOp().Topic(), input,
				gala.NewHeaders([]string{"email", "subscriber", "resend"}, input)); receipt.Err != nil {
				logx.FromContext(ctx).Error().Err(receipt.Err).Msg("error sending subscriber email")

				return receipt.Err
			}
		}

		return ErrExpiredToken
	}

	return nil
}
