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

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// SLADefinition defines the SLA definition schema.
type SLADefinition struct {
	SchemaFuncs

	ent.Schema
}

// SchemaSLADefinition is the name of the SLA definition schema.
const SchemaSLADefinition = "sla_definition"

// Name returns the name of the SLA definition schema.
func (SLADefinition) Name() string {
	return SchemaSLADefinition
}

// GetType returns the type of the SLA definition schema.
func (SLADefinition) GetType() any {
	return SLADefinition.Type
}

// PluralName returns the plural name of the SLA definition schema.
func (SLADefinition) PluralName() string {
	return pluralize.NewClient().Plural(SchemaSLADefinition)
}

// Fields returns SLA definition fields.
func (SLADefinition) Fields() []ent.Field {
	return []ent.Field{
		field.Int("sla_days").
			Comment("remediation service level agreement in days for the severity level").
			Positive().
			Annotations(
				entgql.OrderField("sla_days"),
				entx.FieldSearchable(),
			),

		field.Enum("security_level").
			Comment("incoming source severity").
			GoType(enums.SecurityLevel("")).
			Default(enums.SecurityLevelNone.String()).
			Annotations(
				entgql.OrderField("security_level"),
				entgql.Skip(entgql.SkipMutationCreateInput|entgql.SkipMutationUpdateInput),
			),
	}
}

// Mixin of the SLADefinition
func (s SLADefinition) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "SLAD",
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(s),
			newCustomEnumMixin(s, withEnumFieldName("severity_level")),
			newGroupPermissionsMixin(),
		},
	}.getMixins(s)
}

// Edges of the SLADefinition
func (SLADefinition) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the SLADefinition
func (SLADefinition) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("security_level", ownerFieldName).
			Unique().
			Annotations(
				entsql.IndexWhere("deleted_at is NULL"),
			),
	}
}

// Annotations of the SLADefinition
func (SLADefinition) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}

// Policy of the SLADefinition
func (SLADefinition) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}

func (SLADefinition) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogVulnerabilityManagementModule,
	}
}
