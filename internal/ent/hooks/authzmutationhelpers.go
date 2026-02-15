package hooks

import (
	"context"
	"strings"

	"entgo.io/ent"
	goUpper "github.com/99designs/gqlgen/codegen/templates"
	"github.com/99designs/gqlgen/graphql"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	ownerFieldName = "owner_id"
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

	if strings.EqualFold(edgeField, "organization_id") {
		edgeField = ownerFieldName
	}

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
		SubjectType: generated.TypeOrganization,
		ObjectID:    objectID,                        // this is the object id being created
		ObjectType:  GetObjectTypeFromEntMutation(m), // this is the object type being created
		Relation:    fgax.ParentRelation,
	}

	t, err := getTuplesToAdd(ctx, m, tr, ownerFieldName)
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

// GetObjectIDsFromMutation gets the object ids from the mutation, if it is a create it will use the ent.Value
// to get the id, requiring the mutation be executed first
// For updates, it will use the `IDs()` function to get the IDs by querying the database and
// returning the entity ids that match the mutation's predicate.
func GetObjectIDsFromMutation(ctx context.Context, m utils.GenericMutation, v ent.Value) ([]string, error) {
	if m.Op().Is(ent.OpCreate) {
		id, err := GetObjectIDFromEntValue(v)
		if err != nil {
			return nil, err
		}

		return []string{id}, nil
	}

	return getMutationIDs(ctx, m), nil
}

// GetObjectIDFromEntValue extracts the object id from a generic ent value return type
// this function should be called after the mutation has been successful
func GetObjectIDFromEntValue(m ent.Value) (string, error) {
	type objectIDer struct {
		ID string `json:"id"`
	}

	var o objectIDer
	if err := jsonx.RoundTrip(m, &o); err != nil {
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
		foundEdge := checkForEdge(parentField, e)
		if foundEdge != "" {
			// we need to parse the graphql input to get the ids
			field := goUpper.ToGo(parentField)
			if m.Op() != ent.OpCreate {
				field = "Add" + field
			}

			return parseMutationForAddedEdgeIDs(ctx, foundEdge, field, m)
		}
	}

	return nil, nil
}

// getParentIDFromEntValue extracts the parent id from a generic ent value return type
// if it is not set, it will return an empty string
// this function does not ensure that the mutation was successful, it only extracts the id
func getRemovedParentIDsFromEntMutation(ctx context.Context, m ent.Mutation, parentField string) ([]string, error) {
	if ok := m.FieldCleared(parentField); ok &&
		(m.Op().Is(ent.OpUpdateOne) || m.Op().Is(ent.OpUpdate)) {
		v, err := m.OldField(ctx, parentField)
		if err != nil {
			return nil, err
		}

		return []string{v.(string)}, nil
	}

	// check if the edges were set on the mutation
	edges := m.RemovedEdges()
	for _, e := range edges {
		foundEdge := checkForEdge(parentField, e)

		if foundEdge != "" {
			// we need to parse the graphql input to get the ids
			field := goUpper.ToGo(parentField)
			if m.Op() != ent.OpCreate {
				field = "Remove" + field
			}

			return parseMutationForRemovedEdgeIDs(ctx, foundEdge, field, m)
		}
	}

	return nil, nil
}

// checkForEdge checks if the edge field is the parent field or the parent field pluralized
// or not an set edge at all
func checkForEdge(parentField, edgeField string) string {
	parentEdge := strings.ReplaceAll(parentField, "_id", "")
	pluralEdge := parentEdge + "s"
	foundEdge := ""

	// determine if the edge is the parent field or the parent field pluralized
	switch edgeField {
	case parentEdge:
		foundEdge = parentEdge
	case pluralEdge:
		foundEdge = pluralEdge
	}

	return foundEdge
}

// parseMutationForAddedEdgeIDs parses the mutation to get the ids for the parent edge by first checking the mutation
// and then the graphql input
func parseMutationForAddedEdgeIDs(ctx context.Context, parentEdge, parentField string, m ent.Mutation) ([]string, error) {
	val := m.AddedIDs(parentEdge)

	if len(val) == 0 {
		return parseGraphqlInputForEdgeIDs(ctx, parentField)
	}

	ids := []string{}

	for _, v := range val {
		ids = append(ids, v.(string))
	}

	return ids, nil
}

