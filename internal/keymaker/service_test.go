package keymaker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/keymaker/oauth"
	"github.com/theopenlane/core/internal/keymaker/pool"
	"github.com/theopenlane/core/internal/keymaker/store"
	"github.com/theopenlane/core/internal/keystore"
)

func TestServiceSessionValidatesRequest(t *testing.T) {
	t.Parallel()

	svc := NewService(&mockRepository{}, nil, ProviderOptions{})

	_, err := svc.Session(context.Background(), SessionRequest{})
	require.ErrorIs(t, err, ErrMissingOrgID)

	_, err = svc.Session(context.Background(), SessionRequest{
		OrgID: "org",
	})
	require.ErrorIs(t, err, ErrMissingIntegration)

	_, err = svc.Session(context.Background(), SessionRequest{
		OrgID:       "org",
		Integration: "integration",
	})
	require.ErrorIs(t, err, ErrMissingProvider)
}

func TestServiceSessionRequiresRegisteredFactory(t *testing.T) {
	t.Parallel()

	now := time.Now().Add(time.Hour).UTC().Truncate(time.Second)

	repo := &mockRepository{
		integrationFn: func(context.Context, string, string) (store.IntegrationRecord, error) {
			return store.IntegrationRecord{
				OrgID:         "org",
				IntegrationID: "integration",
				Provider:      "slack",
			}, nil
		},
		credentialsFn: func(context.Context, string, string) (store.CredentialRecord, error) {
			return makeCredentialRecord(t, "slack", "access", "refresh", now, nil), nil
		},
	}

	svc := NewService(repo, nil, ProviderOptions{})
	_, err := svc.Session(context.Background(), SessionRequest{
		OrgID:       "org",
		Integration: "integration",
		Provider:    "slack",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrFactoryNotRegistered)
}

func TestServiceSessionAcquisitionFailure(t *testing.T) {
	t.Parallel()

	repo := fixedRepository(t)

	svc := NewService(repo, nil, ProviderOptions{
		Factories: []ClientFactory{
			&stubFactory{provider: "slack"},
		},
		Pools: map[string]pool.Manager[any]{
			"slack": &stubPoolManager{
				acquireFn: func(context.Context, pool.Key, pool.Factory[any]) (pool.Handle[any], error) {
					return pool.Handle[any]{}, errors.New("boom")
				},
				releaseFn: func(context.Context, pool.Handle[any]) error { return nil },
				purgeFn:   func(context.Context, pool.Key) error { return nil },
			},
		},
	})

	_, err := svc.Session(context.Background(), SessionRequest{
		OrgID:       "org",
		Integration: "integration",
		Provider:    "slack",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "acquire client")
}

func TestServiceSessionSuccess(t *testing.T) {
	t.Parallel()

	repo := fixedRepository(t)

	var (
		acquiredKey pool.Key
		released    bool
	)

	factory := &spyFactory{
		provider: "slack",
		client:   map[string]string{"client": "ok"},
	}

	manager := &stubPoolManager{
		acquireFn: func(ctx context.Context, key pool.Key, factory pool.Factory[any]) (pool.Handle[any], error) {
			acquiredKey = key

			client, err := factory.New(ctx, key)
			if err != nil {
				return pool.Handle[any]{}, err
			}

			return pool.Handle[any]{Key: key, Client: client}, nil
		},
		releaseFn: func(ctx context.Context, handle pool.Handle[any]) error {
			released = true
			return nil
		},
		purgeFn: func(context.Context, pool.Key) error { return nil },
	}

	exchanger := &stubExchanger{
		refreshFn: func(ctx context.Context, spec oauth.Spec, refreshToken string) (oauth.Token, error) {
			return oauth.Token{
				AccessToken:  "new-access",
				RefreshToken: "new-refresh",
				TokenType:    "Bearer",
				Expiry:       time.Now().Add(2 * time.Hour).UTC().Round(time.Second),
			}, nil
		},
	}

	svc := NewService(repo, exchanger, ProviderOptions{
		Specs: []oauth.Spec{
			{Name: "slack"},
		},
		Factories: []ClientFactory{factory},
		Pools: map[string]pool.Manager[any]{
			"slack": manager,
		},
	})

	req := SessionRequest{
		OrgID:       "org",
		Integration: "integration",
		Provider:    "Slack",
		Scopes:      []string{"scope.z"},
	}

	session, err := svc.Session(context.Background(), req)
	require.NoError(t, err)

	assert.Equal(t, "org", session.OrgID)
	assert.Equal(t, "slack", session.Provider)
	assert.Equal(t, factory.client, session.Client)
	assert.Equal(t, []string{"scope.z"}, session.Credentials.Scopes)
	assert.NotNil(t, session.RefreshHandle)

	expectedHash := hashScopes([]string{"scope.z"})
	assert.Equal(t, expectedHash, acquiredKey.ScopeHash)
	assert.Equal(t, factory.sessionRequest, req)
	assert.Equal(t, factory.sessionCreds.AccessToken, "access")

	// ensure metadata cloned
	assert.Equal(t, map[string]any{"tenantId": "tenant"}, session.Metadata)
	session.Metadata["tenantId"] = "mutated"
	assert.Equal(t, map[string]any{"tenantId": "tenant"}, repo.integration.Metadata)

	// Release should trigger pool manager
	require.NoError(t, svc.Release(context.Background(), session))
	assert.True(t, released)
}

func TestServiceReleaseNoop(t *testing.T) {
	t.Parallel()

	svc := NewService(&mockRepository{}, nil, ProviderOptions{})
	err := svc.Release(context.Background(), Session{})
	require.NoError(t, err)
}

func TestHashScopes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "", hashScopes(nil))
	assert.Equal(t, "", hashScopes([]string{}))
	assert.Equal(t, "a a b", hashScopes([]string{"b", "a", "a"}))
}

