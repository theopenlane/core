package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/models"

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
		field.String("title").
			Comment("the title of the note").
			Optional().
			Nillable(),
		field.Text("text").
			Comment("the text of the note").
			NotEmpty(),
		field.JSON("text_json", []any{}).
			Optional().
			Annotations(
				entgql.Type("[Any!]"),
			).
			Comment("structured details of the note in JSON format"),
		field.String("note_ref").
			Comment("ref location of the note").
			Optional(),
		field.String("discussion_id").
			Comment("the external discussion id this note is associated with").
			Optional(),
		field.Bool("is_edited").
			Comment("whether the note has been edited").
			Default(false),
		field.String("trust_center_id").
			Comment("the trust center this note belongs to, if applicable").
			Optional(),
	}
}

// Mixin of the Note
func (n Note) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix:      "NTE",
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.Note](
				n,
				withParents(InternalPolicy{}, Procedure{}, Control{}, Subcontrol{}, ControlObjective{}, Program{}, Task{}, TrustCenter{}, Risk{}, Evidence{}, Discussion{}),
				withOrganizationOwner(false),
				withOwnerRelation(fgax.OwnerRelation),
				withAllowAnonymousTrustCenterAccess(true),
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
			field:      "trust_center_id",
			ref:        "posts",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: n,
			edgeSchema: Discussion{},
			ref:        "comments",
			field:      "discussion_id",
			annotations: []schema.Annotation{
				// you should only need to be able to view a discussion to add a comment to it
				accessmap.EdgeViewCheck(Discussion{}.Name()),
			},
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
		hooks.HookSlateJSON(),
	}
}
