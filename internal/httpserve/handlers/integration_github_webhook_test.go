package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/utils/rout"
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

	err := h.GitHubIntegrationWebhookHandler(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var reply rout.Reply
	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&reply))
	assert.False(t, reply.Success)
	assert.Equal(t, rout.MissingField(githubWebhookEventHeader).Error(), reply.Error)
}
