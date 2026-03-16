package githubapp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/slacknotify"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/mapx"
)

const (
	githubSignatureHeader       = "X-Hub-Signature-256"
	githubWebhookEventHeader    = "X-GitHub-Event"
	githubWebhookDeliveryHeader = "X-GitHub-Delivery"
)

// Webhook verifies and resolves inbound GitHub App webhook requests
type Webhook struct {
	// Config holds the operator-supplied webhook verification settings.
	Config Config
}

// PingWebhook marks a GitHub App webhook endpoint as verified
type PingWebhook struct{}

// InstallationCreatedWebhook sends the GitHub App installation notification
type InstallationCreatedWebhook struct{}

// DependabotAlertWebhook ingests one Dependabot alert from the webhook payload
type DependabotAlertWebhook struct{}

// CodeScanningAlertWebhook ingests one code scanning alert from the webhook payload
type CodeScanningAlertWebhook struct{}

// SecretScanningAlertWebhook ingests one secret scanning alert from the webhook payload
type SecretScanningAlertWebhook struct{}

type githubWebhookEnvelope struct {
	Action       string                     `json:"action"`
	Installation *githubWebhookInstallation `json:"installation"`
	Repository   *githubWebhookRepository   `json:"repository"`
	Alert        json.RawMessage            `json:"alert"`
}

type githubWebhookInstallation struct {
	ID         int64                 `json:"id"`
	Account    *githubWebhookAccount `json:"account"`
	TargetType string                `json:"target_type"`
}

type githubWebhookAccount struct {
	Login string `json:"login"`
	Type  string `json:"type"`
}

type githubWebhookRepository struct {
	FullName string                 `json:"full_name"`
	Name     string                 `json:"name"`
	HTMLURL  string                 `json:"html_url"`
	Owner    githubWebhookRepoOwner `json:"owner"`
}

type githubWebhookRepoOwner struct {
	Login string `json:"login"`
}

type githubWebhookVerificationStatePatch struct {
	WebhookVerifiedAt time.Time `json:"webhookVerifiedAt"`
}

type githubWebhookVerificationMetadata struct {
	GitHubWebhookVerifiedAt time.Time `json:"githubWebhookVerifiedAt"`
}

// Verify validates the HMAC-SHA256 signature on an inbound GitHub webhook request
func (w Webhook) Verify(ctx context.Context, request types.WebhookVerifyRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if w.Config.WebhookSecret == "" {
		return ErrWebhookSecretMissing
	}

	signature := request.Request.Header.Get(githubSignatureHeader)
	if signature == "" {
		return ErrWebhookSignatureMissing
	}

	mac := hmac.New(sha256.New, []byte(w.Config.WebhookSecret))
	_, _ = mac.Write(request.Payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return ErrWebhookSignatureMismatch
	}

	return nil
}

// Event resolves the inbound webhook payload into one registered GitHub webhook event
func (Webhook) Event(ctx context.Context, request types.WebhookEventRequest) (types.WebhookReceivedEvent, error) {
	if err := ctx.Err(); err != nil {
		return types.WebhookReceivedEvent{}, err
	}

	eventType := request.Request.Header.Get(githubWebhookEventHeader)
	if eventType == "" {
		return types.WebhookReceivedEvent{}, ErrWebhookEventMissing
	}

	var envelope githubWebhookEnvelope
	if err := jsonx.UnmarshalIfPresent(request.Payload, &envelope); err != nil {
		return types.WebhookReceivedEvent{}, err
	}

	name := ""
	switch eventType {
	case "ping":
		name = PingWebhookEvent.Name()
	case "installation":
		if envelope.Action == "created" {
			name = InstallationCreatedWebhookEvent.Name()
		}
	case "dependabot_alert":
		name = DependabotAlertWebhookEvent.Name()
	case "code_scanning_alert":
		name = CodeScanningAlertWebhookEvent.Name()
	case "secret_scanning_alert":
		name = SecretScanningAlertWebhookEvent.Name()
	}

	headers := make(map[string]string, len(request.Request.Header))
	for key, values := range request.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return types.WebhookReceivedEvent{
		Name:       name,
		DeliveryID: request.Request.Header.Get(githubWebhookDeliveryHeader),
		Payload:    jsonx.CloneRawMessage(request.Payload),
		Headers:    headers,
	}, nil
}

