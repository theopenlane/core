package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
)

type Report struct {
	SchemaFuncs

	ent.Schema
}

const SchemaReport = "report"

func (Report) Name() string {
	return SchemaReport
}

func (Report) GetType() any {
	return Report.Type
}

func (Report) PluralName() string {
	return pluralize.NewClient().Plural(SchemaReport)
}

func (Report) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			),
		field.Text("query").
			Comment("the query for this report").
			NotEmpty(),
		field.JSON("variables", map[string]any{}).
			Comment("variables needed to pass into the query if needed").
			Optional(),
		field.Enum("status").
			GoType(enums.ReportStatus("")).
			Default(enums.ReportStatusDraft.String()).
			Annotations(
				entgql.OrderField("STATUS"),
			),
		field.Enum("visibility").
			GoType(enums.ReportVisibility("")).
			Annotations(
				entgql.OrderField("VISIBILITY"),
			).
			Default(enums.ReportVisibilityPrivate.String()),
		field.String("user_id").
			Immutable().
			Comment("the user who created this report").
			NotEmpty(),
	}
}

func (r Report) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: r,
			edgeSchema: User{},
			field:      "user_id",
			required:   true,
			immutable:  true,
		}),
		// edgeToWithPagination(&edgeDefinition{
		// 	fromSchema: r,
		// 	name:       "shared_users",
		// 	t:          User.Type,
		// 	annotations: []schema.Annotation{
		// 		entgql.QueryField(),
		// 		entgql.MultiOrder(),
		// 	},
		// }),
	}
}

func (r Report) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(r),
		},
	}.getMixins(r)
}

func (Report) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogAnyModule,
	}
}

func (Report) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookReport(),
	}
}