func TestCloneMetadata(t *testing.T) {
	t.Parallel()

	assert.Nil(t, cloneMetadata(nil))

	src := map[string]any{"foo": "bar"}
	cloned := cloneMetadata(src)
	require.Equal(t, src, cloned)

	cloned["foo"] = "baz"
	assert.Equal(t, "bar", src["foo"])
}

func TestDecodeCredentials(t *testing.T) {
	t.Parallel()

	expiry := time.Now().Add(time.Hour).UTC().Truncate(time.Second)

	record := makeCredentialRecord(t, "slack", "access", "refresh", expiry, map[string]string{
		"custom": "yes",
	})

	creds, err := decodeCredentials("slack", record)
	require.NoError(t, err)

	assert.Equal(t, "access", creds.AccessToken)
	assert.Equal(t, "refresh", creds.RefreshToken)
	assert.Equal(t, expiry, creds.Expiry)
	assert.Equal(t, "Bearer", creds.TokenType)

	rawPayload, ok := creds.Raw["payload"].(map[string]string)
	require.True(t, ok)
	assert.Equal(t, "yes", rawPayload["slack_custom"])
}

func TestRefreshHandleUnsupported(t *testing.T) {
	t.Parallel()

	handle := &refreshHandle{
		service: &Service{},
		req: SessionRequest{
			Provider: "slack",
		},
	}

	_, err := handle.Refresh(context.Background())
	require.ErrorIs(t, err, ErrRefreshUnsupported)
}

func TestRefreshHandleMissingToken(t *testing.T) {
	t.Parallel()

	encrypted, err := json.Marshal(map[string]string{
		"slack_access_token": "access",
	})
	require.NoError(t, err)

	handle := &refreshHandle{
		service: &Service{
			exchanger: &stubExchanger{
				refreshFn: func(context.Context, oauth.Spec, string) (oauth.Token, error) {
					return oauth.Token{}, nil
				},
			},
		},
		req: SessionRequest{
			Provider: "Slack",
		},
		spec: oauth.Spec{Name: "slack"},
		credRecord: store.CredentialRecord{
			Encrypted: encrypted,
		},
	}

	_, err = handle.Refresh(context.Background())
	require.ErrorIs(t, err, ErrRefreshTokenMissing)
}

