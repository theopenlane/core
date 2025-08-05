package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx/accessmap"
)

// Task holds the schema definition for the Task entity
type Task struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTask is the name of the Task schema.
const SchemaTask = "task"

// Name returns the name of the Task schema.
func (Task) Name() string {
	return SchemaTask
}

// GetType returns the type of the Task schema.
func (Task) GetType() any {
	return Task.Type
}

// PluralName returns the plural name of the Task schema.
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
			GoType(models.DateTime{}).
			Comment("the due date of the task").
			Annotations(
				entgql.OrderField("due"),
			).
			Nillable().
			Optional(),
		field.Time("completed").
			GoType(models.DateTime{}).
			Comment("the completion date of the task").
			Annotations(
				entgql.OrderField("completed"),
			).
			Nillable().
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
			newObjectOwnedMixin[generated.Task](t,
				withParents(InternalPolicy{}, Procedure{}, Control{}, Subcontrol{}, ControlObjective{}, Program{}, Risk{}, Asset{}, Scan{}),
				withOrganizationOwner(true),
			),
		},
	}.getMixins(t)
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
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			name:       "assignee",
			t:          User.Type,
			field:      "assignee_id",
			ref:        "assignee_tasks",
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: t,
			name:       "comments",
			t:          Note.Type,
			comment:    "conversations related to the task",
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Note{}.Name()),
			},
		}),
		defaultEdgeFromWithPagination(t, Group{}),
		defaultEdgeFromWithPagination(t, InternalPolicy{}),
		defaultEdgeFromWithPagination(t, Procedure{}),
		defaultEdgeFromWithPagination(t, Control{}),
		defaultEdgeFromWithPagination(t, Subcontrol{}),
		defaultEdgeFromWithPagination(t, ControlObjective{}),
		defaultEdgeFromWithPagination(t, Program{}),
		defaultEdgeFromWithPagination(t, Risk{}),
		defaultEdgeToWithPagination(t, Evidence{}),
	}
}

func (Task) Features() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
		models.CatalogPolicyManagementAddon,
		models.CatalogRiskManagementAddon,
		models.CatalogBaseModule,
		models.CatalogEntityManagementModule,
	}
}

// Annotations of the Task
func (t Task) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		entx.Exportable{},
	}
}

// Interceptors of the Task
func (t Task) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorFeatures(t.Features()...),
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
func (t Task) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.DenyIfMissingAllFeatures("task", t.Features()...),
			entfga.CheckEditAccess[*generated.TaskMutation](),
		),
	)
}
