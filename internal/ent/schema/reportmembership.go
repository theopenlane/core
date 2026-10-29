package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

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

type ReportMembership struct {
	SchemaFuncs

	ent.Schema
}

const SchemaReportMembership = "report_membership"

func (ReportMembership) Name() string {
	return SchemaReportMembership
}

func (ReportMembership) GetType() any {
	return ReportMembership.Type
}

func (ReportMembership) PluralName() string {
	return pluralize.NewClient().Plural(SchemaReportMembership)
}

func (ReportMembership) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("role").
			GoType(enums.Role("")).
			Values(enums.RoleOwner.String()).
			Default(enums.RoleMember.String()).
			Annotations(
				entgql.OrderField("ROLE"),
			),
		field.String("report_id").
			Immutable(),
		field.String("user_id").
			Immutable(),
	}
}

func (r ReportMembership) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: r,
			edgeSchema: Report{},
			required:   true,
			immutable:  true,
			field:      "report_id",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: r,
			edgeSchema: User{},
			required:   true,
			immutable:  true,
			field:      "user_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(User{}.Name()),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: r,
			edgeSchema: OrgMembership{},
			immutable:  true,
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
	}
}

func (ReportMembership) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.MembershipChecks("report"),
		entx.FGACrudParent(Report{}.Name()),
	}
}

func (ReportMembership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "report_id").
			Unique(),
	}
}

func (r ReportMembership) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags:       true,
		excludeSoftDelete: true,
	}.getMixins(r)
}

func (ReportMembership) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookMembershipSelf("report_memberships"),
		hooks.HookReportMembers(),
	}
}

func (ReportMembership) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.ReportMembershipMutation](),
		),
	)
}

func (ReportMembership) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.FilterListQuery(),
	}
}

func (ReportMembership) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogAnyModule,
	}
}
