package ingest

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/samber/lo"

	integrationtypes "github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/vulnerability"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/pkg/jsonx"
)

// VulnerabilityIngestRequest defines the inputs required for vulnerability ingestion
// It expects alert envelopes produced by integration operations
type VulnerabilityIngestRequest struct {
	// OrgID identifies the organization that owns the vulnerabilities
	OrgID string
	// IntegrationID identifies the integration record
	IntegrationID string
	// Provider identifies the integration provider
	Provider integrationtypes.ProviderType
	// Operation identifies the operation that produced the alerts
	Operation integrationtypes.OperationName
	// IntegrationConfig supplies integration-level configuration for mapping
	IntegrationConfig openapi.IntegrationConfig
	// ProviderState carries provider-specific state for mapping
	ProviderState any
	// OperationConfig supplies operation-level configuration for mapping
	OperationConfig map[string]any
	// Envelopes holds the alert payloads to ingest
	Envelopes []integrationtypes.AlertEnvelope
	// DB provides access to the persistence layer
	DB *generated.Client
}

// Validate checks that required fields are present
func (r *VulnerabilityIngestRequest) Validate() error {
	if r.OrgID == "" {
		return ErrIngestOrgIDRequired
	}
	if r.IntegrationID == "" {
		return ErrIngestIntegrationRequired
	}
	if r.DB == nil {
		return ErrDBClientRequired
	}

	return nil
}

// VulnerabilityIngestSummary reports mapping and persistence stats
type VulnerabilityIngestSummary struct {
	// Total counts total alerts processed
	Total int
	// Mapped counts alerts that produced mapped output
	Mapped int
	// Persisted counts alerts that were persisted
	Persisted int
	// Skipped counts alerts filtered out by mapping
	Skipped int
	// Failed counts alerts that failed mapping or persistence
	Failed int
	// Created counts new vulnerabilities created
	Created int
	// Updated counts existing vulnerabilities updated
	Updated int
}

// VulnerabilityIngestResult captures the results of an ingestion run
type VulnerabilityIngestResult struct {
	// Summary aggregates ingestion totals
	Summary VulnerabilityIngestSummary
	// Errors captures per-alert error messages
	Errors []string
}

type vulnerabilityIngestContext struct {
	mapper               *MappingEvaluator
	schema               integrationgenerated.IntegrationMappingSchema
	overrideIndex        mappingOverrideIndex
	integrationConfigMap map[string]any
	providerStateMap     map[string]any
	storeRaw             bool
}

// SupportsVulnerabilityIngest reports whether default or configured mappings exist
func SupportsVulnerabilityIngest(provider integrationtypes.ProviderType, config openapi.IntegrationConfig) bool {
	if supportsDefaultMapping(provider, mappingSchemaVulnerability) {
		return true
	}

	return newMappingOverrideIndex(config).HasAny(provider, mappingSchemaVulnerability)
}

// newVulnerabilityIngestContext prepares shared state for ingest runs
func newVulnerabilityIngestContext(req VulnerabilityIngestRequest) (vulnerabilityIngestContext, error) {
	mapper, err := NewMappingEvaluator()
	if err != nil {
		return vulnerabilityIngestContext{}, err
	}

	schema, ok := integrationgenerated.IntegrationMappingSchemas[mappingSchemaVulnerability]
	if !ok {
		return vulnerabilityIngestContext{}, ErrMappingSchemaNotFound
	}

	integrationConfigMap, err := jsonx.ToMap(req.IntegrationConfig)
	if err != nil {
		return vulnerabilityIngestContext{}, err
	}
	providerStateMap, err := jsonx.ToMap(req.ProviderState)
	if err != nil {
		return vulnerabilityIngestContext{}, err
	}

	storeRaw := true
	if req.IntegrationConfig.RetentionPolicy != nil {
		storeRaw = req.IntegrationConfig.RetentionPolicy.StoreRawPayload
	}

	return vulnerabilityIngestContext{
		mapper:               mapper,
		schema:               schema,
		overrideIndex:        newMappingOverrideIndex(req.IntegrationConfig),
		integrationConfigMap: integrationConfigMap,
		providerStateMap:     providerStateMap,
		storeRaw:             storeRaw,
	}, nil
}

