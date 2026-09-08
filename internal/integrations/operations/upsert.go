package operations

import (
	"bytes"
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

// provenanceDefaults maps each trusted provenance column to its integration-derived value
func provenanceDefaults(integration *ent.Integration) map[string]string {
	return map[string]string{
		"owner_id":                  integration.OwnerID,
		"integration_id":            integration.ID,
		"platform_id":               integration.PlatformID,
		"source_definition_id":      integration.DefinitionID,
		"source_definition_version": integration.DefinitionVersion,
		"source_instance_id":        integration.InstallationMetadata.Display.ExternalID,
	}
}

// stampProvenance fills the trusted integration-derived columns onto a prepared ingest payload,
// leaving any value the payload already carries; provenance and ownership come from the resolved
// integration rather than from provider-mapped overrides
func stampProvenance(payload json.RawMessage, integration *ent.Integration) json.RawMessage {
	if integration == nil {
		return payload
	}

	return jsonx.EditObject(payload, func(doc map[string]json.RawMessage) bool {
		changed := false

		for key, value := range provenanceDefaults(integration) {
			if value == "" {
				continue
			}

			if raw, ok := doc[key]; ok && !jsonx.IsEmptyRawMessage(raw) && !bytes.Equal(bytes.TrimSpace(raw), []byte(`""`)) {
				continue
			}

			encoded, err := json.Marshal(value)
			if err != nil {
				continue
			}

			doc[key] = encoded
			changed = true
		}

		return changed
	})
}

// stampDirectoryProvenance stamps the trusted integration-derived columns onto a typed directory
// create input, since the directory persist paths build and read the input before reaching the
// shared upsert helpers that stamp everything else
func stampDirectoryProvenance[T any](input T, integration *ent.Integration) (T, error) {
	payload, err := json.Marshal(input)
	if err != nil {
		return input, err
	}

	var stamped T
	if err := json.Unmarshal(stampProvenance(payload, integration), &stamped); err != nil {
		return input, err
	}

	return stamped, nil
}

// persistCatalogUpsert marshals one prepared create input, stamps the trusted provenance columns, and
// persists it through the schema's catalog-driven entityops upsert; changed reports whether the
// upsert wrote the record rather than skipping it as unchanged, and managed reports whether this
// installation's definition manages the record
func persistCatalogUpsert(ctx context.Context, db *ent.Client, schema *entityops.Schema, ownerID string, integration *ent.Integration, createInput any) (string, bool, bool, error) {
	payload, err := json.Marshal(createInput)
	if err != nil {
		return "", false, false, fmt.Errorf("%w: %w", ErrIngestMappedDocumentInvalid, err)
	}

	payload = stampProvenance(payload, integration)

	id, changed, managed, err := schema.Upsert(ctx, db, ownerID, payload)
	switch {
	case err == nil:
		if changed {
			recordIngestChange(ctx)
		}

		return id, changed, managed, nil
	case errors.Is(err, entityops.ErrUpsertKeyMissing):
		return "", false, false, ErrIngestUpsertKeyMissing
	case errors.Is(err, entityops.ErrUpsertConflict):
		// a row exists that the lookup key could not see, so log both the key and the record identity
		lookupField, _ := schema.LookupField()
		doc, _ := jsonx.Decode[map[string]any](payload)

		logx.FromContext(ctx).Error().Err(err).Str(entityops.FieldSchema, schema.Snake).Str("lookup_field", lookupField.Name).Interface("lookup_value", doc[lookupField.InputKey]).Interface("record_name", doc["name"]).Msg("ingest upsert conflict: lookup key found no existing record but the insert violated a unique constraint")

		return "", false, false, fmt.Errorf("%w: %w", ErrIngestUpsertConflict, err)
	default:
		return "", false, false, wrapIngestPersistError(err)
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

// persistLookupUpsert persists one prepared create input against a schema-specific lookup, routing a
// found row through the shared ingest upsert decision and an absent row through the catalog create
func persistLookupUpsert[Existing any](ctx context.Context, db *ent.Client, schema *entityops.Schema, integration *ent.Integration, createInput any, find func(context.Context) (Existing, error)) (string, bool, bool, error) {
	payload, err := json.Marshal(createInput)
	if err != nil {
		return "", false, false, fmt.Errorf("%w: %w", ErrIngestMappedDocumentInvalid, err)
	}

	payload = stampProvenance(payload, integration)

	existing, err := find(ctx)
	switch {
	case err == nil:
		row, marshalErr := json.Marshal(existing)
		if marshalErr != nil {
			return "", false, false, fmt.Errorf("%w: %w", ErrIngestPersistFailed, marshalErr)
		}

		id, changed, managed, upsertErr := schema.Upsert(ctx, db, integration.OwnerID, payload, []json.RawMessage{row})
		if upsertErr != nil {
			return "", false, false, wrapIngestPersistError(upsertErr)
		}

		if changed {
			recordIngestChange(ctx)
		}

		return id, changed, managed, nil
	case ent.IsNotFound(err):
		id, createErr := schema.Create(ctx, db, payload)
		if createErr != nil {
			return "", false, false, wrapIngestPersistError(createErr)
		}

		recordIngestChange(ctx)

		return id, true, true, nil
	default:
		return "", false, false, wrapIngestPersistError(err)
	}
}
