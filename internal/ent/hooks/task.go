package hooks

import (
	"context"
	"slices"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

// HookTaskCreate runs on task create mutations to set default values that are not provided
// this will set the assigner to the current user if it is not provided
func HookTaskCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TaskFunc(func(ctx context.Context, m *generated.TaskMutation) (generated.Value, error) {
			if assigner, _ := m.Assigner(); assigner == "" {
				assigner, err := auth.GetUserIDFromContext(ctx)
				if err != nil {
					return nil, err
				}

				m.SetAssigner(assigner)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

func HookTaskAssignee() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TaskFunc(func(ctx context.Context, m *generated.TaskMutation) (generated.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// get the assignee from the mutation and create a tuple for it
			// this will allow the assignee to see and edit the task
			if slices.Contains(m.Fields(), "assignee") {
				if assignee, _ := m.Assignee(); assignee != "" {
					originalAssignee, _ := m.OldAssignee(ctx)
					if originalAssignee != assignee {
						taskID, ok := m.ID()

						if ok {
							addTuple := fgax.GetTupleKey(fgax.TupleRequest{
								SubjectID:   assignee,
								SubjectType: "user",
								ObjectID:    taskID,
								ObjectType:  m.Type(),
								Relation:    "assignee",
							})

							if _, err := generated.FromContext(ctx).Authz.WriteTupleKeys(ctx, []fgax.TupleKey{addTuple}, nil); err != nil {
								return nil, err
							}

							log.Debug().Str("task_id", taskID).Str("assignee", assignee).Msg("Added assignee tuple")
						}

						// remove the old assignee tuple if it exists
						if originalAssignee != "" {
							removeTuple := fgax.GetTupleKey(fgax.TupleRequest{
								SubjectID:   originalAssignee,
								SubjectType: "user",
								ObjectID:    taskID,
								ObjectType:  m.Type(),
								Relation:    "assignee",
							})

							if _, err := generated.FromContext(ctx).Authz.WriteTupleKeys(ctx, nil, []fgax.TupleKey{removeTuple}); err != nil {
								return nil, err
							}

							log.Debug().Str("task_id", taskID).Str("assignee", originalAssignee).Msg("Removed old assignee tuple")
						}
					}
				}
			}

			return retVal, nil
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}
