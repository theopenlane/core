package keystore

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	ent "github.com/theopenlane/core/internal/ent/generated"
	hushschema "github.com/theopenlane/core/internal/ent/generated/hush"
	integrationschema "github.com/theopenlane/core/internal/ent/generated/integration"
)

const (
	secretRecordCapacity = 5
)

// Store encapsulates integration persistence against ent.
type Store struct {
	db *ent.Client
}

// NewStore returns a persistence shim backed by ent.
func NewStore(db *ent.Client) *Store {
	return &Store{db: db}
}

// SecretRecord describes a hush secret that should be synced with an integration.
type SecretRecord struct {
	Name        string
	DisplayName string
	Description string
	Kind        string
	Value       string
}

// OAuthTokens describes the payload required to persist OAuth credentials.
type OAuthTokens struct {
	OrgID             string
	Provider          string
	Username          string
	UserID            string
	Email             string
	AccessToken       string
	RefreshToken      string
	ExpiresAt         *time.Time
	StoreRefreshToken bool
	Attributes        map[string]string
	Metadata          map[string]any
	Scopes            []string
}

// SaveRequest captures the data necessary to upsert an integration record.
type SaveRequest struct {
	OrgID                  string
	Provider               string
	IntegrationName        string
	IntegrationDescription string
	Secrets                []SecretRecord
	DeleteSecrets          []string
	Metadata               map[string]any
	Scopes                 []string
}

// TokenBundle represents stored OAuth credentials plus associated metadata.
type TokenBundle struct {
	Integration      *ent.Integration
	AccessToken      string
	RefreshToken     string
	ExpiresAt        *time.Time
	ProviderUserID   string
	ProviderUsername string
	ProviderEmail    string
	Attributes       map[string]string
	Metadata         map[string]any
	Scopes           []string
}

// UpsertIntegration creates or updates an integration and synchronizes secrets.
func (s *Store) UpsertIntegration(ctx context.Context, req SaveRequest) (*ent.Integration, error) {
	integration, err := s.db.Integration.Query().
		Where(integrationschema.And(
			integrationschema.OwnerID(req.OrgID),
			integrationschema.Kind(req.Provider),
		)).
		Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return nil, err
		}
		metadata := mergeMetadata(nil, req.Metadata, req.Scopes)
		create := s.db.Integration.Create().
			SetOwnerID(req.OrgID).
			SetName(req.IntegrationName).
			SetDescription(req.IntegrationDescription).
			SetKind(req.Provider)
		if len(metadata) > 0 {
			create.SetMetadata(metadata)
		}
		integration, err = create.Save(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		metadata := mergeMetadata(integration.Metadata, req.Metadata, req.Scopes)
		update := integration.Update().
			SetName(req.IntegrationName).
			SetDescription(req.IntegrationDescription)
		if len(metadata) > 0 {
			update.SetMetadata(metadata)
		}
		integration, err = update.Save(ctx)
		if err != nil {
			return nil, err
		}
	}

	for _, secret := range req.Secrets {
		if err := s.upsertSecret(ctx, integration, secret); err != nil {
			return nil, err
		}
	}

	for _, name := range req.DeleteSecrets {
		if err := s.deleteSecret(ctx, integration, name); err != nil {
			return nil, err
		}
	}

	return integration, nil
}

// UpsertOAuthTokens persists OAuth credentials and related metadata.
func (s *Store) UpsertOAuthTokens(ctx context.Context, payload OAuthTokens) (*ent.Integration, error) {
	helper := NewHelper(payload.Provider, payload.Username)
	integrationName := helper.Name()
	description := helper.Description()

	secrets := make([]SecretRecord, 0, secretRecordCapacity)
	deleteSecrets := make([]string, 0, 1)

	addSecret := func(field, value string) {
		if value == "" {
			return
		}
		secrets = append(secrets, SecretRecord{
			Name:        helper.SecretName(field),
			DisplayName: helper.SecretDisplayName(integrationName, field),
			Description: helper.SecretDescription(field),
			Kind:        OAuthTokenKind,
			Value:       value,
		})
	}

	addSecret(AccessTokenField, payload.AccessToken)

	if payload.StoreRefreshToken {
		addSecret(RefreshTokenField, payload.RefreshToken)
	} else {
		deleteSecrets = append(deleteSecrets, helper.SecretName(RefreshTokenField))
	}

	if payload.ExpiresAt != nil {
		addSecret(ExpiresAtField, payload.ExpiresAt.Format(time.RFC3339))
	}

	addSecret(ProviderUserIDField, payload.UserID)
	addSecret(ProviderUsernameField, payload.Username)
	addSecret(ProviderEmailField, payload.Email)

	for key, value := range payload.Attributes {
		if value == "" {
			continue
		}
		secrets = append(secrets, SecretRecord{
			Name:        helper.SecretName(key),
			DisplayName: helper.SecretDisplayName(integrationName, key),
			Description: helper.SecretDescription(key),
			Kind:        MetadataKind,
			Value:       value,
		})
	}

	return s.UpsertIntegration(ctx, SaveRequest{
		OrgID:                  payload.OrgID,
		Provider:               payload.Provider,
		IntegrationName:        integrationName,
		IntegrationDescription: description,
		Secrets:                secrets,
		DeleteSecrets:          deleteSecrets,
		Metadata:               mergeMetadata(nil, payload.Metadata, payload.Scopes),
		Scopes:                 payload.Scopes,
	})
}

