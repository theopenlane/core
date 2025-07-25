// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/dialect/sql/sqljson"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/mappabledomainhistory"
	"github.com/theopenlane/core/internal/ent/generated/predicate"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// MappableDomainHistoryUpdate is the builder for updating MappableDomainHistory entities.
type MappableDomainHistoryUpdate struct {
	config
	hooks     []Hook
	mutation  *MappableDomainHistoryMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the MappableDomainHistoryUpdate builder.
func (mdhu *MappableDomainHistoryUpdate) Where(ps ...predicate.MappableDomainHistory) *MappableDomainHistoryUpdate {
	mdhu.mutation.Where(ps...)
	return mdhu
}

// SetUpdatedAt sets the "updated_at" field.
func (mdhu *MappableDomainHistoryUpdate) SetUpdatedAt(t time.Time) *MappableDomainHistoryUpdate {
	mdhu.mutation.SetUpdatedAt(t)
	return mdhu
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (mdhu *MappableDomainHistoryUpdate) ClearUpdatedAt() *MappableDomainHistoryUpdate {
	mdhu.mutation.ClearUpdatedAt()
	return mdhu
}

// SetUpdatedBy sets the "updated_by" field.
func (mdhu *MappableDomainHistoryUpdate) SetUpdatedBy(s string) *MappableDomainHistoryUpdate {
	mdhu.mutation.SetUpdatedBy(s)
	return mdhu
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (mdhu *MappableDomainHistoryUpdate) SetNillableUpdatedBy(s *string) *MappableDomainHistoryUpdate {
	if s != nil {
		mdhu.SetUpdatedBy(*s)
	}
	return mdhu
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (mdhu *MappableDomainHistoryUpdate) ClearUpdatedBy() *MappableDomainHistoryUpdate {
	mdhu.mutation.ClearUpdatedBy()
	return mdhu
}

// SetDeletedAt sets the "deleted_at" field.
func (mdhu *MappableDomainHistoryUpdate) SetDeletedAt(t time.Time) *MappableDomainHistoryUpdate {
	mdhu.mutation.SetDeletedAt(t)
	return mdhu
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (mdhu *MappableDomainHistoryUpdate) SetNillableDeletedAt(t *time.Time) *MappableDomainHistoryUpdate {
	if t != nil {
		mdhu.SetDeletedAt(*t)
	}
	return mdhu
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (mdhu *MappableDomainHistoryUpdate) ClearDeletedAt() *MappableDomainHistoryUpdate {
	mdhu.mutation.ClearDeletedAt()
	return mdhu
}

// SetDeletedBy sets the "deleted_by" field.
func (mdhu *MappableDomainHistoryUpdate) SetDeletedBy(s string) *MappableDomainHistoryUpdate {
	mdhu.mutation.SetDeletedBy(s)
	return mdhu
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (mdhu *MappableDomainHistoryUpdate) SetNillableDeletedBy(s *string) *MappableDomainHistoryUpdate {
	if s != nil {
		mdhu.SetDeletedBy(*s)
	}
	return mdhu
}

// ClearDeletedBy clears the value of the "deleted_by" field.
func (mdhu *MappableDomainHistoryUpdate) ClearDeletedBy() *MappableDomainHistoryUpdate {
	mdhu.mutation.ClearDeletedBy()
	return mdhu
}

// SetTags sets the "tags" field.
func (mdhu *MappableDomainHistoryUpdate) SetTags(s []string) *MappableDomainHistoryUpdate {
	mdhu.mutation.SetTags(s)
	return mdhu
}

// AppendTags appends s to the "tags" field.
func (mdhu *MappableDomainHistoryUpdate) AppendTags(s []string) *MappableDomainHistoryUpdate {
	mdhu.mutation.AppendTags(s)
	return mdhu
}

// ClearTags clears the value of the "tags" field.
func (mdhu *MappableDomainHistoryUpdate) ClearTags() *MappableDomainHistoryUpdate {
	mdhu.mutation.ClearTags()
	return mdhu
}

// Mutation returns the MappableDomainHistoryMutation object of the builder.
func (mdhu *MappableDomainHistoryUpdate) Mutation() *MappableDomainHistoryMutation {
	return mdhu.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (mdhu *MappableDomainHistoryUpdate) Save(ctx context.Context) (int, error) {
	if err := mdhu.defaults(); err != nil {
		return 0, err
	}
	return withHooks(ctx, mdhu.sqlSave, mdhu.mutation, mdhu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (mdhu *MappableDomainHistoryUpdate) SaveX(ctx context.Context) int {
	affected, err := mdhu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (mdhu *MappableDomainHistoryUpdate) Exec(ctx context.Context) error {
	_, err := mdhu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (mdhu *MappableDomainHistoryUpdate) ExecX(ctx context.Context) {
	if err := mdhu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (mdhu *MappableDomainHistoryUpdate) defaults() error {
	if _, ok := mdhu.mutation.UpdatedAt(); !ok && !mdhu.mutation.UpdatedAtCleared() {
		if mappabledomainhistory.UpdateDefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized mappabledomainhistory.UpdateDefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := mappabledomainhistory.UpdateDefaultUpdatedAt()
		mdhu.mutation.SetUpdatedAt(v)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (mdhu *MappableDomainHistoryUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *MappableDomainHistoryUpdate {
	mdhu.modifiers = append(mdhu.modifiers, modifiers...)
	return mdhu
}

func (mdhu *MappableDomainHistoryUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(mappabledomainhistory.Table, mappabledomainhistory.Columns, sqlgraph.NewFieldSpec(mappabledomainhistory.FieldID, field.TypeString))
	if ps := mdhu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if mdhu.mutation.RefCleared() {
		_spec.ClearField(mappabledomainhistory.FieldRef, field.TypeString)
	}
	if mdhu.mutation.CreatedAtCleared() {
		_spec.ClearField(mappabledomainhistory.FieldCreatedAt, field.TypeTime)
	}
	if value, ok := mdhu.mutation.UpdatedAt(); ok {
		_spec.SetField(mappabledomainhistory.FieldUpdatedAt, field.TypeTime, value)
	}
	if mdhu.mutation.UpdatedAtCleared() {
		_spec.ClearField(mappabledomainhistory.FieldUpdatedAt, field.TypeTime)
	}
	if mdhu.mutation.CreatedByCleared() {
		_spec.ClearField(mappabledomainhistory.FieldCreatedBy, field.TypeString)
	}
	if value, ok := mdhu.mutation.UpdatedBy(); ok {
		_spec.SetField(mappabledomainhistory.FieldUpdatedBy, field.TypeString, value)
	}
	if mdhu.mutation.UpdatedByCleared() {
		_spec.ClearField(mappabledomainhistory.FieldUpdatedBy, field.TypeString)
	}
	if value, ok := mdhu.mutation.DeletedAt(); ok {
		_spec.SetField(mappabledomainhistory.FieldDeletedAt, field.TypeTime, value)
	}
	if mdhu.mutation.DeletedAtCleared() {
		_spec.ClearField(mappabledomainhistory.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := mdhu.mutation.DeletedBy(); ok {
		_spec.SetField(mappabledomainhistory.FieldDeletedBy, field.TypeString, value)
	}
	if mdhu.mutation.DeletedByCleared() {
		_spec.ClearField(mappabledomainhistory.FieldDeletedBy, field.TypeString)
	}
	if value, ok := mdhu.mutation.Tags(); ok {
		_spec.SetField(mappabledomainhistory.FieldTags, field.TypeJSON, value)
	}
	if value, ok := mdhu.mutation.AppendedTags(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, mappabledomainhistory.FieldTags, value)
		})
	}
	if mdhu.mutation.TagsCleared() {
		_spec.ClearField(mappabledomainhistory.FieldTags, field.TypeJSON)
	}
	_spec.Node.Schema = mdhu.schemaConfig.MappableDomainHistory
	ctx = internal.NewSchemaConfigContext(ctx, mdhu.schemaConfig)
	_spec.AddModifiers(mdhu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, mdhu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{mappabledomainhistory.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	mdhu.mutation.done = true
	return n, nil
}

// MappableDomainHistoryUpdateOne is the builder for updating a single MappableDomainHistory entity.
type MappableDomainHistoryUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *MappableDomainHistoryMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetUpdatedAt sets the "updated_at" field.
func (mdhuo *MappableDomainHistoryUpdateOne) SetUpdatedAt(t time.Time) *MappableDomainHistoryUpdateOne {
	mdhuo.mutation.SetUpdatedAt(t)
	return mdhuo
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (mdhuo *MappableDomainHistoryUpdateOne) ClearUpdatedAt() *MappableDomainHistoryUpdateOne {
	mdhuo.mutation.ClearUpdatedAt()
	return mdhuo
}

// SetUpdatedBy sets the "updated_by" field.
func (mdhuo *MappableDomainHistoryUpdateOne) SetUpdatedBy(s string) *MappableDomainHistoryUpdateOne {
	mdhuo.mutation.SetUpdatedBy(s)
	return mdhuo
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (mdhuo *MappableDomainHistoryUpdateOne) SetNillableUpdatedBy(s *string) *MappableDomainHistoryUpdateOne {
	if s != nil {
		mdhuo.SetUpdatedBy(*s)
	}
	return mdhuo
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (mdhuo *MappableDomainHistoryUpdateOne) ClearUpdatedBy() *MappableDomainHistoryUpdateOne {
	mdhuo.mutation.ClearUpdatedBy()
	return mdhuo
}

// SetDeletedAt sets the "deleted_at" field.
func (mdhuo *MappableDomainHistoryUpdateOne) SetDeletedAt(t time.Time) *MappableDomainHistoryUpdateOne {
	mdhuo.mutation.SetDeletedAt(t)
	return mdhuo
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (mdhuo *MappableDomainHistoryUpdateOne) SetNillableDeletedAt(t *time.Time) *MappableDomainHistoryUpdateOne {
	if t != nil {
		mdhuo.SetDeletedAt(*t)
	}
	return mdhuo
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (mdhuo *MappableDomainHistoryUpdateOne) ClearDeletedAt() *MappableDomainHistoryUpdateOne {
	mdhuo.mutation.ClearDeletedAt()
	return mdhuo
}

// SetDeletedBy sets the "deleted_by" field.
func (mdhuo *MappableDomainHistoryUpdateOne) SetDeletedBy(s string) *MappableDomainHistoryUpdateOne {
	mdhuo.mutation.SetDeletedBy(s)
	return mdhuo
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (mdhuo *MappableDomainHistoryUpdateOne) SetNillableDeletedBy(s *string) *MappableDomainHistoryUpdateOne {
	if s != nil {
		mdhuo.SetDeletedBy(*s)
	}
	return mdhuo
}

// ClearDeletedBy clears the value of the "deleted_by" field.
func (mdhuo *MappableDomainHistoryUpdateOne) ClearDeletedBy() *MappableDomainHistoryUpdateOne {
	mdhuo.mutation.ClearDeletedBy()
	return mdhuo
}

// SetTags sets the "tags" field.
func (mdhuo *MappableDomainHistoryUpdateOne) SetTags(s []string) *MappableDomainHistoryUpdateOne {
	mdhuo.mutation.SetTags(s)
	return mdhuo
}

// AppendTags appends s to the "tags" field.
func (mdhuo *MappableDomainHistoryUpdateOne) AppendTags(s []string) *MappableDomainHistoryUpdateOne {
	mdhuo.mutation.AppendTags(s)
	return mdhuo
}

// ClearTags clears the value of the "tags" field.
func (mdhuo *MappableDomainHistoryUpdateOne) ClearTags() *MappableDomainHistoryUpdateOne {
	mdhuo.mutation.ClearTags()
	return mdhuo
}

// Mutation returns the MappableDomainHistoryMutation object of the builder.
func (mdhuo *MappableDomainHistoryUpdateOne) Mutation() *MappableDomainHistoryMutation {
	return mdhuo.mutation
}

// Where appends a list predicates to the MappableDomainHistoryUpdate builder.
func (mdhuo *MappableDomainHistoryUpdateOne) Where(ps ...predicate.MappableDomainHistory) *MappableDomainHistoryUpdateOne {
	mdhuo.mutation.Where(ps...)
	return mdhuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (mdhuo *MappableDomainHistoryUpdateOne) Select(field string, fields ...string) *MappableDomainHistoryUpdateOne {
	mdhuo.fields = append([]string{field}, fields...)
	return mdhuo
}

// Save executes the query and returns the updated MappableDomainHistory entity.
func (mdhuo *MappableDomainHistoryUpdateOne) Save(ctx context.Context) (*MappableDomainHistory, error) {
	if err := mdhuo.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, mdhuo.sqlSave, mdhuo.mutation, mdhuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (mdhuo *MappableDomainHistoryUpdateOne) SaveX(ctx context.Context) *MappableDomainHistory {
	node, err := mdhuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (mdhuo *MappableDomainHistoryUpdateOne) Exec(ctx context.Context) error {
	_, err := mdhuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (mdhuo *MappableDomainHistoryUpdateOne) ExecX(ctx context.Context) {
	if err := mdhuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (mdhuo *MappableDomainHistoryUpdateOne) defaults() error {
	if _, ok := mdhuo.mutation.UpdatedAt(); !ok && !mdhuo.mutation.UpdatedAtCleared() {
		if mappabledomainhistory.UpdateDefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized mappabledomainhistory.UpdateDefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := mappabledomainhistory.UpdateDefaultUpdatedAt()
		mdhuo.mutation.SetUpdatedAt(v)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (mdhuo *MappableDomainHistoryUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *MappableDomainHistoryUpdateOne {
	mdhuo.modifiers = append(mdhuo.modifiers, modifiers...)
	return mdhuo
}

func (mdhuo *MappableDomainHistoryUpdateOne) sqlSave(ctx context.Context) (_node *MappableDomainHistory, err error) {
	_spec := sqlgraph.NewUpdateSpec(mappabledomainhistory.Table, mappabledomainhistory.Columns, sqlgraph.NewFieldSpec(mappabledomainhistory.FieldID, field.TypeString))
	id, ok := mdhuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`generated: missing "MappableDomainHistory.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := mdhuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, mappabledomainhistory.FieldID)
		for _, f := range fields {
			if !mappabledomainhistory.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("generated: invalid field %q for query", f)}
			}
			if f != mappabledomainhistory.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := mdhuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if mdhuo.mutation.RefCleared() {
		_spec.ClearField(mappabledomainhistory.FieldRef, field.TypeString)
	}
	if mdhuo.mutation.CreatedAtCleared() {
		_spec.ClearField(mappabledomainhistory.FieldCreatedAt, field.TypeTime)
	}
	if value, ok := mdhuo.mutation.UpdatedAt(); ok {
		_spec.SetField(mappabledomainhistory.FieldUpdatedAt, field.TypeTime, value)
	}
	if mdhuo.mutation.UpdatedAtCleared() {
		_spec.ClearField(mappabledomainhistory.FieldUpdatedAt, field.TypeTime)
	}
	if mdhuo.mutation.CreatedByCleared() {
		_spec.ClearField(mappabledomainhistory.FieldCreatedBy, field.TypeString)
	}
	if value, ok := mdhuo.mutation.UpdatedBy(); ok {
		_spec.SetField(mappabledomainhistory.FieldUpdatedBy, field.TypeString, value)
	}
	if mdhuo.mutation.UpdatedByCleared() {
		_spec.ClearField(mappabledomainhistory.FieldUpdatedBy, field.TypeString)
	}
	if value, ok := mdhuo.mutation.DeletedAt(); ok {
		_spec.SetField(mappabledomainhistory.FieldDeletedAt, field.TypeTime, value)
	}
	if mdhuo.mutation.DeletedAtCleared() {
		_spec.ClearField(mappabledomainhistory.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := mdhuo.mutation.DeletedBy(); ok {
		_spec.SetField(mappabledomainhistory.FieldDeletedBy, field.TypeString, value)
	}
	if mdhuo.mutation.DeletedByCleared() {
		_spec.ClearField(mappabledomainhistory.FieldDeletedBy, field.TypeString)
	}
	if value, ok := mdhuo.mutation.Tags(); ok {
		_spec.SetField(mappabledomainhistory.FieldTags, field.TypeJSON, value)
	}
	if value, ok := mdhuo.mutation.AppendedTags(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, mappabledomainhistory.FieldTags, value)
		})
	}
	if mdhuo.mutation.TagsCleared() {
		_spec.ClearField(mappabledomainhistory.FieldTags, field.TypeJSON)
	}
	_spec.Node.Schema = mdhuo.schemaConfig.MappableDomainHistory
	ctx = internal.NewSchemaConfigContext(ctx, mdhuo.schemaConfig)
	_spec.AddModifiers(mdhuo.modifiers...)
	_node = &MappableDomainHistory{config: mdhuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, mdhuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{mappabledomainhistory.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	mdhuo.mutation.done = true
	return _node, nil
}
