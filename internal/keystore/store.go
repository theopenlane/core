package keystore

import (
	"context"
	"errors"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/internal/ent/generated"
	hushschema "github.com/theopenlane/core/internal/ent/generated/hush"
	integration "github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	integrationstate "github.com/theopenlane/core/internal/integrations/state"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
)

const providerIntegrationLookupLimit = 2

// Store persists credential sets using Ent-backed integrations and hush secrets.
type Store struct {
	// db provides access to the Ent ORM client for database operations.
	db *ent.Client
}

// NewStore returns a Store backed by the supplied Ent client.
func NewStore(db *ent.Client) (*Store, error) {
	if db == nil {
		return nil, ErrStoreNotInitialized
	}

	return &Store{
		db: db,
	}, nil
}

// SaveCredential upserts credentials for the given org/provider pair.
func (s *Store) SaveCredential(ctx context.Context, orgID string, provider types.ProviderType, authKind types.AuthKind, credential models.CredentialSet) (models.CredentialSet, error) {
	if provider == types.ProviderUnknown {
		return models.CredentialSet{}, ErrProviderRequired
	}

	if orgID == "" {
		return models.CredentialSet{}, ErrOrgIDRequired
	}

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)
	systemCtx = auth.WithCaller(systemCtx, auth.NewKeystoreCaller())

	integrationRecord, err := s.ensureIntegration(systemCtx, orgID, provider)
	if err != nil {
		logx.FromContext(systemCtx).Error().Err(err).Msg("failed to ensure integration record")
		return models.CredentialSet{}, err
	}

	return s.saveCredentialForIntegrationRecord(systemCtx, orgID, provider, authKind, credential, integrationRecord)
}

// SaveCredentialForIntegration upserts credentials for a specific integration record.
func (s *Store) SaveCredentialForIntegration(ctx context.Context, orgID string, integrationID string, provider types.ProviderType, authKind types.AuthKind, credential models.CredentialSet) (models.CredentialSet, error) {
	if provider == types.ProviderUnknown {
		return models.CredentialSet{}, ErrProviderRequired
	}

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)
	systemCtx = auth.WithCaller(systemCtx, auth.NewKeystoreCaller())

	integrationRecord, err := s.integrationByID(systemCtx, orgID, provider, integrationID, false, false)
	if err != nil {
		logx.FromContext(systemCtx).Error().Err(err).Str("integration_id", integrationID).Msg("failed to load integration record for credential save")
		return models.CredentialSet{}, err
	}

	return s.saveCredentialForIntegrationRecord(systemCtx, orgID, provider, authKind, credential, integrationRecord)
}

// EnsureIntegration guarantees an integration record exists for the given org/provider pair.
func (s *Store) EnsureIntegration(ctx context.Context, orgID string, provider types.ProviderType) (*ent.Integration, error) {
	return s.ensureIntegration(ctx, orgID, provider)
}

// LoadCredential retrieves credentials and metadata for the given org/provider pair.
func (s *Store) LoadCredential(ctx context.Context, orgID string, provider types.ProviderType) (models.CredentialSet, types.AuthKind, *integrationstate.IntegrationProviderState, error) {
	if provider == types.ProviderUnknown {
		return models.CredentialSet{}, types.AuthKindUnknown, nil, ErrProviderRequired
	}

	if orgID == "" {
		return models.CredentialSet{}, types.AuthKindUnknown, nil, ErrOrgIDRequired
	}

	integrationRecord, found, err := s.providerIntegration(ctx, orgID, provider, true)
	if err != nil {
		return models.CredentialSet{}, types.AuthKindUnknown, nil, err
	}
	if !found {
		return models.CredentialSet{}, types.AuthKindUnknown, nil, ErrCredentialNotFound
	}

	return s.loadCredentialFromIntegrationRecord(provider, integrationRecord)
}

// LoadCredentialForIntegration retrieves credentials and metadata for a specific integration record.
func (s *Store) LoadCredentialForIntegration(ctx context.Context, orgID string, provider types.ProviderType, integrationID string) (models.CredentialSet, types.AuthKind, *integrationstate.IntegrationProviderState, error) {
	integrationRecord, err := s.integrationByID(ctx, orgID, provider, integrationID, true, true)
	if err != nil {
		return models.CredentialSet{}, types.AuthKindUnknown, nil, err
	}

	return s.loadCredentialFromIntegrationRecord(provider, integrationRecord)
}

// loadCredentialForIntegrationRecord attempts to load credentials for the given integration record.
// It returns found=false when no persisted secret exists.
func (s *Store) loadCredentialForIntegrationRecord(ctx context.Context, orgID string, provider types.ProviderType, integrationID string) (models.CredentialSet, types.AuthKind, *integrationstate.IntegrationProviderState, bool, error) {
	integrationRecord, err := s.integrationByID(ctx, orgID, provider, integrationID, true, true)
	if err != nil {
		return models.CredentialSet{}, types.AuthKindUnknown, nil, false, err
	}

	credential, authKind, providerState, err := s.loadCredentialFromIntegrationRecord(provider, integrationRecord)
	if err == nil {
		return credential, authKind, providerState, true, nil
	}
	if !errors.Is(err, ErrCredentialNotFound) {
		return models.CredentialSet{}, types.AuthKindUnknown, nil, false, err
	}

	state := integrationRecord.ProviderState

	return models.CredentialSet{}, types.AuthKindUnknown, &state, false, nil
}

