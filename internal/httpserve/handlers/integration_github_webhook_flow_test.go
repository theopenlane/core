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

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/integrationwebhook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/definitions/githubapp"
	"github.com/theopenlane/core/internal/slacknotify"
	"github.com/theopenlane/core/pkg/slacktemplates"
)

const githubAppWebhookPath = "/github/app/webhook"

func defaultGitHubAppSpec() githubapp.Config {
	return githubapp.Config{
		AppID:         "123",
		AppSlug:       "openlane",
		PrivateKey:    "private-key",
		WebhookSecret: "secret",
	}
}

func (suite *HandlerTestSuite) registerGitHubAppWebhookRoute() {
	op := suite.createImpersonationOperation("GitHubAppWebhook", "Handle GitHub App security alert webhooks")
	suite.registerRouteOnce(http.MethodPost, githubAppWebhookPath, op, suite.h.GitHubAppWebhookHandler)
}

func (suite *HandlerTestSuite) TestGitHubAppWebhookDoesNotRequireCaller() {
	t := suite.T()

	restore := suite.withGitHubAppIntegrationRuntime(t, defaultGitHubAppSpec())
	defer restore()

	suite.registerGitHubAppWebhookRoute()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	installAttrs, _ := json.Marshal(githubapp.InstallationMetadata{InstallationID: "456"})
	_, err := suite.db.Integration.Create().
		SetOwnerID(user.OrganizationID).
		SetName("GitHub App").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Attributes: installAttrs}).
		SetDefinitionID(githubAppDefinitionID).
		SetDefinitionSlug(githubAppSlug).
		Save(user.UserCtx)
	require.NoError(t, err)

	payload := []byte(`{"zen":"keep it logically awesome","installation":{"id":456}}`)
	req := httptest.NewRequest(http.MethodPost, githubAppWebhookPath, strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", githubWebhookSignature("secret", payload))
	req = req.WithContext(privacy.DecisionContext(req.Context(), privacy.Allow))
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func (suite *HandlerTestSuite) TestGitHubWebhookPingUpdatesIntegrationMetadata() {
	t := suite.T()

	restore := suite.withGitHubAppIntegrationRuntime(t, defaultGitHubAppSpec())
	defer restore()

	suite.registerGitHubAppWebhookRoute()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	installAttrs, _ := json.Marshal(githubapp.InstallationMetadata{InstallationID: "1001"})
	integrationRecord, err := suite.db.Integration.Create().
		SetOwnerID(user.OrganizationID).
		SetName("GitHub App").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Attributes: installAttrs}).
		SetDefinitionID(githubAppDefinitionID).
		SetDefinitionSlug(githubAppSlug).
		Save(user.UserCtx)
	require.NoError(t, err)

	payload := []byte(`{"zen":"keep it logically awesome","installation":{"id":1001}}`)
	req := httptest.NewRequest(http.MethodPost, githubAppWebhookPath, strings.NewReader(string(payload)))
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", githubWebhookSignature("secret", payload))
	req = req.WithContext(user.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	// Wait for in-memory Gala pool to finish processing the dispatched webhook event
	suite.h.IntegrationsRuntime.Gala().WaitIdle()

	updated, err := suite.db.Integration.Get(user.UserCtx, integrationRecord.ID)
	require.NoError(t, err)
	verifiedAtValue, ok := updated.Metadata["githubWebhookVerifiedAt"]
	require.True(t, ok)
	verifiedAtString, ok := verifiedAtValue.(string)
	require.True(t, ok)
	require.NotEmpty(t, verifiedAtString)
}

func (suite *HandlerTestSuite) TestGitHubWebhookPingRejectsInvalidSignature() {
	t := suite.T()

	restore := suite.withGitHubAppIntegrationRuntime(t, defaultGitHubAppSpec())
	defer restore()

	suite.registerGitHubAppWebhookRoute()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	installAttrs, _ := json.Marshal(githubapp.InstallationMetadata{InstallationID: "1004"})
	integrationRecord, err := suite.db.Integration.Create().
		SetOwnerID(user.OrganizationID).
		SetName("GitHub App").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Attributes: installAttrs}).
		SetDefinitionID(githubAppDefinitionID).
		SetDefinitionSlug(githubAppSlug).
		Save(user.UserCtx)
	require.NoError(t, err)

	payload := []byte(`{"zen":"keep it logically awesome","installation":{"id":1004}}`)
	req := httptest.NewRequest(http.MethodPost, githubAppWebhookPath, strings.NewReader(string(payload)))
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", githubWebhookSignature("wrong-secret", payload))
	req = req.WithContext(user.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	updated, err := suite.db.Integration.Get(user.UserCtx, integrationRecord.ID)
	require.NoError(t, err)

	_, hasMetadata := updated.Metadata["githubWebhookVerifiedAt"]
	assert.False(t, hasMetadata)
}