// LoadTokens fetches integration secrets and returns a token bundle.
func (s *Store) LoadTokens(ctx context.Context, orgID, provider string) (*TokenBundle, error) {
	integration, err := s.db.Integration.Query().
		Where(integrationschema.And(
			integrationschema.OwnerID(orgID),
			integrationschema.Kind(provider),
		)).
		WithSecrets().
		Only(ctx)
	if err != nil {
		return nil, err
	}

	bundle := &TokenBundle{
		Integration: integration,
		Attributes:  make(map[string]string),
	}

	if meta := integration.Metadata; len(meta) > 0 {
		b := cloneMetadata(meta)
		if len(b) > 0 {
			bundle.Metadata = b
		}
		if scopes := stringSliceFromAny(meta["scopes"]); len(scopes) > 0 {
			bundle.Scopes = scopes
		}
	}

	prefix := provider + SecretNameSeparator

	for _, secret := range integration.Edges.Secrets {
		if secret.SecretName == "" {
			continue
		}
		if secret.SecretValue == "" {
			continue
		}

		switch secret.SecretName {
		case prefix + AccessTokenField:
			bundle.AccessToken = secret.SecretValue
		case prefix + RefreshTokenField:
			bundle.RefreshToken = secret.SecretValue
		case prefix + ExpiresAtField:
			if ts, err := time.Parse(time.RFC3339, secret.SecretValue); err == nil {
				bundle.ExpiresAt = &ts
			}
		case prefix + ProviderUserIDField:
			bundle.ProviderUserID = secret.SecretValue
		case prefix + ProviderUsernameField:
			bundle.ProviderUsername = secret.SecretValue
		case prefix + ProviderEmailField:
			bundle.ProviderEmail = secret.SecretValue
		default:
			shortKey := strings.TrimPrefix(secret.SecretName, prefix)
			bundle.Attributes[shortKey] = secret.SecretValue
		}
	}

	return bundle, nil
}

func (s *Store) upsertSecret(ctx context.Context, integration *ent.Integration, secret SecretRecord) error {
	if secret.Name == "" {
		return nil
	}

	if secret.Value == "" {
		return s.deleteSecret(ctx, integration, secret.Name)
	}

	existing, err := s.db.Hush.Query().
		Where(hushschema.And(
			hushschema.OwnerID(integration.OwnerID),
			hushschema.SecretName(secret.Name),
			hushschema.HasIntegrationsWith(integrationschema.ID(integration.ID)),
		)).
		Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return err
	}

	create := func() error {
		mutation := s.db.Hush.Create().
			SetOwnerID(integration.OwnerID).
			SetSecretName(secret.Name).
			SetSecretValue(secret.Value).
			AddIntegrations(integration)
		if secret.DisplayName != "" {
			mutation.SetName(secret.DisplayName)
		}
		if secret.Description != "" {
			mutation.SetDescription(secret.Description)
		}
		if secret.Kind != "" {
			mutation.SetKind(secret.Kind)
		}
		return mutation.Exec(ctx)
	}

	if ent.IsNotFound(err) {
		return create()
	}

	if err := s.db.Hush.DeleteOne(existing).Exec(ctx); err != nil {
		return err
	}
	return create()
}

func (s *Store) deleteSecret(ctx context.Context, integration *ent.Integration, name string) error {
	if name == "" {
		return nil
	}

	existing, err := s.db.Hush.Query().
		Where(hushschema.And(
			hushschema.OwnerID(integration.OwnerID),
			hushschema.SecretName(name),
			hushschema.HasIntegrationsWith(integrationschema.ID(integration.ID)),
		)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return err
	}

	return s.db.Hush.DeleteOne(existing).Exec(ctx)
}

// LoadHushSecret fetches the secret value for the supplied owner/secret name pair.
func (s *Store) LoadHushSecret(ctx context.Context, ownerID, name string) (string, error) {
	if ownerID == "" || strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("load hush secret: ownerID and name are required")
	}

	secret, err := s.db.Hush.Query().
		Where(
			hushschema.OwnerID(ownerID),
			hushschema.SecretName(name),
		).
		Only(ctx)
	if err != nil {
		return "", err
	}
	return secret.SecretValue, nil
}

func mergeMetadata(existing map[string]any, updates map[string]any, scopes []string) map[string]any {
	merged := cloneMetadata(existing)
	if merged == nil {
		merged = make(map[string]any)
	}
	for k, v := range updates {
		if v == nil {
			delete(merged, k)
			continue
		}
		merged[k] = v
	}
	if len(scopes) > 0 {
		merged["scopes"] = uniqueStrings(scopes)
	}
	return merged
}

func cloneMetadata(src map[string]any) map[string]any {
	if len(src) == 0 {
		return nil
	}
	cloned := make(map[string]any, len(src))
	for k, v := range src {
		cloned[k] = v
	}
	return cloned
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]string, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; !ok {
			seen[key] = trimmed
		}
	}
	if len(seen) == 0 {
		return nil
	}
	out := make([]string, 0, len(seen))
	for _, value := range seen {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func stringSliceFromAny(value any) []string {
	switch v := value.(type) {
	case []string:
		return uniqueStrings(v)
	case []any:
		items := make([]string, 0, len(v))
		for _, item := range v {
			items = append(items, fmt.Sprint(item))
		}
		return uniqueStrings(items)
	case string:
		return uniqueStrings([]string{v})
	default:
		if value == nil {
			return nil
		}
		return uniqueStrings([]string{fmt.Sprint(value)})
	}
}
