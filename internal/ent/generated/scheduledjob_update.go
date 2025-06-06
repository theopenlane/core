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
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/scheduledjob"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// ScheduledJobUpdate is the builder for updating ScheduledJob entities.
type ScheduledJobUpdate struct {
	config
	hooks     []Hook
	mutation  *ScheduledJobMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the ScheduledJobUpdate builder.
func (sju *ScheduledJobUpdate) Where(ps ...predicate.ScheduledJob) *ScheduledJobUpdate {
	sju.mutation.Where(ps...)
	return sju
}

// SetUpdatedAt sets the "updated_at" field.
func (sju *ScheduledJobUpdate) SetUpdatedAt(t time.Time) *ScheduledJobUpdate {
	sju.mutation.SetUpdatedAt(t)
	return sju
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (sju *ScheduledJobUpdate) ClearUpdatedAt() *ScheduledJobUpdate {
	sju.mutation.ClearUpdatedAt()
	return sju
}

// SetUpdatedBy sets the "updated_by" field.
func (sju *ScheduledJobUpdate) SetUpdatedBy(s string) *ScheduledJobUpdate {
	sju.mutation.SetUpdatedBy(s)
	return sju
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (sju *ScheduledJobUpdate) SetNillableUpdatedBy(s *string) *ScheduledJobUpdate {
	if s != nil {
		sju.SetUpdatedBy(*s)
	}
	return sju
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (sju *ScheduledJobUpdate) ClearUpdatedBy() *ScheduledJobUpdate {
	sju.mutation.ClearUpdatedBy()
	return sju
}

// SetDeletedAt sets the "deleted_at" field.
func (sju *ScheduledJobUpdate) SetDeletedAt(t time.Time) *ScheduledJobUpdate {
	sju.mutation.SetDeletedAt(t)
	return sju
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (sju *ScheduledJobUpdate) SetNillableDeletedAt(t *time.Time) *ScheduledJobUpdate {
	if t != nil {
		sju.SetDeletedAt(*t)
	}
	return sju
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (sju *ScheduledJobUpdate) ClearDeletedAt() *ScheduledJobUpdate {
	sju.mutation.ClearDeletedAt()
	return sju
}

// SetDeletedBy sets the "deleted_by" field.
func (sju *ScheduledJobUpdate) SetDeletedBy(s string) *ScheduledJobUpdate {
	sju.mutation.SetDeletedBy(s)
	return sju
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (sju *ScheduledJobUpdate) SetNillableDeletedBy(s *string) *ScheduledJobUpdate {
	if s != nil {
		sju.SetDeletedBy(*s)
	}
	return sju
}

// ClearDeletedBy clears the value of the "deleted_by" field.
func (sju *ScheduledJobUpdate) ClearDeletedBy() *ScheduledJobUpdate {
	sju.mutation.ClearDeletedBy()
	return sju
}

// SetTags sets the "tags" field.
func (sju *ScheduledJobUpdate) SetTags(s []string) *ScheduledJobUpdate {
	sju.mutation.SetTags(s)
	return sju
}

// AppendTags appends s to the "tags" field.
func (sju *ScheduledJobUpdate) AppendTags(s []string) *ScheduledJobUpdate {
	sju.mutation.AppendTags(s)
	return sju
}

// ClearTags clears the value of the "tags" field.
func (sju *ScheduledJobUpdate) ClearTags() *ScheduledJobUpdate {
	sju.mutation.ClearTags()
	return sju
}

// SetOwnerID sets the "owner_id" field.
func (sju *ScheduledJobUpdate) SetOwnerID(s string) *ScheduledJobUpdate {
	sju.mutation.SetOwnerID(s)
	return sju
}

// SetNillableOwnerID sets the "owner_id" field if the given value is not nil.
func (sju *ScheduledJobUpdate) SetNillableOwnerID(s *string) *ScheduledJobUpdate {
	if s != nil {
		sju.SetOwnerID(*s)
	}
	return sju
}

// ClearOwnerID clears the value of the "owner_id" field.
func (sju *ScheduledJobUpdate) ClearOwnerID() *ScheduledJobUpdate {
	sju.mutation.ClearOwnerID()
	return sju
}

// SetTitle sets the "title" field.
func (sju *ScheduledJobUpdate) SetTitle(s string) *ScheduledJobUpdate {
	sju.mutation.SetTitle(s)
	return sju
}

// SetNillableTitle sets the "title" field if the given value is not nil.
func (sju *ScheduledJobUpdate) SetNillableTitle(s *string) *ScheduledJobUpdate {
	if s != nil {
		sju.SetTitle(*s)
	}
	return sju
}

// SetDescription sets the "description" field.
func (sju *ScheduledJobUpdate) SetDescription(s string) *ScheduledJobUpdate {
	sju.mutation.SetDescription(s)
	return sju
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (sju *ScheduledJobUpdate) SetNillableDescription(s *string) *ScheduledJobUpdate {
	if s != nil {
		sju.SetDescription(*s)
	}
	return sju
}

// ClearDescription clears the value of the "description" field.
func (sju *ScheduledJobUpdate) ClearDescription() *ScheduledJobUpdate {
	sju.mutation.ClearDescription()
	return sju
}

// SetJobType sets the "job_type" field.
func (sju *ScheduledJobUpdate) SetJobType(et enums.JobType) *ScheduledJobUpdate {
	sju.mutation.SetJobType(et)
	return sju
}

// SetNillableJobType sets the "job_type" field if the given value is not nil.
func (sju *ScheduledJobUpdate) SetNillableJobType(et *enums.JobType) *ScheduledJobUpdate {
	if et != nil {
		sju.SetJobType(*et)
	}
	return sju
}

// SetScript sets the "script" field.
func (sju *ScheduledJobUpdate) SetScript(s string) *ScheduledJobUpdate {
	sju.mutation.SetScript(s)
	return sju
}

// SetNillableScript sets the "script" field if the given value is not nil.
func (sju *ScheduledJobUpdate) SetNillableScript(s *string) *ScheduledJobUpdate {
	if s != nil {
		sju.SetScript(*s)
	}
	return sju
}

// ClearScript clears the value of the "script" field.
func (sju *ScheduledJobUpdate) ClearScript() *ScheduledJobUpdate {
	sju.mutation.ClearScript()
	return sju
}

// SetConfiguration sets the "configuration" field.
func (sju *ScheduledJobUpdate) SetConfiguration(mc models.JobConfiguration) *ScheduledJobUpdate {
	sju.mutation.SetConfiguration(mc)
	return sju
}

// SetNillableConfiguration sets the "configuration" field if the given value is not nil.
func (sju *ScheduledJobUpdate) SetNillableConfiguration(mc *models.JobConfiguration) *ScheduledJobUpdate {
	if mc != nil {
		sju.SetConfiguration(*mc)
	}
	return sju
}

// SetCadence sets the "cadence" field.
func (sju *ScheduledJobUpdate) SetCadence(mc models.JobCadence) *ScheduledJobUpdate {
	sju.mutation.SetCadence(mc)
	return sju
}

// SetNillableCadence sets the "cadence" field if the given value is not nil.
func (sju *ScheduledJobUpdate) SetNillableCadence(mc *models.JobCadence) *ScheduledJobUpdate {
	if mc != nil {
		sju.SetCadence(*mc)
	}
	return sju
}

// ClearCadence clears the value of the "cadence" field.
func (sju *ScheduledJobUpdate) ClearCadence() *ScheduledJobUpdate {
	sju.mutation.ClearCadence()
	return sju
}

// SetCron sets the "cron" field.
func (sju *ScheduledJobUpdate) SetCron(m models.Cron) *ScheduledJobUpdate {
	sju.mutation.SetCron(m)
	return sju
}

// SetNillableCron sets the "cron" field if the given value is not nil.
func (sju *ScheduledJobUpdate) SetNillableCron(m *models.Cron) *ScheduledJobUpdate {
	if m != nil {
		sju.SetCron(*m)
	}
	return sju
}

// ClearCron clears the value of the "cron" field.
func (sju *ScheduledJobUpdate) ClearCron() *ScheduledJobUpdate {
	sju.mutation.ClearCron()
	return sju
}

// SetOwner sets the "owner" edge to the Organization entity.
func (sju *ScheduledJobUpdate) SetOwner(o *Organization) *ScheduledJobUpdate {
	return sju.SetOwnerID(o.ID)
}

// Mutation returns the ScheduledJobMutation object of the builder.
func (sju *ScheduledJobUpdate) Mutation() *ScheduledJobMutation {
	return sju.mutation
}

// ClearOwner clears the "owner" edge to the Organization entity.
func (sju *ScheduledJobUpdate) ClearOwner() *ScheduledJobUpdate {
	sju.mutation.ClearOwner()
	return sju
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (sju *ScheduledJobUpdate) Save(ctx context.Context) (int, error) {
	if err := sju.defaults(); err != nil {
		return 0, err
	}
	return withHooks(ctx, sju.sqlSave, sju.mutation, sju.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (sju *ScheduledJobUpdate) SaveX(ctx context.Context) int {
	affected, err := sju.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (sju *ScheduledJobUpdate) Exec(ctx context.Context) error {
	_, err := sju.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sju *ScheduledJobUpdate) ExecX(ctx context.Context) {
	if err := sju.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (sju *ScheduledJobUpdate) defaults() error {
	if _, ok := sju.mutation.UpdatedAt(); !ok && !sju.mutation.UpdatedAtCleared() {
		if scheduledjob.UpdateDefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized scheduledjob.UpdateDefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := scheduledjob.UpdateDefaultUpdatedAt()
		sju.mutation.SetUpdatedAt(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (sju *ScheduledJobUpdate) check() error {
	if v, ok := sju.mutation.Title(); ok {
		if err := scheduledjob.TitleValidator(v); err != nil {
			return &ValidationError{Name: "title", err: fmt.Errorf(`generated: validator failed for field "ScheduledJob.title": %w`, err)}
		}
	}
	if v, ok := sju.mutation.JobType(); ok {
		if err := scheduledjob.JobTypeValidator(v); err != nil {
			return &ValidationError{Name: "job_type", err: fmt.Errorf(`generated: validator failed for field "ScheduledJob.job_type": %w`, err)}
		}
	}
	if v, ok := sju.mutation.Cadence(); ok {
		if err := v.Validate(); err != nil {
			return &ValidationError{Name: "cadence", err: fmt.Errorf(`generated: validator failed for field "ScheduledJob.cadence": %w`, err)}
		}
	}
	if v, ok := sju.mutation.Cron(); ok {
		if err := v.Validate(); err != nil {
			return &ValidationError{Name: "cron", err: fmt.Errorf(`generated: validator failed for field "ScheduledJob.cron": %w`, err)}
		}
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (sju *ScheduledJobUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *ScheduledJobUpdate {
	sju.modifiers = append(sju.modifiers, modifiers...)
	return sju
}

func (sju *ScheduledJobUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := sju.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(scheduledjob.Table, scheduledjob.Columns, sqlgraph.NewFieldSpec(scheduledjob.FieldID, field.TypeString))
	if ps := sju.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if sju.mutation.CreatedAtCleared() {
		_spec.ClearField(scheduledjob.FieldCreatedAt, field.TypeTime)
	}
	if value, ok := sju.mutation.UpdatedAt(); ok {
		_spec.SetField(scheduledjob.FieldUpdatedAt, field.TypeTime, value)
	}
	if sju.mutation.UpdatedAtCleared() {
		_spec.ClearField(scheduledjob.FieldUpdatedAt, field.TypeTime)
	}
	if sju.mutation.CreatedByCleared() {
		_spec.ClearField(scheduledjob.FieldCreatedBy, field.TypeString)
	}
	if value, ok := sju.mutation.UpdatedBy(); ok {
		_spec.SetField(scheduledjob.FieldUpdatedBy, field.TypeString, value)
	}
	if sju.mutation.UpdatedByCleared() {
		_spec.ClearField(scheduledjob.FieldUpdatedBy, field.TypeString)
	}
	if value, ok := sju.mutation.DeletedAt(); ok {
		_spec.SetField(scheduledjob.FieldDeletedAt, field.TypeTime, value)
	}
	if sju.mutation.DeletedAtCleared() {
		_spec.ClearField(scheduledjob.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := sju.mutation.DeletedBy(); ok {
		_spec.SetField(scheduledjob.FieldDeletedBy, field.TypeString, value)
	}
	if sju.mutation.DeletedByCleared() {
		_spec.ClearField(scheduledjob.FieldDeletedBy, field.TypeString)
	}
	if value, ok := sju.mutation.Tags(); ok {
		_spec.SetField(scheduledjob.FieldTags, field.TypeJSON, value)
	}
	if value, ok := sju.mutation.AppendedTags(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, scheduledjob.FieldTags, value)
		})
	}
	if sju.mutation.TagsCleared() {
		_spec.ClearField(scheduledjob.FieldTags, field.TypeJSON)
	}
	if sju.mutation.SystemOwnedCleared() {
		_spec.ClearField(scheduledjob.FieldSystemOwned, field.TypeBool)
	}
	if value, ok := sju.mutation.Title(); ok {
		_spec.SetField(scheduledjob.FieldTitle, field.TypeString, value)
	}
	if value, ok := sju.mutation.Description(); ok {
		_spec.SetField(scheduledjob.FieldDescription, field.TypeString, value)
	}
	if sju.mutation.DescriptionCleared() {
		_spec.ClearField(scheduledjob.FieldDescription, field.TypeString)
	}
	if value, ok := sju.mutation.JobType(); ok {
		_spec.SetField(scheduledjob.FieldJobType, field.TypeEnum, value)
	}
	if value, ok := sju.mutation.Script(); ok {
		_spec.SetField(scheduledjob.FieldScript, field.TypeString, value)
	}
	if sju.mutation.ScriptCleared() {
		_spec.ClearField(scheduledjob.FieldScript, field.TypeString)
	}
	if value, ok := sju.mutation.Configuration(); ok {
		_spec.SetField(scheduledjob.FieldConfiguration, field.TypeJSON, value)
	}
	if value, ok := sju.mutation.Cadence(); ok {
		_spec.SetField(scheduledjob.FieldCadence, field.TypeJSON, value)
	}
	if sju.mutation.CadenceCleared() {
		_spec.ClearField(scheduledjob.FieldCadence, field.TypeJSON)
	}
	if value, ok := sju.mutation.Cron(); ok {
		_spec.SetField(scheduledjob.FieldCron, field.TypeString, value)
	}
	if sju.mutation.CronCleared() {
		_spec.ClearField(scheduledjob.FieldCron, field.TypeString)
	}
	if sju.mutation.OwnerCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   scheduledjob.OwnerTable,
			Columns: []string{scheduledjob.OwnerColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(organization.FieldID, field.TypeString),
			},
		}
		edge.Schema = sju.schemaConfig.ScheduledJob
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sju.mutation.OwnerIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   scheduledjob.OwnerTable,
			Columns: []string{scheduledjob.OwnerColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(organization.FieldID, field.TypeString),
			},
		}
		edge.Schema = sju.schemaConfig.ScheduledJob
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.Node.Schema = sju.schemaConfig.ScheduledJob
	ctx = internal.NewSchemaConfigContext(ctx, sju.schemaConfig)
	_spec.AddModifiers(sju.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, sju.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{scheduledjob.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	sju.mutation.done = true
	return n, nil
}

// ScheduledJobUpdateOne is the builder for updating a single ScheduledJob entity.
type ScheduledJobUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *ScheduledJobMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetUpdatedAt sets the "updated_at" field.
func (sjuo *ScheduledJobUpdateOne) SetUpdatedAt(t time.Time) *ScheduledJobUpdateOne {
	sjuo.mutation.SetUpdatedAt(t)
	return sjuo
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (sjuo *ScheduledJobUpdateOne) ClearUpdatedAt() *ScheduledJobUpdateOne {
	sjuo.mutation.ClearUpdatedAt()
	return sjuo
}

// SetUpdatedBy sets the "updated_by" field.
func (sjuo *ScheduledJobUpdateOne) SetUpdatedBy(s string) *ScheduledJobUpdateOne {
	sjuo.mutation.SetUpdatedBy(s)
	return sjuo
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (sjuo *ScheduledJobUpdateOne) SetNillableUpdatedBy(s *string) *ScheduledJobUpdateOne {
	if s != nil {
		sjuo.SetUpdatedBy(*s)
	}
	return sjuo
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (sjuo *ScheduledJobUpdateOne) ClearUpdatedBy() *ScheduledJobUpdateOne {
	sjuo.mutation.ClearUpdatedBy()
	return sjuo
}

// SetDeletedAt sets the "deleted_at" field.
func (sjuo *ScheduledJobUpdateOne) SetDeletedAt(t time.Time) *ScheduledJobUpdateOne {
	sjuo.mutation.SetDeletedAt(t)
	return sjuo
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (sjuo *ScheduledJobUpdateOne) SetNillableDeletedAt(t *time.Time) *ScheduledJobUpdateOne {
	if t != nil {
		sjuo.SetDeletedAt(*t)
	}
	return sjuo
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (sjuo *ScheduledJobUpdateOne) ClearDeletedAt() *ScheduledJobUpdateOne {
	sjuo.mutation.ClearDeletedAt()
	return sjuo
}

// SetDeletedBy sets the "deleted_by" field.
func (sjuo *ScheduledJobUpdateOne) SetDeletedBy(s string) *ScheduledJobUpdateOne {
	sjuo.mutation.SetDeletedBy(s)
	return sjuo
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (sjuo *ScheduledJobUpdateOne) SetNillableDeletedBy(s *string) *ScheduledJobUpdateOne {
	if s != nil {
		sjuo.SetDeletedBy(*s)
	}
	return sjuo
}

// ClearDeletedBy clears the value of the "deleted_by" field.
func (sjuo *ScheduledJobUpdateOne) ClearDeletedBy() *ScheduledJobUpdateOne {
	sjuo.mutation.ClearDeletedBy()
	return sjuo
}

// SetTags sets the "tags" field.
func (sjuo *ScheduledJobUpdateOne) SetTags(s []string) *ScheduledJobUpdateOne {
	sjuo.mutation.SetTags(s)
	return sjuo
}

// AppendTags appends s to the "tags" field.
func (sjuo *ScheduledJobUpdateOne) AppendTags(s []string) *ScheduledJobUpdateOne {
	sjuo.mutation.AppendTags(s)
	return sjuo
}

// ClearTags clears the value of the "tags" field.
func (sjuo *ScheduledJobUpdateOne) ClearTags() *ScheduledJobUpdateOne {
	sjuo.mutation.ClearTags()
	return sjuo
}

// SetOwnerID sets the "owner_id" field.
func (sjuo *ScheduledJobUpdateOne) SetOwnerID(s string) *ScheduledJobUpdateOne {
	sjuo.mutation.SetOwnerID(s)
	return sjuo
}

// SetNillableOwnerID sets the "owner_id" field if the given value is not nil.
func (sjuo *ScheduledJobUpdateOne) SetNillableOwnerID(s *string) *ScheduledJobUpdateOne {
	if s != nil {
		sjuo.SetOwnerID(*s)
	}
	return sjuo
}

// ClearOwnerID clears the value of the "owner_id" field.
func (sjuo *ScheduledJobUpdateOne) ClearOwnerID() *ScheduledJobUpdateOne {
	sjuo.mutation.ClearOwnerID()
	return sjuo
}

// SetTitle sets the "title" field.
func (sjuo *ScheduledJobUpdateOne) SetTitle(s string) *ScheduledJobUpdateOne {
	sjuo.mutation.SetTitle(s)
	return sjuo
}

// SetNillableTitle sets the "title" field if the given value is not nil.
func (sjuo *ScheduledJobUpdateOne) SetNillableTitle(s *string) *ScheduledJobUpdateOne {
	if s != nil {
		sjuo.SetTitle(*s)
	}
	return sjuo
}

// SetDescription sets the "description" field.
func (sjuo *ScheduledJobUpdateOne) SetDescription(s string) *ScheduledJobUpdateOne {
	sjuo.mutation.SetDescription(s)
	return sjuo
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (sjuo *ScheduledJobUpdateOne) SetNillableDescription(s *string) *ScheduledJobUpdateOne {
	if s != nil {
		sjuo.SetDescription(*s)
	}
	return sjuo
}

// ClearDescription clears the value of the "description" field.
func (sjuo *ScheduledJobUpdateOne) ClearDescription() *ScheduledJobUpdateOne {
	sjuo.mutation.ClearDescription()
	return sjuo
}

// SetJobType sets the "job_type" field.
func (sjuo *ScheduledJobUpdateOne) SetJobType(et enums.JobType) *ScheduledJobUpdateOne {
	sjuo.mutation.SetJobType(et)
	return sjuo
}

// SetNillableJobType sets the "job_type" field if the given value is not nil.
func (sjuo *ScheduledJobUpdateOne) SetNillableJobType(et *enums.JobType) *ScheduledJobUpdateOne {
	if et != nil {
		sjuo.SetJobType(*et)
	}
	return sjuo
}

// SetScript sets the "script" field.
func (sjuo *ScheduledJobUpdateOne) SetScript(s string) *ScheduledJobUpdateOne {
	sjuo.mutation.SetScript(s)
	return sjuo
}

// SetNillableScript sets the "script" field if the given value is not nil.
func (sjuo *ScheduledJobUpdateOne) SetNillableScript(s *string) *ScheduledJobUpdateOne {
	if s != nil {
		sjuo.SetScript(*s)
	}
	return sjuo
}

// ClearScript clears the value of the "script" field.
func (sjuo *ScheduledJobUpdateOne) ClearScript() *ScheduledJobUpdateOne {
	sjuo.mutation.ClearScript()
	return sjuo
}

// SetConfiguration sets the "configuration" field.
func (sjuo *ScheduledJobUpdateOne) SetConfiguration(mc models.JobConfiguration) *ScheduledJobUpdateOne {
	sjuo.mutation.SetConfiguration(mc)
	return sjuo
}

// SetNillableConfiguration sets the "configuration" field if the given value is not nil.
func (sjuo *ScheduledJobUpdateOne) SetNillableConfiguration(mc *models.JobConfiguration) *ScheduledJobUpdateOne {
	if mc != nil {
		sjuo.SetConfiguration(*mc)
	}
	return sjuo
}

// SetCadence sets the "cadence" field.
func (sjuo *ScheduledJobUpdateOne) SetCadence(mc models.JobCadence) *ScheduledJobUpdateOne {
	sjuo.mutation.SetCadence(mc)
	return sjuo
}

// SetNillableCadence sets the "cadence" field if the given value is not nil.
func (sjuo *ScheduledJobUpdateOne) SetNillableCadence(mc *models.JobCadence) *ScheduledJobUpdateOne {
	if mc != nil {
		sjuo.SetCadence(*mc)
	}
	return sjuo
}

// ClearCadence clears the value of the "cadence" field.
func (sjuo *ScheduledJobUpdateOne) ClearCadence() *ScheduledJobUpdateOne {
	sjuo.mutation.ClearCadence()
	return sjuo
}

// SetCron sets the "cron" field.
func (sjuo *ScheduledJobUpdateOne) SetCron(m models.Cron) *ScheduledJobUpdateOne {
	sjuo.mutation.SetCron(m)
	return sjuo
}

// SetNillableCron sets the "cron" field if the given value is not nil.
func (sjuo *ScheduledJobUpdateOne) SetNillableCron(m *models.Cron) *ScheduledJobUpdateOne {
	if m != nil {
		sjuo.SetCron(*m)
	}
	return sjuo
}

// ClearCron clears the value of the "cron" field.
func (sjuo *ScheduledJobUpdateOne) ClearCron() *ScheduledJobUpdateOne {
	sjuo.mutation.ClearCron()
	return sjuo
}

// SetOwner sets the "owner" edge to the Organization entity.
func (sjuo *ScheduledJobUpdateOne) SetOwner(o *Organization) *ScheduledJobUpdateOne {
	return sjuo.SetOwnerID(o.ID)
}

// Mutation returns the ScheduledJobMutation object of the builder.
func (sjuo *ScheduledJobUpdateOne) Mutation() *ScheduledJobMutation {
	return sjuo.mutation
}

// ClearOwner clears the "owner" edge to the Organization entity.
func (sjuo *ScheduledJobUpdateOne) ClearOwner() *ScheduledJobUpdateOne {
	sjuo.mutation.ClearOwner()
	return sjuo
}

// Where appends a list predicates to the ScheduledJobUpdate builder.
func (sjuo *ScheduledJobUpdateOne) Where(ps ...predicate.ScheduledJob) *ScheduledJobUpdateOne {
	sjuo.mutation.Where(ps...)
	return sjuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (sjuo *ScheduledJobUpdateOne) Select(field string, fields ...string) *ScheduledJobUpdateOne {
	sjuo.fields = append([]string{field}, fields...)
	return sjuo
}

// Save executes the query and returns the updated ScheduledJob entity.
func (sjuo *ScheduledJobUpdateOne) Save(ctx context.Context) (*ScheduledJob, error) {
	if err := sjuo.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, sjuo.sqlSave, sjuo.mutation, sjuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (sjuo *ScheduledJobUpdateOne) SaveX(ctx context.Context) *ScheduledJob {
	node, err := sjuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (sjuo *ScheduledJobUpdateOne) Exec(ctx context.Context) error {
	_, err := sjuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sjuo *ScheduledJobUpdateOne) ExecX(ctx context.Context) {
	if err := sjuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (sjuo *ScheduledJobUpdateOne) defaults() error {
	if _, ok := sjuo.mutation.UpdatedAt(); !ok && !sjuo.mutation.UpdatedAtCleared() {
		if scheduledjob.UpdateDefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized scheduledjob.UpdateDefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := scheduledjob.UpdateDefaultUpdatedAt()
		sjuo.mutation.SetUpdatedAt(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (sjuo *ScheduledJobUpdateOne) check() error {
	if v, ok := sjuo.mutation.Title(); ok {
		if err := scheduledjob.TitleValidator(v); err != nil {
			return &ValidationError{Name: "title", err: fmt.Errorf(`generated: validator failed for field "ScheduledJob.title": %w`, err)}
		}
	}
	if v, ok := sjuo.mutation.JobType(); ok {
		if err := scheduledjob.JobTypeValidator(v); err != nil {
			return &ValidationError{Name: "job_type", err: fmt.Errorf(`generated: validator failed for field "ScheduledJob.job_type": %w`, err)}
		}
	}
	if v, ok := sjuo.mutation.Cadence(); ok {
		if err := v.Validate(); err != nil {
			return &ValidationError{Name: "cadence", err: fmt.Errorf(`generated: validator failed for field "ScheduledJob.cadence": %w`, err)}
		}
	}
	if v, ok := sjuo.mutation.Cron(); ok {
		if err := v.Validate(); err != nil {
			return &ValidationError{Name: "cron", err: fmt.Errorf(`generated: validator failed for field "ScheduledJob.cron": %w`, err)}
		}
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (sjuo *ScheduledJobUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *ScheduledJobUpdateOne {
	sjuo.modifiers = append(sjuo.modifiers, modifiers...)
	return sjuo
}

func (sjuo *ScheduledJobUpdateOne) sqlSave(ctx context.Context) (_node *ScheduledJob, err error) {
	if err := sjuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(scheduledjob.Table, scheduledjob.Columns, sqlgraph.NewFieldSpec(scheduledjob.FieldID, field.TypeString))
	id, ok := sjuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`generated: missing "ScheduledJob.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := sjuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, scheduledjob.FieldID)
		for _, f := range fields {
			if !scheduledjob.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("generated: invalid field %q for query", f)}
			}
			if f != scheduledjob.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := sjuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if sjuo.mutation.CreatedAtCleared() {
		_spec.ClearField(scheduledjob.FieldCreatedAt, field.TypeTime)
	}
	if value, ok := sjuo.mutation.UpdatedAt(); ok {
		_spec.SetField(scheduledjob.FieldUpdatedAt, field.TypeTime, value)
	}
	if sjuo.mutation.UpdatedAtCleared() {
		_spec.ClearField(scheduledjob.FieldUpdatedAt, field.TypeTime)
	}
	if sjuo.mutation.CreatedByCleared() {
		_spec.ClearField(scheduledjob.FieldCreatedBy, field.TypeString)
	}
	if value, ok := sjuo.mutation.UpdatedBy(); ok {
		_spec.SetField(scheduledjob.FieldUpdatedBy, field.TypeString, value)
	}
	if sjuo.mutation.UpdatedByCleared() {
		_spec.ClearField(scheduledjob.FieldUpdatedBy, field.TypeString)
	}
	if value, ok := sjuo.mutation.DeletedAt(); ok {
		_spec.SetField(scheduledjob.FieldDeletedAt, field.TypeTime, value)
	}
	if sjuo.mutation.DeletedAtCleared() {
		_spec.ClearField(scheduledjob.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := sjuo.mutation.DeletedBy(); ok {
		_spec.SetField(scheduledjob.FieldDeletedBy, field.TypeString, value)
	}
	if sjuo.mutation.DeletedByCleared() {
		_spec.ClearField(scheduledjob.FieldDeletedBy, field.TypeString)
	}
	if value, ok := sjuo.mutation.Tags(); ok {
		_spec.SetField(scheduledjob.FieldTags, field.TypeJSON, value)
	}
	if value, ok := sjuo.mutation.AppendedTags(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, scheduledjob.FieldTags, value)
		})
	}
	if sjuo.mutation.TagsCleared() {
		_spec.ClearField(scheduledjob.FieldTags, field.TypeJSON)
	}
	if sjuo.mutation.SystemOwnedCleared() {
		_spec.ClearField(scheduledjob.FieldSystemOwned, field.TypeBool)
	}
	if value, ok := sjuo.mutation.Title(); ok {
		_spec.SetField(scheduledjob.FieldTitle, field.TypeString, value)
	}
	if value, ok := sjuo.mutation.Description(); ok {
		_spec.SetField(scheduledjob.FieldDescription, field.TypeString, value)
	}
	if sjuo.mutation.DescriptionCleared() {
		_spec.ClearField(scheduledjob.FieldDescription, field.TypeString)
	}
	if value, ok := sjuo.mutation.JobType(); ok {
		_spec.SetField(scheduledjob.FieldJobType, field.TypeEnum, value)
	}
	if value, ok := sjuo.mutation.Script(); ok {
		_spec.SetField(scheduledjob.FieldScript, field.TypeString, value)
	}
	if sjuo.mutation.ScriptCleared() {
		_spec.ClearField(scheduledjob.FieldScript, field.TypeString)
	}
	if value, ok := sjuo.mutation.Configuration(); ok {
		_spec.SetField(scheduledjob.FieldConfiguration, field.TypeJSON, value)
	}
	if value, ok := sjuo.mutation.Cadence(); ok {
		_spec.SetField(scheduledjob.FieldCadence, field.TypeJSON, value)
	}
	if sjuo.mutation.CadenceCleared() {
		_spec.ClearField(scheduledjob.FieldCadence, field.TypeJSON)
	}
	if value, ok := sjuo.mutation.Cron(); ok {
		_spec.SetField(scheduledjob.FieldCron, field.TypeString, value)
	}
	if sjuo.mutation.CronCleared() {
		_spec.ClearField(scheduledjob.FieldCron, field.TypeString)
	}
	if sjuo.mutation.OwnerCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   scheduledjob.OwnerTable,
			Columns: []string{scheduledjob.OwnerColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(organization.FieldID, field.TypeString),
			},
		}
		edge.Schema = sjuo.schemaConfig.ScheduledJob
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sjuo.mutation.OwnerIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   scheduledjob.OwnerTable,
			Columns: []string{scheduledjob.OwnerColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(organization.FieldID, field.TypeString),
			},
		}
		edge.Schema = sjuo.schemaConfig.ScheduledJob
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.Node.Schema = sjuo.schemaConfig.ScheduledJob
	ctx = internal.NewSchemaConfigContext(ctx, sjuo.schemaConfig)
	_spec.AddModifiers(sjuo.modifiers...)
	_node = &ScheduledJob{config: sjuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, sjuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{scheduledjob.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	sjuo.mutation.done = true
	return _node, nil
}
