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
	integrationID string
	clientID      types.ClientID
	digest        string
}

func (k clientCacheKey) String() string {
	return k.integrationID + ":" + k.clientID.String() + ":" + k.digest
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

// LoadCredential resolves one persisted credential slot for one installation record
func (s *Store) LoadCredential(ctx context.Context, installation *ent.Integration, credentialRef types.CredentialSlotID) (types.CredentialSet, bool, error) {
	if credentialRef == (types.CredentialSlotID{}) {
		return types.CredentialSet{}, false, ErrCredentialNotFound
	}

	record, ok, err := s.activeCredentialRecord(integrationSystemContext(ctx), installation.ID, credentialRef)
	if err != nil {
		return types.CredentialSet{}, false, err
	}
	if !ok {
		return types.CredentialSet{}, false, nil
	}

	return types.CredentialSet(record.CredentialSet), true, nil
}

// LoadCredentials resolves the requested credential slots for one installation record
func (s *Store) LoadCredentials(ctx context.Context, installation *ent.Integration, credentialRefs []types.CredentialSlotID) (types.CredentialBindings, error) {
	records, err := s.activeCredentialRecords(integrationSystemContext(ctx), installation.ID, credentialRefs)
	if err != nil {
		return nil, err
	}

	bindings := make(types.CredentialBindings, 0, len(records))
	for _, ref := range credentialRefs {
		record, ok := records[ref.String()]
		if !ok {
			continue
		}

		bindings = append(bindings, types.CredentialBinding{
			Ref:        ref,
			Credential: cloneCredentialSet(types.CredentialSet(record.CredentialSet)),
		})
	}

	return bindings, nil
}

// SaveCredential upserts one credential slot for one installation record
func (s *Store) SaveCredential(ctx context.Context, installation *ent.Integration, credentialRef types.CredentialSlotID, credential types.CredentialSet) error {
	if credentialRef == (types.CredentialSlotID{}) {
		return ErrCredentialNotFound
	}

	systemCtx := integrationSystemContext(ctx)

	existing, ok, err := s.activeCredentialRecord(systemCtx, installation.ID, credentialRef)
	if err != nil {
		return err
	}

	secretName := credentialRef.String()
	if !ok {
		if err := s.db.Hush.Create().
			SetOwnerID(installation.OwnerID).
			SetName(secretName).
			SetSecretName(secretName).
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

// SaveInstallationCredential loads the installation record by ID and upserts one credential slot
func (s *Store) SaveInstallationCredential(ctx context.Context, integrationID string, credentialRef types.CredentialSlotID, credential types.CredentialSet) error {
	if integrationID == "" {
		return ErrInstallationIDRequired
	}
	if credentialRef == (types.CredentialSlotID{}) {
		return ErrCredentialNotFound
	}

	systemCtx := integrationSystemContext(ctx)

	installation, err := s.db.Integration.Get(systemCtx, integrationID)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrCredentialNotFound
		}

		return err
	}

	return s.SaveCredential(ctx, installation, credentialRef, credential)
}

// DeleteCredential removes all credentials for one installation by identifier
func (s *Store) DeleteCredential(ctx context.Context, integrationID string) error {
	if integrationID == "" {
		return ErrInstallationIDRequired
	}

	systemCtx := integrationSystemContext(ctx)

	_, err := s.db.Hush.Delete().
		Where(enthush.HasIntegrationsWith(entintegration.IDEQ(integrationID))).
		Exec(systemCtx)
	if err != nil {
		return err
	}

	s.InvalidateClients(integrationID)

	return nil
}

// BuildClient resolves one named client for an installation using explicit credential bundles
func (s *Store) BuildClient(ctx context.Context, installation *ent.Integration, registration types.ClientRegistration, credentials types.CredentialBindings, config json.RawMessage, force bool) (any, error) {
	cacheKey := clientCacheKey{
		integrationID: installation.ID,
		clientID:      registration.Ref,
		digest:        clientCacheDigest(credentials, config),
	}

	if force {
		s.clientPool.RemoveClient(cacheKey)
	} else if cached := s.clientPool.GetClient(cacheKey); cached.IsPresent() {
		return cached.MustGet(), nil
	}

	client, err := registration.Build(ctx, types.ClientBuildRequest{
		Integration: installation,
		Credentials: cloneCredentialBindings(credentials),
		Config:      jsonx.CloneRawMessage(config),
	})
	if err != nil {
		return nil, err
	}

	s.clientPool.SetClient(cacheKey, client)
	s.trackClientKey(cacheKey)

	return client, nil
}

// InvalidateClients drops all pooled clients for one installation
func (s *Store) InvalidateClients(integrationID string) {
	s.mu.Lock()
	keys := s.clientKeys[integrationID]
	delete(s.clientKeys, integrationID)
	s.mu.Unlock()

	for key := range keys {
		s.clientPool.RemoveClient(key)
	}
}

// trackClientKey tracks a client cache key for an installation
func (s *Store) trackClientKey(key clientCacheKey) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.clientKeys[key.integrationID] == nil {
		s.clientKeys[key.integrationID] = map[clientCacheKey]struct{}{}
	}

	s.clientKeys[key.integrationID][key] = struct{}{}
}

