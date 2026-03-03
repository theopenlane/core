package keystore

import (
	"context"
	"errors"
	"maps"
	"time"

	"github.com/samber/lo"
	"golang.org/x/oauth2"

	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/internal/ent/generated"
	hushschema "github.com/theopenlane/core/internal/ent/generated/hush"
	integration "github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
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

	if orgID == "" {
		return types.CredentialPayload{}, ErrOrgIDRequired
	}

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)
	systemCtx = auth.WithCaller(systemCtx, auth.NewKeystoreCaller())

	integrationRecord, err := s.ensureIntegration(systemCtx, orgID, payload.Provider)
	if err != nil {
		logx.FromContext(systemCtx).Error().Err(err).Msg("failed to ensure integration record")
		return types.CredentialPayload{}, err
	}

	return s.saveCredentialForIntegrationRecord(systemCtx, orgID, payload, integrationRecord)
}

// SaveCredentialForIntegration upserts credentials for a specific integration record.
func (s *Store) SaveCredentialForIntegration(ctx context.Context, orgID string, integrationID string, payload types.CredentialPayload) (types.CredentialPayload, error) {
	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)
	systemCtx = contextx.With(systemCtx, auth.KeyStoreContextKey{})

	integrationRecord, err := s.db.Integration.Query().
		Where(
			integration.IDEQ(integrationID),
			integration.OwnerIDEQ(orgID),
			integration.KindEQ(string(payload.Provider)),
		).
		Only(systemCtx)
	if err != nil {
		logx.FromContext(systemCtx).Error().Err(err).Str("integration_id", integrationID).Msg("failed to load integration record for credential save")
		return types.CredentialPayload{}, err
	}

	return s.saveCredentialForIntegrationRecord(systemCtx, orgID, payload, integrationRecord)
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

	if orgID == "" {
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

	return s.loadCredentialFromIntegrationRecord(provider, integrationRecord)
}

// LoadCredentialForIntegration retrieves the credential payload for a specific integration record.
func (s *Store) LoadCredentialForIntegration(ctx context.Context, orgID string, provider types.ProviderType, integrationID string) (types.CredentialPayload, error) {
	integrationRecord, err := s.integrationRecordWithSecrets(ctx, orgID, provider, integrationID)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	return s.loadCredentialFromIntegrationRecord(provider, integrationRecord)
}

func (s *Store) loadCredentialSubjectForIntegration(ctx context.Context, orgID string, provider types.ProviderType, integrationID string) (types.CredentialPayload, bool, error) {
	integrationRecord, err := s.integrationRecordWithSecrets(ctx, orgID, provider, integrationID)
	if err != nil {
		return types.CredentialPayload{}, false, err
	}

	payload, err := s.loadCredentialFromIntegrationRecord(provider, integrationRecord)
	if err == nil {
		return payload, true, nil
	}
	if !errors.Is(err, ErrCredentialNotFound) {
		return types.CredentialPayload{}, false, err
	}

	providerState := integrationRecord.ProviderState

	return types.CredentialPayload{
		Provider:      provider,
		Kind:          types.CredentialKindMetadata,
		ProviderState: &providerState,
	}, false, nil
}

// DeleteIntegration removes the integration and associated secrets for the given org
func (s *Store) DeleteIntegration(ctx context.Context, orgID string, integrationID string) (types.ProviderType, string, error) {
	if s == nil || s.db == nil {
		return types.ProviderUnknown, "", ErrStoreNotInitialized
	}

	if orgID == "" {
		return types.ProviderUnknown, "", integrations.ErrOrgIDRequired
	}
	if integrationID == "" {
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
		secretIDs := lo.Map(record.Edges.Secrets, func(s *ent.Hush, _ int) string { return s.ID })
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

func (s *Store) saveCredentialForIntegrationRecord(ctx context.Context, orgID string, payload types.CredentialPayload, integrationRecord *ent.Integration) (types.CredentialPayload, error) {
	if integrationRecord == nil {
		return types.CredentialPayload{}, ErrCredentialNotFound
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
		Only(ctx)

	if err != nil {
		if !ent.IsNotFound(err) {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to query existing credential")
			return types.CredentialPayload{}, err
		}

		createErr := s.db.Hush.Create().
			SetOwnerID(orgID).
			SetName(secretName).
			SetSecretName(secretName).
			SetKind(string(payload.Kind)).
			SetCredentialSet(envelope).
			AddIntegrations(integrationRecord).
			Exec(ctx)
		if createErr != nil {
			logx.FromContext(ctx).Error().Err(createErr).Msg("failed to create credential record")
			return types.CredentialPayload{}, createErr
		}
	} else {
		updateErr := existing.Update().
			SetCredentialSet(envelope).
			SetKind(string(payload.Kind)).
			Exec(ctx)
		if updateErr != nil {
			logx.FromContext(ctx).Error().Err(updateErr).Msg("failed to update credential record")
			return types.CredentialPayload{}, updateErr
		}
	}

	return payload, nil
}

func (s *Store) loadCredentialFromIntegrationRecord(provider types.ProviderType, integrationRecord *ent.Integration) (types.CredentialPayload, error) {
	if integrationRecord == nil {
		return types.CredentialPayload{}, ErrCredentialNotFound
	}

	secretName := string(provider)
	for _, secret := range integrationRecord.Edges.Secrets {
		if secret.SecretName != secretName {
			continue
		}

		envelope := secret.CredentialSet

		payload, err := credentialSetToPayload(provider, envelope)
		if err != nil {
			return types.CredentialPayload{}, err
		}
		providerState := integrationRecord.ProviderState
		payload.ProviderState = &providerState

		return payload, nil
	}

	return types.CredentialPayload{}, ErrCredentialNotFound
}

func (s *Store) integrationRecordWithSecrets(ctx context.Context, orgID string, provider types.ProviderType, integrationID string) (*ent.Integration, error) {
	integrationRecord, err := s.db.Integration.Query().
		Where(
			integration.IDEQ(integrationID),
			integration.OwnerIDEQ(orgID),
			integration.KindEQ(string(provider)),
		).
		WithSecrets().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrCredentialNotFound
		}

		return nil, err
	}

	return integrationRecord, nil
}

// payloadToCredentialSet converts a CredentialPayload into a storable CredentialSet
func payloadToCredentialSet(payload types.CredentialPayload, now func() time.Time) models.CredentialSet {
	cloned := payload.Data
	// Copy map to avoid mutating caller
	cloned.ProviderData = maps.Clone(payload.Data.ProviderData)

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
	if set.OAuthAccessToken == "" && set.OAuthRefreshToken == "" {
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

	out, err := jsonx.ToMap(claims)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// claimsFromSet deserializes OIDC claims from a stored map
func claimsFromSet(raw map[string]any) *oidc.IDTokenClaims {
	if len(raw) == 0 {
		return nil
	}

	var claims oidc.IDTokenClaims
	if err := jsonx.RoundTrip(raw, &claims); err != nil {
		return nil
	}

	return &claims
}
