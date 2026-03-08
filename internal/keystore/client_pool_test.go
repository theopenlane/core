package keystore

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/types"
)

type fakeClient struct {
	Token string
}

type fakeCredentialSource struct {
	payload     models.CredentialSet
	mintPayload models.CredentialSet
	getErr      error
	mintErr     error
	getCalls    int
	mintCalls   int
}

func (f *fakeCredentialSource) Get(context.Context, string, types.ProviderType) (models.CredentialSet, error) {
	f.getCalls++
	if f.getErr != nil {
		return models.CredentialSet{}, f.getErr
	}
	return types.CloneCredentialSet(f.payload), nil
}

func (f *fakeCredentialSource) Mint(context.Context, string, types.ProviderType) (models.CredentialSet, error) {
	f.mintCalls++
	if f.mintErr != nil {
		return models.CredentialSet{}, f.mintErr
	}
	if !types.IsCredentialSetEmpty(f.mintPayload) {
		f.payload = types.CloneCredentialSet(f.mintPayload)
		return types.CloneCredentialSet(f.mintPayload), nil
	}
	return types.CloneCredentialSet(f.payload), nil
}

type trackingBuilder struct {
	provider types.ProviderType
	builds   int
}

func (b *trackingBuilder) Build(_ context.Context, payload models.CredentialSet, _ json.RawMessage) (*fakeClient, error) {
	b.builds++
	return &fakeClient{Token: payload.OAuthAccessToken}, nil
}

func (b *trackingBuilder) ProviderType() types.ProviderType {
	return b.provider
}

func TestClientPoolCachesClients(t *testing.T) {
	t.Parallel()

	source := &fakeCredentialSource{
		payload: newOAuthPayload("token-one", time.Now().Add(time.Hour)),
	}
	builder := &trackingBuilder{provider: types.ProviderType("github")}

	pool, err := NewClientPool(source, builder)
	require.NoError(t, err)

	ctx := context.Background()
	client, err := pool.Get(ctx, "org-1")
	require.NoError(t, err)
	require.Equal(t, "token-one", client.Token)
	require.Equal(t, 1, builder.builds)

	client, err = pool.Get(ctx, "org-1")
	require.NoError(t, err)
	require.Equal(t, "token-one", client.Token)
	require.Equal(t, 1, builder.builds, "second call should reuse cached client")
}

func TestClientPoolRefreshesExpiredToken(t *testing.T) {
	t.Parallel()

	expired := newOAuthPayload("token-old", time.Now().Add(-time.Minute))
	refreshed := newOAuthPayload("token-new", time.Now().Add(time.Hour))

	source := &fakeCredentialSource{
		payload:     expired,
		mintPayload: refreshed,
	}
	builder := &trackingBuilder{provider: types.ProviderType("github")}

	pool, err := NewClientPool(source, builder)
	require.NoError(t, err)

	// freeze clock so refresh logic considers the token expired
	pool.now = func() time.Time {
		return time.Now()
	}

	client, err := pool.Get(context.Background(), "org-1")
	require.NoError(t, err)
	require.Equal(t, "token-new", client.Token)
	require.Equal(t, 1, source.mintCalls)
}

func TestClientPoolForceRefresh(t *testing.T) {
	t.Parallel()

	current := newOAuthPayload("token-current", time.Now().Add(time.Hour))
	next := newOAuthPayload("token-next", time.Now().Add(2*time.Hour))

	source := &fakeCredentialSource{
		payload:     current,
		mintPayload: next,
	}
	builder := &trackingBuilder{provider: types.ProviderType("github")}

	pool, err := NewClientPool(source, builder)
	require.NoError(t, err)

	client, err := pool.Get(context.Background(), "org-1", WithClientForceRefresh())
	require.NoError(t, err)
	require.Equal(t, "token-next", client.Token)
	require.Equal(t, 1, source.mintCalls)
}

func newOAuthPayload(token string, expiry time.Time) models.CredentialSet {
	payload := models.CredentialSet{
		OAuthAccessToken:  token,
		OAuthRefreshToken: "refresh-" + token,
		OAuthTokenType:    "bearer",
	}
	payload.OAuthExpiry = &expiry
	return payload
}
