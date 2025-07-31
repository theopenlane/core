package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerFileUploadRoute registers the file upload route
func registerFileUploadRoute(router *Router) (err error) { // nolint:unused
	config := Config{
		Path:        "/upload",
		Method:      http.MethodPost,
		Name:        "FileUpload",
		Description: "Upload files to the server storage",
		Tags:        []string{"files"},
		OperationID: "FileUpload",
		Security:    handlers.AuthenticatedSecurity,
		Middlewares: *AuthenticatedEndpoint,
		Handler:     router.Handler.FileUploadHandler,
	}

	return router.AddV1HandlerRoute(config)
}
