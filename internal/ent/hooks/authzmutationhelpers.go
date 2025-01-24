package hooks

import (
	"context"
	"encoding/json"
	"strings"

	"entgo.io/ent"
	goUpper "github.com/99designs/gqlgen/codegen/templates"
	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// GetObjectTypeFromEntMutation gets the object type from the ent mutation
func GetObjectTypeFromEntMutation(m ent.Mutation) string {
	return strcase.SnakeCase(m.Type())
}

// getTuplesToAdd is the generic function to get the tuples that need to be added to the authz service based on the edges that were added
// it is recommend to use the helper functions that call this instead of calling this directly
// for example, to add a parent relationship, use createParentTuples, or for an org owner relationship, use createOrgOwnerParentTuple
// this takes in the tuple request and sets the subject and subject id based on the edge field and tuple set relation
func getTuplesToAdd(ctx context.Context, m ent.Mutation, tr fgax.TupleRequest, edgeField string) ([]fgax.TupleKey, error) {
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
		// set the subject id for the tuple
		tr.SubjectID = subjectID

		addTuples = append(addTuples, fgax.GetTupleKey(tr))
	}

	return addTuples, nil
}

// createParentTuples gets the tuples that need to be added to the authz service based on the edges that were added
// with the parent relation
func createParentTuples(ctx context.Context, m ent.Mutation, objectID string, parents []string) ([]fgax.TupleKey, error) {
	var addTuples []fgax.TupleKey

	for _, parent := range parents {
		subjectType := strings.ReplaceAll(parent, "_id", "")

		// create the tuple for the parent relationship without the subject id
		// this will be filled in by getTuplesToAdd based on the parent field
		tr := fgax.TupleRequest{
			SubjectType: subjectType,
			ObjectID:    objectID,                        // this is the object id being created
			ObjectType:  GetObjectTypeFromEntMutation(m), // this is the object type being created
			Relation:    fgax.ParentRelation,
		}

		t, err := getTuplesToAdd(ctx, m, tr, parent)
		if err != nil {
			return nil, err
		}

		addTuples = append(addTuples, t...)
	}

	return addTuples, nil
}

// createOrgOwnerParentTuple creates the tuple for the parent org owner relationship
func createOrgOwnerParentTuple(ctx context.Context, m ent.Mutation, objectID string) ([]fgax.TupleKey, error) {
	var addTuples []fgax.TupleKey

	// create the tuple for the parent org owner relationship without the subject id
	// this will be filled in by getTuplesToAdd based on the owner id field
	tr := fgax.TupleRequest{
		SubjectType: "organization",
		ObjectID:    objectID,                        // this is the object id being created
		ObjectType:  GetObjectTypeFromEntMutation(m), // this is the object type being created
		Relation:    fgax.ParentRelation,
	}

	t, err := getTuplesToAdd(ctx, m, tr, "owner_id")
	if err != nil {
		return nil, err
	}

	addTuples = append(addTuples, t...)

	return addTuples, nil
}

// createTuplesByRelation creates the tuple for the specified relationship based on an edge field
// with the member relation for the subject, e.g. group#member
func createTuplesByRelation(ctx context.Context, m ent.Mutation, objectID string, relation fgax.Relation, objects map[string]string) ([]fgax.TupleKey, error) {
	var addTuples []fgax.TupleKey

	for edgeField, subjectType := range objects {
		// create the tuple for the editor relationship without the subject id
		// this will be filled in by getTuplesToAdd based on the edge field
		tr := fgax.TupleRequest{
			SubjectType:     subjectType,
			SubjectRelation: fgax.MemberRelation,
			ObjectID:        objectID,                        // this is the object id being created
			ObjectType:      GetObjectTypeFromEntMutation(m), // this is the object type being created
			Relation:        relation.String(),
		}

		// this will create tuples such as group:ulid-of-group#member with the relation provided
		t, err := getTuplesToAdd(ctx, m, tr, edgeField)
		if err != nil {
			return nil, err
		}

		addTuples = append(addTuples, t...)
	}

	return addTuples, nil
}

