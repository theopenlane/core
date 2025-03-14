package handlers

import (
	"embed"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/pkg/models"
)

//go:embed csv/*.csv
var examplecsv embed.FS

// ForgotPassword will send an forgot password email if the provided email exists
func (h *Handler) ExampleCSV(ctx echo.Context) error {
	var in models.ExampleCSVRequest
	if err := ctx.Bind(&in); err != nil {
		return h.BadRequest(ctx, err)
	}

	if err := in.Validate(); err != nil {
		return h.InvalidInput(ctx, err)
	}

	log.Warn().Msgf("input was validated")

	file, err := examplecsv.Open("csv/sample_" + in.Filename + ".csv")
	if err != nil {
		log.Warn().Msgf("failed to open example csv file: %s", in.Filename)
		return h.InternalServerError(ctx, err)
	}
	defer file.Close()

	log.Warn().Msgf("example csv file: %s", in.Filename)

	content, err := io.ReadAll(file)
	if err != nil {
		log.Warn().Msgf("failed to read example csv file: %s", in.Filename)
		return h.InternalServerError(ctx, err)
	}

	return ctx.Blob(http.StatusOK, "text/csv", content)
}
