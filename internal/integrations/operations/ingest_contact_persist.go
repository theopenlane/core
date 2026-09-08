package operations

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/contact"
)

// persistContactInput upserts one Contact record using the ingest lookup key fields
// ExternalID takes priority when present; email is used as the fallback identifier
// Using both fields in a single AND query would produce false negatives when a contact
// exists by email but has not yet had an external_id assigned
func persistContactInput(ctx context.Context, db *ent.Client, integration *ent.Integration, createInput ent.CreateContactInput) (string, bool, bool, error) {
	hasExternalID := createInput.ExternalID != nil && *createInput.ExternalID != ""
	hasEmail := createInput.Email != nil && *createInput.Email != ""

	if !hasExternalID && !hasEmail {
		return "", false, false, ErrIngestUpsertKeyMissing
	}

	if createInput.IntegrationID == nil {
		createInput.IntegrationID = &integration.ID
	}

	q := db.Contact.Query().Where(contact.OwnerID(integration.OwnerID))

	switch {
	case hasExternalID:
		q = q.Where(contact.ExternalID(*createInput.ExternalID))
	default:
		q = q.Where(contact.Email(*createInput.Email))
	}

	return persistLookupUpsert(ctx, db, entityops.SchemaContact, integration, createInput, q.Only)
}
