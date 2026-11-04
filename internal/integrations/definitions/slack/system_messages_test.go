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

	"github.com/samber/lo"
	slackgo "github.com/slack-go/slack"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestIntegrationInstalledTemplateRender(t *testing.T) {
	t.Parallel()

	input := IntegrationInstalledMessage{
		IntegrationName:  "Slack",
		OrganizationName: "Acme Inc",
		OrganizationID:   "01HABCD",
	}

	var buf bytes.Buffer
	assert.NilError(t, integrationInstalledTemplate.Execute(&buf, input))

	out := buf.String()
	assert.Assert(t, is.Contains(out, "Slack"))
	assert.Assert(t, is.Contains(out, "Acme Inc"))
	assert.Assert(t, is.Contains(out, "01HABCD"))
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
	assert.NilError(t, demoRequestTemplate.Execute(&buf, input))

	out := buf.String()
	assert.Assert(t, is.Contains(out, "Acme"))
	assert.Assert(t, is.Contains(out, "buyer@example.com"))
	assert.Assert(t, is.Contains(out, "acme.com, acme.io"))
	assert.Assert(t, is.Contains(out, "size"))
	assert.Assert(t, is.Contains(out, "Demo requested"))
}

func TestOrganizationsPendingDeletionTemplateRender(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    OrganizationsPendingDeletionMessage
		contains []string
	}{
		{
			name: "without dry run",
			input: OrganizationsPendingDeletionMessage{
				Count: 3,
				Organizations: []string{
					"Openlane Inc (openlane-1)",
					"Github Inc (github)",
					"Google (google)",
				},
			},
			contains: []string{
				"Organization deletion reminders scheduled",
				"Organizations marked for deletion: 3",
				"- Github Inc",
				"- Openlane Inc",
			},
		},
		{
			name: "with dry run",
			input: OrganizationsPendingDeletionMessage{
				Count:         1,
				Organizations: []string{"Openlane Inc (openlane-1)"},
				DryRun:        true,
			},
			contains: []string{
				"[Dry run] Organization deletion reminders scheduled",
				"- Openlane Inc (openlane-1)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			assert.NilError(t, orgDeletionReminderTemplate.Execute(&buf, tt.input))

			for _, expected := range tt.contains {
				assert.Assert(t, is.Contains(buf.String(), expected))
			}
		})
	}
}

func TestSlackClientSendTextEmpty(t *testing.T) {
	t.Parallel()

	client := &Client{WebhookURL: "https://hooks.slack.com/services/T/B/X"}
	err := client.sendText(context.Background(), "", "")
	assert.ErrorIs(t, err, ErrMessageEmpty)
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
		assert.NilError(t, json.NewDecoder(req.Body).Decode(&payload))

		mu.Lock()
		received = append(received, payload.Text)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := &Client{WebhookURL: server.URL}
	err := client.sendText(context.Background(), "hello webhook", "")
	assert.NilError(t, err)

	mu.Lock()
	defer mu.Unlock()
	assert.Assert(t, is.DeepEqual([]string{"hello webhook"}, received))
}

func TestSlackClientSendTextUsesDefaultChannel(t *testing.T) {
	t.Parallel()

	recorder := newSystemMessageRecorder(t)
	server := httptest.NewServer(recorder.handler())
	defer server.Close()

	client := &Client{
		API:            slackgo.New("xoxb-test", slackgo.OptionAPIURL(server.URL+"/")),
		DefaultChannel: "C123",
	}

	err := client.sendText(context.Background(), "hello channel", "")
	assert.NilError(t, err)
	assert.Assert(t, is.DeepEqual([]string{"C123"}, recorder.channels()))
	assert.Assert(t, is.DeepEqual([]string{"hello channel"}, recorder.texts()))
}

func TestSlackClientSendTextChannelOverridesDefault(t *testing.T) {
	t.Parallel()

	recorder := newSystemMessageRecorder(t)
	server := httptest.NewServer(recorder.handler())
	defer server.Close()

	client := &Client{
		API:            slackgo.New("xoxb-test", slackgo.OptionAPIURL(server.URL+"/")),
		DefaultChannel: "C123",
	}

	err := client.sendText(context.Background(), "hello", "C999")
	assert.NilError(t, err)
	assert.Assert(t, is.DeepEqual([]string{"C999"}, recorder.channels()))
}

func TestSlackClientSendTextNoChannelConfigured(t *testing.T) {
	t.Parallel()

	client := &Client{API: slackgo.New("xoxb-test")}

	err := client.sendText(context.Background(), "hello", "")
	assert.ErrorIs(t, err, ErrDefaultChannelMissing)
}

func TestAllSlackSystemMessagesWiredToMessageClient(t *testing.T) {
	t.Parallel()

	def, err := Builder(Config{}, nil, false)()
	assert.NilError(t, err)

	clientNames := lo.Keys(def.ConnectionList()[0].Clients)
	assert.Assert(t, is.Len(clientNames, 1))

	regs := AllSlackSystemMessages()
	assert.Assert(t, len(regs) > 0)

	for _, reg := range regs {
		if reg.ClientRef != clientNames[0] {
			t.Fatalf("op %s has ClientRef %v, want %v", reg.Name, reg.ClientRef, clientNames[0])
		}

		assert.Assert(t, reg.Policy.SkipRunRecord, "op %s must be SkipRunRecord", reg.Name)
	}
}

