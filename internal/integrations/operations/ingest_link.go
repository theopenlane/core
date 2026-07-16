package operations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/entityops"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// injectLinks resolves the mapping's cross-object link rules and writes the matched target ids into
// the mapped create-input payload under each edge's create-input key, so the record is created (or
// emitted for async creation) with its edges already set rather than linked in a post-persist step.
// rules are the effective link rules (installation override or definition default); payload is the
// mapped create-input JSON, re-keyed into the snake_case source context the rule matches read
func injectLinks(ctx context.Context, db *ent.Client, ownerID string, rules []types.LinkRule, schemaName string, payload json.RawMessage) (json.RawMessage, error) {
	if len(rules) == 0 {
		return payload, nil
	}

	sourceSchema, ok := entityops.LookupSchema(schemaName)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrLinkTargetSchemaNotFound, schemaName)
	}

	source := payload
	if sourceSchema.SourceContext != nil {
		source = sourceSchema.SourceContext(payload)
	}

	// translate each provider link rule (keyed by target schema) into a shared LinkSpec keyed by the
	// resolved edge name; the source entity's own mapped fields back the match via SourceContext
	links := make([]entityops.LinkSpec, 0, len(rules))

	for _, rule := range rules {
		edge, found := lo.Find(sourceSchema.Edges, func(e entityops.EdgeDescriptor) bool {
			return e.TargetType == rule.TargetSchema
		})
		if !found {
			return nil, fmt.Errorf("%w: %s has no edge to %s", ErrLinkEdgeNotFound, schemaName, rule.TargetSchema)
		}

		selector := entityops.TargetSelector{SourceContext: source}

		if rule.TargetField != "" {
			selector.KeyMatch = &entityops.KeyMatch{
				TargetField: rule.TargetField,
				SourceField: rule.SourceField,
				SourceList:  rule.SourceList,
			}
		} else {
			selector.Expression = rule.Expression
		}

		links = append(links, entityops.LinkSpec{Edge: edge.Name, Target: selector})
	}

	payload, err := entityops.InjectCreateLinks(ctx, db, ownerID, sourceSchema, payload, links)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("schema", schemaName).Msg("ingest link target resolution failed")

		return nil, fmt.Errorf("%w: %w", ErrLinkFailed, err)
	}

	return payload, nil
}
