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
				entgql.OrderField("title"),
			).
			NotEmpty(),
		field.String("description").
			Annotations(
				entx.FieldSearchable(),
			).
			Comment("the description of the task").
			Optional(),
		field.Text("details").
			Comment("the details of the task").
			Optional(),
		field.Enum("status").
			GoType(enums.TaskStatus("")).
			Comment("the status of the task").
			Annotations(
				entgql.OrderField("STATUS"),
			).
			Default(enums.TaskStatusOpen.String()),
		field.String("category").
			Comment("the category of the task, e.g. evidence upload, risk review, policy review, etc.").
			Annotations(
				entgql.OrderField("category"),
			).
			Optional(),
		field.Time("due").
			Comment("the due date of the task").
			Annotations(
				entgql.OrderField("due"),
			).
			Optional(),
		field.Time("completed").
			Comment("the completion date of the task").
			Annotations(
				entgql.OrderField("completed"),
			).
			Optional(),
		field.String("assignee_id").
			Comment("the id of the user who was assigned the task").
			Optional(),
		field.String("assigner_id").
			Optional().
			Comment("the id of the user who assigned the task, can be left empty if created by the system or a service token"),
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
			FieldNames:            []string{"internal_policy_id", "procedure_id", "control_id", "subcontrol_id", "control_objective_id", "program_id"},
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
			Unique(),
		edge.From("assignee", User.Type).
			Ref("assignee_tasks").
			Field("assignee_id").
			Unique(),
		edge.To("comments", Note.Type).
			Annotations(entgql.RelayConnection()).
			Comment("conversations related to the task"),
		edge.From("group", Group.Type).
			Annotations(entgql.RelayConnection()).
			Ref("tasks"),
		edge.From("internal_policy", InternalPolicy.Type).
			Annotations(entgql.RelayConnection()).
			Ref("tasks"),
		edge.From("procedure", Procedure.Type).
			Annotations(entgql.RelayConnection()).
			Ref("tasks"),
		edge.From("control", Control.Type).
			Annotations(entgql.RelayConnection()).
			Ref("tasks"),
		edge.From("control_objective", ControlObjective.Type).
			Annotations(entgql.RelayConnection()).
			Ref("tasks"),
		edge.From("subcontrol", Subcontrol.Type).
			Annotations(entgql.RelayConnection()).
			Ref("tasks"),
		edge.From("program", Program.Type).
			Annotations(entgql.RelayConnection()).
			Ref("tasks"),
		edge.To("evidence", Evidence.Type).Annotations(entgql.RelayConnection()),
	}
}

// Annotations of the Task
func (Task) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.MultiOrder(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.SelfAccessChecks(),
	}
}

// Hooks of the Task
func (Task) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookTaskCreate(),
		hooks.HookTaskPermissions(),
	}
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
