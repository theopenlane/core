package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/entx"
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
			Annotations(
				entx.FieldSearchable(),
			).
			NotEmpty(),
		field.String("description").
			Annotations(
				entx.FieldSearchable(),
			).
			Comment("the description of the task").
			Optional(),
		field.JSON("details", map[string]any{}).
			Comment("the details of the task").
			Optional(),
		field.Enum("status").
			GoType(enums.TaskStatus("")).
			Comment("the status of the task").
			Default(enums.TaskStatusOpen.String()),
		field.Time("due").
			Comment("the due date of the task").
			Optional(),
		field.Enum("priority").
			GoType(enums.Priority("")).
			Comment("the priority of the task").
			Default(enums.PriorityMedium.String()),
		field.Time("completed").
			Comment("the completion date of the task").
			Optional(),
		field.String("assignee_id").
			Comment("the id of the user who was assigned the task").
			Optional(),
		field.String("assigner_id").
			Comment("the id of the user who assigned the task"),
	}
}

// Mixin of the Task
func (Task) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.NewIDMixinWithPrefixedID("TSK"),
		mixin.SoftDeleteMixin{},
		emixin.TagMixin{},
		NewObjectOwnedMixin(ObjectOwnedMixin{
			FieldNames:            []string{"group_id", "policy_id", "procedure_id", "control_id", "subcontrol_id", "control_objective_id", "program_id"},
			WithOrganizationOwner: true,
			Ref:                   "tasks",
		}),
	}
}

// Edges of the Task
func (Task) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("assigner", User.Type).
			Ref("assigner_tasks").
			Field("assigner_id").
			Required().
			Unique(),
		edge.From("assignee", User.Type).
			Ref("assignee_tasks").
			Field("assignee_id").
			Unique(),
		edge.From("group", Group.Type).
			Ref("tasks"),
		edge.From("internal_policy", InternalPolicy.Type).
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
		edge.To("evidence", Evidence.Type),
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
		entfga.SelfAccessChecks(),
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
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.TaskMutation](),
		),
	)
}