func TestRefreshHandleSuccess(t *testing.T) {
	t.Parallel()

	expiry := time.Now().Add(30 * time.Minute).UTC().Round(time.Second)

	encrypted := makeEncrypted(t, "slack", map[string]string{
		keystore.AccessTokenField:  "access-old",
		keystore.RefreshTokenField: "refresh-old",
	})

	var saved store.CredentialRecord

	handle := &refreshHandle{
		service: &Service{
			exchanger: &stubExchanger{
				refreshFn: func(context.Context, oauth.Spec, string) (oauth.Token, error) {
					return oauth.Token{
						AccessToken:  "access-new",
						RefreshToken: "refresh-new",
						TokenType:    "Bearer",
						Expiry:       expiry,
					}, nil
				},
			},
			store: &mockRepository{
				saveCredentialsFn: func(_ context.Context, cred store.CredentialRecord) error {
					saved = cred
					return nil
				},
			},
		},
		req: SessionRequest{
			Provider: "slack",
			Scopes:   []string{"scope"},
		},
		spec: oauth.Spec{Name: "slack"},
		credRecord: store.CredentialRecord{
			OrgID:         "org",
			IntegrationID: "integration",
			Provider:      "slack",
			Encrypted:     encrypted,
			Version:       2,
		},
	}

	got, err := handle.Refresh(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "access-new", got.AccessToken)
	assert.Equal(t, "refresh-new", got.RefreshToken)
	assert.Equal(t, expiry, got.Expiry)
	assert.Equal(t, []string{"scope"}, got.Scopes)

	require.NotNil(t, saved.Encrypted)
	require.Equal(t, 3, saved.Version)

	payload := make(map[string]string)
	require.NoError(t, json.Unmarshal(saved.Encrypted, &payload))
	require.Equal(t, "access-new", payload["slack_"+keystore.AccessTokenField])
	require.Equal(t, "refresh-new", payload["slack_"+keystore.RefreshTokenField])
	require.Equal(t, expiry.Format(time.RFC3339), payload["slack_"+keystore.ExpiresAtField])

	rawVersion, ok := got.Raw["version"].(int)
	require.True(t, ok)
	require.Equal(t, 3, rawVersion)

	rawPayload, ok := got.Raw["payload"].(map[string]string)
	require.True(t, ok)
	require.Equal(t, payload, rawPayload)
}

// fixedRepository returns a repository with deterministic records.
func fixedRepository(t *testing.T) *mockRepository {
	t.Helper()

	expiry := time.Now().Add(time.Hour).UTC().Truncate(time.Second)

	integration := store.IntegrationRecord{
		OrgID:         "org",
		IntegrationID: "integration",
		Provider:      "slack",
		Metadata: map[string]any{
			"tenantId": "tenant",
		},
		Scopes: []string{"scope.a", "scope.b"},
	}

	return &mockRepository{
		integrationFn: func(context.Context, string, string) (store.IntegrationRecord, error) {
			return integration, nil
		},
		credentialsFn: func(context.Context, string, string) (store.CredentialRecord, error) {
			return makeCredentialRecord(t, "slack", "access", "refresh", expiry, nil), nil
		},
		integration: integration,
	}
}

func makeCredentialRecord(t *testing.T, provider, access, refresh string, expiry time.Time, extra map[string]string) store.CredentialRecord {
	t.Helper()

	payload := map[string]string{
		provider + keystore.SecretNameSeparator + keystore.AccessTokenField:  access,
		provider + keystore.SecretNameSeparator + keystore.RefreshTokenField: refresh,
	}
	if !expiry.IsZero() {
		payload[provider+keystore.SecretNameSeparator+keystore.ExpiresAtField] = expiry.Format(time.RFC3339)
	}
	for k, v := range extra {
		payload[provider+keystore.SecretNameSeparator+k] = v
	}

	encoded, err := json.Marshal(payload)
	require.NoError(t, err)

	return store.CredentialRecord{
		OrgID:         "org",
		IntegrationID: "integration",
		Provider:      provider,
		Encrypted:     encoded,
		Version:       len(payload),
	}
}

