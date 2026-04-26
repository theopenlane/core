package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/common/enums"
)

type severityScoreMutation interface {
	Score() (float64, bool)
	SetSecurityLevel(enums.SecurityLevel)
}

type severityMutation interface {
	Severity() (string, bool)
	SetSecurityLevel(enums.SecurityLevel)
}

// HookSeverityLevel sets the security_level based on the score field using CVSS v4.0 ranges
func HookSeverityLevel() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mut, ok := m.(severityScoreMutation)
			if ok {
				score, ok := mut.Score()
				if ok && score > 0 {
					mut.SetSecurityLevel(severityLevelFromScore(score))
					return next.Mutate(ctx, m)
				}
			}

			mutSev, ok := m.(severityMutation)
			if !ok {
				return next.Mutate(ctx, m)
			}

			sev, ok := mutSev.Severity()
			if !ok || sev == "" {
				return next.Mutate(ctx, m)
			}

			level := enums.ToSecurityLevel(sev)
			if level != nil && level != &enums.SecurityLevelInvalid {
				mutSev.SetSecurityLevel(*level)
			}

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
