package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/riverqueue/river"
	"github.com/stripe/stripe-go/v84"
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

	ctx = intobvs.WithIntegration(ctx, integration.ID, integration.DefinitionID)

	var errs []error

	for _, op := range def.Operations {
		if !op.Policy.Reconcile {
			continue
		}

		if op.Disabled != nil && op.Disabled(integration.Config.ClientConfig) {
			logx.FromContext(ctx).Debug().Str(intobvs.FieldOperation, op.Name).Msg("operation is disabled, skipping reconcile")

			continue
		}

		oc := types.NewOperationContext(integration.OwnerID, op.Name, types.IntegrationSource{
			IntegrationID: integration.ID,
			DefinitionID:  integration.DefinitionID,
			RunType:       enums.IntegrationRunTypeReconcile,
		})

		receipt := r.Gala().EmitWithHeaders(gala.WithOperationContext(ctx, oc), operations.ReconcileTopic, operations.ReconcileEnvelope{
			OperationContext: oc,
		}, gala.Headers{
			Properties: oc.Properties(),
		})

		if receipt.Err != nil {
			logx.FromContext(ctx).Error().Err(receipt.Err).Str(intobvs.FieldOperation, op.Name).Msg("failed to emit reconcile envelope")
			errs = append(errs, receipt.Err)

			continue
		}

		logx.FromContext(ctx).Info().Str(intobvs.FieldOperation, op.Name).Msg("reconcile envelope emitted")
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

	ctx = ensureCallerOrg(ctx, installation.OwnerID)

	if installation.Status != enums.IntegrationStatusConnected {
		logx.FromContext(ctx).Info().Str("integration_id", installation.ID).
			Str("status", installation.Status.String()).
			Msg("integration is not connected, skipping current run")

		return 0, operations.ErrOperationDisabled
	}

	ok, err := r.isOrgSubscriptionActive(ctx, installation.OwnerID)
	if err != nil {
		return 0, err
	}

	if !ok {
		logx.FromContext(ctx).Info().Str("integration_id", installation.ID).Str("owner_id", installation.OwnerID).Msg("owner subscription is not active, stopping reconcile cycle")

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

	runRecord, err := operations.CreatePendingRun(ctx, db, installation, types.DispatchRequest{
		IntegrationID: src.IntegrationID,
		Operation:     envelope.Operation,
		RunType:       enums.IntegrationRunTypeReconcile,
	})
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
	ingestOptions.CompleteDirectorySnapshot = true

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
		ctx = intobvs.WithIntegration(ctx, integration.ID, integration.DefinitionID)
	} else {
		ctx = intobvs.WithIntegration(ctx, "", definitionID)
		ctx = gala.WithOperationContext(ctx, types.NewOperationContext("", operation.Name, types.IntegrationSource{
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

	response, _, err := r.executeResolvedOperation(ctx, integration, operation, credentials, config, false, operations.IngestOptions{
		CompleteDirectorySnapshot: true,
	})

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
		ctx = ensureCallerOrg(ctx, integration.OwnerID)
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
	ingestOptions.CompleteDirectorySnapshot = true

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
		if err := r.seedReconcileJobsForInstallation(ensureCallerOrg(systemCtx, inst.OwnerID), inst); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// SeedReconcileJobsForInstallation checks every reconcilable operation on the given
// installation and emits a ReconcileEnvelope for any that do not have an active River job
func (r *Runtime) SeedReconcileJobsForInstallation(ctx context.Context, inst *ent.Integration) error {
	systemCtx := auth.WithCaller(privacy.DecisionContext(ctx, privacy.Allow), &auth.Caller{
		Capabilities: auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation,
	})

	return r.seedReconcileJobsForInstallation(ensureCallerOrg(systemCtx, inst.OwnerID), inst)
}

// seedReconcileJobsForInstallation is the shared implementation used by both
// SeedReconcileJobs and SeedReconcileJobsForInstallation
func (r *Runtime) seedReconcileJobsForInstallation(ctx context.Context, inst *ent.Integration) error {
	if inst.Status != enums.IntegrationStatusConnected {
		return nil
	}

	active, err := r.isOrgSubscriptionActive(ctx, inst.OwnerID)
	if err != nil {
		return err
	}

	if !active {
		logx.FromContext(ctx).Info().Str("integration_id", inst.ID).Str("owner_id", inst.OwnerID).Msg("owner subscription is not active, skipping reconcile seed")

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

		fragment, err := reconcileMetadataFragment(inst.ID, op.Name)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		active, err := r.Gala().HasActiveJobWithMetadata(ctx, fragment)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("integration_id", inst.ID).Str(intobvs.FieldOperation, op.Name).Msg("failed to check for active reconcile job")
			errs = append(errs, err)
			continue
		}

		if active {
			logx.FromContext(ctx).Debug().Str("integration_id", inst.ID).Str(intobvs.FieldOperation, op.Name).Msg("reconcile job already active, skipping seed")
			continue
		}

		logx.FromContext(ctx).Info().Str("integration_id", inst.ID).Str(intobvs.FieldOperation, op.Name).Msg("seeding missing reconcile job")

		oc := types.NewOperationContext(inst.OwnerID, op.Name, types.IntegrationSource{
			IntegrationID: inst.ID,
			DefinitionID:  inst.DefinitionID,
			RunType:       enums.IntegrationRunTypeReconcile,
		})

		receipt := r.Gala().EmitWithHeaders(
			gala.WithOperationContext(ctx, oc),
			operations.ReconcileTopic,
			operations.ReconcileEnvelope{OperationContext: oc},
			gala.Headers{Properties: oc.Properties()},
		)
		if receipt.Err != nil {
			logx.FromContext(ctx).Error().Err(receipt.Err).Str("integration_id", inst.ID).Str(intobvs.FieldOperation, op.Name).Msg("failed to seed reconcile job")
			errs = append(errs, receipt.Err)
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

// ensureCallerOrg sets the organization on the existing caller if not already present
func ensureCallerOrg(ctx context.Context, orgID string) context.Context {
	if orgID == "" {
		return ctx
	}

	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		return ctx
	}

	if _, hasOrg := caller.ActiveOrg(); hasOrg {
		return ctx
	}

	scoped := *caller
	scoped.OrganizationID = orgID
	scoped.OrganizationIDs = append([]string{orgID}, caller.OrgIDs()...)

	return auth.WithCaller(ctx, &scoped)
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

// reconcileMetadataFragment builds the JSONB containment fragment used to query
// River for an active reconcile job for the given integration and operation
func reconcileMetadataFragment(integrationID, operationName string) (string, error) {
	fragment := struct {
		Properties struct {
			InstallationID string `json:"installationId"`
			Operation      string `json:"operation"`
		} `json:"properties"`
	}{}

	fragment.Properties.InstallationID = integrationID
	fragment.Properties.Operation = operationName

	b, err := json.Marshal(fragment)
	if err != nil {
		return "", err
	}

	return string(b), nil
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
