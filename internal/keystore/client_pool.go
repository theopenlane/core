package keystore

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/theopenlane/eddy"

	"github.com/theopenlane/core/common/helpers"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
)

const (
	// defaultClientPoolTTL is the duration that cached clients remain valid
	defaultClientPoolTTL = 5 * time.Minute
	// refreshSkew defines how far before token expiry refresh operations begin
	refreshSkew = cacheSkew
)

// CredentialSource exposes the subset of broker operations required by the client pool
type CredentialSource interface {
	// Get retrieves the latest credential payload for the given org/provider pair
	Get(ctx context.Context, orgID string, provider types.ProviderType) (types.CredentialPayload, error)
	// Mint obtains a fresh credential payload for the given org/provider pair
	Mint(ctx context.Context, orgID string, provider types.ProviderType) (types.CredentialPayload, error)
}

// ClientBuilder constructs provider SDK clients from credential payloads
type ClientBuilder[T any, Config any] interface {
	// Build constructs a new client instance using the supplied credential payload and configuration
	Build(ctx context.Context, payload types.CredentialPayload, config Config) (T, error)
	// ProviderType returns the provider identifier handled by this builder
	ProviderType() types.ProviderType
}

// ClientBuilderFunc adapts a function to the ClientBuilder interface
type ClientBuilderFunc[T any, Config any] struct {
	// Provider identifies which provider this builder handles
	Provider types.ProviderType
	// BuildFn is the function that constructs the client
	BuildFn func(context.Context, types.CredentialPayload, Config) (T, error)
}

// Build constructs the client using the configured function
func (f ClientBuilderFunc[T, Config]) Build(ctx context.Context, payload types.CredentialPayload, config Config) (T, error) {
	var zero T
	if f.BuildFn == nil {
		return zero, ErrClientBuilderRequired
	}

	return f.BuildFn(ctx, payload, config)
}

// ProviderType returns the provider identifier for the builder
func (f ClientBuilderFunc[T, Config]) ProviderType() types.ProviderType {
	return f.Provider
}

// ClientPool orchestrates credential retrieval and client caching for a specific provider type
type ClientPool[T any, Config any] struct {
	// source provides credential retrieval and refresh capabilities
	source CredentialSource
	// provider identifies which provider this pool serves
	provider types.ProviderType
	// builder constructs new client instances
	builder eddy.Builder[T, types.CredentialPayload, Config]
	// service manages the underlying client cache
	service *eddy.ClientService[T, types.CredentialPayload, Config]
	// now returns the current time, overridable for testing
	now func() time.Time
}

// ClientPoolOption customizes client pool construction
type ClientPoolOption[T any, Config any] func(*ClientPool[T, Config], *clientPoolSettings[Config])

// clientPoolSettings configures client pool behavior
type clientPoolSettings[Config any] struct {
	// ttl specifies how long clients remain cached
	ttl time.Duration
	// configClone creates copies of configuration objects
	configClone func(Config) Config
}

// WithClientPoolTTL overrides the default client cache TTL
func WithClientPoolTTL[T any, Config any](ttl time.Duration) ClientPoolOption[T, Config] {
	return func(_ *ClientPool[T, Config], settings *clientPoolSettings[Config]) {
		if ttl > 0 {
			settings.ttl = ttl
		}
	}
}

// WithClientConfigClone configures how per-request config structs are cloned before invoking the builder
func WithClientConfigClone[T any, Config any](clone func(Config) Config) ClientPoolOption[T, Config] {
	return func(_ *ClientPool[T, Config], settings *clientPoolSettings[Config]) {
		if clone != nil {
			settings.configClone = clone
		}
	}
}

