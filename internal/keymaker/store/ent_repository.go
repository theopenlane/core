package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	ent "github.com/theopenlane/core/internal/ent/generated"
	hushschema "github.com/theopenlane/core/internal/ent/generated/hush"
	integrationSchema "github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/keystore"
)

// EntRepository implements Repository backed by ent.
type EntRepository struct {
	db *ent.Client
}

// NewEntRepository creates a Repository backed by ent.
func NewEntRepository(db *ent.Client) *EntRepository {
	return &EntRepository{db: db}
}

// Integration fetches the persisted integration metadata.
func (r *EntRepository) Integration(ctx context.Context, orgID, integrationID string) (IntegrationRecord, error) {
	if orgID == "" || integrationID == "" {
		return IntegrationRecord{}, ErrMissingIdentifiers
	}

	integration, err := r.db.Integration.Query().
		Where(integrationSchema.And(
			integrationSchema.OwnerID(orgID),
			integrationSchema.ID(integrationID),
		)).
		Only(ctx)
	if err != nil {
		return IntegrationRecord{}, err
	}

	record := IntegrationRecord{
		OrgID:         integration.OwnerID,
		IntegrationID: integration.ID,
		Provider:      integration.Kind,
		Metadata:      integration.Metadata,
	}

	if tenantRaw, ok := integration.Metadata["tenantId"]; ok {
		record.TenantID = fmt.Sprint(tenantRaw)
	}

	if scopes, ok := integration.Metadata["scopes"]; ok {
		switch v := scopes.(type) {
		case []any:
			record.Scopes = make([]string, 0, len(v))
			for _, item := range v {
				record.Scopes = append(record.Scopes, fmt.Sprint(item))
			}
		case []string:
			record.Scopes = append(record.Scopes, v...)
		case string:
			record.Scopes = append(record.Scopes, strings.Split(v, " ")...)
		}
	}

	return record, nil
}

// Credentials loads hush-backed credentials and serializes them for keymaker consumption.
func (r *EntRepository) Credentials(ctx context.Context, orgID, integrationID string) (CredentialRecord, error) {
	if orgID == "" || integrationID == "" {
		return CredentialRecord{}, ErrMissingIdentifiers
	}

	integration, err := r.db.Integration.Query().
		Where(integrationSchema.And(
			integrationSchema.OwnerID(orgID),
			integrationSchema.ID(integrationID),
		)).
		WithSecrets().
		Only(ctx)
	if err != nil {
		return CredentialRecord{}, err
	}

	payload := make(map[string]string, len(integration.Edges.Secrets))
	for _, secret := range integration.Edges.Secrets {
		payload[secret.SecretName] = secret.SecretValue
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return CredentialRecord{}, err
	}

	return CredentialRecord{
		OrgID:         orgID,
		IntegrationID: integrationID,
		Provider:      integration.Kind,
		Encrypted:     encoded,
		Version:       len(payload),
	}, nil
}

// SaveCredentials writes credential material to hush.
func (r *EntRepository) SaveCredentials(ctx context.Context, cred CredentialRecord) error {
	if cred.OrgID == "" || cred.IntegrationID == "" {
		return ErrMissingIdentifiers
	}

	integration, err := r.db.Integration.Query().
		Where(integrationSchema.And(
			integrationSchema.OwnerID(cred.OrgID),
			integrationSchema.ID(cred.IntegrationID),
		)).
		Only(ctx)
	if err != nil {
		return err
	}

	var payload map[string]string
	if len(cred.Encrypted) > 0 {
		if err := json.Unmarshal(cred.Encrypted, &payload); err != nil {
			return err
		}
	}

	for name, value := range payload {
		if err := r.upsertSecret(ctx, integration, name, value); err != nil {
			return err
		}
	}

	return nil
}

// DeleteCredentials removes stored credentials for an integration.
func (r *EntRepository) DeleteCredentials(ctx context.Context, orgID, integrationID string) error {
	if orgID == "" || integrationID == "" {
		return ErrMissingIdentifiers
	}

	integration, err := r.db.Integration.Query().
		Where(integrationSchema.And(
			integrationSchema.OwnerID(orgID),
			integrationSchema.ID(integrationID),
		)).
		WithSecrets().
		Only(ctx)
	if err != nil {
		return err
	}

	for _, secret := range integration.Edges.Secrets {
		if err := r.db.Hush.DeleteOne(secret).Exec(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (r *EntRepository) upsertSecret(ctx context.Context, integration *ent.Integration, name, value string) error {
	if name == "" {
		return nil
	}
	if value == "" {
		return r.deleteSecret(ctx, integration, name)
	}

	existing, err := r.db.Hush.Query().
		Where(hushschema.And(
			hushschema.OwnerID(integration.OwnerID),
			hushschema.SecretName(name),
			hushschema.HasIntegrationsWith(integrationSchema.ID(integration.ID)),
		)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return err
	}

	if ent.IsNotFound(err) {
		return r.db.Hush.Create().
			SetOwnerID(integration.OwnerID).
			SetSecretName(name).
			SetSecretValue(value).
			AddIntegrations(integration).
			Exec(ctx)
	}

	if err := r.db.Hush.DeleteOne(existing).Exec(ctx); err != nil {
		return err
	}

	return r.db.Hush.Create().
		SetOwnerID(integration.OwnerID).
		SetSecretName(name).
		SetSecretValue(value).
		AddIntegrations(integration).
		Exec(ctx)
}

func (r *EntRepository) deleteSecret(ctx context.Context, integration *ent.Integration, name string) error {
	if name == "" {
		return nil
	}
	_, err := r.db.Hush.
		Delete().
		Where(hushschema.And(
			hushschema.OwnerID(integration.OwnerID),
			hushschema.SecretName(name),
			hushschema.HasIntegrationsWith(integrationSchema.ID(integration.ID)),
		)).
		Exec(ctx)
	return err
}

// EncodeTokens marshals a token bundle into the credential format understood by keymaker.
func EncodeTokens(provider string, bundle tokensBundle) ([]byte, error) {
	prefix := provider + keystore.SecretNameSeparator
	payload := map[string]string{
		prefix + keystore.AccessTokenField:  bundle.AccessToken,
		prefix + keystore.RefreshTokenField: bundle.RefreshToken,
	}
	for key, value := range bundle.Attributes {
		payload[prefix+key] = value
	}
	return json.Marshal(payload)
}

type tokensBundle struct {
	AccessToken  string
	RefreshToken string
	Attributes   map[string]string
}
