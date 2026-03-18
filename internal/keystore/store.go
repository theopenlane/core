package keystore

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/eddy"

	"github.com/theopenlane/core/common/helpers"
	ent "github.com/theopenlane/core/internal/ent/generated"
	enthush "github.com/theopenlane/core/internal/ent/generated/hush"
	entintegration "github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/mapx"
	"github.com/theopenlane/iam/auth"
)

// Store persists and retrieves installation credentials via Ent-backed hush secrets
type Store struct {
	// db provides access to the Ent ORM client
	db *ent.Client
	// clientPool caches installation-scoped initialized clients
	clientPool *eddy.ClientPool[any]
	// clientKeys indexes pooled cache keys by installation for targeted invalidation
	clientKeys map[string]map[clientCacheKey]struct{}
	// mu protects clientKeys
	mu sync.Mutex
}

const defaultClientPoolTTL = 5 * time.Minute

type clientCacheKey struct {
	installationID string
	clientID       types.ClientID
	digest         string
}

func (k clientCacheKey) String() string {
	return k.installationID + ":" + k.clientID.String() + ":" + k.digest
}

// NewStore constructs the credential store backed by the supplied Ent client
func NewStore(db *ent.Client) (*Store, error) {
	if db == nil {
		return nil, ErrStoreNotInitialized
	}

	return &Store{
		db:         db,
		clientPool: eddy.NewClientPool[any](defaultClientPoolTTL),
		clientKeys: map[string]map[clientCacheKey]struct{}{},
	}, nil
}

// LoadCredential resolves persisted credentials for one installation record
func (s *Store) LoadCredential(ctx context.Context, installation *ent.Integration) (types.CredentialSet, bool, error) {
	if installation == nil {
		return types.CredentialSet{}, false, ErrCredentialNotFound
	}

	record, ok, err := s.activeCredentialRecord(integrationSystemContext(ctx), installation.ID)
	if err != nil {
		return types.CredentialSet{}, false, err
	}
	if !ok {
		return types.CredentialSet{}, false, nil
	}

	return types.CredentialSet(record.CredentialSet), true, nil
}

// installationCredentialName is the canonical secret name used for all installation credentials
const installationCredentialName = "default"

// SaveCredential upserts the credential for one installation record
func (s *Store) SaveCredential(ctx context.Context, installation *ent.Integration, credential types.CredentialSet) error {
	if installation == nil {
		return ErrInstallationIDRequired
	}

	systemCtx := integrationSystemContext(ctx)

	existing, ok, err := s.activeCredentialRecord(systemCtx, installation.ID)
	if err != nil {
		return err
	}

	if !ok {
		if err := s.db.Hush.Create().
			SetOwnerID(installation.OwnerID).
			SetName(installationCredentialName).
			SetSecretName(installationCredentialName).
			SetCredentialSet(credential).
			AddIntegrationIDs(installation.ID).
			Exec(systemCtx); err != nil {
			return err
		}
	} else {
		if err := existing.Update().
			SetCredentialSet(credential).
			Exec(systemCtx); err != nil {
			return err
		}
	}

	s.InvalidateClients(installation.ID)

	return nil
}

// SaveInstallationCredential loads the installation record by ID and upserts its credential
func (s *Store) SaveInstallationCredential(ctx context.Context, installationID string, credential types.CredentialSet) error {
	if installationID == "" {
		return ErrInstallationIDRequired
	}

	systemCtx := integrationSystemContext(ctx)

	installation, err := s.db.Integration.Get(systemCtx, installationID)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrCredentialNotFound
		}

		return err
	}

	return s.SaveCredential(ctx, installation, credential)
}

// DeleteCredential removes all credentials for one installation by identifier
func (s *Store) DeleteCredential(ctx context.Context, installationID string) error {
	if installationID == "" {
		return ErrInstallationIDRequired
	}

	systemCtx := integrationSystemContext(ctx)

	_, err := s.db.Hush.Delete().
		Where(enthush.HasIntegrationsWith(entintegration.IDEQ(installationID))).
		Exec(systemCtx)
	if err != nil {
		return err
	}

	s.InvalidateClients(installationID)

	return nil
}

// BuildClient resolves one named client for an installation using an explicit credential bundle
func (s *Store) BuildClient(ctx context.Context, installation *ent.Integration, registration types.ClientRegistration, credential types.CredentialSet, config json.RawMessage, force bool) (any, error) {
	if installation == nil {
		return nil, ErrInstallationIDRequired
	}

	cacheKey := clientCacheKey{
		installationID: installation.ID,
		clientID:       registration.Ref,
		digest:         clientCacheDigest(credential, config),
	}

	if force {
		s.clientPool.RemoveClient(cacheKey)
	} else if cached := s.clientPool.GetClient(cacheKey); cached.IsPresent() {
		return cached.MustGet(), nil
	}

	client, err := registration.Build(ctx, types.ClientBuildRequest{
		Installation: installation,
		Credential:   cloneCredentialSet(credential),
		Config:       jsonx.CloneRawMessage(config),
	})
	if err != nil {
		return nil, err
	}

	s.clientPool.SetClient(cacheKey, client)
	s.trackClientKey(cacheKey)

	return client, nil
}

// InvalidateClients drops all pooled clients for one installation
func (s *Store) InvalidateClients(installationID string) {
	s.mu.Lock()
	keys := s.clientKeys[installationID]
	delete(s.clientKeys, installationID)
	s.mu.Unlock()

	for key := range keys {
		s.clientPool.RemoveClient(key)
	}
}

func (s *Store) trackClientKey(key clientCacheKey) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.clientKeys[key.installationID] == nil {
		s.clientKeys[key.installationID] = map[clientCacheKey]struct{}{}
	}

	s.clientKeys[key.installationID][key] = struct{}{}
}

func clientCacheDigest(credential types.CredentialSet, config json.RawMessage) string {
	encodedCredential, err := json.Marshal(credential)
	if err != nil {
		encodedCredential = nil
	}

	return helpers.NewHashBuilder().
		WriteStrings(string(encodedCredential), string(config)).
		Hex()
}

func cloneCredentialSet(credential types.CredentialSet) types.CredentialSet {
	clone := credential
	clone.ProviderData = jsonx.CloneRawMessage(credential.ProviderData)
	clone.Claims = mapx.DeepCloneMapAny(credential.Claims)

	if credential.OAuthExpiry != nil {
		expiry := *credential.OAuthExpiry
		clone.OAuthExpiry = &expiry
	}

	return clone
}

func (s *Store) activeCredentialRecord(ctx context.Context, installationID string) (*ent.Hush, bool, error) {
	records, err := s.db.Hush.Query().
		Where(enthush.HasIntegrationsWith(entintegration.IDEQ(installationID))).
		Order(
			enthush.ByUpdatedAt(sql.OrderDesc()),
			enthush.ByCreatedAt(sql.OrderDesc()),
		).
		All(ctx)
	if err != nil {
		return nil, false, err
	}
	if len(records) == 0 {
		return nil, false, nil
	}

	// Prefer the most recently updated credential so sequential rotation keeps working
	// even when older rows still exist for the same installation.
	return records[0], true, nil
}

func integrationSystemContext(ctx context.Context) context.Context {
	callCtx := privacy.DecisionContext(ctx, privacy.Allow)
	return auth.WithCaller(callCtx, auth.NewKeystoreCaller())
}
