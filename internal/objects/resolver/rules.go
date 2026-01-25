package resolver

import (
	"context"

	"github.com/theopenlane/eddy"
	"github.com/theopenlane/eddy/helpers"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
)

// providerBuilder is a type alias for readability
type providerBuilder = eddy.Builder[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]

// providerBuilders groups the provider builders required to assemble resolver rules
type providerBuilders struct {
	s3   providerBuilder
	r2   providerBuilder
	disk providerBuilder
	db   providerBuilder
}

// RuleOption configures aspects of ruleCoordinator
type RuleOption func(*ruleCoordinator)

// configureProviderRules adds the resolver rules that determine which provider to use for a request
func configureProviderRules(resolver *providerResolver, opts ...RuleOption) {
	coordinator := newRuleCoordinator(resolver, opts...)
	coordinator.configure()
}

// ruleCoordinator groups state required to add resolver rules in a readable way
type ruleCoordinator struct {
	resolver *providerResolver
	config   storage.ProviderConfig
	builders providerBuilders
	runtime  serviceOptions
}

// newRuleCoordinator returns a helper for building provider rules
func newRuleCoordinator(resolver *providerResolver, opts ...RuleOption) *ruleCoordinator {
	rc := &ruleCoordinator{resolver: resolver}

	for _, opt := range opts {
		if opt == nil {
			continue
		}

		opt(rc)
	}

	return rc
}

// configure wires together the configured set of provider rules.
func (rc *ruleCoordinator) configure() {
	if rc.handleDevMode() {
		return
	}

	rc.addKnownProviderRule()
	rc.addTemplateKindRule(enums.TemplateKindTrustCenterNda, storage.R2Provider)
	rc.addModuleRule(models.CatalogTrustCenterModule, storage.R2Provider)
	rc.addModuleRule(models.CatalogComplianceModule, storage.S3Provider)
	rc.addDefaultProviderRule()
}

// WithProviderConfig supplies the provider configuration used when resolving options.
func WithProviderConfig(config storage.ProviderConfig) RuleOption {
	return func(rc *ruleCoordinator) {
		rc.config = config
	}
}

// WithProviderBuilders supplies the provider builders used for rule resolution.
func WithProviderBuilders(builders providerBuilders) RuleOption {
	return func(rc *ruleCoordinator) {
		rc.builders = builders
	}
}

// WithRuntimeOptions supplies runtime configuration such as token manager hooks.
func WithRuntimeOptions(runtime serviceOptions) RuleOption {
	return func(rc *ruleCoordinator) {
		rc.runtime = runtime
	}
}

// getBuilder returns the appropriate builder for a provider type
func (rc *ruleCoordinator) getBuilder(provider storage.ProviderType) providerBuilder {
	switch provider {
	case storage.S3Provider:
		return rc.builders.s3
	case storage.R2Provider:
		return rc.builders.r2
	case storage.DiskProvider:
		return rc.builders.disk
	case storage.DatabaseProvider:
		return rc.builders.db
	default:
		return nil
	}
}

func devModeOptions() *storage.ProviderOptions {
	return storage.NewProviderOptions(
		storage.WithBucket(objects.DefaultDevStorageBucket),
		storage.WithBasePath(objects.DefaultDevStorageBucket),
		storage.WithProxyPresignEnabled(true),
		storage.WithEndpoint(objects.DefaultLocalDiskURL),
		storage.WithProxyPresignConfig(&storage.ProxyPresignConfig{
			BaseURL: objects.DefaultLocalDiskURL,
		}),
		storage.WithLocalURL(objects.DefaultLocalDiskURL),
		storage.WithExtra("dev_mode", true),
	)
}

// handleDevMode adds a dev-only disk rule when configured and returns true if handled
func (rc *ruleCoordinator) handleDevMode() bool {
	if !rc.config.DevMode {
		return false
	}

	devRule := &helpers.ConditionalRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Predicate: func(_ context.Context) bool {
			return true
		},
		Resolver: func(_ context.Context) (*eddy.ResolvedProvider[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions], error) {
			return &eddy.ResolvedProvider[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
				Builder: rc.builders.disk,
				Output:  storage.ProviderCredentials{},
				Config:  devModeOptions().Clone(),
			}, nil
		},
	}

	rc.resolver.AddRule(devRule)

	return true
}

