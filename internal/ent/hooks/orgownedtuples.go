package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
)

// HookOrgOwnedTuples is a hook that adds organization owned tuples for the object being created
// it will add the user and parent (organization owner_id) permissions to the object
// on creation
// by default, it will always add an admin user permission to the object
func HookOrgOwnedTuples(skipUser bool) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			objectID, err := getObjectIDFromEntValue(retVal)
			if err != nil {
				return nil, err
			}

			var addTuples []fgax.TupleKey

			// add user permissions to the object on creation
			if !skipUser && m.Op() == ent.OpCreate {
				a, err := auth.GetAuthenticatedUserContext(ctx)
				if err != nil {
					return nil, err
				}

				subject := "user"
				if a.AuthenticationType == auth.APITokenAuthentication {
					subject = "service"
				}

				// add user permissions to the object as the parent on creation
				userTuple := fgax.GetTupleKey(fgax.TupleRequest{
					SubjectID:   a.SubjectID,
					SubjectType: subject,
					ObjectID:    objectID,                        // this is the object id being created
					ObjectType:  getObjectTypeFromEntMutation(m), // this is the object type being created
					Relation:    fgax.AdminRelation,
				})

				addTuples = append(addTuples, userTuple)
			}

			additionalAddTuples, err := createOrgOwnerParentTuple(ctx, m, objectID)
			if err != nil {
				return nil, err
			}

			addTuples = append(addTuples, additionalAddTuples...)

			// write the tuples to the authz service
			if len(addTuples) != 0 {
				if _, err := generated.FromContext(ctx).Authz.WriteTupleKeys(ctx, addTuples, nil); err != nil {
					return nil, err
				}
			}

			log.Debug().Interface("tuples", addTuples).Msg("added organization permissions")

			return retVal, err
		},
		)
	}
}
