package handlers

import (
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/utils/rout"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/logx"
)

// UnsubscribeHandler is the handler for the unsubscribe endpoint. The token sent in update
// emails is an unguessable bearer token, so possession is sufficient proof to unsubscribe;
// no TTL is enforced so unsubscribe links remain valid indefinitely
func (h *Handler) UnsubscribeHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleUnsubscribeRequest, models.ExampleUnsubscribeResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	// setup viewer context so the lookup and update are authorized for the subscriber
	ctxWithToken := token.NewContextWithVerifyToken(reqCtx, in.Token)

	entSubscriber, err := h.getSubscriberByToken(ctxWithToken, in.Token)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.BadRequest(ctx, err, openapi)
		}

		logx.FromContext(reqCtx).Error().Err(err).Msg("error retrieving subscriber")

		return h.InternalServerError(ctx, ErrUnableToUnsubscribe, openapi)
	}

	// add org to the authenticated context so the update passes org-scoped checks
	reqCtx = auth.WithCaller(ctxWithToken, &auth.Caller{
		OrganizationID:  entSubscriber.OwnerID,
		OrganizationIDs: []string{entSubscriber.OwnerID},
	})

	ctxWithToken = token.NewContextWithVerifyToken(reqCtx, in.Token)

	if !entSubscriber.Unsubscribed {
		if err := h.setSubscriberUnsubscribed(ctxWithToken, entSubscriber.ID); err != nil {
			logx.FromContext(reqCtx).Error().Err(err).Msg("error unsubscribing subscriber")

			return h.InternalServerError(ctx, ErrUnableToUnsubscribe, openapi)
		}
	}

	out := &models.UnsubscribeReply{
		Reply:   rout.Reply{Success: true},
		Message: "You have been unsubscribed and will no longer receive updates.",
	}

	return h.Success(ctx, out, openapi)
}
