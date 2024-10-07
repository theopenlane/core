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

	//	mw = append(mw, router.Handler.ObjectStorage.UploadHandlerCopy(("uploadFile")))

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return router.Handler.FileUploadHandler(c, "uploadFile")
		},
	}
	//	switchOperation := router.Handler.BindFileUploadHandler()

	if err := router.AddEchoOnlyRoute(path, method, route); err != nil {
		return err
	}

	return nil
}
