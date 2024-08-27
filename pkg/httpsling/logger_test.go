package httpsling

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockLoggerRecorder struct {
	Records []string
}

// Write writes the given bytes to the recorder
func (m *mockLoggerRecorder) Write(p []byte) (n int, err error) {
	m.Records = append(m.Records, string(p))

	return len(p), nil
}

func TestDefaultLoggerLevels(t *testing.T) {
	rec := &mockLoggerRecorder{}
	logger := NewDefaultLogger(rec, LevelDebug)

	logger.Debugf("debug %s", "message")
	logger.Infof("info %s", "message")
	logger.Warnf("warn %s", "message")
	logger.Errorf("error %s", "message")

	assert.Len(t, rec.Records, 4, "Should log 4 messages")
	assert.Contains(t, rec.Records[0], "debug message", "Debug log message should match")
	assert.Contains(t, rec.Records[1], "info message", "Info log message should match")
	assert.Contains(t, rec.Records[2], "warn message", "Warn log message should match")
	assert.Contains(t, rec.Records[3], "error message", "Error log message should match")
}

type mockLogger struct {
	Infos  []string
	Errors []string
}

func (m *mockLogger) Debugf(format string, v ...any) {
	m.Infos = append(m.Infos, fmt.Sprintf(format, v...))
}
func (m *mockLogger) Infof(format string, v ...any) {
	m.Infos = append(m.Infos, fmt.Sprintf(format, v...))
}
func (m *mockLogger) Warnf(format string, v ...any) {
	m.Infos = append(m.Infos, fmt.Sprintf(format, v...))
}
func (m *mockLogger) Errorf(format string, v ...any) {
	m.Errors = append(m.Errors, fmt.Sprintf(format, v...))
}

func (m *mockLogger) SetLevel(level Level) {}
func TestRetryLogMessage(t *testing.T) {
	// Initialize attempt counter
	var attempts int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			// Fail initially to trigger a retry
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			// Succeed in the next attempt
			w.WriteHeader(http.StatusOK)
		}
	}))

	defer server.Close()

	mockLogger := &mockLogger{}
	client := Create(&Config{
		BaseURL: server.URL,
		Logger:  mockLogger,
	}).SetMaxRetries(1).SetRetryStrategy(func(attempt int) time.Duration {
		return 0 // No delay for testing
	})

	// Making a request that should trigger a retry
	_, err := client.Get("/test").Send(context.Background())
	assert.Nil(t, err, "Did not expect an error after retry")

	// Check if the retry log message was recorded
	expectedLogMessage := "Retrying request (attempt 1) after backoff"
	found := false

	for _, logMsg := range mockLogger.Infos {
		if strings.Contains(logMsg, expectedLogMessage) {
			found = true
			break
		}
	}

	assert.True(t, found, "Expected retry log message was not recorded")
}
