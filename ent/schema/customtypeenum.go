package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/ent/hooks"
	"github.com/theopenlane/ent/mixin"
	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/ent/validator"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"
	"github.com/theopenlane/shared/models"
)

// CustomTypeEnum holds the schema definition for the CustomTypeEnum entity
type CustomTypeEnum struct {
	SchemaFuncs

	ent.Schema
}

// SchemaCustomTypeEnum is the name of the schema in snake case
const SchemaCustomTypeEnum = "custom_type_enum"

// Name is the name of the schema in snake case
func (CustomTypeEnum) Name() string {
	return SchemaCustomTypeEnum
}

// GetType returns the type of the schema
func (CustomTypeEnum) GetType() any {
	return CustomTypeEnum.Type
}

// PluralName returns the plural name of the schema
func (CustomTypeEnum) PluralName() string {
	return pluralize.NewClient().Plural(SchemaCustomTypeEnum)
}

// Fields of the CustomTypeEnum
func (CustomTypeEnum) Fields() []ent.Field {
	return []ent.Field{
		field.String("object_type").
			Comment("the kind of object the type applies to, for example task").
			NotEmpty().
			Immutable().
			Validate(validateObjectType).
			Annotations(
				entx.FieldSearchable(),
			),
		field.String("field").
			Comment("the field on the object the type applies to, for example kind or category").
			Default("kind").
			Immutable(),
		field.String("name").
			Comment("The name of the enum value, for example evidence request").
			Annotations(entx.FieldSearchable()).
			NotEmpty().
			Immutable().
			SchemaType(map[string]string{
				dialect.Postgres: "citext",
			}),
		field.String("description").
			Comment("The description of the custom type").
			Optional(),
		field.String("color").
			Comment("The color of the tag definition in hex format").
			Validate(validator.HexColorValidator).
			DefaultFunc(defaultRandomColor).
			Optional(),
		field.String("icon").
			Comment("The icon of the custom type enum in SVG format").
			Optional(),
	}
}

// Mixin of the CustomTypeEnum
func (t CustomTypeEnum) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(t),
			mixin.NewSystemOwnedMixin(),
		},
	}.getMixins(t)
}

// Edges of the CustomTypeEnum
func (t CustomTypeEnum) Edges() []ent.Edge {
	return []ent.Edge{
		edgeToWithPagination(&edgeDefinition{
			fromSchema: t,
			ref:        "task_kind",
			edgeSchema: Task{},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: t,
			ref:        "control_kind",
			edgeSchema: Control{},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: t,
			ref:        "subcontrol_kind",
			edgeSchema: Subcontrol{},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: t,
			ref:        "risk_kind",
			edgeSchema: Risk{},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: t,
			ref:        "risk_category",
			name:       "risk_categories",
			t:          Risk.Type,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: t,
			ref:        "internal_policy_kind",
			edgeSchema: InternalPolicy{},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: t,
			ref:        "procedure_kind",
			edgeSchema: Procedure{},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: t,
			ref:        "action_plan_kind",
			edgeSchema: ActionPlan{},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: t,
			ref:        "program_kind",
			edgeSchema: Program{},
		}),
	}
}

// Indexes of the CustomTypeEnum
func (CustomTypeEnum) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", ownerFieldName).
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
		index.Fields("object_type"),
	}
}

// Annotations of the CustomTypeEnum
func (CustomTypeEnum) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the CustomTypeEnum
func (CustomTypeEnum) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookCustomTypeEnumDelete(),
	}
}

// Interceptors of the CustomTypeEnum
func (CustomTypeEnum) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Modules this schema has access to
func (CustomTypeEnum) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Policy of the CustomTypeEnum
func (CustomTypeEnum) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}
