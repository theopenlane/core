package consolelog_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/theopenlane/shared/logx/consolelog"
)

func ExampleNewConsoleWriter() {
	output := consolelog.NewConsoleWriter()
	logger := zerolog.New(&output)

	logger.Info().Str("foo", "bar").Msg("hello world")
	// Output: INF hello world foo=bar
}

func ExampleNewConsoleWriter_custom() {
	output := consolelog.NewConsoleWriter(
		// Customize time formatting
		//
		func(w *consolelog.ConsoleWriter) {
			w.TimeFormat = time.RFC822
		},
		// Customize "level" formatting
		//
		func(w *consolelog.ConsoleWriter) {
			w.SetFormatter(
				zerolog.LevelFieldName,
				func(i any) string { return strings.ToUpper(fmt.Sprintf("%-5s", i)) })
		},
	)

	logger := zerolog.New(&output).With().Timestamp().Logger()

	logger.Info().Str("foo", "bar").Msg("hello world")
	// => 19 Jul 18 15:50 CEST INFO  hello world foo=bar
}
