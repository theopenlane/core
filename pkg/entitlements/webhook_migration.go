package entitlements

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/stripe/stripe-go/v83"
)

const (
	versionQueryParam = "api_version"
	migrationStateNew = "new"
	migrationStateOld = "old"
)

type webhookMigrationOptions struct {
	CurrentVersion string
}

// WebhookMigrationOption is a functional option for GetWebhookMigrationState
type WebhookMigrationOption func(*webhookMigrationOptions)

// WithCurrentVersion sets the current API version from config
func WithCurrentVersion(version string) WebhookMigrationOption {
	return func(o *webhookMigrationOptions) {
		o.CurrentVersion = version
	}
}

// WebhookMigrationState represents the current state of webhook migration
type WebhookMigrationState struct {
	OldWebhook        *stripe.WebhookEndpoint
	NewWebhook        *stripe.WebhookEndpoint
	CurrentSDKVersion string
	CanMigrate        bool
	MigrationStage    string
}

// MigrationStage represents the stages of webhook migration
type MigrationStage string

const (
	// MigrationStageNone indicates no migration is in progress
	MigrationStageNone MigrationStage = "none"
	// MigrationStageReady indicates ready to start migration
	MigrationStageReady MigrationStage = "ready"
	// MigrationStageNewCreated indicates new webhook has been created but disabled
	MigrationStageNewCreated MigrationStage = "new_created"
	// MigrationStageDualProcessing indicates both webhooks are enabled
	MigrationStageDualProcessing MigrationStage = "dual_processing"
	// MigrationStageComplete indicates migration is complete, only new webhook is active
	MigrationStageComplete MigrationStage = "complete"
)

// GetWebhookMigrationState analyzes the current webhook configuration and returns migration state
func (sc *StripeClient) GetWebhookMigrationState(ctx context.Context, baseURL string, opts ...WebhookMigrationOption) (*WebhookMigrationState, error) {
	options := &webhookMigrationOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if baseURL == "" {
		return nil, ErrWebhookURLRequired
	}

	endpoints, err := sc.ListWebhookEndpoints(ctx)
	if err != nil {
		return nil, err
	}

	state := &WebhookMigrationState{
		CurrentSDKVersion: stripe.APIVersion,
		CanMigrate:        false,
		MigrationStage:    string(MigrationStageNone),
	}

	baseURLClean := cleanURL(baseURL)

	configCurrentVersion := options.CurrentVersion

	if configCurrentVersion == "" && sc.Config != nil {
		configCurrentVersion = sc.Config.StripeWebhookAPIVersion
	}

	for _, endpoint := range endpoints {
		endpointURLClean := cleanURL(endpoint.URL)

		if endpointURLClean == baseURLClean {
			if endpoint.APIVersion == configCurrentVersion {
				state.NewWebhook = endpoint
			} else if endpoint.Status == "enabled" {
				state.OldWebhook = endpoint
			}
		}
	}

	state.MigrationStage = determineMigrationStage(state)
	state.CanMigrate = canProceedWithMigration(state)

	return state, nil
}

// CreateNewWebhookForMigration creates a new webhook endpoint with the provided API version
func (sc *StripeClient) CreateNewWebhookForMigration(ctx context.Context, baseURL string, events []string, apiVersion string) (*stripe.WebhookEndpoint, error) {
	if baseURL == "" {
		return nil, ErrWebhookURLRequired
	}

	if apiVersion == "" {
		return nil, ErrAPIVersionRequired
	}

	endpoints, err := sc.ListWebhookEndpoints(ctx)
	if err != nil {
		return nil, err
	}

	baseURLClean := cleanURL(baseURL)

	for _, endpoint := range endpoints {
		endpointURLClean := cleanURL(endpoint.URL)
		if endpointURLClean == baseURLClean && endpoint.APIVersion == apiVersion {
			return nil, fmt.Errorf("webhook with version %s already exists: %w", apiVersion, ErrNewWebhookAlreadyExists)
		}
	}

	if len(events) == 0 {
		for _, endpoint := range endpoints {
			endpointURLClean := cleanURL(endpoint.URL)
			if endpointURLClean == baseURLClean && endpoint.Status == "enabled" {
				events = endpoint.EnabledEvents
				break
			}
		}
	}

	if len(events) == 0 {
		events = SupportedEventTypeStrings()
	}

	return sc.CreateWebhookEndpoint(ctx, baseURL, events, apiVersion, false)
}

