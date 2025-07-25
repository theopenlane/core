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
	"github.com/theopenlane/core/internal/ent/generated/trustcentersubprocessorhistory"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// TrustCenterSubprocessorHistoryUpdate is the builder for updating TrustCenterSubprocessorHistory entities.
type TrustCenterSubprocessorHistoryUpdate struct {
	config
	hooks     []Hook
	mutation  *TrustCenterSubprocessorHistoryMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the TrustCenterSubprocessorHistoryUpdate builder.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) Where(ps ...predicate.TrustCenterSubprocessorHistory) *TrustCenterSubprocessorHistoryUpdate {
	tcshu.mutation.Where(ps...)
	return tcshu
}

// SetUpdatedAt sets the "updated_at" field.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) SetUpdatedAt(t time.Time) *TrustCenterSubprocessorHistoryUpdate {
	tcshu.mutation.SetUpdatedAt(t)
	return tcshu
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) ClearUpdatedAt() *TrustCenterSubprocessorHistoryUpdate {
	tcshu.mutation.ClearUpdatedAt()
	return tcshu
}

// SetUpdatedBy sets the "updated_by" field.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) SetUpdatedBy(s string) *TrustCenterSubprocessorHistoryUpdate {
	tcshu.mutation.SetUpdatedBy(s)
	return tcshu
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) SetNillableUpdatedBy(s *string) *TrustCenterSubprocessorHistoryUpdate {
	if s != nil {
		tcshu.SetUpdatedBy(*s)
	}
	return tcshu
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) ClearUpdatedBy() *TrustCenterSubprocessorHistoryUpdate {
	tcshu.mutation.ClearUpdatedBy()
	return tcshu
}

// SetDeletedAt sets the "deleted_at" field.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) SetDeletedAt(t time.Time) *TrustCenterSubprocessorHistoryUpdate {
	tcshu.mutation.SetDeletedAt(t)
	return tcshu
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) SetNillableDeletedAt(t *time.Time) *TrustCenterSubprocessorHistoryUpdate {
	if t != nil {
		tcshu.SetDeletedAt(*t)
	}
	return tcshu
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) ClearDeletedAt() *TrustCenterSubprocessorHistoryUpdate {
	tcshu.mutation.ClearDeletedAt()
	return tcshu
}

// SetDeletedBy sets the "deleted_by" field.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) SetDeletedBy(s string) *TrustCenterSubprocessorHistoryUpdate {
	tcshu.mutation.SetDeletedBy(s)
	return tcshu
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) SetNillableDeletedBy(s *string) *TrustCenterSubprocessorHistoryUpdate {
	if s != nil {
		tcshu.SetDeletedBy(*s)
	}
	return tcshu
}

// ClearDeletedBy clears the value of the "deleted_by" field.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) ClearDeletedBy() *TrustCenterSubprocessorHistoryUpdate {
	tcshu.mutation.ClearDeletedBy()
	return tcshu
}

// SetSubprocessorID sets the "subprocessor_id" field.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) SetSubprocessorID(s string) *TrustCenterSubprocessorHistoryUpdate {
	tcshu.mutation.SetSubprocessorID(s)
	return tcshu
}

// SetNillableSubprocessorID sets the "subprocessor_id" field if the given value is not nil.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) SetNillableSubprocessorID(s *string) *TrustCenterSubprocessorHistoryUpdate {
	if s != nil {
		tcshu.SetSubprocessorID(*s)
	}
	return tcshu
}

// SetTrustCenterID sets the "trust_center_id" field.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) SetTrustCenterID(s string) *TrustCenterSubprocessorHistoryUpdate {
	tcshu.mutation.SetTrustCenterID(s)
	return tcshu
}

// SetNillableTrustCenterID sets the "trust_center_id" field if the given value is not nil.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) SetNillableTrustCenterID(s *string) *TrustCenterSubprocessorHistoryUpdate {
	if s != nil {
		tcshu.SetTrustCenterID(*s)
	}
	return tcshu
}

// ClearTrustCenterID clears the value of the "trust_center_id" field.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) ClearTrustCenterID() *TrustCenterSubprocessorHistoryUpdate {
	tcshu.mutation.ClearTrustCenterID()
	return tcshu
}

