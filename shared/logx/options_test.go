package logx_test

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/shared/logx"
)

func TestConfigureWithCaller(t *testing.T) {
	b := &bytes.Buffer{}

	logger := logx.Configure(logx.LoggerConfig{
		Writer:        b,
		WithEcho:      true,
		IncludeCaller: true,
	}).Echo

	logger.Print("foobar")

	var entry map[string]any
	err := json.Unmarshal(b.Bytes(), &entry)
	assert.NoError(t, err)

	segments := strings.Split(entry["caller"].(string), ":")
	filePath := filepath.Base(segments[0])

	assert.Equal(t, "logger.go", filePath)
}

type hookLog struct {
	level   zerolog.Level
	message string
}

func TestConfigureWithHookFunc(t *testing.T) {
	b := &bytes.Buffer{}
	var logs []hookLog

	logger := logx.Configure(logx.LoggerConfig{
		Writer:   b,
		WithEcho: true,
		Hooks: []zerolog.HookFunc{
			func(e *zerolog.Event, level zerolog.Level, message string) {
				logs = append(logs, hookLog{
					level:   level,
					message: message,
				})
			},
		},
	}).Echo

	logger.Info("Foo")
	logger.Warn("Bar")

	assert.Len(t, logs, 2)
	assert.Equal(t, zerolog.InfoLevel, logs[0].level)
	assert.Equal(t, "Foo", logs[0].message)
	assert.Equal(t, zerolog.WarnLevel, logs[1].level)
	assert.Equal(t, "Bar", logs[1].message)
}

func TestConfigureWithLevel(t *testing.T) {
	b := &bytes.Buffer{}

	logger := logx.Configure(logx.LoggerConfig{
		Writer:   b,
		WithEcho: true,
		Level:    zerolog.WarnLevel,
	}).Echo

	logger.Debug("Test")
	assert.Equal(t, "", b.String())

	logger.Warn("Foobar")

	var entry map[string]any
	err := json.Unmarshal(b.Bytes(), &entry)
	assert.NoError(t, err)
	assert.Equal(t, "Foobar", entry["message"])
	assert.Equal(t, "warn", entry["level"])
}

func TestConfigureWithTimestamp(t *testing.T) {
	b := &bytes.Buffer{}

	logger := logx.Configure(logx.LoggerConfig{
		Writer:   b,
		WithEcho: true,
	}).Echo

	logger.Print("foobar")

	var entry struct {
		Level   string    `json:"level"`
		Message string    `json:"message"`
		Time    time.Time `json:"time"`
	}

	err := json.Unmarshal(b.Bytes(), &entry)

	assert.NoError(t, err)
	assert.NotEmpty(t, entry.Time)
	assert.Equal(t, "foobar", entry.Message)
}
