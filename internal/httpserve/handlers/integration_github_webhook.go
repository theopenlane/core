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
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/ingest"
	"github.com/theopenlane/core/internal/integrations/providers/github"
	"github.com/theopenlane/core/pkg/gala"
)

// GitHub webhook header names and limits.
const (
	githubWebhookSignatureHeader = "X-Hub-Signature-256"
	githubWebhookEventHeader     = "X-GitHub-Event"
	githubWebhookDeliveryHeader  = "X-GitHub-Delivery"
	maxGitHubWebhookBodyBytes    = int64(1024 * 1024)
)

var (
	githubAppWebhookReceivedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openlane_github_app_webhook_received_total",
			Help: "Total number of GitHub App webhooks received by event type",
		},
		[]string{"event_type"},
	)

	githubAppWebhookResponseCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openlane_github_app_webhook_responses_total",
			Help: "Total number of GitHub App webhook responses by event, status code, and result",
		},
		[]string{"event_type", "status_code", "result"},
	)

	githubAppWebhookProcessingLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "openlane_github_app_webhook_processing_latency_seconds",
			Help:    "Latency of GitHub App webhook processing in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"event_type"},
	)

	githubAppWebhookAlertsQueuedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openlane_github_app_webhook_alerts_queued_total",
			Help: "Total number of GitHub App vulnerability alerts queued for ingest by event and alert type",
		},
		[]string{"event_type", "alert_type"},
	)

	githubAppWebhookEmitErrorsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openlane_github_app_webhook_emit_errors_total",
			Help: "Total number of GitHub App webhook ingest emit errors by event type",
		},
		[]string{"event_type"},
	)
)

func init() {
	prometheus.MustRegister(githubAppWebhookReceivedCounter)
	prometheus.MustRegister(githubAppWebhookResponseCounter)
	prometheus.MustRegister(githubAppWebhookProcessingLatency)
	prometheus.MustRegister(githubAppWebhookAlertsQueuedCounter)
	prometheus.MustRegister(githubAppWebhookEmitErrorsCounter)
}

func normalizeGitHubWebhookEventType(eventType string) string {
	var normalized types.LowerString
	if err := normalized.UnmarshalText([]byte(eventType)); err != nil {
		return "unknown"
	}

	if normalized == "" {
		return "unknown"
	}

	return normalized.String()
}

func recordGitHubWebhookResponse(eventType string, statusCode int, result string) {
	githubAppWebhookResponseCounter.WithLabelValues(
		normalizeGitHubWebhookEventType(eventType),
		strconv.Itoa(statusCode),
		result,
	).Inc()
}

