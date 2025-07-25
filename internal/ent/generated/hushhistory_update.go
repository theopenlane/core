// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/hushhistory"
	"github.com/theopenlane/core/internal/ent/generated/predicate"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// HushHistoryUpdate is the builder for updating HushHistory entities.
type HushHistoryUpdate struct {
	config
	hooks     []Hook
	mutation  *HushHistoryMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the HushHistoryUpdate builder.
func (hhu *HushHistoryUpdate) Where(ps ...predicate.HushHistory) *HushHistoryUpdate {
	hhu.mutation.Where(ps...)
	return hhu
}

// SetUpdatedAt sets the "updated_at" field.
func (hhu *HushHistoryUpdate) SetUpdatedAt(t time.Time) *HushHistoryUpdate {
	hhu.mutation.SetUpdatedAt(t)
	return hhu
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (hhu *HushHistoryUpdate) ClearUpdatedAt() *HushHistoryUpdate {
	hhu.mutation.ClearUpdatedAt()
	return hhu
}

// SetUpdatedBy sets the "updated_by" field.
func (hhu *HushHistoryUpdate) SetUpdatedBy(s string) *HushHistoryUpdate {
	hhu.mutation.SetUpdatedBy(s)
	return hhu
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (hhu *HushHistoryUpdate) SetNillableUpdatedBy(s *string) *HushHistoryUpdate {
	if s != nil {
		hhu.SetUpdatedBy(*s)
	}
	return hhu
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (hhu *HushHistoryUpdate) ClearUpdatedBy() *HushHistoryUpdate {
	hhu.mutation.ClearUpdatedBy()
	return hhu
}

// SetDeletedAt sets the "deleted_at" field.
func (hhu *HushHistoryUpdate) SetDeletedAt(t time.Time) *HushHistoryUpdate {
	hhu.mutation.SetDeletedAt(t)
	return hhu
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (hhu *HushHistoryUpdate) SetNillableDeletedAt(t *time.Time) *HushHistoryUpdate {
	if t != nil {
		hhu.SetDeletedAt(*t)
	}
	return hhu
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (hhu *HushHistoryUpdate) ClearDeletedAt() *HushHistoryUpdate {
	hhu.mutation.ClearDeletedAt()
	return hhu
}

// SetDeletedBy sets the "deleted_by" field.
func (hhu *HushHistoryUpdate) SetDeletedBy(s string) *HushHistoryUpdate {
	hhu.mutation.SetDeletedBy(s)
	return hhu
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (hhu *HushHistoryUpdate) SetNillableDeletedBy(s *string) *HushHistoryUpdate {
	if s != nil {
		hhu.SetDeletedBy(*s)
	}
	return hhu
}

// ClearDeletedBy clears the value of the "deleted_by" field.
func (hhu *HushHistoryUpdate) ClearDeletedBy() *HushHistoryUpdate {
	hhu.mutation.ClearDeletedBy()
	return hhu
}

// SetOwnerID sets the "owner_id" field.
func (hhu *HushHistoryUpdate) SetOwnerID(s string) *HushHistoryUpdate {
	hhu.mutation.SetOwnerID(s)
	return hhu
}

// SetNillableOwnerID sets the "owner_id" field if the given value is not nil.
func (hhu *HushHistoryUpdate) SetNillableOwnerID(s *string) *HushHistoryUpdate {
	if s != nil {
		hhu.SetOwnerID(*s)
	}
	return hhu
}

// ClearOwnerID clears the value of the "owner_id" field.
func (hhu *HushHistoryUpdate) ClearOwnerID() *HushHistoryUpdate {
	hhu.mutation.ClearOwnerID()
	return hhu
}

// SetName sets the "name" field.
func (hhu *HushHistoryUpdate) SetName(s string) *HushHistoryUpdate {
	hhu.mutation.SetName(s)
	return hhu
}

// SetNillableName sets the "name" field if the given value is not nil.
func (hhu *HushHistoryUpdate) SetNillableName(s *string) *HushHistoryUpdate {
	if s != nil {
		hhu.SetName(*s)
	}
	return hhu
}

// SetDescription sets the "description" field.
func (hhu *HushHistoryUpdate) SetDescription(s string) *HushHistoryUpdate {
	hhu.mutation.SetDescription(s)
	return hhu
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (hhu *HushHistoryUpdate) SetNillableDescription(s *string) *HushHistoryUpdate {
	if s != nil {
		hhu.SetDescription(*s)
	}
	return hhu
}

// ClearDescription clears the value of the "description" field.
func (hhu *HushHistoryUpdate) ClearDescription() *HushHistoryUpdate {
	hhu.mutation.ClearDescription()
	return hhu
}

// SetKind sets the "kind" field.
func (hhu *HushHistoryUpdate) SetKind(s string) *HushHistoryUpdate {
	hhu.mutation.SetKind(s)
	return hhu
}

// SetNillableKind sets the "kind" field if the given value is not nil.
func (hhu *HushHistoryUpdate) SetNillableKind(s *string) *HushHistoryUpdate {
	if s != nil {
		hhu.SetKind(*s)
	}
	return hhu
}

// ClearKind clears the value of the "kind" field.
func (hhu *HushHistoryUpdate) ClearKind() *HushHistoryUpdate {
	hhu.mutation.ClearKind()
	return hhu
}

// Mutation returns the HushHistoryMutation object of the builder.
func (hhu *HushHistoryUpdate) Mutation() *HushHistoryMutation {
	return hhu.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (hhu *HushHistoryUpdate) Save(ctx context.Context) (int, error) {
	if err := hhu.defaults(); err != nil {
		return 0, err
	}
	return withHooks(ctx, hhu.sqlSave, hhu.mutation, hhu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (hhu *HushHistoryUpdate) SaveX(ctx context.Context) int {
	affected, err := hhu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (hhu *HushHistoryUpdate) Exec(ctx context.Context) error {
	_, err := hhu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (hhu *HushHistoryUpdate) ExecX(ctx context.Context) {
	if err := hhu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (hhu *HushHistoryUpdate) defaults() error {
	if _, ok := hhu.mutation.UpdatedAt(); !ok && !hhu.mutation.UpdatedAtCleared() {
		if hushhistory.UpdateDefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized hushhistory.UpdateDefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := hushhistory.UpdateDefaultUpdatedAt()
		hhu.mutation.SetUpdatedAt(v)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (hhu *HushHistoryUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *HushHistoryUpdate {
	hhu.modifiers = append(hhu.modifiers, modifiers...)
	return hhu
}

func (hhu *HushHistoryUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(hushhistory.Table, hushhistory.Columns, sqlgraph.NewFieldSpec(hushhistory.FieldID, field.TypeString))
	if ps := hhu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if hhu.mutation.RefCleared() {
		_spec.ClearField(hushhistory.FieldRef, field.TypeString)
	}
	if hhu.mutation.CreatedAtCleared() {
		_spec.ClearField(hushhistory.FieldCreatedAt, field.TypeTime)
	}
	if value, ok := hhu.mutation.UpdatedAt(); ok {
		_spec.SetField(hushhistory.FieldUpdatedAt, field.TypeTime, value)
	}
	if hhu.mutation.UpdatedAtCleared() {
		_spec.ClearField(hushhistory.FieldUpdatedAt, field.TypeTime)
	}
	if hhu.mutation.CreatedByCleared() {
		_spec.ClearField(hushhistory.FieldCreatedBy, field.TypeString)
	}
	if value, ok := hhu.mutation.UpdatedBy(); ok {
		_spec.SetField(hushhistory.FieldUpdatedBy, field.TypeString, value)
	}
	if hhu.mutation.UpdatedByCleared() {
		_spec.ClearField(hushhistory.FieldUpdatedBy, field.TypeString)
	}
	if value, ok := hhu.mutation.DeletedAt(); ok {
		_spec.SetField(hushhistory.FieldDeletedAt, field.TypeTime, value)
	}
	if hhu.mutation.DeletedAtCleared() {
		_spec.ClearField(hushhistory.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := hhu.mutation.DeletedBy(); ok {
		_spec.SetField(hushhistory.FieldDeletedBy, field.TypeString, value)
	}
	if hhu.mutation.DeletedByCleared() {
		_spec.ClearField(hushhistory.FieldDeletedBy, field.TypeString)
	}
	if value, ok := hhu.mutation.OwnerID(); ok {
		_spec.SetField(hushhistory.FieldOwnerID, field.TypeString, value)
	}
	if hhu.mutation.OwnerIDCleared() {
		_spec.ClearField(hushhistory.FieldOwnerID, field.TypeString)
	}
	if value, ok := hhu.mutation.Name(); ok {
		_spec.SetField(hushhistory.FieldName, field.TypeString, value)
	}
	if value, ok := hhu.mutation.Description(); ok {
		_spec.SetField(hushhistory.FieldDescription, field.TypeString, value)
	}
	if hhu.mutation.DescriptionCleared() {
		_spec.ClearField(hushhistory.FieldDescription, field.TypeString)
	}
	if value, ok := hhu.mutation.Kind(); ok {
		_spec.SetField(hushhistory.FieldKind, field.TypeString, value)
	}
	if hhu.mutation.KindCleared() {
		_spec.ClearField(hushhistory.FieldKind, field.TypeString)
	}
	if hhu.mutation.SecretNameCleared() {
		_spec.ClearField(hushhistory.FieldSecretName, field.TypeString)
	}
	if hhu.mutation.SecretValueCleared() {
		_spec.ClearField(hushhistory.FieldSecretValue, field.TypeString)
	}
	_spec.Node.Schema = hhu.schemaConfig.HushHistory
	ctx = internal.NewSchemaConfigContext(ctx, hhu.schemaConfig)
	_spec.AddModifiers(hhu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, hhu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{hushhistory.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	hhu.mutation.done = true
	return n, nil
}

// HushHistoryUpdateOne is the builder for updating a single HushHistory entity.
type HushHistoryUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *HushHistoryMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetUpdatedAt sets the "updated_at" field.
func (hhuo *HushHistoryUpdateOne) SetUpdatedAt(t time.Time) *HushHistoryUpdateOne {
	hhuo.mutation.SetUpdatedAt(t)
	return hhuo
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (hhuo *HushHistoryUpdateOne) ClearUpdatedAt() *HushHistoryUpdateOne {
	hhuo.mutation.ClearUpdatedAt()
	return hhuo
}

// SetUpdatedBy sets the "updated_by" field.
func (hhuo *HushHistoryUpdateOne) SetUpdatedBy(s string) *HushHistoryUpdateOne {
	hhuo.mutation.SetUpdatedBy(s)
	return hhuo
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (hhuo *HushHistoryUpdateOne) SetNillableUpdatedBy(s *string) *HushHistoryUpdateOne {
	if s != nil {
		hhuo.SetUpdatedBy(*s)
	}
	return hhuo
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (hhuo *HushHistoryUpdateOne) ClearUpdatedBy() *HushHistoryUpdateOne {
	hhuo.mutation.ClearUpdatedBy()
	return hhuo
}

// SetDeletedAt sets the "deleted_at" field.
func (hhuo *HushHistoryUpdateOne) SetDeletedAt(t time.Time) *HushHistoryUpdateOne {
	hhuo.mutation.SetDeletedAt(t)
	return hhuo
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (hhuo *HushHistoryUpdateOne) SetNillableDeletedAt(t *time.Time) *HushHistoryUpdateOne {
	if t != nil {
		hhuo.SetDeletedAt(*t)
	}
	return hhuo
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (hhuo *HushHistoryUpdateOne) ClearDeletedAt() *HushHistoryUpdateOne {
	hhuo.mutation.ClearDeletedAt()
	return hhuo
}

// SetDeletedBy sets the "deleted_by" field.
func (hhuo *HushHistoryUpdateOne) SetDeletedBy(s string) *HushHistoryUpdateOne {
	hhuo.mutation.SetDeletedBy(s)
	return hhuo
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (hhuo *HushHistoryUpdateOne) SetNillableDeletedBy(s *string) *HushHistoryUpdateOne {
	if s != nil {
		hhuo.SetDeletedBy(*s)
	}
	return hhuo
}

// ClearDeletedBy clears the value of the "deleted_by" field.
func (hhuo *HushHistoryUpdateOne) ClearDeletedBy() *HushHistoryUpdateOne {
	hhuo.mutation.ClearDeletedBy()
	return hhuo
}

// SetOwnerID sets the "owner_id" field.
func (hhuo *HushHistoryUpdateOne) SetOwnerID(s string) *HushHistoryUpdateOne {
	hhuo.mutation.SetOwnerID(s)
	return hhuo
}

// SetNillableOwnerID sets the "owner_id" field if the given value is not nil.
func (hhuo *HushHistoryUpdateOne) SetNillableOwnerID(s *string) *HushHistoryUpdateOne {
	if s != nil {
		hhuo.SetOwnerID(*s)
	}
	return hhuo
}

// ClearOwnerID clears the value of the "owner_id" field.
func (hhuo *HushHistoryUpdateOne) ClearOwnerID() *HushHistoryUpdateOne {
	hhuo.mutation.ClearOwnerID()
	return hhuo
}

// SetName sets the "name" field.
func (hhuo *HushHistoryUpdateOne) SetName(s string) *HushHistoryUpdateOne {
	hhuo.mutation.SetName(s)
	return hhuo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (hhuo *HushHistoryUpdateOne) SetNillableName(s *string) *HushHistoryUpdateOne {
	if s != nil {
		hhuo.SetName(*s)
	}
	return hhuo
}

// SetDescription sets the "description" field.
func (hhuo *HushHistoryUpdateOne) SetDescription(s string) *HushHistoryUpdateOne {
	hhuo.mutation.SetDescription(s)
	return hhuo
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (hhuo *HushHistoryUpdateOne) SetNillableDescription(s *string) *HushHistoryUpdateOne {
	if s != nil {
		hhuo.SetDescription(*s)
	}
	return hhuo
}

// ClearDescription clears the value of the "description" field.
func (hhuo *HushHistoryUpdateOne) ClearDescription() *HushHistoryUpdateOne {
	hhuo.mutation.ClearDescription()
	return hhuo
}

// SetKind sets the "kind" field.
func (hhuo *HushHistoryUpdateOne) SetKind(s string) *HushHistoryUpdateOne {
	hhuo.mutation.SetKind(s)
	return hhuo
}

// SetNillableKind sets the "kind" field if the given value is not nil.
func (hhuo *HushHistoryUpdateOne) SetNillableKind(s *string) *HushHistoryUpdateOne {
	if s != nil {
		hhuo.SetKind(*s)
	}
	return hhuo
}

// ClearKind clears the value of the "kind" field.
func (hhuo *HushHistoryUpdateOne) ClearKind() *HushHistoryUpdateOne {
	hhuo.mutation.ClearKind()
	return hhuo
}

// Mutation returns the HushHistoryMutation object of the builder.
func (hhuo *HushHistoryUpdateOne) Mutation() *HushHistoryMutation {
	return hhuo.mutation
}

// Where appends a list predicates to the HushHistoryUpdate builder.
func (hhuo *HushHistoryUpdateOne) Where(ps ...predicate.HushHistory) *HushHistoryUpdateOne {
	hhuo.mutation.Where(ps...)
	return hhuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (hhuo *HushHistoryUpdateOne) Select(field string, fields ...string) *HushHistoryUpdateOne {
	hhuo.fields = append([]string{field}, fields...)
	return hhuo
}

// Save executes the query and returns the updated HushHistory entity.
func (hhuo *HushHistoryUpdateOne) Save(ctx context.Context) (*HushHistory, error) {
	if err := hhuo.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, hhuo.sqlSave, hhuo.mutation, hhuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (hhuo *HushHistoryUpdateOne) SaveX(ctx context.Context) *HushHistory {
	node, err := hhuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (hhuo *HushHistoryUpdateOne) Exec(ctx context.Context) error {
	_, err := hhuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (hhuo *HushHistoryUpdateOne) ExecX(ctx context.Context) {
	if err := hhuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (hhuo *HushHistoryUpdateOne) defaults() error {
	if _, ok := hhuo.mutation.UpdatedAt(); !ok && !hhuo.mutation.UpdatedAtCleared() {
		if hushhistory.UpdateDefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized hushhistory.UpdateDefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := hushhistory.UpdateDefaultUpdatedAt()
		hhuo.mutation.SetUpdatedAt(v)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (hhuo *HushHistoryUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *HushHistoryUpdateOne {
	hhuo.modifiers = append(hhuo.modifiers, modifiers...)
	return hhuo
}

func (hhuo *HushHistoryUpdateOne) sqlSave(ctx context.Context) (_node *HushHistory, err error) {
	_spec := sqlgraph.NewUpdateSpec(hushhistory.Table, hushhistory.Columns, sqlgraph.NewFieldSpec(hushhistory.FieldID, field.TypeString))
	id, ok := hhuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`generated: missing "HushHistory.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := hhuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, hushhistory.FieldID)
		for _, f := range fields {
			if !hushhistory.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("generated: invalid field %q for query", f)}
			}
			if f != hushhistory.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := hhuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if hhuo.mutation.RefCleared() {
		_spec.ClearField(hushhistory.FieldRef, field.TypeString)
	}
	if hhuo.mutation.CreatedAtCleared() {
		_spec.ClearField(hushhistory.FieldCreatedAt, field.TypeTime)
	}
	if value, ok := hhuo.mutation.UpdatedAt(); ok {
		_spec.SetField(hushhistory.FieldUpdatedAt, field.TypeTime, value)
	}
	if hhuo.mutation.UpdatedAtCleared() {
		_spec.ClearField(hushhistory.FieldUpdatedAt, field.TypeTime)
	}
	if hhuo.mutation.CreatedByCleared() {
		_spec.ClearField(hushhistory.FieldCreatedBy, field.TypeString)
	}
	if value, ok := hhuo.mutation.UpdatedBy(); ok {
		_spec.SetField(hushhistory.FieldUpdatedBy, field.TypeString, value)
	}
	if hhuo.mutation.UpdatedByCleared() {
		_spec.ClearField(hushhistory.FieldUpdatedBy, field.TypeString)
	}
	if value, ok := hhuo.mutation.DeletedAt(); ok {
		_spec.SetField(hushhistory.FieldDeletedAt, field.TypeTime, value)
	}
	if hhuo.mutation.DeletedAtCleared() {
		_spec.ClearField(hushhistory.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := hhuo.mutation.DeletedBy(); ok {
		_spec.SetField(hushhistory.FieldDeletedBy, field.TypeString, value)
	}
	if hhuo.mutation.DeletedByCleared() {
		_spec.ClearField(hushhistory.FieldDeletedBy, field.TypeString)
	}
	if value, ok := hhuo.mutation.OwnerID(); ok {
		_spec.SetField(hushhistory.FieldOwnerID, field.TypeString, value)
	}
	if hhuo.mutation.OwnerIDCleared() {
		_spec.ClearField(hushhistory.FieldOwnerID, field.TypeString)
	}
	if value, ok := hhuo.mutation.Name(); ok {
		_spec.SetField(hushhistory.FieldName, field.TypeString, value)
	}
	if value, ok := hhuo.mutation.Description(); ok {
		_spec.SetField(hushhistory.FieldDescription, field.TypeString, value)
	}
	if hhuo.mutation.DescriptionCleared() {
		_spec.ClearField(hushhistory.FieldDescription, field.TypeString)
	}
	if value, ok := hhuo.mutation.Kind(); ok {
		_spec.SetField(hushhistory.FieldKind, field.TypeString, value)
	}
	if hhuo.mutation.KindCleared() {
		_spec.ClearField(hushhistory.FieldKind, field.TypeString)
	}
	if hhuo.mutation.SecretNameCleared() {
		_spec.ClearField(hushhistory.FieldSecretName, field.TypeString)
	}
	if hhuo.mutation.SecretValueCleared() {
		_spec.ClearField(hushhistory.FieldSecretValue, field.TypeString)
	}
	_spec.Node.Schema = hhuo.schemaConfig.HushHistory
	ctx = internal.NewSchemaConfigContext(ctx, hhuo.schemaConfig)
	_spec.AddModifiers(hhuo.modifiers...)
	_node = &HushHistory{config: hhuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, hhuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{hushhistory.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	hhuo.mutation.done = true
	return _node, nil
}
