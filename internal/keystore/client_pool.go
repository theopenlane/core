package keystore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"sort"
	"strings"
	"time"

	"github.com/theopenlane/eddy"

	"github.com/theopenlane/core/internal/integrations/helpers"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/models"
)

const (
	defaultClientPoolTTL = 5 * time.Minute
	refreshSkew          = cacheSkew
)

// CredentialSource exposes the subset of broker operations required by the client pool
type CredentialSource interface {
	Get(ctx context.Context, orgID string, provider types.ProviderType) (types.CredentialPayload, error)
	Mint(ctx context.Context, orgID string, provider types.ProviderType) (types.CredentialPayload, error)
}

// ClientBuilder constructs provider SDK clients from credential payloads
type ClientBuilder[T any, Config any] interface {
	Build(ctx context.Context, payload types.CredentialPayload, config Config) (T, error)
	ProviderType() types.ProviderType
}

// ClientBuilderFunc adapts a function to the ClientBuilder interface
type ClientBuilderFunc[T any, Config any] struct {
	Provider types.ProviderType
	BuildFn  func(context.Context, types.CredentialPayload, Config) (T, error)
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
	source   CredentialSource
	provider types.ProviderType
	builder  eddy.Builder[T, types.CredentialPayload, Config]
	service  *eddy.ClientService[T, types.CredentialPayload, Config]
	now      func() time.Time
}

// ClientPoolOption customizes client pool construction
type ClientPoolOption[T any, Config any] func(*ClientPool[T, Config], *clientPoolSettings[Config])

type clientPoolSettings[Config any] struct {
	ttl         time.Duration
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

type clientRequest[Config any] struct {
	orgID        string
	config       Config
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

func shouldRefreshCredential(payload types.CredentialPayload, now func() time.Time) bool {
	if payload.Token == nil || payload.Token.Expiry.IsZero() {
		return false
	}

	refreshAt := payload.Token.Expiry.Add(-refreshSkew)
	return now().After(refreshAt)
}

type clientCacheKey struct {
	OrgID    string
	Provider types.ProviderType
	Version  string
}

func (k clientCacheKey) String() string {
	base := fmt.Sprintf("%s:%s", k.OrgID, k.Provider)
	version := strings.TrimSpace(k.Version)
	if version == "" {
		return base
	}
	return base + ":" + version
}

type clientBuilderAdapter[T any, Config any] struct {
	builder ClientBuilder[T, Config]
}

func (a clientBuilderAdapter[T, Config]) Build(ctx context.Context, payload types.CredentialPayload, config Config) (T, error) {
	return a.builder.Build(ctx, payload, config)
}

func (a clientBuilderAdapter[T, Config]) ProviderType() string {
	return string(a.builder.ProviderType())
}

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

func cloneCredentialSet(set models.CredentialSet) models.CredentialSet {
	clone := set
	if set.ProviderData != nil {
		clone.ProviderData = maps.Clone(set.ProviderData)
	}
	if set.Claims != nil {
		clone.Claims = maps.Clone(set.Claims)
	}
	if set.OAuthExpiry != nil {
		expiry := *set.OAuthExpiry
		clone.OAuthExpiry = &expiry
	}
	return clone
}

func credentialVersion(payload types.CredentialPayload) string {
	if payload.Provider == types.ProviderUnknown {
		return ""
	}

	hasher := sha256.New()
	write := func(values ...string) {
		for _, value := range values {
			if value == "" {
				continue
			}
			_, _ = hasher.Write([]byte(value))
		}
	}

	write(string(payload.Provider), string(payload.Kind))

	if payload.Token != nil {
		write(payload.Token.AccessToken, payload.Token.RefreshToken, payload.Token.TokenType)
		if !payload.Token.Expiry.IsZero() {
			write(payload.Token.Expiry.UTC().Format(time.RFC3339Nano))
		}
	}

	write(
		payload.Data.AccessKeyID,
		payload.Data.SecretAccessKey,
		payload.Data.ProjectID,
		payload.Data.AccountID,
		payload.Data.APIToken,
		payload.Data.OAuthAccessToken,
		payload.Data.OAuthRefreshToken,
		payload.Data.OAuthTokenType,
	)

	if payload.Data.OAuthExpiry != nil {
		write(payload.Data.OAuthExpiry.UTC().Format(time.RFC3339Nano))
	}

	if len(payload.Data.ProviderData) > 0 {
		keys := make([]string, 0, len(payload.Data.ProviderData))
		for key := range payload.Data.ProviderData {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			write(key)
			value := payload.Data.ProviderData[key]
			switch v := value.(type) {
			case string:
				write(v)
			default:
				if encoded, err := json.Marshal(v); err == nil {
					_, _ = hasher.Write(encoded)
				}
			}
		}
	}

	if len(payload.Data.Claims) > 0 {
		keys := make([]string, 0, len(payload.Data.Claims))
		for key := range payload.Data.Claims {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			write(key)
			if encoded, err := json.Marshal(payload.Data.Claims[key]); err == nil {
				_, _ = hasher.Write(encoded)
			}
		}
	}

	return hex.EncodeToString(hasher.Sum(nil))
}
