package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/notifications"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// integrationUnhealthyObjectType is the notification object type for a stopped installation
const integrationUnhealthyObjectType = "INTEGRATION_RECONFIGURATION_REQUIRED"

// integrationHealthyObjectType is the notification object type for a recovered installation
const integrationHealthyObjectType = "INTEGRATION_RECONNECTED"

// integrationDegradedObjectType is the notification object type for a partially working installation
const integrationDegradedObjectType = "INTEGRATION_OPERATION_DEGRADED"

// ConnectionHealthResult reports the connection-level health check outcome
type ConnectionHealthResult struct {
	// Healthy reports whether the connection's credentials passed the health check
	Healthy bool `json:"healthy"`
	// Reason explains the failure when unhealthy
	Reason string `json:"reason,omitempty"`
}

// OperationHealthResult reports one operation's health outcome
type OperationHealthResult struct {
	// Name is the operation name
	Name string `json:"name"`
	// Healthy reports whether the operation's prerequisites are satisfied
	Healthy bool `json:"healthy"`
	// Reason explains the failure when unhealthy
	Reason string `json:"reason,omitempty"`
}

// HealthAssessment reports the outcome of one installation health run
type HealthAssessment struct {
	// Status is the installation status after the assessment
	Status enums.IntegrationStatus `json:"status"`
	// Connection is the connection-level check outcome
	Connection ConnectionHealthResult `json:"connection"`
	// Operations lists per-operation probe outcomes and recorded failures
	Operations []OperationHealthResult `json:"operations,omitempty"`
}

// MarkIntegrationUnhealthy flags one installation as errored with a user-facing reason and
// notifies the owning organization; recurring cycles stop on their next status check. An
// installation that is already errored is left as is so repeated failures don't stack
// duplicate notifications
func (r *Runtime) MarkIntegrationUnhealthy(ctx context.Context, installation *ent.Integration, reason string) error {
	health := installation.Health
	health.UnhealthyReason = reason

	// health marking runs from worker contexts without a privileged caller; both writes are
	// server-internal and require the allow decision per the notification mutation policy
	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	transitioned, err := r.DB().Integration.Update().
		Where(integration.ID(installation.ID), integration.StatusNEQ(enums.IntegrationStatusErrored)).
		SetStatus(enums.IntegrationStatusErrored).
		SetHealth(health).
		Save(systemCtx)
	if err != nil {
		return err
	}

	if transitioned == 0 {
		return nil
	}

	installation.Status = enums.IntegrationStatusErrored
	installation.Health = health

	displayName := r.integrationDisplayName(installation)

	logx.FromContext(ctx).Warn().Str("reason", reason).Msg("integration marked unhealthy, recurring operations will stop")

	return r.notifyIntegrationHealth(systemCtx, installation, integrationUnhealthyObjectType,
		fmt.Sprintf("%s has stopped syncing", displayName),
		fmt.Sprintf("The %s integration has stopped syncing: %s. Reconnect it to resume.", displayName, reason),
		map[string]any{
			"integration_id": installation.ID,
			"definition_id":  installation.DefinitionID,
			"reason":         reason,
			// the console addresses integrations by definition id
			"url": entityops.ConsoleObjectPath(ent.TypeIntegration, installation.DefinitionID),
		})
}

// ClearIntegrationUnhealthy returns an errored installation to connected, wipes its recorded
// failures so probes and runtime signals re-derive them, notifies the owning organization, and
// reseeds its recurring operations; a non-errored installation is left as is so concurrent
// recoveries don't stack duplicate notifications
func (r *Runtime) ClearIntegrationUnhealthy(ctx context.Context, installation *ent.Integration) error {
	health := installation.Health
	health.UnhealthyReason = ""
	health.UnhealthyOperations = nil

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	transitioned, err := r.DB().Integration.Update().
		Where(integration.ID(installation.ID), integration.StatusEQ(enums.IntegrationStatusErrored)).
		SetStatus(enums.IntegrationStatusConnected).
		SetHealth(health).
		Save(systemCtx)
	if err != nil {
		return err
	}

	if transitioned == 0 {
		return nil
	}

	installation.Status = enums.IntegrationStatusConnected
	installation.Health = health

	displayName := r.integrationDisplayName(installation)

	logx.FromContext(ctx).Info().Msg("integration recovered, recurring operations resume")

	if err := r.notifyIntegrationHealth(systemCtx, installation, integrationHealthyObjectType,
		fmt.Sprintf("%s is syncing again", displayName),
		fmt.Sprintf("The %s integration reconnected and syncing has resumed.", displayName),
		map[string]any{
			"integration_id": installation.ID,
			"definition_id":  installation.DefinitionID,
			"url":            entityops.ConsoleObjectPath(ent.TypeIntegration, installation.DefinitionID),
		}); err != nil {
		return err
	}

	return r.SeedReconcileJobsForInstallation(ctx, installation)
}

