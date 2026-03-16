package operations

import (
	"context"
	"encoding/json"

	"github.com/samber/lo"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/vulnerability"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// ProcessIngest resolves declared ingest contracts, applies CEL mappings, and upserts normalized records.
func ProcessIngest(ctx context.Context, reg *registry.Registry, db *ent.Client, installation *ent.Integration, operation types.OperationRegistration, response json.RawMessage) error {
	if len(operation.Ingest) == 0 {
		return nil
	}

	definition, ok := reg.Definition(installation.DefinitionID)
	if !ok {
		return ErrIngestDefinitionNotFound
	}

	var payloadSets []types.IngestPayloadSet
	if err := jsonx.RoundTrip(response, &payloadSets); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Str("operation", operation.Name).Msg("failed to decode ingest payloads")
		return ErrIngestPayloadsInvalid
	}

	for _, contract := range operation.Ingest {
		ingestSchema, ok := integrationgenerated.IntegrationIngestSchemas[string(contract.Schema)]
		if !ok {
			return ErrIngestSchemaNotFound
		}

		payloadSet, found := lo.Find(payloadSets, func(item types.IngestPayloadSet) bool {
			return item.Schema == contract.Schema
		})
		if !found {
			continue
		}

		mappingSchema, ok := integrationgenerated.IntegrationMappingSchemas[ingestSchema.Name]
		if !ok {
			return ErrIngestSchemaNotFound
		}

		for _, envelope := range payloadSet.Envelopes {
			mapping, found := findMappingRegistration(definition, contract.Schema, envelope.Variant)
			if !found {
				logx.FromContext(ctx).Error().Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Str("schema", ingestSchema.Name).Str("variant", envelope.Variant).Msg("ingest mapping not found")
				return ErrIngestMappingNotFound
			}

			passed, err := providerkit.EvalFilter(ctx, mapping.Spec.FilterExpr, envelope)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Str("schema", ingestSchema.Name).Str("variant", envelope.Variant).Msg("ingest filter evaluation failed")
				return ErrIngestFilterFailed
			}
			if !passed {
				continue
			}

			mappedRaw, err := providerkit.EvalMap(ctx, mapping.Spec.MapExpr, envelope)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Str("schema", ingestSchema.Name).Str("variant", envelope.Variant).Msg("ingest map evaluation failed")
				return ErrIngestTransformFailed
			}

			mappedDocument, err := jsonx.ToMap(mappedRaw)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Str("schema", ingestSchema.Name).Str("variant", envelope.Variant).Msg("ingest mapped document decode failed")
				return ErrIngestMappedDocumentInvalid
			}

			if contract.EnsurePayloads {
				if _, ok := mappingSchema.AllowedKeys["rawPayload"]; ok {
					if _, exists := mappedDocument["rawPayload"]; !exists {
						mappedDocument["rawPayload"] = jsonx.DecodeAnyOrNil(envelope.Payload)
					}
				}
			}

			if err := validateMappedDocument(mappingSchema, mappedDocument); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Str("schema", ingestSchema.Name).Str("variant", envelope.Variant).Msg("ingest mapped document validation failed")
				return err
			}

			if err := upsertMappedDocument(ctx, db, installation, ingestSchema.Name, mappedDocument); err != nil {
				return err
			}
		}
	}

	return nil
}

func findMappingRegistration(definition types.Definition, schema string, variant string) (types.MappingRegistration, bool) {
	if mapping, ok := lo.Find(definition.Mappings, func(item types.MappingRegistration) bool {
		return item.Schema == schema && item.Variant == variant
	}); ok {
		return mapping, true
	}

	return lo.Find(definition.Mappings, func(item types.MappingRegistration) bool {
		return item.Schema == schema && item.Variant == ""
	})
}

