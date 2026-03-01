package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
)

// OrgOwnedTuplesHookWithAdmin is a hook that adds organization owned tuples for the object being created
// it will add the user and parent (organization owner_id) permissions to the object
// on creation, and will also add an admin user permission to the object
func OrgOwnedTuplesHookWithAdmin() ent.Hook {
	return hookOrgOwnedTuples(true)
}

// OrgOwnedTuplesHook is a hook that adds organization owned tuples for the object being created
// it will only add the parent organization permissions, and no specific user permissions
func OrgOwnedTuplesHook() ent.Hook {
	return hookOrgOwnedTuples(false)
}

// hookOrgOwnedTuples is a hook that adds organization owned tuples for the object being created
// it will add the user and parent (organization owner_id) permissions to the object
// on creation
// by default, it will always add an admin user permission to the object
func hookOrgOwnedTuples(includeAdminRelation bool) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			objectID, err := GetObjectIDFromEntValue(retVal)
			if err != nil {
				return nil, err
			}

			var addTuples []fgax.TupleKey

			// add user and org owner tuples to the object on creation
			if m.Op() == ent.OpCreate {
				if includeAdminRelation {
					orgOwnedCaller, ok := auth.CallerFromContext(ctx)
					if !ok || orgOwnedCaller == nil {
						return nil, auth.ErrNoAuthUser
					}

					// add user permissions to the object as the parent on creation
					userTuple := fgax.GetTupleKey(fgax.TupleRequest{
						SubjectID:   orgOwnedCaller.SubjectID,
						SubjectType: orgOwnedCaller.SubjectType(),
						ObjectID:    objectID,                        // this is the object id being created
						ObjectType:  GetObjectTypeFromEntMutation(m), // this is the object type being created
						Relation:    fgax.AdminRelation,
					})

					addTuples = append(addTuples, userTuple)
				}

				additionalAddTuples, err := createOrgOwnerParentTuple(ctx, m, objectID)
				if err != nil {
					return nil, err
				}

				addTuples = append(addTuples, additionalAddTuples...)
			}

			// write the tuples to the authz service
			if len(addTuples) != 0 {
				if _, err := utils.AuthzClientFromContext(ctx).WriteTupleKeys(ctx, addTuples, nil); err != nil {
					return nil, err
				}
			}

			logx.FromContext(ctx).Debug().Interface("tuples", addTuples).Msg("added organization permissions")

			return retVal, err
		},
		)
	}
}
