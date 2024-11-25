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

// CanCreateObjectsInProgram is a rule that returns allow decision if user has edit access in the program
// which allows them to create program owned objects
func CanCreateObjectsInProgram() privacy.MutationRuleFunc {
	return privacy.MutationRuleFunc(func(ctx context.Context, m generated.Mutation) error {
		pID, err := getProgramIDFromEntMutation(m)
		if err != nil {
			return privacy.Denyf("unable to get program id from mutation, %s", err.Error())
		}

		if pID == "" {
			return privacy.Skipf("no owner set on request, cannot check access")
		}

		log.Debug().Msg("checking mutation access")

		relation := fgax.CanEdit

		userID, err := auth.GetUserIDFromContext(ctx)
		if err != nil {
			return err
		}

		log.Info().Str("relation", relation).
			Str("program_id", pID).
			Msg("checking relationship tuples")

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

		if access {
			log.Debug().Str("relation", relation).
				Str("program_id", pID).
				Msg("access allowed")

			return privacy.Allow
		}

		// deny if it was a mutation is not allowed
		return generated.ErrPermissionDenied
	})
}

// getOwnerIDFromEntMutation extracts the object id from a the mutation
// by attempting to cast the mutation to a risk mutation
// if additional object types are needed, they should be added to this function
func getProgramIDFromEntMutation(m generated.Mutation) (string, error) {
	if o, ok := m.(*generated.RiskMutation); ok {
		if pID, ok := o.ProgramID(); ok {
			return pID, nil
		}

		return "", nil
	}

	return "", nil
}
