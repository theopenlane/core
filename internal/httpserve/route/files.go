package route

import (
	"fmt"
	"mime"
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
		Path:        "/files/:orgid/:id/:name",
		Method:      http.MethodGet,
		Name:        "Files",
		Description: "Serve uploaded files from local storage (development only)",
		Tags:        []string{"files"},
		OperationID: "Files",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		SimpleHandler: func(ctx echo.Context) error {
			fileSystem := os.DirFS(router.LocalFilePath)

			// Build the file path from the URL parameters
			orgID := ctx.PathParam("orgid")
			objectID := ctx.PathParam("id")
			fileName := ctx.PathParam("name")

			// Clean and construct the full file path
			name := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(fmt.Sprintf("%s/%s/%s", orgID, objectID, fileName), "/")))

			// Detect and set the correct content type based on file extension
			ext := filepath.Ext(name)
			if contentType := mime.TypeByExtension(ext); contentType != "" {
				ctx.Response().Header().Set(echo.HeaderContentType, contentType)
			}

			return ctx.FileFS(name, fileSystem)
		},
	}

	return router.AddUnversionedHandlerRoute(config)
}

// registerFileDownloadHandler registers the file download handler and route
func registerFileDownloadHandler(router *Router) error {
	config := Config{
		Path:        "/files/:id/download",
		Method:      http.MethodGet,
		Name:        "File Download",
		Description: handlers.AuthEndpointDesc("Download", "files via proxy-signed URLs"),
		Tags:        []string{"files"},
		OperationID: "FileDownload",
		Security:    handlers.AuthenticatedSecurity,
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.FileDownloadHandler,
	}

	return router.AddV1HandlerRoute(config)
}
