package handlers

import (
	"embed"
	"net/http"

	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/pkg/models"
)

//go:embed csv/*.csv
var examplecsv embed.FS

// ExampleCSV will return an example csv file that can be used for bulk uploads of the object
func (h *Handler) ExampleCSV(ctx echo.Context) error {
	var in models.ExampleCSVRequest
	if err := ctx.Bind(&in); err != nil {
		return h.BadRequest(ctx, err)
	}

	if err := in.Validate(); err != nil {
		return h.InvalidInput(ctx, err)
	}

	content, err := examplecsv.ReadFile("csv/sample_" + in.Filename + ".csv")
	if err != nil {
		log.Warn().Msgf("failed to read example csv file: %s", in.Filename)
		return h.InternalServerError(ctx, err)
	}

	return ctx.Blob(http.StatusOK, "text/csv", content)
}
