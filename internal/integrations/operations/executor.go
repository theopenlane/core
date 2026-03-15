package operations

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/clients"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

// Executor handles queued operation events
type Executor struct {
	registry    registry.DefinitionRegistry
	db          *ent.Client
	credentials types.CredentialResolver
	clients     *clients.Service
	runs        *RunStore
}

type listenerHandler struct {
	executor *Executor
}

// NewExecutor constructs the operation executor
func NewExecutor(reg registry.DefinitionRegistry, db *ent.Client, credentialResolver types.CredentialResolver, clientService *clients.Service, runs *RunStore) (*Executor, error) {
	if reg == nil {
		return nil, ErrRegistryRequired
	}

	if db == nil {
		return nil, ErrDBClientRequired
	}

	if credentialResolver == nil {
		return nil, ErrCredentialResolverRequired
	}

	if clientService == nil {
		return nil, ErrClientServiceRequired
	}

	if runs == nil {
		return nil, ErrRunStoreRequired
	}

	return &Executor{
		registry:    reg,
		db:          db,
		credentials: credentialResolver,
		clients:     clientService,
		runs:        runs,
	}, nil
}

// RegisterListeners attaches one Gala listener per registered operation topic
func (e *Executor) RegisterListeners(runtime *gala.Gala) error {
	if runtime == nil {
		return ErrGalaRequired
	}

	handler := listenerHandler{executor: e}

	for _, operation := range e.registry.Listeners() {
		if _, err := gala.RegisterListeners(runtime.Registry(), gala.Definition[Envelope]{
			Topic: gala.Topic[Envelope]{
				Name: operation.Topic,
			},
			Name:   fmt.Sprintf("integrationsv2.%s.%s", operation.Topic, operation.Name),
			Handle: handler.Handle,
		}); err != nil {
			return err
		}
	}

	return nil
}

// ExecuteOperation runs one operation inline without run tracking.
// Used for synchronous health checks and validation before persisting configuration.
func (e *Executor) ExecuteOperation(ctx context.Context, installation *ent.Integration, operationName types.OperationName, credential types.CredentialSet, config json.RawMessage) (json.RawMessage, error) {
	operation, err := e.registry.OperationFromString(installation.DefinitionID, string(operationName))
	if err != nil {
		return nil, err
	}

	var client any
	if operation.Client != "" {
		client, err = e.clients.Build(ctx, installation, operation.Client, credential, config, false)
		if err != nil {
			return nil, err
		}
	}

	return operation.Handle(ctx, installation, credential, client, jsonx.CloneRawMessage(config))
}

// Handle processes one queued operation event
func (e *Executor) Handle(ctx context.Context, envelope Envelope) error {
	startedAt := time.Now()

	if err := e.runs.MarkRunning(ctx, envelope.RunID); err != nil {
		return err
	}

	installationRecord, err := e.db.Integration.Get(ctx, envelope.InstallationID)
	if err != nil {
		_ = e.runs.Complete(ctx, envelope.RunID, startedAt, RunResult{
			Status: enums.IntegrationRunStatusFailed,
			Error:  err.Error(),
		})
		return err
	}

	operation, err := e.registry.OperationFromString(installationRecord.DefinitionID, envelope.Operation)
	if err != nil {
		_ = e.runs.Complete(ctx, envelope.RunID, startedAt, RunResult{
			Status: enums.IntegrationRunStatusFailed,
			Error:  err.Error(),
		})
		return err
	}

	credential, found, err := e.credentials.LoadCredential(ctx, installationRecord)
	if err != nil {
		_ = e.runs.Complete(ctx, envelope.RunID, startedAt, RunResult{
			Status: enums.IntegrationRunStatusFailed,
			Error:  err.Error(),
		})
		return err
	}
	if !found {
		credential = types.CredentialSet{}
	}

	var client any
	if operation.Client != "" {
		client, err = e.clients.Build(ctx, installationRecord, operation.Client, credential, envelope.Config, envelope.ClientForce)
		if err != nil {
			_ = e.runs.Complete(ctx, envelope.RunID, startedAt, RunResult{
				Status: enums.IntegrationRunStatusFailed,
				Error:  err.Error(),
			})
			return err
		}
	}

	response, err := operation.Handle(ctx, installationRecord, credential, client, jsonx.CloneRawMessage(envelope.Config))
	if err != nil {
		_ = e.runs.Complete(ctx, envelope.RunID, startedAt, RunResult{
			Status: enums.IntegrationRunStatusFailed,
			Error:  err.Error(),
			Metrics: map[string]any{
				"response": decodeResponse(response),
			},
		})
		return err
	}

	return e.runs.Complete(ctx, envelope.RunID, startedAt, RunResult{
		Status:  enums.IntegrationRunStatusSuccess,
		Summary: "operation completed",
		Metrics: map[string]any{
			"response": decodeResponse(response),
		},
	})
}

// decodeResponse converts one raw JSON response into a persisted metric value
func decodeResponse(value json.RawMessage) any {
	if len(value) == 0 {
		return map[string]any{}
	}

	var decoded any
	if err := json.Unmarshal(value, &decoded); err != nil {
		return string(value)
	}

	return decoded
}

// Handle processes one Gala envelope through the shared executor
func (h listenerHandler) Handle(ctx gala.HandlerContext, envelope Envelope) error {
	return h.executor.Handle(ctx.Context, envelope)
}
