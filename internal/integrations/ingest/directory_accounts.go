package ingest

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

type directoryAccountIngestContext struct {
	schemaIngestContext
}

// newDirectoryAccountIngestContext prepares shared state for ingest runs
func newDirectoryAccountIngestContext(req IngestRequest) (directoryAccountIngestContext, error) {
	ingestCtx, err := newSchemaIngestContext(req.IntegrationConfig, req.ProviderState, req.MappingIndex, mappingSchemaDirectoryAccount)
	if err != nil {
		return directoryAccountIngestContext{}, err
	}

	return directoryAccountIngestContext{
		schemaIngestContext: ingestCtx,
	}, nil
}

// mapEnvelopeToDirectoryAccount applies mapping rules to a single account envelope
func (c *directoryAccountIngestContext) mapEnvelopeToDirectoryAccount(ctx context.Context, req IngestRequest, envelope integrationtypes.AlertEnvelope) (map[string]any, bool, error) {
	return mapIngestEnvelope(ctx, c.schemaIngestContext, envelopeMappingRequest{
		Provider:        req.Provider,
		Operation:       req.Operation,
		OrgID:           req.OrgID,
		IntegrationID:   req.IntegrationID,
		OperationConfig: req.OperationConfig,
	}, mappingSchemaDirectoryAccount, envelope)
}

// DirectoryAccounts maps provider payloads into directory account inputs and persists them
func DirectoryAccounts(ctx context.Context, req IngestRequest) (IngestResult, error) {
	result := IngestResult{}

	if err := req.Validate(); err != nil {
		return result, err
	}

	ingestCtx, err := newDirectoryAccountIngestContext(req)
	if err != nil {
		return result, err
	}

	summary, errors := processIngestEnvelopes(req.Envelopes, func(envelope integrationtypes.AlertEnvelope) (bool, bool, error) {
		mapped, allowed, err := ingestCtx.mapEnvelopeToDirectoryAccount(ctx, req, envelope)
		if err != nil {
			return false, false, err
		}
		if !allowed {
			return false, false, nil
		}

		input, err := integrationgenerated.DecodeInput[generated.CreateDirectoryAccountInput](mapped)
		if err != nil {
			return false, false, err
		}

		created, err := upsertDirectoryAccount(ctx, req.DB, req.OrgID, req.IntegrationID, input)
		if err != nil {
			return false, false, err
		}

		return true, created, nil
	})

	result.Summary = summary
	result.Errors = errors

	return result, nil
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
