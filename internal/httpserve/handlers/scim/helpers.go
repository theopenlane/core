package scim

import (
	"context"
	"fmt"
	"time"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	scimoptional "github.com/elimity-com/scim/optional"

	"github.com/theopenlane/core/internal/ent/generated"
	definitionscim "github.com/theopenlane/core/internal/integrations/definitions/scim"
	"github.com/theopenlane/core/pkg/logx"
)

// lookupByID queries a single entity by ID, mapping ent not-found to a SCIM resource-not-found error
func lookupByID[T any](ctx context.Context, id string, query func(context.Context) (T, error)) (T, error) {
	record, err := query(ctx)
	if err != nil {
		var zero T

		if generated.IsNotFound(err) {
			return zero, scimerrors.ScimErrorResourceNotFound(id)
		}

		logx.FromContext(ctx).Error().Err(err).Str("id", id).Msg("database error looking up resource")

		return zero, scimerrors.ScimErrorInternal
	}

	return record, nil
}

// buildSCIMResource constructs a scim.Resource from common entity fields and pre-built attributes.
// The attrs map should contain only resource-specific attributes (not id, externalId, or meta)
// as the library injects those from the Resource fields during response serialization
func buildSCIMResource(id, externalID string, createdAt, updatedAt time.Time, attrs scim.ResourceAttributes) scim.Resource {
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

// applyPatchValue applies a SCIM PATCH add or replace operation to the attribute map.
// The library has already validated the operation against the composed schema, so the
// path and value types are guaranteed valid
func applyPatchValue(attributes scim.ResourceAttributes, op scim.PatchOperation) {
	if op.Path == nil {
		valueMap, ok := op.Value.(map[string]any)
		if !ok {
			return
		}

		definitionscim.MergeSCIMMap(attributes, valueMap)

		return
	}

	attrName := op.Path.AttributePath.AttributeName
	subAttr := op.Path.AttributePath.SubAttributeName()

	if subAttr != "" {
		child := definitionscim.EnsureSCIMMap(attributes, attrName)
		child[subAttr] = op.Value

		return
	}

	attributes[attrName] = op.Value
}

// removePatchValue applies a SCIM PATCH remove operation to the attribute map.
// The library has already validated the operation against the composed schema
func removePatchValue(attributes scim.ResourceAttributes, op scim.PatchOperation) {
	if op.Path == nil {
		return
	}

	attrName := op.Path.AttributePath.AttributeName
	subAttr := op.Path.AttributePath.SubAttributeName()

	if subAttr != "" {
		if child, ok := attributes[attrName].(map[string]any); ok {
			delete(child, subAttr)
		}

		return
	}

	delete(attributes, attrName)
}
