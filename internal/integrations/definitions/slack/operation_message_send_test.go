package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	slackgo "github.com/slack-go/slack"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/templatekit"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestMessageSendRunDestinations(t *testing.T) {
	t.Parallel()

	recorded := newSlackPostMessageRecorder()
	server := httptest.NewServer(recorded.handler(t))
	defer server.Close()

	client := &SlackClient{API: slackgo.New("testing-token", slackgo.OptionAPIURL(server.URL+"/"))}

	resultBytes, err := MessageSend{}.Run(context.Background(), types.OperationRequest{}, client, MessageSendOperation{
		Destinations: []string{"C11111", "C22222"},
		Text:         "hello from destinations",
	})
	require.NoError(t, err)

	var result MessageSend
	require.NoError(t, json.Unmarshal(resultBytes, &result))
	require.Len(t, result.Deliveries, 2)
	require.Equal(t, "C11111", result.Channel)
	require.Equal(t, "C11111", result.Deliveries[0].Channel)
	require.Equal(t, "C22222", result.Deliveries[1].Channel)
	require.Equal(t, []string{"C11111", "C22222"}, recorded.channelValues())
	require.Equal(t, []string{"hello from destinations", "hello from destinations"}, recorded.texts())
}

func TestMessageSendRunDedupesChannelAndDestinations(t *testing.T) {
	t.Parallel()

	recorded := newSlackPostMessageRecorder()
	server := httptest.NewServer(recorded.handler(t))
	defer server.Close()

	client := &SlackClient{API: slackgo.New("testing-token", slackgo.OptionAPIURL(server.URL+"/"))}

	resultBytes, err := MessageSend{}.Run(context.Background(), types.OperationRequest{}, client, MessageSendOperation{
		Channel:      "C11111",
		Destinations: []string{"C11111", "C22222", "C22222"},
		Text:         "hello dedupe",
	})
	require.NoError(t, err)

	var result MessageSend
	require.NoError(t, json.Unmarshal(resultBytes, &result))
	require.Len(t, result.Deliveries, 2)
	require.Equal(t, []string{"C11111", "C22222"}, recorded.channelValues())
}

func TestResolveOperationTemplateNoop(t *testing.T) {
	t.Parallel()

	cfg := MessageSendOperation{
		Channel: "C11111",
		Text:    "hello",
	}

	err := templatekit.ResolveOperationTemplate(context.Background(), types.OperationRequest{}, cfg.TemplateID, cfg.TemplateKey, &cfg)
	require.NoError(t, err)
	require.Equal(t, "C11111", cfg.Channel)
	require.Equal(t, "hello", cfg.Text)
}

func TestResolveOperationTemplateBothRefsUseResolutionPath(t *testing.T) {
	t.Parallel()

	cfg := MessageSendOperation{
		TemplateID:  "some-id",
		TemplateKey: "some-key",
		Channel:     "C11111",
		Text:        "hello",
	}

	err := templatekit.ResolveOperationTemplate(context.Background(), types.OperationRequest{}, cfg.TemplateID, cfg.TemplateKey, &cfg)
	require.ErrorIs(t, err, templatekit.ErrTemplateNotFound)
}

// slackPostMessageRecorder captures chat.postMessage requests made against a test HTTP server
type slackPostMessageRecorder struct {
	// mu guards concurrent access to channels and text slices
	mu sync.Mutex
	// channels holds the channel parameter from each recorded request in order
	channels []string
	// text holds the text parameter from each recorded request in order
	text []string
}

// newSlackPostMessageRecorder returns an empty slackPostMessageRecorder
func newSlackPostMessageRecorder() *slackPostMessageRecorder {
	return &slackPostMessageRecorder{}
}

// handler returns an HTTP handler that records Slack chat.postMessage requests
func (r *slackPostMessageRecorder) handler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			t.Errorf("unexpected method: got %s want %s", req.Method, http.MethodPost)
			http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		if req.URL.Path != "/chat.postMessage" {
			t.Errorf("unexpected path: got %s want %s", req.URL.Path, "/chat.postMessage")
			http.Error(w, "unexpected path", http.StatusNotFound)
			return
		}
		if err := req.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}

		channel := req.PostForm.Get("channel")
		text := req.PostForm.Get("text")

		r.mu.Lock()
		r.channels = append(r.channels, channel)
		r.text = append(r.text, text)
		index := len(r.channels)
		r.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"channel": channel,
			"ts":      fmt.Sprintf("1700000000.%06d", index),
		}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	})
}

// channelValues returns a copy of the recorded channel parameters
func (r *slackPostMessageRecorder) channelValues() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]string(nil), r.channels...)
}

// texts returns a copy of the recorded text parameters
func (r *slackPostMessageRecorder) texts() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]string(nil), r.text...)
}
