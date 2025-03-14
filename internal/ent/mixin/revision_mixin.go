package mixin

import (
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"golang.org/x/mod/semver"

	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/utils/rout"
)

// RevisionMixin implements the revision pattern for schemas.
type RevisionMixin struct {
	mixin.Schema
}

// Fields of the RevisionMixin.
func (RevisionMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("revision").
			Default(models.DefaultRevision).
			Validate(func(s string) error {
				ok := semver.IsValid(s)
				if !ok {
					return fmt.Errorf("%w, invalid semver value", rout.InvalidField("revision"))
				}

				return nil
			}).
			Optional().
			Comment("revision of the object as a semver (e.g. v1.0.0), by default any update will bump the patch version, unless the revision_bump field is set"),
		// field.Enum("revision_bump").
		// 	Optional().
		// 	GoType(models.VersionBump("")).
		// 	Comment("revision bump type - major, minor, patch, or draft").
		// 	StorageKey("revision").
		// 	Annotations(
		// 		entgql.Skip(
		// 			^entgql.SkipEnumField & ^entgql.SkipMutationUpdateInput,
		// 		),
		// 		entsql.Skip(), // this is not a database field
		// 	),
	}
}

// Hooks of the RevisionMixin.
func (d RevisionMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookRevisionUpdate(),
	}
}
