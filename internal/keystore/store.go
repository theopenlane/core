package keystore

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/samber/lo"
	"golang.org/x/oauth2"

	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/models"
	ent "github.com/theopenlane/ent/generated"
	hushschema "github.com/theopenlane/ent/generated/hush"
	integration "github.com/theopenlane/ent/generated/integration"
	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"
)

// Store persists credential payloads using Ent-backed integrations and hush secrets
type Store struct {
	// db provides access to the Ent ORM client for database operations
	db *ent.Client
	// now returns the current time, overridable for testing
	now func() time.Time
}

// NewStore returns a Store backed by the supplied Ent client
func NewStore(db *ent.Client) *Store {
	return &Store{
		db:  db,
		now: time.Now,
	}
}

// SaveCredential upserts the credential payload for the given org/provider pair
func (s *Store) SaveCredential(ctx context.Context, orgID string, payload types.CredentialPayload) (types.CredentialPayload, error) {
	if payload.Provider == types.ProviderUnknown {
		return types.CredentialPayload{}, ErrProviderRequired
	}

	if strings.TrimSpace(orgID) == "" {
		return types.CredentialPayload{}, ErrOrgIDRequired
	}

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)
	systemCtx = contextx.With(systemCtx, auth.KeyStoreContextKey{})

	integrationRecord, err := s.ensureIntegration(systemCtx, orgID, payload.Provider)
	if err != nil {
		logx.FromContext(systemCtx).Error().Err(err).Msg("failed to ensure integration record")
		return types.CredentialPayload{}, err
	}

	secretName := string(payload.Provider)
	envelope := payloadToCredentialSet(payload, s.now)

	existing, err := s.db.Hush.Query().
		Where(
			hushschema.And(
				hushschema.OwnerID(orgID),
				hushschema.SecretName(secretName),
				hushschema.HasIntegrationsWith(integration.ID(integrationRecord.ID)),
			),
		).
		Only(systemCtx)

	if err != nil {
		if !ent.IsNotFound(err) {
			logx.FromContext(systemCtx).Error().Err(err).Msg("failed to query existing credential")
			return types.CredentialPayload{}, err
		}

		createErr := s.db.Hush.Create().
			SetOwnerID(orgID).
			SetName(secretName).
			SetSecretName(secretName).
			SetKind(string(payload.Kind)).
			SetCredentialSet(envelope).
			AddIntegrations(integrationRecord).
			Exec(systemCtx)
		if createErr != nil {
			logx.FromContext(systemCtx).Error().Err(createErr).Msg("failed to create credential record")
			return types.CredentialPayload{}, createErr
		}
	} else {
		updateErr := existing.Update().
			SetCredentialSet(envelope).
			SetKind(string(payload.Kind)).
			Exec(systemCtx)
		if updateErr != nil {
			logx.FromContext(systemCtx).Error().Err(updateErr).Msg("failed to update credential record")
			return types.CredentialPayload{}, updateErr
		}
	}

	return payload, nil
}

// EnsureIntegration guarantees an integration record exists for the given org/provider pair
func (s *Store) EnsureIntegration(ctx context.Context, orgID string, provider types.ProviderType) (*ent.Integration, error) {
	if s == nil {
		return nil, ErrStoreNotInitialized
	}

	return s.ensureIntegration(ctx, orgID, provider)
}

// LoadCredential retrieves the credential payload for the given org/provider pair
func (s *Store) LoadCredential(ctx context.Context, orgID string, provider types.ProviderType) (types.CredentialPayload, error) {
	if provider == types.ProviderUnknown {
		return types.CredentialPayload{}, ErrProviderRequired
	}

	if strings.TrimSpace(orgID) == "" {
		return types.CredentialPayload{}, ErrOrgIDRequired
	}

	integrationRecord, err := s.db.Integration.Query().Where(
		integration.And(
			integration.OwnerIDEQ(orgID),
			integration.KindEQ(string(provider)),
		),
	).WithSecrets().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return types.CredentialPayload{}, ErrCredentialNotFound
		}

		return types.CredentialPayload{}, err
	}

	secretName := string(provider)

	for _, secret := range integrationRecord.Edges.Secrets {
		if secret.SecretName != secretName {
			continue
		}

		envelope := secret.CredentialSet

		return credentialSetToPayload(provider, envelope)
	}

	return types.CredentialPayload{}, ErrCredentialNotFound
}

