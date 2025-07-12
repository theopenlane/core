package rule

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/subprocessor"
)

func SystemOwnedSubprocessor() privacy.SubprocessorMutationRuleFunc {
	return privacy.SubprocessorMutationRuleFunc(func(ctx context.Context, m *generated.SubprocessorMutation) error {
		systemOwned, _ := m.SystemOwned()

		allowAdmin, err := CheckIsSystemAdminWithContext(ctx)
		if err != nil {
			return err
		}

		if allowAdmin {
			return privacy.Allow
		}

		if !systemOwned {
			switch m.Op() {
			case ent.OpCreate:
				// on create check if system owned is being set, if not continue
				return privacy.Skipf("no system owned field set")
			default:
				ids, err := m.IDs(ctx)
				if err != nil {
					return err
				}
				subprocessors, err := m.Client().Subprocessor.Query().Where(subprocessor.IDIn(ids...)).Select(subprocessor.FieldSystemOwned).All(ctx)
				if err != nil {
					return err
				}

				for _, s := range subprocessors {
					if s.SystemOwned {
						systemOwned = true
						break
					}
				}
			}
		}
		if systemOwned && !allowAdmin {
			log.Debug().Msg("user attempted to set system owned field without being a system admin")
			return ErrAdminOnlyField
		}

		return privacy.Skip
	})
}
