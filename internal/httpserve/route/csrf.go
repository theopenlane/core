package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerCSRFHandler serves up the csrf token for the UI to use
func registerCSRFHandler(router *Router) (err error) {
	config := Config{
		Path:        "/csrf",
		Method:      http.MethodGet,
		Name:        "CSRF",
		Description: "Get CSRF token for form submissions",
		Tags:        []string{"security"},
		OperationID: "CSRF",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		SimpleHandler: func(ctx echo.Context) error {
			token := ctx.Get(middleware.DefaultCSRFConfig.ContextKey)

			return ctx.JSON(http.StatusOK, echo.Map{
				"csrf": token,
			})
		},
	}

	return router.AddUnversionedHandlerRoute(config)
}
