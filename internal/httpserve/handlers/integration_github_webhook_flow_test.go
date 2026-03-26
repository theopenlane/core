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
	"github.com/theopenlane/core/internal/ent/generated/vulnerability"
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
			integrationwebhook.ProviderEQ(githubAppDefinitionID),
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

func (suite *HandlerTestSuite) TestGitHubWebhookMultiOrgInstallationRoutesToCorrectIntegration() {
	t := suite.T()

	restore := suite.withGitHubAppIntegrationRuntime(t, defaultGitHubAppSpec())
	defer restore()

	suite.registerGitHubAppWebhookRoute()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	// Create two integrations in the same Openlane org, each representing a different GitHub org installation
	installAttrsOrgA, _ := json.Marshal(githubapp.InstallationMetadata{InstallationID: "7001"})
	integrationA, err := suite.db.Integration.Create().
		SetOwnerID(user.OrganizationID).
		SetName("GitHub App - Org A").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Attributes: installAttrsOrgA}).
		SetDefinitionID(githubAppDefinitionID).
		Save(user.UserCtx)
	require.NoError(t, err)

	installAttrsOrgB, _ := json.Marshal(githubapp.InstallationMetadata{InstallationID: "7002"})
	integrationB, err := suite.db.Integration.Create().
		SetOwnerID(user.OrganizationID).
		SetName("GitHub App - Org B").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Attributes: installAttrsOrgB}).
		SetDefinitionID(githubAppDefinitionID).
		Save(user.UserCtx)
	require.NoError(t, err)

	// Send a ping webhook for installation 7002 (Org B) — should resolve to integrationB
	payloadB := []byte(`{"zen":"keep it logically awesome","installation":{"id":7002}}`)
	reqB := httptest.NewRequest(http.MethodPost, githubAppWebhookPath, strings.NewReader(string(payloadB)))
	reqB.Header.Set("X-GitHub-Event", "ping")
	reqB.Header.Set("X-Hub-Signature-256", githubWebhookSignature("secret", payloadB))
	reqB = reqB.WithContext(privacy.DecisionContext(reqB.Context(), privacy.Allow))
	recB := httptest.NewRecorder()

	suite.e.ServeHTTP(recB, reqB)
	require.Equal(t, http.StatusOK, recB.Code)

	suite.h.IntegrationsRuntime.Gala().WaitIdle()

	// Verify integrationB was updated with webhook verification metadata
	updatedB, err := suite.db.Integration.Get(user.UserCtx, integrationB.ID)
	require.NoError(t, err)
	_, hasBVerified := updatedB.Metadata["githubWebhookVerifiedAt"]
	assert.True(t, hasBVerified, "integration B should have verification metadata")

	// Verify integrationA was NOT updated
	updatedA, err := suite.db.Integration.Get(user.UserCtx, integrationA.ID)
	require.NoError(t, err)
	_, hasAVerified := updatedA.Metadata["githubWebhookVerifiedAt"]
	assert.False(t, hasAVerified, "integration A should not have verification metadata")

	// Now send a ping webhook for installation 7001 (Org A) — should resolve to integrationA
	payloadA := []byte(`{"zen":"keep it logically awesome","installation":{"id":7001}}`)
	reqA := httptest.NewRequest(http.MethodPost, githubAppWebhookPath, strings.NewReader(string(payloadA)))
	reqA.Header.Set("X-GitHub-Event", "ping")
	reqA.Header.Set("X-Hub-Signature-256", githubWebhookSignature("secret", payloadA))
	reqA = reqA.WithContext(privacy.DecisionContext(reqA.Context(), privacy.Allow))
	recA := httptest.NewRecorder()

	suite.e.ServeHTTP(recA, reqA)
	require.Equal(t, http.StatusOK, recA.Code)

	suite.h.IntegrationsRuntime.Gala().WaitIdle()

	updatedA, err = suite.db.Integration.Get(user.UserCtx, integrationA.ID)
	require.NoError(t, err)
	_, hasAVerified = updatedA.Metadata["githubWebhookVerifiedAt"]
	assert.True(t, hasAVerified, "integration A should now have verification metadata")
}

