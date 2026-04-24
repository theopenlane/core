package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/riverqueue/river"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// bootstrapHandlerContext resolves the integration and returns a fully prepared
// system-level context. When metadata.Runtime is true, no DB lookup is performed
// and the client is resolved from the registry at execution time
func (r *Runtime) bootstrapHandlerContext(ctx context.Context, metadata types.ExecutionMetadata) (context.Context, *ent.Integration, types.ExecutionMetadata, error) {
	if metadata.Runtime {
		return ctx, nil, metadata, nil
	}

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	integration, err := r.DB().Integration.Get(systemCtx, metadata.IntegrationID)
	if err != nil {
		return ctx, nil, metadata, err
	}

	metadata.OwnerID = integration.OwnerID
	metadata.IntegrationID = integration.ID
	metadata.DefinitionID = integration.DefinitionID

	systemCtx = auth.WithCaller(systemCtx, auth.NewWebhookCaller(integration.OwnerID))
	systemCtx = types.WithExecutionMetadata(systemCtx, metadata)

	return systemCtx, integration, metadata, nil
}

// reconcileOperations emits one reconciliation envelope per reconcilable operation,
// starting an independent adaptive scheduling cycle for each
func (r *Runtime) reconcileOperations(ctx context.Context, integration *ent.Integration) error {
	def, ok := r.Registry().Definition(integration.DefinitionID)
	if !ok {
		return ErrDefinitionNotFound
	}

	var errs []error

	for _, op := range def.Operations {
		if !op.Policy.Reconcile {
			continue
		}

		metadata := types.ExecutionMetadata{
			OwnerID:       integration.OwnerID,
			IntegrationID: integration.ID,
			DefinitionID:  integration.DefinitionID,
			Operation:     op.Name,
			RunType:       enums.IntegrationRunTypeReconcile,
		}

		receipt := r.Gala().EmitWithHeaders(types.WithExecutionMetadata(ctx, metadata), operations.ReconcileTopic, operations.ReconcileEnvelope{
			ExecutionMetadata: metadata,
		}, gala.Headers{
			Properties: metadata.Properties(),
		})

		if receipt.Err != nil {
			logx.FromContext(ctx).Error().Err(receipt.Err).Str("integration_id", integration.ID).Str("operation", op.Name).Msg("failed to emit reconcile envelope")
			errs = append(errs, receipt.Err)

			continue
		}

		logx.FromContext(ctx).Info().Str("integration_id", integration.ID).Str("operation", op.Name).Msg("reconcile envelope emitted")
	}

	return errors.Join(errs...)
}

// reconcileOutput is the structured output recorded on reconcile River jobs for UI visibility
type reconcileOutput struct {
	// IntegrationID is the target integration identifier
	IntegrationID string `json:"integration_id"`
	// DefinitionID is the integration definition identifier
	DefinitionID string `json:"definition_id"`
	// Operation is the operation that was executed
	Operation string `json:"operation"`
	// RunID is the integration run record identifier
	RunID string `json:"run_id"`
	// Records is the number of ingest records processed
	Records int `json:"records,omitempty"`
	// Status is the final run status
	Status enums.IntegrationRunStatus `json:"status"`
	// Error is the error message on failure
	Error string `json:"error,omitempty"`
	// DurationMS is the execution duration in milliseconds
	DurationMS int64 `json:"duration_ms"`
}

