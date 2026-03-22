package runtime

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
	"github.com/samber/lo"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/definitions/catalog"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/pkg/gala"
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
	DefinitionBuilders []definition.Builder
	// Keystore provides credential persistence and installation-scoped client pooling
	Keystore *keystore.Store
	// RedisClient provides the shared Redis client used for ephemeral integration auth state
	RedisClient *redis.Client
	// CatalogConfig supplies operator-level credentials for all built-in definitions
	CatalogConfig catalog.Config
	// SkipExecutorListeners disables automatic Gala listener registration for the executor
	SkipExecutorListeners bool
}

// Runtime bundles the integrations services behind a do injector
type Runtime struct {
	// injector holds all wired integration dependencies
	injector do.Injector
}

// DB returns the Ent client from the injector
func (r *Runtime) DB() *ent.Client {
	return do.MustInvoke[*ent.Client](r.injector)
}

// Keystore returns the credential store from the injector
func (r *Runtime) Keystore() *keystore.Store {
	return do.MustInvoke[*keystore.Store](r.injector)
}

// Gala returns the event runtime from the injector
func (r *Runtime) Gala() *gala.Gala {
	return do.MustInvoke[*gala.Gala](r.injector)
}

// Keymaker returns the auth flow service from the injector
func (r *Runtime) Keymaker() *keymaker.Service {
	return do.MustInvoke[*keymaker.Service](r.injector)
}

// Registry returns the definition registry
func (r *Runtime) Registry() *registry.Registry {
	return do.MustInvoke[*registry.Registry](r.injector)
}

// Catalog returns all registered definition specs in stable id order
func (r *Runtime) Catalog() []types.DefinitionSpec {
	return r.Registry().Catalog()
}

// Definition returns one definition by canonical identifier
func (r *Runtime) Definition(id string) (types.Definition, bool) {
	return r.Registry().Definition(id)
}

// Dispatch enqueues one integration operation through the runtime-managed dispatcher
func (r *Runtime) Dispatch(ctx context.Context, req operations.DispatchRequest) (operations.DispatchResult, error) {
	result, err := operations.Dispatch(ctx, r.Registry(), r.DB(), r.Gala(), req)
	if err != nil {
		return operations.DispatchResult{}, normalizeDispatchError(err)
	}

	return result, err
}

func normalizeDispatchError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, registry.ErrDefinitionNotFound):
		return ErrDefinitionNotFound
	case errors.Is(err, registry.ErrOperationNotFound):
		return ErrOperationNotFound
	case errors.Is(err, operations.ErrDispatchInputInvalid):
		return operations.ErrDispatchInputInvalid
	default:
		return err
	}
}

// NewForTesting constructs a Runtime backed only by the supplied registry.
// Use only in unit tests that exercise registry lookup without executing operations
func NewForTesting(reg *registry.Registry) *Runtime {
	injector := do.New()
	do.ProvideValue(injector, reg)

	return &Runtime{
		injector: injector,
	}
}

// New wires the integrations runtime
func New(config Config) (*Runtime, error) {
	injector := do.New()
	rt := &Runtime{
		injector: injector,
	}

	do.ProvideValue(injector, config.DB)
	do.ProvideValue(injector, config.Gala)
	do.ProvideValue(injector, config.Keystore)
	do.Provide(injector, func(do.Injector) (keymaker.AuthStateStore, error) {
		if config.RedisClient != nil {
			return keymaker.NewRedisAuthStateStore(config.RedisClient), nil
		}

		return keymaker.NewInMemoryAuthStateStore(), nil
	})

	do.Provide(injector, func(do.Injector) (*registry.Registry, error) {
		registryInstance := config.Registry
		if registryInstance == nil {
			registryInstance = registry.New()
		}

		builders := config.DefinitionBuilders
		if len(builders) == 0 && config.Registry == nil {
			builders = catalog.Builders(config.CatalogConfig)
		}

		if len(builders) > 0 {
			if err := definition.RegisterAll(registryInstance, builders...); err != nil {
				return nil, err
			}
		}

		return registryInstance, nil
	})
	do.Provide(injector, func(i do.Injector) (*keymaker.Service, error) {
		return keymaker.NewService(
			rt.Definition,
			func(ctx context.Context, installationID string, credentialRef types.CredentialRef, def types.Definition, result types.AuthCompleteResult) error {
				installation, err := rt.ResolveInstallation(ctx, "", installationID, def.ID)
				if err != nil {
					return err
				}

				if connection, err := def.ConnectionRegistration(credentialRef); err != nil {
					return err
				} else {
					return rt.Reconcile(ctx, installation, nil, connection.Auth.CredentialRef, &result.Credential, result.InstallationInput)
				}
			},
			rt.lookupKeymakerInstallation,
			do.MustInvoke[keymaker.AuthStateStore](i),
		), nil
	})

	if _, err := do.Invoke[*registry.Registry](injector); err != nil {
		return nil, err
	}

	if _, err := do.Invoke[*keymaker.Service](injector); err != nil {
		return nil, err
	}

	if err := operations.RegisterRuntimeListeners(
		rt.Gala(),
		rt.Registry(),
		lo.Ternary(!config.SkipExecutorListeners, rt.HandleOperation, nil),
		rt.HandleWebhookEvent,
	); err != nil {
		return nil, err
	}

	return rt, nil
}
