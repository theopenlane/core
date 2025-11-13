package keystore

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/internal/integrations/types"
)

type fakeCredentialSource struct {
	payload     types.CredentialPayload
	mintPayload types.CredentialPayload
	getErr      error
	mintErr     error
	getCalls    int
	mintCalls   int
}

func (f *fakeCredentialSource) Get(context.Context, string, types.ProviderType) (types.CredentialPayload, error) {
	f.getCalls++
	if f.getErr != nil {
		return types.CredentialPayload{}, f.getErr
	}
	return cloneCredentialPayload(f.payload), nil
}

func (f *fakeCredentialSource) Mint(context.Context, string, types.ProviderType) (types.CredentialPayload, error) {
	f.mintCalls++
	if f.mintErr != nil {
		return types.CredentialPayload{}, f.mintErr
	}
	if f.mintPayload.Provider != types.ProviderUnknown {
		f.payload = cloneCredentialPayload(f.mintPayload)
		return cloneCredentialPayload(f.mintPayload), nil
	}
	return cloneCredentialPayload(f.payload), nil
}

type fakeClient struct {
	Token string
	Note  string
}

type fakeConfig struct {
	Note string
}

type trackingBuilder struct {
	provider types.ProviderType
	builds   int
}

func (b *trackingBuilder) Build(_ context.Context, payload types.CredentialPayload, cfg fakeConfig) (*fakeClient, error) {
	b.builds++
	token := ""
	if payload.Token != nil {
		token = payload.Token.AccessToken
	}
	return &fakeClient{
		Token: token,
		Note:  cfg.Note,
	}, nil
}

func (b *trackingBuilder) ProviderType() types.ProviderType {
	return b.provider
}

func TestClientPoolCachesClients(t *testing.T) {
	t.Parallel()

	source := &fakeCredentialSource{
		payload: newOAuthPayload(types.ProviderType("github"), "token-one", time.Now().Add(time.Hour)),
	}
	builder := &trackingBuilder{provider: types.ProviderType("github")}

	pool, err := NewClientPool(source, builder)
	require.NoError(t, err)

	ctx := context.Background()
	client, err := pool.Get(ctx, "org-1", WithClientConfig(fakeConfig{Note: "first"}))
	require.NoError(t, err)
	require.Equal(t, "token-one", client.Token)
	require.Equal(t, 1, builder.builds)

	client, err = pool.Get(ctx, "org-1", WithClientConfig(fakeConfig{Note: "ignored"}))
	require.NoError(t, err)
	require.Equal(t, "token-one", client.Token)
	require.Equal(t, 1, builder.builds, "second call should reuse cached client")
}

func TestClientPoolRefreshesExpiredToken(t *testing.T) {
	t.Parallel()

	expired := newOAuthPayload(types.ProviderType("github"), "token-old", time.Now().Add(-time.Minute))
	refreshed := newOAuthPayload(types.ProviderType("github"), "token-new", time.Now().Add(time.Hour))

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

	current := newOAuthPayload(types.ProviderType("github"), "token-current", time.Now().Add(time.Hour))
	next := newOAuthPayload(types.ProviderType("github"), "token-next", time.Now().Add(2*time.Hour))

	source := &fakeCredentialSource{
		payload:     current,
		mintPayload: next,
	}
	builder := &trackingBuilder{provider: types.ProviderType("github")}

	pool, err := NewClientPool(source, builder)
	require.NoError(t, err)

	client, err := pool.Get(context.Background(), "org-1", WithClientForceRefresh[fakeConfig]())
	require.NoError(t, err)
	require.Equal(t, "token-next", client.Token)
	require.Equal(t, 1, source.mintCalls)
}

func newOAuthPayload(provider types.ProviderType, token string, expiry time.Time) types.CredentialPayload {
	oauthToken := &oauth2.Token{
		AccessToken:  token,
		RefreshToken: "refresh-" + token,
		TokenType:    "bearer",
		Expiry:       expiry,
	}

	payload, _ := types.NewCredentialBuilder(provider).
		With(types.WithOAuthToken(oauthToken)).
		Build()

	return payload
}