func validateMappedDocument(schema integrationgenerated.IntegrationMappingSchema, mapped map[string]any) error {
	for key := range mapped {
		if _, ok := schema.AllowedKeys[key]; !ok {
			return ErrIngestMappedDocumentInvalid
		}
	}

	for _, key := range schema.RequiredKeys {
		value, ok := mapped[key]
		if !ok || value == nil {
			return ErrIngestRequiredKeyMissing
		}

		if stringValue, ok := value.(string); ok && stringValue == "" {
			return ErrIngestRequiredKeyMissing
		}
	}

	return nil
}

func upsertMappedDocument(ctx context.Context, db *ent.Client, installation *ent.Integration, schemaName string, mapped map[string]any) error {
	switch schemaName {
	case integrationgenerated.IntegrationMappingSchemaVulnerability:
		return upsertVulnerability(ctx, db, installation, mapped)
	default:
		return ErrIngestUnsupportedSchema
	}
}

func upsertVulnerability(ctx context.Context, db *ent.Client, installation *ent.Integration, mapped map[string]any) error {
	createInput, err := integrationgenerated.DecodeInput[ent.CreateVulnerabilityInput](mapped)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Msg("failed to decode vulnerability create input")
		return ErrIngestMappedDocumentInvalid
	}

	updateInput, err := integrationgenerated.DecodeInput[ent.UpdateVulnerabilityInput](mapped)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Msg("failed to decode vulnerability update input")
		return ErrIngestMappedDocumentInvalid
	}

	ownerID := installation.OwnerID
	createInput.OwnerID = &ownerID
	if !lo.Contains(createInput.IntegrationIDs, installation.ID) {
		createInput.IntegrationIDs = append(createInput.IntegrationIDs, installation.ID)
	}

	upsertPredicates := vulnerabilityUpsertPredicates(mapped)
	if len(upsertPredicates) == 0 {
		return ErrIngestUpsertKeyMissing
	}

	record, err := db.Vulnerability.Query().Where(vulnerability.OwnerIDEQ(installation.OwnerID), vulnerability.Or(upsertPredicates...)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			if _, err := db.Vulnerability.Create().SetInput(createInput).Save(ctx); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Msg("failed to create vulnerability from ingest")
				return ErrIngestPersistFailed
			}

			return nil
		}

		if ent.IsNotSingular(err) {
			logx.FromContext(ctx).Error().Err(err).Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Msg("vulnerability ingest upsert keys matched multiple records")
			return ErrIngestUpsertConflict
		}

		logx.FromContext(ctx).Error().Err(err).Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Msg("failed to load vulnerability for ingest upsert")
		return ErrIngestPersistFailed
	}

	linked, err := db.Vulnerability.Query().Where(vulnerability.IDEQ(record.ID), vulnerability.HasIntegrationsWith(integration.IDEQ(installation.ID))).Exist(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Str("vulnerability_id", record.ID).Msg("failed to load vulnerability integration linkage")
		return ErrIngestPersistFailed
	}

	if !linked && !lo.Contains(updateInput.AddIntegrationIDs, installation.ID) {
		updateInput.AddIntegrationIDs = append(updateInput.AddIntegrationIDs, installation.ID)
	}

	if err := db.Vulnerability.UpdateOneID(record.ID).SetInput(updateInput).Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Str("vulnerability_id", record.ID).Msg("failed to update vulnerability from ingest")
		return ErrIngestPersistFailed
	}

	return nil
}

func vulnerabilityUpsertPredicates(mapped map[string]any) []predicate.Vulnerability {
	predicates := make([]predicate.Vulnerability, 0, 2)

	externalID, _ := mapped[integrationgenerated.IntegrationMappingVulnerabilityExternalID].(string)
	if externalID != "" {
		predicates = append(predicates, vulnerability.ExternalIDEQ(externalID))
	}

	cveID, _ := mapped[integrationgenerated.IntegrationMappingVulnerabilityCveID].(string)
	if cveID != "" {
		predicates = append(predicates, vulnerability.CveIDEQ(cveID))
	}

	return predicates
}
