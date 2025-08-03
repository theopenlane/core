package hooks

import (
	"context"
	"errors"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/gertd/go-pluralize"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// HookObjectOwnedTuples is a hook that adds object owned tuples for the object being created
// given a set of parent id fields, it will add the user and parent permissions to the object
// on creation
// by default, it will always add a user permission to the object
// ownerRelation should normally be set to fgax.ParentRelation, but in some cases
// this is set to owner to account for different inherited permissions from parent objects
// vs. the user/service owner of the object (see notes as an example)
func HookObjectOwnedTuples(parents []string, ownerRelation string) ent.Hook {
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

			// add user permissions to the object on creation
			if m.Op() == ent.OpCreate {
				subjectID, err := auth.GetSubjectIDFromContext(ctx)
				if err != nil {
					return nil, err
				}

				// add user permissions to the object as the parent on creation
				userTuple := fgax.GetTupleKey(fgax.TupleRequest{
					SubjectID:   subjectID,
					SubjectType: auth.GetAuthzSubjectType(ctx),
					ObjectID:    objectID,                        // this is the object id being created
					ObjectType:  GetObjectTypeFromEntMutation(m), // this is the object type being created
					Relation:    ownerRelation,
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
				if _, err := utils.AuthzClient(ctx, m).WriteTupleKeys(ctx, addTuples, removeTuples); err != nil {
					return nil, err
				}

				if len(addTuples) != 0 {
					zerolog.Ctx(ctx).Debug().Interface("tuples", addTuples).Msg("added object permissions")
				}

				if len(removeTuples) != 0 {
					zerolog.Ctx(ctx).Debug().Interface("tuples", removeTuples).Msg("removed object permissions")
				}
			}

			return retVal, err
		},
		)
	}
}

// HookGroupPermissionsTuples is a hook that adds group permissions tuples for the object being created
// this is the reverse edge of the object owned tuples, meaning these run on group mutations
// whereas the other hooks run on the object mutations
func HookGroupPermissionsTuples() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			mut := m.(*generated.GroupMutation)

			subjectID, ok := mut.ID()
			if !ok {
				subjectID, err = GetObjectIDFromEntValue(retVal)
				if err != nil {
					return nil, err
				}
			}

			addTuples, removeTuples, err := getTuplesForGroupEdgeChanges(m, subjectID)
			if err != nil {
				return nil, err
			}

			// write the tuples to the authz service
			if len(addTuples) != 0 || len(removeTuples) != 0 {
				if _, err := utils.AuthzClient(ctx, m).WriteTupleKeys(ctx, addTuples, removeTuples); err != nil {
					return nil, err
				}

				zerolog.Ctx(ctx).Debug().Interface("tuples", addTuples).Msg("added tuples")
				zerolog.Ctx(ctx).Debug().Interface("tuples", removeTuples).Msg("removed tuples")
			}

			return retVal, err
		},
		)
	}
}

// HookRelationTuples is a hook that adds tuples for the object being created
// the objects input is a map of object id fields to the object type
// these tuples based are based on the direct relation, e.g. a group#member to another object
// this is the reverse of the HookGroupPermissionsTuples
func HookRelationTuples(objects map[string]string, relation fgax.Relation) ent.Hook {
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

			// write the tuples to the authz service, the permissions to the edges
			// were already checked by the global edge permissions hook
			if len(addTuples) != 0 || len(removeTuples) != 0 {
				// first check permissions, if the user doesn't have access
				// these is the easiest place to check and roll back the transaction
				if err := checkAccessToObjectsFromTuples(ctx, m, addTuples); err != nil {
					return nil, err
				}

				if err := checkAccessToObjectsFromTuples(ctx, m, removeTuples); err != nil {
					return nil, err
				}

				if _, err := utils.AuthzClient(ctx, m).WriteTupleKeys(ctx, addTuples, removeTuples); err != nil {
					return nil, err
				}

				zerolog.Ctx(ctx).Debug().Interface("tuples", addTuples).Msg("added tuples")
				zerolog.Ctx(ctx).Debug().Interface("tuples", removeTuples).Msg("removed tuples")
			}

			return retVal, err
		},
		)
	}
}

// checkAccessForEdges checks if the user has access to the object they are trying to give permissions to
// by looking at all the AddedEdges and RemovedEdges
func checkAccessForEdges(ctx context.Context, m ent.Mutation) error {
	addedEdges := m.AddedEdges()
	removedEdges := m.RemovedEdges()

	// check added edges
	if len(addedEdges) > 0 {
		if err := checkEdgesForAddedAccess(ctx, m, addedEdges); err != nil {
			return err
		}
	}

	// check removed edges
	if len(removedEdges) > 0 {
		if err := checkEdgesForRemovedAccess(ctx, m, removedEdges); err != nil {
			return err
		}
	}

	return nil
}

