//go:build test

package handlers_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/integrationwebhook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/integrations/definitions/githubapp"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/slacktemplates"
)

// defaultGitHubAppSpec returns a fully-configured GitHub App provider spec for flow tests.
func defaultGitHubAppSpec() githubapp.Config {
	return githubapp.Config{
		AppID:         "123",
		AppSlug:       "openlane",
		PrivateKey:    "private-key",
		WebhookSecret: "secret",
	}
}

// TestGitHubWebhookPingUpdatesIntegrationMetadata verifies ping webhook handling updates integration metadata for UI visibility
func (suite *HandlerTestSuite) TestGitHubWebhookPingUpdatesIntegrationMetadata() {
	t := suite.T()

	restore := suite.withGitHubAppIntegrationRuntime(t, defaultGitHubAppSpec(), "")
	defer restore()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	integrationRecord, err := suite.db.Integration.Create().
		SetOwnerID(user.OrganizationID).
		SetName("GitHub App").
		SetDefinitionID(githubapp.DefinitionID).
		SetDefinitionSlug(githubapp.DefinitionSlug).
		SetProviderState(func() types.IntegrationProviderState {
			doc := types.IntegrationProviderState{}
			_, mergeErr := doc.MergeProviderData(githubapp.DefinitionSlug, []byte(`{"appId":"123","installationId":"456"}`))
			require.NoError(t, mergeErr)
			return doc
		}()).
		Save(user.UserCtx)
	require.NoError(t, err)

	payload := []byte(`{"zen":"keep it logically awesome","installation":{"id":456}}`)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(payload)))
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", githubWebhookSignature("secret", payload))
	req = req.WithContext(user.UserCtx)

	rec := httptest.NewRecorder()
	ctx := suite.e.NewContext(req, rec)

	err = suite.h.GitHubIntegrationWebhookHandler(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	updated, err := suite.db.Integration.Get(user.UserCtx, integrationRecord.ID)
	require.NoError(t, err)
	providerState, err := jsonx.ToMap(updated.ProviderState.ProviderData(githubapp.DefinitionSlug))
	require.NoError(t, err)
	require.NotNil(t, providerState)
	webhookVerifiedAt, ok := providerState["webhookVerifiedAt"].(string)
	require.True(t, ok)
	require.NotEmpty(t, webhookVerifiedAt)

	verifiedAtValue, ok := updated.Metadata["githubWebhookVerifiedAt"]
	require.True(t, ok)
	verifiedAtString, ok := verifiedAtValue.(string)
	require.True(t, ok)
	require.NotEmpty(t, verifiedAtString)
}

// TestGitHubWebhookPingRejectsInvalidSignature verifies invalid signatures are rejected before integration lookup/update.
func (suite *HandlerTestSuite) TestGitHubWebhookPingRejectsInvalidSignature() {
	t := suite.T()

	restore := suite.withGitHubAppIntegrationRuntime(t, defaultGitHubAppSpec(), "")
	defer restore()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	integrationRecord, err := suite.db.Integration.Create().
		SetOwnerID(user.OrganizationID).
		SetName("GitHub App").
		SetDefinitionID(githubapp.DefinitionID).
		SetDefinitionSlug(githubapp.DefinitionSlug).
		SetProviderState(func() types.IntegrationProviderState {
			doc := types.IntegrationProviderState{}
			_, mergeErr := doc.MergeProviderData(githubapp.DefinitionSlug, []byte(`{"appId":"123","installationId":"456"}`))
			require.NoError(t, mergeErr)
			return doc
		}()).
		Save(user.UserCtx)
	require.NoError(t, err)

	payload := []byte(`{"zen":"keep it logically awesome","installation":{"id":456}}`)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(payload)))
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", githubWebhookSignature("wrong-secret", payload))
	req = req.WithContext(user.UserCtx)

	rec := httptest.NewRecorder()
	ctx := suite.e.NewContext(req, rec)

	err = suite.h.GitHubIntegrationWebhookHandler(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	updated, err := suite.db.Integration.Get(user.UserCtx, integrationRecord.ID)
	require.NoError(t, err)

	providerState, err := jsonx.ToMap(updated.ProviderState.ProviderData(githubapp.DefinitionSlug))
	require.NoError(t, err)
	_, hasVerifiedAt := providerState["webhookVerifiedAt"]
	assert.False(t, hasVerifiedAt)

	_, hasMetadata := updated.Metadata["githubWebhookVerifiedAt"]
	assert.False(t, hasMetadata)
}

