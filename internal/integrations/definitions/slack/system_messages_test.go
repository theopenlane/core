package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	slackgo "github.com/slack-go/slack"
	"github.com/stretchr/testify/require"
)

func TestNewSubscriberTemplateRender(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, newSubscriberTemplate.Execute(&buf, NewSubscriberMessage{Email: "alice@example.com"}))
	require.Contains(t, buf.String(), "alice@example.com")
}

func TestGitHubAppInstalledTemplateRender(t *testing.T) {
	t.Parallel()

	input := GitHubAppInstalledMessage{
		GitHubOrganization:         "acme",
		GitHubAccountType:          "Organization",
		OpenlaneOrganization:       "Acme Inc",
		OpenlaneOrganizationID:     "01HABCD",
		ShowOpenlaneOrganizationID: true,
	}

	var buf bytes.Buffer
	require.NoError(t, githubAppInstalledTemplate.Execute(&buf, input))

	out := buf.String()
	require.Contains(t, out, "acme")
	require.Contains(t, out, "Organization")
	require.Contains(t, out, "Acme Inc")
	require.Contains(t, out, "01HABCD")
}

func TestDemoRequestTemplateRender(t *testing.T) {
	t.Parallel()

	input := DemoRequestMessage{
		CompanyName:    "Acme",
		Email:          "buyer@example.com",
		Domains:        []string{"acme.com", "acme.io"},
		CompanyDetails: map[string]any{"size": "500"},
		DemoRequested:  true,
	}

	var buf bytes.Buffer
	require.NoError(t, demoRequestTemplate.Execute(&buf, input))

	out := buf.String()
	require.Contains(t, out, "Acme")
	require.Contains(t, out, "buyer@example.com")
	require.Contains(t, out, "acme.com, acme.io")
	require.Contains(t, out, "size")
	require.Contains(t, out, "Demo requested")
}

func TestSlackClientSendTextEmpty(t *testing.T) {
	t.Parallel()

	client := &SlackClient{WebhookURL: "https://hooks.slack.com/services/T/B/X"}
	err := client.sendText(context.Background(), "", "")
	require.ErrorIs(t, err, ErrMessageEmpty)
}

