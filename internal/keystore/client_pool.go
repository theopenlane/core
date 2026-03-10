package keystore

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/theopenlane/eddy"

	"github.com/theopenlane/core/common/helpers"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	// defaultClientPoolTTL is the duration that cached clients remain valid
	defaultClientPoolTTL = 5 * time.Minute
)

// CredentialSource exposes the subset of broker operations required by the client pool
type CredentialSource interface {
	// Get retrieves the latest credential payload for the given org/provider pair
	Get(ctx context.Context, orgID string, provider types.ProviderType) (models.CredentialSet, error)
	// Mint obtains a fresh credential payload for the given org/provider pair
	Mint(ctx context.Context, orgID string, provider types.ProviderType) (models.CredentialSet, error)
}

// ClientBuilder constructs provider SDK clients from credential payloads
type ClientBuilder[T any] interface {
	// Build constructs a new client instance using the supplied credential payload and configuration
	Build(ctx context.Context, payload models.CredentialSet, config json.RawMessage) (T, error)
	// ProviderType returns the provider identifier handled by this builder
	ProviderType() types.ProviderType
}

// ClientBuilderFunc adapts a function to the ClientBuilder interface
type ClientBuilderFunc[T any] struct {
	// Provider identifies which provider this builder handles
	Provider types.ProviderType
	// BuildFn is the function that constructs the client
	BuildFn func(context.Context, models.CredentialSet, json.RawMessage) (T, error)
}

// Build constructs the client using the configured function
func (f ClientBuilderFunc[T]) Build(ctx context.Context, payload models.CredentialSet, config json.RawMessage) (T, error) {
	var zero T
	if f.BuildFn == nil {
		return zero, ErrClientBuilderRequired
	}

	return f.BuildFn(ctx, payload, config)
}

// ProviderType returns the provider identifier for the builder
func (f ClientBuilderFunc[T]) ProviderType() types.ProviderType {
	return f.Provider
}

// ClientPool orchestrates credential retrieval and client caching for a specific provider type
type ClientPool[T any] struct {
	// source provides credential retrieval and refresh capabilities
	source CredentialSource
	// provider identifies which provider this pool serves
	provider types.ProviderType
	// builder constructs new client instances
	builder eddy.Builder[T, models.CredentialSet, json.RawMessage]
	// service manages the underlying client cache
	service *eddy.ClientService[T, models.CredentialSet, json.RawMessage]
	// now returns the current time, overridable for testing
	now func() time.Time
}

// ClientPoolOption customizes client pool construction
type ClientPoolOption[T any] func(*ClientPool[T], *clientPoolSettings)

// clientPoolSettings configures client pool behavior
type clientPoolSettings struct {
	// ttl specifies how long clients remain cached
	ttl time.Duration
	// configClone creates copies of configuration objects
	configClone func(json.RawMessage) json.RawMessage
}

// WithClientPoolTTL overrides the default client cache TTL
func WithClientPoolTTL[T any](ttl time.Duration) ClientPoolOption[T] {
	return func(_ *ClientPool[T], settings *clientPoolSettings) {
		if ttl > 0 {
			settings.ttl = ttl
		}
	}
}

// WithClientConfigClone configures how per-request config is cloned before invoking the builder
func WithClientConfigClone[T any](clone func(json.RawMessage) json.RawMessage) ClientPoolOption[T] {
	return func(_ *ClientPool[T], settings *clientPoolSettings) {
		if clone != nil {
			settings.configClone = clone
		}
	}
}

// NewClientPool builds a client pool that reuses provider SDK clients using eddy's caching primitives
func NewClientPool[T any](source CredentialSource, builder ClientBuilder[T], opts ...ClientPoolOption[T]) (*ClientPool[T], error) {
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

	settings := &clientPoolSettings{
		ttl: defaultClientPoolTTL,
		configClone: func(cfg json.RawMessage) json.RawMessage {
			return jsonx.CloneRawMessage(cfg)
		},
	}

	pool := &ClientPool[T]{
		source:   source,
		provider: provider,
		builder:  clientBuilderAdapter[T]{builder: builder},
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
		eddy.WithConfigClone[T, models.CredentialSet](settings.configClone),
		eddy.WithOutputClone[T, models.CredentialSet, json.RawMessage](types.CloneCredentialSet),
	)

	return pool, nil
}