// TestGitHubWebhookPingUnknownInstallationAccepted verifies signed webhooks with unknown installations are safely ignored.
func (suite *HandlerTestSuite) TestGitHubWebhookPingUnknownInstallationAccepted() {
	t := suite.T()

	restore := suite.withGitHubAppIntegrationRuntime(t, defaultGitHubAppSpec(), "")
	defer restore()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	payload := []byte(`{"zen":"keep it logically awesome","installation":{"id":999999}}`)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(payload)))
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", githubWebhookSignature("secret", payload))
	req = req.WithContext(user.UserCtx)

	rec := httptest.NewRecorder()
	ctx := suite.e.NewContext(req, rec)

	err := suite.h.GitHubIntegrationWebhookHandler(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	count, err := suite.db.Integration.Query().
		Where(
			integration.OwnerIDEQ(user.OrganizationID),
			integration.DefinitionIDEQ(githubapp.DefinitionID),
		).
		Count(user.UserCtx)
	require.NoError(t, err)
	require.Zero(t, count)
}

// TestGitHubWebhookInstallationCreatedSendsTemplatedSlackNotification verifies installation webhooks look up the integration and send template-rendered Slack messages.
func (suite *HandlerTestSuite) TestGitHubWebhookInstallationCreatedSendsTemplatedSlackNotification() {
	t := suite.T()

	restore := suite.withGitHubAppIntegrationRuntime(t, defaultGitHubAppSpec(), "")
	defer restore()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	err := suite.db.Organization.UpdateOneID(user.OrganizationID).
		SetDisplayName("Acme Security").
		Exec(user.UserCtx)
	require.NoError(t, err)

	_, err = suite.db.Integration.Create().
		SetOwnerID(user.OrganizationID).
		SetName("GitHub App").
		SetDefinitionID(githubapp.DefinitionID).
		SetDefinitionSlug(githubapp.DefinitionSlug).
		SetProviderState(func() types.IntegrationProviderState {
			doc := types.IntegrationProviderState{}
			_, mergeErr := doc.MergeProviderData(githubapp.DefinitionSlug, []byte(`{"appId":"123","installationId":"456"}`))
			require.NoError(t, mergeErr)
			return doc
		}()).
		Save(user.UserCtx)
	require.NoError(t, err)

	recorder := newSlackWebhookRecorder(t)
	defer recorder.Close()

	hooks.SetSlackConfig(hooks.SlackConfig{WebhookURL: recorder.URL()})
	t.Cleanup(func() {
		hooks.SetSlackConfig(hooks.SlackConfig{})
	})

	payload := []byte(`{"action":"created","installation":{"id":456,"account":{"login":"acme-github-org","type":"Organization"}}}`)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(payload)))
	req.Header.Set("X-GitHub-Event", "installation")
	req.Header.Set("X-Hub-Signature-256", githubWebhookSignature("secret", payload))
	req = req.WithContext(user.UserCtx)

	rec := httptest.NewRecorder()
	ctx := suite.e.NewContext(req, rec)

	err = suite.h.GitHubIntegrationWebhookHandler(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	bodies := recorder.Bodies()
	require.Len(t, bodies, 1)

	text := slackMessageText(t, bodies[0])
	expected := renderGitHubAppInstallTemplate(t, map[string]any{
		"GitHubOrganization":         "acme-github-org",
		"GitHubAccountType":          "Organization",
		"OpenlaneOrganization":       "Acme Security",
		"OpenlaneOrganizationID":     user.OrganizationID,
		"ShowOpenlaneOrganizationID": true,
	})
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(text))
}

// TestGitHubWebhookInstallationCreatedUnknownInstallationDoesNotNotify verifies unknown installations are ignored before notification.
func (suite *HandlerTestSuite) TestGitHubWebhookInstallationCreatedUnknownInstallationDoesNotNotify() {
	t := suite.T()

	restore := suite.withGitHubAppIntegrationRuntime(t, defaultGitHubAppSpec(), "")
	defer restore()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	_, err := suite.db.Integration.Create().
		SetOwnerID(user.OrganizationID).
		SetName("GitHub App").
		SetDefinitionID(githubapp.DefinitionID).
		SetDefinitionSlug(githubapp.DefinitionSlug).
		SetProviderState(func() types.IntegrationProviderState {
			doc := types.IntegrationProviderState{}
			_, mergeErr := doc.MergeProviderData(githubapp.DefinitionSlug, []byte(`{"appId":"123","installationId":"456"}`))
			require.NoError(t, mergeErr)
			return doc
		}()).
		Save(user.UserCtx)
	require.NoError(t, err)

	recorder := newSlackWebhookRecorder(t)
	defer recorder.Close()

	hooks.SetSlackConfig(hooks.SlackConfig{WebhookURL: recorder.URL()})
	t.Cleanup(func() {
		hooks.SetSlackConfig(hooks.SlackConfig{})
	})

	payload := []byte(`{"action":"created","installation":{"id":999999,"account":{"login":"unknown-org","type":"Organization"}}}`)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(payload)))
	req.Header.Set("X-GitHub-Event", "installation")
	req.Header.Set("X-Hub-Signature-256", githubWebhookSignature("secret", payload))
	req = req.WithContext(user.UserCtx)

	rec := httptest.NewRecorder()
	ctx := suite.e.NewContext(req, rec)

	err = suite.h.GitHubIntegrationWebhookHandler(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	assert.Empty(t, recorder.Bodies())
}

