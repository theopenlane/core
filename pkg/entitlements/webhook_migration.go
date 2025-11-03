package entitlements

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/stripe/stripe-go/v83"
)

const (
	versionQueryParam = "stripe_api_version"
	migrationStateNew = "new"
	migrationStateOld = "old"
)

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
	// MigrationStageTransitioned indicates new webhook is active, old is disabled
	MigrationStageTransitioned MigrationStage = "transitioned"
	// MigrationStageComplete indicates old webhook can be safely deleted
	MigrationStageComplete MigrationStage = "complete"
)

// GetWebhookMigrationState analyzes the current webhook configuration and returns migration state
func (sc *StripeClient) GetWebhookMigrationState(ctx context.Context, baseURL string) (*WebhookMigrationState, error) {
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

	for _, endpoint := range endpoints {
		endpointURLClean := cleanURL(endpoint.URL)

		if endpointURLClean == baseURLClean {
			versionParam := extractVersionParam(endpoint.URL)
			switch versionParam {
			case "", migrationStateOld:
				state.OldWebhook = endpoint
			case migrationStateNew:
				state.NewWebhook = endpoint
			}
		}
	}

	state.MigrationStage = determineMigrationStage(state)
	state.CanMigrate = canProceedWithMigration(state)

	return state, nil
}

// CreateNewWebhookForMigration creates a new webhook endpoint with the current SDK API version
func (sc *StripeClient) CreateNewWebhookForMigration(ctx context.Context, baseURL string, events []string) (*stripe.WebhookEndpoint, error) {
	state, err := sc.GetWebhookMigrationState(ctx, baseURL)
	if err != nil {
		return nil, err
	}

	if state.NewWebhook != nil {
		return nil, ErrNewWebhookAlreadyExists
	}

	if state.OldWebhook == nil {
		return nil, ErrOldWebhookNotFound
	}

	if len(events) == 0 {
		events = state.OldWebhook.EnabledEvents
	}

	newURL := addVersionParam(baseURL, migrationStateNew)

	params := &stripe.WebhookEndpointCreateParams{
		URL:           stripe.String(newURL),
		EnabledEvents: stripe.StringSlice(events),
		APIVersion:    stripe.String(stripe.APIVersion),
	}

	endpoint, err := sc.Client.V1WebhookEndpoints.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	_, err = sc.DisableWebhookEndpoint(ctx, endpoint.ID)
	if err != nil {
		return nil, err
	}

	return endpoint, nil
}

// EnableNewWebhook enables the new webhook endpoint to begin dual processing
func (sc *StripeClient) EnableNewWebhook(ctx context.Context, baseURL string) (*stripe.WebhookEndpoint, error) {
	state, err := sc.GetWebhookMigrationState(ctx, baseURL)
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

// DisableOldWebhook disables the old webhook endpoint to complete the migration
func (sc *StripeClient) DisableOldWebhook(ctx context.Context, baseURL string) (*stripe.WebhookEndpoint, error) {
	state, err := sc.GetWebhookMigrationState(ctx, baseURL)
	if err != nil {
		return nil, err
	}

	if state.OldWebhook == nil {
		return nil, ErrOldWebhookNotFound
	}

	if state.MigrationStage != string(MigrationStageDualProcessing) {
		return nil, fmt.Errorf("%w: expected stage %s, got %s", ErrInvalidMigrationState, MigrationStageDualProcessing, state.MigrationStage)
	}

	return sc.DisableWebhookEndpoint(ctx, state.OldWebhook.ID)
}

// RollbackMigration rolls back the migration by disabling the new webhook and re-enabling the old one
func (sc *StripeClient) RollbackMigration(ctx context.Context, baseURL string) error {
	state, err := sc.GetWebhookMigrationState(ctx, baseURL)
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

// CleanupOldWebhook deletes the old webhook endpoint after successful migration
func (sc *StripeClient) CleanupOldWebhook(ctx context.Context, baseURL string) (*stripe.WebhookEndpoint, error) {
	state, err := sc.GetWebhookMigrationState(ctx, baseURL)
	if err != nil {
		return nil, err
	}

	if state.OldWebhook == nil {
		return nil, ErrOldWebhookNotFound
	}

	if state.MigrationStage != string(MigrationStageTransitioned) && state.MigrationStage != string(MigrationStageComplete) {
		return nil, fmt.Errorf("%w: cannot cleanup in stage %s", ErrInvalidMigrationState, state.MigrationStage)
	}

	return sc.DeleteWebhookEndpoint(ctx, state.OldWebhook.ID)
}

// PromoteNewWebhook removes the version query parameter from the new webhook URL to make it the primary webhook
func (sc *StripeClient) PromoteNewWebhook(ctx context.Context, baseURL string) (*stripe.WebhookEndpoint, error) {
	state, err := sc.GetWebhookMigrationState(ctx, baseURL)
	if err != nil {
		return nil, err
	}

	if state.NewWebhook == nil {
		return nil, ErrWebhookNotFound
	}

	if state.OldWebhook != nil {
		return nil, fmt.Errorf("%w: old webhook must be deleted before promotion", ErrInvalidMigrationState)
	}

	oldURL := addVersionParam(baseURL, migrationStateOld)

	params := &stripe.WebhookEndpointUpdateParams{
		URL: stripe.String(oldURL),
	}

	return sc.UpdateWebhookEndpoint(ctx, state.NewWebhook.ID, params)
}

func cleanURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	parsedURL.RawQuery = ""
	return parsedURL.String()
}

func extractVersionParam(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return parsedURL.Query().Get(versionQueryParam)
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
		if state.OldWebhook.APIVersion == state.CurrentSDKVersion {
			return string(MigrationStageNone)
		}
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
			return string(MigrationStageTransitioned)
		}

	case state.OldWebhook == nil && state.NewWebhook != nil:
		return string(MigrationStageComplete)
	}

	return string(MigrationStageNone)
}

func canProceedWithMigration(state *WebhookMigrationState) bool {
	stage := MigrationStage(state.MigrationStage)

	switch stage {
	case MigrationStageReady, MigrationStageNewCreated, MigrationStageDualProcessing, MigrationStageTransitioned:
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
		return "Create new webhook endpoint with current SDK API version."
	case MigrationStageNewCreated:
		return "Enable new webhook endpoint to begin dual processing."
	case MigrationStageDualProcessing:
		return "Update code to process new version events. Then disable old webhook."
	case MigrationStageTransitioned:
		return "Monitor new webhook. When confident, cleanup old webhook."
	case MigrationStageComplete:
		return "Migration complete. Optionally promote new webhook to remove version query parameter."
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
