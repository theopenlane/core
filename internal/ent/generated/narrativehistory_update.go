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
	"github.com/theopenlane/core/internal/ent/generated/narrativehistory"
	"github.com/theopenlane/core/internal/ent/generated/predicate"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// NarrativeHistoryUpdate is the builder for updating NarrativeHistory entities.
type NarrativeHistoryUpdate struct {
	config
	hooks     []Hook
	mutation  *NarrativeHistoryMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the NarrativeHistoryUpdate builder.
func (nhu *NarrativeHistoryUpdate) Where(ps ...predicate.NarrativeHistory) *NarrativeHistoryUpdate {
	nhu.mutation.Where(ps...)
	return nhu
}

// SetUpdatedAt sets the "updated_at" field.
func (nhu *NarrativeHistoryUpdate) SetUpdatedAt(t time.Time) *NarrativeHistoryUpdate {
	nhu.mutation.SetUpdatedAt(t)
	return nhu
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (nhu *NarrativeHistoryUpdate) ClearUpdatedAt() *NarrativeHistoryUpdate {
	nhu.mutation.ClearUpdatedAt()
	return nhu
}

// SetUpdatedBy sets the "updated_by" field.
func (nhu *NarrativeHistoryUpdate) SetUpdatedBy(s string) *NarrativeHistoryUpdate {
	nhu.mutation.SetUpdatedBy(s)
	return nhu
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (nhu *NarrativeHistoryUpdate) SetNillableUpdatedBy(s *string) *NarrativeHistoryUpdate {
	if s != nil {
		nhu.SetUpdatedBy(*s)
	}
	return nhu
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (nhu *NarrativeHistoryUpdate) ClearUpdatedBy() *NarrativeHistoryUpdate {
	nhu.mutation.ClearUpdatedBy()
	return nhu
}

// SetDeletedAt sets the "deleted_at" field.
func (nhu *NarrativeHistoryUpdate) SetDeletedAt(t time.Time) *NarrativeHistoryUpdate {
	nhu.mutation.SetDeletedAt(t)
	return nhu
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (nhu *NarrativeHistoryUpdate) SetNillableDeletedAt(t *time.Time) *NarrativeHistoryUpdate {
	if t != nil {
		nhu.SetDeletedAt(*t)
	}
	return nhu
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (nhu *NarrativeHistoryUpdate) ClearDeletedAt() *NarrativeHistoryUpdate {
	nhu.mutation.ClearDeletedAt()
	return nhu
}

// SetDeletedBy sets the "deleted_by" field.
func (nhu *NarrativeHistoryUpdate) SetDeletedBy(s string) *NarrativeHistoryUpdate {
	nhu.mutation.SetDeletedBy(s)
	return nhu
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (nhu *NarrativeHistoryUpdate) SetNillableDeletedBy(s *string) *NarrativeHistoryUpdate {
	if s != nil {
		nhu.SetDeletedBy(*s)
	}
	return nhu
}

// ClearDeletedBy clears the value of the "deleted_by" field.
func (nhu *NarrativeHistoryUpdate) ClearDeletedBy() *NarrativeHistoryUpdate {
	nhu.mutation.ClearDeletedBy()
	return nhu
}

// SetTags sets the "tags" field.
func (nhu *NarrativeHistoryUpdate) SetTags(s []string) *NarrativeHistoryUpdate {
	nhu.mutation.SetTags(s)
	return nhu
}

// AppendTags appends s to the "tags" field.
func (nhu *NarrativeHistoryUpdate) AppendTags(s []string) *NarrativeHistoryUpdate {
	nhu.mutation.AppendTags(s)
	return nhu
}

// ClearTags clears the value of the "tags" field.
func (nhu *NarrativeHistoryUpdate) ClearTags() *NarrativeHistoryUpdate {
	nhu.mutation.ClearTags()
	return nhu
}

// SetName sets the "name" field.
func (nhu *NarrativeHistoryUpdate) SetName(s string) *NarrativeHistoryUpdate {
	nhu.mutation.SetName(s)
	return nhu
}

// SetNillableName sets the "name" field if the given value is not nil.
func (nhu *NarrativeHistoryUpdate) SetNillableName(s *string) *NarrativeHistoryUpdate {
	if s != nil {
		nhu.SetName(*s)
	}
	return nhu
}

// SetDescription sets the "description" field.
func (nhu *NarrativeHistoryUpdate) SetDescription(s string) *NarrativeHistoryUpdate {
	nhu.mutation.SetDescription(s)
	return nhu
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (nhu *NarrativeHistoryUpdate) SetNillableDescription(s *string) *NarrativeHistoryUpdate {
	if s != nil {
		nhu.SetDescription(*s)
	}
	return nhu
}

// ClearDescription clears the value of the "description" field.
func (nhu *NarrativeHistoryUpdate) ClearDescription() *NarrativeHistoryUpdate {
	nhu.mutation.ClearDescription()
	return nhu
}

// SetDetails sets the "details" field.
func (nhu *NarrativeHistoryUpdate) SetDetails(s string) *NarrativeHistoryUpdate {
	nhu.mutation.SetDetails(s)
	return nhu
}

// SetNillableDetails sets the "details" field if the given value is not nil.
func (nhu *NarrativeHistoryUpdate) SetNillableDetails(s *string) *NarrativeHistoryUpdate {
	if s != nil {
		nhu.SetDetails(*s)
	}
	return nhu
}

// ClearDetails clears the value of the "details" field.
func (nhu *NarrativeHistoryUpdate) ClearDetails() *NarrativeHistoryUpdate {
	nhu.mutation.ClearDetails()
	return nhu
}

// Mutation returns the NarrativeHistoryMutation object of the builder.
func (nhu *NarrativeHistoryUpdate) Mutation() *NarrativeHistoryMutation {
	return nhu.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (nhu *NarrativeHistoryUpdate) Save(ctx context.Context) (int, error) {
	if err := nhu.defaults(); err != nil {
		return 0, err
	}
	return withHooks(ctx, nhu.sqlSave, nhu.mutation, nhu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (nhu *NarrativeHistoryUpdate) SaveX(ctx context.Context) int {
	affected, err := nhu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (nhu *NarrativeHistoryUpdate) Exec(ctx context.Context) error {
	_, err := nhu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (nhu *NarrativeHistoryUpdate) ExecX(ctx context.Context) {
	if err := nhu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (nhu *NarrativeHistoryUpdate) defaults() error {
	if _, ok := nhu.mutation.UpdatedAt(); !ok && !nhu.mutation.UpdatedAtCleared() {
		if narrativehistory.UpdateDefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized narrativehistory.UpdateDefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := narrativehistory.UpdateDefaultUpdatedAt()
		nhu.mutation.SetUpdatedAt(v)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (nhu *NarrativeHistoryUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *NarrativeHistoryUpdate {
	nhu.modifiers = append(nhu.modifiers, modifiers...)
	return nhu
}

func (nhu *NarrativeHistoryUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(narrativehistory.Table, narrativehistory.Columns, sqlgraph.NewFieldSpec(narrativehistory.FieldID, field.TypeString))
	if ps := nhu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if nhu.mutation.RefCleared() {
		_spec.ClearField(narrativehistory.FieldRef, field.TypeString)
	}
	if nhu.mutation.CreatedAtCleared() {
		_spec.ClearField(narrativehistory.FieldCreatedAt, field.TypeTime)
	}
	if value, ok := nhu.mutation.UpdatedAt(); ok {
		_spec.SetField(narrativehistory.FieldUpdatedAt, field.TypeTime, value)
	}
	if nhu.mutation.UpdatedAtCleared() {
		_spec.ClearField(narrativehistory.FieldUpdatedAt, field.TypeTime)
	}
	if nhu.mutation.CreatedByCleared() {
		_spec.ClearField(narrativehistory.FieldCreatedBy, field.TypeString)
	}
	if value, ok := nhu.mutation.UpdatedBy(); ok {
		_spec.SetField(narrativehistory.FieldUpdatedBy, field.TypeString, value)
	}
	if nhu.mutation.UpdatedByCleared() {
		_spec.ClearField(narrativehistory.FieldUpdatedBy, field.TypeString)
	}
	if value, ok := nhu.mutation.DeletedAt(); ok {
		_spec.SetField(narrativehistory.FieldDeletedAt, field.TypeTime, value)
	}
	if nhu.mutation.DeletedAtCleared() {
		_spec.ClearField(narrativehistory.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := nhu.mutation.DeletedBy(); ok {
		_spec.SetField(narrativehistory.FieldDeletedBy, field.TypeString, value)
	}
	if nhu.mutation.DeletedByCleared() {
		_spec.ClearField(narrativehistory.FieldDeletedBy, field.TypeString)
	}
	if value, ok := nhu.mutation.Tags(); ok {
		_spec.SetField(narrativehistory.FieldTags, field.TypeJSON, value)
	}
	if value, ok := nhu.mutation.AppendedTags(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, narrativehistory.FieldTags, value)
		})
	}
	if nhu.mutation.TagsCleared() {
		_spec.ClearField(narrativehistory.FieldTags, field.TypeJSON)
	}
	if nhu.mutation.OwnerIDCleared() {
		_spec.ClearField(narrativehistory.FieldOwnerID, field.TypeString)
	}
	if value, ok := nhu.mutation.Name(); ok {
		_spec.SetField(narrativehistory.FieldName, field.TypeString, value)
	}
	if value, ok := nhu.mutation.Description(); ok {
		_spec.SetField(narrativehistory.FieldDescription, field.TypeString, value)
	}
	if nhu.mutation.DescriptionCleared() {
		_spec.ClearField(narrativehistory.FieldDescription, field.TypeString)
	}
	if value, ok := nhu.mutation.Details(); ok {
		_spec.SetField(narrativehistory.FieldDetails, field.TypeString, value)
	}
	if nhu.mutation.DetailsCleared() {
		_spec.ClearField(narrativehistory.FieldDetails, field.TypeString)
	}
	_spec.Node.Schema = nhu.schemaConfig.NarrativeHistory
	ctx = internal.NewSchemaConfigContext(ctx, nhu.schemaConfig)
	_spec.AddModifiers(nhu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, nhu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{narrativehistory.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	nhu.mutation.done = true
	return n, nil
}

// NarrativeHistoryUpdateOne is the builder for updating a single NarrativeHistory entity.
type NarrativeHistoryUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *NarrativeHistoryMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetUpdatedAt sets the "updated_at" field.
func (nhuo *NarrativeHistoryUpdateOne) SetUpdatedAt(t time.Time) *NarrativeHistoryUpdateOne {
	nhuo.mutation.SetUpdatedAt(t)
	return nhuo
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (nhuo *NarrativeHistoryUpdateOne) ClearUpdatedAt() *NarrativeHistoryUpdateOne {
	nhuo.mutation.ClearUpdatedAt()
	return nhuo
}

// SetUpdatedBy sets the "updated_by" field.
func (nhuo *NarrativeHistoryUpdateOne) SetUpdatedBy(s string) *NarrativeHistoryUpdateOne {
	nhuo.mutation.SetUpdatedBy(s)
	return nhuo
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (nhuo *NarrativeHistoryUpdateOne) SetNillableUpdatedBy(s *string) *NarrativeHistoryUpdateOne {
	if s != nil {
		nhuo.SetUpdatedBy(*s)
	}
	return nhuo
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (nhuo *NarrativeHistoryUpdateOne) ClearUpdatedBy() *NarrativeHistoryUpdateOne {
	nhuo.mutation.ClearUpdatedBy()
	return nhuo
}

// SetDeletedAt sets the "deleted_at" field.
func (nhuo *NarrativeHistoryUpdateOne) SetDeletedAt(t time.Time) *NarrativeHistoryUpdateOne {
	nhuo.mutation.SetDeletedAt(t)
	return nhuo
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (nhuo *NarrativeHistoryUpdateOne) SetNillableDeletedAt(t *time.Time) *NarrativeHistoryUpdateOne {
	if t != nil {
		nhuo.SetDeletedAt(*t)
	}
	return nhuo
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (nhuo *NarrativeHistoryUpdateOne) ClearDeletedAt() *NarrativeHistoryUpdateOne {
	nhuo.mutation.ClearDeletedAt()
	return nhuo
}

// SetDeletedBy sets the "deleted_by" field.
func (nhuo *NarrativeHistoryUpdateOne) SetDeletedBy(s string) *NarrativeHistoryUpdateOne {
	nhuo.mutation.SetDeletedBy(s)
	return nhuo
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (nhuo *NarrativeHistoryUpdateOne) SetNillableDeletedBy(s *string) *NarrativeHistoryUpdateOne {
	if s != nil {
		nhuo.SetDeletedBy(*s)
	}
	return nhuo
}

// ClearDeletedBy clears the value of the "deleted_by" field.
func (nhuo *NarrativeHistoryUpdateOne) ClearDeletedBy() *NarrativeHistoryUpdateOne {
	nhuo.mutation.ClearDeletedBy()
	return nhuo
}

// SetTags sets the "tags" field.
func (nhuo *NarrativeHistoryUpdateOne) SetTags(s []string) *NarrativeHistoryUpdateOne {
	nhuo.mutation.SetTags(s)
	return nhuo
}

// AppendTags appends s to the "tags" field.
func (nhuo *NarrativeHistoryUpdateOne) AppendTags(s []string) *NarrativeHistoryUpdateOne {
	nhuo.mutation.AppendTags(s)
	return nhuo
}

// ClearTags clears the value of the "tags" field.
func (nhuo *NarrativeHistoryUpdateOne) ClearTags() *NarrativeHistoryUpdateOne {
	nhuo.mutation.ClearTags()
	return nhuo
}

// SetName sets the "name" field.
func (nhuo *NarrativeHistoryUpdateOne) SetName(s string) *NarrativeHistoryUpdateOne {
	nhuo.mutation.SetName(s)
	return nhuo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (nhuo *NarrativeHistoryUpdateOne) SetNillableName(s *string) *NarrativeHistoryUpdateOne {
	if s != nil {
		nhuo.SetName(*s)
	}
	return nhuo
}

// SetDescription sets the "description" field.
func (nhuo *NarrativeHistoryUpdateOne) SetDescription(s string) *NarrativeHistoryUpdateOne {
	nhuo.mutation.SetDescription(s)
	return nhuo
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (nhuo *NarrativeHistoryUpdateOne) SetNillableDescription(s *string) *NarrativeHistoryUpdateOne {
	if s != nil {
		nhuo.SetDescription(*s)
	}
	return nhuo
}

// ClearDescription clears the value of the "description" field.
func (nhuo *NarrativeHistoryUpdateOne) ClearDescription() *NarrativeHistoryUpdateOne {
	nhuo.mutation.ClearDescription()
	return nhuo
}

// SetDetails sets the "details" field.
func (nhuo *NarrativeHistoryUpdateOne) SetDetails(s string) *NarrativeHistoryUpdateOne {
	nhuo.mutation.SetDetails(s)
	return nhuo
}

// SetNillableDetails sets the "details" field if the given value is not nil.
func (nhuo *NarrativeHistoryUpdateOne) SetNillableDetails(s *string) *NarrativeHistoryUpdateOne {
	if s != nil {
		nhuo.SetDetails(*s)
	}
	return nhuo
}

// ClearDetails clears the value of the "details" field.
func (nhuo *NarrativeHistoryUpdateOne) ClearDetails() *NarrativeHistoryUpdateOne {
	nhuo.mutation.ClearDetails()
	return nhuo
}

// Mutation returns the NarrativeHistoryMutation object of the builder.
func (nhuo *NarrativeHistoryUpdateOne) Mutation() *NarrativeHistoryMutation {
	return nhuo.mutation
}

// Where appends a list predicates to the NarrativeHistoryUpdate builder.
func (nhuo *NarrativeHistoryUpdateOne) Where(ps ...predicate.NarrativeHistory) *NarrativeHistoryUpdateOne {
	nhuo.mutation.Where(ps...)
	return nhuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (nhuo *NarrativeHistoryUpdateOne) Select(field string, fields ...string) *NarrativeHistoryUpdateOne {
	nhuo.fields = append([]string{field}, fields...)
	return nhuo
}

// Save executes the query and returns the updated NarrativeHistory entity.
func (nhuo *NarrativeHistoryUpdateOne) Save(ctx context.Context) (*NarrativeHistory, error) {
	if err := nhuo.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, nhuo.sqlSave, nhuo.mutation, nhuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (nhuo *NarrativeHistoryUpdateOne) SaveX(ctx context.Context) *NarrativeHistory {
	node, err := nhuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (nhuo *NarrativeHistoryUpdateOne) Exec(ctx context.Context) error {
	_, err := nhuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (nhuo *NarrativeHistoryUpdateOne) ExecX(ctx context.Context) {
	if err := nhuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (nhuo *NarrativeHistoryUpdateOne) defaults() error {
	if _, ok := nhuo.mutation.UpdatedAt(); !ok && !nhuo.mutation.UpdatedAtCleared() {
		if narrativehistory.UpdateDefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized narrativehistory.UpdateDefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := narrativehistory.UpdateDefaultUpdatedAt()
		nhuo.mutation.SetUpdatedAt(v)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (nhuo *NarrativeHistoryUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *NarrativeHistoryUpdateOne {
	nhuo.modifiers = append(nhuo.modifiers, modifiers...)
	return nhuo
}

func (nhuo *NarrativeHistoryUpdateOne) sqlSave(ctx context.Context) (_node *NarrativeHistory, err error) {
	_spec := sqlgraph.NewUpdateSpec(narrativehistory.Table, narrativehistory.Columns, sqlgraph.NewFieldSpec(narrativehistory.FieldID, field.TypeString))
	id, ok := nhuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`generated: missing "NarrativeHistory.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := nhuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, narrativehistory.FieldID)
		for _, f := range fields {
			if !narrativehistory.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("generated: invalid field %q for query", f)}
			}
			if f != narrativehistory.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := nhuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if nhuo.mutation.RefCleared() {
		_spec.ClearField(narrativehistory.FieldRef, field.TypeString)
	}
	if nhuo.mutation.CreatedAtCleared() {
		_spec.ClearField(narrativehistory.FieldCreatedAt, field.TypeTime)
	}
	if value, ok := nhuo.mutation.UpdatedAt(); ok {
		_spec.SetField(narrativehistory.FieldUpdatedAt, field.TypeTime, value)
	}
	if nhuo.mutation.UpdatedAtCleared() {
		_spec.ClearField(narrativehistory.FieldUpdatedAt, field.TypeTime)
	}
	if nhuo.mutation.CreatedByCleared() {
		_spec.ClearField(narrativehistory.FieldCreatedBy, field.TypeString)
	}
	if value, ok := nhuo.mutation.UpdatedBy(); ok {
		_spec.SetField(narrativehistory.FieldUpdatedBy, field.TypeString, value)
	}
	if nhuo.mutation.UpdatedByCleared() {
		_spec.ClearField(narrativehistory.FieldUpdatedBy, field.TypeString)
	}
	if value, ok := nhuo.mutation.DeletedAt(); ok {
		_spec.SetField(narrativehistory.FieldDeletedAt, field.TypeTime, value)
	}
	if nhuo.mutation.DeletedAtCleared() {
		_spec.ClearField(narrativehistory.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := nhuo.mutation.DeletedBy(); ok {
		_spec.SetField(narrativehistory.FieldDeletedBy, field.TypeString, value)
	}
	if nhuo.mutation.DeletedByCleared() {
		_spec.ClearField(narrativehistory.FieldDeletedBy, field.TypeString)
	}
	if value, ok := nhuo.mutation.Tags(); ok {
		_spec.SetField(narrativehistory.FieldTags, field.TypeJSON, value)
	}
	if value, ok := nhuo.mutation.AppendedTags(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, narrativehistory.FieldTags, value)
		})
	}
	if nhuo.mutation.TagsCleared() {
		_spec.ClearField(narrativehistory.FieldTags, field.TypeJSON)
	}
	if nhuo.mutation.OwnerIDCleared() {
		_spec.ClearField(narrativehistory.FieldOwnerID, field.TypeString)
	}
	if value, ok := nhuo.mutation.Name(); ok {
		_spec.SetField(narrativehistory.FieldName, field.TypeString, value)
	}
	if value, ok := nhuo.mutation.Description(); ok {
		_spec.SetField(narrativehistory.FieldDescription, field.TypeString, value)
	}
	if nhuo.mutation.DescriptionCleared() {
		_spec.ClearField(narrativehistory.FieldDescription, field.TypeString)
	}
	if value, ok := nhuo.mutation.Details(); ok {
		_spec.SetField(narrativehistory.FieldDetails, field.TypeString, value)
	}
	if nhuo.mutation.DetailsCleared() {
		_spec.ClearField(narrativehistory.FieldDetails, field.TypeString)
	}
	_spec.Node.Schema = nhuo.schemaConfig.NarrativeHistory
	ctx = internal.NewSchemaConfigContext(ctx, nhuo.schemaConfig)
	_spec.AddModifiers(nhuo.modifiers...)
	_node = &NarrativeHistory{config: nhuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, nhuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{narrativehistory.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	nhuo.mutation.done = true
	return _node, nil
}
