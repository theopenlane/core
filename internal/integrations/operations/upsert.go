package operations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/internal/ent/entityops"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/jsonx"
)

// legacyScientificKey converts a numeric key like "147884153" into the "1.47884153e+08" form the
// old CEL double conversion stored, so we can still find rows written before the fix; returns false
// for non-numeric keys and for numbers small enough that the two forms are identical
func legacyScientificKey(value string) (string, bool) {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "", false
	}

	legacy := strconv.FormatFloat(parsed, 'g', -1, 64)
	if legacy == value {
		return "", false
	}

	return legacy, true
}

// findWithLegacyKeyAdoption runs the upsert lookup for one external id, retrying with the legacy
// scientific notation form when the canonical form misses; a row found under the legacy key gets
// its key repaired via the repair callback before it is returned, so the update path sees a row
// that already carries the canonical key
func findWithLegacyKeyAdoption[T any](ctx context.Context, externalID string, find func(context.Context, string) (T, error), repair func(context.Context, T) error) (T, error) {
	existing, err := find(ctx, externalID)
	if err == nil || !ent.IsNotFound(err) {
		return existing, err
	}

	legacy, ok := legacyScientificKey(externalID)
	if !ok {
		return existing, err
	}

	adopted, legacyErr := find(ctx, legacy)
	switch {
	case ent.IsNotFound(legacyErr):
		return existing, err
	case legacyErr != nil:
		return existing, legacyErr
	}

	if repairErr := repair(ctx, adopted); repairErr != nil {
		return existing, repairErr
	}

	return adopted, nil
}

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
