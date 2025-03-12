package mixin

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/pkg/models"
)

// RevisionMixin implements the revision pattern for schemas.
type RevisionMixin struct {
	mixin.Schema
}

// Fields of the RevisionMixin.
func (RevisionMixin) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("revision", &models.SemverVersion{}).
			Default(&models.SemverVersion{
				Patch: 1, // default to v0.0.1
			}).
			Annotations(entgql.Type("String")).
			Optional().
			Comment("revision of the object, by default any update will bump the patch version, unless the revision_bump field is set"),
		field.Enum("revision_bump").
			Optional().
			GoType(models.VersionBump("")).
			Comment("revision bump type - major, minor, patch, or draft").
			Annotations(
				entgql.Skip(
					^entgql.SkipMutationUpdateInput, // only allow updates to update the revision field
				),
				entsql.Skip(),
			),
	}
}

// Hooks of the RevisionMixin.
func (d RevisionMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookRevisionUpdate(),
	}
}
