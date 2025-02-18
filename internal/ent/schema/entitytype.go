package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
)

const (
	maxEntityNameLen = 64
)

// EntityType holds the schema definition for the EntityType entity
type EntityType struct {
	ent.Schema
}

// Fields of the EntityType
func (EntityType) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the entity").
			SchemaType(map[string]string{
				dialect.Postgres: "citext",
			}).
			Validate(validator.SpecialCharValidator).
			MaxLen(maxEntityNameLen).
			NotEmpty().
			Annotations(
				entgql.OrderField("name"),
			),
	}
}

// Mixin of the EntityType
func (EntityType) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
		emixin.TagMixin{},
		NewOrgOwnMixinWithRef("entity_types"),
	}
}

// Edges of the EntityType
func (EntityType) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entities", Entity.Type),
	}
}

// Indexes of the EntityType
func (EntityType) Indexes() []ent.Index {
	return []ent.Index{
		// names should be unique by owner, but ignore deleted names
		index.Fields("name", "owner_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Annotations of the EntityType
func (EntityType) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.OrganizationInheritedChecks(),
	}
}

// Hooks of the EntityType
func (EntityType) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the EntityType
func (EntityType) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the EntityType
func (EntityType) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.EntityTypeMutation](),
		),
	)
}
