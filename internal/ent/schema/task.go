package schema

import (
	"context"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/pkg/enums"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"
)

// Task holds the schema definition for the Task entity
type Task struct {
	ent.Schema
}

// Fields of the Task
func (Task) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").
			Comment("the title of the task").
			NotEmpty(),
		field.String("description").
			Comment("the description of the task").
			Optional(),
		field.JSON("details", map[string]interface{}{}).
			Comment("the details of the task").
			Optional(),
		field.Enum("status").
			GoType(enums.TaskStatus("")).
			Comment("the status of the task").
			Default(enums.TaskStatusOpen.String()),
		field.Time("due").
			Comment("the due date of the task").
			Optional(),
		field.Time("completed").
			Comment("the completion date of the task").
			Optional(),
	}
}

// Mixin of the Task
func (Task) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
		emixin.TagMixin{},
		NewObjectOwnedMixin(ObjectOwnedMixin{
			FieldNames: []string{"organization_id", "group_id", "policy_id", "procedure_id", "control_id", "subcontrol_id", "control_objective_id"},
			Required:   false,
		}),
	}
}

// Edges of the Task
func (Task) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("assigner", User.Type).
			Ref("assigner_tasks").
			Required().
			Unique(),
		edge.From("assignee", User.Type).
			Ref("assignee_tasks").
			Unique(),
		edge.From("organization", Organization.Type).
			Ref("tasks"),
		edge.From("group", Group.Type).
			Ref("tasks"),
		edge.From("policy", InternalPolicy.Type).
			Ref("tasks"),
		edge.From("procedure", Procedure.Type).
			Ref("tasks"),
		edge.From("control", Control.Type).
			Ref("tasks"),
		edge.From("control_objective", ControlObjective.Type).
			Ref("tasks"),
		edge.From("subcontrol", Subcontrol.Type).
			Ref("tasks"),
		edge.From("program", Program.Type).
			Ref("tasks"),
	}
}

// Indexes of the Task
func (Task) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the Task
func (Task) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.Annotations{
			ObjectType:   "task",
			IncludeHooks: false,
		},
	}
}

// Hooks of the Task
func (Task) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookTaskCreate(),
		hooks.HookTaskAssignee(),
	}
}

// Interceptors of the Task
func (Task) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the Task
func (Task) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			privacy.TaskMutationRuleFunc(func(ctx context.Context, m *generated.TaskMutation) error {
				return m.CheckAccessForEdit(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.TaskQueryRuleFunc(func(ctx context.Context, q *generated.TaskQuery) error {
				return q.CheckAccess(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
	}
}