// Handle marks the GitHub webhook as verified for the installation
func (PingWebhook) Handle(ctx context.Context, request types.WebhookHandleRequest) error {
	statePatch, err := json.Marshal(githubWebhookVerificationStatePatch{
		WebhookVerifiedAt: time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	nextState := request.Integration.ProviderState
	if _, err := nextState.MergeProviderData(Slug, statePatch); err != nil {
		return err
	}

	metadataPatch, err := jsonx.ToMap(githubWebhookVerificationMetadata{
		GitHubWebhookVerifiedAt: time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	return request.DB.Integration.UpdateOneID(request.Integration.ID).
		SetProviderState(nextState).
		SetMetadata(mapx.DeepMergeMapAny(mapx.DeepCloneMapAny(request.Integration.Metadata), metadataPatch)).
		Exec(ctx)
}

// Handle sends the GitHub App installation-created Slack notification
func (InstallationCreatedWebhook) Handle(ctx context.Context, request types.WebhookHandleRequest) error {
	if !slacknotify.NotificationsEnabled() {
		return nil
	}

	var envelope githubWebhookEnvelope
	if err := jsonx.UnmarshalIfPresent(request.Event.Payload, &envelope); err != nil {
		return err
	}

	org, err := request.DB.Organization.Get(ctx, request.Integration.OwnerID)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Str("organization_id", request.Integration.OwnerID).Msg("failed to resolve openlane organization name for github installation webhook")
		return nil
	}

	openlaneOrgName := org.ID
	if org.DisplayName != "" {
		openlaneOrgName = org.DisplayName
	} else if org.Name != "" {
		openlaneOrgName = org.Name
	}

	githubOrg := ""
	githubAccountType := ""
	if envelope.Installation != nil && envelope.Installation.Account != nil {
		githubOrg = envelope.Installation.Account.Login
		githubAccountType = envelope.Installation.Account.Type
	}

	message, err := slacknotify.RenderGitHubAppInstallMessage(githubOrg, githubAccountType, openlaneOrgName, request.Integration.OwnerID)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Str("integration_id", request.Integration.ID).Msg("failed to render github installation webhook slack message")
		return nil
	}

	if err := slacknotify.SendNotification(ctx, message); err != nil {
		logx.FromContext(ctx).Warn().Err(err).Str("integration_id", request.Integration.ID).Msg("failed to send github installation webhook slack notification")
	}

	return nil
}

// Handle ingests one Dependabot alert from the webhook payload
func (DependabotAlertWebhook) Handle(ctx context.Context, request types.WebhookHandleRequest) error {
	return ingestGitHubAlert(ctx, request, githubAlertTypeDependabot)
}

// Handle ingests one code scanning alert from the webhook payload
func (CodeScanningAlertWebhook) Handle(ctx context.Context, request types.WebhookHandleRequest) error {
	return ingestGitHubAlert(ctx, request, githubAlertTypeCodeScanning)
}

// Handle ingests one secret scanning alert from the webhook payload
func (SecretScanningAlertWebhook) Handle(ctx context.Context, request types.WebhookHandleRequest) error {
	return ingestGitHubAlert(ctx, request, githubAlertTypeSecretScan)
}

func ingestGitHubAlert(ctx context.Context, request types.WebhookHandleRequest, variant string) error {
	var envelope githubWebhookEnvelope
	if err := jsonx.UnmarshalIfPresent(request.Event.Payload, &envelope); err != nil {
		return err
	}

	resource := githubRepoFromWebhook(envelope.Repository)
	if resource == "" || len(envelope.Alert) == 0 {
		return nil
	}

	return request.Ingest(ctx, []types.IngestPayloadSet{
		{
			Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
			Envelopes: []types.MappingEnvelope{
				{
					Variant:  variant,
					Resource: resource,
					Action:   envelope.Action,
					Payload:  jsonx.CloneRawMessage(envelope.Alert),
				},
			},
		},
	})
}

func githubRepoFromWebhook(repo *githubWebhookRepository) string {
	if repo == nil {
		return ""
	}

	if repo.FullName != "" {
		return repo.FullName
	}

	if repo.Owner.Login != "" && repo.Name != "" {
		return repo.Owner.Login + "/" + repo.Name
	}

	return repo.Name
}
