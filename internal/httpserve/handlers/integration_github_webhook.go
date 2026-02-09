package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/theopenlane/core/pkg/logx"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/utils/rout"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"

	"github.com/theopenlane/core/common/integrations/types"
	apimodels "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/providers/github"
)

// GitHub webhook header names and limits.
const (
	githubWebhookSignatureHeader = "X-Hub-Signature-256"
	githubWebhookEventHeader     = "X-GitHub-Event"
	githubWebhookDeliveryHeader  = "X-GitHub-Delivery"
	maxGitHubWebhookBodyBytes    = int64(1024 * 1024)
)

// GitHubIntegrationWebhookHandler ingests GitHub App security alert webhooks.
func (h *Handler) GitHubIntegrationWebhookHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	if isRegistrationContext(ctx) {
		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{
			Reply:     rout.Reply{Success: true},
			Persisted: map[string]any{"created": 0, "updated": 0, "skipped": 0},
		}, openapi)
	}

	if err := h.validateGitHubAppConfig(); err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	req := ctx.Request()
	res := ctx.Response()
	payload, err := io.ReadAll(http.MaxBytesReader(res.Writer, req.Body, maxGitHubWebhookBodyBytes))
	if err != nil {
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}
	if len(payload) == 0 {
		return h.BadRequest(ctx, errPayloadEmpty, openapi)
	}

	eventType := strings.TrimSpace(req.Header.Get(githubWebhookEventHeader))
	if eventType == "" {
		return h.BadRequest(ctx, rout.MissingField(githubWebhookEventHeader), openapi)
	}

	var envelope githubWebhookEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	installationID := githubInstallationID(envelope.Installation)
	if installationID == "" {
		logx.FromContext(req.Context()).Info().Str("event", eventType).Msg("github webhook missing installation id")

		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	integrationRecord, err := h.findGitHubAppIntegrationByInstallationID(req.Context(), installationID)
	if err != nil {
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}
	if integrationRecord == nil {
		logx.FromContext(req.Context()).Info().Str("installation_id", installationID).Msg("no integration configured for github app installation")

		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	orgID := integrationRecord.OwnerID
	webhookSecret := strings.TrimSpace(h.IntegrationGitHubApp.WebhookSecret)
	if webhookSecret == "" {
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	if !validateGitHubWebhookSignature(webhookSecret, req.Header.Get(githubWebhookSignatureHeader), payload) {
		return h.BadRequest(ctx, rout.InvalidField(githubWebhookSignatureHeader), openapi)
	}

	allowCtx := privacy.DecisionContext(req.Context(), privacy.Allow)
	if deliveryID := strings.TrimSpace(req.Header.Get(githubWebhookDeliveryHeader)); deliveryID != "" {
		exists, err := h.checkForEventID(allowCtx, deliveryID)
		if err != nil {
			return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
		}

		if exists {
			return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
		}

		if _, err := h.createEvent(allowCtx, ent.CreateEventInput{EventID: &deliveryID}); err != nil {
			return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
		}
	}

	if eventType == "ping" {
		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	repo := githubRepoFromWebhook(envelope.Repository)
	if repo == "" {
		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	alertType, ok := githubAlertTypeFromEvent(eventType)
	if !ok {
		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	result := types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Processed GitHub %s webhook", eventType),
		Details: map[string]any{
			"alerts": []types.AlertEnvelope{
				{
					AlertType: alertType,
					Resource:  repo,
					Action:    envelope.Action,
					Payload:   envelope.Alert,
				},
			},
		},
	}

	persistSummary, err := h.persistIntegrationVulnerabilities(allowCtx, orgID, github.TypeGitHubApp, result)
	if err != nil {
		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	return h.Success(ctx, apimodels.GitHubAppWebhookResponse{
		Reply:     rout.Reply{Success: true},
		Persisted: persistSummary,
	}, openapi)
}

// githubWebhookEnvelope captures GitHub webhook fields required for alert processing.
type githubWebhookEnvelope struct {
	// Action represents the GitHub webhook action (created, resolved, etc).
	Action string `json:"action"`
	// Installation describes the GitHub App installation context.
	Installation *githubWebhookInstallation `json:"installation"`
	// Repository describes the repository associated with the alert.
	Repository *githubWebhookRepository `json:"repository"`
	// Alert contains the raw alert payload.
	Alert json.RawMessage `json:"alert"`
}

// githubWebhookInstallation identifies a GitHub App installation.
type githubWebhookInstallation struct {
	// ID is the installation identifier.
	ID int64 `json:"id"`
}

// githubWebhookRepository captures repository details from a webhook payload.
type githubWebhookRepository struct {
	// FullName is the repository full name (owner/name).
	FullName string `json:"full_name"`
	// Name is the repository name.
	Name string `json:"name"`
	// HTMLURL is the repository URL in GitHub.
	HTMLURL string `json:"html_url"`
	// Owner identifies the repository owner.
	Owner githubWebhookRepoOwner `json:"owner"`
}

// githubWebhookRepoOwner captures repository owner info.
type githubWebhookRepoOwner struct {
	// Login is the owner login name.
	Login string `json:"login"`
}

// githubInstallationID returns the installation ID as a string.
func githubInstallationID(installation *githubWebhookInstallation) string {
	if installation == nil || installation.ID == 0 {
		return ""
	}
	return fmt.Sprintf("%d", installation.ID)
}

// githubRepoFromWebhook derives the repo name from a webhook repository payload.
func githubRepoFromWebhook(repo *githubWebhookRepository) string {
	if repo == nil {
		return ""
	}
	if strings.TrimSpace(repo.FullName) != "" {
		return strings.TrimSpace(repo.FullName)
	}
	owner := strings.TrimSpace(repo.Owner.Login)
	name := strings.TrimSpace(repo.Name)
	if owner != "" && name != "" {
		return owner + "/" + name
	}
	return name
}

// githubAlertTypeFromEvent maps GitHub webhook event names to alert types.
func githubAlertTypeFromEvent(eventType string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(eventType)) {
	case "dependabot_alert":
		return "dependabot", true
	case "code_scanning_alert":
		return "code_scanning", true
	case "secret_scanning_alert":
		return "secret_scanning", true
	default:
		return "", false
	}
}

// validateGitHubWebhookSignature checks the webhook signature using the shared secret.
func validateGitHubWebhookSignature(secret string, signatureHeader string, payload []byte) bool {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return false
	}

	signatureHeader = strings.TrimSpace(signatureHeader)
	if signatureHeader == "" {
		return false
	}

	const prefix = "sha256="
	if !strings.HasPrefix(signatureHeader, prefix) {
		return false
	}

	provided := strings.TrimPrefix(signatureHeader, prefix)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(provided))
}

// findGitHubAppIntegrationByInstallationID locates the integration record for an installation ID.
func (h *Handler) findGitHubAppIntegrationByInstallationID(ctx context.Context, installationID string) (*ent.Integration, error) {
	if h.DBClient == nil {
		return nil, errDBClientNotConfigured
	}
	installationID = strings.TrimSpace(installationID)
	if installationID == "" {
		return nil, nil
	}

	query := h.DBClient.Integration.Query().
		Where(
			integration.KindEQ(string(github.TypeGitHubApp)),
			func(s *sql.Selector) {
				s.Where(sqljson.ValueEQ(integration.FieldProviderState, installationID, sqljson.Path("github", "installationId")))
			},
		)
	record, err := query.Only(ctx)
	if err != nil {
		if ent.IsNotSingular(err) {
			record, err = query.First(ctx)
		}
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return record, nil
}