// SetCountries sets the "countries" field.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) SetCountries(s []string) *TrustCenterSubprocessorHistoryUpdate {
	tcshu.mutation.SetCountries(s)
	return tcshu
}

// AppendCountries appends s to the "countries" field.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) AppendCountries(s []string) *TrustCenterSubprocessorHistoryUpdate {
	tcshu.mutation.AppendCountries(s)
	return tcshu
}

// ClearCountries clears the value of the "countries" field.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) ClearCountries() *TrustCenterSubprocessorHistoryUpdate {
	tcshu.mutation.ClearCountries()
	return tcshu
}

// SetCategory sets the "category" field.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) SetCategory(s string) *TrustCenterSubprocessorHistoryUpdate {
	tcshu.mutation.SetCategory(s)
	return tcshu
}

// SetNillableCategory sets the "category" field if the given value is not nil.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) SetNillableCategory(s *string) *TrustCenterSubprocessorHistoryUpdate {
	if s != nil {
		tcshu.SetCategory(*s)
	}
	return tcshu
}

// Mutation returns the TrustCenterSubprocessorHistoryMutation object of the builder.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) Mutation() *TrustCenterSubprocessorHistoryMutation {
	return tcshu.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) Save(ctx context.Context) (int, error) {
	if err := tcshu.defaults(); err != nil {
		return 0, err
	}
	return withHooks(ctx, tcshu.sqlSave, tcshu.mutation, tcshu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) SaveX(ctx context.Context) int {
	affected, err := tcshu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) Exec(ctx context.Context) error {
	_, err := tcshu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) ExecX(ctx context.Context) {
	if err := tcshu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) defaults() error {
	if _, ok := tcshu.mutation.UpdatedAt(); !ok && !tcshu.mutation.UpdatedAtCleared() {
		if trustcentersubprocessorhistory.UpdateDefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized trustcentersubprocessorhistory.UpdateDefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := trustcentersubprocessorhistory.UpdateDefaultUpdatedAt()
		tcshu.mutation.SetUpdatedAt(v)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (tcshu *TrustCenterSubprocessorHistoryUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *TrustCenterSubprocessorHistoryUpdate {
	tcshu.modifiers = append(tcshu.modifiers, modifiers...)
	return tcshu
}

func (tcshu *TrustCenterSubprocessorHistoryUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(trustcentersubprocessorhistory.Table, trustcentersubprocessorhistory.Columns, sqlgraph.NewFieldSpec(trustcentersubprocessorhistory.FieldID, field.TypeString))
	if ps := tcshu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if tcshu.mutation.RefCleared() {
		_spec.ClearField(trustcentersubprocessorhistory.FieldRef, field.TypeString)
	}
	if tcshu.mutation.CreatedAtCleared() {
		_spec.ClearField(trustcentersubprocessorhistory.FieldCreatedAt, field.TypeTime)
	}
	if value, ok := tcshu.mutation.UpdatedAt(); ok {
		_spec.SetField(trustcentersubprocessorhistory.FieldUpdatedAt, field.TypeTime, value)
	}
	if tcshu.mutation.UpdatedAtCleared() {
		_spec.ClearField(trustcentersubprocessorhistory.FieldUpdatedAt, field.TypeTime)
	}
	if tcshu.mutation.CreatedByCleared() {
		_spec.ClearField(trustcentersubprocessorhistory.FieldCreatedBy, field.TypeString)
	}
	if value, ok := tcshu.mutation.UpdatedBy(); ok {
		_spec.SetField(trustcentersubprocessorhistory.FieldUpdatedBy, field.TypeString, value)
	}
	if tcshu.mutation.UpdatedByCleared() {
		_spec.ClearField(trustcentersubprocessorhistory.FieldUpdatedBy, field.TypeString)
	}
	if value, ok := tcshu.mutation.DeletedAt(); ok {
		_spec.SetField(trustcentersubprocessorhistory.FieldDeletedAt, field.TypeTime, value)
	}
	if tcshu.mutation.DeletedAtCleared() {
		_spec.ClearField(trustcentersubprocessorhistory.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := tcshu.mutation.DeletedBy(); ok {
		_spec.SetField(trustcentersubprocessorhistory.FieldDeletedBy, field.TypeString, value)
	}
	if tcshu.mutation.DeletedByCleared() {
		_spec.ClearField(trustcentersubprocessorhistory.FieldDeletedBy, field.TypeString)
	}
	if value, ok := tcshu.mutation.SubprocessorID(); ok {
		_spec.SetField(trustcentersubprocessorhistory.FieldSubprocessorID, field.TypeString, value)
	}
	if value, ok := tcshu.mutation.TrustCenterID(); ok {
		_spec.SetField(trustcentersubprocessorhistory.FieldTrustCenterID, field.TypeString, value)
	}
	if tcshu.mutation.TrustCenterIDCleared() {
		_spec.ClearField(trustcentersubprocessorhistory.FieldTrustCenterID, field.TypeString)
	}
	if value, ok := tcshu.mutation.Countries(); ok {
		_spec.SetField(trustcentersubprocessorhistory.FieldCountries, field.TypeJSON, value)
	}
	if value, ok := tcshu.mutation.AppendedCountries(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, trustcentersubprocessorhistory.FieldCountries, value)
		})
	}
	if tcshu.mutation.CountriesCleared() {
		_spec.ClearField(trustcentersubprocessorhistory.FieldCountries, field.TypeJSON)
	}
	if value, ok := tcshu.mutation.Category(); ok {
		_spec.SetField(trustcentersubprocessorhistory.FieldCategory, field.TypeString, value)
	}
	_spec.Node.Schema = tcshu.schemaConfig.TrustCenterSubprocessorHistory
	ctx = internal.NewSchemaConfigContext(ctx, tcshu.schemaConfig)
	_spec.AddModifiers(tcshu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, tcshu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{trustcentersubprocessorhistory.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	tcshu.mutation.done = true
	return n, nil
}

// TrustCenterSubprocessorHistoryUpdateOne is the builder for updating a single TrustCenterSubprocessorHistory entity.
type TrustCenterSubprocessorHistoryUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *TrustCenterSubprocessorHistoryMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetUpdatedAt sets the "updated_at" field.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) SetUpdatedAt(t time.Time) *TrustCenterSubprocessorHistoryUpdateOne {
	tcshuo.mutation.SetUpdatedAt(t)
	return tcshuo
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) ClearUpdatedAt() *TrustCenterSubprocessorHistoryUpdateOne {
	tcshuo.mutation.ClearUpdatedAt()
	return tcshuo
}

// SetUpdatedBy sets the "updated_by" field.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) SetUpdatedBy(s string) *TrustCenterSubprocessorHistoryUpdateOne {
	tcshuo.mutation.SetUpdatedBy(s)
	return tcshuo
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) SetNillableUpdatedBy(s *string) *TrustCenterSubprocessorHistoryUpdateOne {
	if s != nil {
		tcshuo.SetUpdatedBy(*s)
	}
	return tcshuo
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) ClearUpdatedBy() *TrustCenterSubprocessorHistoryUpdateOne {
	tcshuo.mutation.ClearUpdatedBy()
	return tcshuo
}

// SetDeletedAt sets the "deleted_at" field.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) SetDeletedAt(t time.Time) *TrustCenterSubprocessorHistoryUpdateOne {
	tcshuo.mutation.SetDeletedAt(t)
	return tcshuo
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) SetNillableDeletedAt(t *time.Time) *TrustCenterSubprocessorHistoryUpdateOne {
	if t != nil {
		tcshuo.SetDeletedAt(*t)
	}
	return tcshuo
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) ClearDeletedAt() *TrustCenterSubprocessorHistoryUpdateOne {
	tcshuo.mutation.ClearDeletedAt()
	return tcshuo
}

// SetDeletedBy sets the "deleted_by" field.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) SetDeletedBy(s string) *TrustCenterSubprocessorHistoryUpdateOne {
	tcshuo.mutation.SetDeletedBy(s)
	return tcshuo
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) SetNillableDeletedBy(s *string) *TrustCenterSubprocessorHistoryUpdateOne {
	if s != nil {
		tcshuo.SetDeletedBy(*s)
	}
	return tcshuo
}

// ClearDeletedBy clears the value of the "deleted_by" field.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) ClearDeletedBy() *TrustCenterSubprocessorHistoryUpdateOne {
	tcshuo.mutation.ClearDeletedBy()
	return tcshuo
}

