//go:build test

package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// SlackTestMessage represents a message captured by the mock Slack server
type SlackTestMessage struct {
	// Channel is the target channel for the message
	Channel string
	// Text is the plain-text content
	Text string
}

// SlackMessageRecorder captures messages sent through a mock Slack API server
type SlackMessageRecorder struct {
	mu       sync.Mutex
	messages []SlackTestMessage
}

// Messages returns a copy of all recorded messages
func (r *SlackMessageRecorder) Messages() []SlackTestMessage {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]SlackTestMessage, len(r.messages))
	copy(out, r.messages)

	return out
}

// Reset clears all recorded messages
func (r *SlackMessageRecorder) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.messages = nil
}

func (r *SlackMessageRecorder) record(channel, text string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.messages = append(r.messages, SlackTestMessage{Channel: channel, Text: text})
}

// MockSlackRuntime holds mock Slack infrastructure for integration test suites
type MockSlackRuntime struct {
	// Server is the mock HTTP server backing the Slack API
	Server *httptest.Server
	// Recorder captures all messages sent through the mock
	Recorder *SlackMessageRecorder
}

// NewMockSlackRuntime creates a mock Slack test server and recorder
func NewMockSlackRuntime() *MockSlackRuntime {
	recorder := &SlackMessageRecorder{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := req.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		channel := req.PostForm.Get("channel")
		text := req.PostForm.Get("text")

		recorder.record(channel, text)

		w.Header().Set("Content-Type", "application/json")

		idx := len(recorder.Messages())

		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"channel": channel,
			"ts":      fmt.Sprintf("1700000000.%06d", idx),
		})
	}))

	return &MockSlackRuntime{
		Server:   server,
		Recorder: recorder,
	}
}

// Close shuts down the mock HTTP server
func (m *MockSlackRuntime) Close() {
	m.Server.Close()
}

// Builder returns a Slack definition builder backed by the mock server.
// All client builds return a SlackClient with the API pointed at the mock,
// bypassing credential resolution entirely
func (m *MockSlackRuntime) Builder() registry.Builder {
	mockAPIURL := m.Server.URL + "/"

	return registry.Builder(func() (types.Definition, error) {
		def, err := Builder(Config{}, nil)()
		if err != nil {
			return types.Definition{}, err
		}

		for i := range def.Clients {
			def.Clients[i].Build = func(_ context.Context, _ types.ClientBuildRequest) (any, error) {
				return &SlackClient{
					API: slackgo.New("mock-token", slackgo.OptionAPIURL(mockAPIURL)),
				}, nil
			}
		}

		return def, nil
	})
}
