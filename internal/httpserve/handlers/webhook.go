package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v81/webhook"
	echo "github.com/theopenlane/echox"

	ent "github.com/theopenlane/core/internal/ent/generated"
)

// WebhookReceiverHandler handles incoming stripe webhook events
func (h *Handler) WebhookReceiverHandler(ctx echo.Context) error {
	req := ctx.Request()
	res := ctx.Response()

	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(res.Writer, req.Body, MaxBodyBytes)

	payload, err := io.ReadAll(req.Body)
	if err != nil {
		return ctx.String(http.StatusServiceUnavailable, fmt.Errorf("problem with request. Error: %w", err).Error())
	}

	event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"), h.Entitlements.Config.StripeWebhookSecret)
	if err != nil {
		return ctx.String(http.StatusBadRequest, fmt.Errorf("error verifying webhook signature. Error: %w", err).Error())
	}

	exists, err := h.checkForEventID(req.Context(), event.ID)
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	if !exists {
		_, err := h.Entitlements.HandleEvent(req.Context(), &event)
		if err != nil {
			return h.InternalServerError(ctx, err)
		}

		input := ent.CreateEventInput{
			EventID:   &event.ID,
			EventType: "stripe",
			// TODO unmarshall event data into internal event
		}

		meowevent, err := h.createEvent(req.Context(), input)
		if err != nil {
			return h.InternalServerError(ctx, err)
		}

		log.Debug().Msgf("Internal event: %v", meowevent)
	}

	out := WebhookResponse{
		Message: "Received!",
	}

	return h.Success(ctx, out)
}

// WebhookRequest is the request object for the webhook handler
type WebhookRequest struct {
	// TODO determine if there's any request data actually need or needs to be validated given the signature verification that's already occurring
}

// WebhookResponse is the response object for the webhook handler
type WebhookResponse struct {
	Message string
}

// Validate validates the webhook request
func (r *WebhookRequest) Validate() error {
	return nil
}