// mapEnvelopeToVulnerability applies mapping rules to a single alert envelope
func (c *vulnerabilityIngestContext) mapEnvelopeToVulnerability(ctx context.Context, req VulnerabilityIngestRequest, envelope integrationtypes.AlertEnvelope) (map[string]any, bool, error) {
	payloadMap, err := decodeAlertPayload(envelope.Payload)
	if err != nil {
		return nil, false, err
	}

	spec, ok := resolveMappingSpecWithIndex(c.overrideIndex, req.Provider, mappingSchemaVulnerability, envelope.AlertType)
	if !ok {
		return nil, false, ErrMappingNotFound
	}

	vars := MappingVars{
		Payload:           payloadMap,
		Resource:          envelope.Resource,
		AlertType:         envelope.AlertType,
		Provider:          req.Provider,
		Operation:         req.Operation,
		OrgID:             req.OrgID,
		IntegrationID:     req.IntegrationID,
		Config:            req.OperationConfig,
		IntegrationConfig: c.integrationConfigMap,
		ProviderState:     c.providerStateMap,
	}.Map()

	allowed, err := c.mapper.EvaluateFilter(ctx, spec.FilterExpr, vars)
	if err != nil {
		return nil, false, err
	}
	if !allowed {
		return nil, false, nil
	}

	mapped, err := c.mapper.EvaluateMap(ctx, spec.MapExpr, vars)
	if err != nil {
		return nil, false, err
	}

	mapped = filterMappingOutput(c.schema, mapped)
	if err := validateMappingOutput(c.schema, mapped); err != nil {
		return nil, false, err
	}

	return mapped, true, nil
}

// VulnerabilityAlerts maps provider alerts into vulnerability inputs and persists them
func VulnerabilityAlerts(ctx context.Context, req VulnerabilityIngestRequest) (VulnerabilityIngestResult, error) {
	result := VulnerabilityIngestResult{}

	if req.DB == nil {
		return result, ErrDBClientRequired
	}

	ingestCtx, err := newVulnerabilityIngestContext(req)
	if err != nil {
		return result, err
	}

	for _, envelope := range req.Envelopes {
		result.Summary.Total++

		mapped, allowed, err := ingestCtx.mapEnvelopeToVulnerability(ctx, req, envelope)
		if err != nil {
			result.Summary.Failed++
			result.Errors = append(result.Errors, err.Error())
			continue
		}
		if !allowed {
			result.Summary.Skipped++
			continue
		}

		if !ingestCtx.storeRaw {
			delete(mapped, integrationgenerated.IntegrationMappingVulnerabilityRawPayload)
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
		if req.IntegrationID != "" && !lo.Contains(input.IntegrationIDs, req.IntegrationID) {
			input.IntegrationIDs = append(input.IntegrationIDs, req.IntegrationID)
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
	if err := jsonx.RoundTrip(data, &input); err != nil {
		return input, err
	}

	return input, nil
}

// upsertVulnerability inserts or updates a vulnerability based on external identifiers
func upsertVulnerability(ctx context.Context, db *generated.Client, orgID string, integrationID string, input generated.CreateVulnerabilityInput) (bool, error) {
	externalID := strings.TrimSpace(input.ExternalID)
	if externalID == "" {
		return false, ErrExternalIDRequired
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
			exists, err := db.Vulnerability.Query().Where(vulnerability.IDEQ(existingID)).QueryIntegrations().Where(integration.IDEQ(integrationID)).Exist(ctx)
			if err != nil {
				return false, err
			}

			if !exists {
				update.AddIntegrationIDs(integrationID)
			}
		}

		if err := update.Exec(ctx); err != nil {
			return false, err
		}

		return false, nil
	}

	create := db.Vulnerability.Create().SetInput(input)
	if integrationID != "" {
		create.AddIntegrationIDs(integrationID)
	}

	if err := create.Exec(ctx); err != nil {
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
