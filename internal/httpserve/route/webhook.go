package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerWebhookHandler registers a webhook endpoint handler behind the /stripe/ path for handling inbound event receivers from stripe
func registerWebhookHandler(router *Router) (err error) {
	path := "/stripe/webhook"
	method := http.MethodPost
	name := "StripeWebhook"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: baseMW, // leaves off the additional middleware(including csrf)
		Handler: func(c echo.Context) error {
			return router.Handler.WebhookReceiverHandler(c)
		},
	}

	if err := router.AddEchoOnlyRoute(route); err != nil {
		return err
	}

	return nil
}