// checkEdgesForAddedAccess checks if the user has access to the object they are trying to add permissions to
// it will look at the AddedEdges and check if the user has access to the object
func checkEdgesForAddedAccess(ctx context.Context, m ent.Mutation, edges []string) error {
	return checkEdgesEditAccess(ctx, m, edges, true)
}

// checkEdgesForRemovedAccess checks if the user has access to the object they are trying to remove permissions from
func checkEdgesForRemovedAccess(ctx context.Context, m ent.Mutation, edges []string) error {
	return checkEdgesEditAccess(ctx, m, edges, false)
}

// checkEdgesEditAccess takes a list of edges and looks for the permissions edges to confirm the user has edit access
func checkEdgesEditAccess(ctx context.Context, m ent.Mutation, edges []string, added bool) error {
	actor, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("unable to get user id from context")

		return err
	}

	for _, edge := range edges {
		relationCheck := fgax.CanEdit
		edgeMap := mapEdgeToObjectType(m.Type(), edge)
		if edgeMap.SkipEditCheck {
			if edgeMap.CheckViewAccess {
				relationCheck = fgax.CanView
			} else {
				// not required to check the edge, so skip
				continue
			}
		}

		var ids []ent.Value
		if added {
			ids = m.AddedIDs(edge)
		} else {
			ids = m.RemovedIDs(edge)
		}

		for _, id := range ids {
			idStr, ok := id.(string)
			if !ok {
				zerolog.Ctx(ctx).Warn().Interface("id", id).Msg("id is not a string, unable to check access")

				continue
			}

			if idStr == "" {
				zerolog.Ctx(ctx).Debug().Msg("id is empty, nothing to check, validation will catch this later")

				continue
			}

			if edgeMap.ObjectType == organization.Label && edge != "organizations" {
				orgID, err := auth.GetOrganizationIDFromContext(ctx)
				if err != nil {
					zerolog.Ctx(ctx).Error().Err(err).Msg("unable to get organization id from context")

					return err
				}

				if err := ensureObjectInOrganization(ctx, m, edge, idStr, orgID); err != nil {
					zerolog.Ctx(ctx).Error().Err(err).Msg("object is not part of the organization")

					return err
				}

				idStr = orgID
			}

			ac := fgax.AccessCheck{
				Relation:    relationCheck,
				ObjectID:    idStr,
				ObjectType:  fgax.Kind(edgeMap.ObjectType),
				SubjectID:   actor.SubjectID,
				SubjectType: auth.GetAuthzSubjectType(ctx),
				Context:     utils.NewOrganizationContextKey(actor.SubjectEmail),
			}

			if allow, err := utils.AuthzClient(ctx, m).CheckAccess(ctx, ac); err != nil || !allow {
				log.Error().Err(err).Str("edge", edge).Str("relation", ac.Relation).Msg("user does not have access to the object for edge permissions")

				return generated.ErrPermissionDenied
			}
		}

	}

	return nil
}

func mapEdgeToObjectType(schema, edge string) generated.EdgeAccess {
	schemaType := strcase.SnakeCase(schema)

	schemaMap, ok := generated.EdgeAccessMap[schemaType]
	if !ok {
		log.Error().Str("schema", schema).Msg("schema not found in edge access map")
		return generated.EdgeAccess{}
	}

	edgeAccess, ok := schemaMap[edge]
	if !ok {
		log.Error().Str("edge", edge).Msg("edge not found in edge access map for schema")

		return generated.EdgeAccess{}
	}

	return edgeAccess
}

func ensureObjectInOrganization(ctx context.Context, m ent.Mutation, edge string, objectID, orgID string) error {
	// also ensure the id is part of the organization
	mut, ok := m.(GenericMutation)
	if !ok {
		zerolog.Ctx(ctx).Error().Msg("unable to determine access")
		return privacy.Deny
	}

	// check view access to the organization instead
	if edge == organization.Label {
		if err := rule.CheckCurrentOrgAccess(ctx, m, fgax.CanView); errors.Is(err, privacy.Allow) {
			return nil
		}

		zerolog.Ctx(ctx).Error().Msg("user does not have access to the organization")

		return privacy.Deny

	}

	table := pluralize.NewClient().Plural(edge)
	query := "SELECT EXISTS (SELECT 1 FROM " + table + " WHERE id = $1 and (owner_id = $2 or owner_id IS NULL))"

	var rows sql.Rows
	if err := mut.Client().Driver().Query(ctx, query, []any{objectID, orgID}, &rows); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("failed to check for object in organization")

		return privacy.Deny
	}

	defer rows.Close()

	if rows.Next() {
		var exists bool
		if err := rows.Scan(&exists); err == nil && exists {
			return nil
		}
	}

	// fall back to deny if the object is not in the organization
	return privacy.Deny
}