// parseMutationForRemovedEdgeIDs parses the mutation to get the ids removed for the parent field by first checking the mutation
// and then the graphql input
func parseMutationForRemovedEdgeIDs(ctx context.Context, parentEdge, parentField string, m ent.Mutation) ([]string, error) {
	val := m.RemovedIDs(parentEdge)

	if len(val) == 0 {
		return parseGraphqlInputForEdgeIDs(ctx, parentField)
	}

	ids := []string{}

	for _, v := range val {
		ids = append(ids, v.(string))
	}

	return ids, nil
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

	var v map[string]interface{}
	if err := jsonx.RoundTrip(input, &v); err != nil {
		return nil, err
	}

	// check for the edge
	out, ok := v[parentField+"s"] // plural first
	if !ok {
		out, ok = v[parentField] // check for singular
		if !ok {
			return nil, nil
		}

		if strOut, ok := out.(string); ok {
			return []string{strOut}, nil
		}
	}

	// return the ids if they are set
	var ids []string
	if err := jsonx.RoundTrip(out, &ids); err != nil {
		return nil, err
	}

	return ids, nil
}

// addTokenEditPermissions adds the edit permissions for the api token to the object
func addTokenEditPermissions(ctx context.Context, m generated.Mutation, oID string, objectType string) error {
	// get auth info from context
	ac, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to get subject id from context, cannot update token permissions")

		return err
	}

	req := fgax.TupleRequest{
		SubjectID:   ac.SubjectID,
		SubjectType: auth.GetAuthzSubjectType(ctx),
		Relation:    fgax.CanEdit,
		ObjectID:    oID,
		ObjectType:  objectType,
	}

	logx.FromContext(ctx).Debug().Interface("request", req).
		Msg("creating edit tuples for api token")

	if _, err := utils.AuthzClient(ctx, m).WriteTupleKeys(ctx, []fgax.TupleKey{fgax.GetTupleKey(req)}, nil); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create relationship tuple")

		return ErrInternalServerError
	}

	return nil
}

// getOrgMemberID gets the org member id for the user in the organization if they are a member
func getOrgMemberID(ctx context.Context, m utils.GenericMutation, userID string, orgID string) (string, error) {
	// ensure user is a member of the organization
	orgMemberID, err := m.Client().OrgMembership.Query().
		Where(orgmembership.UserID(userID)).
		Where(orgmembership.OrganizationID(orgID)).
		OnlyID(ctx)
	if err != nil || orgMemberID == "" {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get org membership, cannot add user to group")

		return "", ErrUserNotInOrg
	}

	return orgMemberID, nil
}

// addUserRelation adds a relation to fga based on the authenticated user context, mutation
// and relation type provided
func addUserRelation(ctx context.Context, m generated.Mutation, relation string) error {
	mut, _ := m.(utils.GenericMutation)

	objID, exists := mut.ID()
	if !exists {
		return nil
	}

	ac, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to get subject id from context, cannot update token permissions")

		return err
	}

	req := fgax.TupleRequest{
		SubjectID:   ac.SubjectID,
		SubjectType: auth.GetAuthzSubjectType(ctx),
		Relation:    relation,
		ObjectID:    objID,
		ObjectType:  GetObjectTypeFromEntMutation(m),
	}

	logx.FromContext(ctx).Debug().Interface("request", req).
		Msg("creating can_view tuple for user")

	if _, err := utils.AuthzClient(ctx, m).WriteTupleKeys(ctx, []fgax.TupleKey{fgax.GetTupleKey(req)}, nil); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create can_view relationship tuple for user")

		return ErrInternalServerError
	}

	return nil
}

// addUserCanViewRelation adds the can_view relation to fga based on the authenticated user context and mutation provided
func addUserCanViewRelation(ctx context.Context, m generated.Mutation) error {
	return addUserRelation(ctx, m, fgax.CanView)
}
