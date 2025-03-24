package echolog_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/pkg/logx/echolog"
)

func TestNew(t *testing.T) {
	b := &bytes.Buffer{}

	l := echolog.New(b)

	l.Print("foo")

	assert.Equal(
		t,
		`{"level":"-","message":"foo"}
`,
		b.String(),
	)
}

func TestNewWithZerolog(t *testing.T) {
	b := &bytes.Buffer{}
	zl := zerolog.New(b)

	l := echolog.New(zl.With().Str("key", "test").Logger())

	l.Print("foo")

	assert.Equal(
		t,
		`{"key":"test","level":"-","message":"foo"}
`,
		b.String(),
	)
}

func TestFrom(t *testing.T) {
	b := &bytes.Buffer{}

	zl := zerolog.New(b)
	l := echolog.From(zl.With().Str("key", "test").Logger())

	l.Print("foo")

	assert.Equal(
		t,
		`{"key":"test","level":"-","message":"foo"}
`,
		b.String(),
	)
}

func TestLogger_SetPrefix(t *testing.T) {
	//	b := &bytes.Buffer{}
	//
	//	l := echolog.New(b)
	//
	//	l.Print("t-e-s-t")
	//
	//	assert.Equal(
	//		t,
	//		`{"level":"-","message":"t-e-s-t"}
	//`,
	//		b.String(),
	//	)
	//
	//	b.Reset()
	//
	//	l.SetPrefix("foo")
	//	l.Print("test")
	//
	//	assert.Equal(
	//		t,
	//		`{"prefix":"foo","level":"-","message":"test"}
	//`,
	//		b.String(),
	//	)
	//
	//	b.Reset()
	//
	//	l.SetPrefix("bar")
	//	l.Print("test-test")
	//
	//	assert.Equal(
	//		t,
	//		`{"prefix":"bar","level":"-","message":"test-test"}
	//`,
	//		b.String(),
	//	)
}

func TestLogger_Output(t *testing.T) {
	out1 := &bytes.Buffer{}

	l := echolog.New(out1)

	l.Print("foo")
	l.Print("bar")

	out2 := &bytes.Buffer{}
	l.SetOutput(out2)

	l.Print("baz")

	assert.Equal(
		t,
		`{"level":"-","message":"foo"}
{"level":"-","message":"bar"}
`,
		out1.String(),
	)

	assert.Equal(
		t,
		`{"level":"-","message":"baz"}
`,
		out2.String(),
	)
}

func TestLogger_SetLevel(t *testing.T) {
	b := &bytes.Buffer{}

	l := echolog.New(b)

	l.Debug("foo")

	assert.Equal(
		t,
		`{"level":"debug","message":"foo"}
`,
		b.String(),
	)

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