// Provider returns the provider type handled by this pool
func (p *ClientPool[T]) Provider() types.ProviderType {
	return p.provider
}

// ClientRequestOption customizes Get requests
type ClientRequestOption func(*clientRequest)

// WithClientConfig supplies provider-specific builder configuration
func WithClientConfig(config json.RawMessage) ClientRequestOption {
	return func(req *clientRequest) {
		req.config = config
	}
}

// WithClientForceRefresh bypasses cached credentials and forces a mint operation
func WithClientForceRefresh() ClientRequestOption {
	return func(req *clientRequest) {
		req.forceRefresh = true
	}
}

// clientRequest captures the parameters for a single client retrieval
type clientRequest struct {
	// orgID identifies the organization requesting the client
	orgID string
	// config provides provider-specific configuration
	config json.RawMessage
	// forceRefresh bypasses caches and forces credential refresh
	forceRefresh bool
}

// Get returns a provider-specific client for the supplied organization, reusing cached instances when possible
func (p *ClientPool[T]) Get(ctx context.Context, orgID string, opts ...ClientRequestOption) (T, error) {
	var zero T

	req := clientRequest{
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
func (p *ClientPool[T]) resolveCredential(ctx context.Context, req clientRequest) (models.CredentialSet, error) {
	// sorry this is hard to read its more work than its worth to change it
	return resolveCredentialWithPolicy(
		ctx,
		req.forceRefresh,
		p.now,
		func(callCtx context.Context) (models.CredentialSet, error) {
			return p.source.Get(callCtx, req.orgID, p.provider)
		},
		func(callCtx context.Context, previous models.CredentialSet) (models.CredentialSet, error) {
			return p.refreshCredential(callCtx, req.orgID, previous)
		},
	)
}

// refreshCredential obtains a fresh credential and evicts stale cache entries
func (p *ClientPool[T]) refreshCredential(ctx context.Context, orgID string, previous models.CredentialSet) (models.CredentialSet, error) {
	p.evict(orgID, credentialVersion(previous))

	refreshed, err := p.source.Mint(ctx, orgID, p.provider)
	if err != nil {
		return models.CredentialSet{}, err
	}

	return refreshed, nil
}

// evict removes a client from the cache by org and credential version
func (p *ClientPool[T]) evict(orgID, version string) {
	if orgID == "" {
		return
	}

	p.service.Pool().RemoveClient(clientCacheKey{
		OrgID:    orgID,
		Provider: p.provider,
		Version:  version,
	})
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
type clientBuilderAdapter[T any] struct {
	// builder is the underlying client builder
	builder ClientBuilder[T]
}

// Build constructs a client using the adapted builder
func (a clientBuilderAdapter[T]) Build(ctx context.Context, payload models.CredentialSet, config json.RawMessage) (T, error) {
	return a.builder.Build(ctx, payload, config)
}

// ProviderType returns the provider identifier as a string
func (a clientBuilderAdapter[T]) ProviderType() string {
	return string(a.builder.ProviderType())
}

// credentialVersion computes a hash representing the credential's content
func credentialVersion(payload models.CredentialSet) string {
	if types.IsCredentialSetEmpty(payload) {
		return ""
	}

	builder := helpers.NewHashBuilder()
	builder.WriteStrings(
		payload.AccessKeyID,
		payload.SecretAccessKey,
		payload.SessionToken,
		payload.ClientID,
		payload.ClientSecret,
		payload.ServiceAccountKey,
		payload.SubjectToken,
		payload.ProjectID,
		payload.AccountID,
		payload.APIToken,
		payload.OAuthAccessToken,
		payload.OAuthRefreshToken,
		payload.OAuthTokenType,
	).WriteTimePtr(payload.OAuthExpiry)

	builder.WriteStrings(string(payload.ProviderData)).
		WriteSortedMap(payload.Claims)

	return builder.Hex()
}