// MarkOperationUnhealthy records one operation as failing on an installation and stops its
// recurring loop; when no healthy workload operation remains the whole installation is marked
// unhealthy instead. An operation already recorded is left as is so repeated failures don't
// stack duplicate notifications
func (r *Runtime) MarkOperationUnhealthy(ctx context.Context, installation *ent.Integration, operationName, reason string) error {
	health := installation.Health
	if _, recorded := health.UnhealthyOperations[operationName]; recorded {
		return nil
	}

	def, ok := r.Registry().Definition(installation.DefinitionID)
	if !ok {
		return nil
	}

	unhealthy := make(map[string]string, len(health.UnhealthyOperations)+1)
	for name, why := range health.UnhealthyOperations {
		unhealthy[name] = why
	}

	unhealthy[operationName] = reason
	health.UnhealthyOperations = unhealthy
	installation.Health = health

	healthyRemain := lo.SomeBy(workloadOperations(def, installation), func(op types.OperationRegistration) bool {
		_, failing := unhealthy[op.Name]

		return !failing
	})

	if !healthyRemain {
		return r.MarkIntegrationUnhealthy(ctx, installation, reason)
	}

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	transitioned, err := r.DB().Integration.Update().
		Where(integration.ID(installation.ID), integration.StatusIn(enums.IntegrationOperationalStatuses...)).
		SetStatus(enums.IntegrationStatusDegraded).
		SetHealth(health).
		Save(systemCtx)
	if err != nil {
		return err
	}

	if transitioned == 0 {
		return nil
	}

	installation.Status = enums.IntegrationStatusDegraded

	fragment, err := types.PropertiesFragment(map[string]string{
		"entityId":  installation.ID,
		"operation": operationName,
		"runType":   enums.IntegrationRunTypeReconcile.String(),
	})
	if err != nil {
		return err
	}

	// the loop also self-cancels on its next status check, so a purge failure only logs
	if _, err := r.Gala().PurgeActiveJobsWithMetadata(ctx, fragment); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("operation", operationName).Msg("failed purging unhealthy operation jobs")
	}

	displayName := r.integrationDisplayName(installation)

	logx.FromContext(ctx).Warn().Str("operation", operationName).Str("reason", reason).Msg("integration operation marked unhealthy, its recurring loop will stop")

	return r.notifyIntegrationHealth(systemCtx, installation, integrationDegradedObjectType,
		fmt.Sprintf("%s is partially working", displayName),
		fmt.Sprintf("The %s integration's %s operation has stopped: %s. Other operations continue to run.", displayName, operationName, reason),
		map[string]any{
			"integration_id": installation.ID,
			"definition_id":  installation.DefinitionID,
			"operation":      operationName,
			"reason":         reason,
			"url":            entityops.ConsoleObjectPath(ent.TypeIntegration, installation.DefinitionID),
		})
}

// ClearOperationUnhealthy removes one operation's failure record and reseeds its recurring
// loop; the installation returns to connected when no failing operation remains
func (r *Runtime) ClearOperationUnhealthy(ctx context.Context, installation *ent.Integration, operationName string) error {
	health := installation.Health
	if _, recorded := health.UnhealthyOperations[operationName]; !recorded {
		return nil
	}

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	unhealthy := make(map[string]string, len(health.UnhealthyOperations))

	for name, why := range health.UnhealthyOperations {
		if name != operationName {
			unhealthy[name] = why
		}
	}

	health.UnhealthyOperations = unhealthy
	if len(unhealthy) == 0 {
		health.UnhealthyOperations = nil
	}

	installation.Health = health

	if len(unhealthy) == 0 {
		transitioned, err := r.DB().Integration.Update().
			Where(integration.ID(installation.ID), integration.StatusEQ(enums.IntegrationStatusDegraded)).
			SetStatus(enums.IntegrationStatusConnected).
			SetHealth(health).
			Save(systemCtx)
		if err != nil {
			return err
		}

		if transitioned == 0 {
			return r.DB().Integration.UpdateOneID(installation.ID).SetHealth(health).Exec(systemCtx)
		}

		installation.Status = enums.IntegrationStatusConnected

		logx.FromContext(ctx).Info().Str("operation", operationName).Msg("all operations recovered, integration returns to connected")

		displayName := r.integrationDisplayName(installation)

		// the status transition hook reseeds every loop
		return r.notifyIntegrationHealth(systemCtx, installation, integrationHealthyObjectType,
			fmt.Sprintf("%s is fully operational", displayName),
			fmt.Sprintf("The %s integration's operations all recovered and syncing has resumed.", displayName),
			map[string]any{
				"integration_id": installation.ID,
				"definition_id":  installation.DefinitionID,
				"url":            entityops.ConsoleObjectPath(ent.TypeIntegration, installation.DefinitionID),
			})
	}

	if err := r.DB().Integration.UpdateOneID(installation.ID).SetHealth(health).Exec(systemCtx); err != nil {
		return err
	}

	logx.FromContext(ctx).Info().Str("operation", operationName).Msg("operation recovered, its recurring loop resumes")

	op, err := r.Registry().Operation(installation.DefinitionID, operationName)
	if err != nil {
		return err
	}

	if !op.Policy.Reconcile {
		return nil
	}

	active, err := r.isOrgSubscriptionActive(ctx, installation.OwnerID)
	if err != nil {
		return err
	}

	if !active {
		return nil
	}

	return r.emitReconcileLoop(ctx, installation, operationName)
}

