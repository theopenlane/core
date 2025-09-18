package handlers

import (
	echo "github.com/theopenlane/echox"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/objects"
	models "github.com/theopenlane/core/pkg/openapi"
)

// FileUploadHandler is responsible for uploading files
func (h *Handler) FileUploadHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	r := ctx.Request()

	// create the output struct
	out := models.UploadFilesReply{}

	// files are upload via the middleware and stored in the context
	files, err := objects.FilesFromContext(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to get files from context")

		return h.BadRequest(ctx, err, openapi)
	}

	// check if any files were uploaded
	// loop through keys
	for _, file := range files {
		// per key, loop through files
		for _, f := range file {
			outFile := models.File{
				ID:           f.ID,
				Name:         f.Name,
				PresignedURL: f.PresignedURL,
				MimeType:     f.ContentType,
				ContentType:  f.ContentType,
				MD5:          f.MD5,
				Size:         f.Size,
				CreatedAt:    f.CreatedAt,
				UpdatedAt:    f.UpdatedAt,
			}

			out.Files = append(out.Files, outFile)
			out.FileCount++
		}
	}

	out.Message = "file(s) uploaded successfully"
	out.Success = true

	// return the response
	return h.SuccessBlob(ctx, out)
}
