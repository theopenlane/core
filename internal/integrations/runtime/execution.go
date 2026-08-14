package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/riverqueue/river"
	"github.com/stripe/stripe-go/v86"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	intobvs "github.com/theopenlane/core/internal/integrations/observability"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// reconcileOperations emits one reconciliation envelope per reconcilable operation,
// starting an independent adaptive scheduling cycle for each
func (r *Runtime) reconcileOperations(ctx context.Context, integration *ent.Integration) error {
	def, ok := r.Registry().Definition(integration.DefinitionID)
	if !ok {
		return ErrDefinitionNotFound
	}

	ctx = intobvs.WithInstallation(ctx, integration)

	var errs []error

	for _, op := range def.Operations {
		if !op.Policy.Reconcile {
			continue
		}

		opCtx := intobvs.WithOperation(ctx, op.Name)

		if op.Disabled != nil && op.Disabled(integration.Config.ClientConfig) {
			logx.FromContext(opCtx).Debug().Msg("operation is disabled, skipping reconcile")

			continue
		}

		if err := r.emitReconcileLoop(opCtx, integration, op.Name); err != nil {
			logx.FromContext(opCtx).Error().Err(err).Msg("failed to emit reconcile envelope")
			errs = append(errs, err)
		}
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

// HandleReconcile executes one recurring operation cycle inline and returns the delta
// for adaptive scheduling; envelopes with no integration ID run the scheduled runtime path
func (r *Runtime) HandleReconcile(ctx context.Context, envelope operations.ReconcileEnvelope) (int, error) {
	oc := envelope.OperationContext
	src := types.IntegrationSourceFrom(oc)
	ctx = intobvs.WithContext(ctx, oc)

	if src.IntegrationID == "" {
		return r.handleScheduledCycle(ctx, envelope)
	}

	installation, err := r.ResolveIntegration(ctx, IntegrationLookup{IntegrationID: src.IntegrationID})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("reconcile bootstrap failed")

		return 0, err
	}

	ctx = auth.EnsureIntegrationCaller(ctx, installation.OwnerID)

	if installation.Status != enums.IntegrationStatusConnected {
		logx.FromContext(ctx).Info().Str("status", installation.Status.String()).Msg("integration is not connected, skipping current run")

		return 0, operations.ErrOperationDisabled
	}

	ok, err := r.isOrgSubscriptionActive(ctx, installation.OwnerID)
	if err != nil {
		return 0, err
	}

	if !ok {
		logx.FromContext(ctx).Info().Msg("owner subscription is not active, stopping reconcile cycle")

		return 0, operations.ErrOperationDisabled
	}

	db := r.DB()
	startedAt := time.Now()

	logx.FromContext(ctx).Info().Msg("reconcile cycle started")

	operation, err := r.Registry().Operation(src.DefinitionID, envelope.Operation)
	if err != nil {
		return 0, err
	}

	if operation.Disabled != nil && operation.Disabled(installation.Config.ClientConfig) {
		logx.FromContext(ctx).Debug().Msg("operation is disabled, stopping reconcile cycle")

		return 0, operations.ErrOperationDisabled
	}

	runRecord, err := operations.CreatePendingRun(ctx, db, installation, envelope.Operation, enums.IntegrationRunTypeReconcile, nil)
	if err != nil {
		return 0, err
	}

	if err := operations.MarkRunRunning(ctx, db, runRecord.ID); err != nil {
		return 0, err
	}

	src.RunID = runRecord.ID
	_ = gala.SetAttributes(&oc, src)
	ctx = intobvs.WithContext(ctx, oc)

	ingestOptions := operations.IngestOptionsFromOperationContext(oc)

	response, recordCount, execErr := r.executeResolvedOperation(ctx, installation, operation, nil, nil, false, ingestOptions)

	if execErr != nil {
		logx.FromContext(ctx).Error().Err(execErr).Msg("reconcile operation failed")

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
			IntegrationID: src.IntegrationID,
			DefinitionID:  src.DefinitionID,
			Operation:     envelope.Operation,
			RunID:         runRecord.ID,
			Status:        enums.IntegrationRunStatusFailed,
			Error:         execErr.Error(),
			DurationMS:    time.Since(startedAt).Milliseconds(),
		}); outputErr != nil {
			logx.FromContext(ctx).Error().Err(outputErr).Msg("failed to record river output")
		}

		if unhealthy, ok := types.UnhealthyFrom(execErr); ok {
			if markErr := r.MarkIntegrationUnhealthy(ctx, installation, unhealthy.Reason); markErr != nil {
				logx.FromContext(ctx).Error().Err(markErr).Msg("failed marking integration unhealthy after terminal operation failure")
			}
		}

		return 0, execErr
	}

	logx.FromContext(ctx).Info().Int("records", recordCount).Msg("reconcile operation completed")

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
		IntegrationID: src.IntegrationID,
		DefinitionID:  src.DefinitionID,
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

	ctx = auth.EnsureIntegrationCaller(ctx, integration.OwnerID)

	return r.executeOperationInline(ctx, integration, integration.DefinitionID, operation, credentials, config)
}

