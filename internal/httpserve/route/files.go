package route

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerUploadsHandler serves up static files from the upload directory
// this is used for development without saving files to S3
// it is *ONLY* registered when the disk mode object storage is used
func registerUploadsHandler(router *Router) (err error) {
	config := Config{
		Path:        "/files/:name",
		Method:      http.MethodGet,
		Name:        "Files",
		Description: "Serve uploaded files from local storage (development only)",
		Tags:        []string{"files"},
		OperationID: "Files",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		SimpleHandler: func(ctx echo.Context) error {
			fileSystem := os.DirFS(router.LocalFilePath)

			p := ctx.PathParam("name")
			name := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(p, "/")))

			return ctx.FileFS(name, fileSystem)
		},
	}

	return router.AddUnversionedHandlerRoute(config)
}
