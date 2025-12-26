package mixin

import (
	"fmt"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"golang.org/x/mod/semver"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/common/models"
	"github.com/theopenlane/core/internal/ent/hooks"
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
			Annotations(
				entgql.OrderField("revision"),
			).
			Comment("revision of the object as a semver (e.g. v1.0.0), by default any update will bump the patch version, unless the revision_bump field is set"),
	}
}

// Hooks of the RevisionMixin.
func (d RevisionMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookRevisionUpdate(),
	}
}
