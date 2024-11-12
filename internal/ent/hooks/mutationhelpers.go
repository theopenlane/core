package hooks

import (
	"context"
	"encoding/json"
	"strings"

	"entgo.io/ent"
	goUpper "github.com/99designs/gqlgen/codegen/templates"
	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/iam/fgax"
)

// getTuplesToAdd gets the tuples that need to be added to the authz service based on the edges that were added
func getTuplesToAdd(ctx context.Context, m ent.Mutation, objectID string, parents []string) ([]fgax.TupleKey, error) {
	var addTuples []fgax.TupleKey

	for _, parent := range parents {
		subjectIDs, err := getAddedParentIDsFromEntMutation(ctx, m, parent)
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

			addTuples = append(addTuples, parentTuple)
		}
	}

	return addTuples, nil
}

// getTuplesToRemove gets the tuples that need to be removed from the authz service based on the edges that were removed
func getTuplesToRemove(ctx context.Context, m ent.Mutation, objectID string, parents []string) ([]fgax.TupleKey, error) {
	var removeTuples []fgax.TupleKey

	for _, parent := range parents {
		subjectIDs, err := getRemovedParentIDsFromEntMutation(ctx, m, parent)
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

			removeTuples = append(removeTuples, parentTuple)
		}
	}

	return removeTuples, nil
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
func getAddedParentIDsFromEntMutation(ctx context.Context, m ent.Mutation, parentField string) ([]string, error) {
	if v, ok := m.Field(parentField); ok {
		return []string{v.(string)}, nil
	}

	// check if the edges were set on the mutation
	edges := m.AddedEdges()
	for _, e := range edges {
		parentEdge := strings.ReplaceAll(parentField, "_id", "")
		if e == parentEdge {
			// we need to parse the graphql input to get the ids
			field := goUpper.ToGo(parentField) + "s"
			if m.Op() != ent.OpCreate {
				field = "Add" + goUpper.ToGo(parentField) + "s"
			}

			return parseGraphqlInputForEdgeIDs(ctx, field)
		}
	}

	return nil, nil
}

// getParentIDFromEntValue extracts the parent id from a generic ent value return type
// if it is not set, it will return an empty string
// this function does not ensure that the mutation was successful, it only extracts the id
func getRemovedParentIDsFromEntMutation(ctx context.Context, m ent.Mutation, parentField string) ([]string, error) {
	if v, ok := m.Field(parentField); ok {
		return []string{v.(string)}, nil
	}

	// check if the edges were set on the mutation
	edges := m.RemovedEdges()
	for _, e := range edges {
		parentEdge := strings.ReplaceAll(parentField, "_id", "")
		if e == parentEdge {
			// we need to parse the graphql input to get the ids
			field := goUpper.ToGo(parentField) + "s"
			if m.Op() != ent.OpCreate {
				field = "Remove" + goUpper.ToGo(parentField) + "s"
			}

			return parseGraphqlInputForEdgeIDs(ctx, field)
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
	out := v[parentField]

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
