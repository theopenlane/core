package logx_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/shared/logx"
)

func TestNew(t *testing.T) {
	b := &bytes.Buffer{}

	result := logx.Configure(logx.LoggerConfig{
		Writer:   b,
		WithEcho: true,
	})

	l := result.Echo

	l.Print("foo")

	var entry map[string]any
	err := json.Unmarshal(b.Bytes(), &entry)
	assert.NoError(t, err)
	assert.Equal(t, "-", entry["level"])
	assert.Equal(t, "foo", entry["message"])
}

func TestLogger_SetPrefix(t *testing.T) {
	b := &bytes.Buffer{}

	l := logx.Configure(logx.LoggerConfig{
		Writer:   b,
		WithEcho: true,
	}).Echo

	l.Print("t-e-s-t")

	var entry map[string]any
	err := json.Unmarshal(b.Bytes(), &entry)
	assert.NoError(t, err)
	assert.Equal(t, "t-e-s-t", entry["message"])

	b.Reset()

	l.SetPrefix("foo")
	l.Print("test")

	err = json.Unmarshal(b.Bytes(), &entry)
	assert.NoError(t, err)
	assert.Equal(t, "foo", entry["prefix"])
	assert.Equal(t, "test", entry["message"])
}

func TestLogger_Output(t *testing.T) {
	out1 := &bytes.Buffer{}

	l := logx.Configure(logx.LoggerConfig{
		Writer:   out1,
		WithEcho: true,
	}).Echo

	l.Print("foo")
	l.Print("bar")

	out2 := &bytes.Buffer{}
	l.SetOutput(out2)

	l.Print("baz")

	lines := strings.Split(strings.TrimSpace(out1.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	for idx, expected := range []string{"foo", "bar"} {
		var entry map[string]any
		err := json.Unmarshal([]byte(lines[idx]), &entry)
		assert.NoError(t, err)
		assert.Equal(t, expected, entry["message"])
	}

	var entry map[string]any
	err := json.Unmarshal(out2.Bytes(), &entry)
	assert.NoError(t, err)
	assert.Equal(t, "baz", entry["message"])
}

func TestLogger_SetLevel(t *testing.T) {
	b := &bytes.Buffer{}

	l := logx.Configure(logx.LoggerConfig{
		Writer:   b,
		WithEcho: true,
	}).Echo

	l.Debug("foo")

	var entry map[string]any
	err := json.Unmarshal(b.Bytes(), &entry)
	assert.NoError(t, err)
	assert.Equal(t, "foo", entry["message"])
	assert.Equal(t, "debug", entry["level"])

	b.Reset()

	l.SetLevel(log.WARN)

	l.Debug("foo")

	assert.Equal(t, "", b.String())
}

type SimpleLog struct {
	Level zerolog.Level
	Fn    func(i ...interface{})
}

type FormattedLog struct {
	Level zerolog.Level
	Fn    func(format string, i ...interface{})
}

type JSONLog struct {
	Level zerolog.Level
	Fn    func(fields map[string]interface{})
}

func TestLogger(t *testing.T) {
	var b bytes.Buffer
	l := zerolog.New(&b)

	simpleLogs := []SimpleLog{
		{
			Level: zerolog.DebugLevel,
			Fn:    func(i ...interface{}) { l.Debug().Msg(fmt.Sprint(i...)) },
		},
		{
			Level: zerolog.InfoLevel,
			Fn:    func(i ...interface{}) { l.Info().Msg(fmt.Sprint(i...)) },
		},
		{
			Level: zerolog.WarnLevel,
			Fn:    func(i ...interface{}) { l.Warn().Msg(fmt.Sprint(i...)) },
		},
		{
			Level: zerolog.ErrorLevel,
			Fn:    func(i ...interface{}) { l.Error().Msg(fmt.Sprint(i...)) },
		},
	}

	for _, log := range simpleLogs {
		b.Reset()

		log.Fn("foobar")
		assert.Equal(t, fmt.Sprintf(`{"level":"%s","message":"foobar"}
`, log.Level),
			b.String())
	}

	formattedLogs := []FormattedLog{
		{
			Level: zerolog.DebugLevel,
			Fn:    l.Debug().Msgf,
		},
		{
			Level: zerolog.InfoLevel,
			Fn:    l.Info().Msgf,
		},
		{
			Level: zerolog.WarnLevel,
			Fn:    l.Warn().Msgf,
		},
		{
			Level: zerolog.ErrorLevel,
			Fn:    l.Error().Msgf,
		},
	}

	for _, log := range formattedLogs {
		b.Reset()

		log.Fn("foobar %s", "baz")
		assert.Equal(t, fmt.Sprintf(`{"level":"%s","message":"foobar baz"}
`, log.Level),
			b.String())
	}

	jsonLogs := []JSONLog{
		{
			Level: zerolog.DebugLevel,
			Fn:    func(fields map[string]interface{}) { l.Debug().Fields(fields).Send() },
		},
		{
			Level: zerolog.InfoLevel,
			Fn:    func(fields map[string]interface{}) { l.Info().Fields(fields).Send() },
		},
		{
			Level: zerolog.WarnLevel,
			Fn:    func(fields map[string]interface{}) { l.Warn().Fields(fields).Send() },
		},
		{
			Level: zerolog.ErrorLevel,
			Fn:    func(fields map[string]interface{}) { l.Error().Fields(fields).Send() },
		},
	}

	for _, log := range jsonLogs {
		b.Reset()

		log.Fn(map[string]interface{}{
			"message": "foobar",
		})
		assert.Equal(t, fmt.Sprintf(`{"level":"%s","message":"foobar"}
`, log.Level),
			b.String())
	}
}
