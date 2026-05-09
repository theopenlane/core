package entdb

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

func TestGooseLoggerPrintfDisabled(t *testing.T) {
	var buf bytes.Buffer

	previousLogger := log.Logger
	previousLevel := zerolog.GlobalLevel()

	log.Logger = zerolog.New(&buf)
	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	t.Cleanup(func() {
		log.Logger = previousLogger
		zerolog.SetGlobalLevel(previousLevel)
	})

	newGooseLogger(false).Printf("OK   %s", "migration.sql")

	require.Empty(t, buf.String())
}

func TestGooseLoggerPrintfEnabled(t *testing.T) {
	var buf bytes.Buffer

	previousLogger := log.Logger
	previousLevel := zerolog.GlobalLevel()

	log.Logger = zerolog.New(&buf)
	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	t.Cleanup(func() {
		log.Logger = previousLogger
		zerolog.SetGlobalLevel(previousLevel)
	})

	newGooseLogger(true).Printf("OK   %s", "migration.sql")

	require.Contains(t, buf.String(), `"component":"goose"`)
	require.Contains(t, buf.String(), `"message":"OK   migration.sql"`)
}
