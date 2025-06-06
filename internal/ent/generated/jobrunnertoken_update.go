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
	"github.com/theopenlane/core/internal/ent/generated/jobrunner"
	"github.com/theopenlane/core/internal/ent/generated/jobrunnertoken"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/predicate"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// JobRunnerTokenUpdate is the builder for updating JobRunnerToken entities.
type JobRunnerTokenUpdate struct {
	config
	hooks     []Hook
	mutation  *JobRunnerTokenMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the JobRunnerTokenUpdate builder.
func (jrtu *JobRunnerTokenUpdate) Where(ps ...predicate.JobRunnerToken) *JobRunnerTokenUpdate {
	jrtu.mutation.Where(ps...)
	return jrtu
}

// SetUpdatedAt sets the "updated_at" field.
func (jrtu *JobRunnerTokenUpdate) SetUpdatedAt(t time.Time) *JobRunnerTokenUpdate {
	jrtu.mutation.SetUpdatedAt(t)
	return jrtu
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (jrtu *JobRunnerTokenUpdate) ClearUpdatedAt() *JobRunnerTokenUpdate {
	jrtu.mutation.ClearUpdatedAt()
	return jrtu
}

// SetUpdatedBy sets the "updated_by" field.
func (jrtu *JobRunnerTokenUpdate) SetUpdatedBy(s string) *JobRunnerTokenUpdate {
	jrtu.mutation.SetUpdatedBy(s)
	return jrtu
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (jrtu *JobRunnerTokenUpdate) SetNillableUpdatedBy(s *string) *JobRunnerTokenUpdate {
	if s != nil {
		jrtu.SetUpdatedBy(*s)
	}
	return jrtu
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (jrtu *JobRunnerTokenUpdate) ClearUpdatedBy() *JobRunnerTokenUpdate {
	jrtu.mutation.ClearUpdatedBy()
	return jrtu
}

// SetDeletedAt sets the "deleted_at" field.
func (jrtu *JobRunnerTokenUpdate) SetDeletedAt(t time.Time) *JobRunnerTokenUpdate {
	jrtu.mutation.SetDeletedAt(t)
	return jrtu
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (jrtu *JobRunnerTokenUpdate) SetNillableDeletedAt(t *time.Time) *JobRunnerTokenUpdate {
	if t != nil {
		jrtu.SetDeletedAt(*t)
	}
	return jrtu
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (jrtu *JobRunnerTokenUpdate) ClearDeletedAt() *JobRunnerTokenUpdate {
	jrtu.mutation.ClearDeletedAt()
	return jrtu
}

// SetDeletedBy sets the "deleted_by" field.
func (jrtu *JobRunnerTokenUpdate) SetDeletedBy(s string) *JobRunnerTokenUpdate {
	jrtu.mutation.SetDeletedBy(s)
	return jrtu
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (jrtu *JobRunnerTokenUpdate) SetNillableDeletedBy(s *string) *JobRunnerTokenUpdate {
	if s != nil {
		jrtu.SetDeletedBy(*s)
	}
	return jrtu
}

// ClearDeletedBy clears the value of the "deleted_by" field.
func (jrtu *JobRunnerTokenUpdate) ClearDeletedBy() *JobRunnerTokenUpdate {
	jrtu.mutation.ClearDeletedBy()
	return jrtu
}

// SetTags sets the "tags" field.
func (jrtu *JobRunnerTokenUpdate) SetTags(s []string) *JobRunnerTokenUpdate {
	jrtu.mutation.SetTags(s)
	return jrtu
}

// AppendTags appends s to the "tags" field.
func (jrtu *JobRunnerTokenUpdate) AppendTags(s []string) *JobRunnerTokenUpdate {
	jrtu.mutation.AppendTags(s)
	return jrtu
}

// ClearTags clears the value of the "tags" field.
func (jrtu *JobRunnerTokenUpdate) ClearTags() *JobRunnerTokenUpdate {
	jrtu.mutation.ClearTags()
	return jrtu
}

// SetOwnerID sets the "owner_id" field.
func (jrtu *JobRunnerTokenUpdate) SetOwnerID(s string) *JobRunnerTokenUpdate {
	jrtu.mutation.SetOwnerID(s)
	return jrtu
}

// SetNillableOwnerID sets the "owner_id" field if the given value is not nil.
func (jrtu *JobRunnerTokenUpdate) SetNillableOwnerID(s *string) *JobRunnerTokenUpdate {
	if s != nil {
		jrtu.SetOwnerID(*s)
	}
	return jrtu
}

// ClearOwnerID clears the value of the "owner_id" field.
func (jrtu *JobRunnerTokenUpdate) ClearOwnerID() *JobRunnerTokenUpdate {
	jrtu.mutation.ClearOwnerID()
	return jrtu
}

// SetExpiresAt sets the "expires_at" field.
func (jrtu *JobRunnerTokenUpdate) SetExpiresAt(t time.Time) *JobRunnerTokenUpdate {
	jrtu.mutation.SetExpiresAt(t)
	return jrtu
}

// SetNillableExpiresAt sets the "expires_at" field if the given value is not nil.
func (jrtu *JobRunnerTokenUpdate) SetNillableExpiresAt(t *time.Time) *JobRunnerTokenUpdate {
	if t != nil {
		jrtu.SetExpiresAt(*t)
	}
	return jrtu
}

// ClearExpiresAt clears the value of the "expires_at" field.
func (jrtu *JobRunnerTokenUpdate) ClearExpiresAt() *JobRunnerTokenUpdate {
	jrtu.mutation.ClearExpiresAt()
	return jrtu
}

// SetLastUsedAt sets the "last_used_at" field.
func (jrtu *JobRunnerTokenUpdate) SetLastUsedAt(t time.Time) *JobRunnerTokenUpdate {
	jrtu.mutation.SetLastUsedAt(t)
	return jrtu
}

// SetNillableLastUsedAt sets the "last_used_at" field if the given value is not nil.
func (jrtu *JobRunnerTokenUpdate) SetNillableLastUsedAt(t *time.Time) *JobRunnerTokenUpdate {
	if t != nil {
		jrtu.SetLastUsedAt(*t)
	}
	return jrtu
}

// ClearLastUsedAt clears the value of the "last_used_at" field.
func (jrtu *JobRunnerTokenUpdate) ClearLastUsedAt() *JobRunnerTokenUpdate {
	jrtu.mutation.ClearLastUsedAt()
	return jrtu
}

// SetIsActive sets the "is_active" field.
func (jrtu *JobRunnerTokenUpdate) SetIsActive(b bool) *JobRunnerTokenUpdate {
	jrtu.mutation.SetIsActive(b)
	return jrtu
}

// SetNillableIsActive sets the "is_active" field if the given value is not nil.
func (jrtu *JobRunnerTokenUpdate) SetNillableIsActive(b *bool) *JobRunnerTokenUpdate {
	if b != nil {
		jrtu.SetIsActive(*b)
	}
	return jrtu
}

// ClearIsActive clears the value of the "is_active" field.
func (jrtu *JobRunnerTokenUpdate) ClearIsActive() *JobRunnerTokenUpdate {
	jrtu.mutation.ClearIsActive()
	return jrtu
}

// SetRevokedReason sets the "revoked_reason" field.
func (jrtu *JobRunnerTokenUpdate) SetRevokedReason(s string) *JobRunnerTokenUpdate {
	jrtu.mutation.SetRevokedReason(s)
	return jrtu
}

// SetNillableRevokedReason sets the "revoked_reason" field if the given value is not nil.
func (jrtu *JobRunnerTokenUpdate) SetNillableRevokedReason(s *string) *JobRunnerTokenUpdate {
	if s != nil {
		jrtu.SetRevokedReason(*s)
	}
	return jrtu
}

// ClearRevokedReason clears the value of the "revoked_reason" field.
func (jrtu *JobRunnerTokenUpdate) ClearRevokedReason() *JobRunnerTokenUpdate {
	jrtu.mutation.ClearRevokedReason()
	return jrtu
}

// SetRevokedBy sets the "revoked_by" field.
func (jrtu *JobRunnerTokenUpdate) SetRevokedBy(s string) *JobRunnerTokenUpdate {
	jrtu.mutation.SetRevokedBy(s)
	return jrtu
}

// SetNillableRevokedBy sets the "revoked_by" field if the given value is not nil.
func (jrtu *JobRunnerTokenUpdate) SetNillableRevokedBy(s *string) *JobRunnerTokenUpdate {
	if s != nil {
		jrtu.SetRevokedBy(*s)
	}
	return jrtu
}

// ClearRevokedBy clears the value of the "revoked_by" field.
func (jrtu *JobRunnerTokenUpdate) ClearRevokedBy() *JobRunnerTokenUpdate {
	jrtu.mutation.ClearRevokedBy()
	return jrtu
}

// SetRevokedAt sets the "revoked_at" field.
func (jrtu *JobRunnerTokenUpdate) SetRevokedAt(t time.Time) *JobRunnerTokenUpdate {
	jrtu.mutation.SetRevokedAt(t)
	return jrtu
}

// SetNillableRevokedAt sets the "revoked_at" field if the given value is not nil.
func (jrtu *JobRunnerTokenUpdate) SetNillableRevokedAt(t *time.Time) *JobRunnerTokenUpdate {
	if t != nil {
		jrtu.SetRevokedAt(*t)
	}
	return jrtu
}

// ClearRevokedAt clears the value of the "revoked_at" field.
func (jrtu *JobRunnerTokenUpdate) ClearRevokedAt() *JobRunnerTokenUpdate {
	jrtu.mutation.ClearRevokedAt()
	return jrtu
}

// SetOwner sets the "owner" edge to the Organization entity.
func (jrtu *JobRunnerTokenUpdate) SetOwner(o *Organization) *JobRunnerTokenUpdate {
	return jrtu.SetOwnerID(o.ID)
}

// AddJobRunnerIDs adds the "job_runners" edge to the JobRunner entity by IDs.
func (jrtu *JobRunnerTokenUpdate) AddJobRunnerIDs(ids ...string) *JobRunnerTokenUpdate {
	jrtu.mutation.AddJobRunnerIDs(ids...)
	return jrtu
}

// AddJobRunners adds the "job_runners" edges to the JobRunner entity.
func (jrtu *JobRunnerTokenUpdate) AddJobRunners(j ...*JobRunner) *JobRunnerTokenUpdate {
	ids := make([]string, len(j))
	for i := range j {
		ids[i] = j[i].ID
	}
	return jrtu.AddJobRunnerIDs(ids...)
}

// Mutation returns the JobRunnerTokenMutation object of the builder.
func (jrtu *JobRunnerTokenUpdate) Mutation() *JobRunnerTokenMutation {
	return jrtu.mutation
}

// ClearOwner clears the "owner" edge to the Organization entity.
func (jrtu *JobRunnerTokenUpdate) ClearOwner() *JobRunnerTokenUpdate {
	jrtu.mutation.ClearOwner()
	return jrtu
}

// ClearJobRunners clears all "job_runners" edges to the JobRunner entity.
func (jrtu *JobRunnerTokenUpdate) ClearJobRunners() *JobRunnerTokenUpdate {
	jrtu.mutation.ClearJobRunners()
	return jrtu
}

// RemoveJobRunnerIDs removes the "job_runners" edge to JobRunner entities by IDs.
func (jrtu *JobRunnerTokenUpdate) RemoveJobRunnerIDs(ids ...string) *JobRunnerTokenUpdate {
	jrtu.mutation.RemoveJobRunnerIDs(ids...)
	return jrtu
}

// RemoveJobRunners removes "job_runners" edges to JobRunner entities.
func (jrtu *JobRunnerTokenUpdate) RemoveJobRunners(j ...*JobRunner) *JobRunnerTokenUpdate {
	ids := make([]string, len(j))
	for i := range j {
		ids[i] = j[i].ID
	}
	return jrtu.RemoveJobRunnerIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (jrtu *JobRunnerTokenUpdate) Save(ctx context.Context) (int, error) {
	if err := jrtu.defaults(); err != nil {
		return 0, err
	}
	return withHooks(ctx, jrtu.sqlSave, jrtu.mutation, jrtu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (jrtu *JobRunnerTokenUpdate) SaveX(ctx context.Context) int {
	affected, err := jrtu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (jrtu *JobRunnerTokenUpdate) Exec(ctx context.Context) error {
	_, err := jrtu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (jrtu *JobRunnerTokenUpdate) ExecX(ctx context.Context) {
	if err := jrtu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (jrtu *JobRunnerTokenUpdate) defaults() error {
	if _, ok := jrtu.mutation.UpdatedAt(); !ok && !jrtu.mutation.UpdatedAtCleared() {
		if jobrunnertoken.UpdateDefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized jobrunnertoken.UpdateDefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := jobrunnertoken.UpdateDefaultUpdatedAt()
		jrtu.mutation.SetUpdatedAt(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (jrtu *JobRunnerTokenUpdate) check() error {
	if v, ok := jrtu.mutation.OwnerID(); ok {
		if err := jobrunnertoken.OwnerIDValidator(v); err != nil {
			return &ValidationError{Name: "owner_id", err: fmt.Errorf(`generated: validator failed for field "JobRunnerToken.owner_id": %w`, err)}
		}
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (jrtu *JobRunnerTokenUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *JobRunnerTokenUpdate {
	jrtu.modifiers = append(jrtu.modifiers, modifiers...)
	return jrtu
}

func (jrtu *JobRunnerTokenUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := jrtu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(jobrunnertoken.Table, jobrunnertoken.Columns, sqlgraph.NewFieldSpec(jobrunnertoken.FieldID, field.TypeString))
	if ps := jrtu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if jrtu.mutation.CreatedAtCleared() {
		_spec.ClearField(jobrunnertoken.FieldCreatedAt, field.TypeTime)
	}
	if value, ok := jrtu.mutation.UpdatedAt(); ok {
		_spec.SetField(jobrunnertoken.FieldUpdatedAt, field.TypeTime, value)
	}
	if jrtu.mutation.UpdatedAtCleared() {
		_spec.ClearField(jobrunnertoken.FieldUpdatedAt, field.TypeTime)
	}
	if jrtu.mutation.CreatedByCleared() {
		_spec.ClearField(jobrunnertoken.FieldCreatedBy, field.TypeString)
	}
	if value, ok := jrtu.mutation.UpdatedBy(); ok {
		_spec.SetField(jobrunnertoken.FieldUpdatedBy, field.TypeString, value)
	}
	if jrtu.mutation.UpdatedByCleared() {
		_spec.ClearField(jobrunnertoken.FieldUpdatedBy, field.TypeString)
	}
	if value, ok := jrtu.mutation.DeletedAt(); ok {
		_spec.SetField(jobrunnertoken.FieldDeletedAt, field.TypeTime, value)
	}
	if jrtu.mutation.DeletedAtCleared() {
		_spec.ClearField(jobrunnertoken.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := jrtu.mutation.DeletedBy(); ok {
		_spec.SetField(jobrunnertoken.FieldDeletedBy, field.TypeString, value)
	}
	if jrtu.mutation.DeletedByCleared() {
		_spec.ClearField(jobrunnertoken.FieldDeletedBy, field.TypeString)
	}
	if value, ok := jrtu.mutation.Tags(); ok {
		_spec.SetField(jobrunnertoken.FieldTags, field.TypeJSON, value)
	}
	if value, ok := jrtu.mutation.AppendedTags(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, jobrunnertoken.FieldTags, value)
		})
	}
	if jrtu.mutation.TagsCleared() {
		_spec.ClearField(jobrunnertoken.FieldTags, field.TypeJSON)
	}
	if value, ok := jrtu.mutation.ExpiresAt(); ok {
		_spec.SetField(jobrunnertoken.FieldExpiresAt, field.TypeTime, value)
	}
	if jrtu.mutation.ExpiresAtCleared() {
		_spec.ClearField(jobrunnertoken.FieldExpiresAt, field.TypeTime)
	}
	if value, ok := jrtu.mutation.LastUsedAt(); ok {
		_spec.SetField(jobrunnertoken.FieldLastUsedAt, field.TypeTime, value)
	}
	if jrtu.mutation.LastUsedAtCleared() {
		_spec.ClearField(jobrunnertoken.FieldLastUsedAt, field.TypeTime)
	}
	if value, ok := jrtu.mutation.IsActive(); ok {
		_spec.SetField(jobrunnertoken.FieldIsActive, field.TypeBool, value)
	}
	if jrtu.mutation.IsActiveCleared() {
		_spec.ClearField(jobrunnertoken.FieldIsActive, field.TypeBool)
	}
	if value, ok := jrtu.mutation.RevokedReason(); ok {
		_spec.SetField(jobrunnertoken.FieldRevokedReason, field.TypeString, value)
	}
	if jrtu.mutation.RevokedReasonCleared() {
		_spec.ClearField(jobrunnertoken.FieldRevokedReason, field.TypeString)
	}
	if value, ok := jrtu.mutation.RevokedBy(); ok {
		_spec.SetField(jobrunnertoken.FieldRevokedBy, field.TypeString, value)
	}
	if jrtu.mutation.RevokedByCleared() {
		_spec.ClearField(jobrunnertoken.FieldRevokedBy, field.TypeString)
	}
	if value, ok := jrtu.mutation.RevokedAt(); ok {
		_spec.SetField(jobrunnertoken.FieldRevokedAt, field.TypeTime, value)
	}
	if jrtu.mutation.RevokedAtCleared() {
		_spec.ClearField(jobrunnertoken.FieldRevokedAt, field.TypeTime)
	}
	if jrtu.mutation.OwnerCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   jobrunnertoken.OwnerTable,
			Columns: []string{jobrunnertoken.OwnerColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(organization.FieldID, field.TypeString),
			},
		}
		edge.Schema = jrtu.schemaConfig.JobRunnerToken
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := jrtu.mutation.OwnerIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   jobrunnertoken.OwnerTable,
			Columns: []string{jobrunnertoken.OwnerColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(organization.FieldID, field.TypeString),
			},
		}
		edge.Schema = jrtu.schemaConfig.JobRunnerToken
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if jrtu.mutation.JobRunnersCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   jobrunnertoken.JobRunnersTable,
			Columns: jobrunnertoken.JobRunnersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(jobrunner.FieldID, field.TypeString),
			},
		}
		edge.Schema = jrtu.schemaConfig.JobRunnerJobRunnerTokens
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := jrtu.mutation.RemovedJobRunnersIDs(); len(nodes) > 0 && !jrtu.mutation.JobRunnersCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   jobrunnertoken.JobRunnersTable,
			Columns: jobrunnertoken.JobRunnersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(jobrunner.FieldID, field.TypeString),
			},
		}
		edge.Schema = jrtu.schemaConfig.JobRunnerJobRunnerTokens
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := jrtu.mutation.JobRunnersIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   jobrunnertoken.JobRunnersTable,
			Columns: jobrunnertoken.JobRunnersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(jobrunner.FieldID, field.TypeString),
			},
		}
		edge.Schema = jrtu.schemaConfig.JobRunnerJobRunnerTokens
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.Node.Schema = jrtu.schemaConfig.JobRunnerToken
	ctx = internal.NewSchemaConfigContext(ctx, jrtu.schemaConfig)
	_spec.AddModifiers(jrtu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, jrtu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{jobrunnertoken.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	jrtu.mutation.done = true
	return n, nil
}

// JobRunnerTokenUpdateOne is the builder for updating a single JobRunnerToken entity.
type JobRunnerTokenUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *JobRunnerTokenMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetUpdatedAt sets the "updated_at" field.
func (jrtuo *JobRunnerTokenUpdateOne) SetUpdatedAt(t time.Time) *JobRunnerTokenUpdateOne {
	jrtuo.mutation.SetUpdatedAt(t)
	return jrtuo
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (jrtuo *JobRunnerTokenUpdateOne) ClearUpdatedAt() *JobRunnerTokenUpdateOne {
	jrtuo.mutation.ClearUpdatedAt()
	return jrtuo
}

// SetUpdatedBy sets the "updated_by" field.
func (jrtuo *JobRunnerTokenUpdateOne) SetUpdatedBy(s string) *JobRunnerTokenUpdateOne {
	jrtuo.mutation.SetUpdatedBy(s)
	return jrtuo
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (jrtuo *JobRunnerTokenUpdateOne) SetNillableUpdatedBy(s *string) *JobRunnerTokenUpdateOne {
	if s != nil {
		jrtuo.SetUpdatedBy(*s)
	}
	return jrtuo
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (jrtuo *JobRunnerTokenUpdateOne) ClearUpdatedBy() *JobRunnerTokenUpdateOne {
	jrtuo.mutation.ClearUpdatedBy()
	return jrtuo
}

// SetDeletedAt sets the "deleted_at" field.
func (jrtuo *JobRunnerTokenUpdateOne) SetDeletedAt(t time.Time) *JobRunnerTokenUpdateOne {
	jrtuo.mutation.SetDeletedAt(t)
	return jrtuo
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (jrtuo *JobRunnerTokenUpdateOne) SetNillableDeletedAt(t *time.Time) *JobRunnerTokenUpdateOne {
	if t != nil {
		jrtuo.SetDeletedAt(*t)
	}
	return jrtuo
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (jrtuo *JobRunnerTokenUpdateOne) ClearDeletedAt() *JobRunnerTokenUpdateOne {
	jrtuo.mutation.ClearDeletedAt()
	return jrtuo
}

// SetDeletedBy sets the "deleted_by" field.
func (jrtuo *JobRunnerTokenUpdateOne) SetDeletedBy(s string) *JobRunnerTokenUpdateOne {
	jrtuo.mutation.SetDeletedBy(s)
	return jrtuo
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (jrtuo *JobRunnerTokenUpdateOne) SetNillableDeletedBy(s *string) *JobRunnerTokenUpdateOne {
	if s != nil {
		jrtuo.SetDeletedBy(*s)
	}
	return jrtuo
}

// ClearDeletedBy clears the value of the "deleted_by" field.
func (jrtuo *JobRunnerTokenUpdateOne) ClearDeletedBy() *JobRunnerTokenUpdateOne {
	jrtuo.mutation.ClearDeletedBy()
	return jrtuo
}

// SetTags sets the "tags" field.
func (jrtuo *JobRunnerTokenUpdateOne) SetTags(s []string) *JobRunnerTokenUpdateOne {
	jrtuo.mutation.SetTags(s)
	return jrtuo
}

// AppendTags appends s to the "tags" field.
func (jrtuo *JobRunnerTokenUpdateOne) AppendTags(s []string) *JobRunnerTokenUpdateOne {
	jrtuo.mutation.AppendTags(s)
	return jrtuo
}

// ClearTags clears the value of the "tags" field.
func (jrtuo *JobRunnerTokenUpdateOne) ClearTags() *JobRunnerTokenUpdateOne {
	jrtuo.mutation.ClearTags()
	return jrtuo
}

// SetOwnerID sets the "owner_id" field.
func (jrtuo *JobRunnerTokenUpdateOne) SetOwnerID(s string) *JobRunnerTokenUpdateOne {
	jrtuo.mutation.SetOwnerID(s)
	return jrtuo
}

// SetNillableOwnerID sets the "owner_id" field if the given value is not nil.
func (jrtuo *JobRunnerTokenUpdateOne) SetNillableOwnerID(s *string) *JobRunnerTokenUpdateOne {
	if s != nil {
		jrtuo.SetOwnerID(*s)
	}
	return jrtuo
}

// ClearOwnerID clears the value of the "owner_id" field.
func (jrtuo *JobRunnerTokenUpdateOne) ClearOwnerID() *JobRunnerTokenUpdateOne {
	jrtuo.mutation.ClearOwnerID()
	return jrtuo
}

// SetExpiresAt sets the "expires_at" field.
func (jrtuo *JobRunnerTokenUpdateOne) SetExpiresAt(t time.Time) *JobRunnerTokenUpdateOne {
	jrtuo.mutation.SetExpiresAt(t)
	return jrtuo
}

// SetNillableExpiresAt sets the "expires_at" field if the given value is not nil.
func (jrtuo *JobRunnerTokenUpdateOne) SetNillableExpiresAt(t *time.Time) *JobRunnerTokenUpdateOne {
	if t != nil {
		jrtuo.SetExpiresAt(*t)
	}
	return jrtuo
}

// ClearExpiresAt clears the value of the "expires_at" field.
func (jrtuo *JobRunnerTokenUpdateOne) ClearExpiresAt() *JobRunnerTokenUpdateOne {
	jrtuo.mutation.ClearExpiresAt()
	return jrtuo
}

// SetLastUsedAt sets the "last_used_at" field.
func (jrtuo *JobRunnerTokenUpdateOne) SetLastUsedAt(t time.Time) *JobRunnerTokenUpdateOne {
	jrtuo.mutation.SetLastUsedAt(t)
	return jrtuo
}

// SetNillableLastUsedAt sets the "last_used_at" field if the given value is not nil.
func (jrtuo *JobRunnerTokenUpdateOne) SetNillableLastUsedAt(t *time.Time) *JobRunnerTokenUpdateOne {
	if t != nil {
		jrtuo.SetLastUsedAt(*t)
	}
	return jrtuo
}

// ClearLastUsedAt clears the value of the "last_used_at" field.
func (jrtuo *JobRunnerTokenUpdateOne) ClearLastUsedAt() *JobRunnerTokenUpdateOne {
	jrtuo.mutation.ClearLastUsedAt()
	return jrtuo
}

// SetIsActive sets the "is_active" field.
func (jrtuo *JobRunnerTokenUpdateOne) SetIsActive(b bool) *JobRunnerTokenUpdateOne {
	jrtuo.mutation.SetIsActive(b)
	return jrtuo
}

// SetNillableIsActive sets the "is_active" field if the given value is not nil.
func (jrtuo *JobRunnerTokenUpdateOne) SetNillableIsActive(b *bool) *JobRunnerTokenUpdateOne {
	if b != nil {
		jrtuo.SetIsActive(*b)
	}
	return jrtuo
}

// ClearIsActive clears the value of the "is_active" field.
func (jrtuo *JobRunnerTokenUpdateOne) ClearIsActive() *JobRunnerTokenUpdateOne {
	jrtuo.mutation.ClearIsActive()
	return jrtuo
}

// SetRevokedReason sets the "revoked_reason" field.
func (jrtuo *JobRunnerTokenUpdateOne) SetRevokedReason(s string) *JobRunnerTokenUpdateOne {
	jrtuo.mutation.SetRevokedReason(s)
	return jrtuo
}

// SetNillableRevokedReason sets the "revoked_reason" field if the given value is not nil.
func (jrtuo *JobRunnerTokenUpdateOne) SetNillableRevokedReason(s *string) *JobRunnerTokenUpdateOne {
	if s != nil {
		jrtuo.SetRevokedReason(*s)
	}
	return jrtuo
}

// ClearRevokedReason clears the value of the "revoked_reason" field.
func (jrtuo *JobRunnerTokenUpdateOne) ClearRevokedReason() *JobRunnerTokenUpdateOne {
	jrtuo.mutation.ClearRevokedReason()
	return jrtuo
}

// SetRevokedBy sets the "revoked_by" field.
func (jrtuo *JobRunnerTokenUpdateOne) SetRevokedBy(s string) *JobRunnerTokenUpdateOne {
	jrtuo.mutation.SetRevokedBy(s)
	return jrtuo
}

// SetNillableRevokedBy sets the "revoked_by" field if the given value is not nil.
func (jrtuo *JobRunnerTokenUpdateOne) SetNillableRevokedBy(s *string) *JobRunnerTokenUpdateOne {
	if s != nil {
		jrtuo.SetRevokedBy(*s)
	}
	return jrtuo
}

// ClearRevokedBy clears the value of the "revoked_by" field.
func (jrtuo *JobRunnerTokenUpdateOne) ClearRevokedBy() *JobRunnerTokenUpdateOne {
	jrtuo.mutation.ClearRevokedBy()
	return jrtuo
}

// SetRevokedAt sets the "revoked_at" field.
func (jrtuo *JobRunnerTokenUpdateOne) SetRevokedAt(t time.Time) *JobRunnerTokenUpdateOne {
	jrtuo.mutation.SetRevokedAt(t)
	return jrtuo
}

// SetNillableRevokedAt sets the "revoked_at" field if the given value is not nil.
func (jrtuo *JobRunnerTokenUpdateOne) SetNillableRevokedAt(t *time.Time) *JobRunnerTokenUpdateOne {
	if t != nil {
		jrtuo.SetRevokedAt(*t)
	}
	return jrtuo
}

// ClearRevokedAt clears the value of the "revoked_at" field.
func (jrtuo *JobRunnerTokenUpdateOne) ClearRevokedAt() *JobRunnerTokenUpdateOne {
	jrtuo.mutation.ClearRevokedAt()
	return jrtuo
}

// SetOwner sets the "owner" edge to the Organization entity.
func (jrtuo *JobRunnerTokenUpdateOne) SetOwner(o *Organization) *JobRunnerTokenUpdateOne {
	return jrtuo.SetOwnerID(o.ID)
}

// AddJobRunnerIDs adds the "job_runners" edge to the JobRunner entity by IDs.
func (jrtuo *JobRunnerTokenUpdateOne) AddJobRunnerIDs(ids ...string) *JobRunnerTokenUpdateOne {
	jrtuo.mutation.AddJobRunnerIDs(ids...)
	return jrtuo
}

// AddJobRunners adds the "job_runners" edges to the JobRunner entity.
func (jrtuo *JobRunnerTokenUpdateOne) AddJobRunners(j ...*JobRunner) *JobRunnerTokenUpdateOne {
	ids := make([]string, len(j))
	for i := range j {
		ids[i] = j[i].ID
	}
	return jrtuo.AddJobRunnerIDs(ids...)
}

// Mutation returns the JobRunnerTokenMutation object of the builder.
func (jrtuo *JobRunnerTokenUpdateOne) Mutation() *JobRunnerTokenMutation {
	return jrtuo.mutation
}

// ClearOwner clears the "owner" edge to the Organization entity.
func (jrtuo *JobRunnerTokenUpdateOne) ClearOwner() *JobRunnerTokenUpdateOne {
	jrtuo.mutation.ClearOwner()
	return jrtuo
}

// ClearJobRunners clears all "job_runners" edges to the JobRunner entity.
func (jrtuo *JobRunnerTokenUpdateOne) ClearJobRunners() *JobRunnerTokenUpdateOne {
	jrtuo.mutation.ClearJobRunners()
	return jrtuo
}

// RemoveJobRunnerIDs removes the "job_runners" edge to JobRunner entities by IDs.
func (jrtuo *JobRunnerTokenUpdateOne) RemoveJobRunnerIDs(ids ...string) *JobRunnerTokenUpdateOne {
	jrtuo.mutation.RemoveJobRunnerIDs(ids...)
	return jrtuo
}

// RemoveJobRunners removes "job_runners" edges to JobRunner entities.
func (jrtuo *JobRunnerTokenUpdateOne) RemoveJobRunners(j ...*JobRunner) *JobRunnerTokenUpdateOne {
	ids := make([]string, len(j))
	for i := range j {
		ids[i] = j[i].ID
	}
	return jrtuo.RemoveJobRunnerIDs(ids...)
}

// Where appends a list predicates to the JobRunnerTokenUpdate builder.
func (jrtuo *JobRunnerTokenUpdateOne) Where(ps ...predicate.JobRunnerToken) *JobRunnerTokenUpdateOne {
	jrtuo.mutation.Where(ps...)
	return jrtuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (jrtuo *JobRunnerTokenUpdateOne) Select(field string, fields ...string) *JobRunnerTokenUpdateOne {
	jrtuo.fields = append([]string{field}, fields...)
	return jrtuo
}

// Save executes the query and returns the updated JobRunnerToken entity.
func (jrtuo *JobRunnerTokenUpdateOne) Save(ctx context.Context) (*JobRunnerToken, error) {
	if err := jrtuo.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, jrtuo.sqlSave, jrtuo.mutation, jrtuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (jrtuo *JobRunnerTokenUpdateOne) SaveX(ctx context.Context) *JobRunnerToken {
	node, err := jrtuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (jrtuo *JobRunnerTokenUpdateOne) Exec(ctx context.Context) error {
	_, err := jrtuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (jrtuo *JobRunnerTokenUpdateOne) ExecX(ctx context.Context) {
	if err := jrtuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (jrtuo *JobRunnerTokenUpdateOne) defaults() error {
	if _, ok := jrtuo.mutation.UpdatedAt(); !ok && !jrtuo.mutation.UpdatedAtCleared() {
		if jobrunnertoken.UpdateDefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized jobrunnertoken.UpdateDefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := jobrunnertoken.UpdateDefaultUpdatedAt()
		jrtuo.mutation.SetUpdatedAt(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (jrtuo *JobRunnerTokenUpdateOne) check() error {
	if v, ok := jrtuo.mutation.OwnerID(); ok {
		if err := jobrunnertoken.OwnerIDValidator(v); err != nil {
			return &ValidationError{Name: "owner_id", err: fmt.Errorf(`generated: validator failed for field "JobRunnerToken.owner_id": %w`, err)}
		}
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (jrtuo *JobRunnerTokenUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *JobRunnerTokenUpdateOne {
	jrtuo.modifiers = append(jrtuo.modifiers, modifiers...)
	return jrtuo
}

func (jrtuo *JobRunnerTokenUpdateOne) sqlSave(ctx context.Context) (_node *JobRunnerToken, err error) {
	if err := jrtuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(jobrunnertoken.Table, jobrunnertoken.Columns, sqlgraph.NewFieldSpec(jobrunnertoken.FieldID, field.TypeString))
	id, ok := jrtuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`generated: missing "JobRunnerToken.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := jrtuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, jobrunnertoken.FieldID)
		for _, f := range fields {
			if !jobrunnertoken.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("generated: invalid field %q for query", f)}
			}
			if f != jobrunnertoken.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := jrtuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if jrtuo.mutation.CreatedAtCleared() {
		_spec.ClearField(jobrunnertoken.FieldCreatedAt, field.TypeTime)
	}
	if value, ok := jrtuo.mutation.UpdatedAt(); ok {
		_spec.SetField(jobrunnertoken.FieldUpdatedAt, field.TypeTime, value)
	}
	if jrtuo.mutation.UpdatedAtCleared() {
		_spec.ClearField(jobrunnertoken.FieldUpdatedAt, field.TypeTime)
	}
	if jrtuo.mutation.CreatedByCleared() {
		_spec.ClearField(jobrunnertoken.FieldCreatedBy, field.TypeString)
	}
	if value, ok := jrtuo.mutation.UpdatedBy(); ok {
		_spec.SetField(jobrunnertoken.FieldUpdatedBy, field.TypeString, value)
	}
	if jrtuo.mutation.UpdatedByCleared() {
		_spec.ClearField(jobrunnertoken.FieldUpdatedBy, field.TypeString)
	}
	if value, ok := jrtuo.mutation.DeletedAt(); ok {
		_spec.SetField(jobrunnertoken.FieldDeletedAt, field.TypeTime, value)
	}
	if jrtuo.mutation.DeletedAtCleared() {
		_spec.ClearField(jobrunnertoken.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := jrtuo.mutation.DeletedBy(); ok {
		_spec.SetField(jobrunnertoken.FieldDeletedBy, field.TypeString, value)
	}
	if jrtuo.mutation.DeletedByCleared() {
		_spec.ClearField(jobrunnertoken.FieldDeletedBy, field.TypeString)
	}
	if value, ok := jrtuo.mutation.Tags(); ok {
		_spec.SetField(jobrunnertoken.FieldTags, field.TypeJSON, value)
	}
	if value, ok := jrtuo.mutation.AppendedTags(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, jobrunnertoken.FieldTags, value)
		})
	}
	if jrtuo.mutation.TagsCleared() {
		_spec.ClearField(jobrunnertoken.FieldTags, field.TypeJSON)
	}
	if value, ok := jrtuo.mutation.ExpiresAt(); ok {
		_spec.SetField(jobrunnertoken.FieldExpiresAt, field.TypeTime, value)
	}
	if jrtuo.mutation.ExpiresAtCleared() {
		_spec.ClearField(jobrunnertoken.FieldExpiresAt, field.TypeTime)
	}
	if value, ok := jrtuo.mutation.LastUsedAt(); ok {
		_spec.SetField(jobrunnertoken.FieldLastUsedAt, field.TypeTime, value)
	}
	if jrtuo.mutation.LastUsedAtCleared() {
		_spec.ClearField(jobrunnertoken.FieldLastUsedAt, field.TypeTime)
	}
	if value, ok := jrtuo.mutation.IsActive(); ok {
		_spec.SetField(jobrunnertoken.FieldIsActive, field.TypeBool, value)
	}
	if jrtuo.mutation.IsActiveCleared() {
		_spec.ClearField(jobrunnertoken.FieldIsActive, field.TypeBool)
	}
	if value, ok := jrtuo.mutation.RevokedReason(); ok {
		_spec.SetField(jobrunnertoken.FieldRevokedReason, field.TypeString, value)
	}
	if jrtuo.mutation.RevokedReasonCleared() {
		_spec.ClearField(jobrunnertoken.FieldRevokedReason, field.TypeString)
	}
	if value, ok := jrtuo.mutation.RevokedBy(); ok {
		_spec.SetField(jobrunnertoken.FieldRevokedBy, field.TypeString, value)
	}
	if jrtuo.mutation.RevokedByCleared() {
		_spec.ClearField(jobrunnertoken.FieldRevokedBy, field.TypeString)
	}
	if value, ok := jrtuo.mutation.RevokedAt(); ok {
		_spec.SetField(jobrunnertoken.FieldRevokedAt, field.TypeTime, value)
	}
	if jrtuo.mutation.RevokedAtCleared() {
		_spec.ClearField(jobrunnertoken.FieldRevokedAt, field.TypeTime)
	}
	if jrtuo.mutation.OwnerCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   jobrunnertoken.OwnerTable,
			Columns: []string{jobrunnertoken.OwnerColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(organization.FieldID, field.TypeString),
			},
		}
		edge.Schema = jrtuo.schemaConfig.JobRunnerToken
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := jrtuo.mutation.OwnerIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   jobrunnertoken.OwnerTable,
			Columns: []string{jobrunnertoken.OwnerColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(organization.FieldID, field.TypeString),
			},
		}
		edge.Schema = jrtuo.schemaConfig.JobRunnerToken
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if jrtuo.mutation.JobRunnersCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   jobrunnertoken.JobRunnersTable,
			Columns: jobrunnertoken.JobRunnersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(jobrunner.FieldID, field.TypeString),
			},
		}
		edge.Schema = jrtuo.schemaConfig.JobRunnerJobRunnerTokens
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := jrtuo.mutation.RemovedJobRunnersIDs(); len(nodes) > 0 && !jrtuo.mutation.JobRunnersCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   jobrunnertoken.JobRunnersTable,
			Columns: jobrunnertoken.JobRunnersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(jobrunner.FieldID, field.TypeString),
			},
		}
		edge.Schema = jrtuo.schemaConfig.JobRunnerJobRunnerTokens
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := jrtuo.mutation.JobRunnersIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   jobrunnertoken.JobRunnersTable,
			Columns: jobrunnertoken.JobRunnersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(jobrunner.FieldID, field.TypeString),
			},
		}
		edge.Schema = jrtuo.schemaConfig.JobRunnerJobRunnerTokens
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.Node.Schema = jrtuo.schemaConfig.JobRunnerToken
	ctx = internal.NewSchemaConfigContext(ctx, jrtuo.schemaConfig)
	_spec.AddModifiers(jrtuo.modifiers...)
	_node = &JobRunnerToken{config: jrtuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, jrtuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{jobrunnertoken.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	jrtuo.mutation.done = true
	return _node, nil
}
