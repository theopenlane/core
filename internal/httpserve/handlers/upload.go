package handlers

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	echo "github.com/theopenlane/echox"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects"
)

// FileUploadHandler is responsible for uploading files
func (h *Handler) FileUploadHandler(ctx echo.Context) error {
	r := ctx.Request()

	// create the output struct
	out := models.UploadFilesReply{}

	// files are upload via the middleware and stored in the context
	files, err := objects.FilesFromContext(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to get files from context")

		return h.BadRequest(ctx, err)
	}

	// check if any files were uploaded
	// loop through keys
	for _, file := range files {
		// per key, loop through files
		for _, f := range file {
			outFile := models.File{
				ID:           f.ID,
				Name:         f.UploadedFileName,
				PresignedURL: f.PresignedURL,
				MimeType:     f.MimeType,
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

// BindUploadBander binds the upload handler to the OpenAPI schema
func (h *Handler) BindUploadBander() *openapi3.Operation {
	uploadHandler := openapi3.NewOperation()
	uploadHandler.Description = "Upload files such as images, documents, using the multipart form data"
	uploadHandler.OperationID = "Upload"

	h.AddRequestBody("UploadRequest", models.ExampleUploadFileRequest, uploadHandler)
	h.AddResponse("UploadReply", "success", models.ExampleUploadFilesSuccessResponse.Reply, uploadHandler, http.StatusOK)
	uploadHandler.AddResponse(http.StatusInternalServerError, internalServerError())
	uploadHandler.AddResponse(http.StatusBadRequest, badRequest())
	uploadHandler.AddResponse(http.StatusUnauthorized, unauthorized())

	return uploadHandler
}
