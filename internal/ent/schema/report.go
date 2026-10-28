package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/ent/interceptors"
	"github.com/theopenlane/core/v2/internal/ent/privacy/policy"
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
			Comment("Name of this report").
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			),
		field.Text("query").
			Comment("the saved query to be re-executed or shared").
			NotEmpty(),
		field.JSON("variables", map[string]any{}).
			Comment("the variables that would be needed to completely execute the query").
			Optional(),
		field.Enum("status").
			GoType(enums.ReportStatus("")).
			Default(enums.ReportStatusDraft.String()).
			Annotations(
				entgql.OrderField("STATUS"),
			),
		field.Enum("visibility").
			GoType(enums.ReportVisibility("")).
			Optional().
			Annotations(
				entgql.OrderField("VISIBILITY"),
			),
		field.String("report_owner_id").
			Comment("the id of the user who is responsible for this report").
			Optional().
			Annotations(
				entx.CSVRef().FromColumn("ReportOwnerEmail").
					MatchOn("email"),
			),
	}
}

func (r Report) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("users", User.Type).
			Ref("reports").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
				entgql.RelayConnection(),
				entgql.QueryField(),
				entgql.MultiOrder(),
			).
			Through("members", ReportMembership.Type),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: r,
			name:       "report_owner",
			t:          User.Type,
			field:      "report_owner_id",
			ref:        "reports_owned",
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
	}
}

func (r Report) Mixin() []ent.Mixin {
	return mixinConfig{
		includeRequestor: true,
		prefix:           "RPT",
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(r),
			newGroupPermissionsMixin(),
			newCustomEnumMixin(r),
		},
	}.getMixins(r)
}

func (Report) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.CascadeThroughAnnotationField([]entx.ThroughCleanup{
			{
				Field:   "Report",
				Through: "ReportMembership",
			},
		}),
		entx.ConsoleRoute(),
		entfga.SelfAccessChecks(),
	}
}

func (Report) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogAnyModule,
	}
}

func (Report) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookValidateReportQuery(),
		hooks.HookReportAuthz(),
	}
}

func (Report) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.FilterQueryResults[generated.Report](nil),
	}
}

func (Report) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.ReportMutation](),
		),
	)
}