// RunHealthAssessment executes the connection health check and every operation probe for one
// installation, records the resulting health state, and returns the assessment
func (r *Runtime) RunHealthAssessment(ctx context.Context, installation *ent.Integration) (HealthAssessment, error) {
	def, err := r.resolveDefinitionForInstallation(installation)
	if err != nil {
		return HealthAssessment{}, err
	}

	// connectionless definitions (push-based providers) have no credentials to exercise
	var connection types.ConnectionRegistration

	if len(def.Connections) > 0 {
		connection, err = r.resolvePersistedConnection(def, installation)
		if err != nil {
			return HealthAssessment{}, err
		}
	}

	if connection.HealthCheck != nil {
		bindings, err := r.loadCredentials(privacy.DecisionContext(ctx, privacy.Allow), installation, connection.CredentialRefs)
		if err != nil {
			return HealthAssessment{}, err
		}

		if checkErr := r.runConnectionHealthCheck(ctx, installation, connection, bindings); checkErr != nil {
			if markErr := r.MarkIntegrationUnhealthy(ctx, installation, checkErr.Error()); markErr != nil {
				logx.FromContext(ctx).Error().Err(markErr).Msg("failed marking integration unhealthy after failed health check")
			}

			return HealthAssessment{
				Status:     enums.IntegrationStatusErrored,
				Connection: ConnectionHealthResult{Reason: checkErr.Error()},
				Operations: appendRecordedResults(nil, installation),
			}, nil
		}
	}

	if installation.Status == enums.IntegrationStatusErrored {
		if err := r.ClearIntegrationUnhealthy(ctx, installation); err != nil {
			return HealthAssessment{}, err
		}
	}

	results := appendRecordedResults(r.assessOperationHealth(ctx, installation, def), installation)

	if err := r.stampHealthCheck(ctx, installation); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to persist health check timestamp")
	}

	return HealthAssessment{
		Status:     installation.Status,
		Connection: ConnectionHealthResult{Healthy: true},
		Operations: results,
	}, nil
}

// assessOperationHealth probes every workload operation with a registered health check and
// records the per-operation outcome; probe failures degrade the operation instead of failing
// the caller
func (r *Runtime) assessOperationHealth(ctx context.Context, installation *ent.Integration, def types.Definition) []OperationHealthResult {
	var results []OperationHealthResult

	for _, op := range workloadOperations(def, installation) {
		if op.HealthCheck == nil {
			continue
		}

		err := r.runOperationProbe(ctx, installation, op)
		if err == nil {
			results = append(results, OperationHealthResult{Name: op.Name, Healthy: true})

			if clearErr := r.ClearOperationUnhealthy(ctx, installation, op.Name); clearErr != nil {
				logx.FromContext(ctx).Error().Err(clearErr).Str("operation", op.Name).Msg("failed clearing recovered operation")
			}

			continue
		}

		reason := err.Error()
		if degraded, ok := types.DegradedFrom(err); ok {
			reason = degraded.Reason
		}

		results = append(results, OperationHealthResult{Name: op.Name, Healthy: false, Reason: reason})

		if markErr := r.MarkOperationUnhealthy(ctx, installation, op.Name, reason); markErr != nil {
			logx.FromContext(ctx).Error().Err(markErr).Str("operation", op.Name).Msg("failed marking operation unhealthy")
		}
	}

	return results
}

