package oidclocal

import (
	"time"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the local OIDC integration definition
	definitionID = types.NewDefinitionRef("def_01K0OIDCLOCAL000000000000001")
	// oidcConnection is the connection handle for the auth-managed local OIDC credential
	oidcConnection = types.ConnectionOf[oidcLocalCred]()
	// installation is the typed installation metadata handle for the local OIDC definition
	installation = types.InstallationOf[InstallationMetadata]()
)

// Client carries the stored credential the local OIDC operations read, since the definition has no provider client
type Client struct {
	// Credential is the decoded auth-managed credential
	Credential oidcLocalCred
}

// oidcLocalCred holds the provider-owned credential material for a local OIDC installation
type oidcLocalCred struct {
	// AccessToken is the OAuth2 access token
	AccessToken string `json:"accessToken"`
	// RefreshToken is the OAuth2 refresh token
	RefreshToken string `json:"refreshToken,omitempty"`
	// Expiry is the token expiration time
	Expiry *time.Time `json:"expiry,omitempty"`
	// Issuer is the OIDC issuer claim
	Issuer string `json:"issuer,omitempty"`
	// Subject is the OIDC subject claim
	Subject string `json:"subject,omitempty"`
	// Email is the OIDC email claim when present
	Email string `json:"email,omitempty"`
	// Name is the OIDC display name claim when present
	Name string `json:"name,omitempty"`
	// PreferredUsername is the OIDC preferred_username claim when present
	PreferredUsername string `json:"preferredUsername,omitempty"`
	// Groups captures any group names emitted in the ID token claims
	Groups []string `json:"groups,omitempty"`
	// Claims keeps the decoded OIDC claims for inspection during local testing
	Claims map[string]any `json:"claims,omitempty"`
}

// InstallationMetadata holds the stable OIDC identity selected for one local test installation
type InstallationMetadata struct {
	// Issuer is the OIDC issuer claim
	Issuer string `json:"issuer,omitempty" jsonschema:"title=Issuer"`
	// Subject is the OIDC subject claim
	Subject string `json:"subject,omitempty" jsonschema:"title=Subject"`
	// Email is the OIDC email claim
	Email string `json:"email,omitempty" jsonschema:"title=Email"`
	// PreferredUsername is the OIDC preferred_username claim
	PreferredUsername string `json:"preferredUsername,omitempty" jsonschema:"title=Preferred Username"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	externalName := m.PreferredUsername
	if externalName == "" {
		externalName = m.Email
	}

	return types.IntegrationInstallationIdentity{
		ExternalID:   m.Subject,
		ExternalName: externalName,
	}
}

// ClaimsInspect returns the stored OIDC claims for inspection
type ClaimsInspect struct {
	// Claims is the full set of OIDC claims stored in the auth-managed credential
	Claims map[string]any `json:"claims,omitempty"`
}
