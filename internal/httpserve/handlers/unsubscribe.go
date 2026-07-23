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

// UnsubscribeHandler unsubscribes a subscriber by their bearer token. The token is embedded in every
// campaign email's unsubscribe link, so it stays valid while subscribed; a replay is an idempotent no-op
func (h *Handler) UnsubscribeHandler(ctx echo.Context) error {
	in, err := BindAndValidate[models.UnsubscribeRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	// setup viewer context so the lookup and update are authorized for the subscriber
	ctxWithToken := token.NewContextWithVerifyToken(reqCtx, in.Token)

	entSubscriber, err := h.getSubscriberByToken(ctxWithToken, in.Token)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.BadRequest(ctx, err)
		}

		logx.FromContext(reqCtx).Error().Err(err).Msg("error retrieving subscriber")

		return h.InternalServerError(ctx, ErrUnableToUnsubscribe)
	}

	// scope the caller to the subscriber's owning org so the update passes the org-ownership pre-policy
	// (DenyIfNotInOrganization); the verify token set above is preserved and authorizes the mutation
	ctxWithToken = auth.WithCaller(ctxWithToken, &auth.Caller{
		OrganizationID:  entSubscriber.OwnerID,
		OrganizationIDs: []string{entSubscriber.OwnerID},
	})

	// idempotent replay
	if entSubscriber.Unsubscribed {
		out := &models.UnsubscribeResponse{
			Reply:   rout.Reply{Success: true},
			Message: "You are already unsubscribed and will not receive updates.",
		}

		return h.Success(ctx, out)
	}

	if err := h.setSubscriberUnsubscribed(ctxWithToken, entSubscriber.ID); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error unsubscribing subscriber")

		return h.InternalServerError(ctx, ErrUnableToUnsubscribe)
	}

	out := &models.UnsubscribeResponse{
		Reply:   rout.Reply{Success: true},
		Message: "You have been unsubscribed and will no longer receive updates.",
	}

	return h.Success(ctx, out)
}
