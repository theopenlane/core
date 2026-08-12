package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

type Audience struct {
	SchemaFuncs

	ent.Schema
}

const SchemaAudience = "audience"

func (Audience) Name() string {
	return SchemaAudience
}

func (Audience) GetType() any {
	return Audience.Type
}

func (Audience) PluralName() string {
	return pluralize.NewClient().Plural(SchemaAudience)
}

func (Audience) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the audience").
			SchemaType(map[string]string{
				dialect.Postgres: "citext",
			}).
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			),
		field.String("description").
			Comment("the description of the audience").
			Optional(),
		field.Enum("audience_type").
			Comment("the type of audience").
			GoType(enums.AudienceType("")).
			Default(enums.AudienceTypeManual.String()).
			Annotations(
				entgql.OrderField("TYPE"),
			),
		field.JSON("filters", map[string]any{}).
			Comment("filter definition used to resolve dynamic audience members").
			Optional(),
	}
}

func (a Audience) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "AUD",
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(a),
		},
	}.getMixins(a)
}

func (a Audience) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(a, AudienceMember{}),
		defaultEdgeFromWithPagination(a, Campaign{}),
	}
}

func (Audience) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", ownerFieldName).
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

func (Audience) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
		models.CatalogTrustCenterModule,
	}
}

func (Audience) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.NewExportable(),
	}
}

func (Audience) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
			policy.CheckCreateAccess(),
		),
	)
}
