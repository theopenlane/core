package mixin

import (
	"context"
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
)

// SoftDeleteMixin implements the soft delete pattern for schemas.
type SoftDeleteMixin struct {
	mixin.Schema
}

// Fields of the SoftDeleteMixin.
func (SoftDeleteMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("deleted_at").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipAll),
			),
		field.String("deleted_by").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipAll),
			),
	}
}

// Interceptors of the SoftDeleteMixin.
func (d SoftDeleteMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			// Skip soft-delete, means include soft-deleted entities.
			if skip, _ := ctx.Value(entx.SoftDeleteSkipKey{}).(bool); skip {
				return nil
			}

			d.P(q)

			return nil
		}),
	}
}

// SoftDeleteHook will soft delete records, by changing the delete mutation to an update and setting
// the deleted_at and deleted_by fields, unless the softDeleteSkipKey is set
func (d SoftDeleteMixin) SoftDeleteHook(next ent.Mutator) ent.Mutator {
	type SoftDelete interface {
		SetOp(ent.Op)
		Client() *generated.Client
		SetDeletedAt(time.Time)
		SetDeletedBy(string)
		WhereP(...func(*sql.Selector))
	}

	return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
		if skip, _ := ctx.Value(entx.SoftDeleteSkipKey{}).(bool); skip {
			return next.Mutate(ctx, m)
		}

		actor := "unknown"
		if caller, ok := auth.CallerFromContext(ctx); ok && caller != nil && caller.SubjectID != "" {
			actor = caller.SubjectID
		}

		sd, ok := m.(SoftDelete)
		if !ok {
			return nil, newUnexpectedMutationTypeError(m)
		}

		d.P(sd)
		sd.SetOp(ent.OpUpdate)

		// set that the transaction is a soft-delete
		ctx = entx.IsSoftDelete(ctx, m.Type())

		sd.SetDeletedAt(time.Now())
		sd.SetDeletedBy(actor)

		return sd.Client().Mutate(ctx, m)
	})
}

// Hooks of the SoftDeleteMixin.
func (d SoftDeleteMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(
			d.SoftDeleteHook,
			ent.OpDeleteOne|ent.OpDelete,
		),
	}
}

// P adds a storage-level predicate to the queries and mutations.
func (d SoftDeleteMixin) P(w interface{ WhereP(...func(*sql.Selector)) }) {
	w.WhereP(
		sql.FieldIsNull(d.Fields()[0].Descriptor().Name),
	)
}