// ExecuteRuntimeOperation runs one system-initiated operation inline against a definition's cached runtime client,
// with no Integration installation and no run tracking. Used for operator-owned calls that need their result back synchronously
func (r *Runtime) ExecuteRuntimeOperation(ctx context.Context, definitionID, operationName string, config json.RawMessage) (json.RawMessage, error) {
	operation, err := r.Registry().Operation(definitionID, operationName)
	if err != nil {
		return nil, err
	}

	return r.executeOperationInline(ctx, nil, definitionID, operation, nil, config)
}

// executeOperationInline runs one integration operation inline without run tracking, if there is no integration ID it runs as an runtime client
func (r *Runtime) executeOperationInline(ctx context.Context, integration *ent.Integration, definitionID string, operation types.OperationRegistration, credentials types.CredentialBindings, config json.RawMessage) (json.RawMessage, error) {
	if integration != nil {
		ctx = intobvs.WithInstallation(ctx, integration)
	} else {
		ctx = intobvs.WithContext(ctx, types.NewOperationContext("", operation.Name, types.IntegrationSource{
			DefinitionID: definitionID,
			Runtime:      true,
		}))
	}

	ctx = intobvs.WithOperation(ctx, operation.Name)

	if len(config) > 0 {
		if err := validatePayload(ctx, operation.ConfigSchema, config, ErrOperationConfigInvalid); err != nil {
			return nil, err
		}
	}

	response, _, err := r.executeResolvedOperation(ctx, integration, operation, credentials, config, false, operations.IngestOptions{})

	return response, err
}

