package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/notifications"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// integrationUnhealthyObjectType is the notification object type for a stopped installation
const integrationUnhealthyObjectType = "INTEGRATION_RECONFIGURATION_REQUIRED"

// integrationHealthyObjectType is the notification object type for a recovered installation
const integrationHealthyObjectType = "INTEGRATION_RECONNECTED"

// integrationDegradedObjectType is the notification type for a partially working installation
const integrationDegradedObjectType = "INTEGRATION_OPERATION_DEGRADED"

// notificationFieldReason is the notification data key carrying the failure reason
const notificationFieldReason = "reason"

// notificationFieldOperation is the notification data key carrying the affected operation name
const notificationFieldOperation = "operation"

// notificationFieldURL is the notification data key carrying the console link
const notificationFieldURL = "url"

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

// transitionStatus conditionally moves the installation to status with the given health, mirroring the row on the record when it transitioned
func (r *Runtime) transitionStatus(ctx context.Context, installation *ent.Integration, from predicate.Integration, status enums.IntegrationStatus, health models.IntegrationHealth, clearExpiry bool) (bool, error) {
	update := r.DB().Integration.Update().
		Where(integration.ID(installation.ID), from).
		SetStatus(status).
		SetHealth(health)

	if clearExpiry {
		update = update.ClearExpiresAt()
	}

	transitioned, err := update.Save(privacy.DecisionContext(ctx, privacy.Allow))
	if err != nil {
		return false, err
	}

	if transitioned == 0 {
		return false, nil
	}

	installation.Status = status
	installation.Health = health

	return true, nil
}

// notifyHealthChange sends a health notification carrying the installation identity, console link, and any extra fields
func (r *Runtime) notifyHealthChange(ctx context.Context, installation *ent.Integration, objectType, title, body string, extra map[string]any) error {
	data := lo.Assign(map[string]any{
		intobvs.FieldIntegrationID: installation.ID,
		intobvs.FieldDefinitionID:  installation.DefinitionID,
		notificationFieldURL:       entityops.ConsoleObjectPath(ent.TypeIntegration, installation.DefinitionID),
	}, extra)

	ids, err := notifications.OrgUserIDsByRole(ctx, r.DB(), installation.OwnerID, enums.RoleOwner, enums.RoleSuperAdmin)
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

// MarkIntegrationUnhealthy flags an installation errored and notifies the owning organization
func (r *Runtime) MarkIntegrationUnhealthy(ctx context.Context, installation *ent.Integration, reason string) error {
	health := installation.Health
	health.UnhealthyReason = reason

	transitioned, err := r.transitionStatus(ctx, installation, integration.StatusNEQ(enums.IntegrationStatusErrored), enums.IntegrationStatusErrored, health, false)
	if err != nil || !transitioned {
		return err
	}

	displayName := r.integrationDisplayName(installation)

	logx.FromContext(ctx).Warn().Str(notificationFieldReason, reason).Msg("integration marked unhealthy, recurring operations will stop")

	return r.notifyHealthChange(privacy.DecisionContext(ctx, privacy.Allow), installation, integrationUnhealthyObjectType,
		fmt.Sprintf("%s has stopped syncing", displayName),
		fmt.Sprintf("The %s integration has stopped syncing: %s. Reconnect it to resume.", displayName, reason),
		map[string]any{notificationFieldReason: reason})
}

// ClearIntegrationUnhealthy returns an errored installation to connected and reseeds its loops
func (r *Runtime) ClearIntegrationUnhealthy(ctx context.Context, installation *ent.Integration) error {
	health := installation.Health
	health.UnhealthyReason = ""
	health.UnhealthyOperations = nil

	transitioned, err := r.transitionStatus(ctx, installation, integration.StatusEQ(enums.IntegrationStatusErrored), enums.IntegrationStatusConnected, health, true)
	if err != nil || !transitioned {
		return err
	}

	displayName := r.integrationDisplayName(installation)

	logx.FromContext(ctx).Info().Msg("integration recovered, recurring operations resume")

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	if err := r.notifyHealthChange(systemCtx, installation, integrationHealthyObjectType,
		fmt.Sprintf("%s is syncing again", displayName),
		fmt.Sprintf("The %s integration reconnected and syncing has resumed.", displayName),
		nil); err != nil {
		return err
	}

	return r.ResetReconcileLoops(systemCtx, installation)
}

// MarkOperationUnhealthy records an operation as failing and stops its recurring loop
func (r *Runtime) MarkOperationUnhealthy(ctx context.Context, installation *ent.Integration, operationName, reason string) error {
	health := installation.Health
	if _, recorded := health.UnhealthyOperations[operationName]; recorded {
		return nil
	}

	def, ok := r.Registry().Definition(installation.DefinitionID)
	if !ok {
		return nil
	}

	unhealthy := lo.Assign(health.UnhealthyOperations, map[string]string{operationName: reason})
	health.UnhealthyOperations = unhealthy

	healthyRemain := lo.SomeBy(workloadOperations(def, installation), func(op types.OperationRegistration) bool {
		_, failing := unhealthy[op.Name]

		return !failing
	})

	if !healthyRemain {
		installation.Health = health

		return r.MarkIntegrationUnhealthy(ctx, installation, reason)
	}

	transitioned, err := r.transitionStatus(ctx, installation, integration.StatusIn(enums.IntegrationOperationalStatuses...), enums.IntegrationStatusDegraded, health, false)
	if err != nil || !transitioned {
		return err
	}

	if _, err := r.purgeReconcileLoop(ctx, installation.ID, operationName); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str(notificationFieldOperation, operationName).Msg("failed purging unhealthy operation jobs")
	}

	displayName := r.integrationDisplayName(installation)

	logx.FromContext(ctx).Warn().Str(notificationFieldOperation, operationName).Str(notificationFieldReason, reason).Msg("integration operation marked unhealthy, its recurring loop will stop")

	return r.notifyHealthChange(privacy.DecisionContext(ctx, privacy.Allow), installation, integrationDegradedObjectType,
		fmt.Sprintf("%s is partially working", displayName),
		fmt.Sprintf("The %s integration's %s operation has stopped: %s. Other operations continue to run.", displayName, operationName, reason),
		map[string]any{notificationFieldOperation: operationName, notificationFieldReason: reason})
}