// NewClientPool builds a client pool that reuses provider SDK clients using eddy's caching primitives
func NewClientPool[T any, Config any](source CredentialSource, builder ClientBuilder[T, Config], opts ...ClientPoolOption[T, Config]) (*ClientPool[T, Config], error) {
	if source == nil {
		return nil, ErrBrokerRequired
	}
	if builder == nil {
		return nil, ErrClientBuilderRequired
	}

	provider := builder.ProviderType()
	if provider == types.ProviderUnknown {
		return nil, ErrProviderRequired
	}

	settings := &clientPoolSettings[Config]{
		ttl: defaultClientPoolTTL,
		configClone: func(cfg Config) Config {
			return cfg
		},
	}

	pool := &ClientPool[T, Config]{
		source:   source,
		provider: provider,
		builder:  clientBuilderAdapter[T, Config]{builder: builder},
		now:      time.Now,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(pool, settings)
		}
	}

	eddyPool := eddy.NewClientPool[T](settings.ttl)
	pool.service = eddy.NewClientService(
		eddyPool,
		eddy.WithConfigClone[T, types.CredentialPayload, Config](settings.configClone),
		eddy.WithOutputClone[T, types.CredentialPayload, Config](cloneCredentialPayload),
	)

	return pool, nil
}

// Provider returns the provider type handled by this pool
func (p *ClientPool[T, Config]) Provider() types.ProviderType {
	if p == nil {
		return types.ProviderUnknown
	}

	return p.provider
}

// ClientRequestOption customizes Get requests
type ClientRequestOption[Config any] func(*clientRequest[Config])

// WithClientConfig supplies provider-specific builder configuration
func WithClientConfig[Config any](config Config) ClientRequestOption[Config] {
	return func(req *clientRequest[Config]) {
		req.config = config
	}
}

// WithClientForceRefresh bypasses cached credentials and forces a mint operation
func WithClientForceRefresh[Config any]() ClientRequestOption[Config] {
	return func(req *clientRequest[Config]) {
		req.forceRefresh = true
	}
}

// clientRequest captures the parameters for a single client retrieval
type clientRequest[Config any] struct {
	// orgID identifies the organization requesting the client
	orgID string
	// config provides provider-specific configuration
	config Config
	// forceRefresh bypasses caches and forces credential refresh
	forceRefresh bool
}

// Get returns a provider-specific client for the supplied organization, reusing cached instances when possible
func (p *ClientPool[T, Config]) Get(ctx context.Context, orgID string, opts ...ClientRequestOption[Config]) (T, error) {
	var zero T
	if p == nil {
		return zero, ErrClientUnavailable
	}

	req := clientRequest[Config]{
		orgID: strings.TrimSpace(orgID),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&req)
		}
	}

	if req.orgID == "" {
		return zero, ErrOrgIDRequired
	}

	payload, err := p.resolveCredential(ctx, req)
	if err != nil {
		return zero, err
	}

	version := credentialVersion(payload)
	key := clientCacheKey{
		OrgID:    req.orgID,
		Provider: p.provider,
		Version:  version,
	}

	client := p.service.GetClient(ctx, key, p.builder, payload, req.config)
	if !client.IsPresent() {
		return zero, ErrClientUnavailable
	}

	return client.MustGet(), nil
}

// resolveCredential retrieves or refreshes the credential for the given request
func (p *ClientPool[T, Config]) resolveCredential(ctx context.Context, req clientRequest[Config]) (types.CredentialPayload, error) {
	if req.forceRefresh {
		return p.refreshCredential(ctx, req.orgID, types.CredentialPayload{})
	}

	payload, err := p.source.Get(ctx, req.orgID, p.provider)
	if err != nil {
		if errors.Is(err, ErrCredentialNotFound) {
			return p.refreshCredential(ctx, req.orgID, types.CredentialPayload{})
		}
		return types.CredentialPayload{}, err
	}

	if shouldRefreshCredential(payload, p.now) {
		return p.refreshCredential(ctx, req.orgID, payload)
	}

	return payload, nil
}

// refreshCredential obtains a fresh credential and evicts stale cache entries
func (p *ClientPool[T, Config]) refreshCredential(ctx context.Context, orgID string, previous types.CredentialPayload) (types.CredentialPayload, error) {
	if previous.Provider != types.ProviderUnknown {
		p.evict(orgID, credentialVersion(previous))
	}

	refreshed, err := p.source.Mint(ctx, orgID, p.provider)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	return refreshed, nil
}

