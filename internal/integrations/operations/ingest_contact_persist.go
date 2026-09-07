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
func persistContactInput(ctx context.Context, db *ent.Client, integration *ent.Integration, createInput ent.CreateContactInput) (string, error) {
	hasExternalID := createInput.ExternalID != nil && *createInput.ExternalID != ""
	hasEmail := createInput.Email != nil && *createInput.Email != ""

	if !hasExternalID && !hasEmail {
		return "", ErrIngestUpsertKeyMissing
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

	return persistRoundTripUpsert(
		ctx,
		createInput,
		func(existing *ent.Contact, input ent.UpdateContactInput) (bool, error) {
			if !entityops.ContactIngestUnchanged(existing, input) {
				return false, nil
			}

			// contacts adopt across installations by owner-scoped lookup, so an unchanged row's
			// integration linkage still converges through the relink path; a row another
			// definition manages keeps its linkage
			if existing.IntegrationID != integration.ID && (existing.SourceDefinitionID == "" || existing.SourceDefinitionID == integration.DefinitionID) {
				if err := relinkIngestIntegration(ctx, db, entityops.SchemaContact.Snake, existing.ID, integration.ID); err != nil {
					return false, err
				}
			}

			return true, nil
		},
		func(ctx context.Context) (*ent.Contact, error) {
			return q.Only(ctx)
		},
		func(ctx context.Context, input ent.CreateContactInput) (string, error) {
			c, err := db.Contact.Create().SetInput(input).Save(ctx)
			if err != nil {
				return "", err
			}
			return c.ID, nil
		},
		func(ctx context.Context, existing *ent.Contact, input ent.UpdateContactInput) error {
			return db.Contact.UpdateOneID(existing.ID).SetInput(input).Exec(ctx)
		},
		func(c *ent.Contact) string { return c.ID },
	)
}
