package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/resend/resend-go/v3"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/campaigntarget"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
)

// ResendWebhookHandler handles inbound Resend webhook events.
func (h *Handler) ResendWebhookHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	if isRegistrationContext(ctx) {
		return nil
	}

	startTime := time.Now()
	eventLabel := "resend"

	if !h.CampaignWebhook.Enabled || h.CampaignWebhook.ResendSecret == "" {
		webhookResponseCounter.WithLabelValues(eventLabel+"_disabled", "200").Inc()

		return h.Success(ctx, "webhook disabled")
	}

	req := ctx.Request()
	res := ctx.Response()

	payload, err := io.ReadAll(http.MaxBytesReader(res.Writer, req.Body, maxBodyBytes))
	if err != nil {
		webhookResponseCounter.WithLabelValues(eventLabel+"_payload_read", "500").Inc()
		logx.FromContext(req.Context()).Error().Err(err).Msg("failed to read resend webhook payload")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	if len(payload) == 0 {
		webhookResponseCounter.WithLabelValues(eventLabel+"_empty_payload", "400").Inc()

		return h.BadRequest(ctx, errPayloadEmpty, openapi)
	}

	headers := resend.WebhookHeaders{
		Id:        req.Header.Get("svix-id"),
		Timestamp: req.Header.Get("svix-timestamp"),
		Signature: req.Header.Get("svix-signature"),
	}

	if headers.Id == "" {
		webhookResponseCounter.WithLabelValues(eventLabel+"_missing_id", "400").Inc()

		return h.BadRequest(ctx, ErrResendWebhookMissingID, openapi)
	}

	client := resend.NewClient(h.CampaignWebhook.ResendAPIKey)

	if err := client.Webhooks.Verify(&resend.VerifyWebhookOptions{
		Payload:       string(payload),
		Headers:       headers,
		WebhookSecret: h.CampaignWebhook.ResendSecret,
	}); err != nil {
		webhookResponseCounter.WithLabelValues(eventLabel+"_invalid_signature", "400").Inc()
		logx.FromContext(req.Context()).Error().Err(err).Msg("resend webhook verification failed")

		return h.BadRequest(ctx, err, openapi)
	}

	var event resendWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		webhookResponseCounter.WithLabelValues(eventLabel+"_invalid_payload", "400").Inc()
		logx.FromContext(req.Context()).Error().Err(err).Msg("failed to parse resend webhook payload")

		return h.BadRequest(ctx, err, openapi)
	}

	webhookReceivedCounter.WithLabelValues(eventLabel).Inc()
	if event.Type != "" && !supportedResendEventTypes[event.Type] {
		webhookResponseCounter.WithLabelValues(eventLabel+"_discarded", "200").Inc()

		return h.Success(ctx, "unsupported event type", openapi)
	}

	allowCtx := privacy.DecisionContext(req.Context(), privacy.Allow)
	allowCtx = auth.WithCaller(allowCtx, auth.NewWebhookCaller(""))

	if event.Type != "" {
		eventLabel = "resend." + event.Type
	}

	exists, err := h.checkForEventID(allowCtx, headers.Id)
	if err != nil {
		webhookResponseCounter.WithLabelValues(eventLabel+"_event_lookup", "500").Inc()
		logx.FromContext(req.Context()).Error().Err(err).Msg("failed to check resend event id")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	if exists {
		duration := time.Since(startTime).Seconds()
		webhookProcessingLatency.WithLabelValues(eventLabel).Observe(duration)
		webhookResponseCounter.WithLabelValues(eventLabel+"_duplicate", "200").Inc()

		return h.Success(ctx, "duplicate webhook event", openapi)
	}

	metadata := map[string]any{
		"event_type": event.Type,
		"email_id":   event.Data.EmailID,
		"created_at": event.Data.CreatedAt,
	}

	if len(event.Data.Tags) > 0 {
		tags := make(map[string]string, len(event.Data.Tags))
		for _, tag := range event.Data.Tags {
			if tag.Name == "" || tag.Value == "" {
				continue
			}

			tags[strings.ToLower(tag.Name)] = tag.Value
		}

		if len(tags) > 0 {
			metadata["tags"] = tags
		}
	}

	eventType := eventLabel
	if eventType == "resend" {
		eventType = "resend.webhook"
	}

	if _, err := h.createEvent(allowCtx, generated.CreateEventInput{
		EventID:   &headers.Id,
		EventType: eventType,
		Metadata:  metadata,
	}); err != nil {
		webhookResponseCounter.WithLabelValues(eventLabel+"_event_create", "500").Inc()
		logx.FromContext(req.Context()).Error().Err(err).Msg("failed to create resend event")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	if err := h.handleResendWebhookEvent(allowCtx, event); err != nil {
		webhookResponseCounter.WithLabelValues(eventLabel+"_processing_failed", "500").Inc()
		logx.FromContext(req.Context()).Error().Err(err).Msg("failed to handle resend webhook event")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	duration := time.Since(startTime).Seconds()
	webhookProcessingLatency.WithLabelValues(eventLabel).Observe(duration)
	webhookResponseCounter.WithLabelValues(eventLabel+"_processed", "200").Inc()

	return h.Success(ctx, nil)
}

// resendWebhookEvent represents the structure of a Resend webhook event payload.
type resendWebhookEvent struct {
	// Type is the event type, e.g. "email.sent", "email.delivered".
	Type string `json:"type"`
	// Data contains the event-specific data.
	Data resendWebhookData `json:"data"`
}

// resendWebhookData contains the data portion of a Resend webhook event.
type resendWebhookData struct {
	// EmailID is the unique identifier for the email in Resend.
	EmailID string `json:"email_id"`
	// CreatedAt is the timestamp when the event occurred.
	CreatedAt string `json:"created_at"`
	// Tags are the custom tags attached to the email for tracking purposes.
	Tags []resendWebhookTag `json:"tags"`
}

// resendWebhookTag represents a single tag from a Resend webhook event.
type resendWebhookTag struct {
	// Name is the tag key.
	Name string `json:"name"`
	// Value is the tag value.
	Value string `json:"value"`
}

// supportedResendEventTypes is a set of Resend event types that this handler processes.
// Events not in this set are acknowledged but not processed.
var supportedResendEventTypes = func() map[string]bool {
	return map[string]bool{
		resend.EventEmailSent:      true,
		resend.EventEmailDelivered: true,
		resend.EventEmailOpened:    true,
		resend.EventEmailClicked:   true,
		resend.EventEmailBounced:   true,
		resend.EventEmailFailed:    true,
	}
}()

// handleResendWebhookEvent routes Resend events to the relevant update flows based on tags.
func (h *Handler) handleResendWebhookEvent(ctx context.Context, event resendWebhookEvent) error {
	if event.Type == "" {
		return nil
	}

	tagMap := make(map[string]string)
	for _, tag := range event.Data.Tags {
		if tag.Name == "" || tag.Value == "" {
			continue
		}
		tagMap[strings.ToLower(tag.Name)] = tag.Value
	}

	isTest := false
	if value, ok := tagMap["is_test"]; ok {
		switch strings.ToLower(value) {
		case "true", "1", "yes", "y":
			isTest = true
		}
	}

	assessmentResponseID := tagMap["assessment_response_id"]
	campaignTargetID := tagMap["campaign_target_id"]

	if assessmentResponseID != "" {
		if err := h.updateAssessmentResponseFromResend(ctx, assessmentResponseID, event, isTest); err != nil {
			return err
		}
	}

	if campaignTargetID != "" && !isTest {
		if err := h.updateCampaignTargetFromResend(ctx, campaignTargetID, "", event); err != nil {
			return err
		}
	}

	return nil
}

// updateAssessmentResponseFromResend applies delivery events to an assessment response record.
// The isTest parameter indicates whether the webhook event has the is_test tag set, which
// prevents cascading updates to campaign targets.
func (h *Handler) updateAssessmentResponseFromResend(ctx context.Context, responseID string, event resendWebhookEvent, isTest bool) error {
	resp, err := h.DBClient.AssessmentResponse.Get(ctx, responseID)
	if err != nil {
		return err
	}

	update := h.DBClient.AssessmentResponse.UpdateOneID(responseID)
	eventTime := parseResendEventTime(event.Data.CreatedAt)

	if eventTime != nil {
		update.SetLastEmailEventAt(*eventTime)
	}

	switch event.Type {
	case resend.EventEmailSent:
		update.SetStatus(enums.AssessmentResponseStatusSent)
	case resend.EventEmailDelivered:
		if eventTime != nil {
			update.SetEmailDeliveredAt(*eventTime)
		}
	case resend.EventEmailOpened:
		if eventTime != nil {
			update.SetEmailOpenedAt(*eventTime)
		}
		update.AddEmailOpenCount(1)
	case resend.EventEmailClicked:
		if eventTime != nil {
			update.SetEmailClickedAt(*eventTime)
		}
		update.AddEmailClickCount(1)
	case resend.EventEmailBounced, resend.EventEmailFailed:
		metadata := resp.EmailMetadata
		if metadata == nil {
			metadata = map[string]any{}
		}
		metadata["resend_event"] = event.Type
		metadata["resend_email_id"] = event.Data.EmailID
		metadata["resend_timestamp"] = event.Data.CreatedAt
		update.SetEmailMetadata(metadata)
	}

	if _, err := update.Save(ctx); err != nil {
		return err
	}

	if resp.CampaignID != "" && !resp.IsTest && !isTest {
		return h.updateCampaignTargetFromResend(ctx, "", resp.CampaignID, event, resp.Email)
	}

	return nil
}

// updateCampaignTargetFromResend applies delivery events to a campaign target record.
// It can locate the target either by targetID directly or by campaignID and email combination.
func (h *Handler) updateCampaignTargetFromResend(ctx context.Context, targetID, campaignID string, event resendWebhookEvent, email ...string) error {
	var (
		target *generated.CampaignTarget
		err    error
	)

	switch {
	case targetID != "":
		target, err = h.DBClient.CampaignTarget.Get(ctx, targetID)
	case campaignID != "" && len(email) > 0 && email[0] != "":
		target, err = h.DBClient.CampaignTarget.Query().
			Where(
				campaigntarget.CampaignIDEQ(campaignID),
				campaigntarget.EmailEqualFold(email[0]),
			).
			Only(ctx)
	default:
		return nil
	}

	if err != nil {
		if generated.IsNotFound(err) {
			return nil
		}
		return err
	}

	update := h.DBClient.CampaignTarget.UpdateOneID(target.ID)

	eventTime := parseResendEventTime(event.Data.CreatedAt)
	switch event.Type {
	case resend.EventEmailSent, resend.EventEmailDelivered, resend.EventEmailOpened, resend.EventEmailClicked:
		update.SetStatus(enums.AssessmentResponseStatusSent)
		if eventTime != nil {
			update.SetSentAt(models.DateTime(*eventTime))
		}
	case resend.EventEmailBounced, resend.EventEmailFailed:
		metadata := target.Metadata
		if metadata == nil {
			metadata = map[string]any{}
		}
		metadata["resend_event"] = event.Type
		metadata["resend_email_id"] = event.Data.EmailID
		metadata["resend_timestamp"] = event.Data.CreatedAt
		update.SetMetadata(metadata)
	}

	_, err = update.Save(ctx)
	return err
}

// parseResendEventTime parses a Resend timestamp string and returns a time.Time pointer.
// It supports both RFC3339 and RFC3339Nano formats. Returns nil if the value is empty or unparsable.
func parseResendEventTime(value string) *time.Time {
	if value == "" {
		return nil
	}

	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return &parsed
	}

	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return &parsed
	}

	return nil
}
