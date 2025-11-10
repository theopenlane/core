package keystore

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	// IntegrationNameSuffix is appended to provider names for integration display names.
	IntegrationNameSuffix = " Integration" //nolint:gosec
	// IntegrationDescPrefix prefixes integration descriptions.
	IntegrationDescPrefix = "OAuth integration with "

	// SecretNameSeparator joins provider and field in secret names.
	SecretNameSeparator = "_" //nolint:gosec
	// SecretDescSeparator replaces separator characters for display strings.
	SecretDescSeparator = " "
	// SecretDescSuffix formats human-readable secret descriptions.
	SecretDescSuffix = " for %s integration" //nolint:gosec

	// Secret field identifiers used across integrations.
	AccessTokenField      = "access_token"
	RefreshTokenField     = "refresh_token"
	ExpiresAtField        = "expires_at"
	ProviderUserIDField   = "provider_user_id"
	ProviderUsernameField = "provider_username"
	ProviderEmailField    = "provider_email"

	// OAuthTokenKind is the ent hush kind used for OAuth secrets.
	OAuthTokenKind = "oauth_token" //nolint:gosec
	// MetadataKind is used for non-secret integration metadata persisted in hush.
	MetadataKind = "integration_metadata"
)

// Helper provides common naming utilities for integrations.
type Helper struct {
	provider string
	username string
}

// NewHelper returns a helper for a provider/username combination.
func NewHelper(provider, username string) *Helper {
	return &Helper{
		provider: strings.ToLower(provider),
		username: username,
	}
}

// Provider returns the normalized provider identifier.
func (h *Helper) Provider() string {
	return h.provider
}

// TitleCase returns a title-cased string using English rules.
func TitleCase(value string) string {
	return cases.Title(language.English).String(value)
}

// Name builds the integration display name.
func (h *Helper) Name() string {
	name := TitleCase(h.provider) + IntegrationNameSuffix
	if h.username != "" {
		return name + " (" + h.username + ")"
	}
	return name
}

// Description builds the integration description.
func (h *Helper) Description() string {
	desc := IntegrationDescPrefix + h.provider
	if h.username != "" {
		desc += " for " + h.username
	}
	return desc
}

// SecretName computes the hush secret name for a field.
func (h *Helper) SecretName(field string) string {
	return h.provider + SecretNameSeparator + field
}

// SecretDisplayName returns the display name for a secret.
func (h *Helper) SecretDisplayName(integrationName, field string) string {
	return integrationName + SecretDescSeparator + strings.ReplaceAll(field, SecretNameSeparator, SecretDescSeparator)
}

// SecretDescription returns a readable description for a secret.
func (h *Helper) SecretDescription(field string) string {
	displayName := TitleCase(strings.ReplaceAll(field, SecretNameSeparator, SecretDescSeparator))
	return fmt.Sprintf(displayName+SecretDescSuffix, h.provider)
}