// DeleteIntegration removes the integration and associated secrets for the given org
func (s *Store) DeleteIntegration(ctx context.Context, orgID string, integrationID string) (types.ProviderType, string, error) {
	if s == nil || s.db == nil {
		return types.ProviderUnknown, "", ErrStoreNotInitialized
	}

	if strings.TrimSpace(orgID) == "" {
		return types.ProviderUnknown, "", integrations.ErrOrgIDRequired
	}
	if strings.TrimSpace(integrationID) == "" {
		return types.ProviderUnknown, "", integrations.ErrIntegrationIDRequired
	}

	record, err := s.db.Integration.Query().
		Where(
			integration.IDEQ(integrationID),
			integration.OwnerIDEQ(orgID),
		).
		WithSecrets().
		Only(ctx)
	if err != nil {
		return types.ProviderUnknown, "", err
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return types.ProviderUnknown, "", err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}

		if cerr := tx.Commit(); cerr != nil {
			err = cerr
		}
	}()

	if len(record.Edges.Secrets) > 0 {
		secretIDs := make([]string, 0, len(record.Edges.Secrets))
		for _, secret := range record.Edges.Secrets {
			secretIDs = append(secretIDs, secret.ID)
		}

		if _, err = tx.Hush.Delete().Where(hushschema.IDIn(secretIDs...)).Exec(ctx); err != nil {
			return types.ProviderUnknown, "", err
		}
	}

	if err = tx.Integration.DeleteOneID(record.ID).Exec(ctx); err != nil {
		return types.ProviderUnknown, "", err
	}

	return types.ProviderTypeFromString(record.Kind), record.ID, nil
}

// ensureIntegration guarantees an integration record exists for the given org/provider pair
func (s *Store) ensureIntegration(ctx context.Context, orgID string, provider types.ProviderType) (*ent.Integration, error) {
	record, err := s.db.Integration.Query().
		Where(
			integration.And(
				integration.OwnerIDEQ(orgID),
				integration.KindEQ(string(provider)),
			),
		).
		Only(ctx)
	if err == nil {
		return record, nil
	}

	if !ent.IsNotFound(err) {
		return nil, err
	}

	record, createErr := s.db.Integration.Create().
		SetOwnerID(orgID).
		SetKind(string(provider)).
		SetName(string(provider)).
		Save(ctx)
	if createErr != nil {
		return nil, createErr
	}
	return record, nil
}

// payloadToCredentialSet converts a CredentialPayload into a storable CredentialSet
func payloadToCredentialSet(payload types.CredentialPayload, now func() time.Time) models.CredentialSet {
	cloned := payload.Data
	// Copy map to avoid mutating caller
	cloned.ProviderData = lo.Assign(map[string]any{}, payload.Data.ProviderData)

	if payload.Token != nil {
		cloned.OAuthAccessToken = payload.Token.AccessToken
		cloned.OAuthRefreshToken = payload.Token.RefreshToken
		cloned.OAuthTokenType = payload.Token.TokenType
		if !payload.Token.Expiry.IsZero() {
			exp := payload.Token.Expiry.UTC()
			cloned.OAuthExpiry = &exp
		} else {
			cloned.OAuthExpiry = nil
		}
	}

	if payload.Claims != nil {
		if claimsMap, err := claimsToMap(payload.Claims); err == nil {
			cloned.Claims = claimsMap
		}
	}

	if cloned.OAuthExpiry == nil && payload.Token == nil {
		ts := now().UTC()
		cloned.OAuthExpiry = &ts
	}

	return cloned
}

// credentialSetToPayload converts a stored CredentialSet back into a CredentialPayload
func credentialSetToPayload(provider types.ProviderType, set models.CredentialSet) (types.CredentialPayload, error) {
	baseSet := set
	baseSet.OAuthAccessToken = ""
	baseSet.OAuthRefreshToken = ""
	baseSet.OAuthTokenType = ""
	baseSet.OAuthExpiry = nil
	baseSet.Claims = nil

	builder := types.NewCredentialBuilder(provider).
		With(types.WithCredentialSet(baseSet))

	if token := tokenFromSet(set); token != nil {
		builder = builder.With(types.WithOAuthToken(token))
	}

	if claims := claimsFromSet(set.Claims); claims != nil {
		builder = builder.With(types.WithOIDCClaims(claims))
	}

	return builder.Build()
}

// tokenFromSet constructs an oauth2.Token from the CredentialSet if token data is present
func tokenFromSet(set models.CredentialSet) *oauth2.Token {
	if strings.TrimSpace(set.OAuthAccessToken) == "" && strings.TrimSpace(set.OAuthRefreshToken) == "" {
		return nil
	}

	token := &oauth2.Token{
		AccessToken:  set.OAuthAccessToken,
		RefreshToken: set.OAuthRefreshToken,
		TokenType:    set.OAuthTokenType,
	}

	if set.OAuthExpiry != nil {
		token.Expiry = set.OAuthExpiry.UTC()
	}

	return token
}

// claimsToMap serializes OIDC claims into a map for storage
func claimsToMap(claims *oidc.IDTokenClaims) (map[string]any, error) {
	if claims == nil {
		return nil, nil
	}

	bytes, err := json.Marshal(claims)
	if err != nil {
		return nil, err
	}

	var out map[string]any
	if err := json.Unmarshal(bytes, &out); err != nil {
		return nil, err
	}

	return out, nil
}

// claimsFromSet deserializes OIDC claims from a stored map
func claimsFromSet(raw map[string]any) *oidc.IDTokenClaims {
	if len(raw) == 0 {
		return nil
	}

	bytes, err := json.Marshal(raw)
	if err != nil {
		return nil
	}

	var claims oidc.IDTokenClaims
	if err := json.Unmarshal(bytes, &claims); err != nil {
		return nil
	}

	return &claims
}
