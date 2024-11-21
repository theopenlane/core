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

	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// getTuplesParentToAdd gets the tuples that need to be added to the authz service based on the edges that were added
// with the parent relation
func getTuplesParentToAdd(ctx context.Context, m ent.Mutation, objectID string, parents []string) ([]fgax.TupleKey, error) {
	var addTuples []fgax.TupleKey

	for _, parent := range parents {
		subjectType := strings.ReplaceAll(parent, "_id", "")

		t, err := getTuplesToAdd(ctx, m, objectID, m.Type(), fgax.ParentRelation, subjectType, parent)
		if err != nil {
			return nil, err
		}

		addTuples = append(addTuples, t...)
	}

	return addTuples, nil
}

// getTuplesToAdd is the generic function to get the tuples that need to be added to the authz service based on the edges that were added
// it is recommend to use the helper functions that call this instead of calling this directly
// for example, to add a parent relationship, use getTuplesParentToAdd, or for an org owner relationship, use createOrgOwnerParentTuple
func getTuplesToAdd(ctx context.Context, m ent.Mutation, objectID, objectType, relation, subjectType, edgeField string) ([]fgax.TupleKey, error) {
	var addTuples []fgax.TupleKey

	subjectIDs, err := getAddedParentIDsFromEntMutation(ctx, m, edgeField)
	if err != nil {
		return nil, err
	}

	// edge is not set, no need to add a tuple
	if len(subjectIDs) == 0 {
		return addTuples, nil
	}

	for _, subjectID := range subjectIDs {
		tuple := fgax.GetTupleKey(fgax.TupleRequest{
			SubjectID:   subjectID,
			SubjectType: subjectType,
			ObjectID:    objectID,   // this is the object id being created
			ObjectType:  objectType, // this is the object type being created
			Relation:    relation,
		})

		addTuples = append(addTuples, tuple)
	}

	return addTuples, nil
}

// createOrgOwnerParentTuple creates the tuple for the parent org owner relationship
func createOrgOwnerParentTuple(ctx context.Context, m ent.Mutation, objectID string) ([]fgax.TupleKey, error) {
	var addTuples []fgax.TupleKey

	parentField := "owner_id"
	subjectType := "organization"

	t, err := getTuplesToAdd(ctx, m, objectID, m.Type(), fgax.ParentRelation, subjectType, parentField)
	if err != nil {
		return nil, err
	}

	addTuples = append(addTuples, t...)

	return addTuples, nil
}

// createEditorTuple creates the tuple for the editor relationship based on an edge field
func createEditorTuple(ctx context.Context, m ent.Mutation, objectID string, objects map[string]string) ([]fgax.TupleKey, error) {
	var addTuples []fgax.TupleKey

	for parentField, subjectType := range objects {
		t, err := getTuplesToAdd(ctx, m, objectID, m.Type(), fgax.EditorRelation, subjectType, parentField)
		if err != nil {
			return nil, err
		}

		addTuples = append(addTuples, t...)
	}

	return addTuples, nil
}

// createBlockedTuple creates the tuple for the blocked relationship based on an edge field
func createBlockedTuple(ctx context.Context, m ent.Mutation, objectID string, objects map[string]string) ([]fgax.TupleKey, error) {
	var addTuples []fgax.TupleKey

	for parentField, subjectType := range objects {
		t, err := getTuplesToAdd(ctx, m, objectID, m.Type(), fgax.BlockedRelation, subjectType, parentField)
		if err != nil {
			return nil, err
		}

		addTuples = append(addTuples, t...)
	}

	return addTuples, nil
}

// getTuplesToRemove is the generic function to get the tuples that need to be added to the authz service based on the edges that were added
// it is recommend to use the helper functions that call this instead of calling this directly
// for example, to add a parent relationship, use getParentTuplesToRemove
func getTuplesToRemove(ctx context.Context, m ent.Mutation, objectID, objectType, relation, subjectType, edgeField string) ([]fgax.TupleKey, error) {
	var removeTuples []fgax.TupleKey

	subjectIDs, err := getRemovedParentIDsFromEntMutation(ctx, m, edgeField)
	if err != nil {
		return nil, err
	}

	// edge is not set, no need to add a tuple
	if len(subjectIDs) == 0 {
		return removeTuples, nil
	}

	for _, subjectID := range subjectIDs {
		parentTuple := fgax.GetTupleKey(fgax.TupleRequest{
			SubjectID:   subjectID,
			SubjectType: subjectType,
			ObjectID:    objectID,   // this is the object id being created
			ObjectType:  objectType, // this is the object type being created
			Relation:    relation,
		})

		removeTuples = append(removeTuples, parentTuple)
	}

	return removeTuples, nil
}

// getParentTuplesToRemove gets the tuples that need to be removed from the authz service based on the edges that were removed
// with the parent relation
func getParentTuplesToRemove(ctx context.Context, m ent.Mutation, objectID string, parents []string) ([]fgax.TupleKey, error) {
	var removeTuples []fgax.TupleKey

	for _, parent := range parents {
		subjectType := strings.ReplaceAll(parent, "_id", "")

		t, err := getTuplesToRemove(ctx, m, objectID, m.Type(), fgax.ParentRelation, subjectType, parent)
		if err != nil {
			return nil, err
		}

		removeTuples = append(removeTuples, t...)
	}

	return removeTuples, nil
}

// removeEditorTuples removes the tuple for the editor relationship based on an edge field
func removeEditorTuples(ctx context.Context, m ent.Mutation, objectID string, objects map[string]string) ([]fgax.TupleKey, error) {
	var removeTuples []fgax.TupleKey

	for parentField, subjectType := range objects {
		t, err := getTuplesToRemove(ctx, m, objectID, m.Type(), fgax.EditorRelation, subjectType, parentField)
		if err != nil {
			return nil, err
		}

		removeTuples = append(removeTuples, t...)
	}

	return removeTuples, nil
}

// removeBlockedTuples removes the tuple for the blocked relationship based on an edge field
func removeBlockedTuples(ctx context.Context, m ent.Mutation, objectID string, objects map[string]string) ([]fgax.TupleKey, error) {
	var removeTuples []fgax.TupleKey

	for parentField, subjectType := range objects {
		t, err := getTuplesToRemove(ctx, m, objectID, m.Type(), fgax.BlockedRelation, subjectType, parentField)
		if err != nil {
			return nil, err
		}

		removeTuples = append(removeTuples, t...)
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
		// check if the edge is the parent field or the parent field pluralized
		if e == parentEdge || e == parentEdge+"s" {
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
		if e == parentEdge || e == parentEdge+"s" {
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

// addTokenEditPermissions adds the edit permissions for the api token to the object
func addTokenEditPermissions(ctx context.Context, oID string, objectType string) error {
	// get auth info from context
	ac, err := auth.GetAuthenticatedUserContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("unable to get subject id from context, cannot update token permissions")

		return err
	}

	req := fgax.TupleRequest{
		SubjectID:   ac.SubjectID,
		SubjectType: auth.GetAuthzSubjectType(ctx),
		Relation:    fgax.CanEdit,
		ObjectID:    oID,
		ObjectType:  objectType,
	}

	log.Debug().Interface("request", req).
		Msg("creating edit tuples for api token")

	if _, err := utils.AuthzClientFromContext(ctx).WriteTupleKeys(ctx, []fgax.TupleKey{fgax.GetTupleKey(req)}, nil); err != nil {
		log.Error().Err(err).Msg("failed to create relationship tuple")

		return ErrInternalServerError
	}

	return nil
}
