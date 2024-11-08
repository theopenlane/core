package hooks

import (
	"context"
	"encoding/json"
	"strings"

	"entgo.io/ent"
	goUpper "github.com/99designs/gqlgen/codegen/templates"
	"github.com/99designs/gqlgen/graphql"
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
				subjectIDs, err := getParentIDsFromEntMutation(ctx, m, parent)
				if err != nil {
					return nil, err
				}

				// edge is not set, no need to add a tuple
				if len(subjectIDs) == 0 {
					continue
				}

				subjectType := strings.ReplaceAll(parent, "_id", "")

				for _, subjectID := range subjectIDs {
					parentTuple := fgax.GetTupleKey(fgax.TupleRequest{
						SubjectID:   subjectID,
						SubjectType: subjectType,
						ObjectID:    objectID, // this is the object id being created
						ObjectType:  m.Type(), // this is the object type being created
						Relation:    fgax.ParentRelation,
					})

					tuples = append(tuples, parentTuple)
				}
			}

			if _, err := generated.FromContext(ctx).Authz.WriteTupleKeys(ctx, tuples, nil); err != nil {
				return nil, err
			}

			log.Debug().Interface("tuples", tuples).Msg("added object permissions")

			return retVal, err
		},
		)
	}
}

// getObjectIDFromEntValue extracts the object id from a generic ent value return type
// this function should be called after the mutation has been successful
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
// this function does not ensure that the mutation was successful, it only extracts the id
func getParentIDsFromEntMutation(ctx context.Context, m ent.Mutation, parentField string) ([]string, error) {
	if v, ok := m.Field(parentField); ok {
		return []string{v.(string)}, nil
	}

	// check if the edges were set on the mutation
	edges := m.AddedEdges()
	for _, e := range edges {
		parentEdge := strings.ReplaceAll(parentField, "_id", "")
		if e == parentEdge {
			// we need to parse the graphql input to get the ids
			return parseGraphqlInputForEdgeIDs(ctx, parentField)
		}
	}

	return nil, nil
}

// parseGraphqlInputForEdgeIDs parses the graphql input to get the ids for the parent field
func parseGraphqlInputForEdgeIDs(ctx context.Context, parentField string) ([]string, error) {
	fCtx := graphql.GetFieldContext(ctx)

	// check if the input is set
	input, ok := fCtx.Args["input"]
	if !ok {
		return nil, nil
	}

	// unmarshal the input
	tmp, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var v map[string]interface{}
	if err := json.Unmarshal(tmp, &v); err != nil {
		return nil, err
	}

	// check for the edge
	field := goUpper.ToGo(parentField) + "s"
	out := v[field]

	tmp, err = json.Marshal(out)
	if err != nil {
		return nil, err
	}

	// return the ids if they are set
	var ids []string
	if err := json.Unmarshal(tmp, &ids); err != nil {
		return nil, err
	}

	return ids, nil
}
