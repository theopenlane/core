// Package store wraps persistence for integrations and hush secrets.
package store

import "context"

// Repository exposes typed helpers for integration metadata and secret storage.
type Repository interface {
	// Integration fetches the persisted integration metadata for the org/provider pair.
	Integration(ctx context.Context, orgID, integrationID string) (IntegrationRecord, error)

	// Credentials fetches the encrypted credential material for the integration.
	Credentials(ctx context.Context, orgID, integrationID string) (CredentialRecord, error)

	// SaveCredentials upserts encrypted credential material for the integration session.
	SaveCredentials(ctx context.Context, cred CredentialRecord) error

	// DeleteCredentials removes credential material associated with the integration.
	DeleteCredentials(ctx context.Context, orgID, integrationID string) error
}

// IntegrationRecord captures the persisted integration configuration.
type IntegrationRecord struct {
	OrgID         string
	IntegrationID string
	Provider      string
	TenantID      string
	Metadata      map[string]any
	Scopes        []string
}

// CredentialRecord stores the hush-backed credential payload and related metadata.
type CredentialRecord struct {
	OrgID         string
	IntegrationID string
	Provider      string
	Encrypted     []byte
	Version       int
}
