package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/samber/lo"
	"github.com/stripe/stripe-go/v86"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgsubscription"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// operationCompletedSummary is the run summary recorded for a successful operation that ingests no records
const operationCompletedSummary = "operation completed"

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

// HandleReconcile executes one recurring operation cycle inline and returns the delta for adaptive scheduling; envelopes with no integration ID run the scheduled runtime path
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

	if err := r.reconcileCyclePreflight(ctx, installation, envelope.Operation); err != nil {
		return 0, err
	}

	db := r.DB()
	startedAt := time.Now()

	logx.FromContext(ctx).Info().Msg("reconcile cycle started")

	operation, err := r.Registry().Operation(src.DefinitionID, envelope.Operation)
	if err != nil {
		return 0, err
	}

	if operation.DisabledFor(installation.Config.ClientConfig) {
		logx.FromContext(ctx).Debug().Msg("operation is disabled, stopping reconcile cycle")

		return 0, operations.ErrOperationDisabled
	}

	runRecord, err := operations.CreatePendingRun(ctx, db, installation, operation, enums.IntegrationRunTypeReconcile, nil)
	if err != nil {
		return 0, err
	}

	if err := operations.MarkRunRunning(ctx, db, runRecord.ID); err != nil {
		return 0, err
	}

	src.RunID = runRecord.ID
	_ = gala.SetAttributes(&oc, src)
	ctx = intobvs.WithContext(ctx, oc)

	cycle := reconcileCycle{installation: installation, operation: envelope.Operation, src: src, runID: runRecord.ID, startedAt: startedAt}

	response, ingestResult, execErr := r.executeResolvedOperation(ctx, installation, operation, nil, nil, false, operations.IngestOptionsFromOperationContext(oc))
	if execErr != nil {
		return 0, r.failReconcileCycle(ctx, cycle, response, ingestResult, execErr)
	}

	return r.completeReconcileCycle(ctx, cycle, operation, response, ingestResult)
}

// reconcileCycle is one recurring operation cycle's run identity
type reconcileCycle struct {
	installation *ent.Integration
	operation    string
	src          types.IntegrationSource
	runID        string
	startedAt    time.Time
}

// reconcileCyclePreflight reports whether the installation may run the operation's reconcile cycle
func (r *Runtime) reconcileCyclePreflight(ctx context.Context, installation *ent.Integration, operationName string) error {
	if !lo.Contains(enums.IntegrationOperationalStatuses, installation.Status) {
		logx.FromContext(ctx).Info().Str("status", installation.Status.String()).Msg("integration is not operational, skipping current run")

		return operations.ErrOperationDisabled
	}

	if _, failing := installation.Health.UnhealthyOperations[operationName]; failing {
		logx.FromContext(ctx).Info().Msg("operation is marked unhealthy, stopping cycle")

		return operations.ErrOperationDisabled
	}

	ok, err := r.isOrgSubscriptionActive(ctx, installation.OwnerID)
	if err != nil {
		return err
	}

	if !ok {
		logx.FromContext(ctx).Info().Msg("owner subscription is not active, stopping reconcile cycle")

		return operations.ErrOperationDisabled
	}

	return nil
}

// failReconcileCycle records a failed cycle on its run, river output, and installation health
func (r *Runtime) failReconcileCycle(ctx context.Context, cycle reconcileCycle, response json.RawMessage, ingestResult operations.IngestResult, execErr error) error {
	logx.FromContext(ctx).Error().Err(execErr).Msg("reconcile operation failed")

	if completeErr := operations.CompleteRun(ctx, r.DB(), cycle.runID, cycle.startedAt, executionRunResult(false, response, ingestResult, execErr)); completeErr != nil {
		return errors.Join(execErr, completeErr)
	}

	if outputErr := river.RecordOutput(ctx, reconcileOutput{
		IntegrationID: cycle.src.IntegrationID,
		DefinitionID:  cycle.src.DefinitionID,
		Operation:     cycle.operation,
		RunID:         cycle.runID,
		Status:        enums.IntegrationRunStatusFailed,
		Error:         execErr.Error(),
		DurationMS:    time.Since(cycle.startedAt).Milliseconds(),
	}); outputErr != nil {
		logx.FromContext(ctx).Error().Err(outputErr).Msg("failed to record river output")
	}

	unhealthy, isUnhealthy := types.UnhealthyFrom(execErr)
	degraded, isDegraded := types.DegradedFrom(execErr)

	switch {
	case isUnhealthy:
		if markErr := r.MarkIntegrationUnhealthy(ctx, cycle.installation, unhealthy.Reason); markErr != nil {
			logx.FromContext(ctx).Error().Err(markErr).Msg("failed marking integration unhealthy after terminal operation failure")
		}
	case isDegraded:
		if markErr := r.MarkOperationUnhealthy(ctx, cycle.installation, cycle.operation, degraded.Reason); markErr != nil {
			logx.FromContext(ctx).Error().Err(markErr).Msg("failed marking operation unhealthy after terminal operation failure")
		}
	}

	return execErr
}

