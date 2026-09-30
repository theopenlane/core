package operations

import (
	"fmt"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// linkSpecs resolves one mapping variant's cross-object link rules into entityops link specs, shared
// by entityops.PrefetchLinkTargets and entityops.InjectCreateLinks across every record in the variant
// group so edge resolution runs once per group instead of once per record
func linkSpecs(sourceSchema *entityops.Schema, rules []types.LinkRule) ([]entityops.LinkSpec, error) {
	specs := make([]entityops.LinkSpec, 0, len(rules))

	for _, rule := range rules {
		edge, err := registry.ResolveLinkEdge(sourceSchema, rule)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrLinkFailed, err)
		}

		selector := entityops.TargetSelector{}

		if rule.TargetField != "" {
			selector.KeyMatch = &entityops.KeyMatch{
				TargetField: rule.TargetField,
				SourceField: rule.SourceField,
				SourceList:  rule.SourceList,
			}
		} else {
			selector.Expression = rule.Expression
		}

		specs = append(specs, entityops.LinkSpec{Edge: edge.Name, Target: selector})
	}

	return specs, nil
}
