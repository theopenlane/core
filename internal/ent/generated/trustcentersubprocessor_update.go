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
	"github.com/theopenlane/core/internal/ent/generated/subprocessor"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/ent/generated/trustcentersubprocessor"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// TrustCenterSubprocessorUpdate is the builder for updating TrustCenterSubprocessor entities.
type TrustCenterSubprocessorUpdate struct {
	config
	hooks     []Hook
	mutation  *TrustCenterSubprocessorMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the TrustCenterSubprocessorUpdate builder.
func (tcsu *TrustCenterSubprocessorUpdate) Where(ps ...predicate.TrustCenterSubprocessor) *TrustCenterSubprocessorUpdate {
	tcsu.mutation.Where(ps...)
	return tcsu
}

// SetUpdatedAt sets the "updated_at" field.
func (tcsu *TrustCenterSubprocessorUpdate) SetUpdatedAt(t time.Time) *TrustCenterSubprocessorUpdate {
	tcsu.mutation.SetUpdatedAt(t)
	return tcsu
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (tcsu *TrustCenterSubprocessorUpdate) ClearUpdatedAt() *TrustCenterSubprocessorUpdate {
	tcsu.mutation.ClearUpdatedAt()
	return tcsu
}

// SetUpdatedBy sets the "updated_by" field.
func (tcsu *TrustCenterSubprocessorUpdate) SetUpdatedBy(s string) *TrustCenterSubprocessorUpdate {
	tcsu.mutation.SetUpdatedBy(s)
	return tcsu
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (tcsu *TrustCenterSubprocessorUpdate) SetNillableUpdatedBy(s *string) *TrustCenterSubprocessorUpdate {
	if s != nil {
		tcsu.SetUpdatedBy(*s)
	}
	return tcsu
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (tcsu *TrustCenterSubprocessorUpdate) ClearUpdatedBy() *TrustCenterSubprocessorUpdate {
	tcsu.mutation.ClearUpdatedBy()
	return tcsu
}

// SetDeletedAt sets the "deleted_at" field.
func (tcsu *TrustCenterSubprocessorUpdate) SetDeletedAt(t time.Time) *TrustCenterSubprocessorUpdate {
	tcsu.mutation.SetDeletedAt(t)
	return tcsu
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (tcsu *TrustCenterSubprocessorUpdate) SetNillableDeletedAt(t *time.Time) *TrustCenterSubprocessorUpdate {
	if t != nil {
		tcsu.SetDeletedAt(*t)
	}
	return tcsu
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (tcsu *TrustCenterSubprocessorUpdate) ClearDeletedAt() *TrustCenterSubprocessorUpdate {
	tcsu.mutation.ClearDeletedAt()
	return tcsu
}

// SetDeletedBy sets the "deleted_by" field.
func (tcsu *TrustCenterSubprocessorUpdate) SetDeletedBy(s string) *TrustCenterSubprocessorUpdate {
	tcsu.mutation.SetDeletedBy(s)
	return tcsu
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (tcsu *TrustCenterSubprocessorUpdate) SetNillableDeletedBy(s *string) *TrustCenterSubprocessorUpdate {
	if s != nil {
		tcsu.SetDeletedBy(*s)
	}
	return tcsu
}

// ClearDeletedBy clears the value of the "deleted_by" field.
func (tcsu *TrustCenterSubprocessorUpdate) ClearDeletedBy() *TrustCenterSubprocessorUpdate {
	tcsu.mutation.ClearDeletedBy()
	return tcsu
}

// SetSubprocessorID sets the "subprocessor_id" field.
func (tcsu *TrustCenterSubprocessorUpdate) SetSubprocessorID(s string) *TrustCenterSubprocessorUpdate {
	tcsu.mutation.SetSubprocessorID(s)
	return tcsu
}

// SetNillableSubprocessorID sets the "subprocessor_id" field if the given value is not nil.
func (tcsu *TrustCenterSubprocessorUpdate) SetNillableSubprocessorID(s *string) *TrustCenterSubprocessorUpdate {
	if s != nil {
		tcsu.SetSubprocessorID(*s)
	}
	return tcsu
}

// SetTrustCenterID sets the "trust_center_id" field.
func (tcsu *TrustCenterSubprocessorUpdate) SetTrustCenterID(s string) *TrustCenterSubprocessorUpdate {
	tcsu.mutation.SetTrustCenterID(s)
	return tcsu
}

// SetNillableTrustCenterID sets the "trust_center_id" field if the given value is not nil.
func (tcsu *TrustCenterSubprocessorUpdate) SetNillableTrustCenterID(s *string) *TrustCenterSubprocessorUpdate {
	if s != nil {
		tcsu.SetTrustCenterID(*s)
	}
	return tcsu
}

// ClearTrustCenterID clears the value of the "trust_center_id" field.
func (tcsu *TrustCenterSubprocessorUpdate) ClearTrustCenterID() *TrustCenterSubprocessorUpdate {
	tcsu.mutation.ClearTrustCenterID()
	return tcsu
}

// SetCountries sets the "countries" field.
func (tcsu *TrustCenterSubprocessorUpdate) SetCountries(s []string) *TrustCenterSubprocessorUpdate {
	tcsu.mutation.SetCountries(s)
	return tcsu
}

// AppendCountries appends s to the "countries" field.
func (tcsu *TrustCenterSubprocessorUpdate) AppendCountries(s []string) *TrustCenterSubprocessorUpdate {
	tcsu.mutation.AppendCountries(s)
	return tcsu
}

// ClearCountries clears the value of the "countries" field.
func (tcsu *TrustCenterSubprocessorUpdate) ClearCountries() *TrustCenterSubprocessorUpdate {
	tcsu.mutation.ClearCountries()
	return tcsu
}

// SetCategory sets the "category" field.
func (tcsu *TrustCenterSubprocessorUpdate) SetCategory(s string) *TrustCenterSubprocessorUpdate {
	tcsu.mutation.SetCategory(s)
	return tcsu
}

// SetNillableCategory sets the "category" field if the given value is not nil.
func (tcsu *TrustCenterSubprocessorUpdate) SetNillableCategory(s *string) *TrustCenterSubprocessorUpdate {
	if s != nil {
		tcsu.SetCategory(*s)
	}
	return tcsu
}

// SetTrustCenter sets the "trust_center" edge to the TrustCenter entity.
func (tcsu *TrustCenterSubprocessorUpdate) SetTrustCenter(t *TrustCenter) *TrustCenterSubprocessorUpdate {
	return tcsu.SetTrustCenterID(t.ID)
}

// SetSubprocessor sets the "subprocessor" edge to the Subprocessor entity.
func (tcsu *TrustCenterSubprocessorUpdate) SetSubprocessor(s *Subprocessor) *TrustCenterSubprocessorUpdate {
	return tcsu.SetSubprocessorID(s.ID)
}

// Mutation returns the TrustCenterSubprocessorMutation object of the builder.
func (tcsu *TrustCenterSubprocessorUpdate) Mutation() *TrustCenterSubprocessorMutation {
	return tcsu.mutation
}

// ClearTrustCenter clears the "trust_center" edge to the TrustCenter entity.
func (tcsu *TrustCenterSubprocessorUpdate) ClearTrustCenter() *TrustCenterSubprocessorUpdate {
	tcsu.mutation.ClearTrustCenter()
	return tcsu
}

// ClearSubprocessor clears the "subprocessor" edge to the Subprocessor entity.
func (tcsu *TrustCenterSubprocessorUpdate) ClearSubprocessor() *TrustCenterSubprocessorUpdate {
	tcsu.mutation.ClearSubprocessor()
	return tcsu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (tcsu *TrustCenterSubprocessorUpdate) Save(ctx context.Context) (int, error) {
	if err := tcsu.defaults(); err != nil {
		return 0, err
	}
	return withHooks(ctx, tcsu.sqlSave, tcsu.mutation, tcsu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (tcsu *TrustCenterSubprocessorUpdate) SaveX(ctx context.Context) int {
	affected, err := tcsu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (tcsu *TrustCenterSubprocessorUpdate) Exec(ctx context.Context) error {
	_, err := tcsu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tcsu *TrustCenterSubprocessorUpdate) ExecX(ctx context.Context) {
	if err := tcsu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (tcsu *TrustCenterSubprocessorUpdate) defaults() error {
	if _, ok := tcsu.mutation.UpdatedAt(); !ok && !tcsu.mutation.UpdatedAtCleared() {
		if trustcentersubprocessor.UpdateDefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized trustcentersubprocessor.UpdateDefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := trustcentersubprocessor.UpdateDefaultUpdatedAt()
		tcsu.mutation.SetUpdatedAt(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (tcsu *TrustCenterSubprocessorUpdate) check() error {
	if v, ok := tcsu.mutation.SubprocessorID(); ok {
		if err := trustcentersubprocessor.SubprocessorIDValidator(v); err != nil {
			return &ValidationError{Name: "subprocessor_id", err: fmt.Errorf(`generated: validator failed for field "TrustCenterSubprocessor.subprocessor_id": %w`, err)}
		}
	}
	if v, ok := tcsu.mutation.TrustCenterID(); ok {
		if err := trustcentersubprocessor.TrustCenterIDValidator(v); err != nil {
			return &ValidationError{Name: "trust_center_id", err: fmt.Errorf(`generated: validator failed for field "TrustCenterSubprocessor.trust_center_id": %w`, err)}
		}
	}
	if v, ok := tcsu.mutation.Category(); ok {
		if err := trustcentersubprocessor.CategoryValidator(v); err != nil {
			return &ValidationError{Name: "category", err: fmt.Errorf(`generated: validator failed for field "TrustCenterSubprocessor.category": %w`, err)}
		}
	}
	if tcsu.mutation.SubprocessorCleared() && len(tcsu.mutation.SubprocessorIDs()) > 0 {
		return errors.New(`generated: clearing a required unique edge "TrustCenterSubprocessor.subprocessor"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (tcsu *TrustCenterSubprocessorUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *TrustCenterSubprocessorUpdate {
	tcsu.modifiers = append(tcsu.modifiers, modifiers...)
	return tcsu
}

func (tcsu *TrustCenterSubprocessorUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := tcsu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(trustcentersubprocessor.Table, trustcentersubprocessor.Columns, sqlgraph.NewFieldSpec(trustcentersubprocessor.FieldID, field.TypeString))
	if ps := tcsu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if tcsu.mutation.CreatedAtCleared() {
		_spec.ClearField(trustcentersubprocessor.FieldCreatedAt, field.TypeTime)
	}
	if value, ok := tcsu.mutation.UpdatedAt(); ok {
		_spec.SetField(trustcentersubprocessor.FieldUpdatedAt, field.TypeTime, value)
	}
	if tcsu.mutation.UpdatedAtCleared() {
		_spec.ClearField(trustcentersubprocessor.FieldUpdatedAt, field.TypeTime)
	}
	if tcsu.mutation.CreatedByCleared() {
		_spec.ClearField(trustcentersubprocessor.FieldCreatedBy, field.TypeString)
	}
	if value, ok := tcsu.mutation.UpdatedBy(); ok {
		_spec.SetField(trustcentersubprocessor.FieldUpdatedBy, field.TypeString, value)
	}
	if tcsu.mutation.UpdatedByCleared() {
		_spec.ClearField(trustcentersubprocessor.FieldUpdatedBy, field.TypeString)
	}
	if value, ok := tcsu.mutation.DeletedAt(); ok {
		_spec.SetField(trustcentersubprocessor.FieldDeletedAt, field.TypeTime, value)
	}
	if tcsu.mutation.DeletedAtCleared() {
		_spec.ClearField(trustcentersubprocessor.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := tcsu.mutation.DeletedBy(); ok {
		_spec.SetField(trustcentersubprocessor.FieldDeletedBy, field.TypeString, value)
	}
	if tcsu.mutation.DeletedByCleared() {
		_spec.ClearField(trustcentersubprocessor.FieldDeletedBy, field.TypeString)
	}
	if value, ok := tcsu.mutation.Countries(); ok {
		_spec.SetField(trustcentersubprocessor.FieldCountries, field.TypeJSON, value)
	}
	if value, ok := tcsu.mutation.AppendedCountries(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, trustcentersubprocessor.FieldCountries, value)
		})
	}
	if tcsu.mutation.CountriesCleared() {
		_spec.ClearField(trustcentersubprocessor.FieldCountries, field.TypeJSON)
	}
	if value, ok := tcsu.mutation.Category(); ok {
		_spec.SetField(trustcentersubprocessor.FieldCategory, field.TypeString, value)
	}
	if tcsu.mutation.TrustCenterCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   trustcentersubprocessor.TrustCenterTable,
			Columns: []string{trustcentersubprocessor.TrustCenterColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(trustcenter.FieldID, field.TypeString),
			},
		}
		edge.Schema = tcsu.schemaConfig.TrustCenterSubprocessor
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := tcsu.mutation.TrustCenterIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   trustcentersubprocessor.TrustCenterTable,
			Columns: []string{trustcentersubprocessor.TrustCenterColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(trustcenter.FieldID, field.TypeString),
			},
		}
		edge.Schema = tcsu.schemaConfig.TrustCenterSubprocessor
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if tcsu.mutation.SubprocessorCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   trustcentersubprocessor.SubprocessorTable,
			Columns: []string{trustcentersubprocessor.SubprocessorColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(subprocessor.FieldID, field.TypeString),
			},
		}
		edge.Schema = tcsu.schemaConfig.TrustCenterSubprocessor
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := tcsu.mutation.SubprocessorIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   trustcentersubprocessor.SubprocessorTable,
			Columns: []string{trustcentersubprocessor.SubprocessorColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(subprocessor.FieldID, field.TypeString),
			},
		}
		edge.Schema = tcsu.schemaConfig.TrustCenterSubprocessor
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.Node.Schema = tcsu.schemaConfig.TrustCenterSubprocessor
	ctx = internal.NewSchemaConfigContext(ctx, tcsu.schemaConfig)
	_spec.AddModifiers(tcsu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, tcsu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{trustcentersubprocessor.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	tcsu.mutation.done = true
	return n, nil
}

// TrustCenterSubprocessorUpdateOne is the builder for updating a single TrustCenterSubprocessor entity.
type TrustCenterSubprocessorUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *TrustCenterSubprocessorMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetUpdatedAt sets the "updated_at" field.
func (tcsuo *TrustCenterSubprocessorUpdateOne) SetUpdatedAt(t time.Time) *TrustCenterSubprocessorUpdateOne {
	tcsuo.mutation.SetUpdatedAt(t)
	return tcsuo
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (tcsuo *TrustCenterSubprocessorUpdateOne) ClearUpdatedAt() *TrustCenterSubprocessorUpdateOne {
	tcsuo.mutation.ClearUpdatedAt()
	return tcsuo
}

// SetUpdatedBy sets the "updated_by" field.
func (tcsuo *TrustCenterSubprocessorUpdateOne) SetUpdatedBy(s string) *TrustCenterSubprocessorUpdateOne {
	tcsuo.mutation.SetUpdatedBy(s)
	return tcsuo
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (tcsuo *TrustCenterSubprocessorUpdateOne) SetNillableUpdatedBy(s *string) *TrustCenterSubprocessorUpdateOne {
	if s != nil {
		tcsuo.SetUpdatedBy(*s)
	}
	return tcsuo
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (tcsuo *TrustCenterSubprocessorUpdateOne) ClearUpdatedBy() *TrustCenterSubprocessorUpdateOne {
	tcsuo.mutation.ClearUpdatedBy()
	return tcsuo
}

// SetDeletedAt sets the "deleted_at" field.
func (tcsuo *TrustCenterSubprocessorUpdateOne) SetDeletedAt(t time.Time) *TrustCenterSubprocessorUpdateOne {
	tcsuo.mutation.SetDeletedAt(t)
	return tcsuo
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (tcsuo *TrustCenterSubprocessorUpdateOne) SetNillableDeletedAt(t *time.Time) *TrustCenterSubprocessorUpdateOne {
	if t != nil {
		tcsuo.SetDeletedAt(*t)
	}
	return tcsuo
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (tcsuo *TrustCenterSubprocessorUpdateOne) ClearDeletedAt() *TrustCenterSubprocessorUpdateOne {
	tcsuo.mutation.ClearDeletedAt()
	return tcsuo
}

// SetDeletedBy sets the "deleted_by" field.
func (tcsuo *TrustCenterSubprocessorUpdateOne) SetDeletedBy(s string) *TrustCenterSubprocessorUpdateOne {
	tcsuo.mutation.SetDeletedBy(s)
	return tcsuo
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (tcsuo *TrustCenterSubprocessorUpdateOne) SetNillableDeletedBy(s *string) *TrustCenterSubprocessorUpdateOne {
	if s != nil {
		tcsuo.SetDeletedBy(*s)
	}
	return tcsuo
}

// ClearDeletedBy clears the value of the "deleted_by" field.
func (tcsuo *TrustCenterSubprocessorUpdateOne) ClearDeletedBy() *TrustCenterSubprocessorUpdateOne {
	tcsuo.mutation.ClearDeletedBy()
	return tcsuo
}

// SetSubprocessorID sets the "subprocessor_id" field.
func (tcsuo *TrustCenterSubprocessorUpdateOne) SetSubprocessorID(s string) *TrustCenterSubprocessorUpdateOne {
	tcsuo.mutation.SetSubprocessorID(s)
	return tcsuo
}

// SetNillableSubprocessorID sets the "subprocessor_id" field if the given value is not nil.
func (tcsuo *TrustCenterSubprocessorUpdateOne) SetNillableSubprocessorID(s *string) *TrustCenterSubprocessorUpdateOne {
	if s != nil {
		tcsuo.SetSubprocessorID(*s)
	}
	return tcsuo
}

// SetTrustCenterID sets the "trust_center_id" field.
func (tcsuo *TrustCenterSubprocessorUpdateOne) SetTrustCenterID(s string) *TrustCenterSubprocessorUpdateOne {
	tcsuo.mutation.SetTrustCenterID(s)
	return tcsuo
}

// SetNillableTrustCenterID sets the "trust_center_id" field if the given value is not nil.
func (tcsuo *TrustCenterSubprocessorUpdateOne) SetNillableTrustCenterID(s *string) *TrustCenterSubprocessorUpdateOne {
	if s != nil {
		tcsuo.SetTrustCenterID(*s)
	}
	return tcsuo
}

// ClearTrustCenterID clears the value of the "trust_center_id" field.
func (tcsuo *TrustCenterSubprocessorUpdateOne) ClearTrustCenterID() *TrustCenterSubprocessorUpdateOne {
	tcsuo.mutation.ClearTrustCenterID()
	return tcsuo
}

// SetCountries sets the "countries" field.
func (tcsuo *TrustCenterSubprocessorUpdateOne) SetCountries(s []string) *TrustCenterSubprocessorUpdateOne {
	tcsuo.mutation.SetCountries(s)
	return tcsuo
}

// AppendCountries appends s to the "countries" field.
func (tcsuo *TrustCenterSubprocessorUpdateOne) AppendCountries(s []string) *TrustCenterSubprocessorUpdateOne {
	tcsuo.mutation.AppendCountries(s)
	return tcsuo
}

// ClearCountries clears the value of the "countries" field.
func (tcsuo *TrustCenterSubprocessorUpdateOne) ClearCountries() *TrustCenterSubprocessorUpdateOne {
	tcsuo.mutation.ClearCountries()
	return tcsuo
}

// SetCategory sets the "category" field.
func (tcsuo *TrustCenterSubprocessorUpdateOne) SetCategory(s string) *TrustCenterSubprocessorUpdateOne {
	tcsuo.mutation.SetCategory(s)
	return tcsuo
}

// SetNillableCategory sets the "category" field if the given value is not nil.
func (tcsuo *TrustCenterSubprocessorUpdateOne) SetNillableCategory(s *string) *TrustCenterSubprocessorUpdateOne {
	if s != nil {
		tcsuo.SetCategory(*s)
	}
	return tcsuo
}

// SetTrustCenter sets the "trust_center" edge to the TrustCenter entity.
func (tcsuo *TrustCenterSubprocessorUpdateOne) SetTrustCenter(t *TrustCenter) *TrustCenterSubprocessorUpdateOne {
	return tcsuo.SetTrustCenterID(t.ID)
}

// SetSubprocessor sets the "subprocessor" edge to the Subprocessor entity.
func (tcsuo *TrustCenterSubprocessorUpdateOne) SetSubprocessor(s *Subprocessor) *TrustCenterSubprocessorUpdateOne {
	return tcsuo.SetSubprocessorID(s.ID)
}

// Mutation returns the TrustCenterSubprocessorMutation object of the builder.
func (tcsuo *TrustCenterSubprocessorUpdateOne) Mutation() *TrustCenterSubprocessorMutation {
	return tcsuo.mutation
}

// ClearTrustCenter clears the "trust_center" edge to the TrustCenter entity.
func (tcsuo *TrustCenterSubprocessorUpdateOne) ClearTrustCenter() *TrustCenterSubprocessorUpdateOne {
	tcsuo.mutation.ClearTrustCenter()
	return tcsuo
}

// ClearSubprocessor clears the "subprocessor" edge to the Subprocessor entity.
func (tcsuo *TrustCenterSubprocessorUpdateOne) ClearSubprocessor() *TrustCenterSubprocessorUpdateOne {
	tcsuo.mutation.ClearSubprocessor()
	return tcsuo
}

// Where appends a list predicates to the TrustCenterSubprocessorUpdate builder.
func (tcsuo *TrustCenterSubprocessorUpdateOne) Where(ps ...predicate.TrustCenterSubprocessor) *TrustCenterSubprocessorUpdateOne {
	tcsuo.mutation.Where(ps...)
	return tcsuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (tcsuo *TrustCenterSubprocessorUpdateOne) Select(field string, fields ...string) *TrustCenterSubprocessorUpdateOne {
	tcsuo.fields = append([]string{field}, fields...)
	return tcsuo
}

// Save executes the query and returns the updated TrustCenterSubprocessor entity.
func (tcsuo *TrustCenterSubprocessorUpdateOne) Save(ctx context.Context) (*TrustCenterSubprocessor, error) {
	if err := tcsuo.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, tcsuo.sqlSave, tcsuo.mutation, tcsuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (tcsuo *TrustCenterSubprocessorUpdateOne) SaveX(ctx context.Context) *TrustCenterSubprocessor {
	node, err := tcsuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (tcsuo *TrustCenterSubprocessorUpdateOne) Exec(ctx context.Context) error {
	_, err := tcsuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tcsuo *TrustCenterSubprocessorUpdateOne) ExecX(ctx context.Context) {
	if err := tcsuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (tcsuo *TrustCenterSubprocessorUpdateOne) defaults() error {
	if _, ok := tcsuo.mutation.UpdatedAt(); !ok && !tcsuo.mutation.UpdatedAtCleared() {
		if trustcentersubprocessor.UpdateDefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized trustcentersubprocessor.UpdateDefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := trustcentersubprocessor.UpdateDefaultUpdatedAt()
		tcsuo.mutation.SetUpdatedAt(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (tcsuo *TrustCenterSubprocessorUpdateOne) check() error {
	if v, ok := tcsuo.mutation.SubprocessorID(); ok {
		if err := trustcentersubprocessor.SubprocessorIDValidator(v); err != nil {
			return &ValidationError{Name: "subprocessor_id", err: fmt.Errorf(`generated: validator failed for field "TrustCenterSubprocessor.subprocessor_id": %w`, err)}
		}
	}
	if v, ok := tcsuo.mutation.TrustCenterID(); ok {
		if err := trustcentersubprocessor.TrustCenterIDValidator(v); err != nil {
			return &ValidationError{Name: "trust_center_id", err: fmt.Errorf(`generated: validator failed for field "TrustCenterSubprocessor.trust_center_id": %w`, err)}
		}
	}
	if v, ok := tcsuo.mutation.Category(); ok {
		if err := trustcentersubprocessor.CategoryValidator(v); err != nil {
			return &ValidationError{Name: "category", err: fmt.Errorf(`generated: validator failed for field "TrustCenterSubprocessor.category": %w`, err)}
		}
	}
	if tcsuo.mutation.SubprocessorCleared() && len(tcsuo.mutation.SubprocessorIDs()) > 0 {
		return errors.New(`generated: clearing a required unique edge "TrustCenterSubprocessor.subprocessor"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (tcsuo *TrustCenterSubprocessorUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *TrustCenterSubprocessorUpdateOne {
	tcsuo.modifiers = append(tcsuo.modifiers, modifiers...)
	return tcsuo
}

func (tcsuo *TrustCenterSubprocessorUpdateOne) sqlSave(ctx context.Context) (_node *TrustCenterSubprocessor, err error) {
	if err := tcsuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(trustcentersubprocessor.Table, trustcentersubprocessor.Columns, sqlgraph.NewFieldSpec(trustcentersubprocessor.FieldID, field.TypeString))
	id, ok := tcsuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`generated: missing "TrustCenterSubprocessor.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := tcsuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, trustcentersubprocessor.FieldID)
		for _, f := range fields {
			if !trustcentersubprocessor.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("generated: invalid field %q for query", f)}
			}
			if f != trustcentersubprocessor.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := tcsuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if tcsuo.mutation.CreatedAtCleared() {
		_spec.ClearField(trustcentersubprocessor.FieldCreatedAt, field.TypeTime)
	}
	if value, ok := tcsuo.mutation.UpdatedAt(); ok {
		_spec.SetField(trustcentersubprocessor.FieldUpdatedAt, field.TypeTime, value)
	}
	if tcsuo.mutation.UpdatedAtCleared() {
		_spec.ClearField(trustcentersubprocessor.FieldUpdatedAt, field.TypeTime)
	}
	if tcsuo.mutation.CreatedByCleared() {
		_spec.ClearField(trustcentersubprocessor.FieldCreatedBy, field.TypeString)
	}
	if value, ok := tcsuo.mutation.UpdatedBy(); ok {
		_spec.SetField(trustcentersubprocessor.FieldUpdatedBy, field.TypeString, value)
	}
	if tcsuo.mutation.UpdatedByCleared() {
		_spec.ClearField(trustcentersubprocessor.FieldUpdatedBy, field.TypeString)
	}
	if value, ok := tcsuo.mutation.DeletedAt(); ok {
		_spec.SetField(trustcentersubprocessor.FieldDeletedAt, field.TypeTime, value)
	}
	if tcsuo.mutation.DeletedAtCleared() {
		_spec.ClearField(trustcentersubprocessor.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := tcsuo.mutation.DeletedBy(); ok {
		_spec.SetField(trustcentersubprocessor.FieldDeletedBy, field.TypeString, value)
	}
	if tcsuo.mutation.DeletedByCleared() {
		_spec.ClearField(trustcentersubprocessor.FieldDeletedBy, field.TypeString)
	}
	if value, ok := tcsuo.mutation.Countries(); ok {
		_spec.SetField(trustcentersubprocessor.FieldCountries, field.TypeJSON, value)
	}
	if value, ok := tcsuo.mutation.AppendedCountries(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, trustcentersubprocessor.FieldCountries, value)
		})
	}
	if tcsuo.mutation.CountriesCleared() {
		_spec.ClearField(trustcentersubprocessor.FieldCountries, field.TypeJSON)
	}
	if value, ok := tcsuo.mutation.Category(); ok {
		_spec.SetField(trustcentersubprocessor.FieldCategory, field.TypeString, value)
	}
	if tcsuo.mutation.TrustCenterCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   trustcentersubprocessor.TrustCenterTable,
			Columns: []string{trustcentersubprocessor.TrustCenterColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(trustcenter.FieldID, field.TypeString),
			},
		}
		edge.Schema = tcsuo.schemaConfig.TrustCenterSubprocessor
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := tcsuo.mutation.TrustCenterIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   trustcentersubprocessor.TrustCenterTable,
			Columns: []string{trustcentersubprocessor.TrustCenterColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(trustcenter.FieldID, field.TypeString),
			},
		}
		edge.Schema = tcsuo.schemaConfig.TrustCenterSubprocessor
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if tcsuo.mutation.SubprocessorCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   trustcentersubprocessor.SubprocessorTable,
			Columns: []string{trustcentersubprocessor.SubprocessorColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(subprocessor.FieldID, field.TypeString),
			},
		}
		edge.Schema = tcsuo.schemaConfig.TrustCenterSubprocessor
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := tcsuo.mutation.SubprocessorIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   trustcentersubprocessor.SubprocessorTable,
			Columns: []string{trustcentersubprocessor.SubprocessorColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(subprocessor.FieldID, field.TypeString),
			},
		}
		edge.Schema = tcsuo.schemaConfig.TrustCenterSubprocessor
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.Node.Schema = tcsuo.schemaConfig.TrustCenterSubprocessor
	ctx = internal.NewSchemaConfigContext(ctx, tcsuo.schemaConfig)
	_spec.AddModifiers(tcsuo.modifiers...)
	_node = &TrustCenterSubprocessor{config: tcsuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, tcsuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{trustcentersubprocessor.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	tcsuo.mutation.done = true
	return _node, nil
}
