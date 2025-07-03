package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
)

// Export holds the schema definition for export records used for exporting various content types.
type Export struct {
	SchemaFuncs
	ent.Schema
}

const SchemaExport = "export"

func (Export) Name() string       { return SchemaExport }
func (Export) GetType() any       { return Export.Type }
func (Export) PluralName() string { return pluralize.NewClient().Plural(SchemaExport) }

// Fields returns export fields.
func (Export) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("export_type").
			Comment("the type of export, e.g., control, policy, etc.").
			Annotations(
				entgql.OrderField("export_type"),
			).
			GoType(enums.ExportType("")),
		field.String("item_id").
			Comment("the id of the item to be exported. this could be a control, policy, etc.").
			Immutable().
			NotEmpty(),
		field.Enum("status").
			Comment("the status of the export, e.g., pending, ready, failed").
			GoType(enums.ExportStatus("")).
			Default(enums.ExportStatusPending.String()).
			Annotations(
				entgql.OrderField("status"),
				entgql.Skip(entgql.SkipMutationCreateInput),
			),
		field.String("requestor_id").
			Comment("the user who initiated the export").
			Immutable().
			Optional().
			NotEmpty().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput),
			),
		field.String("file_id").
			Comment("the id of the generated file").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput),
			),
	}
}

// Edges of the Export
func (e Export) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(e, Event{}),
		defaultEdgeToWithPagination(e, File{}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: e,
			edgeSchema: File{},
			field:      "file_id",
			required:   false,
		}),
	}
}

// Mixin of the Export
func (e Export) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(e),
		},
	}.getMixins()
}

// Annotations of the Export
func (Export) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.Features("base"),
	}
}

// Policy of the Export
func (Export) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(),
		),
		policy.WithOnMutationRules(ent.OpCreate,
			privacy.AlwaysAllowRule(),
		),
		policy.WithOnMutationRules(ent.OpUpdate|ent.OpUpdateOne|ent.OpDelete|ent.OpDeleteOne,
			rule.AllowMutationIfSystemAdmin(),
		),
	)
}

// Hooks of the Export
func (Export) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookExport(),
	}
}
