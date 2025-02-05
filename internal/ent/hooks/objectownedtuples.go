package hooks

import (
	"context"
	"strings"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
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

			objectID, err := GetObjectIDFromEntValue(retVal)
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
					ObjectType:  GetObjectTypeFromEntMutation(m), // this is the object type being created
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
				if _, err := utils.AuthzClient(ctx, m).WriteTupleKeys(ctx, addTuples, removeTuples); err != nil {
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

// HookGroupPermissionsTuples is a hook that adds group permissions tuples for the object being created
// this is the reverse edge of the object owned tuples, meaning these run on group mutations
// whereas the other hooks run on the object mutations
func HookGroupPermissionsTuples() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// ensure the user has access to the object they are trying to give permissions to
			if err := checkAccessForEdges(ctx, m); err != nil {
				return nil, err
			}

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

				log.Debug().Interface("tuples", addTuples).Msg("added tuples")
				log.Debug().Interface("tuples", removeTuples).Msg("removed tuples")
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

			// write the tuples to the authz service
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

				log.Debug().Interface("tuples", addTuples).Msg("added tuples")
				log.Debug().Interface("tuples", removeTuples).Msg("removed tuples")
			}

			return retVal, err
		},
		)
	}
}

// checkAccessForEdges checks if the user has access to the object they are trying to give permissions to
// by looking at all the AddedEdges and RemovedEdges
func checkAccessForEdges(ctx context.Context, m ent.Mutation) error {
	if _, allow := privacy.DecisionFromContext(ctx); allow {
		return nil
	}

	addedEdges := m.AddedEdges()
	removedEdges := m.RemovedEdges()

	// nothing to check
	if addedEdges == nil && removedEdges == nil {
		return nil
	}

	// check added edges
	if err := checkEdgesEditAccess(ctx, m, addedEdges); err != nil {
		return err
	}

	// check removed edges
	return checkEdgesEditAccess(ctx, m, removedEdges)
}

// checkEdgesEditAccess takes a list of edges and looks for the permissions edges to confirm the user has edit access
func checkEdgesEditAccess(ctx context.Context, m ent.Mutation, edges []string) error {
	actorID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("unable to get user id from context")

		return err
	}

	for _, edge := range edges {
		objectType, _, ok := isPermissionsEdge(edge)
		if !ok {
			continue
		}

		ids := m.AddedIDs(edge)

		for _, id := range ids {
			idStr, ok := id.(string)
			if !ok {
				log.Warn().Interface("id", id).Msg("id is not a string, unable to check access")

				continue
			}

			ac := fgax.AccessCheck{
				Relation:    fgax.CanEdit,
				ObjectID:    idStr,
				ObjectType:  fgax.Kind(objectType),
				SubjectID:   actorID,
				SubjectType: auth.GetAuthzSubjectType(ctx),
			}

			if allow, err := utils.AuthzClient(ctx, m).CheckAccess(ctx, ac); err != nil || !allow {
				return generated.ErrPermissionDenied
			}
		}
	}

	return nil
}

// getTuplesForGroupEdgeChanges gets the tuples for a group edge based on the mutation
func getTuplesForGroupEdgeChanges(m ent.Mutation, subjectID string) (addTuples []fgax.TupleKey, removeTuples []fgax.TupleKey, err error) {
	// check edges for added edges
	if m.AddedEdges() != nil {
		addTuples = getTuplesForGroupEdge(m, m.AddedEdges(), subjectID)
	}

	// check edges for added edges
	if m.RemovedEdges() != nil {
		removeTuples = getTuplesForGroupEdge(m, m.RemovedEdges(), subjectID)
	}

	return addTuples, removeTuples, nil
}

// getTuplesForGroupEdge gets the tuples for edges that were added or removed, it will take in the edges
// that were changed and the subject id of the group and return the tuples
// the subject id in this case should be the group id from the mutation
func getTuplesForGroupEdge(m ent.Mutation, edges []string, subjectID string) (tuples []fgax.TupleKey) {
	for _, edge := range edges {
		// looking for edges like `program_editors` or `program_viewers`
		objectType, relation, ok := isPermissionsEdge(edge)
		if !ok {
			continue
		}

		ids := m.AddedIDs(edge)

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
		subjectID, err := auth.GetUserIDFromContext(ctx)
		if err != nil {
			return err
		}

		// does the user making the request have access to the edge object
		ac := fgax.AccessCheck{
			Relation:    fgax.CanEdit,
			SubjectType: auth.GetAuthzSubjectType(ctx),
			SubjectID:   subjectID,
			ObjectID:    objectID,
			ObjectType:  fgax.Kind(objectType),
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