// clientCacheDigest computes a hash digest for a set of credentials and config
func clientCacheDigest(credentials types.CredentialBindings, config json.RawMessage) string {
	encodedCredential, err := json.Marshal(credentials)
	if err != nil {
		encodedCredential = nil
	}

	return helpers.NewHashBuilder().
		WriteStrings(string(encodedCredential), string(config)).
		Hex()
}

// cloneCredentialSet returns a deep copy of a CredentialSet
func cloneCredentialSet(credential types.CredentialSet) types.CredentialSet {
	return types.CredentialSet{
		Data: jsonx.CloneRawMessage(credential.Data),
	}
}

// cloneCredentialBindings returns a deep copy of CredentialBindings
func cloneCredentialBindings(credentials types.CredentialBindings) types.CredentialBindings {
	if len(credentials) == 0 {
		return nil
	}

	cloned := make(types.CredentialBindings, 0, len(credentials))
	for _, binding := range credentials {
		cloned = append(cloned, types.CredentialBinding{
			Ref:        binding.Ref,
			Credential: cloneCredentialSet(binding.Credential),
		})
	}

	return cloned
}

// activeCredentialRecord returns the active credential record for an installation and credential ref
func (s *Store) activeCredentialRecord(ctx context.Context, integrationID string, credentialRef types.CredentialSlotID) (*ent.Hush, bool, error) {
	records, err := s.activeCredentialRecords(ctx, integrationID, []types.CredentialSlotID{credentialRef})
	if err != nil {
		return nil, false, err
	}

	record, ok := records[credentialRef.String()]
	if !ok {
		return nil, false, nil
	}

	return record, true, nil
}

// activeCredentialRecords returns all active credential records for an installation and credential refs
// Credentials are keyed by Hush.SecretName which stores the CredentialSlotID as a unique discriminator
func (s *Store) activeCredentialRecords(ctx context.Context, integrationID string, credentialRefs []types.CredentialSlotID) (map[string]*ent.Hush, error) {
	query := s.db.Hush.Query().Where(enthush.HasIntegrationsWith(entintegration.IDEQ(integrationID)))

	if len(credentialRefs) > 0 {
		secretNames := make([]string, 0, len(credentialRefs))
		for _, ref := range credentialRefs {
			if ref == (types.CredentialSlotID{}) {
				continue
			}

			secretNames = append(secretNames, ref.String())
		}

		if len(secretNames) > 0 {
			query = query.Where(enthush.SecretNameIn(secretNames...))
		}
	}

	records, err := query.
		Order(
			enthush.ByUpdatedAt(sql.OrderDesc()),
			enthush.ByCreatedAt(sql.OrderDesc()),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := make(map[string]*ent.Hush, len(records))
	for _, record := range records {
		if _, exists := out[record.SecretName]; exists {
			continue
		}

		out[record.SecretName] = record
	}

	return out, nil
}

// integrationSystemContext returns a context with system-level privileges for integration operations
func integrationSystemContext(ctx context.Context) context.Context {
	callCtx := privacy.DecisionContext(ctx, privacy.Allow)
	return auth.WithCaller(callCtx, auth.NewKeystoreCaller())
}
