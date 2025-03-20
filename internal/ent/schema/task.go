package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"
)

// Task holds the schema definition for the Task entity
type Task struct {
	SchemaFuncs

	ent.Schema
}

const SchemaTask = "task"

func (Task) Name() string {
	return SchemaTask
}

func (Task) GetType() any {
	return Task.Type
}

func (Task) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTask)
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
func (t Task) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "TSK",
		additionalMixins: []ent.Mixin{
			NewObjectOwnedMixin(ObjectOwnedMixin{
				FieldNames:            []string{"internal_policy_id", "procedure_id", "control_id", "subcontrol_id", "control_objective_id", "program_id"},
				WithOrganizationOwner: true,
				Ref:                   t.PluralName(),
			}),
		},
	}.getMixins()
}

// Edges of the Task
func (t Task) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			name:       "assigner",
			t:          User.Type,
			field:      "assigner_id",
			ref:        "assigner_tasks",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			name:       "assignee",
			t:          User.Type,
			field:      "assignee_id",
			ref:        "assignee_tasks",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: t,
			name:       "comments",
			t:          Note.Type,
			comment:    "conversations related to the task",
		}),
		defaultEdgeFromWithPagination(t, Group{}),
		defaultEdgeFromWithPagination(t, InternalPolicy{}),
		defaultEdgeFromWithPagination(t, Procedure{}),
		defaultEdgeFromWithPagination(t, Control{}),
		defaultEdgeFromWithPagination(t, Subcontrol{}),
		defaultEdgeFromWithPagination(t, ControlObjective{}),
		defaultEdgeFromWithPagination(t, Program{}),
		defaultEdgeToWithPagination(t, Evidence{}),
	}
}

// Annotations of the Task
func (Task) Annotations() []schema.Annotation {
	return []schema.Annotation{
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
