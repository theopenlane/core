package githubapp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/slacknotify"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/mapx"
)

const (
	// githubSignatureHeader is the HTTP header carrying the HMAC-SHA256 webhook signature
	githubSignatureHeader = "X-Hub-Signature-256"
	// githubWebhookEventHeader is the HTTP header carrying the GitHub event type name
	githubWebhookEventHeader = "X-GitHub-Event"
	// githubWebhookDeliveryHeader is the HTTP header carrying the provider-assigned delivery ID
	githubWebhookDeliveryHeader = "X-GitHub-Delivery"
)

// App executes the GitHub App webhook verification logic
type App struct {
	// Config holds the operator-supplied GitHub App settings
	Config Config
}

// PingWebhook marks a GitHub App webhook endpoint as verified
type PingWebhook struct{}

// InstallationCreatedWebhook sends the GitHub App installation notification
type InstallationCreatedWebhook struct{}

// InstallationDeletedWebhook removes a disconnected GitHub App installation after the uninstall webhook arrives
type InstallationDeletedWebhook struct{}

// DependabotAlertWebhook ingests one Dependabot alert from the webhook payload
type DependabotAlertWebhook struct{}

// CodeScanningAlertWebhook ingests one code scanning alert from the webhook payload
type CodeScanningAlertWebhook struct{}

// SecretScanningAlertWebhook ingests one secret scanning alert from the webhook payload
type SecretScanningAlertWebhook struct{}

// githubWebhookEnvelope is the common wrapper decoded from all inbound GitHub webhook payloads
type githubWebhookEnvelope struct {
	// Action is the event action sub-type (e.g. "created", "dismissed")
	Action string `json:"action"`
	// installation identifies the GitHub App installation that sent the event
	Installation *githubWebhookInstallation `json:"installation"`
	// Repository is the repository the event originated from, if any
	Repository *githubWebhookRepository `json:"repository"`
	// Alert is the raw alert payload for security alert event types
	Alert json.RawMessage `json:"alert"`
}

// githubWebhookInstallation represents the installation object within a GitHub webhook payload
type githubWebhookInstallation struct {
	// ID is the numeric GitHub App installation identifier
	ID int64 `json:"id"`
	// Account is the account (user or organization) that owns the installation
	Account *githubWebhookAccount `json:"account"`
	// TargetType indicates whether the installation target is a user or organization
	TargetType string `json:"target_type"`
}

// githubWebhookAccount represents the account object nested within a GitHub webhook installation
type githubWebhookAccount struct {
	// Login is the GitHub account login name
	Login string `json:"login"`
	// Type is the account type (User or Organization)
	Type string `json:"type"`
}

// githubWebhookRepository represents the repository object within a GitHub webhook payload
type githubWebhookRepository struct {
	// FullName is the slug-style full repository name (owner/repo)
	FullName string `json:"full_name"`
	// Name is the short repository name without the owner prefix
	Name string `json:"name"`
	// HTMLURL is the browser URL for the repository
	HTMLURL string `json:"html_url"`
	// Owner is the repository owner login information
	Owner githubWebhookRepoOwner `json:"owner"`
}

// githubWebhookRepoOwner holds the owner login from a GitHub webhook repository object
type githubWebhookRepoOwner struct {
	// Login is the owner's GitHub login name
	Login string `json:"login"`
}

// githubWebhookVerificationMetadata is the integration metadata patch written on successful ping verification
type githubWebhookVerificationMetadata struct {
	// GitHubWebhookVerifiedAt records the UTC timestamp of the verified ping event
	GitHubWebhookVerifiedAt time.Time `json:"githubWebhookVerifiedAt"`
}

// Verify validates the HMAC-SHA256 signature on an inbound GitHub webhook request.
// Follows the GitHub webhook verification pattern: the X-Hub-Signature-256 header
// contains "sha256=<hex-encoded HMAC-SHA256>" computed using the app webhook secret
func (a App) Verify(request types.WebhookInboundRequest) error {
	if a.Config.WebhookSecret == "" {
		return ErrWebhookSecretMissing
	}

	signature := request.Request.Header.Get(githubSignatureHeader)
	if signature == "" {
		return ErrWebhookSignatureMissing
	}

	sigHex, found := strings.CutPrefix(signature, "sha256=")
	if !found {
		return ErrWebhookSignatureMismatch
	}

	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		return ErrWebhookSignatureMismatch
	}

	mac := hmac.New(sha256.New, []byte(a.Config.WebhookSecret))
	mac.Write(request.Payload)

	if !hmac.Equal(sigBytes, mac.Sum(nil)) {
		return ErrWebhookSignatureMismatch
	}

	return nil
}