// HandleOperation executes one queued operation envelope through the runtime-managed dependencies
func (r *Runtime) HandleOperation(ctx context.Context, envelope operations.Envelope) error {
	oc := envelope.OperationContext
	src := types.IntegrationSourceFrom(oc)
	ctx = intobvs.WithContext(ctx, oc)

	startedAt := time.Now()
	db := r.DB()
	tracked := src.RunID != ""

	var (
		integration  *ent.Integration
		bootstrapErr error
	)

	if !src.Runtime {
		integration, bootstrapErr = r.ResolveIntegration(ctx, IntegrationLookup{IntegrationID: src.IntegrationID})
	}

	failRun := func(execErr error, response json.RawMessage) error {
		if tracked {
			if completeErr := operations.CompleteRun(ctx, db, src.RunID, startedAt, operations.RunResult{
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

	logx.FromContext(ctx).Debug().Msg("operation started")

	if bootstrapErr != nil {
		return failRun(bootstrapErr, nil)
	}

	if integration != nil {
		ctx = auth.EnsureIntegrationCaller(ctx, integration.OwnerID)
	}

	if tracked {
		if err := operations.MarkRunRunning(ctx, db, src.RunID); err != nil {
			return failRun(err, nil)
		}
	}

	operation, err := r.Registry().Operation(src.DefinitionID, envelope.Operation)
	if err != nil {
		return failRun(err, nil)
	}

	ingestOptions := operations.IngestOptionsFromOperationContext(oc)

	response, _, err := r.executeResolvedOperation(ctx, integration, operation, nil, envelope.Config, envelope.ForceClientRebuild, ingestOptions)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("operation failed")

		return failRun(err, response)
	}

	logx.FromContext(ctx).Info().Msg("operation completed")

	var completeErr error

	if tracked {
		completeErr = operations.CompleteRun(ctx, db, src.RunID, startedAt, operations.RunResult{
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
	client, credentials, _, err := r.resolveOperationClient(ctx, integration, operation, credentials, config, clientForce)
	if err != nil {
		return nil, 0, err
	}

	var lastRunAt *time.Time

	if db := r.dbOrNil(); db != nil && db.IntegrationRun != nil && integration != nil {
		var lastRunErr error

		lastRunAt, lastRunErr = operations.LastSuccessfulRunAt(ctx, db, integration.ID, operation.Name)
		if lastRunErr != nil {
			logx.FromContext(ctx).Warn().Err(lastRunErr).Msg("could not resolve last successful run time, proceeding without incremental filter")
		}
	}

	if lastRunAt == nil && !operation.SkipDefaultLookback {
		t := time.Now().UTC().Add(-r.defaultLookback)
		lastRunAt = &t
	}

	allowed, err := r.checkRateLimit(ctx, operation)
	if err != nil {
		return nil, 0, err
	}

	if !allowed {
		return nil, 0, ErrOperationRateLimited
	}

	req := types.OperationRequest{
		Integration: integration,
		Credentials: credentials,
		Client:      client,
		Config:      jsonx.CloneRawMessage(config),
		LastRunAt:   lastRunAt,
		DB:          r.DB(),
		Dispatch:    r.Dispatch,
		Services:    r,
	}

	if operation.IngestHandle != nil {
		payloadSets, err := operation.IngestHandle(ctx, req)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("ingest handle failed")

			return nil, 0, err
		}

		var totalEnvelopes int
		for _, ps := range payloadSets {
			totalEnvelopes += len(ps.Envelopes)
		}

		logx.FromContext(ctx).Info().Int("payload_sets", len(payloadSets)).Int("envelopes", totalEnvelopes).Msg("ingest handle completed")

		result, err := operations.EmitPayloadSetsWithResult(ctx, operations.IngestContext{
			Registry:    r.Registry(),
			DB:          r.DB(),
			Runtime:     r.Gala(),
			Integration: integration,
		}, operation.Name, operation.Ingest, payloadSets, ingestOptions)
		if err != nil {
			return nil, 0, err
		}

		response, marshalErr := json.Marshal(result)
		if marshalErr != nil {
			return nil, 0, marshalErr
		}
		return response, result.Attempted, nil
	}

	response, err := operation.Handle(ctx, req)
	if err != nil {
		return response, 0, err
	}

	return response, 0, nil
}

// SeedReconcileJobs ensures every connected integration with reconcilable operations
// has an active River job. It is intended to be called once at startup to recover
// reconcile cycles that were lost due to job deletion or a queue flush
func (r *Runtime) SeedReconcileJobs(ctx context.Context) error {
	definitionIDs := r.reconcilableDefinitionIDs()
	if len(definitionIDs) == 0 {
		return nil
	}

	systemCtx := auth.WithCaller(privacy.DecisionContext(ctx, privacy.Allow), &auth.Caller{
		Capabilities: auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation,
	})

	installations, err := r.DB().Integration.Query().
		Where(
			integration.StatusEQ(enums.IntegrationStatusConnected),
			integration.DefinitionIDIn(definitionIDs...),
		).
		All(systemCtx)
	if err != nil {
		return err
	}

	var errs []error

	logx.FromContext(ctx).Debug().Int("count", len(installations)).Msg("installations found to check for reconciliation")

	for _, inst := range installations {
		if err := r.seedReconcileJobsForInstallation(systemCtx, inst); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// SeedReconcileJobsForInstallation checks every reconcilable operation on the given
// installation and emits a ReconcileEnvelope for any that do not have an active River job
func (r *Runtime) SeedReconcileJobsForInstallation(ctx context.Context, inst *ent.Integration) error {
	return r.seedReconcileJobsForInstallation(privacy.DecisionContext(ctx, privacy.Allow), inst)
}

// seedReconcileJobsForInstallation is the shared implementation used by both
// SeedReconcileJobs and SeedReconcileJobsForInstallation
func (r *Runtime) seedReconcileJobsForInstallation(ctx context.Context, inst *ent.Integration) error {
	if inst.Status != enums.IntegrationStatusConnected {
		return nil
	}

	ctx = intobvs.WithInstallation(ctx, inst)

	active, err := r.isOrgSubscriptionActive(ctx, inst.OwnerID)
	if err != nil {
		return err
	}

	if !active {
		logx.FromContext(ctx).Info().Msg("owner subscription is not active, skipping reconcile seed")

		return nil
	}

	def, ok := r.Registry().Definition(inst.DefinitionID)
	if !ok {
		return nil
	}

	var errs []error

	for _, op := range def.Operations {
		if !op.Policy.Reconcile {
			continue
		}

		if op.Disabled != nil && op.Disabled(inst.Config.ClientConfig) {
			continue
		}

		opCtx := intobvs.WithOperation(ctx, op.Name)

		// successor cycles carry per-cycle unique keys, so a live loop only surfaces
		// through its metadata, never through the seed's insert-time key
		fragment, err := types.PropertiesFragment(map[string]string{
			"entityId":  inst.ID,
			"operation": op.Name,
			"runType":   enums.IntegrationRunTypeReconcile.String(),
		})
		if err != nil {
			errs = append(errs, err)
			continue
		}

		active, err := r.Gala().HasActiveJobWithMetadata(opCtx, fragment)
		if err != nil {
			logx.FromContext(opCtx).Error().Err(err).Msg("failed to check for active reconcile job")
			errs = append(errs, err)

			continue
		}

		if active {
			continue
		}

		logx.FromContext(opCtx).Info().Msg("seeding reconcile loop")

		if err := r.emitReconcileLoop(opCtx, inst, op.Name); err != nil {
			logx.FromContext(opCtx).Error().Err(err).Msg("failed to seed reconcile job")
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (r *Runtime) isOrgSubscriptionActive(ctx context.Context, orgID string) (bool, error) {
	client := r.DB()

	if client.EntitlementManager == nil || client.EntitlementManager.Config == nil || !client.EntitlementManager.Config.IsEnabled() {
		return true, nil
	}

	if orgID == "" {
		return false, nil
	}

	return client.OrgSubscription.Query().
		Where(
			orgsubscription.OwnerIDEQ(orgID),
			orgsubscription.Or(
				orgsubscription.ActiveEQ(true),
				orgsubscription.StripeSubscriptionStatusEQ(string(stripe.SubscriptionStatusTrialing)),
			),
			orgsubscription.StripeSubscriptionStatusNEQ(string(stripe.SubscriptionStatusCanceled)),
		).
		Exist(privacy.DecisionContext(ctx, privacy.Allow))
}

// reconcilableDefinitionIDs returns the IDs of all registered definitions that
// have at least one operation with Policy.Reconcile set
func (r *Runtime) reconcilableDefinitionIDs() []string {
	var ids []string

	for _, spec := range r.Registry().Catalog() {
		def, ok := r.Registry().Definition(spec.ID)
		if !ok {
			continue
		}

		if !def.Active {
			continue
		}

		for _, op := range def.Operations {
			if op.Policy.Reconcile {
				ids = append(ids, spec.ID)
				break
			}
		}
	}

	return ids
}

// CancelInstallationJobs cancels every queued River job bound to the installation across all
// job families and returns how many were cancelled. Operation-context jobs (reconcile loops,
// event operations) carry the installation as properties.entityId; ingest record jobs carry
// properties.integration_id
func (r *Runtime) CancelInstallationJobs(ctx context.Context, integrationID string) (int, error) {
	fragments, err := installationJobFragments(integrationID)
	if err != nil {
		return 0, err
	}

	var cancelled int

	for _, fragment := range fragments {
		count, err := r.Gala().CancelActiveJobsWithMetadata(ctx, fragment)
		if err != nil {
			return cancelled, err
		}

		cancelled += count
	}

	return cancelled, nil
}

// installationJobFragments builds the JSONB containment fragments matching every job family
// bound to one installation
func installationJobFragments(integrationID string) ([]string, error) {
	operationJobs, err := types.PropertiesFragment(map[string]string{"entityId": integrationID, "entityType": "integration"})
	if err != nil {
		return nil, err
	}

	ingestJobs, err := types.PropertiesFragment(map[string]string{"integration_id": integrationID})
	if err != nil {
		return nil, err
	}

	return []string{operationJobs, ingestJobs}, nil
}

// resolveOperationClient resolves the client for an operation. When integration
// is non-nil, credentials are loaded from the keystore and the client is built
// via the registered builder. When integration is nil, the pre-built runtime
// client is retrieved from the registry
func (r *Runtime) resolveOperationClient(ctx context.Context, integration *ent.Integration, operation types.OperationRegistration, credentials types.CredentialBindings, config json.RawMessage, clientForce bool) (any, types.CredentialBindings, string, error) {
	if !operation.ClientRef.Valid() {
		if integration != nil {
			return nil, credentials, integration.DefinitionID, nil
		}

		oc, _ := gala.OperationContextFromContext(ctx)
		definitionID := types.IntegrationSourceFrom(oc).DefinitionID

		return nil, credentials, definitionID, nil
	}

	if integration == nil {
		oc, _ := gala.OperationContextFromContext(ctx)
		definitionID := types.IntegrationSourceFrom(oc).DefinitionID

		client, ok := r.Registry().RuntimeClient(definitionID)
		if !ok {
			return nil, credentials, definitionID, ErrRuntimeClientNotFound
		}

		logx.FromContext(ctx).Debug().Msg("runtime client resolved")

		return client, credentials, definitionID, nil
	}

	registration, err := r.Registry().Client(integration.DefinitionID, operation.ClientRef)
	if err != nil {
		return nil, credentials, integration.DefinitionID, err
	}

	if credentials == nil {
		credentials, err = r.loadCredentials(ctx, integration, registration.CredentialRefs)
		if err != nil {
			return nil, credentials, integration.DefinitionID, err
		}
	}

	client, err := r.keystore().BuildClient(ctx, integration, registration, credentials, config, clientForce)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("client build failed")

		return nil, credentials, integration.DefinitionID, err
	}

	logx.FromContext(ctx).Debug().Msg("client initialized")

	return client, credentials, integration.DefinitionID, nil
}
