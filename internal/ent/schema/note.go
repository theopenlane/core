package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/pkg/models"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// Note holds the schema definition for the Note entity
type Note struct {
	SchemaFuncs

	ent.Schema
}

const SchemaNote = "note"

func (Note) Name() string {
	return SchemaNote
}

func (Note) GetType() any {
	return Note.Type
}

func (Note) PluralName() string {
	return pluralize.NewClient().Plural(SchemaNote)
}

// Fields of the Note
func (Note) Fields() []ent.Field {
	return []ent.Field{
		field.Text("text").
			Comment("the text of the note").
			NotEmpty(),
		field.String("note_ref").
			Comment("ref location of the note").
			Optional(),
		field.String("discussion_id").
			Comment("the external discussion id this note is associated with").
			Optional(),
		field.Bool("is_edited").
			Comment("whether the note has been edited").
			Default(false),
	}
}

// id
// comment_ref (string)
// discussion_id (number)
// user_id (string)
// content (json)
// is_edited (boolean) default false
// created_at (DateTime)
// updated_at (DateTime)

// Mixin of the Note
func (n Note) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix:      "NTE",
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.Note](
				n,
				withParents(InternalPolicy{}, Procedure{}, Control{}, Subcontrol{}, ControlObjective{}, Program{}, Task{}, TrustCenter{}, Risk{}, Evidence{}),
				withOrganizationOwner(false),
				withOwnerRelation(fgax.OwnerRelation),
			),
		},
	}.getMixins(n)
}

// Edges of the Note
func (n Note) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: n,
			edgeSchema: Task{},
			ref:        "comments",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: n,
			edgeSchema: Control{},
			ref:        "comments",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: n,
			edgeSchema: Subcontrol{},
			ref:        "comments",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: n,
			edgeSchema: Procedure{},
			ref:        "comments",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: n,
			edgeSchema: Risk{},
			ref:        "comments",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: n,
			edgeSchema: InternalPolicy{},
			ref:        "comments",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: n,
			edgeSchema: Evidence{},
			ref:        "comments",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: n,
			edgeSchema: TrustCenter{},
			ref:        "posts",
		}),
		defaultEdgeToWithPagination(n, File{}),
	}
}

func (Note) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Annotations of the Note
func (n Note) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		// skip generating the schema for this type, this schema is used through extended types
		entx.SchemaGenSkip(true),
		entx.QueryGenSkip(true),
	}
}

// Policy of the Note
func (n Note) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.AllowCreate(),
			entfga.CheckEditAccess[*generated.NoteMutation](),
		),
	)
}

// Hooks of the Note
func (Note) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookNoteFiles(),
		hooks.HookTrustcenterCacheInvalidation(),
	}
}
