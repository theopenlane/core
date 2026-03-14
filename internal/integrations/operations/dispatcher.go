package operations

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

// Dispatcher validates and enqueues operation execution requests
type Dispatcher struct {
	registry registry.DefinitionRegistry
	db       *ent.Client
	runs     *RunStore
	gala     *gala.Gala
}

// NewDispatcher constructs the operation dispatcher
func NewDispatcher(reg registry.DefinitionRegistry, db *ent.Client, runs *RunStore, runtime *gala.Gala) (*Dispatcher, error) {
	if reg == nil {
		return nil, ErrRegistryRequired
	}

	if db == nil {
		return nil, ErrDBClientRequired
	}

	if runs == nil {
		return nil, ErrRunStoreRequired
	}

	if runtime == nil {
		return nil, ErrGalaRequired
	}

	return &Dispatcher{
		registry: reg,
		db:       db,
		runs:     runs,
		gala:     runtime,
	}, nil
}

// Dispatch records the run and emits the execution event
func (d *Dispatcher) Dispatch(ctx context.Context, req DispatchRequest) (DispatchResult, error) {
	if req.InstallationID == "" {
		return DispatchResult{}, ErrInstallationIDRequired
	}

	if req.Operation == "" {
		return DispatchResult{}, ErrOperationNameRequired
	}

	installationRecord, err := d.db.Integration.Get(ctx, req.InstallationID)
	if err != nil {
		return DispatchResult{}, err
	}

	operation, err := d.registry.OperationFromString(installationRecord.DefinitionID, string(req.Operation))
	if err != nil {
		return DispatchResult{}, err
	}

	if err := validateConfig(operation.ConfigSchema, req.Config); err != nil {
		return DispatchResult{}, err
	}

	runType := req.RunType
	if runType == "" {
		runType = enums.IntegrationRunTypeManual
	}

	runRecord, err := d.runs.CreatePending(ctx, installationRecord, DispatchRequest{
		InstallationID: req.InstallationID,
		Operation:      req.Operation,
		Config:         cloneRawMessage(req.Config),
		Force:          req.Force,
		ClientForce:    req.ClientForce,
		RunType:        runType,
	})
	if err != nil {
		return DispatchResult{}, err
	}

	receipt := d.gala.EmitWithHeaders(ctx, operation.Topic, Envelope{
		RunID:          runRecord.ID,
		InstallationID: installationRecord.ID,
		DefinitionID:   installationRecord.DefinitionID,
		Operation:      string(req.Operation),
		Config:         cloneRawMessage(req.Config),
		Force:          req.Force,
		ClientForce:    req.ClientForce,
		WorkflowMeta:   req.WorkflowMeta,
	}, gala.Headers{
		IdempotencyKey: runRecord.ID,
		Properties: map[string]string{
			"installation_id": installationRecord.ID,
			"definition_id":   installationRecord.DefinitionID,
			"operation":       string(req.Operation),
		},
	})
	if receipt.Err != nil {
		return DispatchResult{}, receipt.Err
	}

	return DispatchResult{
		RunID:   runRecord.ID,
		EventID: string(receipt.EventID),
		Status:  enums.IntegrationRunStatusPending,
	}, nil
}

// validateConfig validates one raw configuration payload against the operation schema
func validateConfig(schema json.RawMessage, value json.RawMessage) error {
	if len(schema) == 0 {
		return nil
	}

	var document any = map[string]any{}
	if len(value) > 0 {
		if err := json.Unmarshal(value, &document); err != nil {
			return err
		}
	}

	result, err := jsonx.ValidateSchema(schema, document)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	}

	messages := jsonx.ValidationErrorStrings(result)
	if len(messages) == 0 {
		return nil
	}

	return errors.New(strings.Join(messages, "; "))
}

// cloneRawMessage copies one JSON payload
func cloneRawMessage(value json.RawMessage) json.RawMessage {
	if len(value) == 0 {
		return nil
	}

	return append(json.RawMessage(nil), value...)
}
