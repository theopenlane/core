package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

var (
	// uploadKeys are the keys that can be used to upload files in a multipart form
	uploadKeys = []string{"uploadFile"}
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
			return router.Handler.FileUploadHandler(c, uploadKeys...)
		},
	}

	uploadOperation := router.Handler.BindUploadBander()

	if err := router.Addv1Route(path, method, uploadOperation, route); err != nil {
		return err
	}

	return nil
}
