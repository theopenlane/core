package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerWebhookHandler registers the file upload route
func registerWebhookRoute(router *Router) (err error) {
	path := "/stripe/webhook"
	method := http.MethodPost
	name := "StripeWebhook"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.WebhookHandler(c)
		},
	}

	uploadOperation := router.Handler.BindUploadBander()

	if err := router.Addv1Route(path, method, uploadOperation, route); err != nil {
		return err
	}

	return nil
}
