// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/predicate"

	"github.com/theopenlane/core/internal/ent/generated/internal"
	"github.com/theopenlane/core/internal/ent/generated/orgprice"
)

// OrgPriceDelete is the builder for deleting a OrgPrice entity.
type OrgPriceDelete struct {
	config
	hooks    []Hook
	mutation *OrgPriceMutation
}

// Where appends a list predicates to the OrgPriceDelete builder.
func (opd *OrgPriceDelete) Where(ps ...predicate.OrgPrice) *OrgPriceDelete {
	opd.mutation.Where(ps...)
	return opd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (opd *OrgPriceDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, opd.sqlExec, opd.mutation, opd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (opd *OrgPriceDelete) ExecX(ctx context.Context) int {
	n, err := opd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (opd *OrgPriceDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(orgprice.Table, sqlgraph.NewFieldSpec(orgprice.FieldID, field.TypeString))
	_spec.Node.Schema = opd.schemaConfig.OrgPrice
	ctx = internal.NewSchemaConfigContext(ctx, opd.schemaConfig)
	if ps := opd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, opd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	opd.mutation.done = true
	return affected, err
}

// OrgPriceDeleteOne is the builder for deleting a single OrgPrice entity.
type OrgPriceDeleteOne struct {
	opd *OrgPriceDelete
}

// Where appends a list predicates to the OrgPriceDelete builder.
func (opdo *OrgPriceDeleteOne) Where(ps ...predicate.OrgPrice) *OrgPriceDeleteOne {
	opdo.opd.mutation.Where(ps...)
	return opdo
}

// Exec executes the deletion query.
func (opdo *OrgPriceDeleteOne) Exec(ctx context.Context) error {
	n, err := opdo.opd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{orgprice.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (opdo *OrgPriceDeleteOne) ExecX(ctx context.Context) {
	if err := opdo.Exec(ctx); err != nil {
		panic(err)
	}
}
