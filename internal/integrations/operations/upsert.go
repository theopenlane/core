package operations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/internal/ent/entityops"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/jsonx"
)

// persistCatalogUpsert marshals one prepared create input and persists it through the schema's
// catalog-driven entityops upsert, mapping the entityops sentinels onto the ingest error classes.
// Schemas whose lookup is not a single org-scoped key column keep persistRoundTripUpsert instead
func persistCatalogUpsert(ctx context.Context, db *ent.Client, schema *entityops.Schema, ownerID string, createInput any) (string, error) {
	payload, err := json.Marshal(createInput)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrIngestMappedDocumentInvalid, err)
	}

	id, err := schema.Upsert(ctx, db, ownerID, payload)
	switch {
	case err == nil:
		return id, nil
	case errors.Is(err, entityops.ErrUpsertKeyMissing):
		return "", ErrIngestUpsertKeyMissing
	case errors.Is(err, entityops.ErrUpsertConflict):
		return "", fmt.Errorf("%w: %w", ErrIngestUpsertConflict, err)
	default:
		return "", wrapIngestPersistError(err)
	}
}

// roundTripUpdateInput converts one create input into its matching update input using JSON round-tripping
func roundTripUpdateInput[Create any, Update any](createInput Create) (Update, error) {
	var updateInput Update
	if err := jsonx.RoundTrip(createInput, &updateInput); err != nil {
		log.Error().Err(err).Msg("integration: invalid ingest")

		return updateInput, ErrIngestMappedDocumentInvalid
	}

	return updateInput, nil
}

// persistUpsert centralizes the common ingest upsert flow while allowing schema-specific lookup and mutation logic
// the function input signature is ugly and hard to read but the call sites are much cleaner
func persistUpsert[Create any, Update any, Existing any](ctx context.Context, createInput Create, toUpdate func(Create) (Update, error), findExisting func(context.Context) (Existing, error), create func(context.Context, Create) (string, error), update func(context.Context, Existing, Update) error, existingID func(Existing) string) (string, error) {
	existing, err := findExisting(ctx)
	switch {
	case err == nil:
		// update existing record
	case ent.IsNotFound(err):
		id, createErr := create(ctx, createInput)
		return id, wrapIngestPersistError(createErr)
	default:
		return "", wrapIngestPersistError(err)
	}

	updateInput, err := toUpdate(createInput)
	if err != nil {
		return "", err
	}

	return existingID(existing), wrapIngestPersistError(update(ctx, existing, updateInput))
}

// persistRoundTripUpsert centralizes the common ingest upsert flow for schemas whose update input can be derived by round-tripping the create input
func persistRoundTripUpsert[Create any, Update any, Existing any](ctx context.Context, createInput Create, findExisting func(context.Context) (Existing, error), create func(context.Context, Create) (string, error), update func(context.Context, Existing, Update) error, existingID func(Existing) string) (string, error) {
	return persistUpsert(ctx, createInput, roundTripUpdateInput, findExisting, create, update, existingID)
}
