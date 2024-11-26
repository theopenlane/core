package rule

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// CanCreateObjectsInProgram is a rule that returns allow decision if user has edit access in the program(s)
// which allows them to create objects associated with the program
func CanCreateObjectsInProgram() privacy.MutationRuleFunc {
	return privacy.MutationRuleFunc(func(ctx context.Context, m generated.Mutation) error {
		pIDs, err := getProgramIDFromEntMutation(m)
		if err != nil {
			return privacy.Denyf("unable to get program id from mutation, %s", err.Error())
		}

		if len(pIDs) == 0 {
			return privacy.Skipf("no program set on request, skipping")
		}

		log.Debug().Msg("checking mutation access")

		relation := fgax.CanEdit

		userID, err := auth.GetUserIDFromContext(ctx)
		if err != nil {
			return err
		}

		log.Info().Str("relation", relation).
			Strs("program_id", pIDs).
			Msg("checking relationship tuples")

		for _, pID := range pIDs {
			ac := fgax.AccessCheck{
				SubjectID:   userID,
				SubjectType: auth.GetAuthzSubjectType(ctx),
				ObjectID:    pID,
				ObjectType:  "program",
				Relation:    relation,
			}

			access, err := utils.AuthzClientFromContext(ctx).CheckAccess(ctx, ac)
			if err != nil {
				return privacy.Skipf("unable to check access, %s", err.Error())
			}

			if !access {
				log.Debug().Str("relation", relation).
					Str("program_id", pID).
					Msg("access allowed")

				// no matter the operation, if the user does not have access to the program
				// deny the mutation
				return generated.ErrPermissionDenied
			}
		}

		// if we reach here, user has access to all programs
		// and the mutation is allowed if it is a create operation
		if m.Op() == generated.OpCreate {
			return privacy.Allow
		}

		// if the mutation is not a create operation, continue to the next rule to
		// ensure they have access to the object
		return privacy.Skipf("mutation is not a create operation, skipping")
	})
}

// getOwnerIDFromEntMutation extracts the object id from a the mutation
// by attempting to cast the mutation to a risk mutation
// if additional object types are needed, they should be added to this function
func getProgramIDFromEntMutation(m generated.Mutation) ([]string, error) {
	if o, ok := m.(*generated.RiskMutation); ok {
		return o.ProgramIDs(), nil
	}

	return nil, nil
}