// DeleteIntegration removes the integration and associated secrets for the given org.
func (s *Store) DeleteIntegration(ctx context.Context, orgID string, integrationID string) (types.ProviderType, string, error) {
	record, err := s.integrationByID(ctx, orgID, types.ProviderUnknown, integrationID, true, false)
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

// ensureIntegration guarantees an integration record exists for the given org/provider pair.
func (s *Store) ensureIntegration(ctx context.Context, orgID string, provider types.ProviderType) (*ent.Integration, error) {
	record, found, err := s.providerIntegration(ctx, orgID, provider, false)
	if err != nil {
		return nil, err
	}
	if found {
		return record, nil
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

func (s *Store) saveCredentialForIntegrationRecord(ctx context.Context, orgID string, provider types.ProviderType, authKind types.AuthKind, credential models.CredentialSet, integrationRecord *ent.Integration) (models.CredentialSet, error) {
	secretName := string(provider)
	envelope := types.CloneCredentialSet(credential)
	normalizedKind := normalizeCredentialAuthKind(authKind, envelope)

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
			return models.CredentialSet{}, err
		}

		createErr := s.db.Hush.Create().
			SetOwnerID(orgID).
			SetName(secretName).
			SetSecretName(secretName).
			SetKind(string(normalizedKind)).
			SetCredentialSet(envelope).
			AddIntegrations(integrationRecord).
			Exec(ctx)
		if createErr != nil {
			logx.FromContext(ctx).Error().Err(createErr).Msg("failed to create credential record")
			return models.CredentialSet{}, createErr
		}
	} else {
		updateErr := existing.Update().
			SetCredentialSet(envelope).
			SetKind(string(normalizedKind)).
			Exec(ctx)
		if updateErr != nil {
			logx.FromContext(ctx).Error().Err(updateErr).Msg("failed to update credential record")
			return models.CredentialSet{}, updateErr
		}
	}

	return types.CloneCredentialSet(envelope), nil
}

func (s *Store) loadCredentialFromIntegrationRecord(provider types.ProviderType, integrationRecord *ent.Integration) (models.CredentialSet, types.AuthKind, *integrationstate.IntegrationProviderState, error) {
	secretName := string(provider)
	for _, secret := range integrationRecord.Edges.Secrets {
		if secret.SecretName != secretName {
			continue
		}

		credential := types.CloneCredentialSet(secret.CredentialSet)
		state := integrationRecord.ProviderState

		return credential, types.AuthKindFromString(secret.Kind), &state, nil
	}

	return models.CredentialSet{}, types.AuthKindUnknown, nil, ErrCredentialNotFound
}

// providerIntegrations returns up to two integrations for owner/provider.
// A result length of 2 indicates an ambiguous provider-level lookup.
func (s *Store) providerIntegrations(ctx context.Context, orgID string, provider types.ProviderType, withSecrets bool) ([]*ent.Integration, error) {
	query := s.db.Integration.Query().
		Where(
			integration.OwnerIDEQ(orgID),
			integration.KindEQ(string(provider)),
		).
		Order(ent.Desc(integration.FieldUpdatedAt), ent.Desc(integration.FieldCreatedAt)).
		Limit(providerIntegrationLookupLimit)
	if withSecrets {
		query = query.WithSecrets()
	}

	return query.All(ctx)
}

// providerIntegration resolves a single integration by owner/provider and reports whether one exists.
// Returns ErrIntegrationAmbiguous when multiple integrations are configured for the provider.
func (s *Store) providerIntegration(ctx context.Context, orgID string, provider types.ProviderType, withSecrets bool) (*ent.Integration, bool, error) {
	records, err := s.providerIntegrations(ctx, orgID, provider, withSecrets)
	if err != nil {
		return nil, false, err
	}
	switch len(records) {
	case 0:
		return nil, false, nil
	case 1:
		return records[0], true, nil
	default:
		return nil, false, ErrIntegrationAmbiguous
	}
}

// integrationByID resolves an integration by explicit ID/owner and optional provider.
func (s *Store) integrationByID(ctx context.Context, orgID string, provider types.ProviderType, integrationID string, withSecrets bool, mapNotFound bool) (*ent.Integration, error) {
	query := s.db.Integration.Query().
		Where(
			integration.IDEQ(integrationID),
			integration.OwnerIDEQ(orgID),
		)
	if provider != types.ProviderUnknown {
		query = query.Where(integration.KindEQ(string(provider)))
	}
	if withSecrets {
		query = query.WithSecrets()
	}

	record, err := query.Only(ctx)
	if err != nil {
		if mapNotFound && ent.IsNotFound(err) {
			return nil, ErrCredentialNotFound
		}

		return nil, err
	}

	return record, nil
}

func normalizeCredentialAuthKind(kind types.AuthKind, credential models.CredentialSet) types.AuthKind {
	normalized := kind.Normalize()
	if normalized != types.AuthKindUnknown {
		return normalized
	}

	return types.InferAuthKind(credential)
}
