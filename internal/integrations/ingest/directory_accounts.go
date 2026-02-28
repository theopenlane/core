package ingest

import (
	"context"

	integrationtypes "github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/pkg/jsonx"
)

// DirectoryAccountIngestRequest defines the inputs required for directory account ingestion
// It expects alert envelopes produced by integration operations
type DirectoryAccountIngestRequest struct {
	// OrgID identifies the organization that owns the directory accounts
	OrgID string
	// IntegrationID identifies the integration record
	IntegrationID string
	// Provider identifies the integration provider
	Provider integrationtypes.ProviderType
	// Operation identifies the operation that produced the envelopes
	Operation integrationtypes.OperationName
	// IntegrationConfig supplies integration-level configuration for mapping
	IntegrationConfig openapi.IntegrationConfig
	// ProviderState carries provider-specific state for mapping
	ProviderState any
	// OperationConfig supplies operation-level configuration for mapping
	OperationConfig map[string]any
	// Envelopes holds account payloads to ingest
	Envelopes []integrationtypes.AlertEnvelope
	// DB provides access to the persistence layer
	DB *generated.Client
}

// Validate checks that required fields are present
func (r *DirectoryAccountIngestRequest) Validate() error {
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

// DirectoryAccountIngestSummary reports mapping and persistence stats
type DirectoryAccountIngestSummary struct {
	// Total counts total envelopes processed
	Total int
	// Mapped counts envelopes that produced mapped output
	Mapped int
	// Persisted counts envelopes that were persisted
	Persisted int
	// Skipped counts envelopes filtered out by mapping
	Skipped int
	// Failed counts envelopes that failed mapping or persistence
	Failed int
	// Created counts new directory accounts created
	Created int
	// Updated counts existing directory accounts updated
	Updated int
}

// DirectoryAccountIngestResult captures the results of an ingestion run
type DirectoryAccountIngestResult struct {
	// Summary aggregates ingestion totals
	Summary DirectoryAccountIngestSummary
	// Errors captures per-envelope error messages
	Errors []string
}

type directoryAccountIngestContext struct {
	mapper               *MappingEvaluator
	schema               integrationgenerated.IntegrationMappingSchema
	overrideIndex        mappingOverrideIndex
	integrationConfigMap map[string]any
	providerStateMap     map[string]any
}

// SupportsDirectoryAccountIngest reports whether default or configured mappings exist
func SupportsDirectoryAccountIngest(provider integrationtypes.ProviderType, config openapi.IntegrationConfig) bool {
	if supportsDefaultMapping(provider, mappingSchemaDirectoryAccount) {
		return true
	}

	return newMappingOverrideIndex(config).HasAny(provider, mappingSchemaDirectoryAccount)
}

// newDirectoryAccountIngestContext prepares shared state for ingest runs
func newDirectoryAccountIngestContext(req DirectoryAccountIngestRequest) (directoryAccountIngestContext, error) {
	mapper, err := NewMappingEvaluator()
	if err != nil {
		return directoryAccountIngestContext{}, err
	}

	schema, ok := integrationgenerated.IntegrationMappingSchemas[mappingSchemaDirectoryAccount]
	if !ok {
		return directoryAccountIngestContext{}, ErrMappingSchemaNotFound
	}

	integrationConfigMap, err := jsonx.ToMap(req.IntegrationConfig)
	if err != nil {
		return directoryAccountIngestContext{}, err
	}
	providerStateMap, err := jsonx.ToMap(req.ProviderState)
	if err != nil {
		return directoryAccountIngestContext{}, err
	}

	return directoryAccountIngestContext{
		mapper:               mapper,
		schema:               schema,
		overrideIndex:        newMappingOverrideIndex(req.IntegrationConfig),
		integrationConfigMap: integrationConfigMap,
		providerStateMap:     providerStateMap,
	}, nil
}

// mapEnvelopeToDirectoryAccount applies mapping rules to a single account envelope
func (c *directoryAccountIngestContext) mapEnvelopeToDirectoryAccount(ctx context.Context, req DirectoryAccountIngestRequest, envelope integrationtypes.AlertEnvelope) (map[string]any, bool, error) {
	payloadMap, err := decodeAlertPayload(envelope.Payload)
	if err != nil {
		return nil, false, err
	}

	spec, ok := resolveMappingSpecWithIndex(c.overrideIndex, req.Provider, mappingSchemaDirectoryAccount, envelope.AlertType)
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

// DirectoryAccounts maps provider payloads into directory account inputs and persists them
func DirectoryAccounts(ctx context.Context, req DirectoryAccountIngestRequest) (DirectoryAccountIngestResult, error) {
	result := DirectoryAccountIngestResult{}

	if err := req.Validate(); err != nil {
		return result, err
	}

	ingestCtx, err := newDirectoryAccountIngestContext(req)
	if err != nil {
		return result, err
	}

	for _, envelope := range req.Envelopes {
		result.Summary.Total++

		mapped, allowed, err := ingestCtx.mapEnvelopeToDirectoryAccount(ctx, req, envelope)
		if err != nil {
			result.Summary.Failed++
			result.Errors = append(result.Errors, err.Error())
			continue
		}
		if !allowed {
			result.Summary.Skipped++
			continue
		}

		input, err := decodeDirectoryAccountInput(mapped)
		if err != nil {
			result.Summary.Failed++
			result.Errors = append(result.Errors, err.Error())
			continue
		}

		created, err := upsertDirectoryAccount(ctx, req.DB, req.OrgID, req.IntegrationID, input)
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

// decodeDirectoryAccountInput converts mapped fields into a create input
func decodeDirectoryAccountInput(data map[string]any) (generated.CreateDirectoryAccountInput, error) {
	var input generated.CreateDirectoryAccountInput
	if err := jsonx.RoundTrip(data, &input); err != nil {
		return input, err
	}

	return input, nil
}

// decodeDirectoryAccountUpdateInput converts create input values into an update-safe mutation input
func decodeDirectoryAccountUpdateInput(input generated.CreateDirectoryAccountInput) (generated.UpdateDirectoryAccountInput, error) {
	var updateInput generated.UpdateDirectoryAccountInput
	if err := jsonx.RoundTrip(input, &updateInput); err != nil {
		return generated.UpdateDirectoryAccountInput{}, err
	}

	return updateInput, nil
}

// upsertDirectoryAccount inserts or updates a directory account based on configured upsert keys
func upsertDirectoryAccount(ctx context.Context, db *generated.Client, orgID string, integrationID string, input generated.CreateDirectoryAccountInput) (bool, error) {
	if input.ExternalID == "" {
		return false, ErrExternalIDRequired
	}

	existingID, err := findDirectoryAccountID(ctx, db, orgID, integrationID, input.ExternalID, input.CanonicalEmail)
	if err != nil {
		return false, err
	}

	if existingID != "" {
		updateInput, err := decodeDirectoryAccountUpdateInput(input)
		if err != nil {
			return false, err
		}

		update := db.DirectoryAccount.UpdateOneID(existingID)
		updateInput.Mutate(update.Mutation())
		if err := update.Exec(ctx); err != nil {
			return false, err
		}

		return false, nil
	}

	if input.OwnerID == nil || *input.OwnerID == "" {
		ownerID := orgID
		input.OwnerID = &ownerID
	}
	if input.IntegrationID == nil || *input.IntegrationID == "" {
		integrationIDValue := integrationID
		input.IntegrationID = &integrationIDValue
	}

	if err := db.DirectoryAccount.Create().SetInput(input).Exec(ctx); err != nil {
		return false, err
	}

	return true, nil
}

// findDirectoryAccountID locates an existing directory account by external ID or canonical email
func findDirectoryAccountID(ctx context.Context, db *generated.Client, orgID string, integrationID string, externalID string, canonicalEmail *string) (string, error) {
	queryPredicates := []predicate.DirectoryAccount{
		directoryaccount.ExternalIDEQ(externalID),
	}
	if orgID != "" {
		queryPredicates = append(queryPredicates, directoryaccount.OwnerIDEQ(orgID))
	}
	if integrationID != "" {
		queryPredicates = append(queryPredicates, directoryaccount.IntegrationIDEQ(integrationID))
	}

	id, err := db.DirectoryAccount.Query().
		Where(queryPredicates...).
		Order(generated.Desc(directoryaccount.FieldUpdatedAt)).
		FirstID(ctx)
	if err == nil {
		return id, nil
	}
	if !generated.IsNotFound(err) {
		return "", err
	}

	if canonicalEmail == nil || *canonicalEmail == "" {
		return "", nil
	}

	queryPredicates = []predicate.DirectoryAccount{
		directoryaccount.CanonicalEmailEQ(*canonicalEmail),
	}
	if orgID != "" {
		queryPredicates = append(queryPredicates, directoryaccount.OwnerIDEQ(orgID))
	}
	if integrationID != "" {
		queryPredicates = append(queryPredicates, directoryaccount.IntegrationIDEQ(integrationID))
	}

	id, err = db.DirectoryAccount.Query().
		Where(queryPredicates...).
		Order(generated.Desc(directoryaccount.FieldUpdatedAt)).
		FirstID(ctx)
	if err == nil {
		return id, nil
	}
	if !generated.IsNotFound(err) {
		return "", err
	}

	return "", nil
}