func (suite *HandlerTestSuite) TestGitHubWebhookInstallationCreatedSendsTemplatedSlackNotification() {
	t := suite.T()

	restore := suite.withGitHubAppIntegrationRuntime(t, defaultGitHubAppSpec())
	defer restore()

	suite.registerGitHubAppWebhookRoute()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	err := suite.db.Organization.UpdateOneID(user.OrganizationID).
		SetDisplayName("Acme Security").
		Exec(user.UserCtx)
	require.NoError(t, err)

	installAttrs, _ := json.Marshal(githubapp.InstallationMetadata{InstallationID: "1002"})
	_, err = suite.db.Integration.Create().
		SetOwnerID(user.OrganizationID).
		SetName("GitHub App").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Attributes: installAttrs}).
		SetDefinitionID(githubAppDefinitionID).
		SetDefinitionSlug(githubAppSlug).
		Save(user.UserCtx)
	require.NoError(t, err)

	recorder := newSlackWebhookRecorder(t)
	defer recorder.Close()

	slacknotify.SetConfig(slacknotify.SlackConfig{WebhookURL: recorder.URL()})
	t.Cleanup(func() {
		slacknotify.SetConfig(slacknotify.SlackConfig{})
	})

	payload := []byte(`{"action":"created","installation":{"id":1002,"account":{"login":"acme-github-org","type":"Organization"}}}`)
	req := httptest.NewRequest(http.MethodPost, githubAppWebhookPath, strings.NewReader(string(payload)))
	req.Header.Set("X-GitHub-Event", "installation")
	req.Header.Set("X-Hub-Signature-256", githubWebhookSignature("secret", payload))
	req = req.WithContext(user.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	// Wait for in-memory Gala pool to finish processing the dispatched webhook event
	suite.h.IntegrationsRuntime.Gala().WaitIdle()

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

func (suite *HandlerTestSuite) TestGitHubWebhookDuplicateDeliveryIsIgnored() {
	t := suite.T()

	restore := suite.withGitHubAppIntegrationRuntime(t, defaultGitHubAppSpec())
	defer restore()

	suite.registerGitHubAppWebhookRoute()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	installAttrs, _ := json.Marshal(githubapp.InstallationMetadata{InstallationID: "1003"})
	_, err := suite.db.Integration.Create().
		SetOwnerID(user.OrganizationID).
		SetName("GitHub App").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Attributes: installAttrs}).
		SetDefinitionID(githubAppDefinitionID).
		SetDefinitionSlug(githubAppSlug).
		Save(user.UserCtx)
	require.NoError(t, err)

	payload := []byte(`{"action":"created","installation":{"id":1003},"repository":{"full_name":"acme/repo"},"alert":{"number":1}}`)
	deliveryID := "delivery-dup-1"

	firstReq := httptest.NewRequest(http.MethodPost, githubAppWebhookPath, strings.NewReader(string(payload)))
	firstReq.Header.Set("X-GitHub-Event", "dependabot_alert")
	firstReq.Header.Set("X-Hub-Signature-256", githubWebhookSignature("secret", payload))
	firstReq.Header.Set("X-GitHub-Delivery", deliveryID)
	firstReq = firstReq.WithContext(user.UserCtx)

	firstRec := httptest.NewRecorder()
	suite.e.ServeHTTP(firstRec, firstReq)
	require.Equal(t, http.StatusOK, firstRec.Code)

	secondReq := httptest.NewRequest(http.MethodPost, githubAppWebhookPath, strings.NewReader(string(payload)))
	secondReq.Header.Set("X-GitHub-Event", "dependabot_alert")
	secondReq.Header.Set("X-Hub-Signature-256", githubWebhookSignature("secret", payload))
	secondReq.Header.Set("X-GitHub-Delivery", deliveryID)
	secondReq = secondReq.WithContext(user.UserCtx)

	secondRec := httptest.NewRecorder()
	suite.e.ServeHTTP(secondRec, secondReq)
	require.Equal(t, http.StatusOK, secondRec.Code)

	dedupeCount, err := suite.db.IntegrationWebhook.Query().
		Where(
			integrationwebhook.OwnerIDEQ(user.OrganizationID),
			integrationwebhook.ProviderEQ(githubAppSlug),
			integrationwebhook.ExternalEventIDEQ(deliveryID),
		).
		Count(user.UserCtx)
	require.NoError(t, err)
	require.Equal(t, 1, dedupeCount)
}

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

func githubWebhookSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
