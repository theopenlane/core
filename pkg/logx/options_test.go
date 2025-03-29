package logx_test

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/pkg/logx"
)

func TestWithCaller(t *testing.T) {
	b := &bytes.Buffer{}

	l := logx.New(b, logx.WithCaller())

	l.Print("foobar")

	type Log struct {
		Level   string `json:"level"`
		Caller  string `json:"caller"`
		Message string `json:"message"`
	}

	log := &Log{}

	err := json.Unmarshal(b.Bytes(), log)

	assert.NoError(t, err)

	segments := strings.Split(log.Caller, ":")
	filePath := filepath.Base(segments[0])

	assert.Equal(t, filePath, "logger.go")
}

func TestWithCallerWithSkipFrameCount(t *testing.T) {
	b := &bytes.Buffer{}

	l := logx.New(b, logx.WithCallerWithSkipFrameCount(3))

	l.Print("foobar")

	type Log struct {
		Level   string `json:"level"`
		Caller  string `json:"caller"`
		Message string `json:"message"`
	}

	log := &Log{}

	err := json.Unmarshal(b.Bytes(), log)

	assert.NoError(t, err)

	segments := strings.Split(log.Caller, ":")
	filePath := filepath.Base(segments[0])

	assert.Equal(t, filePath, "options_test.go")
}

func TestWithField(t *testing.T) {
	b := &bytes.Buffer{}

	l := logx.New(b, logx.WithField("service", "logging"))

	l.Print("foobar")

	type Log struct {
		Level   string `json:"level"`
		Service string `json:"service"`
		Message string `json:"message"`
	}

	log := &Log{}

	err := json.Unmarshal(b.Bytes(), log)

	assert.NoError(t, err)
	assert.Equal(t, log.Service, "logging")
}

func TestWithFields(t *testing.T) {
	b := &bytes.Buffer{}

	l := logx.New(b, logx.WithFields(map[string]interface{}{
		"host": "localhost",
		"port": 8080,
	}))

	l.Print("foobar")

	type Log struct {
		Level   string `json:"level"`
		Host    string `json:"host"`
		Port    int    `json:"port"`
		Message string `json:"message"`
	}

	log := &Log{}

	err := json.Unmarshal(b.Bytes(), log)

	assert.NoError(t, err)
	assert.Equal(t, log.Host, "localhost")
	assert.Equal(t, log.Port, 8080)
}

type (
	Hook struct {
		logs []HookLog
	}

	HookLog struct {
		level   zerolog.Level
		message string
	}
)

func (h *Hook) Run(e *zerolog.Event, level zerolog.Level, message string) {
	h.logs = append(h.logs, HookLog{
		level:   level,
		message: message,
	})
}

func TestWithHook(t *testing.T) {
	b := &bytes.Buffer{}
	h := &Hook{}
	l := logx.New(b, logx.WithHook(h))

	l.Info("Foo")
	l.Warn("Bar")

	assert.Len(t, h.logs, 2)
	assert.Equal(t, h.logs[0].level, zerolog.InfoLevel)
	assert.Equal(t, h.logs[0].message, "Foo")
	assert.Equal(t, h.logs[1].level, zerolog.WarnLevel)
	assert.Equal(t, h.logs[1].message, "Bar")
}

func TestWithHookFunc(t *testing.T) {
	b := &bytes.Buffer{}
	logs := make([]HookLog, 0, 2)
	l := logx.New(b, logx.WithHookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
		logs = append(logs, HookLog{
			level:   level,
			message: message,
		})
	}))

	l.Info("Foo")
	l.Warn("Bar")

	assert.Len(t, logs, 2)
	assert.Equal(t, logs[0].level, zerolog.InfoLevel)
	assert.Equal(t, logs[0].message, "Foo")
	assert.Equal(t, logs[1].level, zerolog.WarnLevel)
	assert.Equal(t, logs[1].message, "Bar")
}

func TestWithLevel(t *testing.T) {
	b := &bytes.Buffer{}
	l := logx.New(b, logx.WithLevel(log.WARN))

	l.Debug("Test")

	assert.Equal(t, b.String(), "")

	l.Warn("Foobar")

	assert.Equal(t, b.String(), `{"level":"warn","message":"Foobar"}
`)
}

func TestWithPrefix(t *testing.T) {
	b := &bytes.Buffer{}
	l := logx.New(b, logx.WithPrefix("Test"))

	l.Warn("Foobar")

	assert.Equal(t, b.String(), `{"level":"warn","prefix":"Test","message":"Foobar"}
`)
}

func TestWithTimestamp(t *testing.T) {
	b := &bytes.Buffer{}

	l := logx.New(b, logx.WithTimestamp())

	l.Print("foobar")

	type Log struct {
		Level   string    `json:"level"`
		Message string    `json:"message"`
		Time    time.Time `json:"time"`
	}

	log := &Log{}

	err := json.Unmarshal(b.Bytes(), log)

	assert.NoError(t, err)
	assert.NotEmpty(t, log.Time)
}