// completeReconcileCycle records a successful cycle on its run and river output, returning the change delta
func (r *Runtime) completeReconcileCycle(ctx context.Context, cycle reconcileCycle, operation types.OperationRegistration, response json.RawMessage, ingestResult operations.IngestResult) (int, error) {
	delta := ingestResult.Changed

	logx.FromContext(ctx).Info().Int("records", ingestResult.Attempted).Int("changed", delta).Msg("reconcile operation completed")

	if err := operations.CompleteRun(ctx, r.DB(), cycle.runID, cycle.startedAt, executionRunResult(operation.IngestHandle != nil, response, ingestResult, nil)); err != nil {
		return delta, err
	}

	if outputErr := river.RecordOutput(ctx, reconcileOutput{
		IntegrationID: cycle.src.IntegrationID,
		DefinitionID:  cycle.src.DefinitionID,
		Operation:     cycle.operation,
		RunID:         cycle.runID,
		Records:       ingestResult.Attempted,
		Status:        enums.IntegrationRunStatusSuccess,
		DurationMS:    time.Since(cycle.startedAt).Milliseconds(),
	}); outputErr != nil {
		return delta, outputErr
	}

	return delta, nil
}

// executionRunResult renders one execution's terminal run result from its response, ingest counters, and error; ingest selects the record-count summary for successful ingest operations
func executionRunResult(ingest bool, response json.RawMessage, ingestResult operations.IngestResult, execErr error) operations.RunResult {
	metrics := operations.IngestMetrics(ingestResult)
	metrics["response"] = jsonx.DecodeAnyOrNil(response)

	if execErr != nil {
		return operations.RunResult{Status: enums.IntegrationRunStatusFailed, Error: execErr.Error(), Metrics: metrics}
	}

	result := operations.RunResult{Status: enums.IntegrationRunStatusSuccess, Summary: operationCompletedSummary, Metrics: metrics}

	if ingest {
		result.Summary = operations.IngestRunSummary(ingestResult)
	}

	if ingestResult.Failed > 0 {
		result.Error = operations.RecordFailureSummary(ingestResult)
	}

	return result
}

// ExecuteOperation runs one integration operation inline without run tracking
func (r *Runtime) ExecuteOperation(ctx context.Context, integration *ent.Integration, operation types.OperationRegistration, credentials types.CredentialBindings, config json.RawMessage) (json.RawMessage, error) {
	if integration == nil {
		return nil, ErrInstallationRequired
	}

	ctx = auth.EnsureIntegrationCaller(ctx, integration.OwnerID)

	return r.executeOperationInline(ctx, integration, integration.DefinitionID, operation, credentials, config)
}

// ExecuteRuntimeOperation runs one system-initiated operation inline against a definition's cached runtime client, with no Integration installation and no run tracking
func (r *Runtime) ExecuteRuntimeOperation(ctx context.Context, definitionID, operationName string, config json.RawMessage) (json.RawMessage, error) {
	operation, err := r.Registry().Operation(definitionID, operationName)
	if err != nil {
		return nil, err
	}

	return r.executeOperationInline(ctx, nil, definitionID, operation, nil, config)
}

