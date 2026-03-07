package ingest

import (
	"context"
	"encoding/json"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/vulnerability"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

type vulnerabilityIngestContext struct {
	schemaIngestContext
	storeRaw bool
}

// newVulnerabilityIngestContext prepares shared state for ingest runs
func newVulnerabilityIngestContext(req IngestRequest) (vulnerabilityIngestContext, error) {
	ingestCtx, err := newSchemaIngestContext(req.IntegrationConfig, req.ProviderState, req.MappingIndex, mappingSchemaVulnerability)
	if err != nil {
		return vulnerabilityIngestContext{}, err
	}

	storeRaw := true
	if req.IntegrationConfig.RetentionPolicy != nil {
		storeRaw = req.IntegrationConfig.RetentionPolicy.StoreRawPayload
	}

	return vulnerabilityIngestContext{
		schemaIngestContext: ingestCtx,
		storeRaw:            storeRaw,
	}, nil
}

// mapEnvelopeToVulnerability applies mapping rules to a single alert envelope
func (c *vulnerabilityIngestContext) mapEnvelopeToVulnerability(ctx context.Context, req IngestRequest, envelope integrationtypes.AlertEnvelope) (map[string]any, bool, error) {
	return mapIngestEnvelope(ctx, c.schemaIngestContext, envelopeMappingRequest{
		Provider:        req.Provider,
		Operation:       req.Operation,
		OrgID:           req.OrgID,
		IntegrationID:   req.IntegrationID,
		OperationConfig: req.OperationConfig,
	}, mappingSchemaVulnerability, envelope)
}

// VulnerabilityAlerts maps provider alerts into vulnerability inputs and persists them
func VulnerabilityAlerts(ctx context.Context, req IngestRequest) (IngestResult, error) {
	result := IngestResult{}

	if err := req.Validate(); err != nil {
		return result, err
	}

	ingestCtx, err := newVulnerabilityIngestContext(req)
	if err != nil {
		return result, err
	}

	summary, errors := processIngestEnvelopes(req.Envelopes, func(envelope integrationtypes.AlertEnvelope) (bool, bool, error) {
		mapped, allowed, err := ingestCtx.mapEnvelopeToVulnerability(ctx, req, envelope)
		if err != nil {
			return false, false, err
		}
		if !allowed {
			return false, false, nil
		}

		if !ingestCtx.storeRaw {
			delete(mapped, integrationgenerated.IntegrationMappingVulnerabilityRawPayload)
		}

		input, err := integrationgenerated.DecodeInput[generated.CreateVulnerabilityInput](mapped)
		if err != nil {
			return false, false, err
		}

		if input.OwnerID == nil || *input.OwnerID == "" {
			owner := req.OrgID
			input.OwnerID = &owner
		}
		if !lo.Contains(input.IntegrationIDs, req.IntegrationID) {
			input.IntegrationIDs = append(input.IntegrationIDs, req.IntegrationID)
		}

		created, err := upsertVulnerability(ctx, req.DB, req.OrgID, req.IntegrationID, input)
		if err != nil {
			return false, false, err
		}

		return true, created, nil
	})

	result.Summary = summary
	result.Errors = errors

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

// upsertVulnerability inserts or updates a vulnerability based on external identifiers
func upsertVulnerability(ctx context.Context, db *generated.Client, orgID string, integrationID string, input generated.CreateVulnerabilityInput) (bool, error) {
	externalID := input.ExternalID
	if externalID == "" {
		return false, ErrExternalIDRequired
	}

	input.ExternalID = externalID

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

	if cveID == nil || *cveID == "" {
		return "", nil
	}

	id, err = db.Vulnerability.Query().
		Where(
			vulnerability.OwnerIDEQ(orgID),
			vulnerability.CveIDEQ(*cveID),
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
