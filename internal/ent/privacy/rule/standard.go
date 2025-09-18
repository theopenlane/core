package rule

import (
	"context"

	"entgo.io/ent"

	"github.com/rs/zerolog"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/standard"
)

// SystemOwnedStandards is a privacy rule that checks if the standard is system owned
// and if the user is a system admin
// For Create operations, the rule checks if the system owned field is set to true and denys if it is and the user is not a system admin
// For Update operations, the rule checks if the system owned field is set to true or the standard is system owned
// and denys if it is and the user is not a system admin
func SystemOwnedStandards() privacy.StandardMutationRuleFunc {
	return privacy.StandardMutationRuleFunc(func(ctx context.Context, m *generated.StandardMutation) error {
		systemOwned, _ := m.SystemOwned()
		isPublic, _ := m.IsPublic()
		freeToUse, _ := m.FreeToUse()

		hasAdminField := systemOwned || isPublic || freeToUse

		allowAdmin, err := CheckIsSystemAdminWithContext(ctx)
		if err != nil {
			return err
		}

		if allowAdmin {
			return privacy.Allow
		}

		// if the field was not in the mutation, check the database
		if !hasAdminField {
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

				// we don't need to check the other fields on update, just if the standard is system owned
				standards, err := m.Client().Standard.Query().Where(standard.IDIn(ids...)).Select(standard.FieldSystemOwned).All(ctx)
				if err != nil {
					return err
				}

				// if we have one system owned standard, set to true and continue
				for _, s := range standards {
					if s.SystemOwned {
						hasAdminField = true
						break
					}
				}
			}
		}

		if hasAdminField && !allowAdmin {
			zerolog.Ctx(ctx).Debug().Msg("user attempted to set system owned field without being a system admin")

			return ErrAdminOnlyField
		}

		return privacy.Skip
	})
}
