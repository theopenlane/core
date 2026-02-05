package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	integrationtypes "github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/vulnerability"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
)

// VulnerabilityIngestRequest defines the inputs required for vulnerability ingestion
// It expects alert envelopes produced by integration operations
type VulnerabilityIngestRequest struct {
	OrgID             string
	IntegrationID     string
	Provider          integrationtypes.ProviderType
	Operation         integrationtypes.OperationName
	IntegrationConfig openapi.IntegrationConfig
	ProviderState     any
	OperationConfig   map[string]any
	Envelopes         []integrationtypes.AlertEnvelope
	DB                *generated.Client
}

// VulnerabilityIngestSummary reports mapping and persistence stats
type VulnerabilityIngestSummary struct {
	Total     int
	Mapped    int
	Persisted int
	Skipped   int
	Failed    int
	Created   int
	Updated   int
}

// VulnerabilityIngestResult captures the results of an ingestion run
type VulnerabilityIngestResult struct {
	Summary VulnerabilityIngestSummary
	Errors  []string
}

// SupportsVulnerabilityIngest reports whether default or configured mappings exist
func SupportsVulnerabilityIngest(provider integrationtypes.ProviderType, config openapi.IntegrationConfig) bool {
	if supportsDefaultMapping(provider, mappingSchemaVulnerability) {
		return true
	}

	overrides := mappingOverrideLookup(config)
	providerKey := normalizeMappingKey(string(provider))
	schemaKey := normalizeMappingKey(mappingSchemaVulnerability)
	providerPrefix := ""
	if providerKey != "" {
		providerPrefix = fmt.Sprintf("%s:%s:", providerKey, schemaKey)
	}
	schemaPrefix := ""
	if schemaKey != "" {
		schemaPrefix = fmt.Sprintf("%s:", schemaKey)
	}

	for key := range overrides {
		if key == schemaKey {
			return true
		}
		if providerKey != "" && key == fmt.Sprintf("%s:%s", providerKey, schemaKey) {
			return true
		}
		if schemaPrefix != "" && strings.HasPrefix(key, schemaPrefix) {
			return true
		}
		if providerPrefix != "" && strings.HasPrefix(key, providerPrefix) {
			return true
		}
	}

	return false
}

// IngestVulnerabilityAlerts maps provider alerts into vulnerability inputs and persists them
func IngestVulnerabilityAlerts(ctx context.Context, req VulnerabilityIngestRequest) (VulnerabilityIngestResult, error) {
	result := VulnerabilityIngestResult{}

	if req.DB == nil {
		return result, fmt.Errorf("db client required")
	}

	mapper, err := NewMappingEvaluator()
	if err != nil {
		return result, err
	}

	schema, ok := integrationgenerated.IntegrationMappingSchemas[mappingSchemaVulnerability]
	if !ok {
		return result, fmt.Errorf("mapping schema not found: %s", mappingSchemaVulnerability)
	}

	integrationConfigMap, err := toMap(req.IntegrationConfig)
	if err != nil {
		return result, err
	}
	providerStateMap, err := toMap(req.ProviderState)
	if err != nil {
		return result, err
	}

	storeRaw := true
	if req.IntegrationConfig.RetentionPolicy != nil {
		storeRaw = req.IntegrationConfig.RetentionPolicy.StoreRawPayload
	}

	for _, envelope := range req.Envelopes {
		result.Summary.Total++

		payloadMap, err := decodeAlertPayload(envelope.Payload)
		if err != nil {
			result.Summary.Failed++
			result.Errors = append(result.Errors, err.Error())
			continue
		}

		spec, ok := resolveMappingSpec(req.IntegrationConfig, req.Provider, mappingSchemaVulnerability, envelope.AlertType)
		if !ok {
			result.Summary.Failed++
			result.Errors = append(result.Errors, ErrMappingNotFound.Error())
			continue
		}

		vars := map[string]any{
			"payload":            payloadMap,
			"resource":           envelope.Resource,
			"alert_type":         envelope.AlertType,
			"provider":           string(req.Provider),
			"operation":          string(req.Operation),
			"org_id":             req.OrgID,
			"integration_id":     req.IntegrationID,
			"config":             req.OperationConfig,
			"integration_config": integrationConfigMap,
			"provider_state":     providerStateMap,
		}

		allowed, err := mapper.EvaluateFilter(ctx, spec.FilterExpr, vars)
		if err != nil {
			result.Summary.Failed++
			result.Errors = append(result.Errors, err.Error())
			continue
		}
		if !allowed {
			result.Summary.Skipped++
			continue
		}

		mapped, err := mapper.EvaluateMap(ctx, spec.MapExpr, vars)
		if err != nil {
			result.Summary.Failed++
			result.Errors = append(result.Errors, err.Error())
			continue
		}

		mapped = filterMappingOutput(schema, mapped)
		if err := validateMappingOutput(schema, mapped); err != nil {
			result.Summary.Failed++
			result.Errors = append(result.Errors, err.Error())
			continue
		}

		if !storeRaw {
			delete(mapped, "rawPayload")
		}

		input, err := decodeVulnerabilityInput(mapped)
		if err != nil {
			result.Summary.Failed++
			result.Errors = append(result.Errors, err.Error())
			continue
		}

		if req.OrgID != "" && (input.OwnerID == nil || *input.OwnerID == "") {
			owner := req.OrgID
			input.OwnerID = &owner
		}
		if req.IntegrationID != "" {
			input.IntegrationIDs = appendUnique(input.IntegrationIDs, req.IntegrationID)
		}

		created, err := upsertVulnerability(ctx, req.DB, req.OrgID, req.IntegrationID, input)
		if err != nil {
			result.Summary.Failed++
			result.Errors = append(result.Errors, err.Error())
			continue
		}

		result.Summary.Mapped++
		result.Summary.Persisted++
		if created {
			result.Summary.Created++
		} else {
			result.Summary.Updated++
		}
	}

	return result, nil
}

