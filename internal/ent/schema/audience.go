package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/ent/privacy/policy"
)

// Audience holds the schema definition for reusable recipient audiences.
type Audience struct {
	SchemaFuncs

	ent.Schema
}

// SchemaAudience is the name of the Audience schema.
const SchemaAudience = "audience"

// Name returns the name of the Audience schema.
func (Audience) Name() string {
	return SchemaAudience
}

// GetType returns the type of the Audience schema.
func (Audience) GetType() any {
	return Audience.Type
}

// PluralName returns the plural name of the Audience schema.
func (Audience) PluralName() string {
	return pluralize.NewClient().Plural(SchemaAudience)
}

// Fields of the Audience.
func (Audience) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the audience").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			),
		field.String("description").
			Comment("the description of the audience").
			Optional(),
		field.Enum("audience_type").
			Comment("the audience resolution type").
			GoType(enums.AudienceType("")).
			Default(enums.AudienceTypeManual.String()).
			Annotations(
				entgql.OrderField("AUDIENCE_TYPE"),
			),
		field.JSON("filters", map[string]any{}).
			Comment("selector filters for dynamic audiences").
			Optional(),
		field.JSON("metadata", map[string]any{}).
			Comment("additional metadata about the audience").
			Optional(),
	}
}

// Mixin of the Audience.
func (a Audience) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "AUD",
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(a),
			newGroupPermissionsMixin(),
		},
	}.getMixins(a)
}

// Edges of the Audience.
func (a Audience) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(a, AudienceMember{}),
		defaultEdgeFromWithPagination(a, Campaign{}),
	}
}

// Indexes of the Audience.
func (Audience) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", ownerFieldName).
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Modules this schema has access to.
func (Audience) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
		models.CatalogTrustCenterModule,
	}
}

// Annotations of the Audience.
func (Audience) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		entx.NewExportable(),
	}
}

// Hooks of the Audience.
func (Audience) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookAudienceValidateFilters(),
	}
}

// Policy of the Audience.
func (Audience) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.AudienceMutation](),
		),
	)
}
