package runtime

import (
	"context"

	"github.com/samber/do/v2"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/clients"
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/definitions/catalog"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/pkg/gala"
)

// CatalogConfig is an alias for catalog.Config so callers do not need to import both packages
type CatalogConfig = catalog.Config

// Config defines the dependencies required to build the integrationsv2 runtime
type Config struct {
	// DB is the Ent client used by run stores and direct installation queries
	DB *ent.Client
	// Gala is the event runtime used for operation dispatch and execution
	Gala *gala.Gala
	// Registry overrides the default empty definition registry when provided
	Registry *registry.Registry
	// DefinitionBuilders override the built-in catalog when provided
	DefinitionBuilders []definition.Builder
	// CredentialStore provides both credential read and write access for installations
	CredentialStore types.CredentialStore
	// AuthStateStore persists pending OAuth state across auth start and callback completion.
	AuthStateStore keymaker.AuthStateStore
	// CatalogConfig supplies operator-level credentials for all built-in definitions
	CatalogConfig catalog.Config
	// SuccessRedirectURL is the URL to redirect to after a successful integration auth flow
	SuccessRedirectURL string
	// SkipExecutorListeners disables automatic Gala listener registration for the executor.
	// Set this when a workflow engine will register its own listeners that wrap the executor,
	// to prevent double execution of operations on the same topics.
	SkipExecutorListeners bool
}

// Runtime bundles the integrationsv2 services behind a do injector
type Runtime struct {
	injector           do.Injector
	successRedirectURL string
}

type serviceProviders struct {
	config Config
}

// Injector returns the underlying dependency injector
func (r *Runtime) Injector() do.Injector {
	return r.injector
}

// Registry returns the definition registry
func (r *Runtime) Registry() *registry.Registry {
	return do.MustInvoke[*registry.Registry](r.injector)
}

// Clients returns the client service
func (r *Runtime) Clients() *clients.Service {
	return do.MustInvoke[*clients.Service](r.injector)
}

// Runs returns the operation run store
func (r *Runtime) Runs() *operations.RunStore {
	return do.MustInvoke[*operations.RunStore](r.injector)
}

// Dispatcher returns the operation dispatcher
func (r *Runtime) Dispatcher() *operations.Dispatcher {
	return do.MustInvoke[*operations.Dispatcher](r.injector)
}

// Executor returns the operation executor
func (r *Runtime) Executor() *operations.Executor {
	return do.MustInvoke[*operations.Executor](r.injector)
}

// CredentialStore returns the credential store
func (r *Runtime) CredentialStore() types.CredentialStore {
	return do.MustInvoke[types.CredentialStore](r.injector)
}

// Keymaker returns the keymaker service
func (r *Runtime) Keymaker() *keymaker.Service {
	return do.MustInvoke[*keymaker.Service](r.injector)
}

// SuccessRedirectURL returns the configured success redirect URL
func (r *Runtime) SuccessRedirectURL() string {
	return r.successRedirectURL
}

// NewForTesting constructs a Runtime backed only by the supplied registry.
// Calling methods that require DB, Gala, or CredentialStore will panic.
// Use only in unit tests that exercise registry lookup without executing operations.
func NewForTesting(reg *registry.Registry, successRedirectURL string) *Runtime {
	injector := do.New()
	do.ProvideValue(injector, reg)

	return &Runtime{injector: injector, successRedirectURL: successRedirectURL}
}