// HandleReconcile executes one reconcilable operation inline and returns the
// number of records processed as the delta for adaptive scheduling
func (r *Runtime) HandleReconcile(ctx context.Context, envelope operations.ReconcileEnvelope) (int, error) {
	ctx, installation, metadata, err := r.bootstrapHandlerContext(ctx, envelope.ExecutionMetadata)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("integration_id", envelope.IntegrationID).Str("operation", envelope.Operation).Msg("reconcile bootstrap failed")

		return 0, err
	}

	db := r.DB()
	startedAt := time.Now()

	logx.FromContext(ctx).Info().Str("integration_id", envelope.IntegrationID).Str("definition_id", envelope.DefinitionID).Str("operation", envelope.Operation).Msg("reconcile cycle started")

	operation, err := r.Registry().Operation(metadata.DefinitionID, envelope.Operation)
	if err != nil {
		return 0, err
	}

	runRecord, err := operations.CreatePendingRun(ctx, db, installation, operations.DispatchRequest{
		IntegrationID: envelope.IntegrationID,
		Operation:     envelope.Operation,
		RunType:       enums.IntegrationRunTypeReconcile,
	})
	if err != nil {
		return 0, err
	}

	if err := operations.MarkRunRunning(ctx, db, runRecord.ID); err != nil {
		return 0, err
	}

	metadata.RunID = runRecord.ID
	ctx = types.WithExecutionMetadata(ctx, metadata)

	response, recordCount, execErr := r.executeResolvedOperation(ctx, installation, operation, nil, nil, false, operations.IngestOptionsFromMetadata(integrationgenerated.IntegrationIngestSourceOperation, metadata))

	if execErr != nil {
		logx.FromContext(ctx).Error().Err(execErr).Str("integration_id", envelope.IntegrationID).Str("operation", envelope.Operation).Str("run_id", runRecord.ID).Msg("reconcile operation failed")

		if completeErr := operations.CompleteRun(ctx, db, runRecord.ID, startedAt, operations.RunResult{
			Status: enums.IntegrationRunStatusFailed,
			Error:  execErr.Error(),
			Metrics: map[string]any{
				"response": jsonx.DecodeAnyOrNil(response),
			},
		}); completeErr != nil {
			return 0, errors.Join(execErr, completeErr)
		}

		if outputErr := river.RecordOutput(ctx, reconcileOutput{
			IntegrationID: envelope.IntegrationID,
			DefinitionID:  envelope.DefinitionID,
			Operation:     envelope.Operation,
			RunID:         runRecord.ID,
			Status:        enums.IntegrationRunStatusFailed,
			Error:         execErr.Error(),
			DurationMS:    time.Since(startedAt).Milliseconds(),
		}); outputErr != nil {
			logx.FromContext(ctx).Error().Err(outputErr).Str("run_id", runRecord.ID).Msg("failed to record river output")
		}

		return 0, execErr
	}

	logx.FromContext(ctx).Info().Str("integration_id", envelope.IntegrationID).Str("operation", envelope.Operation).Str("run_id", runRecord.ID).Int("records", recordCount).Msg("reconcile operation completed")

	if err := operations.CompleteRun(ctx, db, runRecord.ID, startedAt, operations.RunResult{
		Status:  enums.IntegrationRunStatusSuccess,
		Summary: "operation completed",
		Metrics: map[string]any{
			"records":  recordCount,
			"response": jsonx.DecodeAnyOrNil(response),
		},
	}); err != nil {
		return recordCount, err
	}

	if outputErr := river.RecordOutput(ctx, reconcileOutput{
		IntegrationID: envelope.IntegrationID,
		DefinitionID:  envelope.DefinitionID,
		Operation:     envelope.Operation,
		RunID:         runRecord.ID,
		Records:       recordCount,
		Status:        enums.IntegrationRunStatusSuccess,
		DurationMS:    time.Since(startedAt).Milliseconds(),
	}); outputErr != nil {
		return recordCount, outputErr
	}

	return recordCount, nil
}

// ExecuteOperation runs one integration operation inline without run tracking
func (r *Runtime) ExecuteOperation(ctx context.Context, integration *ent.Integration, operation types.OperationRegistration, credentials types.CredentialBindings, config json.RawMessage) (json.RawMessage, error) {
	if integration == nil {
		return nil, ErrInstallationRequired
	}

	if len(config) > 0 {
		if err := validatePayload(operation.ConfigSchema, config, ErrOperationConfigInvalid); err != nil {
			return nil, err
		}
	}

	response, _, err := r.executeResolvedOperation(ctx, integration, operation, credentials, config, false, operations.IngestOptions{
		Source: integrationgenerated.IntegrationIngestSourceOperation,
	})

	return response, err
}

// HandleOperation executes one queued operation envelope through the runtime-managed dependencies
func (r *Runtime) HandleOperation(ctx context.Context, envelope operations.Envelope) error {
	ctx, integration, metadata, bootstrapErr := r.bootstrapHandlerContext(ctx, envelope.ExecutionMetadata)

	startedAt := time.Now()
	db := r.DB()
	tracked := envelope.RunID != ""

	failRun := func(execErr error, response json.RawMessage) error {
		if tracked {
			if completeErr := operations.CompleteRun(ctx, db, envelope.RunID, startedAt, operations.RunResult{
				Status: enums.IntegrationRunStatusFailed,
				Error:  execErr.Error(),
				Metrics: map[string]any{
					"response": jsonx.DecodeAnyOrNil(response),
				},
			}); completeErr != nil {
				execErr = errors.Join(execErr, completeErr)
			}
		}

		if r.postExecutionHook != nil {
			r.postExecutionHook(ctx, envelope, execErr)
		}

		return execErr
	}

	logx.FromContext(ctx).Debug().Str("integration_id", envelope.IntegrationID).Str("definition_id", envelope.DefinitionID).Str("operation", envelope.Operation).Str("run_id", envelope.RunID).Msg("operation started")

	if tracked {
		if err := operations.MarkRunRunning(ctx, db, envelope.RunID); err != nil {
			return err
		}
	}

	if bootstrapErr != nil {
		return failRun(bootstrapErr, nil)
	}

	operation, err := r.Registry().Operation(metadata.DefinitionID, envelope.Operation)
	if err != nil {
		return failRun(err, nil)
	}

	response, _, err := r.executeResolvedOperation(ctx, integration, operation, nil, envelope.Config, envelope.ForceClientRebuild, operations.IngestOptionsFromMetadata(integrationgenerated.IntegrationIngestSourceOperation, metadata))
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("integration_id", envelope.IntegrationID).Str("definition_id", metadata.DefinitionID).Str("operation", envelope.Operation).Str("run_id", envelope.RunID).Msg("operation failed")

		return failRun(err, response)
	}

	logx.FromContext(ctx).Info().Str("integration_id", envelope.IntegrationID).Str("definition_id", metadata.DefinitionID).Str("operation", envelope.Operation).Str("run_id", envelope.RunID).Msg("operation completed")

	var completeErr error

	if tracked {
		completeErr = operations.CompleteRun(ctx, db, envelope.RunID, startedAt, operations.RunResult{
			Status:  enums.IntegrationRunStatusSuccess,
			Summary: "operation completed",
			Metrics: map[string]any{
				"response": jsonx.DecodeAnyOrNil(response),
			},
		})
	}

	if r.postExecutionHook != nil {
		r.postExecutionHook(ctx, envelope, completeErr)
	}

	return completeErr
}