// runConnectionHealthCheck executes one connection mode's health check against the supplied bindings
func (r *Runtime) runConnectionHealthCheck(ctx context.Context, installation *ent.Integration, connection types.ConnectionRegistration, bindings types.CredentialBindings) error {
	var client any

	if connection.HealthCheck.ClientRef.Valid() {
		registration, err := r.Registry().Client(installation.DefinitionID, connection.HealthCheck.ClientRef)
		if err != nil {
			return err
		}

		client, err = r.keystore().BuildClient(ctx, installation, registration, bindings, nil, false)
		if err != nil {
			return err
		}
	}

	_, err := connection.HealthCheck.Handle(ctx, types.OperationRequest{
		Integration: installation,
		Credentials: bindings,
		Client:      client,
		DB:          r.DB(),
		Services:    r,
	})

	return err
}

// runOperationProbe executes one operation's health probe under the operation's own client
func (r *Runtime) runOperationProbe(ctx context.Context, installation *ent.Integration, operation types.OperationRegistration) error {
	client, credentials, _, err := r.resolveOperationClient(privacy.DecisionContext(ctx, privacy.Allow), installation, operation, nil, nil, false)
	if err != nil {
		return err
	}

	_, err = operation.HealthCheck(ctx, types.OperationRequest{
		Integration: installation,
		Credentials: credentials,
		Client:      client,
		DB:          r.DB(),
		Services:    r,
	})

	return err
}

// verifyInstallationHealth runs the persisted connection's health check under stored
// credentials; connections without one pass
func (r *Runtime) verifyInstallationHealth(ctx context.Context, installation *ent.Integration, def types.Definition) error {
	connection, err := r.resolvePersistedConnection(def, installation)
	if err != nil {
		return err
	}

	if connection.HealthCheck == nil {
		return nil
	}

	bindings, err := r.loadCredentials(privacy.DecisionContext(ctx, privacy.Allow), installation, connection.CredentialRefs)
	if err != nil {
		return err
	}

	if err := r.runConnectionHealthCheck(ctx, installation, connection, bindings); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("health check failed after configuration update, installation stays errored")

		return err
	}

	return nil
}

// stampHealthCheck records the assessment time on the installation health record
func (r *Runtime) stampHealthCheck(ctx context.Context, installation *ent.Integration) error {
	now := time.Now().UTC()

	health := installation.Health
	health.LastSuccessfulHealthCheck = &now
	installation.Health = health

	return r.DB().Integration.UpdateOneID(installation.ID).
		SetHealth(health).
		Exec(privacy.DecisionContext(ctx, privacy.Allow))
}

// appendRecordedResults adds recorded failures for operations the probe sweep did not cover
func appendRecordedResults(results []OperationHealthResult, installation *ent.Integration) []OperationHealthResult {
	probed := lo.SliceToMap(results, func(result OperationHealthResult) (string, struct{}) { return result.Name, struct{}{} })

	for name, reason := range installation.Health.UnhealthyOperations {
		if _, ok := probed[name]; ok {
			continue
		}

		results = append(results, OperationHealthResult{Name: name, Healthy: false, Reason: reason})
	}

	return results
}

// workloadOperations returns the operations that do work for one installation, excluding
// internal operations and those disabled globally or for the installation's user input
func workloadOperations(def types.Definition, installation *ent.Integration) []types.OperationRegistration {
	return lo.Filter(def.Operations, func(op types.OperationRegistration, _ int) bool {
		switch {
		case op.Internal, op.DisabledForAll:
			return false
		case op.Disabled != nil && op.Disabled(installation.Config.ClientConfig):
			return false
		}

		return true
	})
}

// notifyIntegrationHealth sends one health notification to the owning organization's owners and admins
func (r *Runtime) notifyIntegrationHealth(ctx context.Context, installation *ent.Integration, objectType, title, body string, data map[string]any) error {
	ids, err := notifications.OrgUserIDsByRole(ctx, r.DB(), installation.OwnerID, enums.RoleOwner, enums.RoleAdmin)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	topic := enums.NotificationTopicIntegration

	return entityops.CreateNotifications(ctx, r.DB(), ids, &ent.CreateNotificationInput{
		NotificationType: enums.NotificationTypeOrganization,
		ObjectType:       objectType,
		Title:            title,
		Body:             body,
		Data:             data,
		Topic:            &topic,
		OwnerID:          &installation.OwnerID,
	})
}

// integrationDisplayName resolves the definition display name for one installation, falling
// back to the installation's own name when the definition is no longer registered
func (r *Runtime) integrationDisplayName(installation *ent.Integration) string {
	if def, ok := r.Registry().Definition(installation.DefinitionID); ok {
		return def.DisplayName
	}

	return installation.Name
}
