package hooks

import (
	"context"
	"slices"

	"entgo.io/ent"
	openfga "github.com/openfga/go-sdk"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/ent/generated/task"
	"github.com/theopenlane/ent/privacy/utils"
	"github.com/theopenlane/shared/logx"
)

const (
	// assigneeField is the field name for the assignee tuple
	assigneeField = "assignee"
	// assignerField is the field name for the assigner tuple, this defaults to the creator of the task but can be changed
	assignerField = "assigner"
)

// HookTaskCreate runs on task create mutations to set default values that are not provided
// this will set the assigner to the current user if it is not provided
func HookTaskCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TaskFunc(func(ctx context.Context, m *generated.TaskMutation) (generated.Value, error) {
			if assigner, _ := m.AssignerID(); assigner == "" {
				// if the assigner is not provided, set it to the current user if not using an API token
				if !auth.IsAPITokenAuthentication(ctx) {
					assigner, err := auth.GetSubjectIDFromContext(ctx)
					if err != nil {
						return nil, err
					}

					m.SetAssignerID(assigner)
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

// HookTaskPermissions runs on task create and update mutations to add and remove the assignee tuple
func HookTaskPermissions() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.TaskFunc(func(ctx context.Context, m *generated.TaskMutation) (generated.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// get the assignee from the mutation and create a tuple for it
			// this will allow the assignee to see and edit the task
			taskID, _ := m.ID()
			assignee, ok := m.AssigneeID()

			orgID, ownerOk := m.OwnerID()
			if !ownerOk {
				task, err := m.Client().Task.Query().Where(task.ID(taskID)).Select(ownerFieldName).Only(ctx)
				if err != nil {
					return nil, err
				}

				orgID = task.OwnerID
			}

			if ok || m.AssigneeCleared() || slices.Contains(m.RemovedEdges(), assigneeField) {
				// ensure the assignee is a member of the organization
				if assignee != "" {
					if _, err := getOrgMemberID(ctx, m, assignee, orgID); err != nil {
						return nil, err
					}
				}

				// update the assignee tuple
				if err := updateTaskAssigneeTuples(ctx, m, assignee, taskID); err != nil {
					return nil, err
				}
			}

			assigner, ok := m.AssignerID()
			if ok || m.AssignerCleared() || slices.Contains(m.RemovedEdges(), assignerField) {
				// ensure the assigner is a member of the organization
				if assigner != "" {
					if _, err := getOrgMemberID(ctx, m, assigner, orgID); err != nil {
						return nil, err
					}
				}

				// update the assigner tuple
				if err := updateTaskAssignerTuples(ctx, m, assigner, taskID); err != nil {
					return nil, err
				}
			}

			return retVal, nil
		})
	},
		hook.And(
			hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne),
			hook.Or(
				hook.HasFields("assignee_id"),
				hook.HasFields("assigner_id"),
				hook.HasAddedFields("assignee_id"),
				hook.HasAddedFields("assigner_id"),
				hook.HasClearedFields("assignee_id"),
				hook.HasClearedFields("assigner_id"),
			),
		),
	)
}

// updateTaskAssigneeTuples will add the new user tuple for the relation (assignee) and remove all the old assignee tuples
func updateTaskAssigneeTuples(ctx context.Context, m *generated.TaskMutation, newUser, taskID string) error {
	return updateTaskTuples(ctx, m, newUser, assigneeField, taskID)
}

// updateTaskAssignerTuples will add the new user tuple for the relation (assigner) and remove all the old assigner tuples
func updateTaskAssignerTuples(ctx context.Context, m *generated.TaskMutation, newUser, taskID string) error {
	return updateTaskTuples(ctx, m, newUser, assignerField, taskID)
}

// updateTaskTuples will add the new user tuple for the relation (assignee or assigner) and remove all the old assignee or assigner tuples
// the removal should just be a single user, but allowing for multiple users to be removed incase something went wrong with a previous mutation
func updateTaskTuples(ctx context.Context, m *generated.TaskMutation, newUser, relation, taskID string) error {
	resp, err := utils.AuthzClient(ctx, m).ListUserRequest(ctx, fgax.ListRequest{
		ObjectID:   taskID,
		ObjectType: GetObjectTypeFromEntMutation(m),
		Relation:   relation,
	})
	if err != nil {
		return err
	}

	oldUsers := resp.GetUsers()

	addTuples := []fgax.TupleKey{}

	if newUser != "" && !slices.ContainsFunc(oldUsers, func(u openfga.User) bool {
		return u.Object.Id == newUser
	}) {
		addTuple := fgax.GetTupleKey(fgax.TupleRequest{
			SubjectID:   newUser,
			SubjectType: "user",
			ObjectID:    taskID,
			ObjectType:  GetObjectTypeFromEntMutation(m),
			Relation:    relation,
		})

		addTuples = append(addTuples, addTuple)
	}

	deleteTuples := []fgax.TupleKey{}

	for _, oldUser := range oldUsers {
		// skip if its a duplicate of the new user
		if newUser == oldUser.Object.Id {
			continue
		}

		deleteTuple := fgax.GetTupleKey(fgax.TupleRequest{
			SubjectID:   oldUser.Object.Id,
			SubjectType: "user",
			ObjectID:    taskID,
			ObjectType:  GetObjectTypeFromEntMutation(m),
			Relation:    relation,
		})

		deleteTuples = append(deleteTuples, deleteTuple)
	}

	// add the new assignee and remove the old assignee
	if _, err := utils.AuthzClientFromContext(ctx).WriteTupleKeys(ctx, addTuples, deleteTuples); err != nil {
		return err
	}

	logx.FromContext(ctx).Debug().Str("task_id", taskID).Str(relation, newUser).Msg("added tuple")
	logx.FromContext(ctx).Debug().Str("task_id", taskID).Interface(relation, oldUsers).Str("relation", relation).Msg("removed task tuples")

	return nil
}