func makeEncrypted(t *testing.T, provider string, fields map[string]string) []byte {
	t.Helper()

	payload := make(map[string]string, len(fields))
	for field, value := range fields {
		payload[provider+keystore.SecretNameSeparator+field] = value
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)
	return data
}

type mockRepository struct {
	integrationFn       func(context.Context, string, string) (store.IntegrationRecord, error)
	credentialsFn       func(context.Context, string, string) (store.CredentialRecord, error)
	saveCredentialsFn   func(context.Context, store.CredentialRecord) error
	deleteCredentialsFn func(context.Context, string, string) error

	integration store.IntegrationRecord
}

func (m *mockRepository) Integration(ctx context.Context, orgID, integrationID string) (store.IntegrationRecord, error) {
	if m.integrationFn == nil {
		return store.IntegrationRecord{}, errors.New("unexpected Integration call")
	}
	return m.integrationFn(ctx, orgID, integrationID)
}

func (m *mockRepository) Credentials(ctx context.Context, orgID, integrationID string) (store.CredentialRecord, error) {
	if m.credentialsFn == nil {
		return store.CredentialRecord{}, errors.New("unexpected Credentials call")
	}
	return m.credentialsFn(ctx, orgID, integrationID)
}

func (m *mockRepository) SaveCredentials(ctx context.Context, cred store.CredentialRecord) error {
	if m.saveCredentialsFn != nil {
		return m.saveCredentialsFn(ctx, cred)
	}
	return nil
}

func (m *mockRepository) DeleteCredentials(ctx context.Context, orgID, integrationID string) error {
	if m.deleteCredentialsFn != nil {
		return m.deleteCredentialsFn(ctx, orgID, integrationID)
	}
	return nil
}

type stubFactory struct {
	provider string
}

func (s *stubFactory) Provider() string { return s.provider }
func (s *stubFactory) New(context.Context, SessionRequest, Credentials) (any, error) {
	return struct{}{}, nil
}

type spyFactory struct {
	provider       string
	client         any
	sessionRequest SessionRequest
	sessionCreds   Credentials
}

func (s *spyFactory) Provider() string { return s.provider }

func (s *spyFactory) New(_ context.Context, req SessionRequest, creds Credentials) (any, error) {
	s.sessionRequest = req
	s.sessionCreds = creds
	return s.client, nil
}

type stubPoolManager struct {
	acquireFn func(context.Context, pool.Key, pool.Factory[any]) (pool.Handle[any], error)
	releaseFn func(context.Context, pool.Handle[any]) error
	purgeFn   func(context.Context, pool.Key) error
}

func (s *stubPoolManager) Acquire(ctx context.Context, key pool.Key, factory pool.Factory[any]) (pool.Handle[any], error) {
	return s.acquireFn(ctx, key, factory)
}

func (s *stubPoolManager) Release(ctx context.Context, handle pool.Handle[any]) error {
	return s.releaseFn(ctx, handle)
}

func (s *stubPoolManager) Purge(ctx context.Context, key pool.Key) error {
	return s.purgeFn(ctx, key)
}

type stubExchanger struct {
	refreshFn func(context.Context, oauth.Spec, string) (oauth.Token, error)
}

func (s *stubExchanger) Exchange(context.Context, oauth.Spec, string, oauth.ExchangeOptions) (oauth.Token, error) {
	return oauth.Token{}, errors.New("not implemented")
}

func (s *stubExchanger) Refresh(ctx context.Context, spec oauth.Spec, refreshToken string) (oauth.Token, error) {
	if s.refreshFn != nil {
		return s.refreshFn(ctx, spec, refreshToken)
	}
	return oauth.Token{}, errors.New("not implemented")
}
