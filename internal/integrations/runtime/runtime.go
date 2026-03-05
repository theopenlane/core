package runtime

import (
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/activation"
	"github.com/theopenlane/core/internal/integrations/ingest"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/internal/workflows/engine"
)

// Config carries required and optional dependencies for constructing the integrations runtime.
type Config struct {
	// Registry provides provider descriptors and minting implementations.
	Registry *registry.Registry
	// DB provides persistence for credentials and integration records.
	DB *ent.Client
	// WorkflowEngine receives integration deps for queued operation execution.
	WorkflowEngine *engine.WorkflowEngine
}

// Runtime contains the fully wired integrations runtime components.
type Runtime struct {
	Registry   *registry.Registry
	Store      *keystore.Store
	Broker     *keystore.Broker
	Clients    *keystore.ClientPoolManager
	Operations *keystore.OperationManager
	Keymaker   *keymaker.Service
	Activation *activation.Service
	Ingest     *ingest.Runtime
}

// New constructs the integrations runtime in dependency order and wires workflow deps when provided.
func New(cfg Config) (*Runtime, error) {
	if cfg.Registry == nil {
		return nil, ErrRegistryRequired
	}
	if cfg.DB == nil {
		return nil, ErrDBClientRequired
	}

	store := keystore.NewStore(cfg.DB)
	broker := keystore.NewBroker(store, cfg.Registry)

	clients, err := keystore.NewClientPoolManager(broker, keystore.FlattenDescriptors(cfg.Registry.ClientDescriptorCatalog()))
	if err != nil {
		return nil, err
	}

	operations, err := keystore.NewOperationManager(
		broker,
		keystore.FlattenOperationDescriptors(cfg.Registry.OperationDescriptorCatalog()),
		keystore.WithOperationClients(clients),
	)
	if err != nil {
		return nil, err
	}

	sessions := keymaker.NewMemorySessionStore()
	keymakerSvc, err := keymaker.NewService(cfg.Registry, store, sessions, keymaker.ServiceOptions{})
	if err != nil {
		return nil, err
	}

	activationSvc, err := activation.NewService(store, operations, cfg.Registry)
	if err != nil {
		return nil, err
	}

	ingestRuntime := ingest.NewRuntime(cfg.Registry)

	if cfg.WorkflowEngine != nil {
		if err := cfg.WorkflowEngine.SetIntegrationDeps(engine.IntegrationDeps{
			Registry:     cfg.Registry,
			Store:        store,
			Operations:   operations,
			MappingIndex: ingestRuntime.MappingIndex(),
		}); err != nil {
			return nil, err
		}
	}

	return &Runtime{
		Registry:   cfg.Registry,
		Store:      store,
		Broker:     broker,
		Clients:    clients,
		Operations: operations,
		Keymaker:   keymakerSvc,
		Activation: activationSvc,
		Ingest:     ingestRuntime,
	}, nil
}
