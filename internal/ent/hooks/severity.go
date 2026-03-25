package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/common/enums"
)

type severityMutation interface {
	Score() (float64, bool)
	SetSecurityLevel(enums.SecurityLevel)
}

// HookSeverityLevel sets the security_level based on the score field using CVSS v4.0 ranges
func HookSeverityLevel() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mut, ok := m.(severityMutation)
			if !ok {
				return next.Mutate(ctx, m)
			}

			score, ok := mut.Score()
			if !ok {
				return next.Mutate(ctx, m)
			}

			mut.SetSecurityLevel(severityLevelFromScore(score))

			return next.Mutate(ctx, m)
		})
	}
}

const (
	cvssScoreCritical = 9.0
	cvssScoreHigh     = 7.0
	cvssScoreMedium   = 4.0
	cvssScoreLow      = 0.1
)

// severityLevelFromScore maps a CVSS v4.0 score to a normalized severity level
func severityLevelFromScore(score float64) enums.SecurityLevel {
	switch {
	case score >= cvssScoreCritical:
		return enums.SecurityLevelCritical
	case score >= cvssScoreHigh:
		return enums.SecurityLevelHigh
	case score >= cvssScoreMedium:
		return enums.SecurityLevelMedium
	case score >= cvssScoreLow:
		return enums.SecurityLevelLow
	default:
		return enums.SecurityLevelNone
	}
}