func TestBuildRuntimeMessageClientWebhookOnly(t *testing.T) {
	t.Parallel()

	cfg := RuntimeSlackConfig{WebhookURL: "https://hooks.slack.com/services/T/B/X"}
	raw, err := json.Marshal(cfg)
	assert.NilError(t, err)

	built, err := runtimeSlackClientBuilder(false)(context.Background(), raw)
	assert.NilError(t, err)

	client, ok := built.(*Client)
	assert.Assert(t, ok)
	assert.Assert(t, is.Equal(cfg.WebhookURL, client.WebhookURL))
	assert.Assert(t, is.Nil(client.API))
}

func TestBuildRuntimeMessageClientBotToken(t *testing.T) {
	t.Parallel()

	cfg := RuntimeSlackConfig{
		BotToken:       "xoxb-platform-token",
		DefaultChannel: "C_PLATFORM",
	}
	raw, err := json.Marshal(cfg)
	assert.NilError(t, err)

	built, err := runtimeSlackClientBuilder(false)(context.Background(), raw)
	assert.NilError(t, err)

	client, ok := built.(*Client)
	assert.Assert(t, ok)
	assert.Assert(t, client.API != nil)
	assert.Assert(t, is.Equal("C_PLATFORM", client.DefaultChannel))
	assert.Assert(t, is.Len(client.WebhookURL, 0))
}

func TestBuildRuntimeMessageClientBotTokenWithWebhookFallback(t *testing.T) {
	t.Parallel()

	cfg := RuntimeSlackConfig{
		BotToken:       "xoxb-platform-token",
		WebhookURL:     "https://hooks.slack.com/services/T/B/X",
		DefaultChannel: "C_PLATFORM",
	}
	raw, err := json.Marshal(cfg)
	assert.NilError(t, err)

	built, err := runtimeSlackClientBuilder(false)(context.Background(), raw)
	assert.NilError(t, err)

	client, ok := built.(*Client)
	assert.Assert(t, ok)
	assert.Assert(t, client.API != nil)
	assert.Assert(t, is.Equal(cfg.WebhookURL, client.WebhookURL))
	assert.Assert(t, is.Equal("C_PLATFORM", client.DefaultChannel))
}

func TestBuildRuntimeMessageClientUnprovisioned(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(RuntimeSlackConfig{})
	assert.NilError(t, err)

	_, err = runtimeSlackClientBuilder(false)(context.Background(), raw)
	assert.Assert(t, errors.Is(err, ErrRuntimeConfigInvalid))
}

func TestDevModeNoCredentials_SystemMessageNoOps(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(RuntimeSlackConfig{})
	assert.NilError(t, err)

	built, err := runtimeSlackClientBuilder(true)(context.Background(), raw)
	assert.NilError(t, err)

	client := built.(*Client)
	assert.NilError(t, client.sendText(context.Background(), "this message should be silently dropped", ""))
}

func TestDevModeWithCredentials_SystemMessageDelivers(t *testing.T) {
	t.Parallel()

	recorder := newSystemMessageRecorder(t)
	server := httptest.NewServer(recorder.handler())
	defer server.Close()

	cfg := RuntimeSlackConfig{
		BotToken:       "xoxb-real-token",
		DefaultChannel: "C_REAL",
	}
	raw, err := json.Marshal(cfg)
	assert.NilError(t, err)

	built, err := runtimeSlackClientBuilder(true)(context.Background(), raw)
	assert.NilError(t, err)

	client := built.(*Client)
	client.API = slackgo.New("xoxb-real-token", slackgo.OptionAPIURL(server.URL+"/"))

	assert.NilError(t, client.sendText(context.Background(), "should deliver normally", ""))
	assert.Assert(t, is.DeepEqual([]string{"C_REAL"}, recorder.channels()))
	assert.Assert(t, is.DeepEqual([]string{"should deliver normally"}, recorder.texts()))
}

func TestSlackClientSendTextAPIPreferredOverWebhook(t *testing.T) {
	t.Parallel()

	recorder := newSystemMessageRecorder(t)
	server := httptest.NewServer(recorder.handler())
	defer server.Close()

	client := &Client{
		API:            slackgo.New("xoxb-test", slackgo.OptionAPIURL(server.URL+"/")),
		WebhookURL:     "https://hooks.slack.com/should-not-be-called",
		DefaultChannel: "C_RUNTIME",
	}

	err := client.sendText(context.Background(), "bot token wins", "")
	assert.NilError(t, err)
	assert.Assert(t, is.DeepEqual([]string{"C_RUNTIME"}, recorder.channels()))
	assert.Assert(t, is.DeepEqual([]string{"bot token wins"}, recorder.texts()))
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

		assert.NilError(r.t, req.ParseForm())

		channel := req.PostForm.Get("channel")
		text := req.PostForm.Get("text")

		r.mu.Lock()
		r.channelsRec = append(r.channelsRec, channel)
		r.textsRec = append(r.textsRec, text)
		index := len(r.channelsRec)
		r.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		assert.NilError(r.t, json.NewEncoder(w).Encode(map[string]any{
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