// ClearOperationUnhealthy clears an operation's failure record and resumes its recurring loop
func (r *Runtime) ClearOperationUnhealthy(ctx context.Context, installation *ent.Integration, operationName string) error {
	health := installation.Health
	if _, recorded := health.UnhealthyOperations[operationName]; !recorded {
		return nil
	}

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	health.UnhealthyOperations = lo.OmitByKeys(health.UnhealthyOperations, []string{operationName})
	if len(health.UnhealthyOperations) == 0 {
		health.UnhealthyOperations = nil
	}

	installation.Health = health

	if health.UnhealthyOperations != nil {
		if err := r.DB().Integration.UpdateOneID(installation.ID).SetHealth(health).Exec(systemCtx); err != nil {
			return err
		}

		logx.FromContext(ctx).Info().Str(notificationFieldOperation, operationName).Msg("operation recovered, its recurring loop resumes")

		return r.ResetReconcileLoops(ctx, installation)
	}

	transitioned, err := r.transitionStatus(ctx, installation, integration.StatusEQ(enums.IntegrationStatusDegraded), enums.IntegrationStatusConnected, health, false)
	if err != nil {
		return err
	}

	if !transitioned {
		return r.DB().Integration.UpdateOneID(installation.ID).SetHealth(health).Exec(systemCtx)
	}

	logx.FromContext(ctx).Info().Str(notificationFieldOperation, operationName).Msg("all operations recovered, integration returns to connected")

	displayName := r.integrationDisplayName(installation)

	return r.notifyHealthChange(systemCtx, installation, integrationHealthyObjectType,
		fmt.Sprintf("%s is fully operational", displayName),
		fmt.Sprintf("The %s integration's operations all recovered and syncing has resumed.", displayName),
		nil)
}

