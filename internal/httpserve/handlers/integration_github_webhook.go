package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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

	apimodels "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/providers/github"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/mapx"
)

// GitHub webhook header names and limits
const (
	githubWebhookSignatureHeader = "X-Hub-Signature-256"
	githubWebhookEventHeader     = "X-GitHub-Event"
	githubWebhookDeliveryHeader  = "X-GitHub-Delivery"
	maxGitHubWebhookBodyBytes    = int64(1024 * 1024)
)

// githubWebhookVerificationStatePatch captures provider state fields persisted after webhook verification
type githubWebhookVerificationStatePatch struct {
	// InstallationID is the installed GitHub App installation identifier
	InstallationID string `json:"installationId"`
	// WebhookVerifiedAt marks the verified timestamp
	WebhookVerifiedAt time.Time `json:"webhookVerifiedAt"`
}

// githubWebhookVerificationMetadata captures integration metadata fields persisted after webhook verification
type githubWebhookVerificationMetadata struct {
	// GitHubWebhookVerifiedAt marks the verified timestamp for UI and operational visibility
	GitHubWebhookVerifiedAt time.Time `json:"githubWebhookVerifiedAt"`
}

// githubWebhookPersistedRegistration summarizes registration-safe webhook persisted counters
type githubWebhookPersistedRegistration struct {
	// Created is the number of created records
	Created int `json:"created"`
	// Updated is the number of updated records
	Updated int `json:"updated"`
	// Skipped is the number of skipped records
	Skipped int `json:"skipped"`
}

// githubWebhookPersistedQueue summarizes queued envelope counters
type githubWebhookPersistedQueue struct {
	// Queued is the number of queued alerts
	Queued int `json:"queued"`
	// Total is the total number of queued alerts
	Total int `json:"total"`
}

// githubWebhookPersistedNotification summarizes installation notification results
type githubWebhookPersistedNotification struct {
	// Notified indicates whether installation notification was sent
	Notified bool `json:"notified"`
}

var (
	errGitHubAppIntegrationAmbiguous = errors.New("multiple github app integrations found for installation")

	githubAppWebhookReceivedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openlane_githubapp_webhook_received_total",
			Help: "Total number of GitHub App webhooks received by event type",
		},
		[]string{"event_type"},
	)

	githubAppWebhookResponseCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openlane_githubapp_webhook_responses_total",
			Help: "Total number of GitHub App webhook responses by event, status code, and result",
		},
		[]string{"event_type", "status_code", "result"},
	)

	githubAppWebhookProcessingLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "openlane_githubapp_webhook_processing_latency_seconds",
			Help:    "Latency of GitHub App webhook processing in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"event_type"},
	)

	githubAppWebhookAlertsQueuedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openlane_githubapp_webhook_alerts_queued_total",
			Help: "Total number of GitHub App vulnerability alerts queued for ingest by event and alert type",
		},
		[]string{"event_type", "alert_type"},
	)

	githubAppWebhookEmitErrorsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openlane_githubapp_webhook_emit_errors_total",
			Help: "Total number of GitHub App webhook ingest emit errors by event type",
		},
		[]string{"event_type"},
	)
)

// init registers GitHub App webhook metrics
func init() {
	prometheus.MustRegister(githubAppWebhookReceivedCounter)
	prometheus.MustRegister(githubAppWebhookResponseCounter)
	prometheus.MustRegister(githubAppWebhookProcessingLatency)
	prometheus.MustRegister(githubAppWebhookAlertsQueuedCounter)
	prometheus.MustRegister(githubAppWebhookEmitErrorsCounter)
}

// githubWebhookPersistedMap encodes persisted response payloads into map form for OpenAPI response types
func githubWebhookPersistedMap(value any) map[string]any {
	persisted, err := jsonx.ToMap(value)
	if err != nil {
		return nil
	}

	return persisted
}

// normalizeGitHubWebhookEventType normalizes the event type used in metric labels
func normalizeGitHubWebhookEventType(eventType string) string {
	if eventType == "" {
		return "unknown"
	}

	return eventType
}

