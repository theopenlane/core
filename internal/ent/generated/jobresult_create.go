// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/jobresult"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/scheduledjob"
	"github.com/theopenlane/core/pkg/enums"
)

// JobResultCreate is the builder for creating a JobResult entity.
type JobResultCreate struct {
	config
	mutation *JobResultMutation
	hooks    []Hook
}

// SetCreatedAt sets the "created_at" field.
func (jrc *JobResultCreate) SetCreatedAt(t time.Time) *JobResultCreate {
	jrc.mutation.SetCreatedAt(t)
	return jrc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (jrc *JobResultCreate) SetNillableCreatedAt(t *time.Time) *JobResultCreate {
	if t != nil {
		jrc.SetCreatedAt(*t)
	}
	return jrc
}

// SetUpdatedAt sets the "updated_at" field.
func (jrc *JobResultCreate) SetUpdatedAt(t time.Time) *JobResultCreate {
	jrc.mutation.SetUpdatedAt(t)
	return jrc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (jrc *JobResultCreate) SetNillableUpdatedAt(t *time.Time) *JobResultCreate {
	if t != nil {
		jrc.SetUpdatedAt(*t)
	}
	return jrc
}

// SetCreatedBy sets the "created_by" field.
func (jrc *JobResultCreate) SetCreatedBy(s string) *JobResultCreate {
	jrc.mutation.SetCreatedBy(s)
	return jrc
}

// SetNillableCreatedBy sets the "created_by" field if the given value is not nil.
func (jrc *JobResultCreate) SetNillableCreatedBy(s *string) *JobResultCreate {
	if s != nil {
		jrc.SetCreatedBy(*s)
	}
	return jrc
}

// SetUpdatedBy sets the "updated_by" field.
func (jrc *JobResultCreate) SetUpdatedBy(s string) *JobResultCreate {
	jrc.mutation.SetUpdatedBy(s)
	return jrc
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (jrc *JobResultCreate) SetNillableUpdatedBy(s *string) *JobResultCreate {
	if s != nil {
		jrc.SetUpdatedBy(*s)
	}
	return jrc
}

// SetDeletedAt sets the "deleted_at" field.
func (jrc *JobResultCreate) SetDeletedAt(t time.Time) *JobResultCreate {
	jrc.mutation.SetDeletedAt(t)
	return jrc
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (jrc *JobResultCreate) SetNillableDeletedAt(t *time.Time) *JobResultCreate {
	if t != nil {
		jrc.SetDeletedAt(*t)
	}
	return jrc
}

// SetDeletedBy sets the "deleted_by" field.
func (jrc *JobResultCreate) SetDeletedBy(s string) *JobResultCreate {
	jrc.mutation.SetDeletedBy(s)
	return jrc
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (jrc *JobResultCreate) SetNillableDeletedBy(s *string) *JobResultCreate {
	if s != nil {
		jrc.SetDeletedBy(*s)
	}
	return jrc
}

// SetOwnerID sets the "owner_id" field.
func (jrc *JobResultCreate) SetOwnerID(s string) *JobResultCreate {
	jrc.mutation.SetOwnerID(s)
	return jrc
}

// SetNillableOwnerID sets the "owner_id" field if the given value is not nil.
func (jrc *JobResultCreate) SetNillableOwnerID(s *string) *JobResultCreate {
	if s != nil {
		jrc.SetOwnerID(*s)
	}
	return jrc
}

// SetScheduledJobID sets the "scheduled_job_id" field.
func (jrc *JobResultCreate) SetScheduledJobID(s string) *JobResultCreate {
	jrc.mutation.SetScheduledJobID(s)
	return jrc
}

// SetStatus sets the "status" field.
func (jrc *JobResultCreate) SetStatus(ees enums.JobExecutionStatus) *JobResultCreate {
	jrc.mutation.SetStatus(ees)
	return jrc
}

// SetExitCode sets the "exit_code" field.
func (jrc *JobResultCreate) SetExitCode(i int) *JobResultCreate {
	jrc.mutation.SetExitCode(i)
	return jrc
}

// SetFinishedAt sets the "finished_at" field.
func (jrc *JobResultCreate) SetFinishedAt(t time.Time) *JobResultCreate {
	jrc.mutation.SetFinishedAt(t)
	return jrc
}

// SetNillableFinishedAt sets the "finished_at" field if the given value is not nil.
func (jrc *JobResultCreate) SetNillableFinishedAt(t *time.Time) *JobResultCreate {
	if t != nil {
		jrc.SetFinishedAt(*t)
	}
	return jrc
}

// SetStartedAt sets the "started_at" field.
func (jrc *JobResultCreate) SetStartedAt(t time.Time) *JobResultCreate {
	jrc.mutation.SetStartedAt(t)
	return jrc
}

// SetNillableStartedAt sets the "started_at" field if the given value is not nil.
func (jrc *JobResultCreate) SetNillableStartedAt(t *time.Time) *JobResultCreate {
	if t != nil {
		jrc.SetStartedAt(*t)
	}
	return jrc
}

// SetFileID sets the "file_id" field.
func (jrc *JobResultCreate) SetFileID(s string) *JobResultCreate {
	jrc.mutation.SetFileID(s)
	return jrc
}

// SetID sets the "id" field.
func (jrc *JobResultCreate) SetID(s string) *JobResultCreate {
	jrc.mutation.SetID(s)
	return jrc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (jrc *JobResultCreate) SetNillableID(s *string) *JobResultCreate {
	if s != nil {
		jrc.SetID(*s)
	}
	return jrc
}

// SetOwner sets the "owner" edge to the Organization entity.
func (jrc *JobResultCreate) SetOwner(o *Organization) *JobResultCreate {
	return jrc.SetOwnerID(o.ID)
}

// SetScheduledJob sets the "scheduled_job" edge to the ScheduledJob entity.
func (jrc *JobResultCreate) SetScheduledJob(s *ScheduledJob) *JobResultCreate {
	return jrc.SetScheduledJobID(s.ID)
}

// SetFile sets the "file" edge to the File entity.
func (jrc *JobResultCreate) SetFile(f *File) *JobResultCreate {
	return jrc.SetFileID(f.ID)
}

// Mutation returns the JobResultMutation object of the builder.
func (jrc *JobResultCreate) Mutation() *JobResultMutation {
	return jrc.mutation
}

// Save creates the JobResult in the database.
func (jrc *JobResultCreate) Save(ctx context.Context) (*JobResult, error) {
	if err := jrc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, jrc.sqlSave, jrc.mutation, jrc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (jrc *JobResultCreate) SaveX(ctx context.Context) *JobResult {
	v, err := jrc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (jrc *JobResultCreate) Exec(ctx context.Context) error {
	_, err := jrc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (jrc *JobResultCreate) ExecX(ctx context.Context) {
	if err := jrc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (jrc *JobResultCreate) defaults() error {
	if _, ok := jrc.mutation.CreatedAt(); !ok {
		if jobresult.DefaultCreatedAt == nil {
			return fmt.Errorf("generated: uninitialized jobresult.DefaultCreatedAt (forgotten import generated/runtime?)")
		}
		v := jobresult.DefaultCreatedAt()
		jrc.mutation.SetCreatedAt(v)
	}
	if _, ok := jrc.mutation.UpdatedAt(); !ok {
		if jobresult.DefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized jobresult.DefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := jobresult.DefaultUpdatedAt()
		jrc.mutation.SetUpdatedAt(v)
	}
	if _, ok := jrc.mutation.FinishedAt(); !ok {
		if jobresult.DefaultFinishedAt == nil {
			return fmt.Errorf("generated: uninitialized jobresult.DefaultFinishedAt (forgotten import generated/runtime?)")
		}
		v := jobresult.DefaultFinishedAt()
		jrc.mutation.SetFinishedAt(v)
	}
	if _, ok := jrc.mutation.StartedAt(); !ok {
		if jobresult.DefaultStartedAt == nil {
			return fmt.Errorf("generated: uninitialized jobresult.DefaultStartedAt (forgotten import generated/runtime?)")
		}
		v := jobresult.DefaultStartedAt()
		jrc.mutation.SetStartedAt(v)
	}
	if _, ok := jrc.mutation.ID(); !ok {
		if jobresult.DefaultID == nil {
			return fmt.Errorf("generated: uninitialized jobresult.DefaultID (forgotten import generated/runtime?)")
		}
		v := jobresult.DefaultID()
		jrc.mutation.SetID(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (jrc *JobResultCreate) check() error {
	if v, ok := jrc.mutation.OwnerID(); ok {
		if err := jobresult.OwnerIDValidator(v); err != nil {
			return &ValidationError{Name: "owner_id", err: fmt.Errorf(`generated: validator failed for field "JobResult.owner_id": %w`, err)}
		}
	}
	if _, ok := jrc.mutation.ScheduledJobID(); !ok {
		return &ValidationError{Name: "scheduled_job_id", err: errors.New(`generated: missing required field "JobResult.scheduled_job_id"`)}
	}
	if _, ok := jrc.mutation.Status(); !ok {
		return &ValidationError{Name: "status", err: errors.New(`generated: missing required field "JobResult.status"`)}
	}
	if v, ok := jrc.mutation.Status(); ok {
		if err := jobresult.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`generated: validator failed for field "JobResult.status": %w`, err)}
		}
	}
	if _, ok := jrc.mutation.ExitCode(); !ok {
		return &ValidationError{Name: "exit_code", err: errors.New(`generated: missing required field "JobResult.exit_code"`)}
	}
	if v, ok := jrc.mutation.ExitCode(); ok {
		if err := jobresult.ExitCodeValidator(v); err != nil {
			return &ValidationError{Name: "exit_code", err: fmt.Errorf(`generated: validator failed for field "JobResult.exit_code": %w`, err)}
		}
	}
	if _, ok := jrc.mutation.FinishedAt(); !ok {
		return &ValidationError{Name: "finished_at", err: errors.New(`generated: missing required field "JobResult.finished_at"`)}
	}
	if _, ok := jrc.mutation.StartedAt(); !ok {
		return &ValidationError{Name: "started_at", err: errors.New(`generated: missing required field "JobResult.started_at"`)}
	}
	if _, ok := jrc.mutation.FileID(); !ok {
		return &ValidationError{Name: "file_id", err: errors.New(`generated: missing required field "JobResult.file_id"`)}
	}
	if len(jrc.mutation.ScheduledJobIDs()) == 0 {
		return &ValidationError{Name: "scheduled_job", err: errors.New(`generated: missing required edge "JobResult.scheduled_job"`)}
	}
	if len(jrc.mutation.FileIDs()) == 0 {
		return &ValidationError{Name: "file", err: errors.New(`generated: missing required edge "JobResult.file"`)}
	}
	return nil
}

func (jrc *JobResultCreate) sqlSave(ctx context.Context) (*JobResult, error) {
	if err := jrc.check(); err != nil {
		return nil, err
	}
	_node, _spec := jrc.createSpec()
	if err := sqlgraph.CreateNode(ctx, jrc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected JobResult.ID type: %T", _spec.ID.Value)
		}
	}
	jrc.mutation.id = &_node.ID
	jrc.mutation.done = true
	return _node, nil
}

func (jrc *JobResultCreate) createSpec() (*JobResult, *sqlgraph.CreateSpec) {
	var (
		_node = &JobResult{config: jrc.config}
		_spec = sqlgraph.NewCreateSpec(jobresult.Table, sqlgraph.NewFieldSpec(jobresult.FieldID, field.TypeString))
	)
	_spec.Schema = jrc.schemaConfig.JobResult
	if id, ok := jrc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := jrc.mutation.CreatedAt(); ok {
		_spec.SetField(jobresult.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := jrc.mutation.UpdatedAt(); ok {
		_spec.SetField(jobresult.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := jrc.mutation.CreatedBy(); ok {
		_spec.SetField(jobresult.FieldCreatedBy, field.TypeString, value)
		_node.CreatedBy = value
	}
	if value, ok := jrc.mutation.UpdatedBy(); ok {
		_spec.SetField(jobresult.FieldUpdatedBy, field.TypeString, value)
		_node.UpdatedBy = value
	}
	if value, ok := jrc.mutation.DeletedAt(); ok {
		_spec.SetField(jobresult.FieldDeletedAt, field.TypeTime, value)
		_node.DeletedAt = value
	}
	if value, ok := jrc.mutation.DeletedBy(); ok {
		_spec.SetField(jobresult.FieldDeletedBy, field.TypeString, value)
		_node.DeletedBy = value
	}
	if value, ok := jrc.mutation.Status(); ok {
		_spec.SetField(jobresult.FieldStatus, field.TypeEnum, value)
		_node.Status = value
	}
	if value, ok := jrc.mutation.ExitCode(); ok {
		_spec.SetField(jobresult.FieldExitCode, field.TypeInt, value)
		_node.ExitCode = &value
	}
	if value, ok := jrc.mutation.FinishedAt(); ok {
		_spec.SetField(jobresult.FieldFinishedAt, field.TypeTime, value)
		_node.FinishedAt = value
	}
	if value, ok := jrc.mutation.StartedAt(); ok {
		_spec.SetField(jobresult.FieldStartedAt, field.TypeTime, value)
		_node.StartedAt = value
	}
	if nodes := jrc.mutation.OwnerIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   jobresult.OwnerTable,
			Columns: []string{jobresult.OwnerColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(organization.FieldID, field.TypeString),
			},
		}
		edge.Schema = jrc.schemaConfig.JobResult
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.OwnerID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := jrc.mutation.ScheduledJobIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   jobresult.ScheduledJobTable,
			Columns: []string{jobresult.ScheduledJobColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(scheduledjob.FieldID, field.TypeString),
			},
		}
		edge.Schema = jrc.schemaConfig.JobResult
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.ScheduledJobID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := jrc.mutation.FileIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   jobresult.FileTable,
			Columns: []string{jobresult.FileColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(file.FieldID, field.TypeString),
			},
		}
		edge.Schema = jrc.schemaConfig.JobResult
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.FileID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// JobResultCreateBulk is the builder for creating many JobResult entities in bulk.
type JobResultCreateBulk struct {
	config
	err      error
	builders []*JobResultCreate
}

// Save creates the JobResult entities in the database.
func (jrcb *JobResultCreateBulk) Save(ctx context.Context) ([]*JobResult, error) {
	if jrcb.err != nil {
		return nil, jrcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(jrcb.builders))
	nodes := make([]*JobResult, len(jrcb.builders))
	mutators := make([]Mutator, len(jrcb.builders))
	for i := range jrcb.builders {
		func(i int, root context.Context) {
			builder := jrcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*JobResultMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, jrcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, jrcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, jrcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (jrcb *JobResultCreateBulk) SaveX(ctx context.Context) []*JobResult {
	v, err := jrcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (jrcb *JobResultCreateBulk) Exec(ctx context.Context) error {
	_, err := jrcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (jrcb *JobResultCreateBulk) ExecX(ctx context.Context) {
	if err := jrcb.Exec(ctx); err != nil {
		panic(err)
	}
}
