package schema

import (
	"context"
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/entx"
)

// NewAuditMixin creates a new AuditMixin
func NewAuditMixin() AuditMixin {
	return AuditMixin{}
}

// NewAuditMixinWithExcludedEdges creates a new AuditMixin with the edges excluded
func NewAuditMixinWithExcludedEdges() AuditMixin {
	return AuditMixin{
		ExcludeEdge: true,
	}
}

// AuditMixin provides auditing for all records where enabled. The created_at, created_by_i, updated_at, and updated_by_id records are automatically populated when this mixin is enabled.
type AuditMixin struct {
	mixin.Schema
	// ExcludeEdge is a boolean to indicate if the edges should be excluded
	ExcludeEdge bool
}

// Fields of the AuditMixin
func (AuditMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Immutable().
			Optional().
			Default(time.Now).
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput,
				),
				entx.FieldAdminSearchable(false),
			),
		field.Time("updated_at").
			Default(time.Now).
			Optional().
			UpdateDefault(time.Now).
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput,
				),
				entx.FieldAdminSearchable(false),
			),
		field.String("created_by_id").
			Immutable().
			Optional().
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput,
				),
				entx.FieldAdminSearchable(false),
			),
		field.String("updated_by_id").
			Optional().
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput,
				),
				entx.FieldAdminSearchable(false),
			),
	}
}

// Edges of the AuditMixin
func (a AuditMixin) Edges() []ent.Edge {
	if a.ExcludeEdge {
		return nil
	}

	return []ent.Edge{
		edge.To("created_by", ChangeActor.Type).
			Field("created_by_id").
			Unique().
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput,
				),
			).
			Immutable(),
		edge.To("updated_by", ChangeActor.Type).
			Field("updated_by_id").
			Unique().
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput,
				),
			),
	}
}

// Hooks of the AuditMixin
func (AuditMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		AuditHook,
	}
}

// AuditHook sets and returns the created_at, updated_at, etc., fields
func AuditHook(next ent.Mutator) ent.Mutator {
	type AuditLogger interface {
		SetCreatedAt(time.Time)
		CreatedAt() (v time.Time, exists bool) // exists if present before this hook
		SetUpdatedAt(time.Time)
		UpdatedAt() (v time.Time, exists bool)
		SetCreatedByID(string)
		CreatedByID() (id string, exists bool)
		SetUpdatedByID(string)
		UpdatedByID() (id string, exists bool)
	}

	return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
		ml, ok := m.(AuditLogger)
		if !ok {
			return nil, newUnexpectedAuditError(m)
		}

		actor, _ := auth.GetUserIDFromContext(ctx) // ignore error, leave as null if not found

		switch op := m.Op(); {
		case op.Is(ent.OpCreate):
			ml.SetCreatedAt(time.Now())

			if actor != "" {
				ml.SetCreatedByID(actor)
				ml.SetUpdatedByID(actor)
			}

		case op.Is(ent.OpUpdateOne | ent.OpUpdate):
			ml.SetUpdatedAt(time.Now())

			if actor != "" {
				ml.SetUpdatedByID(actor)
			}
		}

		return next.Mutate(ctx, m)
	})
}
