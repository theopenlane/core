package runtime

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/definitions/catalog"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/webhooks"
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
	// RedisClient provides the shared Redis client used for ephemeral integration auth state.
	RedisClient *redis.Client
	// CatalogConfig supplies operator-level credentials for all built-in definitions
	CatalogConfig catalog.Config
	// SkipExecutorListeners disables automatic Gala listener registration for the executor.
	// Set this when a workflow engine will register its own listeners that wrap the executor,
	// to prevent double execution of operations on the same topics.
	SkipExecutorListeners bool
}

// Runtime bundles the integrations services behind a do injector
type Runtime struct {
	injector do.Injector
}

// Injector returns the underlying dependency injector
func (r *Runtime) Injector() do.Injector {
	return r.injector
}

// Registry returns the definition registry
func (r *Runtime) Registry() *registry.Registry {
	return do.MustInvoke[*registry.Registry](r.injector)
}

// Dispatch enqueues one integration operation through the runtime-managed dispatcher.
func (r *Runtime) Dispatch(ctx context.Context, req operations.DispatchRequest) (operations.DispatchResult, error) {
	result, err := operations.Dispatch(
		ctx,
		do.MustInvoke[*registry.Registry](r.injector),
		do.MustInvoke[*ent.Client](r.injector),
		do.MustInvoke[*gala.Gala](r.injector),
		req,
	)
	if err != nil {
		switch {
		case errors.Is(err, registry.ErrDefinitionNotFound):
			return operations.DispatchResult{}, ErrDefinitionNotFound
		case errors.Is(err, registry.ErrOperationNotFound):
			return operations.DispatchResult{}, operations.ErrDispatchInputInvalid
		case errors.Is(err, operations.ErrDispatchInputInvalid):
			return operations.DispatchResult{}, operations.ErrDispatchInputInvalid
		}
	}

	return result, err
}

// NewForTesting constructs a Runtime backed only by the supplied registry.
// Calling methods that require DB, Gala, or Keystore will panic.
// Use only in unit tests that exercise registry lookup without executing operations.
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
			do.MustInvoke[*registry.Registry](i).Definition,
			rt.PersistAuthCompletion,
			func(ctx context.Context, installationID string) (keymaker.InstallationRecord, error) {
				record, err := rt.ResolveInstallation(ctx, "", installationID, "")
				if err != nil {
					switch err {
					case ErrInstallationRequired, ErrInstallationIDRequired:
						return keymaker.InstallationRecord{}, keymaker.ErrInstallationIDRequired
					case ErrInstallationNotFound:
						return keymaker.InstallationRecord{}, keymaker.ErrInstallationNotFound
					default:
						return keymaker.InstallationRecord{}, err
					}
				}

				return keymaker.InstallationRecord{
					ID:           record.ID,
					OwnerID:      record.OwnerID,
					DefinitionID: record.DefinitionID,
				}, nil
			},
			do.MustInvoke[keymaker.AuthStateStore](i),
			0,
		), nil
	})

	if _, err := do.Invoke[*registry.Registry](injector); err != nil {
		return nil, err
	}

	if _, err := do.Invoke[*keymaker.Service](injector); err != nil {
		return nil, err
	}

	if !config.SkipExecutorListeners {
		if err := operations.RegisterListeners(
			do.MustInvoke[*gala.Gala](injector),
			do.MustInvoke[*registry.Registry](injector),
			func(ctx context.Context, envelope operations.Envelope) error {
				return rt.HandleOperation(ctx, envelope)
			},
		); err != nil {
			return nil, err
		}
	}

	if err := webhooks.RegisterListeners(
		do.MustInvoke[*gala.Gala](injector),
		do.MustInvoke[*registry.Registry](injector),
		func(ctx context.Context, envelope webhooks.Envelope) error {
			return rt.HandleWebhookEvent(ctx, envelope)
		},
	); err != nil {
		return nil, err
	}

	return rt, nil
}
