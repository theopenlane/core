// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/predicate"

	"github.com/theopenlane/core/internal/ent/generated/entitlementhistory"
	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// EntitlementHistoryDelete is the builder for deleting a EntitlementHistory entity.
type EntitlementHistoryDelete struct {
	config
	hooks    []Hook
	mutation *EntitlementHistoryMutation
}

// Where appends a list predicates to the EntitlementHistoryDelete builder.
func (ehd *EntitlementHistoryDelete) Where(ps ...predicate.EntitlementHistory) *EntitlementHistoryDelete {
	ehd.mutation.Where(ps...)
	return ehd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (ehd *EntitlementHistoryDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, ehd.sqlExec, ehd.mutation, ehd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (ehd *EntitlementHistoryDelete) ExecX(ctx context.Context) int {
	n, err := ehd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (ehd *EntitlementHistoryDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(entitlementhistory.Table, sqlgraph.NewFieldSpec(entitlementhistory.FieldID, field.TypeString))
	_spec.Node.Schema = ehd.schemaConfig.EntitlementHistory
	ctx = internal.NewSchemaConfigContext(ctx, ehd.schemaConfig)
	if ps := ehd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, ehd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	ehd.mutation.done = true
	return affected, err
}

// EntitlementHistoryDeleteOne is the builder for deleting a single EntitlementHistory entity.
type EntitlementHistoryDeleteOne struct {
	ehd *EntitlementHistoryDelete
}

// Where appends a list predicates to the EntitlementHistoryDelete builder.
func (ehdo *EntitlementHistoryDeleteOne) Where(ps ...predicate.EntitlementHistory) *EntitlementHistoryDeleteOne {
	ehdo.ehd.mutation.Where(ps...)
	return ehdo
}

// Exec executes the deletion query.
func (ehdo *EntitlementHistoryDeleteOne) Exec(ctx context.Context) error {
	n, err := ehdo.ehd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{entitlementhistory.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (ehdo *EntitlementHistoryDeleteOne) ExecX(ctx context.Context) {
	if err := ehdo.Exec(ctx); err != nil {
		panic(err)
	}
}