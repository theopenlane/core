package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerFileUploadRoute registers the file upload route
func registerFileUploadRoute(router *Router) (err error) {
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

	uploadOperation := router.Handler.BindUploadBander()

	if err := router.Addv1Route(path, method, uploadOperation, route); err != nil {
		return err
	}

	return nil
}
