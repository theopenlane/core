package operations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
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

// persistCatalogUpsert marshals one prepared create input and persists it through the schema's
// catalog-driven entityops upsert, mapping the entityops sentinels onto the ingest error classes;
// changed reports whether the upsert wrote the record rather than skipping it as unchanged.
// Schemas whose lookup is not a single org-scoped key column keep persistRoundTripUpsert instead
func persistCatalogUpsert(ctx context.Context, db *ent.Client, schema *entityops.Schema, ownerID string, createInput any) (string, bool, error) {
	payload, err := json.Marshal(createInput)
	if err != nil {
		return "", false, fmt.Errorf("%w: %w", ErrIngestMappedDocumentInvalid, err)
	}

	ctx = entityops.WithIngestNoopMark(ctx)

	id, err := schema.Upsert(ctx, db, ownerID, payload)
	switch {
	case err == nil:
		changed := !entityops.IngestNoopMarked(ctx)
		if changed {
			recordIngestChange(ctx)
		}

		return id, changed, nil
	case errors.Is(err, entityops.ErrUpsertKeyMissing):
		return "", false, ErrIngestUpsertKeyMissing
	case errors.Is(err, entityops.ErrUpsertConflict):
		// a row exists that the lookup key could not see, so log both the key and the record identity
		lookupField, _ := schema.LookupField()
		doc, _ := jsonx.Decode[map[string]any](payload)

		logx.FromContext(ctx).Error().Err(err).Str(entityops.FieldSchema, schema.Snake).Str("lookup_field", lookupField.Name).Interface("lookup_value", doc[lookupField.InputKey]).Interface("record_name", doc["name"]).Msg("ingest upsert conflict: lookup key found no existing record but the insert violated a unique constraint")

		return "", false, fmt.Errorf("%w: %w", ErrIngestUpsertConflict, err)
	default:
		return "", false, wrapIngestPersistError(err)
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
		if createErr != nil {
			logx.FromContext(ctx).Error().Err(createErr).Msg("ingest upsert create failed")

			return id, wrapIngestPersistError(createErr)
		}

		recordIngestChange(ctx)

		return id, nil
	default:
		return "", wrapIngestPersistError(err)
	}

	updateInput, err := toUpdate(createInput)
	if err != nil {
		return "", err
	}

	if err := update(ctx, existing, updateInput); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("ingest upsert update failed")

		return existingID(existing), wrapIngestPersistError(err)
	}

	recordIngestChange(ctx)

	return existingID(existing), nil
}

// persistRoundTripUpsert centralizes the common ingest upsert flow for schemas whose update input can be derived by round-tripping the create input
func persistRoundTripUpsert[Create any, Update any, Existing any](ctx context.Context, createInput Create, findExisting func(context.Context) (Existing, error), create func(context.Context, Create) (string, error), update func(context.Context, Existing, Update) error, existingID func(Existing) string) (string, error) {
	return persistUpsert(ctx, createInput, roundTripUpdateInput, findExisting, create, update, existingID)
}