// EnableNewWebhook enables the new webhook endpoint to begin dual processing
func (sc *StripeClient) EnableNewWebhook(ctx context.Context, baseURL string, opts ...WebhookMigrationOption) (*stripe.WebhookEndpoint, error) {
	state, err := sc.GetWebhookMigrationState(ctx, baseURL, opts...)
	if err != nil {
		return nil, err
	}

	if state.NewWebhook == nil {
		return nil, ErrWebhookNotFound
	}

	if state.MigrationStage != string(MigrationStageNewCreated) {
		return nil, fmt.Errorf("%w: expected stage %s, got %s", ErrInvalidMigrationState, MigrationStageNewCreated, state.MigrationStage)
	}

	return sc.EnableWebhookEndpoint(ctx, state.NewWebhook.ID)
}

// DisableWebhookByVersion disables the webhook endpoint matching the specified API version
func (sc *StripeClient) DisableWebhookByVersion(ctx context.Context, baseURL string, apiVersion string) (*stripe.WebhookEndpoint, error) {
	if baseURL == "" {
		return nil, ErrWebhookURLRequired
	}

	if apiVersion == "" {
		return nil, ErrAPIVersionRequired
	}

	endpoints, err := sc.ListWebhookEndpoints(ctx)
	if err != nil {
		return nil, err
	}

	baseURLClean := cleanURL(baseURL)

	for _, endpoint := range endpoints {
		endpointURLClean := cleanURL(endpoint.URL)
		if endpointURLClean == baseURLClean && endpoint.APIVersion == apiVersion && endpoint.Status == "enabled" {
			return sc.DisableWebhookEndpoint(ctx, endpoint.ID)
		}
	}

	return nil, fmt.Errorf("no enabled webhook found with version %s: %w", apiVersion, ErrEnabledWebhookNotFoundByVersion)
}

// RollbackMigration rolls back the migration by disabling the new webhook and re-enabling the old one
func (sc *StripeClient) RollbackMigration(ctx context.Context, baseURL string, opts ...WebhookMigrationOption) error {
	state, err := sc.GetWebhookMigrationState(ctx, baseURL, opts...)
	if err != nil {
		return err
	}

	if state.NewWebhook != nil && state.NewWebhook.Status != "disabled" {
		_, err = sc.DisableWebhookEndpoint(ctx, state.NewWebhook.ID)
		if err != nil {
			return err
		}
	}

	if state.OldWebhook != nil && state.OldWebhook.Status == "disabled" {
		_, err = sc.EnableWebhookEndpoint(ctx, state.OldWebhook.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func cleanURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	parsedURL.RawQuery = ""
	return parsedURL.String()
}

func addVersionParam(baseURL, version string) string {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	q := parsedURL.Query()
	q.Set(versionQueryParam, version)
	parsedURL.RawQuery = q.Encode()

	return parsedURL.String()
}

func determineMigrationStage(state *WebhookMigrationState) string {
	switch {
	case state.OldWebhook == nil && state.NewWebhook == nil:
		return string(MigrationStageNone)

	case state.OldWebhook != nil && state.NewWebhook == nil:
		return string(MigrationStageReady)

	case state.OldWebhook != nil && state.NewWebhook != nil:
		oldEnabled := state.OldWebhook.Status == "enabled"
		newEnabled := state.NewWebhook.Status == "enabled"

		switch {
		case !newEnabled:
			return string(MigrationStageNewCreated)
		case oldEnabled && newEnabled:
			return string(MigrationStageDualProcessing)
		case !oldEnabled && newEnabled:
			return string(MigrationStageComplete)
		}

	case state.OldWebhook == nil && state.NewWebhook != nil:
		return string(MigrationStageComplete)
	}

	return string(MigrationStageNone)
}

func canProceedWithMigration(state *WebhookMigrationState) bool {
	stage := MigrationStage(state.MigrationStage)

	switch stage {
	case MigrationStageReady, MigrationStageNewCreated, MigrationStageDualProcessing:
		return true
	default:
		return false
	}
}

// GetNextMigrationAction returns a description of the next action to take in the migration
func GetNextMigrationAction(stage string) string {
	switch MigrationStage(stage) {
	case MigrationStageNone:
		return "No migration needed. Webhook is already at current SDK version."
	case MigrationStageReady:
		return "Create new webhook endpoint and enable for dual processing."
	case MigrationStageNewCreated:
		return "New webhook created but not yet enabled. Run migrate create again or enable manually."
	case MigrationStageDualProcessing:
		return "Deploy code that accepts both API versions. Then disable old webhook."
	case MigrationStageComplete:
		return "Migration complete. Only the new webhook is active."
	default:
		return "Unknown migration stage."
	}
}

// CompareWebhookEvents compares two sets of webhook events and returns true if they match
func CompareWebhookEvents(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]bool, len(a))
	for _, event := range a {
		aMap[strings.ToLower(event)] = true
	}

	for _, event := range b {
		if !aMap[strings.ToLower(event)] {
			return false
		}
	}

	return true
}
