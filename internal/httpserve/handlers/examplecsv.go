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

// ForgotPassword will send an forgot password email if the provided email exists
func (h *Handler) ExampleCSV(ctx echo.Context) error {
	var in models.ExampleCSVRequest
	if err := ctx.Bind(&in); err != nil {
		return h.BadRequest(ctx, err)
	}

	if err := in.Validate(); err != nil {
		return h.InvalidInput(ctx, err)
	}

	file, err := examplecsv.ReadFile("csv/sample_" + in.Filename + ".csv")
	if err != nil {
		log.Error().Err(err).Msgf("failed to read example csv file: %s", in.Filename)
		return h.InternalServerError(ctx, err)
	}

	if len(file) == 0 {
		log.Error().Err(err).Msg("failed to read example csv file")
		return h.InternalServerError(ctx, err)
	}

	return h.Stream(ctx, "text/csv", file)
}

func (h *Handler) Stream(ctx echo.Context, filetype string, rep interface{}) error {
	return ctx.Stream(http.StatusOK, filetype, rep.(http.File))
}
