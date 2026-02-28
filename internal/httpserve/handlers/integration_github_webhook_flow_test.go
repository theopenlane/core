//go:build test

package handlers_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/integrations/state"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/integrations/providers/github"
)

// TestGitHubWebhookPingUpdatesIntegrationMetadata verifies ping webhook handling updates integration metadata for UI visibility
func (suite *HandlerTestSuite) TestGitHubWebhookPingUpdatesIntegrationMetadata() {
	t := suite.T()

	suite.h.IntegrationGitHubApp = handlers.IntegrationGitHubAppConfig{
		Enabled:       true,
		AppID:         "123",
		AppSlug:       "openlane",
		PrivateKey:    "private-key",
		WebhookSecret: "secret",
	}

	requestCtx := privacy.DecisionContext(httptest.NewRequest(http.MethodGet, "/", nil).Context(), privacy.Allow)
	user := suite.userBuilderWithInput(requestCtx, &userInput{confirmedUser: true})

	integrationRecord, err := suite.db.Integration.Create().
		SetOwnerID(user.OrganizationID).
		SetName("GitHub App").
		SetKind(string(github.TypeGitHubApp)).
		SetProviderState(state.IntegrationProviderState{
			GitHub: &state.GitHubState{
				AppID:          "123",
				InstallationID: "456",
			},
		}).
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
	require.NotNil(t, updated.ProviderState.GitHub)
	require.NotNil(t, updated.ProviderState.GitHub.WebhookVerifiedAt)
	require.WithinDuration(t, time.Now().UTC(), *updated.ProviderState.GitHub.WebhookVerifiedAt, time.Minute)

	verifiedAtValue, ok := updated.Metadata["githubWebhookVerifiedAt"]
	require.True(t, ok)
	verifiedAtString, ok := verifiedAtValue.(string)
	require.True(t, ok)
	require.NotEmpty(t, verifiedAtString)
}

// githubWebhookSignature builds an HMAC-SHA256 GitHub webhook signature
func githubWebhookSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
