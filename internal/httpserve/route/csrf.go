package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware"
)

// registerCSRFHandler serves up the csrf token for the UI to use
func registerCSRFHandler(router *Router) (err error) {
	path := "/csrf"
	method := http.MethodGet
	name := "CSRF"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			token := c.Get(middleware.DefaultCSRFConfig.ContextKey)

			return c.JSON(http.StatusOK, echo.Map{
				"csrf": token,
			})
		}}

	if err := router.AddEchoOnlyRoute(route); err != nil {
		return err
	}

	return nil
}