// getTuplesForGroupEdgeChanges gets the tuples for a group edge based on the mutation
func getTuplesForGroupEdgeChanges(m ent.Mutation, subjectID string) (addTuples []fgax.TupleKey, removeTuples []fgax.TupleKey, err error) {
	// check edges for added edges
	if m.AddedEdges() != nil {
		addTuples = getAddedTuplesForGroupEdge(m, m.AddedEdges(), subjectID)
	}

	// check edges for added edges
	if m.RemovedEdges() != nil {
		removeTuples = getRemovedTuplesForGroupEdge(m, m.RemovedEdges(), subjectID)
	}

	return addTuples, removeTuples, nil
}

// getAddedTuplesForGroupEdge gets the tuples for edges that were added, it will take in the edges
// that were added and the subject id of the group and return the tuples
func getAddedTuplesForGroupEdge(m ent.Mutation, edges []string, subjectID string) (tuples []fgax.TupleKey) {
	return getTuplesForGroupEdge(m, edges, subjectID, true)
}

// getRemovedTuplesForGroupEdge gets the tuples for edges that were removed, it will take in the edges
// that were removed and the subject id of the group and return the tuples
func getRemovedTuplesForGroupEdge(m ent.Mutation, edges []string, subjectID string) (tuples []fgax.TupleKey) {
	return getTuplesForGroupEdge(m, edges, subjectID, false)
}

// getTuplesForGroupEdge gets the tuples for edges that were added or removed, it will take in the edges
// that were changed and the subject id of the group and return the tuples
// the subject id in this case should be the group id from the mutation
func getTuplesForGroupEdge(m ent.Mutation, edges []string, subjectID string, added bool) (tuples []fgax.TupleKey) {
	for _, edge := range edges {
		// looking for edges like `program_editors` or `program_viewers`
		objectType, relation, ok := isPermissionsEdge(edge)
		if !ok {
			continue
		}

		var ids []ent.Value
		if added {
			ids = m.AddedIDs(edge)
		} else {
			ids = m.RemovedIDs(edge)
		}

		for _, id := range ids {
			idStr, ok := id.(string)
			if !ok {
				log.Warn().Interface("id", id).Msg("id is not a string")

				continue
			}

			tr := fgax.TupleRequest{
				SubjectID:       subjectID,           // this is the group id
				SubjectType:     generated.TypeGroup, // this is the group type
				SubjectRelation: fgax.MemberRelation, // this is the relation between the group and the object
				ObjectID:        idStr,               // this the edge object id
				ObjectType:      objectType,          // this is the edge object type
				Relation:        relation.String(),
			}

			tuples = append(tuples, fgax.GetTupleKey(tr))
		}
	}

	return
}

// isPermissionsEdge checks if the edge is a permissions edge
// and returns the object type, relation for the edge, and true if it is a permissions edge
func isPermissionsEdge(edge string) (string, fgax.Relation, bool) {
	switch {
	case strings.HasSuffix(edge, "editors"):
		return strings.TrimSuffix(edge, "_editors"), fgax.EditorRelation, true
	case strings.HasSuffix(edge, "viewers"):
		return strings.TrimSuffix(edge, "_viewers"), fgax.ViewerRelation, true
	case strings.HasSuffix(edge, "blocked_groups"):
		return strings.TrimSuffix(edge, "_blocked_groups"), fgax.BlockedRelation, true
	}

	return "", "", false
}

// checkAccessToObjectsFromTuples checks if the user has access to the object they are trying to give permissions to
// using the tuple structs that are about to be written
func checkAccessToObjectsFromTuples(ctx context.Context, m ent.Mutation, tuples []fgax.TupleKey) error {
	for _, tuple := range tuples {
		// subject is the group that the permissions are being added to
		// this is the reverse edge
		objectID := tuple.Subject.Identifier
		objectType := string(tuple.Subject.Kind)

		if _, allow := privacy.DecisionFromContext(ctx); allow {
			return nil
		}

		// get the user id or service id from the context
		subject, err := auth.GetAuthenticatedUserFromContext(ctx)
		if err != nil {
			return err
		}

		// does the user making the request have access to the edge object
		ac := fgax.AccessCheck{
			Relation:    fgax.CanEdit,
			SubjectType: auth.GetAuthzSubjectType(ctx),
			SubjectID:   subject.SubjectID,
			ObjectID:    objectID,
			ObjectType:  fgax.Kind(objectType),
			Context:     utils.NewOrganizationContextKey(subject.SubjectEmail),
		}

		access, err := utils.AuthzClient(ctx, m).CheckAccess(ctx, ac)
		if err != nil {
			return err
		}

		// return an error if the user does not have access
		if !access {
			return generated.ErrPermissionDenied
		}
	}

	return nil
}
