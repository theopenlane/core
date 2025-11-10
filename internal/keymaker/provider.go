package keymaker

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/theopenlane/core/internal/keymaker/oauth"
	"github.com/theopenlane/core/internal/keymaker/pool"
	"github.com/theopenlane/core/internal/keymaker/store"
	"github.com/theopenlane/core/internal/keystore"
)

// ClientFactory creates provider-specific clients for a session.
type ClientFactory interface {
	Provider() string
	New(ctx context.Context, req SessionRequest, creds Credentials) (any, error)
}

// ProviderOptions configures optional dependencies for the service.
type ProviderOptions struct {
	Specs     []oauth.Spec
	Factories []ClientFactory
	Pools     map[string]pool.Manager[any]
	Validator oauth.Validator
}

// Service implements the Provider interface backed by store/exchange/pool deps.
type Service struct {
	store     store.Repository
	exchanger oauth.Exchanger
	validator oauth.Validator

	specs     map[string]oauth.Spec
	factories map[string]ClientFactory

	poolsMu sync.Mutex
	pools   map[string]pool.Manager[any]
}

// NewService constructs a Provider implementation from the supplied dependencies.
func NewService(repo store.Repository, exchanger oauth.Exchanger, opts ProviderOptions) *Service {
	specs := make(map[string]oauth.Spec)
	for _, spec := range opts.Specs {
		specs[strings.ToLower(spec.Name)] = spec
	}

	factories := make(map[string]ClientFactory)
	for _, factory := range opts.Factories {
		factories[strings.ToLower(factory.Provider())] = factory
	}

	pools := opts.Pools
	if pools == nil {
		pools = make(map[string]pool.Manager[any])
	}

	return &Service{
		store:     repo,
		exchanger: exchanger,
		validator: opts.Validator,
		specs:     specs,
		factories: factories,
		pools:     pools,
	}
}

// Session materializes a session for the requested provider/org combination.
func (s *Service) Session(ctx context.Context, req SessionRequest) (Session, error) {
	if req.OrgID == "" {
		return Session{}, ErrMissingOrgID
	}
	if req.Integration == "" {
		return Session{}, ErrMissingIntegration
	}
	if req.Provider == "" {
		return Session{}, ErrMissingProvider
	}

	providerKey := strings.ToLower(req.Provider)
	spec, ok := s.specs[providerKey]
	if !ok {
		spec = oauth.Spec{Name: req.Provider}
	}

	record, err := s.store.Integration(ctx, req.OrgID, req.Integration)
	if err != nil {
		return Session{}, fmt.Errorf("keymaker: fetch integration: %w", err)
	}

	if s.validator != nil {
		if err := s.validator.ValidateActivation(ctx, spec, record.Metadata); err != nil {
			return Session{}, fmt.Errorf("keymaker: activation validation failed: %w", err)
		}
	}

	credRecord, err := s.store.Credentials(ctx, req.OrgID, req.Integration)
	if err != nil {
		return Session{}, fmt.Errorf("keymaker: load credentials: %w", err)
	}

	creds, err := decodeCredentials(record.Provider, credRecord)
	if err != nil {
		return Session{}, fmt.Errorf("keymaker: decode credentials: %w", err)
	}
	if len(req.Scopes) == 0 {
		creds.Scopes = append([]string{}, record.Scopes...)
	} else {
		creds.Scopes = append([]string{}, req.Scopes...)
	}

	factory, ok := s.factories[providerKey]
	if !ok {
		return Session{}, fmt.Errorf("%w: %s", ErrFactoryNotRegistered, req.Provider)
	}

	poolManager, ok := s.poolManager(providerKey)
	if !ok {
		poolManager = pool.NewSimpleManager[any]()
		s.setPoolManager(providerKey, poolManager)
	}

	scopeSet := creds.Scopes
	if len(scopeSet) == 0 && len(record.Scopes) > 0 {
		scopeSet = record.Scopes
	}

	handle, err := poolManager.Acquire(ctx, pool.Key{
		OrgID:     req.OrgID,
		Provider:  providerKey,
		ScopeHash: hashScopes(scopeSet),
	}, poolFactoryAdapter{
		factory: factory,
		req:     req,
		creds:   creds,
	})
	if err != nil {
		return Session{}, fmt.Errorf("keymaker: acquire client: %w", err)
	}

	now := time.Now()
	refreshAfter := now
	if !creds.Expiry.IsZero() {
		refreshAfter = creds.Expiry.Add(-time.Minute)
	}

	session := Session{
		OrgID:        req.OrgID,
		Provider:     providerKey,
		Credentials:  creds,
		Client:       handle.Client,
		IssuedAt:     now,
		RefreshAfter: refreshAfter,
		ExpiresAt:    creds.Expiry,
	}
	session.Metadata = cloneMetadata(record.Metadata)

	if s.exchanger != nil && creds.RefreshToken != "" {
		session.RefreshHandle = &refreshHandle{
			service:    s,
			req:        req,
			spec:       spec,
			credRecord: credRecord,
			provider:   providerKey,
		}
	}

	session.releaseFn = func(releaseCtx context.Context) error {
		return poolManager.Release(releaseCtx, handle)
	}

	return session, nil
}