// SetSubprocessorID sets the "subprocessor_id" field.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) SetSubprocessorID(s string) *TrustCenterSubprocessorHistoryUpdateOne {
	tcshuo.mutation.SetSubprocessorID(s)
	return tcshuo
}

// SetNillableSubprocessorID sets the "subprocessor_id" field if the given value is not nil.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) SetNillableSubprocessorID(s *string) *TrustCenterSubprocessorHistoryUpdateOne {
	if s != nil {
		tcshuo.SetSubprocessorID(*s)
	}
	return tcshuo
}

// SetTrustCenterID sets the "trust_center_id" field.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) SetTrustCenterID(s string) *TrustCenterSubprocessorHistoryUpdateOne {
	tcshuo.mutation.SetTrustCenterID(s)
	return tcshuo
}

// SetNillableTrustCenterID sets the "trust_center_id" field if the given value is not nil.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) SetNillableTrustCenterID(s *string) *TrustCenterSubprocessorHistoryUpdateOne {
	if s != nil {
		tcshuo.SetTrustCenterID(*s)
	}
	return tcshuo
}

// ClearTrustCenterID clears the value of the "trust_center_id" field.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) ClearTrustCenterID() *TrustCenterSubprocessorHistoryUpdateOne {
	tcshuo.mutation.ClearTrustCenterID()
	return tcshuo
}