// RunHealthAssessment runs the connection and operation health checks and returns the assessment
func (r *Runtime) RunHealthAssessment(ctx context.Context, installation *ent.Integration) (HealthAssessment, error) {
	def, err := r.resolveDefinitionForInstallation(installation)
	if err != nil {
		return HealthAssessment{}, err
	}

	if err := r.ensureCurrentVersion(ctx, installation); err != nil {
		return HealthAssessment{}, err
	}

	checkErr, err := r.verifyConnection(ctx, installation, def)
	if err != nil {
		return HealthAssessment{}, err
	}

	if checkErr != nil {
		if markErr := r.MarkIntegrationUnhealthy(ctx, installation, checkErr.Error()); markErr != nil {
			logx.FromContext(ctx).Error().Err(markErr).Msg("failed marking integration unhealthy after failed health check")
		}

		return HealthAssessment{
			Status:     enums.IntegrationStatusErrored,
			Connection: ConnectionHealthResult{Reason: checkErr.Error()},
			Operations: appendRecordedResults(nil, installation),
		}, nil
	}

	if err := r.RefreshInstallationMetadata(ctx, installation); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("health assessment: instance id refresh failed")
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

// verifyConnection runs the persisted connection's health check under stored credentials; checkErr is the check's own failure, err a resolution failure
func (r *Runtime) verifyConnection(ctx context.Context, installation *ent.Integration, def types.Definition) (checkErr, err error) {
	if len(def.Connections) == 0 || def.HealthCheck == nil {
		return nil, nil
	}

	connection, err := r.resolvePersistedConnection(def, installation)
	if err != nil {
		return nil, err
	}

	bindings, err := r.loadCredentials(privacy.DecisionContext(ctx, privacy.Allow), installation, connection.CredentialRefs)
	if err != nil {
		return nil, err
	}

	return r.runConnectionHealthCheck(ctx, installation, def.HealthCheck, bindings), nil
}

// assessOperationHealth probes workload operations and records each outcome
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
				logx.FromContext(ctx).Error().Err(clearErr).Str(notificationFieldOperation, op.Name).Msg("failed clearing recovered operation")
			}

			continue
		}

		reason := err.Error()
		if degraded, ok := types.DegradedFrom(err); ok {
			reason = degraded.Reason
		}

		results = append(results, OperationHealthResult{Name: op.Name, Healthy: false, Reason: reason})

		if markErr := r.MarkOperationUnhealthy(ctx, installation, op.Name, reason); markErr != nil {
			logx.FromContext(ctx).Error().Err(markErr).Str(notificationFieldOperation, op.Name).Msg("failed marking operation unhealthy")
		}
	}

	return results
}

// runConnectionHealthCheck runs the definition health check against the connection's credentials
func (r *Runtime) runConnectionHealthCheck(ctx context.Context, installation *ent.Integration, check *types.HealthCheckRegistration, bindings types.CredentialBindings) error {
	var client any

	if check.ClientRef.Valid() {
		registration, err := r.Registry().Client(installation.DefinitionID, check.ClientRef)
		if err != nil {
			return err
		}

		client, err = r.keystore().BuildClient(ctx, installation, registration, bindings, nil, false)
		if err != nil {
			return err
		}
	}

	_, err := check.Handle(ctx, types.OperationRequest{
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
	client, credentials, err := r.resolveOperationClient(privacy.DecisionContext(ctx, privacy.Allow), installation, operation, nil, nil, false)
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

// workloadOperations returns operations that do work, excluding internal or disabled ones
func workloadOperations(def types.Definition, installation *ent.Integration) []types.OperationRegistration {
	return lo.Filter(def.Operations, func(op types.OperationRegistration, _ int) bool {
		return !op.Internal && !op.DisabledFor(installation.OperationConfig.For(op.Name))
	})
}

// integrationDisplayName resolves the definition's display name, or the installation's own name
func (r *Runtime) integrationDisplayName(installation *ent.Integration) string {
	if def, ok := r.Registry().Definition(installation.DefinitionID); ok {
		return def.DisplayName
	}

	return installation.Name
}
