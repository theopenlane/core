package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// HookObjectOwnedTuples is a hook that adds object owned tuples for the object being created
// given a set of parent id fields, it will add the user and parent permissions to the object
// on creation
// by default, it will always add a user permission to the object
func HookObjectOwnedTuples(parents []string, skipUser bool) ent.Hook {
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
					Relation:    fgax.ParentRelation,
				})

				addTuples = append(addTuples, userTuple)
			}

			additionalAddTuples, err := createParentTuples(ctx, m, objectID, parents)
			if err != nil {
				return nil, err
			}

			addTuples = append(addTuples, additionalAddTuples...)

			removeTuples, err := removeParentTuples(ctx, m, objectID, parents)
			if err != nil {
				return nil, err
			}

			// write the tuples to the authz service
			if len(addTuples) != 0 || len(removeTuples) != 0 {
				if _, err := utils.AuthzClientFromContext(ctx).WriteTupleKeys(ctx, addTuples, removeTuples); err != nil {
					return nil, err
				}

				if len(addTuples) != 0 {
					log.Debug().Interface("tuples", addTuples).Msg("added object permissions")
				}

				if len(removeTuples) != 0 {
					log.Debug().Interface("tuples", removeTuples).Msg("removed object permissions")
				}
			}

			return retVal, err
		},
		)
	}
}

// HookRelationTuples is a hook that adds tuples for the object being created
// the objects input is a map of object id fields to the object type
// these tuples based are based on the direct relation, e.g. a group#member to another object
func HookRelationTuples(objects map[string]string, relation fgax.Relation) ent.Hook {
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

			var (
				addTuples    []fgax.TupleKey
				removeTuples []fgax.TupleKey
			)

			addTuples, err = createTuplesByRelation(ctx, m, objectID, relation, objects)
			if err != nil {
				return nil, err
			}

			removeTuples, err = removeTuplesByRelation(ctx, m, objectID, relation, objects)
			if err != nil {
				return nil, err
			}

			// write the tuples to the authz service
			if len(addTuples) != 0 || len(removeTuples) != 0 {
				if _, err := utils.AuthzClientFromContext(ctx).WriteTupleKeys(ctx, addTuples, removeTuples); err != nil {
					return nil, err
				}

				log.Debug().Interface("tuples", addTuples).Msg("added tuples")
				log.Debug().Interface("tuples", removeTuples).Msg("removed tuples")
			}

			return retVal, err
		},
		)
	}
}
