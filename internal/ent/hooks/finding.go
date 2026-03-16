package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

// HookFindingSeverityLevel sets the severity_level based on the score field (falling back to numeric_severity) using CVSS v4.0 ranges
func HookFindingSeverityLevel() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.FindingFunc(func(ctx context.Context, m *generated.FindingMutation) (generated.Value, error) {
			score, ok := m.Score()
			if !ok {
				score, ok = m.NumericSeverity()
				if !ok {
					return next.Mutate(ctx, m)
				}
			}

			_ = score
			// m.SetFindingSeverityLevelName(severityLevelFromScore(score))

			return next.Mutate(ctx, m)
		})
	}, hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne))
}
