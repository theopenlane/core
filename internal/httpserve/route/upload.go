package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerFileUploadRoute registers the file upload route
func registerFileUploadRoute(router *Router) (err error) { // nolint:unused
	path := "/upload"
	method := http.MethodPost
	name := "FileUpload"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.FileUploadHandler(c)
		},
	}

	if err := router.AddEchoOnlyRoute(route); err != nil {
		return err
	}

	return nil
}
