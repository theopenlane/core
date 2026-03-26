package scim

import (
	"context"
	"fmt"
	"time"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	scimoptional "github.com/elimity-com/scim/optional"

	"github.com/theopenlane/core/internal/ent/generated"
)

// lookupByID queries a single entity by ID, mapping ent not-found to a SCIM resource-not-found error
func lookupByID[T any](ctx context.Context, id string, query func(context.Context) (T, error)) (T, error) {
	record, err := query(ctx)
	if err != nil {
		var zero T

		if generated.IsNotFound(err) {
			return zero, scimerrors.ScimErrorResourceNotFound(id)
		}

		return zero, fmt.Errorf("failed to get resource: %w", err)
	}

	return record, nil
}

// buildSCIMResource constructs a scim.Resource from common entity fields and pre-built attributes
func buildSCIMResource(id, externalID string, createdAt, updatedAt time.Time, attrs scim.ResourceAttributes) scim.Resource {
	delete(attrs, "externalId")

	extID := scimoptional.NewString("")
	if externalID != "" {
		extID = scimoptional.NewString(externalID)
	}

	return scim.Resource{
		ID:         id,
		ExternalID: extID,
		Attributes: attrs,
		Meta: scim.Meta{
			Created:      &createdAt,
			LastModified: &updatedAt,
			Version:      fmt.Sprintf("W/\"%d\"", updatedAt.Unix()),
		},
	}
}
