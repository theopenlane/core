package genhelpers

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/logx/consolelog"
)

func SetupLogging() {
	// if you want to try the other console writer, swap this out for pzlog.NewPtermWriter()
	output := consolelog.NewConsoleWriter()
	log.Logger = zerolog.New(os.Stderr).
		With().Timestamp().
		Logger()

	// set the log level
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	log.Logger = log.Logger.With().
		Caller().Logger()

	// pretty logging
	log.Logger = log.Output(output)
}
