package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/entx/accessmap"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

// WorkflowDefinition stores workflow configurations, both system-provided templates and organization-specific instances
type WorkflowDefinition struct {
	SchemaFuncs
	ent.Schema
}

// schemaWorkflowDefinition is the name of the WorkflowDefinition schema in snake case
const schemaWorkflowDefinition = "workflow_definition"

// Name returns the name of the WorkflowDefinition schema
func (WorkflowDefinition) Name() string {
	return schemaWorkflowDefinition
}

// GetType returns the type of the WorkflowDefinition schema
func (WorkflowDefinition) GetType() any {
	return WorkflowDefinition.Type
}

// PluralName returns the plural name of the WorkflowDefinition schema
func (WorkflowDefinition) PluralName() string {
	return pluralize.NewClient().Plural(schemaWorkflowDefinition)
}

// Fields of the WorkflowDefinition
func (WorkflowDefinition) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("Name of the workflow definition").
			NotEmpty(),
		field.Text("description").
			Comment("Optional description of the workflow definition").
			Optional(),
		field.Enum("workflow_kind").
			Comment("Kind of workflow, e.g. APPROVAL, LIFECYCLE, NOTIFICATION").
			GoType(enums.WorkflowKind("")),
		field.String("schema_type").
			Comment("Type of schema this workflow applies to").
			NotEmpty(),
		field.Int("revision").
			Comment("Revision number for this definition").
			Default(1),
		field.Bool("draft").
			Comment("Whether this definition is a draft").
			Default(true),
		field.Time("published_at").
			Comment("When this definition was published").
			Nillable().
			Optional(),
		field.Int("cooldown_seconds").
			Comment("Suppress duplicate triggers within this window per object/definition").
			Default(0),
		field.Bool("is_default").
			Comment("Whether this is the default workflow for the schema type").
			Default(false),
		field.Bool("active").
			Comment("Whether the workflow definition is active").
			Default(true),
		field.Strings("trigger_operations").
			Comment("Derived: normalized operations from definition for prefiltering; not user editable").
			Optional().
			Default([]string{}),
		field.Strings("trigger_fields").
			Comment("Derived: normalized fields from definition for prefiltering; not user editable").
			Optional().
			Default([]string{}),
		field.JSON("definition_json", models.WorkflowDefinitionDocument{}).
			Comment("Typed document describing triggers, conditions, and actions").
			Optional(),
		field.Strings("tracked_fields").
			Comment("Cached list of fields that should trigger workflow evaluation").
			Optional(),
	}
}

// Edges of the WorkflowDefinition
func (WorkflowDefinition) Edges() []ent.Edge {
	return []ent.Edge{
		edgeToWithPagination(&edgeDefinition{
			fromSchema: WorkflowDefinition{},
			name:       "target_tags",
			edgeSchema: TagDefinition{},
			comment:    "Tags this workflow targets for scoping",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(TagDefinition{}.Name()),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: WorkflowDefinition{},
			name:       "target_groups",
			edgeSchema: Group{},
			comment:    "Groups this workflow targets for scoping",
		}),
	}
}

// Mixin of the WorkflowDefinition.
func (w WorkflowDefinition) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "WFD",
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.WorkflowDefinition](w,
				withOrganizationOwner(true),
			),
			mixin.NewSystemOwnedMixin(),
		},
	}.getMixins(w)
}

// Modules this schema has access to.
func (WorkflowDefinition) Modules() []models.OrgModule {
	return []models.OrgModule{models.CatalogBaseModule}
}

// Annotations of the WorkflowDefinition
func (WorkflowDefinition) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Policy of the WorkflowDefinition.
func (WorkflowDefinition) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(),
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			//			entfga.CheckEditAccess[*generated.WorkflowDefinitionMutation](),
		),
	)
}
