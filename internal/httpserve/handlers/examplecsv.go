package handlers

import (
	"embed"
	"net/http"

	"github.com/rs/zerolog/log"
	models "github.com/theopenlane/core/pkg/openapi"
	echo "github.com/theopenlane/echox"
)

//go:embed csv/*.csv
var examplecsv embed.FS

// ExampleCSV will return an example csv file that can be used for bulk uploads of the object
func (h *Handler) ExampleCSV(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleCSVRequest{}, models.ExampleUploadFilesSuccessResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	content, err := examplecsv.ReadFile("csv/sample_" + in.Filename + ".csv")
	if err != nil {
		log.Warn().Msgf("failed to read example csv file: %s", in.Filename)
		return h.InternalServerError(ctx, err, openapi)
	}

	return ctx.Blob(http.StatusOK, "text/csv", content)
}
