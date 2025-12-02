package hooks

import (
	"context"
	"strings"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/ent/privacy/utils"
)

type skipCreateUserPermissions func(context.Context, ent.Mutation) bool

// HookObjectOwnedTuples is a hook that adds object owned tuples for the object being created
// given a set of parent id fields, it will add the user and parent permissions to the object
// on creation
// by default, it will always add a user permission to the object
// ownerRelation should normally be set to fgax.ParentRelation, but in some cases
// this is set to owner to account for different inherited permissions from parent objects
// vs. the user/service owner of the object (see notes as an example)
func HookObjectOwnedTuples(parents []string, ownerRelation string, skipCreateUserPermissions skipCreateUserPermissions) ent.Hook {
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

			if skip := skipCreateUserPermissions(ctx, m); !skip {
				// add user permissions to the object on creation
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
					logx.FromContext(ctx).Debug().Interface("tuples", addTuples).Msg("added object permissions")
				}

				if len(removeTuples) != 0 {
					logx.FromContext(ctx).Debug().Interface("tuples", removeTuples).Msg("removed object permissions")
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

				logx.FromContext(ctx).Debug().Interface("tuples", addTuples).Msg("added tuples")
				logx.FromContext(ctx).Debug().Interface("tuples", removeTuples).Msg("removed tuples")
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
				if _, err := utils.AuthzClient(ctx, m).WriteTupleKeys(ctx, addTuples, removeTuples); err != nil {
					return nil, err
				}

				logx.FromContext(ctx).Debug().Interface("tuples", addTuples).Msg("added tuples")
				logx.FromContext(ctx).Debug().Interface("tuples", removeTuples).Msg("removed tuples")
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
		if err := policy.CheckEdgesForAddedAccess(ctx, m, addedEdges); err != nil {
			return err
		}
	}

	// check removed edges
	if len(removedEdges) > 0 {
		if err := policy.CheckEdgesForRemovedAccess(ctx, m, removedEdges); err != nil {
			return err
		}
	}

	return nil
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
