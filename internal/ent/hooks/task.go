package hooks

import (
	"context"
	"slices"
	"strings"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// HookTaskCreate runs on task create mutations to set default values that are not provided
// this will set the assigner to the current user if it is not provided
func HookTaskCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TaskFunc(func(ctx context.Context, m *generated.TaskMutation) (generated.Value, error) {
			if assigner, _ := m.AssignerID(); assigner == "" {
				assigner, err := auth.GetUserIDFromContext(ctx)
				if err != nil {
					return nil, err
				}

				m.SetAssignerID(assigner)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

// HookTaskAssignee runs on task create and update mutations to add and remove the assignee tuple
func HookTaskAssignee() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TaskFunc(func(ctx context.Context, m *generated.TaskMutation) (generated.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// get the assignee from the mutation and create a tuple for it
			// this will allow the assignee to see and edit the task
			if slices.Contains(m.AddedEdges(), "assignee") {
				if assignee, _ := m.AssigneeID(); assignee != "" {
					taskID, ok := m.ID()

					if ok {
						// add the assignee tuple
						addTuple := fgax.GetTupleKey(fgax.TupleRequest{
							SubjectID:   assignee,
							SubjectType: "user",
							ObjectID:    taskID,
							ObjectType:  m.Type(),
							Relation:    "assignee",
						})

						// get the current assignee and remove them
						resp, err := utils.AuthzClientFromContext(ctx).ListUserRequest(ctx, fgax.ListRequest{
							ObjectID:   taskID,
							ObjectType: strings.ToLower(m.Type()),
							Relation:   "assignee",
						})
						if err != nil {
							return nil, err
						}

						currentAssignee := resp.GetUsers()

						// remove the current assignee from the task, this should only be one user but we will loop through it
						// to ensure we remove all assignees
						var deleteTuples []fgax.TupleKey

						for _, user := range currentAssignee {
							deleteTuple := fgax.GetTupleKey(fgax.TupleRequest{
								SubjectID:   user.Object.Id,
								SubjectType: user.Object.Type,
								ObjectID:    taskID,
								ObjectType:  m.Type(),
								Relation:    "assignee",
							})

							deleteTuples = append(deleteTuples, deleteTuple)
						}

						// add the new assignee and remove the old assignee
						if _, err := utils.AuthzClientFromContext(ctx).WriteTupleKeys(ctx, []fgax.TupleKey{addTuple}, deleteTuples); err != nil {
							return nil, err
						}

						log.Debug().Str("task_id", taskID).Str("assignee", assignee).Msg("Added assignee tuple")
						log.Debug().Str("task_id", taskID).Interface("tuples", deleteTuples).Msg("Removed assignee tuples")
					}
				}
			}

			return retVal, nil
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}
