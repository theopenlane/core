package operations

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/entity"
	"github.com/theopenlane/core/v2/internal/ent/generated/entitytype"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// vendorEntityTypeName is the EntityType.Name that marks an Entity as a vendor
const vendorEntityTypeName = "vendor"

// VendorMatchCandidates builds the catalogue match candidates for a vendor in priority order: domain, name, display name, alias
func VendorMatchCandidates(name, domain string) []entityops.MatchCandidate {
	return []entityops.MatchCandidate{
		{Field: entity.FieldDomains, Value: domain},
		{Field: entity.FieldName, Value: name},
		{Field: entity.FieldDisplayName, Value: name},
		{Field: entity.FieldAliases, Value: name},
	}
}

// AdoptVendor returns the organization's adopted copy of the catalogue vendor, creating it typed as the organization's vendor entity type when absent
func AdoptVendor(ctx context.Context, client *generated.Client, ownerID, catalogID string) (id string, created bool, err error) {
	entityTypeID, err := client.EntityType.Query().
		Where(
			entitytype.NameEqualFold(vendorEntityTypeName),
			entitytype.OwnerID(ownerID),
		).
		OnlyID(ctx)

	switch {
	case generated.IsNotFound(err):
		return "", false, ErrVendorEntityTypeMissing
	case err != nil:
		return "", false, err
	}

	overlay, err := jsonx.ToRawMessage(map[string]any{
		"entity_type_id":   entityTypeID,
		"approved_for_use": true,
	})
	if err != nil {
		return "", false, err
	}

	return entityops.SchemaEntity.Adopt(ctx, client, catalogID, ownerID, overlay)
}
