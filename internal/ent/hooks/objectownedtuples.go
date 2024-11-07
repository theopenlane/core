package hooks

import (
	"context"
	"encoding/json"
	"strings"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
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

			var tuples []fgax.TupleKey

			objectID, err := getObjectIDFromEntValue(retVal)
			if err != nil {
				return nil, err
			}

			if !skipUser {
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
					ObjectID:    objectID, // this is the object id being created
					ObjectType:  m.Type(), // this is the object type being created
					Relation:    fgax.ParentRelation,
				})

				tuples = append(tuples, userTuple)
			}

			for _, parent := range parents {
				subjectID, err := getParentIDFromEntValue(retVal, parent)
				if err != nil {
					return nil, err
				}

				// edge is not set, no need to add a tuple
				if subjectID == "" {
					continue
				}

				subjectType := strings.ReplaceAll(parent, "_id", "")

				parentTuple := fgax.GetTupleKey(fgax.TupleRequest{
					SubjectID:   subjectID,
					SubjectType: subjectType,
					ObjectID:    objectID, // this is the object id being created
					ObjectType:  m.Type(), // this is the object type being created
					Relation:    fgax.ParentRelation,
				})

				tuples = append(tuples, parentTuple)
			}

			if _, err := generated.FromContext(ctx).Authz.WriteTupleKeys(ctx, tuples, nil); err != nil {
				return nil, err
			}

			log.Info().Interface("req", tuples).Msg("added object permissions")

			return retVal, err
		},
		)
	}
}

// getObjectIDFromEntValue extracts the object id from a generic ent value return type
func getObjectIDFromEntValue(m ent.Value) (string, error) {
	type objectIDer struct {
		ID string `json:"id"`
	}

	tmp, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	var o objectIDer
	if err := json.Unmarshal(tmp, &o); err != nil {
		return "", err
	}

	return o.ID, nil
}

// getParentIDFromEntValue extracts the parent id from a generic ent value return type
// if it is not set, it will return an empty string
func getParentIDFromEntValue(m ent.Value, parentField string) (string, error) {
	tmp, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	var o map[string]interface{}
	if err := json.Unmarshal(tmp, &o); err != nil {
		return "", err
	}

	if v, ok := o[parentField]; ok {
		return v.(string), nil
	}

	return "", nil
}