// BuildClientForIntegration builds a typed client for a specific integration installation.
// It resolves credentials from the keystore and delegates to the registered client builder
func (r *Runtime) BuildClientForIntegration(ctx context.Context, integration *ent.Integration, clientID types.ClientID) (any, error) {
	registration, err := r.Registry().Client(integration.DefinitionID, clientID)
	if err != nil {
		return nil, err
	}

	credentials, err := r.loadCredentials(ctx, integration, registration.CredentialRefs)
	if err != nil {
		return nil, err
	}

	return r.keystore().BuildClient(ctx, integration, registration, credentials, nil, false)
}

// executeResolvedOperation executes the given operation with the input integration and registered Operation.
// When integration is nil the client is resolved from the registry's runtime client.
// Returns the response payload, the number of ingest records processed (0 for non-ingest operations), and any error
func (r *Runtime) executeResolvedOperation(ctx context.Context, integration *ent.Integration, operation types.OperationRegistration, credentials types.CredentialBindings, config json.RawMessage, clientForce bool, ingestOptions operations.IngestOptions) (json.RawMessage, int, error) {
	client, definitionID, err := r.resolveOperationClient(ctx, integration, operation, credentials, config, clientForce)
	if err != nil {
		return nil, 0, err
	}

	req := types.OperationRequest{
		Integration: integration,
		Credentials: credentials,
		Client:      client,
		Config:      jsonx.CloneRawMessage(config),
		DB:          r.DB(),
	}

	if operation.IngestHandle != nil {
		payloadSets, err := operation.IngestHandle(ctx, req)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("definition_id", definitionID).Str("operation", operation.Name).Msg("ingest handle failed")

			return nil, 0, err
		}

		var totalEnvelopes int
		for _, ps := range payloadSets {
			totalEnvelopes += len(ps.Envelopes)
		}

		logx.FromContext(ctx).Info().Str("definition_id", definitionID).Str("operation", operation.Name).Int("payload_sets", len(payloadSets)).Int("envelopes", totalEnvelopes).Msg("ingest handle completed")

		if err := operations.EmitPayloadSets(ctx, operations.IngestContext{
			Registry:    r.Registry(),
			DB:          r.DB(),
			Runtime:     r.Gala(),
			Integration: integration,
		}, operation.Name, operation.Ingest, payloadSets, ingestOptions); err != nil {
			return nil, 0, err
		}

		return nil, totalEnvelopes, nil
	}

	response, err := operation.Handle(ctx, req)
	if err != nil {
		return response, 0, err
	}

	return response, 0, nil
}

// resolveOperationClient resolves the client for an operation. When integration
// is non-nil, credentials are loaded from the keystore and the client is built
// via the registered builder. When integration is nil, the pre-built runtime
// client is retrieved from the registry
func (r *Runtime) resolveOperationClient(ctx context.Context, integration *ent.Integration, operation types.OperationRegistration, credentials types.CredentialBindings, config json.RawMessage, clientForce bool) (any, string, error) {
	if !operation.ClientRef.Valid() {
		if integration != nil {
			return nil, integration.DefinitionID, nil
		}

		meta, _ := types.ExecutionMetadataFromContext(ctx)

		return nil, meta.DefinitionID, nil
	}

	if integration == nil {
		meta, _ := types.ExecutionMetadataFromContext(ctx)

		client, ok := r.Registry().RuntimeClient(meta.DefinitionID)
		if !ok {
			return nil, meta.DefinitionID, ErrRuntimeClientNotFound
		}

		logx.FromContext(ctx).Info().Str("definition_id", meta.DefinitionID).Str("operation", operation.Name).Msg("runtime client resolved")

		return client, meta.DefinitionID, nil
	}

	registration, err := r.Registry().Client(integration.DefinitionID, operation.ClientRef)
	if err != nil {
		return nil, integration.DefinitionID, err
	}

	if credentials == nil {
		credentials, err = r.loadCredentials(ctx, integration, registration.CredentialRefs)
		if err != nil {
			return nil, integration.DefinitionID, err
		}
	}

	client, err := r.keystore().BuildClient(ctx, integration, registration, credentials, config, clientForce)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("integration_id", integration.ID).Str("operation", operation.Name).Msg("client build failed")

		return nil, integration.DefinitionID, err
	}

	logx.FromContext(ctx).Info().Str("integration_id", integration.ID).Str("operation", operation.Name).Msg("client initialized")

	return client, integration.DefinitionID, nil
}
