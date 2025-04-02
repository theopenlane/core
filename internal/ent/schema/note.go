package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
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
				withParents(InternalPolicy{}, Procedure{}, Control{}, Subcontrol{}, ControlObjective{}, Program{}, Task{}),
				withOrganizationOwner(false),
				withOwnerRelation(fgax.OwnerRelation),
			),
		},
	}.getMixins()
}

// Edges of the Note
func (n Note) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: n,
			edgeSchema: Task{},
			ref:        "comments",
		}),
		defaultEdgeToWithPagination(n, File{}),
	}
}

// Annotations of the Note
func (Note) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		// skip generating the schema for this type, this schema is used through extended types
		entx.SchemaGenSkip(true),
		entx.QueryGenSkip(true),
	}
}

// Policy of the Note
func (Note) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.NoteMutation](),
		),
	)
}
