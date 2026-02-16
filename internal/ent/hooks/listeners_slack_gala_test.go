package hooks

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"entgo.io/ent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/subscriber"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/pkg/gala"
)

// TestRegisterGalaSlackListeners verifies Slack listener registration for Gala mutation topics.
func TestRegisterGalaSlackListeners(t *testing.T) {
	registry := gala.NewRegistry()

	ids, err := RegisterGalaSlackListeners(registry)
	require.NoError(t, err)
	require.Len(t, ids, 2)
	require.True(t, registry.InterestedIn(gala.TopicName(entgen.TypeSubscriber), ent.OpCreate.String()))
	require.True(t, registry.InterestedIn(gala.TopicName(entgen.TypeUser), ent.OpCreate.String()))
	require.False(t, registry.InterestedIn(gala.TopicName(entgen.TypeSubscriber), ent.OpUpdate.String()))
	require.False(t, registry.InterestedIn(gala.TopicName(entgen.TypeUser), ent.OpUpdate.String()))
}

// TestHandleUserMutationGalaSendsSlack verifies create mutations post user template Slack messages.
func TestHandleUserMutationGalaSendsSlack(t *testing.T) {
	recorder := newSlackWebhookRecorder(t)
	defer recorder.Close()

	setSlackConfigForTest(t, SlackConfig{WebhookURL: recorder.URL()})

	handlerContext := gala.HandlerContext{
		Context: context.Background(),
		Envelope: gala.Envelope{
			Headers: gala.Headers{Properties: map[string]string{}},
		},
	}

	payload := eventqueue.MutationGalaPayload{
		Operation: ent.OpCreate.String(),
		ProposedChanges: map[string]any{
			user.FieldEmail: "new.user@example.com",
		},
	}

	require.NoError(t, handleUserMutationGala(handlerContext, payload))
	require.Len(t, recorder.Bodies(), 1)
	assert.Contains(t, recorder.Bodies()[0], "New user registered: new.user@example.com")
}

// TestHandleSubscriberMutationGalaHeaderFallback verifies email fallback to envelope properties.
func TestHandleSubscriberMutationGalaHeaderFallback(t *testing.T) {
	recorder := newSlackWebhookRecorder(t)
	defer recorder.Close()

	setSlackConfigForTest(t, SlackConfig{WebhookURL: recorder.URL()})

	handlerContext := gala.HandlerContext{
		Context: context.Background(),
		Envelope: gala.Envelope{
			Headers: gala.Headers{
				Properties: map[string]string{
					subscriber.FieldEmail: "fallback.subscriber@example.com",
				},
			},
		},
	}

	payload := eventqueue.MutationGalaPayload{
		Operation: ent.OpCreate.String(),
	}

	require.NoError(t, handleSubscriberMutationGala(handlerContext, payload))
	require.Len(t, recorder.Bodies(), 1)
	assert.Contains(t, recorder.Bodies()[0], "New waitlist subscriber: fallback.subscriber@example.com")
}

// TestHandleUserMutationGalaUsesTemplateOverride verifies custom template file override works.
func TestHandleUserMutationGalaUsesTemplateOverride(t *testing.T) {
	recorder := newSlackWebhookRecorder(t)
	defer recorder.Close()

	customTemplatePath := writeSlackTemplateFile(t, "Custom user: {{.Email}}")
	setSlackConfigForTest(t, SlackConfig{
		WebhookURL:         recorder.URL(),
		NewUserMessageFile: customTemplatePath,
	})

	handlerContext := gala.HandlerContext{
		Context:  context.Background(),
		Envelope: gala.Envelope{Headers: gala.Headers{Properties: map[string]string{}}},
	}

	payload := eventqueue.MutationGalaPayload{
		Operation: ent.OpCreate.String(),
		ProposedChanges: map[string]any{
			user.FieldEmail: "custom.user@example.com",
		},
	}

	require.NoError(t, handleUserMutationGala(handlerContext, payload))
	require.Len(t, recorder.Bodies(), 1)
	assert.Contains(t, recorder.Bodies()[0], "Custom user: custom.user@example.com")
}

// setSlackConfigForTest sets Slack config for a test and restores the previous value.
func setSlackConfigForTest(t *testing.T, cfg SlackConfig) {
	t.Helper()

	previous := slackCfg
	SetSlackConfig(cfg)
	t.Cleanup(func() {
		SetSlackConfig(previous)
	})
}

func writeSlackTemplateFile(t *testing.T, body string) string {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "slack-template-*.tmpl")
	require.NoError(t, err)
	require.NoError(t, file.Close())
	require.NoError(t, os.WriteFile(file.Name(), []byte(body), 0o600))

	return file.Name()
}

// slackWebhookRecorder captures webhook request payloads for assertions.
type slackWebhookRecorder struct {
	server *httptest.Server

	mu     sync.Mutex
	bodies []string
}

// newSlackWebhookRecorder constructs a webhook recorder backed by an httptest server.
func newSlackWebhookRecorder(t *testing.T) *slackWebhookRecorder {
	t.Helper()

	recorder := &slackWebhookRecorder{}
	recorder.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body failed", http.StatusInternalServerError)
			return
		}

		recorder.mu.Lock()
		recorder.bodies = append(recorder.bodies, string(body))
		recorder.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))

	return recorder
}

// URL returns the recorder server URL.
func (r *slackWebhookRecorder) URL() string {
	if r == nil || r.server == nil {
		return ""
	}

	return r.server.URL
}

// Bodies returns captured webhook request bodies.
func (r *slackWebhookRecorder) Bodies() []string {
	if r == nil {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]string(nil), r.bodies...)
}

// Close shuts down the underlying httptest server.
func (r *slackWebhookRecorder) Close() {
	if r == nil || r.server == nil {
		return
	}

	r.server.Close()
}

// TestMutationEmailFromGala verifies proposed-change precedence with property fallback.
func TestMutationEmailFromGala(t *testing.T) {
	payload := eventqueue.MutationGalaPayload{
		ProposedChanges: map[string]any{
			user.FieldEmail: "proposed@example.com",
		},
	}

	resolved := mutationEmailFromGala(payload, map[string]string{
		user.FieldEmail: "header@example.com",
	}, user.FieldEmail)
	assert.Equal(t, "proposed@example.com", resolved)

	resolved = mutationEmailFromGala(eventqueue.MutationGalaPayload{}, map[string]string{
		user.FieldEmail: "header@example.com",
	}, user.FieldEmail)
	assert.Equal(t, "header@example.com", resolved)

	resolved = mutationEmailFromGala(eventqueue.MutationGalaPayload{
		ProposedChanges: map[string]any{
			user.FieldEmail: strings.Repeat(" ", 3),
		},
	}, map[string]string{
		user.FieldEmail: "header@example.com",
	}, user.FieldEmail)
	assert.Empty(t, resolved)
}
