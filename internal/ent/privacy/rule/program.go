package rule

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

const (
	// ProgramParent is the parent type for program
	ProgramParent = "program"
	// ControlParent is the parent type for control
	ControlParent = "control"
)

// CanCreateObjectsUnderParent is a rule that returns allow decision if user has edit access in the parent(s)
// which allows them to create objects associated with the parent
func CanCreateObjectsUnderParent[T generated.Mutation](parentType string) privacy.MutationRuleFunc {
	return privacy.MutationRuleFunc(func(ctx context.Context, m generated.Mutation) error {
		pIDs, err := getParentIDFromEntMutation[T](m, parentType)
		if err != nil {
			return privacy.Denyf("unable to get parent id from mutation, %s", err.Error())
		}

		if len(pIDs) == 0 {
			return privacy.Skipf("no parent set on request, skipping")
		}

		relation := fgax.CanEdit

		user, err := auth.GetAuthenticatedUserContext(ctx)
		if err != nil {
			return err
		}

		log.Debug().Str("relation", relation).
			Strs("parent_ids", pIDs).
			Msg("checking relationship tuples")

		for _, pID := range pIDs {
			ac := fgax.AccessCheck{
				SubjectID:   user.SubjectID,
				SubjectType: auth.GetAuthzSubjectType(ctx),
				ObjectID:    pID,
				ObjectType:  fgax.Kind(parentType),
				Relation:    relation,
				Context:     utils.NewOrganizationContextKey(user.SubjectEmail),
			}

			access, err := utils.AuthzClient(ctx, m).CheckAccess(ctx, ac)
			if err != nil {
				return privacy.Skipf("unable to check access, %s", err.Error())
			}

			if !access {
				log.Debug().Interface("access_check", ac).
					Msg("access not allowed")

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

// ProgramParentMutation is an interface that defines the method to get the program ids from the mutation
type ProgramParentMutation interface {
	ProgramsIDs() []string
}

// getProgramIDFromEntMutation returns the program ids from the mutation
func getProgramIDFromEntMutation[T ProgramParentMutation](m generated.Mutation) ([]string, error) {
	return m.(T).ProgramsIDs(), nil
}

// ControlParentMutation is an interface that defines the method to get the control ids from the mutation
type ControlParentMutation interface {
	ControlsIDs() []string
}

// getProgramIDFromEntMutation returns the program ids from the mutation
func getControlIDFromEntMutation[T ControlParentMutation](m generated.Mutation) ([]string, error) {
	return m.(T).ControlsIDs(), nil
}

// getParentIDFromEntMutation returns the parent ids from the mutation
func getParentIDFromEntMutation[T ent.Mutation](m generated.Mutation, parentType string) ([]string, error) {
	switch parentType {
	case ProgramParent:
		return getProgramIDFromEntMutation[ProgramParentMutation](m)
	case ControlParent:
		return getControlIDFromEntMutation[ControlParentMutation](m)
	}

	return []string{}, nil
}
