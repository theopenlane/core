package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/utils/rout"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
)

// TestGitHubIntegrationWebhookHandlerMissingEventHeader verifies the missing event header response.
func TestGitHubIntegrationWebhookHandlerMissingEventHeader(t *testing.T) {
	h := &Handler{IntegrationGitHubApp: IntegrationGitHubAppConfig{
		Enabled:       true,
		AppID:         "123",
		AppSlug:       "openlane",
		PrivateKey:    "private-key",
		WebhookSecret: "secret",
	}}

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	receivedBefore := testutil.ToFloat64(githubAppWebhookReceivedCounter.WithLabelValues("unknown"))
	responseBefore := testutil.ToFloat64(githubAppWebhookResponseCounter.WithLabelValues("unknown", "400", "missing_event_header"))

	err := h.GitHubIntegrationWebhookHandler(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, receivedBefore+1, testutil.ToFloat64(githubAppWebhookReceivedCounter.WithLabelValues("unknown")))
	assert.Equal(t, responseBefore+1, testutil.ToFloat64(githubAppWebhookResponseCounter.WithLabelValues("unknown", "400", "missing_event_header")))

	var reply rout.Reply
	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&reply))
	assert.False(t, reply.Success)
	assert.Equal(t, rout.MissingField(githubWebhookEventHeader).Error(), reply.Error)
}

// TestGitHubIntegrationWebhookHandlerEmptyPayloadMetrics verifies empty payload metrics and status.
func TestGitHubIntegrationWebhookHandlerEmptyPayloadMetrics(t *testing.T) {
	h := &Handler{IntegrationGitHubApp: IntegrationGitHubAppConfig{
		Enabled:       true,
		AppID:         "123",
		AppSlug:       "openlane",
		PrivateKey:    "private-key",
		WebhookSecret: "secret",
	}}

	req := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
	req.Header.Set(githubWebhookEventHeader, "dependabot_alert")
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	receivedBefore := testutil.ToFloat64(githubAppWebhookReceivedCounter.WithLabelValues("dependabot_alert"))
	responseBefore := testutil.ToFloat64(githubAppWebhookResponseCounter.WithLabelValues("dependabot_alert", "400", "empty_payload"))

	err := h.GitHubIntegrationWebhookHandler(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, receivedBefore+1, testutil.ToFloat64(githubAppWebhookReceivedCounter.WithLabelValues("dependabot_alert")))
	assert.Equal(t, responseBefore+1, testutil.ToFloat64(githubAppWebhookResponseCounter.WithLabelValues("dependabot_alert", "400", "empty_payload")))
}

// TestHandleGitHubInstallationWebhookSendsSlack verifies installation.created webhooks notify Slack.
func TestHandleGitHubInstallationWebhookSendsSlack(t *testing.T) {
	var (
		mu       sync.Mutex
		requests []string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		mu.Lock()
		requests = append(requests, string(body))
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	hooks.SetSlackConfig(hooks.SlackConfig{WebhookURL: server.URL})
	t.Cleanup(func() {
		hooks.SetSlackConfig(hooks.SlackConfig{})
	})

	h := &Handler{}
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	responseBefore := testutil.ToFloat64(githubAppWebhookResponseCounter.WithLabelValues("installation", "200", "installation_notification_sent"))

	err := h.handleGitHubInstallationWebhook(ctx, nil, "installation", githubWebhookEnvelope{
		Action: "created",
		Installation: &githubWebhookInstallation{
			ID: 1234,
			Account: &githubWebhookAccount{
				Login: "acme-github-org",
				Type:  "Organization",
			},
		},
	}, &ent.Integration{OwnerID: "openlane-org-id"})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, responseBefore+1, testutil.ToFloat64(githubAppWebhookResponseCounter.WithLabelValues("installation", "200", "installation_notification_sent")))

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, requests, 1)
	assert.Contains(t, requests[0], "GitHub organization: acme-github-org")
	assert.Contains(t, requests[0], "Openlane organization: openlane-org-id")
}

// TestHandleGitHubInstallationWebhookSkipsWhenSlackDisabled verifies no-op when webhook config is absent.
func TestHandleGitHubInstallationWebhookSkipsWhenSlackDisabled(t *testing.T) {
	hooks.SetSlackConfig(hooks.SlackConfig{})
	t.Cleanup(func() {
		hooks.SetSlackConfig(hooks.SlackConfig{})
	})

	h := &Handler{}
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	responseBefore := testutil.ToFloat64(githubAppWebhookResponseCounter.WithLabelValues("installation", "200", "installation_notification_skipped"))

	err := h.handleGitHubInstallationWebhook(ctx, nil, "installation", githubWebhookEnvelope{
		Action: "created",
		Installation: &githubWebhookInstallation{
			ID: 1234,
		},
	}, &ent.Integration{OwnerID: "openlane-org-id"})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, responseBefore+1, testutil.ToFloat64(githubAppWebhookResponseCounter.WithLabelValues("installation", "200", "installation_notification_skipped")))
}