// decodeAlertPayload converts a raw alert payload into a map
func decodeAlertPayload(payload json.RawMessage) (map[string]any, error) {
	var out map[string]any
	if err := json.Unmarshal(payload, &out); err != nil {
		return nil, err
	}

	return out, nil
}

// decodeVulnerabilityInput converts mapped fields into a create input
func decodeVulnerabilityInput(data map[string]any) (generated.CreateVulnerabilityInput, error) {
	var input generated.CreateVulnerabilityInput
	bytes, err := json.Marshal(data)
	if err != nil {
		return input, err
	}

	if err := json.Unmarshal(bytes, &input); err != nil {
		return input, err
	}

	return input, nil
}

// appendUnique adds a value to a slice only if it is not already present
func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}

	return append(values, value)
}

// upsertVulnerability inserts or updates a vulnerability based on external identifiers
func upsertVulnerability(ctx context.Context, db *generated.Client, orgID string, integrationID string, input generated.CreateVulnerabilityInput) (bool, error) {
	if db == nil {
		return false, fmt.Errorf("db client required")
	}
	if orgID == "" {
		return false, fmt.Errorf("org id required")
	}
	externalID := strings.TrimSpace(input.ExternalID)
	if externalID == "" {
		return false, fmt.Errorf("external id required")
	}
	input.ExternalID = externalID
	if input.CveID != nil {
		trimmed := strings.TrimSpace(*input.CveID)
		if trimmed == "" {
			input.CveID = nil
		} else {
			input.CveID = &trimmed
		}
	}

	existingID, err := findVulnerabilityID(ctx, db, orgID, input.ExternalID, input.CveID)
	if err != nil {
		return false, err
	}

	if existingID != "" {
		input.IntegrationIDs = nil
		update := db.Vulnerability.UpdateOneID(existingID)
		input.Mutate(update.Mutation())
		if integrationID != "" {
			addIntegration := true
			if exists, err := db.Vulnerability.Query().Where(vulnerability.IDEQ(existingID)).QueryIntegrations().Where(integration.IDEQ(integrationID)).Exist(ctx); err == nil {
				addIntegration = !exists
			}
			if addIntegration {
				update.AddIntegrationIDs(integrationID)
			}
		}

		if _, err := update.Save(ctx); err != nil {
			return false, err
		}

		return false, nil
	}

	create := db.Vulnerability.Create().SetInput(input)
	if integrationID != "" {
		create.AddIntegrationIDs(integrationID)
	}

	if _, err := create.Save(ctx); err != nil {
		return false, err
	}

	return true, nil
}

// findVulnerabilityID locates an existing vulnerability by external ID or CVE
func findVulnerabilityID(ctx context.Context, db *generated.Client, orgID string, externalID string, cveID *string) (string, error) {
	externalID = strings.TrimSpace(externalID)
	id, err := db.Vulnerability.Query().
		Where(
			vulnerability.OwnerIDEQ(orgID),
			vulnerability.ExternalIDEQ(externalID),
		).
		OnlyID(ctx)
	if err == nil {
		return id, nil
	}
	if !generated.IsNotFound(err) {
		return "", err
	}

	if cveID == nil || strings.TrimSpace(*cveID) == "" {
		return "", nil
	}

	id, err = db.Vulnerability.Query().
		Where(
			vulnerability.OwnerIDEQ(orgID),
			vulnerability.CveIDEQ(strings.TrimSpace(*cveID)),
		).
		OnlyID(ctx)
	if err == nil {
		return id, nil
	}
	if !generated.IsNotFound(err) {
		return "", err
	}

	return "", nil
}
