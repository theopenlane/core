package runtime

import (
	"context"

	"github.com/samber/do/v2"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/definitions/catalog"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/pkg/gala"
)

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
	// Keystore provides credential persistence and installation-scoped client pooling
	Keystore *keystore.Store
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

// Injector returns the underlying dependency injector
func (r *Runtime) Injector() do.Injector {
	return r.injector
}

// Registry returns the definition registry
func (r *Runtime) Registry() *registry.Registry {
	return do.MustInvoke[*registry.Registry](r.injector)
}

// SuccessRedirectURL returns the configured success redirect URL
func (r *Runtime) SuccessRedirectURL() string {
	return r.successRedirectURL
}

// Dispatch enqueues one integration operation through the runtime-managed dispatcher.
func (r *Runtime) Dispatch(ctx context.Context, req operations.DispatchRequest) (operations.DispatchResult, error) {
	return operations.Dispatch(
		ctx,
		do.MustInvoke[*registry.Registry](r.injector),
		do.MustInvoke[*ent.Client](r.injector),
		do.MustInvoke[*gala.Gala](r.injector),
		req,
	)
}

// NewForTesting constructs a Runtime backed only by the supplied registry.
// Calling methods that require DB, Gala, or Keystore will panic.
// Use only in unit tests that exercise registry lookup without executing operations.
func NewForTesting(reg *registry.Registry, successRedirectURL string) *Runtime {
	injector := do.New()
	do.ProvideValue(injector, reg)

	return &Runtime{
		injector:           injector,
		successRedirectURL: successRedirectURL,
	}
}

// New wires the integrationsv2 runtime
func New(config Config) (*Runtime, error) {
	injector := do.New()
	rt := &Runtime{
		injector:           injector,
		successRedirectURL: config.SuccessRedirectURL,
	}

	do.ProvideValue(injector, config.DB)
	do.ProvideValue(injector, config.Gala)
	do.ProvideValue(injector, config.Keystore)

	do.Provide(injector, func(do.Injector) (*registry.Registry, error) {
		registryInstance := config.Registry
		if registryInstance == nil {
			registryInstance = registry.New()
		}

		builders := config.DefinitionBuilders
		if len(builders) == 0 {
			builders = catalog.Builders(config.CatalogConfig)
		}

		if err := definition.RegisterAll(context.Background(), registryInstance, builders...); err != nil {
			return nil, err
		}

		return registryInstance, nil
	})
	do.Provide(injector, func(i do.Injector) (*keymaker.Service, error) {
		return keymaker.NewService(
			do.MustInvoke[*registry.Registry](i).Definition,
			config.Keystore.SaveInstallationCredential,
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
			config.AuthStateStore,
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

	return rt, nil
}
