package keystore

import (
	"context"
	"errors"

	ent "github.com/theopenlane/core/internal/ent/generated"
	enthush "github.com/theopenlane/core/internal/ent/generated/hush"
	entintegration "github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/iam/auth"
)

// Store persists and retrieves v2 installation credentials via Ent-backed hush secrets
type Store struct {
	// db provides access to the Ent ORM client
	db *ent.Client
}

// NewStore constructs the v2 credential store backed by the supplied Ent client
func NewStore(db *ent.Client) (*Store, error) {
	if db == nil {
		return nil, ErrStoreNotInitialized
	}

	return &Store{db: db}, nil
}

// LoadCredential resolves persisted credentials for one installation record
func (s *Store) LoadCredential(ctx context.Context, integration *ent.Integration) (types.CredentialSet, bool, error) {
	if integration == nil {
		return types.CredentialSet{}, false, ErrCredentialNotFound
	}

	record, err := s.db.Hush.Query().
		Where(enthush.HasIntegrationsWith(entintegration.IDEQ(integration.ID))).
		Only(integrationV2SystemContext(ctx))
	if err != nil {
		if ent.IsNotFound(err) {
			return types.CredentialSet{}, false, nil
		}

		if errors.Is(err, ErrCredentialNotFound) {
			return types.CredentialSet{}, false, nil
		}

		return types.CredentialSet{}, false, err
	}

	return types.CredentialSet(record.CredentialSet), true, nil
}

// installationCredentialName is the canonical secret name used for all v2 installation credentials
const installationCredentialName = "default"

// SaveCredential upserts the credential for one v2 installation record
func (s *Store) SaveCredential(ctx context.Context, integration *ent.Integration, credential types.CredentialSet) error {
	if integration == nil {
		return ErrInstallationIDRequired
	}

	systemCtx := integrationV2SystemContext(ctx)

	existing, queryErr := s.db.Hush.Query().
		Where(enthush.HasIntegrationsWith(entintegration.IDEQ(integration.ID))).
		Only(systemCtx)
	if queryErr != nil {
		if !ent.IsNotFound(queryErr) {
			return queryErr
		}

		return s.db.Hush.Create().
			SetOwnerID(integration.OwnerID).
			SetName(installationCredentialName).
			SetSecretName(installationCredentialName).
			SetCredentialSet(credential).
			AddIntegrationIDs(integration.ID).
			Exec(systemCtx)
	}

	return existing.Update().
		SetCredentialSet(credential).
		Exec(systemCtx)
}

// SaveInstallationCredential loads the installation record by ID and upserts its credential.
// This satisfies keymaker.CredentialWriter without requiring the caller to pre-load the record.
func (s *Store) SaveInstallationCredential(ctx context.Context, installationID string, credential types.CredentialSet) error {
	if installationID == "" {
		return ErrInstallationIDRequired
	}

	systemCtx := integrationV2SystemContext(ctx)

	integration, err := s.db.Integration.Get(systemCtx, installationID)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrCredentialNotFound
		}

		return err
	}

	return s.SaveCredential(ctx, integration, credential)
}

// DeleteCredential removes all credentials for one v2 installation by identifier
func (s *Store) DeleteCredential(ctx context.Context, installationID string) error {
	if installationID == "" {
		return ErrInstallationIDRequired
	}

	systemCtx := integrationV2SystemContext(ctx)

	_, err := s.db.Hush.Delete().
		Where(enthush.HasIntegrationsWith(entintegration.IDEQ(installationID))).
		Exec(systemCtx)

	return err
}

// integrationV2SystemContext applies privileged auth and privacy decisions for integrationsv2 credential reads
func integrationV2SystemContext(ctx context.Context) context.Context {
	callCtx := privacy.DecisionContext(ctx, privacy.Allow)
	return auth.WithCaller(callCtx, auth.NewKeystoreCaller())
}
