package ingest

import (
	"context"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// DirectoryAccounts maps provider payloads into directory account inputs and persists them
func DirectoryAccounts(ctx context.Context, req IngestRequest) (IngestResult, error) {
	result := IngestResult{}

	if err := req.Validate(); err != nil {
		return result, err
	}

	ingestCtx, err := newSchemaIngestContext(req.IntegrationConfig, req.ProviderState, req.MappingIndex, mappingSchemaDirectoryAccount)
	if err != nil {
		return result, err
	}

	summary, errors := processIngestEnvelopes(req.Envelopes, func(envelope integrationtypes.AlertEnvelope) (bool, bool, error) {
		mapped, allowed, err := ingestCtx.mapEnvelope(ctx, req, envelope)
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
		input.OwnerID = lo.ToPtr(orgID)
	}
	if input.IntegrationID == nil || *input.IntegrationID == "" {
		input.IntegrationID = lo.ToPtr(integrationID)
	}

	if err := db.DirectoryAccount.Create().SetInput(input).Exec(ctx); err != nil {
		return false, err
	}

	return true, nil
}

// queryDirectoryAccountID queries for a directory account ID using a base predicate plus optional org and integration filters
func queryDirectoryAccountID(ctx context.Context, db *generated.Client, orgID, integrationID string, base predicate.DirectoryAccount) (string, error) {
	preds := []predicate.DirectoryAccount{base}
	if orgID != "" {
		preds = append(preds, directoryaccount.OwnerIDEQ(orgID))
	}
	if integrationID != "" {
		preds = append(preds, directoryaccount.IntegrationIDEQ(integrationID))
	}

	id, err := db.DirectoryAccount.Query().
		Where(preds...).
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

// findDirectoryAccountID locates an existing directory account by external ID or canonical email
func findDirectoryAccountID(ctx context.Context, db *generated.Client, orgID string, integrationID string, externalID string, canonicalEmail *string) (string, error) {
	if id, err := queryDirectoryAccountID(ctx, db, orgID, integrationID, directoryaccount.ExternalIDEQ(externalID)); id != "" || err != nil {
		return id, err
	}

	if canonicalEmail == nil || *canonicalEmail == "" {
		return "", nil
	}

	return queryDirectoryAccountID(ctx, db, orgID, integrationID, directoryaccount.CanonicalEmailEQ(*canonicalEmail))
}
