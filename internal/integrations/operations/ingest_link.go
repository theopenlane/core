package operations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/core/internal/ent/entityops"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// injectLinks resolves the mapping's cross-object link rules and writes the matched target ids into
// the mapped create-input payload under each edge's create-input key, so the record is created (or
// emitted for async creation) with its edges already set rather than linked in a post-persist step
func injectLinks(ctx context.Context, db *ent.Client, ownerID string, rules []types.LinkRule, sourceSchema *entityops.Schema, payload json.RawMessage) (json.RawMessage, error) {
	if len(rules) == 0 {
		return payload, nil
	}

	links := make([]entityops.LinkSpec, 0, len(rules))

	for _, rule := range rules {
		edge, err := registry.ResolveLinkEdge(sourceSchema, rule)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrLinkFailed, err)
		}

		selector := entityops.TargetSelector{SourceContext: payload}

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
		logx.FromContext(ctx).Error().Err(err).Str("schema", sourceSchema.Name).Msg("ingest link target resolution failed")

		return nil, fmt.Errorf("%w: %w", ErrLinkFailed, err)
	}

	return payload, nil
}