// getTuplesToRemove is the generic function to get the tuples that need to be added to the authz service based on the edges that were added
// it is recommend to use the helper functions that call this instead of calling this directly
// for example, to add a parent relationship, use removeParentTuples
// this takes in the tuple request and sets the subject and subject id based on the edge field and tuple set relation
func getTuplesToRemove(ctx context.Context, m ent.Mutation, tr fgax.TupleRequest, edgeField string) ([]fgax.TupleKey, error) {
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
		// set the subject id for the tuple
		tr.SubjectID = subjectID

		removeTuples = append(removeTuples, fgax.GetTupleKey(tr))
	}

	return removeTuples, nil
}

// removeParentTuples gets the tuples that need to be removed from the authz service based on the edges that were removed
// with the parent relation
func removeParentTuples(ctx context.Context, m ent.Mutation, objectID string, parents []string) ([]fgax.TupleKey, error) {
	var removeTuples []fgax.TupleKey

	for _, parent := range parents {
		subjectType := strings.ReplaceAll(parent, "_id", "")

		// create the tuple for the parent relationship without the subject id
		// this will be filled in by getTuplesToRemove based on the parent field
		tr := fgax.TupleRequest{
			SubjectType: subjectType,
			ObjectID:    objectID,                        // this is the object id being created
			ObjectType:  GetObjectTypeFromEntMutation(m), // this is the object type being created
			Relation:    fgax.ParentRelation,
		}

		t, err := getTuplesToRemove(ctx, m, tr, parent)
		if err != nil {
			return nil, err
		}

		removeTuples = append(removeTuples, t...)
	}

	return removeTuples, nil
}

// removeTuplesByRelation removes the tuple for the provided relationship based on an edge field
func removeTuplesByRelation(ctx context.Context, m ent.Mutation, objectID string, relation fgax.Relation, objects map[string]string) ([]fgax.TupleKey, error) {
	var removeTuples []fgax.TupleKey

	for edgeField, subjectType := range objects {
		// create the tuple to remove for the editor relationship without the subject id
		// this will be filled in by getTuplesToRemove based on the edge field
		tr := fgax.TupleRequest{
			SubjectType:     subjectType,
			SubjectRelation: fgax.MemberRelation,
			ObjectID:        objectID,                        // this is the object id being created
			ObjectType:      GetObjectTypeFromEntMutation(m), // this is the object type being created
			Relation:        relation.String(),
		}

		t, err := getTuplesToRemove(ctx, m, tr, edgeField)
		if err != nil {
			return nil, err
		}

		removeTuples = append(removeTuples, t...)
	}

	return removeTuples, nil
}

// GetObjectIDFromEntValue extracts the object id from a generic ent value return type
// this function should be called after the mutation has been successful
func GetObjectIDFromEntValue(m ent.Value) (string, error) {
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
			field := goUpper.ToGo(parentField)
			if m.Op() != ent.OpCreate {
				field = "Add" + field
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
			field := goUpper.ToGo(parentField)
			if m.Op() != ent.OpCreate {
				field = "Remove" + field
			}

			return parseGraphqlInputForEdgeIDs(ctx, field)
		}
	}

	return nil, nil
}

// parseGraphqlInputForEdgeIDs parses the graphql input to get the ids for the parent field
func parseGraphqlInputForEdgeIDs(ctx context.Context, parentField string) ([]string, error) {
	fCtx := graphql.GetFieldContext(ctx)

	if fCtx == nil || fCtx.Args == nil {
		return nil, nil
	}

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
	out, ok := v[parentField+"s"] // plural first
	if !ok {
		out, ok = v[parentField] // check for singular
		if !ok {
			return nil, nil
		} else {
			if strOut, ok := out.(string); ok {
				return []string{strOut}, nil
			}
		}
	}

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