// Event resolves the inbound webhook payload into one registered GitHub webhook event
func (App) Event(request types.WebhookInboundRequest) (types.WebhookReceivedEvent, error) {
	eventType := request.Request.Header.Get(githubWebhookEventHeader)
	if eventType == "" {
		return types.WebhookReceivedEvent{}, ErrWebhookEventMissing
	}

	var envelope githubWebhookEnvelope
	if err := jsonx.UnmarshalIfPresent(request.Payload, &envelope); err != nil {
		return types.WebhookReceivedEvent{}, ErrWebhookPayloadInvalid
	}

	name := ""
	switch eventType {
	case "ping":
		name = pingWebhookEvent.Name()
	case "installation":
		switch envelope.Action {
		case "created":
			name = installationCreatedWebhookEvent.Name()
		case "deleted":
			name = installationDeletedWebhookEvent.Name()
		}
	case "dependabot_alert":
		name = dependabotAlertWebhookEvent.Name()
	case "code_scanning_alert":
		name = codeScanningAlertWebhookEvent.Name()
	case "secret_scanning_alert":
		name = secretScanningAlertWebhookEvent.Name()
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
	metadataPatch, err := jsonx.ToMap(githubWebhookVerificationMetadata{
		GitHubWebhookVerifiedAt: time.Now().UTC(),
	})
	if err != nil {
		return ErrWebhookMetadataEncode
	}

	if err := request.DB.Integration.UpdateOneID(request.Integration.ID).
		SetMetadata(mapx.DeepMergeMapAny(mapx.DeepCloneMapAny(request.Integration.Metadata), metadataPatch)).
		Exec(ctx); err != nil {
		return ErrWebhookPersistFailed
	}

	return nil
}

// Handle sends the GitHub App installation-created Slack notification
func (InstallationCreatedWebhook) Handle(ctx context.Context, request types.WebhookHandleRequest) error {
	if !slacknotify.NotificationsEnabled() {
		return nil
	}

	envelope, err := installationCreatedWebhookEvent.UnmarshalPayload(request.Event.Payload)
	if err != nil {
		return ErrWebhookPayloadInvalid
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

// Handle removes the Openlane installation after GitHub confirms the app was uninstalled
func (InstallationDeletedWebhook) Handle(ctx context.Context, request types.WebhookHandleRequest) error {
	if request.CleanupInstallation == nil {
		return nil
	}

	return request.CleanupInstallation(ctx)
}

// Handle ingests one Dependabot alert from the webhook payload
func (DependabotAlertWebhook) Handle(ctx context.Context, request types.WebhookHandleRequest) error {
	envelope, err := dependabotAlertWebhookEvent.UnmarshalPayload(request.Event.Payload)
	if err != nil {
		return ErrWebhookPayloadInvalid
	}

	return ingestGitHubAlert(ctx, request, githubAlertTypeDependabot, envelope)
}

// Handle ingests one code scanning alert from the webhook payload
func (CodeScanningAlertWebhook) Handle(ctx context.Context, request types.WebhookHandleRequest) error {
	envelope, err := codeScanningAlertWebhookEvent.UnmarshalPayload(request.Event.Payload)
	if err != nil {
		return ErrWebhookPayloadInvalid
	}

	return ingestGitHubAlert(ctx, request, githubAlertTypeCodeScanning, envelope)
}

// Handle ingests one secret scanning alert from the webhook payload
func (SecretScanningAlertWebhook) Handle(ctx context.Context, request types.WebhookHandleRequest) error {
	envelope, err := secretScanningAlertWebhookEvent.UnmarshalPayload(request.Event.Payload)
	if err != nil {
		return ErrWebhookPayloadInvalid
	}

	return ingestGitHubAlert(ctx, request, githubAlertTypeSecretScan, envelope)
}

// ingestGitHubAlert extracts the alert from a webhook payload and routes it for ingest
func ingestGitHubAlert(ctx context.Context, request types.WebhookHandleRequest, variant string, envelope githubWebhookEnvelope) error {
	resource := githubRepoFromWebhook(envelope.Repository)
	if resource == "" || len(envelope.Alert) == 0 {
		return nil
	}

	if err := request.Ingest(ctx, []types.IngestPayloadSet{
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
	}); err != nil {
		return ErrWebhookIngestFailed
	}

	return nil
}

// githubRepoFromWebhook extracts the best available repository identifier from a webhook payload
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
