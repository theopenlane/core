package runtime

import (
	"context"

	"github.com/samber/do/v2"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrationsv2/clients"
	"github.com/theopenlane/core/internal/integrationsv2/definition"
	"github.com/theopenlane/core/internal/integrationsv2/definitions/catalog"
	"github.com/theopenlane/core/internal/integrationsv2/installation"
	"github.com/theopenlane/core/internal/integrationsv2/operations"
	"github.com/theopenlane/core/internal/integrationsv2/registry"
	"github.com/theopenlane/core/internal/integrationsv2/types"
	"github.com/theopenlane/core/pkg/gala"
)

// CatalogConfig is an alias for catalog.Config so callers do not need to import both packages
type CatalogConfig = catalog.Config

// Config defines the dependencies required to build the integrationsv2 runtime
type Config struct {
	// DB is the Ent client used by installation and run stores
	DB *ent.Client
	// Gala is the event runtime used for operation dispatch and execution
	Gala *gala.Gala
	// Registry overrides the default empty definition registry when provided
	Registry *registry.Registry
	// DefinitionBuilders override the built-in catalog when provided
	DefinitionBuilders []definition.Builder
	// CredentialResolver resolves installation-scoped credentials from an external credential service
	CredentialResolver types.CredentialResolver
	// CatalogConfig supplies operator-level credentials for all built-in definitions
	CatalogConfig catalog.Config
}

// Runtime bundles the integrationsv2 services behind a do injector
type Runtime struct {
	injector do.Injector
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

// Installations returns the installation store
func (r *Runtime) Installations() *installation.Store {
	return do.MustInvoke[*installation.Store](r.injector)
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

// New wires the integrationsv2 runtime
func New(config Config) (*Runtime, error) {
	if config.DB == nil {
		return nil, ErrDBClientRequired
	}

	if config.Gala == nil {
		return nil, ErrGalaRequired
	}

	if config.CredentialResolver == nil {
		return nil, ErrCredentialResolverRequired
	}

	injector := do.New()
	services := serviceProviders{config: config}

	do.ProvideValue(injector, config.DB)
	do.ProvideValue(injector, config.Gala)
	do.ProvideValue(injector, config.CredentialResolver)

	do.Provide(injector, services.provideRegistry)
	do.Provide(injector, services.provideInstallationStore)
	do.Provide(injector, services.provideClientService)
	do.Provide(injector, services.provideRunStore)
	do.Provide(injector, services.provideDispatcher)
	do.Provide(injector, services.provideExecutor)

	if _, err := do.Invoke[*registry.Registry](injector); err != nil {
		return nil, err
	}
	if _, err := do.Invoke[*installation.Store](injector); err != nil {
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
	if err := executor.RegisterListeners(do.MustInvoke[*gala.Gala](injector)); err != nil {
		return nil, err
	}

	return &Runtime{injector: injector}, nil
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

// provideInstallationStore builds the installation store for the runtime
func (serviceProviders) provideInstallationStore(i do.Injector) (*installation.Store, error) {
	return installation.NewStore(do.MustInvoke[*ent.Client](i))
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
		do.MustInvoke[*installation.Store](i),
		do.MustInvoke[*operations.RunStore](i),
		do.MustInvoke[*gala.Gala](i),
	)
}

// provideExecutor builds the operation executor for the runtime
func (serviceProviders) provideExecutor(i do.Injector) (*operations.Executor, error) {
	return operations.NewExecutor(
		do.MustInvoke[*registry.Registry](i),
		do.MustInvoke[*installation.Store](i),
		do.MustInvoke[types.CredentialResolver](i),
		do.MustInvoke[*clients.Service](i),
		do.MustInvoke[*operations.RunStore](i),
	)
}