func TestSlackClientSendTextWebhook(t *testing.T) {
	t.Parallel()

	var (
		mu       sync.Mutex
		received []string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var payload struct {
			Text string `json:"text"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&payload))

		mu.Lock()
		received = append(received, payload.Text)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := &SlackClient{WebhookURL: server.URL}
	err := client.sendText(context.Background(), "hello webhook", "")
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, []string{"hello webhook"}, received)
}

func TestSlackClientSendTextUsesDefaultChannel(t *testing.T) {
	t.Parallel()

	recorder := newSystemMessageRecorder(t)
	server := httptest.NewServer(recorder.handler())
	defer server.Close()

	client := &SlackClient{
		API:            slackgo.New("xoxb-test", slackgo.OptionAPIURL(server.URL+"/")),
		DefaultChannel: "C123",
	}

	err := client.sendText(context.Background(), "hello channel", "")
	require.NoError(t, err)
	require.Equal(t, []string{"C123"}, recorder.channels())
	require.Equal(t, []string{"hello channel"}, recorder.texts())
}

func TestSlackClientSendTextChannelOverridesDefault(t *testing.T) {
	t.Parallel()

	recorder := newSystemMessageRecorder(t)
	server := httptest.NewServer(recorder.handler())
	defer server.Close()

	client := &SlackClient{
		API:            slackgo.New("xoxb-test", slackgo.OptionAPIURL(server.URL+"/")),
		DefaultChannel: "C123",
	}

	err := client.sendText(context.Background(), "hello", "C999")
	require.NoError(t, err)
	require.Equal(t, []string{"C999"}, recorder.channels())
}

func TestSlackClientSendTextNoChannelConfigured(t *testing.T) {
	t.Parallel()

	client := &SlackClient{API: slackgo.New("xoxb-test")}

	err := client.sendText(context.Background(), "hello", "")
	require.ErrorIs(t, err, ErrDefaultChannelMissing)
}

func TestAllSlackSystemMessagesWiredToMessageClient(t *testing.T) {
	t.Parallel()

	regs := AllSlackSystemMessages()
	require.NotEmpty(t, regs)

	for _, reg := range regs {
		if reg.ClientRef != slackClient.ID() {
			t.Fatalf("op %s has ClientRef %v, want %v", reg.Name, reg.ClientRef, slackClient.ID())
		}

		require.True(t, reg.Policy.SkipRunRecord, "op %s must be SkipRunRecord", reg.Name)
	}
}

func TestBuildRuntimeMessageClientWebhookOnly(t *testing.T) {
	t.Parallel()

	cfg := RuntimeSlackConfig{WebhookURL: "https://hooks.slack.com/services/T/B/X"}
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	built, err := buildRuntimeSlackClient(context.Background(), raw)
	require.NoError(t, err)

	client, ok := built.(*SlackClient)
	require.True(t, ok)
	require.Equal(t, cfg.WebhookURL, client.WebhookURL)
	require.Nil(t, client.API)
}

func TestBuildRuntimeMessageClientBotToken(t *testing.T) {
	t.Parallel()

	cfg := RuntimeSlackConfig{
		BotToken:       "xoxb-platform-token",
		DefaultChannel: "C_PLATFORM",
	}
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	built, err := buildRuntimeSlackClient(context.Background(), raw)
	require.NoError(t, err)

	client, ok := built.(*SlackClient)
	require.True(t, ok)
	require.NotNil(t, client.API)
	require.Equal(t, "C_PLATFORM", client.DefaultChannel)
	require.Empty(t, client.WebhookURL)
}

func TestBuildRuntimeMessageClientBotTokenWithWebhookFallback(t *testing.T) {
	t.Parallel()

	cfg := RuntimeSlackConfig{
		BotToken:       "xoxb-platform-token",
		WebhookURL:     "https://hooks.slack.com/services/T/B/X",
		DefaultChannel: "C_PLATFORM",
	}
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	built, err := buildRuntimeSlackClient(context.Background(), raw)
	require.NoError(t, err)

	client, ok := built.(*SlackClient)
	require.True(t, ok)
	require.NotNil(t, client.API)
	require.Equal(t, cfg.WebhookURL, client.WebhookURL)
	require.Equal(t, "C_PLATFORM", client.DefaultChannel)
}

func TestBuildRuntimeMessageClientUnprovisioned(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(RuntimeSlackConfig{})
	require.NoError(t, err)

	_, err = buildRuntimeSlackClient(context.Background(), raw)
	require.True(t, errors.Is(err, ErrRuntimeConfigInvalid))
}

func TestSlackClientSendTextAPIPreferredOverWebhook(t *testing.T) {
	t.Parallel()

	recorder := newSystemMessageRecorder(t)
	server := httptest.NewServer(recorder.handler())
	defer server.Close()

	client := &SlackClient{
		API:            slackgo.New("xoxb-test", slackgo.OptionAPIURL(server.URL+"/")),
		WebhookURL:     "https://hooks.slack.com/should-not-be-called",
		DefaultChannel: "C_RUNTIME",
	}

	err := client.sendText(context.Background(), "bot token wins", "")
	require.NoError(t, err)
	require.Equal(t, []string{"C_RUNTIME"}, recorder.channels())
	require.Equal(t, []string{"bot token wins"}, recorder.texts())
}

// systemMessageRecorder captures chat.postMessage requests for the system-message client tests
type systemMessageRecorder struct {
	t           *testing.T
	mu          sync.Mutex
	channelsRec []string
	textsRec    []string
}

// newSystemMessageRecorder constructs an empty systemMessageRecorder bound to the test
func newSystemMessageRecorder(t *testing.T) *systemMessageRecorder {
	t.Helper()

	return &systemMessageRecorder{t: t}
}

// handler returns an HTTP handler that records incoming chat.postMessage requests
func (r *systemMessageRecorder) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !strings.HasSuffix(req.URL.Path, "/chat.postMessage") {
			http.Error(w, "unexpected path", http.StatusNotFound)
			return
		}

		require.NoError(r.t, req.ParseForm())

		channel := req.PostForm.Get("channel")
		text := req.PostForm.Get("text")

		r.mu.Lock()
		r.channelsRec = append(r.channelsRec, channel)
		r.textsRec = append(r.textsRec, text)
		index := len(r.channelsRec)
		r.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		require.NoError(r.t, json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"channel": channel,
			"ts":      fmt.Sprintf("1700000000.%06d", index),
		}))
	})
}

// channels returns a copy of the recorded channel parameters
func (r *systemMessageRecorder) channels() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]string(nil), r.channelsRec...)
}

// texts returns a copy of the recorded text parameters
func (r *systemMessageRecorder) texts() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]string(nil), r.textsRec...)
}