// Release returns the session to the underlying pool.
func (s *Service) Release(ctx context.Context, session Session) error {
	if session.releaseFn == nil {
		return nil
	}
	return session.releaseFn(ctx)
}

func (s *Service) poolManager(provider string) (pool.Manager[any], bool) {
	s.poolsMu.Lock()
	defer s.poolsMu.Unlock()

	manager, ok := s.pools[provider]
	return manager, ok
}

func (s *Service) setPoolManager(provider string, manager pool.Manager[any]) {
	s.poolsMu.Lock()
	defer s.poolsMu.Unlock()

	s.pools[provider] = manager
}

func hashScopes(scopes []string) string {
	if len(scopes) == 0 {
		return ""
	}
	cp := make([]string, len(scopes))
	copy(cp, scopes)
	sort.Strings(cp)
	return strings.Join(cp, " ")
}

func cloneMetadata(src map[string]any) map[string]any {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

type poolFactoryAdapter struct {
	factory ClientFactory
	req     SessionRequest
	creds   Credentials
}

func (p poolFactoryAdapter) New(ctx context.Context, _ pool.Key) (any, error) {
	return p.factory.New(ctx, p.req, p.creds)
}

func decodeCredentials(provider string, record store.CredentialRecord) (Credentials, error) {
	payload := make(map[string]string)
	if len(record.Encrypted) > 0 {
		if err := json.Unmarshal(record.Encrypted, &payload); err != nil {
			return Credentials{}, err
		}
	}

	prefix := provider + keystore.SecretNameSeparator
	read := func(key string) string {
		return payload[prefix+key]
	}

	var expiry time.Time
	if raw := read(keystore.ExpiresAtField); raw != "" {
		if ts, err := time.Parse(time.RFC3339, raw); err == nil {
			expiry = ts
		}
	}

	return Credentials{
		AccessToken:  read(keystore.AccessTokenField),
		RefreshToken: read(keystore.RefreshTokenField),
		TokenType:    "Bearer",
		Expiry:       expiry,
		Raw: map[string]any{
			"payload": payload,
			"version": record.Version,
		},
	}, nil
}

type refreshHandle struct {
	service    *Service
	req        SessionRequest
	spec       oauth.Spec
	provider   string
	credRecord store.CredentialRecord
}

func (h *refreshHandle) Refresh(ctx context.Context) (Credentials, error) {
	if h.service.exchanger == nil {
		return Credentials{}, ErrRefreshUnsupported
	}

	payload := make(map[string]string)
	if len(h.credRecord.Encrypted) > 0 {
		if err := json.Unmarshal(h.credRecord.Encrypted, &payload); err != nil {
			return Credentials{}, err
		}
	}

	providerName := h.provider
	if providerName == "" {
		providerName = strings.ToLower(h.credRecord.Provider)
	}

	prefix := providerName + keystore.SecretNameSeparator
	refreshToken := payload[prefix+keystore.RefreshTokenField]
	if refreshToken == "" {
		return Credentials{}, ErrRefreshTokenMissing
	}

	token, err := h.service.exchanger.Refresh(ctx, h.spec, refreshToken)
	if err != nil {
		return Credentials{}, err
	}

	payload[prefix+keystore.AccessTokenField] = token.AccessToken
	if token.RefreshToken != "" {
		payload[prefix+keystore.RefreshTokenField] = token.RefreshToken
	}
	if !token.Expiry.IsZero() {
		payload[prefix+keystore.ExpiresAtField] = token.Expiry.Format(time.RFC3339)
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return Credentials{}, err
	}

	h.credRecord.Encrypted = encoded
	h.credRecord.Version++

	if err := h.service.store.SaveCredentials(ctx, h.credRecord); err != nil {
		return Credentials{}, err
	}

	return Credentials{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
		Scopes:       h.req.Scopes,
		Raw: map[string]any{
			"payload": payload,
			"version": h.credRecord.Version,
		},
	}, nil
}