// addKnownProviderRule resolves providers when a known provider hint is supplied
func (rc *ruleCoordinator) addKnownProviderRule() {
	rule := &helpers.ConditionalRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Predicate: func(ctx context.Context) bool {
			known, ok := contextx.From[objects.KnownProviderHint](ctx)
			return ok && rc.providerEnabled(storage.ProviderType(known))
		},
		Resolver: func(ctx context.Context) (*eddy.ResolvedProvider[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions], error) {
			known, _ := contextx.From[objects.KnownProviderHint](ctx)
			provider := storage.ProviderType(known)
			return rc.resolveProviderWithBuilder(provider)
		},
	}

	rc.resolver.AddRule(rule)
}

// addModuleRule routes requests for a specific module to the desired provider
func (rc *ruleCoordinator) addModuleRule(module models.OrgModule, provider storage.ProviderType) {
	if !rc.providerEnabled(provider) {
		return
	}

	moduleProvider := provider
	rule := &helpers.ConditionalRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Predicate: func(ctx context.Context) bool {
			hint, ok := contextx.From[objects.ModuleHint](ctx)
			return ok && models.OrgModule(hint) == module
		},
		Resolver: func(_ context.Context) (*eddy.ResolvedProvider[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions], error) {
			return rc.resolveProviderWithBuilder(moduleProvider)
		},
	}

	rc.resolver.AddRule(rule)
}

// addTemplateKindRule routes requests for a specific template kind to the desired provider
func (rc *ruleCoordinator) addTemplateKindRule(kind enums.TemplateKind, provider storage.ProviderType) {
	if !rc.providerEnabled(provider) || kind == "" {
		return
	}

	templateKind := kind
	rule := &helpers.ConditionalRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Predicate: func(ctx context.Context) bool {
			hint, ok := contextx.From[objects.TemplateKindHint](ctx)
			return ok && enums.TemplateKind(hint) == templateKind
		},
		Resolver: func(_ context.Context) (*eddy.ResolvedProvider[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions], error) {
			return rc.resolveProviderWithBuilder(provider)
		},
	}

	rc.resolver.AddRule(rule)
}

// addDefaultProviderRule resolves to the first enabled provider when no other rule applies.
func (rc *ruleCoordinator) addDefaultProviderRule() {
	defaultProvider, ok := rc.defaultProvider()
	if !ok {
		return
	}

	rule := &helpers.ConditionalRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Predicate: func(_ context.Context) bool { return true },
		Resolver: func(_ context.Context) (*eddy.ResolvedProvider[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions], error) {
			return rc.resolveProviderWithBuilder(defaultProvider)
		},
	}

	rc.resolver.AddRule(rule)
}

// defaultProvider determines the provider to use when no other rule applies.
func (rc *ruleCoordinator) defaultProvider() (storage.ProviderType, bool) {
	for _, provider := range []storage.ProviderType{
		storage.S3Provider,
		storage.R2Provider,
		storage.DiskProvider,
		storage.DatabaseProvider,
	} {
		if rc.providerEnabled(provider) {
			return provider, true
		}
	}

	return "", false
}

// resolveProviderWithBuilder resolves provider credentials and returns them with the appropriate builder
func (rc *ruleCoordinator) resolveProviderWithBuilder(provider storage.ProviderType) (*eddy.ResolvedProvider[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions], error) {
	resolved, err := rc.resolveProvider(provider)
	if err != nil {
		return nil, err
	}

	builder := rc.getBuilder(provider)
	if builder == nil {
		return nil, errUnsupportedProvider
	}

	return &eddy.ResolvedProvider[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Builder: builder,
		Output:  resolved.Output,
		Config:  resolved.Config,
	}, nil
}
