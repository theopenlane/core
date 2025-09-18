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
		Name:        "File",
		Description: "Upload files to the server storage",
		Tags:        []string{"files"},
		OperationID: "File",
		Security:    handlers.AuthenticatedSecurity,
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.FileUploadHandler,
	}

	return router.AddV1HandlerRoute(config)
}