// recordGitHubWebhookResponse increments webhook response metrics
func recordGitHubWebhookResponse(eventType string, statusCode int, result string) {
	githubAppWebhookResponseCounter.WithLabelValues(
		normalizeGitHubWebhookEventType(eventType),
		strconv.Itoa(statusCode),
		result,
	).Inc()
}

// GitHubIntegrationWebhookHandler ingests GitHub App security alert webhooks
func (h *Handler) GitHubIntegrationWebhookHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	if isRegistrationContext(ctx) {
		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{
			Reply:     rout.Reply{Success: true},
			Persisted: githubWebhookPersistedMap(githubWebhookPersistedRegistration{}),
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
		return h.BadRequest(ctx, ErrGitHubWebhookEventHeaderMissing, openapi)
	}

	appCfg, ok := h.gitHubAppConfig()
	if !ok || appCfg.WebhookSecret == "" {
		recordGitHubWebhookResponse(eventType, http.StatusInternalServerError, "missing_webhook_secret")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}
	webhookSecret := appCfg.WebhookSecret

	if !validateGitHubWebhookSignature(webhookSecret, req.Header.Get(githubWebhookSignatureHeader), payload) {
		recordGitHubWebhookResponse(eventType, http.StatusBadRequest, "invalid_signature")

		return h.BadRequest(ctx, ErrGitHubWebhookSignatureInvalid, openapi)
	}

	var envelope githubWebhookEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		recordGitHubWebhookResponse(eventType, http.StatusBadRequest, "invalid_payload")

		return h.BadRequest(ctx, err, openapi)
	}

	installationID := githubInstallationID(envelope.Installation)
	allowCtx := privacy.DecisionContext(req.Context(), privacy.Allow)

	if eventType == "ping" {
		if installationID != "" {
			if err := h.markGitHubWebhookVerifiedAt(allowCtx, installationID, time.Now().UTC()); err != nil {
				logx.FromContext(req.Context()).Warn().Err(err).Str("installation_id", installationID).Msg("failed to persist github webhook verification timestamp")
			}
		}

		recordGitHubWebhookResponse(eventType, http.StatusOK, "ping_accepted")

		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	if installationID == "" {
		logx.FromContext(req.Context()).Info().Str("event", eventType).Msg("github webhook missing installation id")
		recordGitHubWebhookResponse(eventType, http.StatusOK, "missing_installation_id")

		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	integrationRecord, err := h.findGitHubAppIntegrationByInstallationID(allowCtx, installationID)
	if err != nil {
		logx.FromContext(req.Context()).Error().Err(err).Str("installation_id", installationID).Msg("failed to resolve github app integration by installation id")

		recordGitHubWebhookResponse(eventType, http.StatusInternalServerError, "integration_lookup_failed")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}
	if integrationRecord == nil {
		logx.FromContext(req.Context()).Info().Str("installation_id", installationID).Msg("no integration configured for github app installation")
		recordGitHubWebhookResponse(eventType, http.StatusOK, "integration_not_configured")

		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	if deliveryID := req.Header.Get(githubWebhookDeliveryHeader); deliveryID != "" {
		duplicate, err := h.registerGitHubWebhookDelivery(allowCtx, integrationRecord, deliveryID)
		if err != nil {
			recordGitHubWebhookResponse(eventType, http.StatusInternalServerError, "delivery_dedupe_record_failed")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
		}

		if duplicate {
			recordGitHubWebhookResponse(eventType, http.StatusOK, "duplicate_delivery_ignored")

			return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
		}
	}

	if eventType == "installation" {
		openlaneOrgName := h.resolveOpenlaneOrganizationName(ctx.Request().Context(), integrationRecord.OwnerID)

		return h.handleGitHubInstallationWebhook(ctx, openapi, eventType, envelope, integrationRecord, openlaneOrgName)
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
		opEnvelope := integrations.NewIntegrationOperationEnvelope(integrations.IntegrationOperationRequestedPayload{
			OrgID:     integrationRecord.OwnerID,
			Provider:  string(github.TypeGitHubApp),
			Operation: string(types.OperationVulnerabilitiesCollect),
			RunType:   enums.IntegrationRunTypeWebhook,
		})
		receipt := h.Gala.EmitWithHeaders(req.Context(), integrations.IntegrationOperationRequestedTopic.Name, opEnvelope, opEnvelope.Headers())
		if receipt.Err != nil {
			logx.FromContext(req.Context()).Warn().Err(receipt.Err).Msg("failed to emit integration ingest event")
		}
	}

	githubAppWebhookAlertsQueuedCounter.WithLabelValues(metricEventType, alertType).Add(float64(len(alerts)))
	recordGitHubWebhookResponse(eventType, http.StatusOK, "alerts_queued")

	return h.Success(ctx, apimodels.GitHubAppWebhookResponse{
		Reply: rout.Reply{Success: true},
		Persisted: githubWebhookPersistedMap(githubWebhookPersistedQueue{
			Queued: len(alerts),
			Total:  len(alerts),
		}),
	}, openapi)
}

// handleGitHubInstallationWebhook sends a Slack notification when an installation is created
func (h *Handler) handleGitHubInstallationWebhook(ctx echo.Context, openapi *OpenAPIContext, eventType string, envelope githubWebhookEnvelope, integrationRecord *ent.Integration, openlaneOrgName string) error {
	if envelope.Action != "created" {
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
	message, err := hooks.RenderGitHubAppInstallSlackMessage(githubOrg, githubAccountType, openlaneOrgName, openlaneOrgID)
	if err != nil {
		recordGitHubWebhookResponse(eventType, http.StatusOK, "installation_notification_failed")

		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	if err := hooks.SendSlackNotification(ctx.Request().Context(), message); err != nil {
		recordGitHubWebhookResponse(eventType, http.StatusOK, "installation_notification_failed")

		return h.Success(ctx, apimodels.GitHubAppWebhookResponse{Reply: rout.Reply{Success: true}}, openapi)
	}

	recordGitHubWebhookResponse(eventType, http.StatusOK, "installation_notification_sent")

	return h.Success(ctx, apimodels.GitHubAppWebhookResponse{
		Reply:     rout.Reply{Success: true},
		Persisted: githubWebhookPersistedMap(githubWebhookPersistedNotification{Notified: true}),
	}, openapi)
}

// githubWebhookEnvelope captures GitHub webhook fields required for alert processing
type githubWebhookEnvelope struct {
	// Action represents the GitHub webhook action (created, resolved, etc)
	Action string `json:"action"`
	// Installation describes the GitHub App installation context
	Installation *githubWebhookInstallation `json:"installation"`
	// Repository describes the repository associated with the alert
	Repository *githubWebhookRepository `json:"repository"`
	// Alert contains the raw alert payload
	Alert json.RawMessage `json:"alert"`
}

// githubWebhookInstallation identifies a GitHub App installation
type githubWebhookInstallation struct {
	// ID is the installation identifier
	ID int64 `json:"id"`
	// Account identifies the user or organization that owns the installation
	Account *githubWebhookAccount `json:"account"`
	// TargetType is the installation target type provided by GitHub
	TargetType string `json:"target_type"`
}

// githubWebhookAccount captures the GitHub account metadata in webhook payloads
type githubWebhookAccount struct {
	// Login is the GitHub account login
	Login string `json:"login"`
	// Type is the GitHub account type
	Type string `json:"type"`
}

// githubWebhookRepository captures repository details from a webhook payload
type githubWebhookRepository struct {
	// FullName is the repository full name (owner/name)
	FullName string `json:"full_name"`
	// Name is the repository name
	Name string `json:"name"`
	// HTMLURL is the repository URL in GitHub
	HTMLURL string `json:"html_url"`
	// Owner identifies the repository owner
	Owner githubWebhookRepoOwner `json:"owner"`
}

// githubWebhookRepoOwner captures repository owner info
type githubWebhookRepoOwner struct {
	// Login is the owner login name
	Login string `json:"login"`
}

// githubInstallationID returns the installation ID as a string
func githubInstallationID(installation *githubWebhookInstallation) string {
	if installation == nil || installation.ID == 0 {
		return ""
	}

	return fmt.Sprintf("%d", installation.ID)
}

// githubRepoFromWebhook derives the repo name from a webhook repository payload
func githubRepoFromWebhook(repo *githubWebhookRepository) string {
	if repo == nil {
		return ""
	}
	if repo.FullName != "" {
		return repo.FullName
	}
	owner := repo.Owner.Login
	name := repo.Name
	if owner != "" && name != "" {
		return owner + "/" + name
	}
	return name
}

// githubAlertTypeFromEvent maps GitHub webhook event names to alert types
func githubAlertTypeFromEvent(eventType string) (string, bool) {
	switch eventType {
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

// validateGitHubWebhookSignature checks the webhook signature using the shared secret
func validateGitHubWebhookSignature(secret string, signatureHeader string, payload []byte) bool {
	if secret == "" {
		return false
	}

	if signatureHeader == "" {
		return false
	}

	const prefix = "sha256="
	if !strings.HasPrefix(signatureHeader, prefix) {
		return false
	}

	provided := signatureHeader[len(prefix):]
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(provided))
}

// findGitHubAppIntegrationByInstallationID locates the integration record for an installation ID
func (h *Handler) findGitHubAppIntegrationByInstallationID(ctx context.Context, installationID string) (*ent.Integration, error) {
	if installationID == "" {
		return nil, nil
	}

	query := h.DBClient.Integration.Query().
		Where(
			integration.KindEQ(string(github.TypeGitHubApp)),
			func(s *sql.Selector) {
				s.Where(sqljson.ValueEQ(integration.FieldProviderState, installationID, sqljson.Path("providers", string(github.TypeGitHubApp), "installationId")))
			},
		)
	record, err := query.Only(ctx)
	if err != nil {
		if ent.IsNotSingular(err) {
			return nil, errGitHubAppIntegrationAmbiguous
		}

		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return record, nil
}

// markGitHubWebhookVerifiedAt updates provider state and metadata when webhook verification succeeds
func (h *Handler) markGitHubWebhookVerifiedAt(ctx context.Context, installationID string, verifiedAt time.Time) error {
	if h.DBClient == nil {
		return errDBClientNotConfigured
	}

	integrationRecord, err := h.findGitHubAppIntegrationByInstallationID(ctx, installationID)
	if err != nil {
		return err
	}
	if integrationRecord == nil {
		return nil
	}

	statePatch, err := json.Marshal(githubWebhookVerificationStatePatch{
		InstallationID:    installationID,
		WebhookVerifiedAt: verifiedAt,
	})
	if err != nil {
		return ErrInvalidStateFormat
	}

	nextState := integrationRecord.ProviderState
	if _, err := nextState.MergeProviderData(string(github.TypeGitHubApp), statePatch); err != nil {
		return ErrInvalidStateFormat
	}

	metadataPatch, err := jsonx.ToMap(githubWebhookVerificationMetadata{
		GitHubWebhookVerifiedAt: verifiedAt,
	})
	if err != nil {
		return ErrInvalidStateFormat
	}

	nextMetadata := mapx.DeepMergeMapAny(mapx.DeepCloneMapAny(integrationRecord.Metadata), metadataPatch)

	return h.DBClient.Integration.UpdateOneID(integrationRecord.ID).
		SetProviderState(nextState).
		SetMetadata(nextMetadata).
		Exec(ctx)
}

// registerGitHubWebhookDelivery inserts a unique delivery marker for idempotent processing.
// Returns duplicate=true when the delivery has already been recorded.
func (h *Handler) registerGitHubWebhookDelivery(ctx context.Context, integrationRecord *ent.Integration, deliveryID string) (duplicate bool, err error) {
	if h.DBClient == nil {
		return false, errDBClientNotConfigured
	}

	createErr := h.DBClient.IntegrationWebhook.Create().
		SetOwnerID(integrationRecord.OwnerID).
		SetIntegrationID(integrationRecord.ID).
		SetProvider(string(github.TypeGitHubApp)).
		SetExternalEventID(deliveryID).
		Exec(ctx)
	if createErr != nil {
		if ent.IsConstraintError(createErr) {
			return true, nil
		}

		return false, createErr
	}

	return false, nil
}