// evict removes a client from the cache by org and credential version
func (p *ClientPool[T, Config]) evict(orgID, version string) {
	if orgID == "" {
		return
	}
	p.service.Pool().RemoveClient(clientCacheKey{
		OrgID:    orgID,
		Provider: p.provider,
		Version:  version,
	})
}

// shouldRefreshCredential determines if a credential needs refreshing based on its expiry
func shouldRefreshCredential(payload types.CredentialPayload, now func() time.Time) bool {
	if payload.Token == nil || payload.Token.Expiry.IsZero() {
		return false
	}

	refreshAt := payload.Token.Expiry.Add(-refreshSkew)
	return now().After(refreshAt)
}

// clientCacheKey uniquely identifies a cached client instance
type clientCacheKey struct {
	// OrgID identifies the organization owning the client
	OrgID string
	// Provider identifies which provider issued the client
	Provider types.ProviderType
	// Version tracks credential version to invalidate stale clients
	Version string
}

// String returns a string representation of the cache key
func (k clientCacheKey) String() string {
	base := fmt.Sprintf("%s:%s", k.OrgID, k.Provider)
	version := strings.TrimSpace(k.Version)
	if version == "" {
		return base
	}
	return base + ":" + version
}

// clientBuilderAdapter adapts ClientBuilder to eddy's Builder interface
type clientBuilderAdapter[T any, Config any] struct {
	// builder is the underlying client builder
	builder ClientBuilder[T, Config]
}

// Build constructs a client using the adapted builder
func (a clientBuilderAdapter[T, Config]) Build(ctx context.Context, payload types.CredentialPayload, config Config) (T, error) {
	return a.builder.Build(ctx, payload, config)
}

// ProviderType returns the provider identifier as a string
func (a clientBuilderAdapter[T, Config]) ProviderType() string {
	return string(a.builder.ProviderType())
}

// cloneCredentialPayload creates a deep copy of a credential payload
func cloneCredentialPayload(payload types.CredentialPayload) types.CredentialPayload {
	clone := types.CredentialPayload{
		Provider: payload.Provider,
		Kind:     payload.Kind,
		Token:    helpers.CloneOAuthToken(payload.Token),
		Claims:   helpers.CloneOIDCClaims(payload.Claims),
		Data:     cloneCredentialSet(payload.Data),
	}
	return clone
}

// cloneCredentialSet creates a deep copy of a credential set
func cloneCredentialSet(set models.CredentialSet) models.CredentialSet {
	clone := set
	clone.ProviderData = helpers.DeepCloneMap(set.ProviderData)
	clone.Claims = helpers.DeepCloneMap(set.Claims)
	if set.OAuthExpiry != nil {
		expiry := *set.OAuthExpiry
		clone.OAuthExpiry = &expiry
	}
	return clone
}

// credentialVersion computes a hash representing the credential's content
func credentialVersion(payload types.CredentialPayload) string {
	if payload.Provider == types.ProviderUnknown {
		return ""
	}

	builder := helpers.NewHashBuilder().
		WriteStrings(string(payload.Provider), string(payload.Kind))

	if payload.Token != nil {
		builder.WriteStrings(
			payload.Token.AccessToken,
			payload.Token.RefreshToken,
			payload.Token.TokenType,
		).WriteTime(payload.Token.Expiry)
	}

	builder.WriteStrings(
		payload.Data.AccessKeyID,
		payload.Data.SecretAccessKey,
		payload.Data.SessionToken,
		payload.Data.ProjectID,
		payload.Data.AccountID,
		payload.Data.APIToken,
		payload.Data.OAuthAccessToken,
		payload.Data.OAuthRefreshToken,
		payload.Data.OAuthTokenType,
	).WriteTimePtr(payload.Data.OAuthExpiry)

	builder.WriteSortedMap(payload.Data.ProviderData).
		WriteSortedMap(payload.Data.Claims)

	return builder.Hex()
}