// executeOperationInline runs one integration operation inline without run tracking, if there is no integration ID it runs as an runtime client
func (r *Runtime) executeOperationInline(ctx context.Context, integration *ent.Integration, definitionID string, operation types.OperationRegistration, credentials types.CredentialBindings, config json.RawMessage) (json.RawMessage, error) {
	switch {
	case integration == nil:
		ctx = intobvs.WithContext(ctx, types.NewOperationContext("", operation.Name, types.IntegrationSource{
			DefinitionID: definitionID,
			Runtime:      true,
		}))
	case operation.DisabledFor(integration.Config.ClientConfig):
		return nil, operations.ErrOperationDisabled
	default:
		ctx = intobvs.WithInstallation(ctx, integration)
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

	finish := func(execErr error, ingest bool, response json.RawMessage, ingestResult operations.IngestResult) error {
		if tracked {
			if completeErr := operations.CompleteRun(ctx, db, src.RunID, startedAt, executionRunResult(ingest, response, ingestResult, execErr)); completeErr != nil {
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
		return finish(bootstrapErr, false, nil, operations.IngestResult{})
	}

	if integration != nil {
		ctx = auth.EnsureIntegrationCaller(ctx, integration.OwnerID)
	}

	if tracked {
		resumedCtx, err := r.resumeTrackedRun(ctx, &oc, &src)
		if err != nil {
			return finish(err, false, nil, operations.IngestResult{})
		}

		ctx = resumedCtx
	}

	operation, err := r.Registry().Operation(src.DefinitionID, envelope.Operation)
	if err != nil {
		return finish(err, false, nil, operations.IngestResult{})
	}

	response, ingestResult, err := r.executeResolvedOperation(ctx, integration, operation, nil, envelope.Config, envelope.ForceClientRebuild, operations.IngestOptionsFromOperationContext(oc))

	switch {
	case err != nil:
		logx.FromContext(ctx).Error().Err(err).Msg("operation failed")
	default:
		logx.FromContext(ctx).Info().Msg("operation completed")
	}

	return finish(err, operation.IngestHandle != nil, response, ingestResult)
}

// resumeTrackedRun marks the envelope's run running, continuing under a retry run when the run already left pending
func (r *Runtime) resumeTrackedRun(ctx context.Context, oc *gala.OperationContext, src *types.IntegrationSource) (context.Context, error) {
	db := r.DB()

	run, err := db.IntegrationRun.Get(ctx, src.RunID)
	if err != nil {
		return ctx, err
	}

	if run.Status != enums.IntegrationRunStatusPending {
		retry, err := operations.RetryRun(ctx, db, run)
		if err != nil {
			return ctx, err
		}

		logx.FromContext(ctx).Info().Str("retry_of", run.ID).Str("run_id", retry.ID).Msg("operation attempt continues under a new run")

		src.RunID = retry.ID
		_ = gala.SetAttributes(oc, *src)
		ctx = intobvs.WithContext(ctx, *oc)
	}

	if err := operations.MarkRunRunning(ctx, db, src.RunID); err != nil {
		return ctx, err
	}

	return ctx, nil
}

// BuildClientForIntegration builds a typed client for a specific integration installation
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

// executeResolvedOperation executes the given operation with the input integration and registered Operation
func (r *Runtime) executeResolvedOperation(ctx context.Context, integration *ent.Integration, operation types.OperationRegistration, credentials types.CredentialBindings, config json.RawMessage, clientForce bool, ingestOptions operations.IngestOptions) (json.RawMessage, operations.IngestResult, error) {
	client, credentials, err := r.resolveOperationClient(ctx, integration, operation, credentials, config, clientForce)
	if err != nil {
		return nil, operations.IngestResult{}, err
	}

	if len(config) == 0 && integration != nil && operation.ConfigResolver != nil {
		config = operation.ConfigResolver(integration.Config.ClientConfig)
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
		return nil, operations.IngestResult{}, err
	}

	if !allowed {
		return nil, operations.IngestResult{}, ErrOperationRateLimited
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

	if operation.IngestHandle == nil {
		response, err := operation.Handle(ctx, req)

		return response, operations.IngestResult{}, err
	}

	payloadSets, err := operation.IngestHandle(ctx, req)
	if err != nil {
		return nil, operations.IngestResult{}, err
	}

	logx.FromContext(ctx).Info().Int("payload_sets", len(payloadSets)).Int("envelopes", lo.SumBy(payloadSets, func(ps types.IngestPayloadSet) int { return len(ps.Envelopes) })).Msg("ingest handle completed")

	result, err := operations.ProcessPayloadSets(ctx, operations.IngestContext{
		Registry:    r.Registry(),
		DB:          r.DB(),
		Runtime:     r.Gala(),
		Integration: integration,
	}, operation.Name, operation.Ingest, operation.Policy, payloadSets, ingestOptions)
	if err != nil {
		return nil, result, err
	}

	response, err := json.Marshal(result)
	if err != nil {
		return nil, operations.IngestResult{}, err
	}

	return response, result, nil
}

// SeedReconcileJobs resets the recurring loops of every operational installation whose active definition declares a reconcilable operation
func (r *Runtime) SeedReconcileJobs(ctx context.Context) error {
	definitionIDs := lo.FilterMap(r.Registry().Definitions(), func(def types.Definition, _ int) (string, bool) {
		return def.ID, def.Active && lo.SomeBy(def.Operations, func(op types.OperationRegistration) bool { return op.Policy.Reconcile })
	})
	if len(definitionIDs) == 0 {
		return nil
	}

	systemCtx := auth.WithSystemSweepContext(ctx)

	installations, err := r.DB().Integration.Query().
		Where(
			integration.StatusIn(enums.IntegrationOperationalStatuses...),
			integration.DefinitionIDIn(definitionIDs...),
		).
		All(systemCtx)
	if err != nil {
		return err
	}

	logx.FromContext(ctx).Debug().Int("count", len(installations)).Msg("installations found to check for reconciliation")

	for _, inst := range installations {
		if err := r.seedReconcileJobsForInstallation(systemCtx, inst); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// SeedReconcileJobsForInstallation checks every reconcilable operation on the given installation and emits a ReconcileEnvelope for any that do not have an active River job
func (r *Runtime) SeedReconcileJobsForInstallation(ctx context.Context, inst *ent.Integration) error {
	return r.seedReconcileJobsForInstallation(ctx, inst)
}

// seedReconcileJobsForInstallation is the shared implementation used by both SeedReconcileJobs and SeedReconcileJobsForInstallation
func (r *Runtime) seedReconcileJobsForInstallation(ctx context.Context, inst *ent.Integration) error {
	if !lo.Contains(enums.IntegrationOperationalStatuses, inst.Status) {
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

	unhealthy := inst.Health.UnhealthyOperations

	for _, op := range def.Operations {
		if !op.Policy.Reconcile {
			continue
		}

		if op.DisabledFor(inst.Config.ClientConfig) {
			continue
		}

		if _, failing := unhealthy[op.Name]; failing {
			continue
		}

		opCtx := intobvs.WithOperation(ctx, op.Name)

		if err := r.emitReconcileLoop(opCtx, inst, op.Name); err != nil {
			logx.FromContext(opCtx).Error().Err(err).Msg("failed to seed reconcile job")
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// isOrgSubscriptionActive reports whether the organization's subscription permits recurring operations, always true when entitlements are disabled
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
		Exist(ctx)
}

// PurgeInstallationJobs removes every queued River job bound to the installation across the operation-context and per-record ingest job families and returns how many were purged
func (r *Runtime) PurgeInstallationJobs(ctx context.Context, integrationID string) (int, error) {
	operationJobs, err := types.PropertiesFragment(map[string]string{"entityId": integrationID, "entityType": "integration"})
	if err != nil {
		return 0, err
	}

	ingestJobs, err := types.PropertiesFragment(map[string]string{"integration_id": integrationID})
	if err != nil {
		return 0, err
	}

	var purged int

	for _, fragment := range []string{operationJobs, ingestJobs} {
		count, err := r.Gala().PurgeActiveJobsWithMetadata(ctx, fragment)
		if err != nil {
			return purged, err
		}

		purged += count
	}

	return purged, nil
}

// resolveOperationClient resolves the client and credentials an operation runs with: none for client-less operations, the definition's cached runtime client without an installation, and the installation's built client otherwise
func (r *Runtime) resolveOperationClient(ctx context.Context, integration *ent.Integration, operation types.OperationRegistration, credentials types.CredentialBindings, config json.RawMessage, clientForce bool) (any, types.CredentialBindings, error) {
	switch {
	case !operation.ClientRef.Valid():
		return nil, credentials, nil
	case integration == nil:
		oc, _ := gala.OperationContextFromContext(ctx)

		client, ok := r.Registry().RuntimeClient(types.IntegrationSourceFrom(oc).DefinitionID)
		if !ok {
			return nil, credentials, ErrRuntimeClientNotFound
		}

		logx.FromContext(ctx).Debug().Msg("runtime client resolved")

		return client, credentials, nil
	}

	registration, err := r.Registry().Client(integration.DefinitionID, operation.ClientRef)
	if err != nil {
		return nil, credentials, err
	}

	if credentials == nil {
		credentials, err = r.loadCredentials(ctx, integration, registration.CredentialRefs)
		if err != nil {
			return nil, credentials, err
		}
	}

	client, err := r.keystore().BuildClient(ctx, integration, registration, credentials, config, clientForce)
	if err != nil {
		return nil, credentials, types.Unhealthy(err, fmt.Sprintf(clientUnresolvedReasonFmt, err))
	}

	logx.FromContext(ctx).Debug().Msg("client initialized")

	return client, credentials, nil
}
