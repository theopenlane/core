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

// Dispatch validates and enqueues one operation execution request
func Dispatch(ctx context.Context, reg *registry.Registry, db *ent.Client, runtime *gala.Gala, req DispatchRequest) (DispatchResult, error) {
	if req.InstallationID == "" {
		return DispatchResult{}, ErrInstallationIDRequired
	}

	if req.Operation == "" {
		return DispatchResult{}, ErrOperationNameRequired
	}

	installationRecord, err := db.Integration.Get(ctx, req.InstallationID)
	if err != nil {
		return DispatchResult{}, err
	}

	operation, err := reg.Operation(installationRecord.DefinitionID, req.Operation)
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

	runRecord, err := CreatePendingRun(ctx, db, installationRecord, DispatchRequest{
		InstallationID: req.InstallationID,
		Operation:      req.Operation,
		Config:         jsonx.CloneRawMessage(req.Config),
		Force:          req.Force,
		ClientForce:    req.ClientForce,
		RunType:        runType,
	})
	if err != nil {
		return DispatchResult{}, err
	}

	receipt := runtime.EmitWithHeaders(ctx, operation.Topic, Envelope{
		RunID:          runRecord.ID,
		InstallationID: installationRecord.ID,
		DefinitionID:   installationRecord.DefinitionID,
		Operation:      req.Operation,
		Config:         jsonx.CloneRawMessage(req.Config),
		Force:          req.Force,
		ClientForce:    req.ClientForce,
		WorkflowMeta:   req.WorkflowMeta,
	}, gala.Headers{
		IdempotencyKey: runRecord.ID,
		Properties: map[string]string{
			"installation_id": installationRecord.ID,
			"definition_id":   installationRecord.DefinitionID,
			"operation":       req.Operation,
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
	if err := jsonx.UnmarshalIfPresent(value, &document); err != nil {
		return err
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
