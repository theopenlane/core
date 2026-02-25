package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/entx/history"

	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
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
			Immutable().
			Annotations(
				entgql.OrderField("export_type"),
			).
			GoType(enums.ExportType("")),
		field.Enum("format").
			Comment("the format of export, e.g., csv and others").
			Default(enums.ExportFormatCsv.String()).
			Immutable().
			Annotations(
				entgql.OrderField("format"),
			).
			GoType(enums.ExportFormat("")),
		field.Enum("status").
			Comment("the status of the export, e.g., pending, ready, failed").
			GoType(enums.ExportStatus("")).
			Default(enums.ExportStatusPending.String()).
			Annotations(
				entgql.OrderField("status"),
				entgql.Skip(entgql.SkipMutationCreateInput),
			),
		field.JSON("fields", []string{}).
			Comment("the specific fields to include in the export (defaults to only the id if not provided)").
			Default([]string{"id"}).
			Optional().
			Immutable(),

		field.String("filters").
			Comment("the specific filters to run against the exported data. This should be a well formatted graphql query").
			Optional().
			Immutable(),

		field.String("error_message").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput),
			).
			Comment("if we try to export and it fails, the error message will be stored here"),
	}
}

// Edges of the Export
func (e Export) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(e, Event{}),
		defaultEdgeToWithPagination(e, File{}),
	}
}

// Mixin of the Export
func (e Export) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags:      true,
		includeRequestor: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(e,
				withSkipForSystemAdmin(true)),
		},
	}.getMixins(e)
}

func (Export) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Annotations of the Export
func (e Export) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
	}
}

// Policy of the Export
func (e Export) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			policy.CheckOrgReadAccess(),
		),
		policy.WithMutationRules(
			policy.AllowCreate(),
		),
		policy.WithOnMutationRules(ent.OpUpdate|ent.OpUpdateOne|ent.OpDelete|ent.OpDeleteOne,
			policy.CheckOrgWriteAccess(),
		),
	)
}

// Hooks of the Export
func (Export) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookExport(),
		hooks.HookUserCanViewTuple(),
	}
}