func (suite *HandlerTestSuite) TestGitHubWebhookDependabotAlertIngestsVulnerability() {
	t := suite.T()

	restore := suite.withDurableGitHubAppIntegrationRuntime(t, defaultGitHubAppSpec())
	defer restore()

	suite.registerGitHubAppWebhookRoute()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	installAttrs, _ := json.Marshal(githubapp.InstallationMetadata{InstallationID: "8001"})
	_, err := suite.db.Integration.Create().
		SetOwnerID(user.OrganizationID).
		SetName("GitHub App").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Attributes: installAttrs}).
		SetDefinitionID(githubAppDefinitionID).
		Save(user.UserCtx)
	require.NoError(t, err)

	payload := []byte(`{
		"action": "created",
		"installation": {"id": 8001},
		"repository": {"full_name": "acme/web-app"},
		"alert": {
			"number": 42,
			"state": "OPEN",
			"html_url": "https://github.com/acme/web-app/security/dependabot/42",
			"created_at": "2026-03-01T00:00:00Z",
			"updated_at": "2026-03-25T12:00:00Z",
			"security_advisory": {
				"ghsa_id": "GHSA-1234-5678-abcd",
				"cve_id": "CVE-2026-12345",
				"severity": "high",
				"summary": "Prototype Pollution in lodash",
				"description": "Versions of lodash before 4.17.21 are vulnerable to prototype pollution."
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, githubAppWebhookPath, strings.NewReader(string(payload)))
	req.Header.Set("X-GitHub-Event", "dependabot_alert")
	req.Header.Set("X-GitHub-Delivery", "delivery-vuln-001")
	req.Header.Set("X-Hub-Signature-256", githubWebhookSignature("secret", payload))
	req = req.WithContext(privacy.DecisionContext(req.Context(), privacy.Allow))

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	suite.h.IntegrationsRuntime.Gala().WaitIdle()

	// Verify the vulnerability record was persisted with correct field values
	vulns, err := suite.db.Vulnerability.Query().
		Where(
			vulnerability.OwnerID(user.OrganizationID),
			vulnerability.ExternalID("github:acme/web-app:dependabot:42"),
		).
		All(user.UserCtx)
	require.NoError(t, err)
	require.Len(t, vulns, 1, "expected exactly one vulnerability record")

	vuln := vulns[0]
	assert.Equal(t, "github", vuln.Source)
	assert.Equal(t, "dependabot", vuln.Category)
	assert.Equal(t, "high", vuln.Severity)
	assert.Equal(t, "Prototype Pollution in lodash", vuln.Summary)
	assert.Equal(t, "Versions of lodash before 4.17.21 are vulnerable to prototype pollution.", vuln.Description)
	assert.Equal(t, "CVE-2026-12345", vuln.CveID)
	assert.Equal(t, "acme/web-app", vuln.ExternalOwnerID)
	assert.Equal(t, "https://github.com/acme/web-app/security/dependabot/42", vuln.ExternalURI)
}

func (suite *HandlerTestSuite) TestGitHubWebhookDependabotAlertUpsertsExistingVulnerability() {
	t := suite.T()

	restore := suite.withDurableGitHubAppIntegrationRuntime(t, defaultGitHubAppSpec())
	defer restore()

	suite.registerGitHubAppWebhookRoute()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	installAttrs, _ := json.Marshal(githubapp.InstallationMetadata{InstallationID: "8002"})
	_, err := suite.db.Integration.Create().
		SetOwnerID(user.OrganizationID).
		SetName("GitHub App").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Attributes: installAttrs}).
		SetDefinitionID(githubAppDefinitionID).
		Save(user.UserCtx)
	require.NoError(t, err)

	// First delivery: alert in OPEN state
	payloadOpen := []byte(`{
		"action": "created",
		"installation": {"id": 8002},
		"repository": {"full_name": "acme/api"},
		"alert": {
			"number": 99,
			"state": "OPEN",
			"html_url": "https://github.com/acme/api/security/dependabot/99",
			"created_at": "2026-03-01T00:00:00Z",
			"updated_at": "2026-03-20T00:00:00Z",
			"security_advisory": {
				"ghsa_id": "GHSA-9999-0000-zzzz",
				"cve_id": "CVE-2026-99999",
				"severity": "critical",
				"summary": "Remote code execution in parser",
				"description": "A critical RCE vulnerability in the YAML parser."
			}
		}
	}`)

	reqOpen := httptest.NewRequest(http.MethodPost, githubAppWebhookPath, strings.NewReader(string(payloadOpen)))
	reqOpen.Header.Set("X-GitHub-Event", "dependabot_alert")
	reqOpen.Header.Set("X-GitHub-Delivery", "delivery-upsert-001")
	reqOpen.Header.Set("X-Hub-Signature-256", githubWebhookSignature("secret", payloadOpen))
	reqOpen = reqOpen.WithContext(privacy.DecisionContext(reqOpen.Context(), privacy.Allow))

	recOpen := httptest.NewRecorder()
	suite.e.ServeHTTP(recOpen, reqOpen)
	require.Equal(t, http.StatusOK, recOpen.Code)

	suite.h.IntegrationsRuntime.Gala().WaitIdle()

	// Verify creation
	vulns, err := suite.db.Vulnerability.Query().
		Where(
			vulnerability.OwnerID(user.OrganizationID),
			vulnerability.ExternalID("github:acme/api:dependabot:99"),
		).
		All(user.UserCtx)
	require.NoError(t, err)
	require.Len(t, vulns, 1)
	assert.Equal(t, "critical", vulns[0].Severity)

	originalID := vulns[0].ID

	// Second delivery: same alert now fixed/dismissed
	payloadFixed := []byte(`{
		"action": "fixed",
		"installation": {"id": 8002},
		"repository": {"full_name": "acme/api"},
		"alert": {
			"number": 99,
			"state": "fixed",
			"html_url": "https://github.com/acme/api/security/dependabot/99",
			"created_at": "2026-03-01T00:00:00Z",
			"updated_at": "2026-03-26T00:00:00Z",
			"security_advisory": {
				"ghsa_id": "GHSA-9999-0000-zzzz",
				"cve_id": "CVE-2026-99999",
				"severity": "critical",
				"summary": "Remote code execution in parser",
				"description": "A critical RCE vulnerability in the YAML parser."
			}
		}
	}`)

	reqFixed := httptest.NewRequest(http.MethodPost, githubAppWebhookPath, strings.NewReader(string(payloadFixed)))
	reqFixed.Header.Set("X-GitHub-Event", "dependabot_alert")
	reqFixed.Header.Set("X-GitHub-Delivery", "delivery-upsert-002")
	reqFixed.Header.Set("X-Hub-Signature-256", githubWebhookSignature("secret", payloadFixed))
	reqFixed = reqFixed.WithContext(privacy.DecisionContext(reqFixed.Context(), privacy.Allow))

	recFixed := httptest.NewRecorder()
	suite.e.ServeHTTP(recFixed, reqFixed)
	require.Equal(t, http.StatusOK, recFixed.Code)

	suite.h.IntegrationsRuntime.Gala().WaitIdle()

	// Verify the update hit the same record, not a new one
	vulns, err = suite.db.Vulnerability.Query().
		Where(
			vulnerability.OwnerID(user.OrganizationID),
			vulnerability.ExternalID("github:acme/api:dependabot:99"),
		).
		All(user.UserCtx)
	require.NoError(t, err)
	require.Len(t, vulns, 1, "should still be one record after upsert")
	assert.Equal(t, originalID, vulns[0].ID, "record ID should be unchanged after upsert")
	assert.Equal(t, "fixed", vulns[0].Status, "status should be updated to fixed")
}

func (suite *HandlerTestSuite) TestGitHubWebhookMultiOrgInstallationIngestsVulnerabilitiesToSameOrg() {
	t := suite.T()

	restore := suite.withDurableGitHubAppIntegrationRuntime(t, defaultGitHubAppSpec())
	defer restore()

	suite.registerGitHubAppWebhookRoute()

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	// Two GitHub org installations under the same Openlane org
	installAttrsOrgA, _ := json.Marshal(githubapp.InstallationMetadata{InstallationID: "9001"})
	_, err := suite.db.Integration.Create().
		SetOwnerID(user.OrganizationID).
		SetName("GitHub App - Org Alpha").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Attributes: installAttrsOrgA}).
		SetDefinitionID(githubAppDefinitionID).
		Save(user.UserCtx)
	require.NoError(t, err)

	installAttrsOrgB, _ := json.Marshal(githubapp.InstallationMetadata{InstallationID: "9002"})
	_, err = suite.db.Integration.Create().
		SetOwnerID(user.OrganizationID).
		SetName("GitHub App - Org Beta").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Attributes: installAttrsOrgB}).
		SetDefinitionID(githubAppDefinitionID).
		Save(user.UserCtx)
	require.NoError(t, err)

	// Send dependabot alert from Org Alpha (installation 9001)
	payloadA := []byte(`{
		"action": "created",
		"installation": {"id": 9001},
		"repository": {"full_name": "alpha-org/service"},
		"alert": {
			"number": 10,
			"state": "OPEN",
			"html_url": "https://github.com/alpha-org/service/security/dependabot/10",
			"created_at": "2026-03-01T00:00:00Z",
			"updated_at": "2026-03-25T00:00:00Z",
			"security_advisory": {
				"ghsa_id": "GHSA-aaaa-bbbb-cccc",
				"cve_id": "CVE-2026-11111",
				"severity": "medium",
				"summary": "XSS in template engine",
				"description": "Template engine fails to escape user input."
			}
		}
	}`)

	reqA := httptest.NewRequest(http.MethodPost, githubAppWebhookPath, strings.NewReader(string(payloadA)))
	reqA.Header.Set("X-GitHub-Event", "dependabot_alert")
	reqA.Header.Set("X-GitHub-Delivery", "delivery-multiorg-001")
	reqA.Header.Set("X-Hub-Signature-256", githubWebhookSignature("secret", payloadA))
	reqA = reqA.WithContext(privacy.DecisionContext(reqA.Context(), privacy.Allow))

	recA := httptest.NewRecorder()
	suite.e.ServeHTTP(recA, reqA)
	require.Equal(t, http.StatusOK, recA.Code)

	// Send dependabot alert from Org Beta (installation 9002)
	payloadB := []byte(`{
		"action": "created",
		"installation": {"id": 9002},
		"repository": {"full_name": "beta-org/platform"},
		"alert": {
			"number": 5,
			"state": "OPEN",
			"html_url": "https://github.com/beta-org/platform/security/dependabot/5",
			"created_at": "2026-03-02T00:00:00Z",
			"updated_at": "2026-03-25T00:00:00Z",
			"security_advisory": {
				"ghsa_id": "GHSA-dddd-eeee-ffff",
				"cve_id": "CVE-2026-22222",
				"severity": "low",
				"summary": "Information disclosure via error messages",
				"description": "Verbose error messages leak internal paths."
			}
		}
	}`)

	reqB := httptest.NewRequest(http.MethodPost, githubAppWebhookPath, strings.NewReader(string(payloadB)))
	reqB.Header.Set("X-GitHub-Event", "dependabot_alert")
	reqB.Header.Set("X-GitHub-Delivery", "delivery-multiorg-002")
	reqB.Header.Set("X-Hub-Signature-256", githubWebhookSignature("secret", payloadB))
	reqB = reqB.WithContext(privacy.DecisionContext(reqB.Context(), privacy.Allow))

	recB := httptest.NewRecorder()
	suite.e.ServeHTTP(recB, reqB)
	require.Equal(t, http.StatusOK, recB.Code)

	suite.h.IntegrationsRuntime.Gala().WaitIdle()

	// Both vulnerabilities should land in the same Openlane org
	vulnA, err := suite.db.Vulnerability.Query().
		Where(
			vulnerability.OwnerID(user.OrganizationID),
			vulnerability.ExternalID("github:alpha-org/service:dependabot:10"),
		).
		All(user.UserCtx)
	require.NoError(t, err)
	require.Len(t, vulnA, 1)
	assert.Equal(t, "medium", vulnA[0].Severity)
	assert.Equal(t, "alpha-org/service", vulnA[0].ExternalOwnerID)
	assert.Equal(t, "CVE-2026-11111", vulnA[0].CveID)

	vulnB, err := suite.db.Vulnerability.Query().
		Where(
			vulnerability.OwnerID(user.OrganizationID),
			vulnerability.ExternalID("github:beta-org/platform:dependabot:5"),
		).
		All(user.UserCtx)
	require.NoError(t, err)
	require.Len(t, vulnB, 1)
	assert.Equal(t, "low", vulnB[0].Severity)
	assert.Equal(t, "beta-org/platform", vulnB[0].ExternalOwnerID)
	assert.Equal(t, "CVE-2026-22222", vulnB[0].CveID)
}

func githubWebhookSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
