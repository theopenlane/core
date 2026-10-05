package runtime

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/catalog"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/keymaker"
	"github.com/theopenlane/core/v2/internal/keystore"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/singleton"
)

const (
	// defaultLookbackDuration is used when no last successful run is recorded for an operation
	defaultLookbackDuration = 90 * 24 * time.Hour
)

// Config defines the dependencies required to build the integrations runtime
type Config struct {
	// DB is the Ent client used by run stores and direct installation queries
	DB *ent.Client
	// Gala is the event runtime used for operation dispatch and execution
	Gala *gala.Gala
	// Registry overrides the default empty definition registry when provided
	Registry *registry.Registry
	// DefinitionBuilders override the built-in catalog when provided
	DefinitionBuilders []registry.Builder
	// Keystore provides credential persistence and installation-scoped client pooling
	Keystore *keystore.Store
	// RedisClient provides the shared Redis client used for ephemeral integration auth state
	RedisClient *redis.Client
	// CatalogConfig supplies operator-level credentials for all built-in definitions
	CatalogConfig catalog.Config
	// FederationIssuer is the issuer URI customer identity providers federate against
	FederationIssuer string
	// DevMode routes supporting integrations to local file-based senders instead of provider APIs
	DevMode bool
	// DefaultLookback sets the fetch-back window when an operation has no prior successful run
	DefaultLookback time.Duration
}

// PostExecutionHook is called after HandleOperation completes with the envelope and any error
type PostExecutionHook func(ctx context.Context, envelope operations.Envelope, err error)

// Runtime bundles the integrations services behind a do injector
type Runtime struct {
	// injector holds all wired integration dependencies
	injector do.Injector
	// postExecutionHook is an optional callback invoked after each HandleOperation call
	postExecutionHook PostExecutionHook
	// defaultLookback is applied as LastRunAt when an operation has no prior successful run
	defaultLookback time.Duration
	// devMode indicates the server is running in development mode
	devMode bool
}

// SetPostExecutionHook registers a callback invoked after each HandleOperation call
func (r *Runtime) SetPostExecutionHook(hook PostExecutionHook) {
	r.postExecutionHook = hook
}

// defaultRuntime holds the process-wide integrations runtime
var defaultRuntime singleton.Value[Runtime]

// SetDefault registers the process-wide integrations runtime
func SetDefault(rt *Runtime) {
	defaultRuntime.Set(rt)
}

// Default returns the process-wide integrations runtime, or nil when none is registered
func Default() *Runtime {
	return defaultRuntime.Get()
}

// DB returns the Ent client from the injector
func (r *Runtime) DB() *ent.Client {
	return do.MustInvoke[*ent.Client](r.injector)
}

// keystore returns the credential store from the injector
func (r *Runtime) keystore() *keystore.Store {
	return do.MustInvoke[*keystore.Store](r.injector)
}

// Gala returns the event runtime from the injector
func (r *Runtime) Gala() *gala.Gala {
	return do.MustInvoke[*gala.Gala](r.injector)
}

// redisClient returns the shared Redis client from the injector, nil when Redis isn't configured
func (r *Runtime) redisClient() *redis.Client {
	return do.MustInvoke[*redis.Client](r.injector)
}

// keymaker returns the auth flow service from the injector
func (r *Runtime) keymaker() *keymaker.Service {
	return do.MustInvoke[*keymaker.Service](r.injector)
}

// Registry returns the definition registry
func (r *Runtime) Registry() *registry.Registry {
	return do.MustInvoke[*registry.Registry](r.injector)
}

// NewForTesting constructs a Runtime backed by the supplied registry and a stub DB client
func NewForTesting(reg *registry.Registry) *Runtime {
	injector := do.New()
	do.ProvideValue(injector, reg)
	do.ProvideValue(injector, &ent.Client{})
	do.ProvideValue(injector, (*redis.Client)(nil))
	do.ProvideValue(injector, (*gala.Gala)(nil))

	return &Runtime{
		injector:        injector,
		defaultLookback: defaultLookbackDuration,
	}
}

// New wires the integrations runtime
func New(config Config) (*Runtime, error) {
	lookback := config.DefaultLookback
	if lookback <= 0 {
		lookback = defaultLookbackDuration
	}

	injector := do.New()
	rt := &Runtime{
		injector:        injector,
		defaultLookback: lookback,
		devMode:         config.DevMode,
	}

	do.ProvideValue(injector, config.DB)
	do.ProvideValue(injector, config.Gala)
	do.ProvideValue(injector, config.Keystore)
	do.ProvideValue(injector, config.RedisClient)

	do.Provide(injector, func(do.Injector) (keymaker.AuthStateStore, error) {
		if config.RedisClient != nil {
			return keymaker.NewRedisAuthStateStore(config.RedisClient), nil
		}

		return keymaker.NewInMemoryAuthStateStore(), nil
	})

	do.Provide(injector, func(do.Injector) (*registry.Registry, error) {
		registryInstance := config.Registry
		builders := config.DefinitionBuilders

		switch {
		case registryInstance != nil:
		case len(builders) > 0:
			registryInstance = registry.New()
		default:
			registryInstance = registry.New(registry.WithSnapshots(registry.Surfaces))
			builders = catalog.Builders(config.CatalogConfig, config.FederationIssuer, config.DevMode)
		}

		if len(builders) > 0 {
			if err := registryInstance.RegisterAll(builders...); err != nil {
				return nil, err
			}
		}

		return registryInstance, nil
	})
	do.Provide(injector, func(i do.Injector) (*keymaker.Service, error) {
		lookupDefinition := func(id string) (types.Definition, bool) { return rt.Registry().Definition(id) }

		return keymaker.NewService(lookupDefinition, func(ctx context.Context, integrationID string, credentialRef types.CredentialSlotID, def types.Definition, result types.AuthCompleteResult) error {
			installation, err := rt.ResolveIntegration(ctx, IntegrationLookup{IntegrationID: integrationID, DefinitionID: def.ID})
			if err != nil {
				return err
			}

			connection, err := def.ConnectionRegistration(credentialRef)
			if err != nil {
				return err
			}

			if err := rt.ReconcileCredential(ctx, installation, connection.Auth.CredentialRef, result.Credential, result.InstallationInput); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("installation_id", installation.ID).Msg("failed to reconcile completed auth credential")

				return err
			}

			return nil
		}, rt.lookupKeymakerInstallation, do.MustInvoke[keymaker.AuthStateStore](i)), nil
	})

	if _, err := do.Invoke[*registry.Registry](injector); err != nil {
		return nil, err
	}

	if _, err := do.Invoke[*keymaker.Service](injector); err != nil {
		return nil, err
	}

	if err := rt.registerListeners(); err != nil {
		return nil, err
	}

	return rt, nil
}
