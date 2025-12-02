package genhelpers

import (
	"os"

	"github.com/rs/zerolog"

	"github.com/theopenlane/core/pkg/logx"
)

// SetupLogging sets up the logging for the code generation process
func SetupLogging() {
	logx.Configure(logx.LoggerConfig{
		Level:         zerolog.DebugLevel,
		Pretty:        true,
		Writer:        os.Stderr,
		IncludeCaller: true,
		SetGlobal:     true,
	})
}
