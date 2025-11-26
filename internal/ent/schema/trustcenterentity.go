package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/pkg/models"
)

// TrustcenterEntity holds the schema definition for the TrustcenterEntity entity
type TrustcenterEntity struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTrustcenterEntityis the name of the schema in snake case
const SchemaTrustcenterEntity = "trustcenter_entity"

// Name is the name of the schema in snake case
func (TrustcenterEntity) Name() string {
	return SchemaTrustcenterEntity
}

// GetType returns the type of the schema
func (TrustcenterEntity) GetType() any {
	return TrustcenterEntity.Type
}

// PluralName returns the plural name of the schema
func (TrustcenterEntity) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTrustcenterEntity)
}

// Fields of the TrustcenterEntity
func (TrustcenterEntity) Fields() []ent.Field {
	return []ent.Field{
		field.String("logo_file_id").
			Comment("The local logo file id").
			Optional().
			Nillable(),
		field.String("url").
			Comment("URL of customer's website").
			MaxLen(urlMaxLen).
			Validate(validator.ValidateURL()).
			Optional().
			Annotations(
				entx.FieldSearchable(),
			),
		field.String("trust_center_id").
			Comment("The trust center this entity belongs to").
			Optional(),
		field.String("name").
			Comment("The name of the tag definition").
			Immutable().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("NAME"),
			),
	}
}

// Mixin of the TrustcenterEntity
func (t TrustcenterEntity) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags:      true,
		additionalMixins: []ent.Mixin{},
	}.getMixins(t)
}

// Edges of the TrustcenterEntity
func (t TrustcenterEntity) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: t,
			name:       "logo_file",
			t:          File.Type,
			field:      "logo_file_id",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: t,
			edgeSchema: TrustCenter{},
			field:      "trust_center_id",
		}),
		defaultEdgeToWithPagination(t, File{}),
	}
}

// Indexes of the TrustcenterEntity
func (TrustcenterEntity) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the TrustcenterEntity
func (TrustcenterEntity) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}

// Hooks of the TrustcenterEntity
func (TrustcenterEntity) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the TrustcenterEntity
func (TrustcenterEntity) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Modules this schema has access to
func (TrustcenterEntity) Modules() []models.OrgModule {
	return []models.OrgModule{}
}

// Policy of the TrustcenterEntity
func (TrustcenterEntity) Policy() ent.Policy {
	// add the new policy here, the default post-policy is to deny all
	// so you need to ensure there are rules in place to allow the actions you want
	return policy.NewPolicy(
		policy.WithMutationRules(
			// add mutation rules here, the below is the recommended default
			policy.CheckCreateAccess(),
			// this needs to be commented out for the first run that had the entfga annotation
			// the first run will generate the functions required based on the entfa annotation
			// entfga.CheckEditAccess[*generated.TrustcenterEntityMutation](),
		),
	)
}
