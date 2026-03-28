package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

// ReconcileOperations emits the first reconciliation envelope for an installation,
// starting the adaptive scheduling cycle that dispatches all non-inline operations
func (r *Runtime) ReconcileOperations(ctx context.Context, installation *ent.Integration) error {
	receipt := r.Gala().EmitWithHeaders(ctx, types.TopicFromType[operations.ReconcileEnvelope](), operations.ReconcileEnvelope{
		InstallationID: installation.ID,
		DefinitionID:   installation.DefinitionID,
	}, gala.Headers{
		Properties: map[string]string{
			"installation_id": installation.ID,
			"definition_id":   installation.DefinitionID,
		},
	})

	return receipt.Err
}

// HandleReconcile processes one reconciliation cycle by dispatching all non-inline
// operations for the installation and returning the count dispatched as the delta
func (r *Runtime) HandleReconcile(ctx context.Context, envelope operations.ReconcileEnvelope) (int, error) {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	def, ok := r.Registry().Definition(envelope.DefinitionID)
	if !ok {
		return 0, ErrDefinitionNotFound
	}

	var (
		dispatched int
		errs       []error
	)

	for _, op := range def.Operations {
		if op.Policy.Inline {
			continue
		}

		_, err := r.Dispatch(ctx, operations.DispatchRequest{
			InstallationID: envelope.InstallationID,
			Operation:      op.Name,
			RunType:        enums.IntegrationRunTypeReconcile,
		})
		if err != nil {
			errs = append(errs, err)
			continue
		}

		dispatched++
	}

	return dispatched, errors.Join(errs...)
}

// ExecuteOperation runs one integration operation inline without run tracking
func (r *Runtime) ExecuteOperation(ctx context.Context, installation *ent.Integration, operation types.OperationRegistration, credentials types.CredentialBindings, config json.RawMessage) (json.RawMessage, error) {
	if installation == nil {
		return nil, ErrInstallationRequired
	}

	if len(config) > 0 {
		if err := validatePayload(operation.ConfigSchema, config, ErrOperationConfigInvalid); err != nil {
			return nil, err
		}
	}

	return r.executeResolvedOperation(ctx, installation, operation, credentials, config, false, operations.IngestOptions{
		Source: integrationgenerated.IntegrationIngestSourceOperation,
	})
}

// HandleOperation executes one queued operation envelope through the runtime-managed dependencies
func (r *Runtime) HandleOperation(ctx context.Context, envelope operations.Envelope) error {
	// Operation handlers are system-level; the privacy bypass set by the original
	// dispatcher is not captured by context codecs in durable dispatch mode
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	startedAt := time.Now()
	db := r.DB()

	failRun := func(execErr error, response json.RawMessage) error {
		result := operations.RunResult{
			Status: enums.IntegrationRunStatusFailed,
			Error:  execErr.Error(),
		}
		if len(response) > 0 {
			result.Metrics = map[string]any{
				"response": jsonx.DecodeAnyOrNil(response),
			}
		}
		if completeErr := operations.CompleteRun(ctx, db, envelope.RunID, startedAt, result); completeErr != nil {
			return errors.Join(execErr, completeErr)
		}

		return execErr
	}

	if err := operations.MarkRunRunning(ctx, db, envelope.RunID); err != nil {
		return err
	}

	ingestOptions := operations.IngestOptions{
		Source: integrationgenerated.IntegrationIngestSourceOperation,
		RunID:  envelope.RunID,
	}

	if envelope.WorkflowMeta != nil {
		ingestOptions.Source = integrationgenerated.IntegrationIngestSourceWorkflow
		ingestOptions.WorkflowMeta = envelope.WorkflowMeta
	}

	installation, err := db.Integration.Get(ctx, envelope.InstallationID)
	if err != nil {
		return failRun(err, nil)
	}

	operation, err := r.Registry().Operation(installation.DefinitionID, envelope.Operation)
	if err != nil {
		return failRun(err, nil)
	}

	response, err := r.executeResolvedOperation(ctx, installation, operation, nil, envelope.Config, envelope.ForceClientRebuild, ingestOptions)
	if err != nil {
		return failRun(err, response)
	}

	return operations.CompleteRun(ctx, db, envelope.RunID, startedAt, operations.RunResult{
		Status:  enums.IntegrationRunStatusSuccess,
		Summary: "operation completed",
		Metrics: map[string]any{
			"response": jsonx.DecodeAnyOrNil(response),
		},
	})
}

// executeResolvedOperation executes the given operation with the input integration and registered Operation
func (r *Runtime) executeResolvedOperation(ctx context.Context, installation *ent.Integration, operation types.OperationRegistration, credentials types.CredentialBindings, config json.RawMessage, clientForce bool, ingestOptions operations.IngestOptions) (json.RawMessage, error) {
	var client any

	if operation.ClientRef.Valid() {
		registration, err := r.Registry().Client(installation.DefinitionID, operation.ClientRef)
		if err != nil {
			return nil, err
		}

		if credentials == nil {
			credentials, err = r.LoadCredentials(ctx, installation, registration.CredentialRefs)
			if err != nil {
				return nil, err
			}
		}

		client, err = r.Keystore().BuildClient(ctx, installation, registration, credentials, config, clientForce)
		if err != nil {
			return nil, err
		}
	}

	req := types.OperationRequest{
		Integration: installation,
		Credentials: credentials,
		Client:      client,
		Config:      jsonx.CloneRawMessage(config),
	}

	if operation.IngestHandle != nil {
		payloadSets, err := operation.IngestHandle(ctx, req)
		if err != nil {
			return nil, err
		}

		if err := operations.EmitPayloadSets(ctx, operations.IngestContext{
			Registry:     r.Registry(),
			DB:           r.DB(),
			Runtime:      r.Gala(),
			Installation: installation,
		}, operation.Name, operation.Ingest, payloadSets, ingestOptions); err != nil {
			return nil, err
		}

		return nil, nil
	}

	response, err := operation.Handle(ctx, req)
	if err != nil {
		return response, err
	}

	return response, nil
}
