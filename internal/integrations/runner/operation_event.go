package runner

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/integrations/types"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integrationrun"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/ingest"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/pkg/events/soiree"
)

const vulnerabilityCollectOperation types.OperationName = "vulnerabilities.collect"

// OperationEventContext provides dependencies for integration operation listeners
type OperationEventContext struct {
	// DB is the ent client used for persistence
	DB *ent.Client
	// Operations executes provider operations
	Operations *keystore.OperationManager
}

// HandleIntegrationOperationRequested executes integration operations triggered by events
func HandleIntegrationOperationRequested(ctx *soiree.EventContext, payload soiree.IntegrationOperationRequestedPayload) error {
	deps, ok := soiree.ClientAs[*OperationEventContext](ctx)
	if !ok || deps == nil || deps.DB == nil || deps.Operations == nil {
		return fmt.Errorf("integration operation context missing")
	}

	runID := strings.TrimSpace(payload.RunID)
	if runID == "" {
		return fmt.Errorf("integration run id required")
	}

	systemCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)
	run, err := deps.DB.IntegrationRun.Query().
		Where(integrationrun.IDEQ(runID)).
		WithIntegration().
		Only(systemCtx)
	if err != nil {
		return err
	}

	integrationRecord := run.Edges.Integration
	if integrationRecord == nil {
		return fmt.Errorf("integration record missing for run")
	}

	provider := types.ProviderTypeFromString(integrationRecord.Kind)
	if provider == types.ProviderUnknown {
		return fmt.Errorf("integration provider unknown: %s", integrationRecord.Kind)
	}

	operationName := types.OperationName(strings.TrimSpace(run.OperationName))
	if operationName == "" {
		return fmt.Errorf("operation name required for run")
	}

	startedAt := time.Now()
	update := deps.DB.IntegrationRun.UpdateOneID(run.ID).
		SetStatus(enums.IntegrationRunStatusRunning).
		SetStartedAt(startedAt)
	if eventID := soiree.EventID(ctx.Event()); eventID != "" {
		update.SetEventID(eventID)
	}
	if err := update.Exec(systemCtx); err != nil {
		return err
	}

	operationConfig := maps.Clone(run.OperationConfig)
	if operationName == vulnerabilityCollectOperation {
		operationConfig = ensureIncludePayloads(operationConfig)
	}

	result, opErr := deps.Operations.Run(systemCtx, types.OperationRequest{
		OrgID:    run.OwnerID,
		Provider: provider,
		Name:     operationName,
		Config:   operationConfig,
		Force:    payload.Force,
	})

	runStatus := enums.IntegrationRunStatusSuccess
	summary := result.Summary
	errorText := ""
	if opErr != nil || result.Status != types.OperationStatusOK {
		runStatus = enums.IntegrationRunStatusFailed
		if opErr != nil {
			errorText = opErr.Error()
		} else {
			errorText = result.Summary
		}
	}

	metrics := buildOperationMetrics(result)

	if runStatus == enums.IntegrationRunStatusSuccess && operationName == vulnerabilityCollectOperation {
		var ingestErr error
		var ingestResult ingest.VulnerabilityIngestResult

		if ingest.SupportsVulnerabilityIngest(provider, integrationRecord.Config) {
			envelopes, err := extractAlertEnvelopes(result.Details)
			if err != nil {
				ingestErr = err
			} else {
				ingestResult, ingestErr = ingest.IngestVulnerabilityAlerts(systemCtx, ingest.VulnerabilityIngestRequest{
					OrgID:             run.OwnerID,
					IntegrationID:     integrationRecord.ID,
					Provider:          provider,
					Operation:         operationName,
					IntegrationConfig: integrationRecord.Config,
					ProviderState:     integrationRecord.ProviderState,
					OperationConfig:   operationConfig,
					Envelopes:         envelopes,
					DB:                deps.DB,
				})
			}
		} else {
			ingestErr = ingest.ErrMappingNotFound
		}

		metrics = appendIngestMetrics(metrics, ingestResult)
		if ingestErr != nil {
			runStatus = enums.IntegrationRunStatusFailed
			errorText = ingestErr.Error()
		}
	}

	finishedAt := time.Now()
	durationMs := int(finishedAt.Sub(startedAt).Milliseconds())
	finalize := deps.DB.IntegrationRun.UpdateOneID(run.ID).
		SetStatus(runStatus).
		SetFinishedAt(finishedAt).
		SetDurationMs(durationMs).
		SetSummary(summary).
		SetMetrics(metrics)
	if errorText != "" {
		finalize.SetError(errorText)
	}

	if err := finalize.Exec(systemCtx); err != nil {
		return err
	}

	if runStatus != enums.IntegrationRunStatusSuccess {
		return fmt.Errorf("integration operation failed: %s", errorText)
	}

	return nil
}

// ensureIncludePayloads forces include_payloads to true
func ensureIncludePayloads(config map[string]any) map[string]any {
	if config == nil {
		config = map[string]any{}
	}
	config["include_payloads"] = true
	return config
}

// extractAlertEnvelopes pulls alert envelopes from operation details
func extractAlertEnvelopes(details map[string]any) ([]types.AlertEnvelope, error) {
	if details == nil {
		return nil, fmt.Errorf("alert payloads missing")
	}
	raw, ok := details["alerts"]
	if !ok || raw == nil {
		return nil, fmt.Errorf("alert payloads missing")
	}

	switch value := raw.(type) {
	case []types.AlertEnvelope:
		return value, nil
	case []any:
		envelopes := make([]types.AlertEnvelope, 0, len(value))
		for _, item := range value {
			switch typed := item.(type) {
			case types.AlertEnvelope:
				envelopes = append(envelopes, typed)
			case map[string]any:
				envelope, err := decodeAlertEnvelope(typed)
				if err != nil {
					return nil, err
				}
				envelopes = append(envelopes, envelope)
			default:
				envelope, err := decodeAlertEnvelope(item)
				if err != nil {
					return nil, err
				}
				envelopes = append(envelopes, envelope)
			}
		}
		return envelopes, nil
	default:
		envelope, err := decodeAlertEnvelope(value)
		if err != nil {
			return nil, err
		}
		return []types.AlertEnvelope{envelope}, nil
	}
}

// decodeAlertEnvelope coerces an envelope from a dynamic payload
func decodeAlertEnvelope(value any) (types.AlertEnvelope, error) {
	var envelope types.AlertEnvelope
	encoded, err := json.Marshal(value)
	if err != nil {
		return envelope, err
	}
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		return envelope, err
	}
	return envelope, nil
}

// buildOperationMetrics builds a metrics map for operation output
func buildOperationMetrics(result types.OperationResult) map[string]any {
	operation := map[string]any{
		"status":  string(result.Status),
		"summary": result.Summary,
	}
	metrics := map[string]any{
		"operation": operation,
	}
	if result.Details != nil {
		operation["details"] = result.Details
	}
	return metrics
}

// appendIngestMetrics attaches ingest results to metrics
func appendIngestMetrics(metrics map[string]any, result ingest.VulnerabilityIngestResult) map[string]any {
	if metrics == nil {
		metrics = map[string]any{}
	}
	metrics["ingest_summary"] = result.Summary
	if len(result.Errors) > 0 {
		metrics["ingest_errors"] = result.Errors
	}
	return metrics
}