// New wires the integrationsv2 runtime
func New(config Config) (*Runtime, error) {
	if config.DB == nil {
		return nil, ErrDBClientRequired
	}

	if config.Gala == nil {
		return nil, ErrGalaRequired
	}

	if config.CredentialStore == nil {
		return nil, ErrCredentialStoreRequired
	}

	if config.AuthStateStore == nil {
		return nil, ErrAuthStateStoreRequired
	}

	injector := do.New()
	services := serviceProviders{config: config}

	do.ProvideValue(injector, config.DB)
	do.ProvideValue(injector, config.Gala)
	// Provide CredentialStore, CredentialResolver, and keymaker.CredentialWriter from the same instance
	do.ProvideValue(injector, config.CredentialStore)
	do.Provide(injector, func(do.Injector) (types.CredentialResolver, error) {
		return config.CredentialStore, nil
	})
	do.Provide(injector, func(do.Injector) (keymaker.CredentialWriter, error) {
		return config.CredentialStore, nil
	})

	do.Provide(injector, services.provideRegistry)
	do.Provide(injector, services.provideClientService)
	do.Provide(injector, services.provideRunStore)
	do.Provide(injector, services.provideDispatcher)
	do.Provide(injector, services.provideExecutor)
	do.Provide(injector, services.provideKeymaker)

	if _, err := do.Invoke[*registry.Registry](injector); err != nil {
		return nil, err
	}
	if _, err := do.Invoke[*clients.Service](injector); err != nil {
		return nil, err
	}
	if _, err := do.Invoke[*operations.RunStore](injector); err != nil {
		return nil, err
	}
	if _, err := do.Invoke[*operations.Dispatcher](injector); err != nil {
		return nil, err
	}

	executor, err := do.Invoke[*operations.Executor](injector)
	if err != nil {
		return nil, err
	}

	if _, err := do.Invoke[*keymaker.Service](injector); err != nil {
		return nil, err
	}

	if !config.SkipExecutorListeners {
		if err := executor.RegisterListeners(do.MustInvoke[*gala.Gala](injector)); err != nil {
			return nil, err
		}
	}

	return &Runtime{injector: injector, successRedirectURL: config.SuccessRedirectURL}, nil
}

// provideRegistry builds the definition registry for the runtime
func (p serviceProviders) provideRegistry(do.Injector) (*registry.Registry, error) {
	registryInstance := p.config.Registry
	if registryInstance == nil {
		registryInstance = registry.New()
	}

	builders := p.config.DefinitionBuilders
	if len(builders) == 0 {
		builders = catalog.Builders(p.config.CatalogConfig)
	}

	if err := definition.RegisterAll(context.Background(), registryInstance, builders...); err != nil {
		return nil, err
	}

	return registryInstance, nil
}

// provideClientService builds the client service for the runtime
func (serviceProviders) provideClientService(i do.Injector) (*clients.Service, error) {
	return clients.NewService(
		do.MustInvoke[*registry.Registry](i),
		do.MustInvoke[types.CredentialResolver](i),
	)
}

// provideRunStore builds the run store for the runtime
func (serviceProviders) provideRunStore(i do.Injector) (*operations.RunStore, error) {
	return operations.NewRunStore(do.MustInvoke[*ent.Client](i))
}

// provideDispatcher builds the operation dispatcher for the runtime
func (serviceProviders) provideDispatcher(i do.Injector) (*operations.Dispatcher, error) {
	return operations.NewDispatcher(
		do.MustInvoke[*registry.Registry](i),
		do.MustInvoke[*ent.Client](i),
		do.MustInvoke[*operations.RunStore](i),
		do.MustInvoke[*gala.Gala](i),
	)
}

// provideExecutor builds the operation executor for the runtime
func (serviceProviders) provideExecutor(i do.Injector) (*operations.Executor, error) {
	return operations.NewExecutor(
		do.MustInvoke[*registry.Registry](i),
		do.MustInvoke[*ent.Client](i),
		do.MustInvoke[types.CredentialResolver](i),
		do.MustInvoke[*clients.Service](i),
		do.MustInvoke[*operations.RunStore](i),
	)
}

// provideKeymaker builds the keymaker auth flow service for the runtime
func (p serviceProviders) provideKeymaker(i do.Injector) (*keymaker.Service, error) {
	return keymaker.NewService(
		do.MustInvoke[*registry.Registry](i),
		do.MustInvoke[keymaker.CredentialWriter](i),
		entInstallationResolver{db: do.MustInvoke[*ent.Client](i)},
		p.config.AuthStateStore,
		keymaker.Options{},
	)
}
