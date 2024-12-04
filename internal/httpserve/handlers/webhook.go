package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stripe/stripe-go/v81"
	echo "github.com/theopenlane/echox"
)

func (h *Handler) WebhookHandler(ctx echo.Context) error {
	// Parse the webhook event
	event, err := h.StripeWebhook.ParseEvent(ctx.Request())
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	// Handle the event
	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}

		// Fulfill the purchase
		// handlePaymentIntentSucceeded(paymentIntent)
	case "payment_method.attached":
		var paymentMethod stripe.PaymentMethod
		err := json.Unmarshal(event.Data.Raw, &paymentMethod)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}

		// handlePaymentMethodAttached(paymentMethod)
	// ... handle other event types
	default:
		// Unhandled event type
		fmt.Printf("Unhandled event type: %s\n", event.Type)
	}

	return ctx.JSON(http.StatusOK, "success")
}
