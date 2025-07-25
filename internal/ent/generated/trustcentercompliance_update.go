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
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/trustcentercompliance"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// TrustCenterComplianceUpdate is the builder for updating TrustCenterCompliance entities.
type TrustCenterComplianceUpdate struct {
	config
	hooks     []Hook
	mutation  *TrustCenterComplianceMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the TrustCenterComplianceUpdate builder.
func (tccu *TrustCenterComplianceUpdate) Where(ps ...predicate.TrustCenterCompliance) *TrustCenterComplianceUpdate {
	tccu.mutation.Where(ps...)
	return tccu
}

// SetUpdatedAt sets the "updated_at" field.
func (tccu *TrustCenterComplianceUpdate) SetUpdatedAt(t time.Time) *TrustCenterComplianceUpdate {
	tccu.mutation.SetUpdatedAt(t)
	return tccu
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (tccu *TrustCenterComplianceUpdate) ClearUpdatedAt() *TrustCenterComplianceUpdate {
	tccu.mutation.ClearUpdatedAt()
	return tccu
}

// SetUpdatedBy sets the "updated_by" field.
func (tccu *TrustCenterComplianceUpdate) SetUpdatedBy(s string) *TrustCenterComplianceUpdate {
	tccu.mutation.SetUpdatedBy(s)
	return tccu
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (tccu *TrustCenterComplianceUpdate) SetNillableUpdatedBy(s *string) *TrustCenterComplianceUpdate {
	if s != nil {
		tccu.SetUpdatedBy(*s)
	}
	return tccu
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (tccu *TrustCenterComplianceUpdate) ClearUpdatedBy() *TrustCenterComplianceUpdate {
	tccu.mutation.ClearUpdatedBy()
	return tccu
}

// SetDeletedAt sets the "deleted_at" field.
func (tccu *TrustCenterComplianceUpdate) SetDeletedAt(t time.Time) *TrustCenterComplianceUpdate {
	tccu.mutation.SetDeletedAt(t)
	return tccu
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (tccu *TrustCenterComplianceUpdate) SetNillableDeletedAt(t *time.Time) *TrustCenterComplianceUpdate {
	if t != nil {
		tccu.SetDeletedAt(*t)
	}
	return tccu
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (tccu *TrustCenterComplianceUpdate) ClearDeletedAt() *TrustCenterComplianceUpdate {
	tccu.mutation.ClearDeletedAt()
	return tccu
}

// SetDeletedBy sets the "deleted_by" field.
func (tccu *TrustCenterComplianceUpdate) SetDeletedBy(s string) *TrustCenterComplianceUpdate {
	tccu.mutation.SetDeletedBy(s)
	return tccu
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (tccu *TrustCenterComplianceUpdate) SetNillableDeletedBy(s *string) *TrustCenterComplianceUpdate {
	if s != nil {
		tccu.SetDeletedBy(*s)
	}
	return tccu
}

// ClearDeletedBy clears the value of the "deleted_by" field.
func (tccu *TrustCenterComplianceUpdate) ClearDeletedBy() *TrustCenterComplianceUpdate {
	tccu.mutation.ClearDeletedBy()
	return tccu
}

// SetTags sets the "tags" field.
func (tccu *TrustCenterComplianceUpdate) SetTags(s []string) *TrustCenterComplianceUpdate {
	tccu.mutation.SetTags(s)
	return tccu
}

// AppendTags appends s to the "tags" field.
func (tccu *TrustCenterComplianceUpdate) AppendTags(s []string) *TrustCenterComplianceUpdate {
	tccu.mutation.AppendTags(s)
	return tccu
}

// ClearTags clears the value of the "tags" field.
func (tccu *TrustCenterComplianceUpdate) ClearTags() *TrustCenterComplianceUpdate {
	tccu.mutation.ClearTags()
	return tccu
}

// Mutation returns the TrustCenterComplianceMutation object of the builder.
func (tccu *TrustCenterComplianceUpdate) Mutation() *TrustCenterComplianceMutation {
	return tccu.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (tccu *TrustCenterComplianceUpdate) Save(ctx context.Context) (int, error) {
	if err := tccu.defaults(); err != nil {
		return 0, err
	}
	return withHooks(ctx, tccu.sqlSave, tccu.mutation, tccu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (tccu *TrustCenterComplianceUpdate) SaveX(ctx context.Context) int {
	affected, err := tccu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (tccu *TrustCenterComplianceUpdate) Exec(ctx context.Context) error {
	_, err := tccu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tccu *TrustCenterComplianceUpdate) ExecX(ctx context.Context) {
	if err := tccu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (tccu *TrustCenterComplianceUpdate) defaults() error {
	if _, ok := tccu.mutation.UpdatedAt(); !ok && !tccu.mutation.UpdatedAtCleared() {
		if trustcentercompliance.UpdateDefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized trustcentercompliance.UpdateDefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := trustcentercompliance.UpdateDefaultUpdatedAt()
		tccu.mutation.SetUpdatedAt(v)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (tccu *TrustCenterComplianceUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *TrustCenterComplianceUpdate {
	tccu.modifiers = append(tccu.modifiers, modifiers...)
	return tccu
}

func (tccu *TrustCenterComplianceUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(trustcentercompliance.Table, trustcentercompliance.Columns, sqlgraph.NewFieldSpec(trustcentercompliance.FieldID, field.TypeString))
	if ps := tccu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if tccu.mutation.CreatedAtCleared() {
		_spec.ClearField(trustcentercompliance.FieldCreatedAt, field.TypeTime)
	}
	if value, ok := tccu.mutation.UpdatedAt(); ok {
		_spec.SetField(trustcentercompliance.FieldUpdatedAt, field.TypeTime, value)
	}
	if tccu.mutation.UpdatedAtCleared() {
		_spec.ClearField(trustcentercompliance.FieldUpdatedAt, field.TypeTime)
	}
	if tccu.mutation.CreatedByCleared() {
		_spec.ClearField(trustcentercompliance.FieldCreatedBy, field.TypeString)
	}
	if value, ok := tccu.mutation.UpdatedBy(); ok {
		_spec.SetField(trustcentercompliance.FieldUpdatedBy, field.TypeString, value)
	}
	if tccu.mutation.UpdatedByCleared() {
		_spec.ClearField(trustcentercompliance.FieldUpdatedBy, field.TypeString)
	}
	if value, ok := tccu.mutation.DeletedAt(); ok {
		_spec.SetField(trustcentercompliance.FieldDeletedAt, field.TypeTime, value)
	}
	if tccu.mutation.DeletedAtCleared() {
		_spec.ClearField(trustcentercompliance.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := tccu.mutation.DeletedBy(); ok {
		_spec.SetField(trustcentercompliance.FieldDeletedBy, field.TypeString, value)
	}
	if tccu.mutation.DeletedByCleared() {
		_spec.ClearField(trustcentercompliance.FieldDeletedBy, field.TypeString)
	}
	if value, ok := tccu.mutation.Tags(); ok {
		_spec.SetField(trustcentercompliance.FieldTags, field.TypeJSON, value)
	}
	if value, ok := tccu.mutation.AppendedTags(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, trustcentercompliance.FieldTags, value)
		})
	}
	if tccu.mutation.TagsCleared() {
		_spec.ClearField(trustcentercompliance.FieldTags, field.TypeJSON)
	}
	_spec.Node.Schema = tccu.schemaConfig.TrustCenterCompliance
	ctx = internal.NewSchemaConfigContext(ctx, tccu.schemaConfig)
	_spec.AddModifiers(tccu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, tccu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{trustcentercompliance.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	tccu.mutation.done = true
	return n, nil
}

// TrustCenterComplianceUpdateOne is the builder for updating a single TrustCenterCompliance entity.
type TrustCenterComplianceUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *TrustCenterComplianceMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetUpdatedAt sets the "updated_at" field.
func (tccuo *TrustCenterComplianceUpdateOne) SetUpdatedAt(t time.Time) *TrustCenterComplianceUpdateOne {
	tccuo.mutation.SetUpdatedAt(t)
	return tccuo
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (tccuo *TrustCenterComplianceUpdateOne) ClearUpdatedAt() *TrustCenterComplianceUpdateOne {
	tccuo.mutation.ClearUpdatedAt()
	return tccuo
}

// SetUpdatedBy sets the "updated_by" field.
func (tccuo *TrustCenterComplianceUpdateOne) SetUpdatedBy(s string) *TrustCenterComplianceUpdateOne {
	tccuo.mutation.SetUpdatedBy(s)
	return tccuo
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (tccuo *TrustCenterComplianceUpdateOne) SetNillableUpdatedBy(s *string) *TrustCenterComplianceUpdateOne {
	if s != nil {
		tccuo.SetUpdatedBy(*s)
	}
	return tccuo
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (tccuo *TrustCenterComplianceUpdateOne) ClearUpdatedBy() *TrustCenterComplianceUpdateOne {
	tccuo.mutation.ClearUpdatedBy()
	return tccuo
}

// SetDeletedAt sets the "deleted_at" field.
func (tccuo *TrustCenterComplianceUpdateOne) SetDeletedAt(t time.Time) *TrustCenterComplianceUpdateOne {
	tccuo.mutation.SetDeletedAt(t)
	return tccuo
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (tccuo *TrustCenterComplianceUpdateOne) SetNillableDeletedAt(t *time.Time) *TrustCenterComplianceUpdateOne {
	if t != nil {
		tccuo.SetDeletedAt(*t)
	}
	return tccuo
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (tccuo *TrustCenterComplianceUpdateOne) ClearDeletedAt() *TrustCenterComplianceUpdateOne {
	tccuo.mutation.ClearDeletedAt()
	return tccuo
}

// SetDeletedBy sets the "deleted_by" field.
func (tccuo *TrustCenterComplianceUpdateOne) SetDeletedBy(s string) *TrustCenterComplianceUpdateOne {
	tccuo.mutation.SetDeletedBy(s)
	return tccuo
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (tccuo *TrustCenterComplianceUpdateOne) SetNillableDeletedBy(s *string) *TrustCenterComplianceUpdateOne {
	if s != nil {
		tccuo.SetDeletedBy(*s)
	}
	return tccuo
}

// ClearDeletedBy clears the value of the "deleted_by" field.
func (tccuo *TrustCenterComplianceUpdateOne) ClearDeletedBy() *TrustCenterComplianceUpdateOne {
	tccuo.mutation.ClearDeletedBy()
	return tccuo
}

// SetTags sets the "tags" field.
func (tccuo *TrustCenterComplianceUpdateOne) SetTags(s []string) *TrustCenterComplianceUpdateOne {
	tccuo.mutation.SetTags(s)
	return tccuo
}

// AppendTags appends s to the "tags" field.
func (tccuo *TrustCenterComplianceUpdateOne) AppendTags(s []string) *TrustCenterComplianceUpdateOne {
	tccuo.mutation.AppendTags(s)
	return tccuo
}

// ClearTags clears the value of the "tags" field.
func (tccuo *TrustCenterComplianceUpdateOne) ClearTags() *TrustCenterComplianceUpdateOne {
	tccuo.mutation.ClearTags()
	return tccuo
}

// Mutation returns the TrustCenterComplianceMutation object of the builder.
func (tccuo *TrustCenterComplianceUpdateOne) Mutation() *TrustCenterComplianceMutation {
	return tccuo.mutation
}

// Where appends a list predicates to the TrustCenterComplianceUpdate builder.
func (tccuo *TrustCenterComplianceUpdateOne) Where(ps ...predicate.TrustCenterCompliance) *TrustCenterComplianceUpdateOne {
	tccuo.mutation.Where(ps...)
	return tccuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (tccuo *TrustCenterComplianceUpdateOne) Select(field string, fields ...string) *TrustCenterComplianceUpdateOne {
	tccuo.fields = append([]string{field}, fields...)
	return tccuo
}

// Save executes the query and returns the updated TrustCenterCompliance entity.
func (tccuo *TrustCenterComplianceUpdateOne) Save(ctx context.Context) (*TrustCenterCompliance, error) {
	if err := tccuo.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, tccuo.sqlSave, tccuo.mutation, tccuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (tccuo *TrustCenterComplianceUpdateOne) SaveX(ctx context.Context) *TrustCenterCompliance {
	node, err := tccuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (tccuo *TrustCenterComplianceUpdateOne) Exec(ctx context.Context) error {
	_, err := tccuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tccuo *TrustCenterComplianceUpdateOne) ExecX(ctx context.Context) {
	if err := tccuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (tccuo *TrustCenterComplianceUpdateOne) defaults() error {
	if _, ok := tccuo.mutation.UpdatedAt(); !ok && !tccuo.mutation.UpdatedAtCleared() {
		if trustcentercompliance.UpdateDefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized trustcentercompliance.UpdateDefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := trustcentercompliance.UpdateDefaultUpdatedAt()
		tccuo.mutation.SetUpdatedAt(v)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (tccuo *TrustCenterComplianceUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *TrustCenterComplianceUpdateOne {
	tccuo.modifiers = append(tccuo.modifiers, modifiers...)
	return tccuo
}

func (tccuo *TrustCenterComplianceUpdateOne) sqlSave(ctx context.Context) (_node *TrustCenterCompliance, err error) {
	_spec := sqlgraph.NewUpdateSpec(trustcentercompliance.Table, trustcentercompliance.Columns, sqlgraph.NewFieldSpec(trustcentercompliance.FieldID, field.TypeString))
	id, ok := tccuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`generated: missing "TrustCenterCompliance.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := tccuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, trustcentercompliance.FieldID)
		for _, f := range fields {
			if !trustcentercompliance.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("generated: invalid field %q for query", f)}
			}
			if f != trustcentercompliance.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := tccuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if tccuo.mutation.CreatedAtCleared() {
		_spec.ClearField(trustcentercompliance.FieldCreatedAt, field.TypeTime)
	}
	if value, ok := tccuo.mutation.UpdatedAt(); ok {
		_spec.SetField(trustcentercompliance.FieldUpdatedAt, field.TypeTime, value)
	}
	if tccuo.mutation.UpdatedAtCleared() {
		_spec.ClearField(trustcentercompliance.FieldUpdatedAt, field.TypeTime)
	}
	if tccuo.mutation.CreatedByCleared() {
		_spec.ClearField(trustcentercompliance.FieldCreatedBy, field.TypeString)
	}
	if value, ok := tccuo.mutation.UpdatedBy(); ok {
		_spec.SetField(trustcentercompliance.FieldUpdatedBy, field.TypeString, value)
	}
	if tccuo.mutation.UpdatedByCleared() {
		_spec.ClearField(trustcentercompliance.FieldUpdatedBy, field.TypeString)
	}
	if value, ok := tccuo.mutation.DeletedAt(); ok {
		_spec.SetField(trustcentercompliance.FieldDeletedAt, field.TypeTime, value)
	}
	if tccuo.mutation.DeletedAtCleared() {
		_spec.ClearField(trustcentercompliance.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := tccuo.mutation.DeletedBy(); ok {
		_spec.SetField(trustcentercompliance.FieldDeletedBy, field.TypeString, value)
	}
	if tccuo.mutation.DeletedByCleared() {
		_spec.ClearField(trustcentercompliance.FieldDeletedBy, field.TypeString)
	}
	if value, ok := tccuo.mutation.Tags(); ok {
		_spec.SetField(trustcentercompliance.FieldTags, field.TypeJSON, value)
	}
	if value, ok := tccuo.mutation.AppendedTags(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, trustcentercompliance.FieldTags, value)
		})
	}
	if tccuo.mutation.TagsCleared() {
		_spec.ClearField(trustcentercompliance.FieldTags, field.TypeJSON)
	}
	_spec.Node.Schema = tccuo.schemaConfig.TrustCenterCompliance
	ctx = internal.NewSchemaConfigContext(ctx, tccuo.schemaConfig)
	_spec.AddModifiers(tccuo.modifiers...)
	_node = &TrustCenterCompliance{config: tccuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, tccuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{trustcentercompliance.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	tccuo.mutation.done = true
	return _node, nil
}