// GitHubIntegrationWebhookHandler ingests GitHub App security alert webhooks.
func (h *Handler) GitHubIntegrationWebhookHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	if isRegistrationContext(ctx) {
		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{
			Reply:     rout.Reply{Success: true},
			Persisted: map[string]any{"created": 0, "updated": 0, "skipped": 0},
		}, openapi)
	}

	req := ctx.Request()
	res := ctx.Response()

	eventType := req.Header.Get(githubWebhookEventHeader)
	metricEventType := normalizeGitHubWebhookEventType(eventType)
	githubAppWebhookReceivedCounter.WithLabelValues(metricEventType).Inc()

	start := time.Now()
	defer func() {
		githubAppWebhookProcessingLatency.WithLabelValues(metricEventType).Observe(time.Since(start).Seconds())
	}()

	if err := h.validateGitHubAppConfig(); err != nil {
		recordGitHubWebhookResponse(eventType, http.StatusBadRequest, "invalid_config")

		return h.BadRequest(ctx, err, openapi)
	}

	payload, err := io.ReadAll(http.MaxBytesReader(res.Writer, req.Body, maxGitHubWebhookBodyBytes))
	if err != nil {
		recordGitHubWebhookResponse(eventType, http.StatusInternalServerError, "payload_read_failed")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}
	if len(payload) == 0 {
		recordGitHubWebhookResponse(eventType, http.StatusBadRequest, "empty_payload")

		return h.BadRequest(ctx, errPayloadEmpty, openapi)
	}

	if eventType == "" {
		recordGitHubWebhookResponse(eventType, http.StatusBadRequest, "missing_event_header")
		return h.BadRequest(ctx, rout.MissingField(githubWebhookEventHeader), openapi)
	}

	var envelope githubWebhookEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		recordGitHubWebhookResponse(eventType, http.StatusBadRequest, "invalid_payload")

		return h.BadRequest(ctx, err, openapi)
	}

	installationID := githubInstallationID(envelope.Installation)
	if installationID == "" {
		logx.FromContext(req.Context()).Info().Str("event", eventType).Msg("github webhook missing installation id")
		recordGitHubWebhookResponse(eventType, http.StatusOK, "missing_installation_id")

		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	integrationRecord, err := h.findGitHubAppIntegrationByInstallationID(req.Context(), installationID)
	if err != nil {
		recordGitHubWebhookResponse(eventType, http.StatusInternalServerError, "integration_lookup_failed")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}
	if integrationRecord == nil {
		logx.FromContext(req.Context()).Info().Str("installation_id", installationID).Msg("no integration configured for github app installation")
		recordGitHubWebhookResponse(eventType, http.StatusOK, "integration_not_configured")

		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	webhookSecret := h.IntegrationGitHubApp.WebhookSecret
	if webhookSecret == "" {
		recordGitHubWebhookResponse(eventType, http.StatusInternalServerError, "missing_webhook_secret")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	if !validateGitHubWebhookSignature(webhookSecret, req.Header.Get(githubWebhookSignatureHeader), payload) {
		recordGitHubWebhookResponse(eventType, http.StatusBadRequest, "invalid_signature")

		return h.BadRequest(ctx, rout.InvalidField(githubWebhookSignatureHeader), openapi)
	}

	allowCtx := privacy.DecisionContext(req.Context(), privacy.Allow)
	if deliveryID := req.Header.Get(githubWebhookDeliveryHeader); deliveryID != "" {
		exists, err := h.checkForEventID(allowCtx, deliveryID)
		if err != nil {
			recordGitHubWebhookResponse(eventType, http.StatusInternalServerError, "delivery_dedupe_check_failed")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
		}

		if exists {
			recordGitHubWebhookResponse(eventType, http.StatusOK, "duplicate_delivery_ignored")

			return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
		}

		if _, err := h.createEvent(allowCtx, ent.CreateEventInput{EventID: &deliveryID}); err != nil {
			recordGitHubWebhookResponse(eventType, http.StatusInternalServerError, "delivery_record_create_failed")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
		}
	}

	if eventType == "ping" {
		recordGitHubWebhookResponse(eventType, http.StatusOK, "ping_ignored")

		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	if strings.EqualFold(eventType, "installation") {
		return h.handleGitHubInstallationWebhook(ctx, openapi, eventType, envelope, integrationRecord)
	}

	repo := githubRepoFromWebhook(envelope.Repository)
	if repo == "" {
		recordGitHubWebhookResponse(eventType, http.StatusOK, "missing_repository")

		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	alertType, ok := githubAlertTypeFromEvent(eventType)
	if !ok {
		recordGitHubWebhookResponse(eventType, http.StatusOK, "unsupported_event_type")

		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	alerts := []types.AlertEnvelope{
		{
			AlertType: alertType,
			Resource:  repo,
			Action:    envelope.Action,
			Payload:   envelope.Alert,
		},
	}

	if h.Gala != nil {
		receipt := h.Gala.EmitWithHeaders(req.Context(), ingest.IntegrationIngestRequestedTopic.Name, ingest.RequestedPayload{
			IntegrationID: integrationRecord.ID,
			Schema:        integrationgenerated.IntegrationMappingSchemaVulnerability,
			Envelopes:     alerts,
		}, gala.Headers{})
		if receipt.Err != nil {
			logx.FromContext(req.Context()).Warn().Err(receipt.Err).Msg("failed to emit integration ingest event")
		}
	}

	githubAppWebhookAlertsQueuedCounter.WithLabelValues(metricEventType, alertType).Add(float64(len(alerts)))
	recordGitHubWebhookResponse(eventType, http.StatusOK, "alerts_queued")

	return h.Success(ctx, apimodels.GitHubAppWebhookResponse{
		Reply: rout.Reply{Success: true},
		Persisted: map[string]any{
			"queued": len(alerts),
			"total":  len(alerts),
		},
	}, openapi)
}

// handleGitHubInstallationWebhook sends a Slack notification when an installation is created.
func (h *Handler) handleGitHubInstallationWebhook(ctx echo.Context, openapi *OpenAPIContext, eventType string, envelope githubWebhookEnvelope, integrationRecord *ent.Integration) error {
	if !strings.EqualFold(envelope.Action, "created") {
		recordGitHubWebhookResponse(eventType, http.StatusOK, "installation_action_ignored")

		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	if integrationRecord == nil {
		recordGitHubWebhookResponse(eventType, http.StatusOK, "integration_not_configured")

		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	if !hooks.SlackNotificationsEnabled() {
		recordGitHubWebhookResponse(eventType, http.StatusOK, "installation_notification_skipped")

		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	githubOrg := ""
	githubAccountType := ""
	if envelope.Installation != nil && envelope.Installation.Account != nil {
		githubOrg = envelope.Installation.Account.Login
		githubAccountType = envelope.Installation.Account.Type
	}

	openlaneOrgID := integrationRecord.OwnerID
	openlaneOrgName := h.resolveOpenlaneOrganizationName(ctx.Request().Context(), openlaneOrgID)
	message := buildGitHubAppInstallSlackMessage(githubOrg, githubAccountType, openlaneOrgName, openlaneOrgID)

	if err := hooks.SendSlackNotification(ctx.Request().Context(), message); err != nil {
		recordGitHubWebhookResponse(eventType, http.StatusOK, "installation_notification_failed")

		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	recordGitHubWebhookResponse(eventType, http.StatusOK, "installation_notification_sent")

	return h.Success(ctx, apimodels.GitHubAppWebhookResponse{
		Reply: rout.Reply{Success: true},
		Persisted: map[string]any{
			"notified": true,
		},
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
	// Account identifies the user or organization that owns the installation.
	Account *githubWebhookAccount `json:"account"`
	// TargetType is the installation target type provided by GitHub.
	TargetType string `json:"target_type"`
}

// githubWebhookAccount captures the GitHub account metadata in webhook payloads.
type githubWebhookAccount struct {
	// Login is the GitHub account login.
	Login string `json:"login"`
	// Type is the GitHub account type.
	Type string `json:"type"`
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
			return query.First(ctx)
		}

		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return record, nil
}