// SetCountries sets the "countries" field.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) SetCountries(s []string) *TrustCenterSubprocessorHistoryUpdateOne {
	tcshuo.mutation.SetCountries(s)
	return tcshuo
}

// AppendCountries appends s to the "countries" field.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) AppendCountries(s []string) *TrustCenterSubprocessorHistoryUpdateOne {
	tcshuo.mutation.AppendCountries(s)
	return tcshuo
}

// ClearCountries clears the value of the "countries" field.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) ClearCountries() *TrustCenterSubprocessorHistoryUpdateOne {
	tcshuo.mutation.ClearCountries()
	return tcshuo
}

// SetCategory sets the "category" field.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) SetCategory(s string) *TrustCenterSubprocessorHistoryUpdateOne {
	tcshuo.mutation.SetCategory(s)
	return tcshuo
}

// SetNillableCategory sets the "category" field if the given value is not nil.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) SetNillableCategory(s *string) *TrustCenterSubprocessorHistoryUpdateOne {
	if s != nil {
		tcshuo.SetCategory(*s)
	}
	return tcshuo
}

// Mutation returns the TrustCenterSubprocessorHistoryMutation object of the builder.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) Mutation() *TrustCenterSubprocessorHistoryMutation {
	return tcshuo.mutation
}

// Where appends a list predicates to the TrustCenterSubprocessorHistoryUpdate builder.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) Where(ps ...predicate.TrustCenterSubprocessorHistory) *TrustCenterSubprocessorHistoryUpdateOne {
	tcshuo.mutation.Where(ps...)
	return tcshuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) Select(field string, fields ...string) *TrustCenterSubprocessorHistoryUpdateOne {
	tcshuo.fields = append([]string{field}, fields...)
	return tcshuo
}