// TestGitHubWebhookDuplicateDeliveryIsIgnored verifies delivery ID dedupe is idempotent and race-safe.
func (suite *HandlerTestSuite) TestGitHubWebhookDuplicateDeliveryIsIgnored() {
	t := suite.T()

	restore := suite.withGitHubAppIntegrationRuntime(t, defaultGitHubAppSpec(), "")
	defer restore()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	_, err := suite.db.Integration.Create().
		SetOwnerID(user.OrganizationID).
		SetName("GitHub App").
		SetDefinitionID(githubapp.DefinitionID).
		SetDefinitionSlug(githubapp.DefinitionSlug).
		SetProviderState(func() types.IntegrationProviderState {
			doc := types.IntegrationProviderState{}
			_, mergeErr := doc.MergeProviderData(githubapp.DefinitionSlug, []byte(`{"appId":"123","installationId":"456"}`))
			require.NoError(t, mergeErr)
			return doc
		}()).
		Save(user.UserCtx)
	require.NoError(t, err)

	payload := []byte(`{"action":"created","installation":{"id":456},"repository":{"full_name":"acme/repo"},"alert":{"number":1}}`)
	deliveryID := "delivery-dup-1"

	firstReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(payload)))
	firstReq.Header.Set("X-GitHub-Event", "dependabot_alert")
	firstReq.Header.Set("X-Hub-Signature-256", githubWebhookSignature("secret", payload))
	firstReq.Header.Set("X-GitHub-Delivery", deliveryID)
	firstReq = firstReq.WithContext(user.UserCtx)

	firstRec := httptest.NewRecorder()
	firstCtx := suite.e.NewContext(firstReq, firstRec)
	err = suite.h.GitHubIntegrationWebhookHandler(firstCtx, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, firstRec.Code)

	secondReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(payload)))
	secondReq.Header.Set("X-GitHub-Event", "dependabot_alert")
	secondReq.Header.Set("X-Hub-Signature-256", githubWebhookSignature("secret", payload))
	secondReq.Header.Set("X-GitHub-Delivery", deliveryID)
	secondReq = secondReq.WithContext(user.UserCtx)

	secondRec := httptest.NewRecorder()
	secondCtx := suite.e.NewContext(secondReq, secondRec)
	err = suite.h.GitHubIntegrationWebhookHandler(secondCtx, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, secondRec.Code)

	dedupeCount, err := suite.db.IntegrationWebhook.Query().
		Where(
			integrationwebhook.OwnerIDEQ(user.OrganizationID),
			integrationwebhook.ProviderEQ(githubapp.DefinitionSlug),
			integrationwebhook.ExternalEventIDEQ(deliveryID),
		).
		Count(user.UserCtx)
	require.NoError(t, err)
	require.Equal(t, 1, dedupeCount)
}

// slackWebhookRecorder captures outgoing Slack webhook payloads.
type slackWebhookRecorder struct {
	server *httptest.Server
	mu     sync.Mutex
	bodies []string
}

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

func (r *slackWebhookRecorder) URL() string {
	if r == nil || r.server == nil {
		return ""
	}

	return r.server.URL
}

func (r *slackWebhookRecorder) Bodies() []string {
	if r == nil {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]string(nil), r.bodies...)
}

func (r *slackWebhookRecorder) Close() {
	if r == nil || r.server == nil {
		return
	}

	r.server.Close()
}

func slackMessageText(t *testing.T, requestBody string) string {
	t.Helper()

	var payload struct {
		Text string `json:"text"`
	}
	require.NoError(t, json.Unmarshal([]byte(requestBody), &payload))
	require.NotEmpty(t, payload.Text)

	return payload.Text
}

func renderGitHubAppInstallTemplate(t *testing.T, data map[string]any) string {
	t.Helper()

	tmpl, err := template.ParseFS(slacktemplates.Templates, slacktemplates.GitHubAppInstallName)
	require.NoError(t, err)

	var rendered strings.Builder
	require.NoError(t, tmpl.Execute(&rendered, data))

	return rendered.String()
}

// githubWebhookSignature builds an HMAC-SHA256 GitHub webhook signature
func githubWebhookSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
