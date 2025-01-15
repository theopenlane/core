package route

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	echo "github.com/theopenlane/echox"
)

// registerUploadsHandler serves up static files from the upload directory
// this is used for development without saving files to S3
// it is *ONLY* registered when the disk mode object storage is used
func registerUploadsHandler(router *Router) (err error) {
	path := "/files/:name"
	method := http.MethodGet
	name := "files"

	route := echo.Route{
		Name:   name,
		Method: method,
		Path:   path,
		// this route is not protected with auth middleware
		// it is only used for development
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			fileSystem := os.DirFS(router.LocalFilePath)

			p := c.PathParam("name")
			name := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(p, "/")))

			return c.FileFS(name, fileSystem)
		},
	}

	if err := router.AddEchoOnlyRoute(path, method, route); err != nil {
		return err
	}

	return nil
}