// Save executes the query and returns the updated TrustCenterSubprocessorHistory entity.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) Save(ctx context.Context) (*TrustCenterSubprocessorHistory, error) {
	if err := tcshuo.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, tcshuo.sqlSave, tcshuo.mutation, tcshuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) SaveX(ctx context.Context) *TrustCenterSubprocessorHistory {
	node, err := tcshuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) Exec(ctx context.Context) error {
	_, err := tcshuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) ExecX(ctx context.Context) {
	if err := tcshuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) defaults() error {
	if _, ok := tcshuo.mutation.UpdatedAt(); !ok && !tcshuo.mutation.UpdatedAtCleared() {
		if trustcentersubprocessorhistory.UpdateDefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized trustcentersubprocessorhistory.UpdateDefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := trustcentersubprocessorhistory.UpdateDefaultUpdatedAt()
		tcshuo.mutation.SetUpdatedAt(v)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *TrustCenterSubprocessorHistoryUpdateOne {
	tcshuo.modifiers = append(tcshuo.modifiers, modifiers...)
	return tcshuo
}

func (tcshuo *TrustCenterSubprocessorHistoryUpdateOne) sqlSave(ctx context.Context) (_node *TrustCenterSubprocessorHistory, err error) {
	_spec := sqlgraph.NewUpdateSpec(trustcentersubprocessorhistory.Table, trustcentersubprocessorhistory.Columns, sqlgraph.NewFieldSpec(trustcentersubprocessorhistory.FieldID, field.TypeString))
	id, ok := tcshuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`generated: missing "TrustCenterSubprocessorHistory.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := tcshuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, trustcentersubprocessorhistory.FieldID)
		for _, f := range fields {
			if !trustcentersubprocessorhistory.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("generated: invalid field %q for query", f)}
			}
			if f != trustcentersubprocessorhistory.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := tcshuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if tcshuo.mutation.RefCleared() {
		_spec.ClearField(trustcentersubprocessorhistory.FieldRef, field.TypeString)
	}
	if tcshuo.mutation.CreatedAtCleared() {
		_spec.ClearField(trustcentersubprocessorhistory.FieldCreatedAt, field.TypeTime)
	}
	if value, ok := tcshuo.mutation.UpdatedAt(); ok {
		_spec.SetField(trustcentersubprocessorhistory.FieldUpdatedAt, field.TypeTime, value)
	}
	if tcshuo.mutation.UpdatedAtCleared() {
		_spec.ClearField(trustcentersubprocessorhistory.FieldUpdatedAt, field.TypeTime)
	}
	if tcshuo.mutation.CreatedByCleared() {
		_spec.ClearField(trustcentersubprocessorhistory.FieldCreatedBy, field.TypeString)
	}
	if value, ok := tcshuo.mutation.UpdatedBy(); ok {
		_spec.SetField(trustcentersubprocessorhistory.FieldUpdatedBy, field.TypeString, value)
	}
	if tcshuo.mutation.UpdatedByCleared() {
		_spec.ClearField(trustcentersubprocessorhistory.FieldUpdatedBy, field.TypeString)
	}
	if value, ok := tcshuo.mutation.DeletedAt(); ok {
		_spec.SetField(trustcentersubprocessorhistory.FieldDeletedAt, field.TypeTime, value)
	}
	if tcshuo.mutation.DeletedAtCleared() {
		_spec.ClearField(trustcentersubprocessorhistory.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := tcshuo.mutation.DeletedBy(); ok {
		_spec.SetField(trustcentersubprocessorhistory.FieldDeletedBy, field.TypeString, value)
	}
	if tcshuo.mutation.DeletedByCleared() {
		_spec.ClearField(trustcentersubprocessorhistory.FieldDeletedBy, field.TypeString)
	}
	if value, ok := tcshuo.mutation.SubprocessorID(); ok {
		_spec.SetField(trustcentersubprocessorhistory.FieldSubprocessorID, field.TypeString, value)
	}
	if value, ok := tcshuo.mutation.TrustCenterID(); ok {
		_spec.SetField(trustcentersubprocessorhistory.FieldTrustCenterID, field.TypeString, value)
	}
	if tcshuo.mutation.TrustCenterIDCleared() {
		_spec.ClearField(trustcentersubprocessorhistory.FieldTrustCenterID, field.TypeString)
	}
	if value, ok := tcshuo.mutation.Countries(); ok {
		_spec.SetField(trustcentersubprocessorhistory.FieldCountries, field.TypeJSON, value)
	}
	if value, ok := tcshuo.mutation.AppendedCountries(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, trustcentersubprocessorhistory.FieldCountries, value)
		})
	}
	if tcshuo.mutation.CountriesCleared() {
		_spec.ClearField(trustcentersubprocessorhistory.FieldCountries, field.TypeJSON)
	}
	if value, ok := tcshuo.mutation.Category(); ok {
		_spec.SetField(trustcentersubprocessorhistory.FieldCategory, field.TypeString, value)
	}
	_spec.Node.Schema = tcshuo.schemaConfig.TrustCenterSubprocessorHistory
	ctx = internal.NewSchemaConfigContext(ctx, tcshuo.schemaConfig)
	_spec.AddModifiers(tcshuo.modifiers...)
	_node = &TrustCenterSubprocessorHistory{config: tcshuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, tcshuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{trustcentersubprocessorhistory.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	tcshuo.mutation.done = true
	return _node, nil
}
