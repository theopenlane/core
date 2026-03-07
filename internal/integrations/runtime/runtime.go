package runtime

import (
	"context"

	"github.com/samber/do/v2"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/activation"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/internal/keystore"
)

// Runtime holds the fully wired integrations runtime via a dependency injector
// Use typed accessor methods for common components, or Injector() for extensibility
type Runtime struct {
	injector do.Injector
}

// Injector returns the underlying dependency injector for external service registration
// and invocation. Callers may register additional services or invoke registered ones
func (r *Runtime) Injector() do.Injector {
	return r.injector
}

// Registry returns the provider registry
func (r *Runtime) Registry() *registry.Registry {
	return do.MustInvoke[*registry.Registry](r.injector)
}

// Store returns the credential and integration record store
func (r *Runtime) Store() *keystore.Store {
	return do.MustInvoke[*keystore.Store](r.injector)
}

// Broker returns the credential minting coordinator
func (r *Runtime) Broker() *keystore.Broker {
	return do.MustInvoke[*keystore.Broker](r.injector)
}

// Clients returns the integration client pool manager
func (r *Runtime) Clients() *keystore.ClientPoolManager {
	return do.MustInvoke[*keystore.ClientPoolManager](r.injector)
}

// Operations returns the integration operation manager
func (r *Runtime) Operations() *keystore.OperationManager {
	return do.MustInvoke[*keystore.OperationManager](r.injector)
}

// Keymaker returns the OAuth keymaker service
func (r *Runtime) Keymaker() *keymaker.Service {
	return do.MustInvoke[*keymaker.Service](r.injector)
}

// Activation returns the integration activation service
func (r *Runtime) Activation() *activation.Service {
	return do.MustInvoke[*activation.Service](r.injector)
}

// SuccessRedirectURL returns the global fallback redirect URL used after successful provider authentication
func (r *Runtime) SuccessRedirectURL() string {
	return do.MustInvoke[successRedirectURL](r.injector).value
}

// successRedirectURL is an unexported wrapper so the string value can be registered in the DI container
type successRedirectURL struct{ value string }

// New constructs the integrations runtime, building the provider registry from
// ProviderSpecs and wiring all dependent components via the injector
func New(cfg Config) (*Runtime, error) {
	if cfg.DB == nil {
		return nil, ErrDBClientRequired
	}

	i := do.New()

	do.ProvideValue(i, cfg.DB)
	do.ProvideValue(i, successRedirectURL{value: cfg.SuccessRedirectURL})

	do.Provide(i, func(_ do.Injector) (*registry.Registry, error) {
		if cfg.Registry != nil {
			return cfg.Registry, nil
		}

		return registry.NewRegistry(context.Background(), cfg.ProviderSpecs)
	})

	do.Provide(i, func(i do.Injector) (*keystore.Store, error) {
		db := do.MustInvoke[*ent.Client](i)

		return keystore.NewStore(db)
	})

	do.Provide(i, func(i do.Injector) (*keystore.Broker, error) {
		store := do.MustInvoke[*keystore.Store](i)
		reg := do.MustInvoke[*registry.Registry](i)

		return keystore.NewBroker(store, reg)
	})

	do.Provide(i, func(i do.Injector) (*keystore.ClientPoolManager, error) {
		broker := do.MustInvoke[*keystore.Broker](i)
		reg := do.MustInvoke[*registry.Registry](i)

		return keystore.NewClientPoolManager(broker, keystore.FlattenDescriptors(reg.ClientDescriptorCatalog()))
	})

	do.Provide(i, func(i do.Injector) (*keystore.OperationManager, error) {
		broker := do.MustInvoke[*keystore.Broker](i)
		clients := do.MustInvoke[*keystore.ClientPoolManager](i)
		reg := do.MustInvoke[*registry.Registry](i)

		return keystore.NewOperationManager(
			broker,
			keystore.FlattenOperationDescriptors(reg.OperationDescriptorCatalog()),
			keystore.WithOperationClients(clients),
		)
	})

	do.Provide(i, func(i do.Injector) (*keymaker.Service, error) {
		reg := do.MustInvoke[*registry.Registry](i)
		store := do.MustInvoke[*keystore.Store](i)
		authStates := cfg.AuthStateStore
		if authStates == nil {
			authStates = keymaker.NewInMemoryAuthStateStore()
		}

		return keymaker.NewService(reg, store, authStates, keymaker.ServiceOptions{})
	})

	do.Provide(i, func(i do.Injector) (*activation.Service, error) {
		store := do.MustInvoke[*keystore.Store](i)
		ops := do.MustInvoke[*keystore.OperationManager](i)
		reg := do.MustInvoke[*registry.Registry](i)

		return activation.NewService(store, ops, reg)
	})

	do.Provide(i, func(i do.Injector) (types.MappingIndex, error) {

		return do.MustInvoke[*registry.Registry](i), nil
	})

	// Eagerly invoke all services so initialization errors surface at startup
	if _, err := do.Invoke[*registry.Registry](i); err != nil {
		return nil, err
	}
	if _, err := do.Invoke[*keystore.Store](i); err != nil {
		return nil, err
	}
	if _, err := do.Invoke[*keystore.Broker](i); err != nil {
		return nil, err
	}
	if _, err := do.Invoke[*keystore.ClientPoolManager](i); err != nil {
		return nil, err
	}
	if _, err := do.Invoke[*keystore.OperationManager](i); err != nil {
		return nil, err
	}
	if _, err := do.Invoke[*keymaker.Service](i); err != nil {
		return nil, err
	}
	if _, err := do.Invoke[*activation.Service](i); err != nil {
		return nil, err
	}
	if _, err := do.Invoke[types.MappingIndex](i); err != nil {
		return nil, err
	}

	return &Runtime{injector: i}, nil
}

// NewFromRegistry builds a minimal runtime with only a pre-built registry registered
func NewFromRegistry(reg *registry.Registry) *Runtime {
	i := do.New()

	do.Provide(i, func(_ do.Injector) (*registry.Registry, error) {
		return reg, nil
	})

	do.ProvideValue(i, successRedirectURL{})

	return &Runtime{injector: i}
}
