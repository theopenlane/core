package operations

import (
	"context"

	"github.com/rs/zerolog/log"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/jsonx"
)

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
func persistUpsert[Create any, Update any, Existing any](ctx context.Context, createInput Create, toUpdate func(Create) (Update, error), findExisting func(context.Context) (Existing, error), create func(context.Context, Create) error, update func(context.Context, Existing, Update) error) error {
	existing, err := findExisting(ctx)
	switch {
	case err == nil:
		// update existing record
	case ent.IsNotFound(err):
		return wrapIngestPersistError(create(ctx, createInput))
	default:
		return wrapIngestPersistError(err)
	}

	updateInput, err := toUpdate(createInput)
	if err != nil {
		return err
	}

	return wrapIngestPersistError(update(ctx, existing, updateInput))
}

// persistRoundTripUpsert centralizes the common ingest upsert flow for schemas whose update input can be derived by round-tripping the create input
func persistRoundTripUpsert[Create any, Update any, Existing any](ctx context.Context, createInput Create, findExisting func(context.Context) (Existing, error), create func(context.Context, Create) error, update func(context.Context, Existing, Update) error) error {
	return persistUpsert(ctx, createInput, roundTripUpdateInput, findExisting, create, update)
}
