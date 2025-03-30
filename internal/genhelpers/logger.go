package genhelpers

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/logx/consolelog"
)

// SetupLogging sets up the logging for the code generation process
func SetupLogging() {
	output := consolelog.NewConsoleWriter()
	log.Logger = zerolog.New(os.Stderr).
		With().Timestamp().
		Logger()

	// set the log level
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	log.Logger = log.Logger.With().
		Caller().Logger()

	// pretty logging
	log.Logger = log.Output(&output)
}
