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

	"github.com/theopenlane/entx"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
)

// Entity holds the schema definition for the Entity entity
type Entity struct {
	ent.Schema
}

// Fields of the Entity
func (Entity) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the entity").
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "citext",
			}).
			MinLen(3).
			Validate(validator.SpecialCharValidator).
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			),
		field.String("display_name").
			Comment("The entity's displayed 'friendly' name").
			MaxLen(nameMaxLen).
			Optional().
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("display_name"),
			),
		field.String("description").
			Comment("An optional description of the entity").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.Strings("domains").
			Comment("domains associated with the entity").
			Validate(validator.ValidateDomains()).
			Optional(),
		field.String("entity_type_id").
			Comment("The type of the entity").
			Optional(),
		field.String("status").
			Comment("status of the entity").
			Default("active").
			Optional(),
	}
}

// Mixin of the Entity
func (Entity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
		emixin.TagMixin{},
		NewOrgOwnMixinWithRef("entities"),
	}
}

// Edges of the Entity
func (Entity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("contacts", Contact.Type),
		edge.To("documents", DocumentData.Type),
		edge.To("notes", Note.Type),
		edge.To("files", File.Type),
		edge.To("entity_type", EntityType.Type).
			Field("entity_type_id").
			Unique(),
	}
}

// Indexes of the Entity
func (Entity) Indexes() []ent.Index {
	return []ent.Index{
		// names should be unique, but ignore deleted names
		index.Fields("name", "owner_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Annotations of the Entity
func (Entity) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entfga.OrganizationInheritedChecks(),
	}
}

// Hooks of the Entity
func (Entity) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookEntityCreate(),
	}
}

// Interceptors of the Entity
func (Entity) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the Entity
func (Entity) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			entfga.CheckReadAccess[*generated.EntityQuery](),
		),
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.EntityMutation](),
		),
	)
}
