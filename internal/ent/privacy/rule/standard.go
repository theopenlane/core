package rule

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/standard"
)

func SystemOwnedStandards() privacy.StandardMutationRuleFunc {
	return privacy.StandardMutationRuleFunc(func(ctx context.Context, m *generated.StandardMutation) error {
		systemOwned, ok := m.SystemOwned()

		if !ok {
			switch m.Op() {
			case ent.OpCreate:
				// on create check if system owned is being set, if not continue
				return privacy.Skipf("no system owned field set")
			default:
				// on update, update one, delete, delete one, always check
				// to ensure the system owned field is set
				ids, err := m.IDs(ctx)
				if err != nil {
					return err
				}

				standards, err := m.Client().Standard.Query().Where(standard.IDIn(ids...)).Select("system_owned").All(ctx)
				if err != nil {
					return err
				}

				// if we have one system owned standard, set to true and continue
				for _, s := range standards {
					if s.SystemOwned {
						systemOwned = true
						break
					}
				}
			}
		}

		allowAdmin, err := CheckIsSystemAdmin(ctx, m)
		if err != nil {
			return err
		}

		if allowAdmin {
			return privacy.Allow
		}

		if systemOwned && !allowAdmin {
			return privacy.Denyf("user is not a system admin")
		}

		return privacy.Skip
	})
}
