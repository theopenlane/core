package entdb

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
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

	assert.Assert(t, is.Equal("", buf.String()))
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

	assert.Assert(t, is.Contains(buf.String(), `"component":"goose"`))
	assert.Assert(t, is.Contains(buf.String(), `"message":"OK   migration.sql"`))
}
