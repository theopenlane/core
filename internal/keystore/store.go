package keystore

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/samber/lo"
	"github.com/theopenlane/eddy"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/helpers"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	enthush "github.com/theopenlane/core/v2/internal/ent/generated/hush"
	entintegration "github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
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
	connection    string
	client        string
	digest        string
}

func (k clientCacheKey) String() string {
	return strings.Join([]string{k.integrationID, k.connection, k.client, k.digest}, ":")
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

// LoadCredential resolves the persisted credential of the named connection for one installation record
func (s *Store) LoadCredential(ctx context.Context, installation *ent.Integration, name string) (types.CredentialSet, bool, error) {
	if name == "" {
		return types.CredentialSet{}, false, ErrCredentialNotFound
	}

	record, ok, err := s.activeCredentialRecord(auth.WithOrgInternalCaller(ctx, installation.OwnerID), installation.ID, credentialRef)
	if err != nil {
		return types.CredentialSet{}, false, err
	}
	if !ok {
		return types.CredentialSet{}, false, nil
	}

	return types.CredentialSet(record.CredentialSet), true, nil
}

// LoadAllCredentials resolves every persisted credential for one installation record keyed by secret name
func (s *Store) LoadAllCredentials(ctx context.Context, installation *ent.Integration) (map[string]types.CredentialSet, error) {
	records, err := s.activeCredentialRecords(integrationSystemContext(ctx), installation.ID, nil)
	if err != nil {
		return nil, err
	}

	out := make(map[string]types.CredentialSet, len(records))
	for secretName, record := range records {
		out[secretName] = cloneCredentialSet(types.CredentialSet(record.CredentialSet))
	}

	return out, nil
}

// SaveCredential upserts the credential of the named connection for one installation record
func (s *Store) SaveCredential(ctx context.Context, installation *ent.Integration, name string, credential types.CredentialSet) error {
	if name == "" {
		return ErrCredentialNotFound
	}

	existing, ok, err := s.activeCredentialRecord(ctx, installation.ID, credentialRef)
	if err != nil {
		return err
	}

	secretName := name
	if !ok {
		if err := s.client(ctx).Hush.Create().
			SetOwnerID(installation.OwnerID).
			SetName(secretName).
			SetSecretName(secretName).
			SetCredentialSet(credential).
			AddIntegrationIDs(installation.ID).
			Exec(ctx); err != nil {
			return err
		}
	} else {
		if err := existing.Update().
			SetCredentialSet(credential).
			Exec(ctx); err != nil {
			return err
		}
	}

	s.InvalidateClients(installation.ID)

	return nil
}

// SaveInstallationCredential loads the installation record by ID and upserts the credential of the named connection
func (s *Store) SaveInstallationCredential(ctx context.Context, integrationID string, name string, credential types.CredentialSet) error {
	if integrationID == "" {
		return ErrInstallationIDRequired
	}
	if name == "" {
		return ErrCredentialNotFound
	}

	installation, err := s.db.Integration.Get(ctx, integrationID)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrCredentialNotFound
		}

		return err
	}

	return s.SaveCredential(ctx, installation, name, credential)
}

// DeleteCredential removes all credentials for one installation by identifier
func (s *Store) DeleteCredential(ctx context.Context, integrationID string) error {
	if integrationID == "" {
		return ErrInstallationIDRequired
	}

	_, err := s.db.Hush.Delete().
		Where(enthush.HasIntegrationsWith(entintegration.IDEQ(integrationID))).
		Exec(ctx)
	if err != nil {
		return err
	}

	s.InvalidateClients(integrationID)

	return nil
}

// ReplaceCredentials reconciles installation credentials keyed by secret name from previous to next
func (s *Store) ReplaceCredentials(ctx context.Context, installation *ent.Integration, previous, next map[string]types.CredentialSet) error {

	for slot, credential := range next {
		existing, tracked := previous[slot]

		if tracked && bytes.Equal(existing.Data, credential.Data) {
			continue
		}

		if !tracked {
			_, exists, err := s.activeCredentialRecord(ctx, installation.ID, slot)
			if err != nil {
				return err
			}

			if exists {
				continue
			}
		}

		if err := s.SaveCredential(ctx, installation, slot, credential); err != nil {
			return err
		}
	}

	removed := lo.FilterMap(lo.Keys(previous), func(slot string, _ int) (string, bool) {
		_, keep := next[slot]

		return slot, !keep
	})

	if len(removed) > 0 {
		if _, err := s.client(ctx).Hush.Delete().
			Where(
				enthush.HasIntegrationsWith(entintegration.IDEQ(installation.ID)),
				enthush.SecretNameIn(removed...),
			).
			Exec(ctx); err != nil {
			return err
		}
	}

	s.InvalidateClients(installation.ID)

	return nil
}

// BuildClient resolves one named client of a connection for an installation from its credential
func (s *Store) BuildClient(ctx context.Context, installation *ent.Integration, connection, client string, build types.ClientBuilderFunc, credential types.CredentialSet, force bool) (any, error) {
	cacheKey := clientCacheKey{
		integrationID: installation.ID,
		connection:    connection,
		client:        client,
		digest:        clientCacheDigest(credential),
	}

	if force {
		s.clientPool.RemoveClient(cacheKey)
	} else if cached := s.clientPool.GetClient(cacheKey); cached.IsPresent() {
		return cached.MustGet(), nil
	}

	built, err := build(ctx, types.ConnectionInput{
		Integration:  installation,
		Credential:   cloneCredentialSet(credential),
		TokenManager: s.db.TokenManager,
	})
	if err != nil {
		return nil, err
	}

	s.clientPool.SetClient(cacheKey, built)
	s.trackClientKey(cacheKey)

	return built, nil
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

// clientCacheDigest computes a hash digest for a credential's data
func clientCacheDigest(credential types.CredentialSet) string {
	return helpers.NewHashBuilder().
		WriteStrings(string(credential.Data)).
		Hex()
}

// cloneCredentialSet returns a deep copy of a CredentialSet
func cloneCredentialSet(credential types.CredentialSet) types.CredentialSet {
	return types.CredentialSet{
		Data: jsonx.CloneRawMessage(credential.Data),
	}
}

// client returns the transaction's client when ctx carries one, else the store's client
func (s *Store) client(ctx context.Context) *ent.Client {
	if tx := ent.TxFromContext(ctx); tx != nil {
		return tx.Client()
	}

	return s.db
}

// activeCredentialRecord returns the active credential record for a connection name
func (s *Store) activeCredentialRecord(ctx context.Context, integrationID string, name string) (*ent.Hush, bool, error) {
	records, err := s.activeCredentialRecords(ctx, integrationID, []string{name})
	if err != nil {
		return nil, false, err
	}

	record, ok := records[name]
	if !ok {
		return nil, false, nil
	}

	return record, true, nil
}

// activeCredentialRecords returns active credential records keyed by Hush.SecretName
func (s *Store) activeCredentialRecords(ctx context.Context, integrationID string, names []string) (map[string]*ent.Hush, error) {
	query := s.client(ctx).Hush.Query().Where(enthush.HasIntegrationsWith(entintegration.IDEQ(integrationID)))

	secretNames := lo.Compact(names)
	if len(secretNames) > 0 {
		query = query.Where(enthush.SecretNameIn(secretNames...))
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
