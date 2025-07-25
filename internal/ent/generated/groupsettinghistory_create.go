// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/groupsettinghistory"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/entx/history"
)

// GroupSettingHistoryCreate is the builder for creating a GroupSettingHistory entity.
type GroupSettingHistoryCreate struct {
	config
	mutation *GroupSettingHistoryMutation
	hooks    []Hook
}

// SetHistoryTime sets the "history_time" field.
func (gshc *GroupSettingHistoryCreate) SetHistoryTime(t time.Time) *GroupSettingHistoryCreate {
	gshc.mutation.SetHistoryTime(t)
	return gshc
}

// SetNillableHistoryTime sets the "history_time" field if the given value is not nil.
func (gshc *GroupSettingHistoryCreate) SetNillableHistoryTime(t *time.Time) *GroupSettingHistoryCreate {
	if t != nil {
		gshc.SetHistoryTime(*t)
	}
	return gshc
}

// SetRef sets the "ref" field.
func (gshc *GroupSettingHistoryCreate) SetRef(s string) *GroupSettingHistoryCreate {
	gshc.mutation.SetRef(s)
	return gshc
}

// SetNillableRef sets the "ref" field if the given value is not nil.
func (gshc *GroupSettingHistoryCreate) SetNillableRef(s *string) *GroupSettingHistoryCreate {
	if s != nil {
		gshc.SetRef(*s)
	}
	return gshc
}

// SetOperation sets the "operation" field.
func (gshc *GroupSettingHistoryCreate) SetOperation(ht history.OpType) *GroupSettingHistoryCreate {
	gshc.mutation.SetOperation(ht)
	return gshc
}

// SetCreatedAt sets the "created_at" field.
func (gshc *GroupSettingHistoryCreate) SetCreatedAt(t time.Time) *GroupSettingHistoryCreate {
	gshc.mutation.SetCreatedAt(t)
	return gshc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (gshc *GroupSettingHistoryCreate) SetNillableCreatedAt(t *time.Time) *GroupSettingHistoryCreate {
	if t != nil {
		gshc.SetCreatedAt(*t)
	}
	return gshc
}

// SetUpdatedAt sets the "updated_at" field.
func (gshc *GroupSettingHistoryCreate) SetUpdatedAt(t time.Time) *GroupSettingHistoryCreate {
	gshc.mutation.SetUpdatedAt(t)
	return gshc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (gshc *GroupSettingHistoryCreate) SetNillableUpdatedAt(t *time.Time) *GroupSettingHistoryCreate {
	if t != nil {
		gshc.SetUpdatedAt(*t)
	}
	return gshc
}

// SetCreatedBy sets the "created_by" field.
func (gshc *GroupSettingHistoryCreate) SetCreatedBy(s string) *GroupSettingHistoryCreate {
	gshc.mutation.SetCreatedBy(s)
	return gshc
}

// SetNillableCreatedBy sets the "created_by" field if the given value is not nil.
func (gshc *GroupSettingHistoryCreate) SetNillableCreatedBy(s *string) *GroupSettingHistoryCreate {
	if s != nil {
		gshc.SetCreatedBy(*s)
	}
	return gshc
}

// SetUpdatedBy sets the "updated_by" field.
func (gshc *GroupSettingHistoryCreate) SetUpdatedBy(s string) *GroupSettingHistoryCreate {
	gshc.mutation.SetUpdatedBy(s)
	return gshc
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (gshc *GroupSettingHistoryCreate) SetNillableUpdatedBy(s *string) *GroupSettingHistoryCreate {
	if s != nil {
		gshc.SetUpdatedBy(*s)
	}
	return gshc
}

// SetDeletedAt sets the "deleted_at" field.
func (gshc *GroupSettingHistoryCreate) SetDeletedAt(t time.Time) *GroupSettingHistoryCreate {
	gshc.mutation.SetDeletedAt(t)
	return gshc
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (gshc *GroupSettingHistoryCreate) SetNillableDeletedAt(t *time.Time) *GroupSettingHistoryCreate {
	if t != nil {
		gshc.SetDeletedAt(*t)
	}
	return gshc
}

// SetDeletedBy sets the "deleted_by" field.
func (gshc *GroupSettingHistoryCreate) SetDeletedBy(s string) *GroupSettingHistoryCreate {
	gshc.mutation.SetDeletedBy(s)
	return gshc
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (gshc *GroupSettingHistoryCreate) SetNillableDeletedBy(s *string) *GroupSettingHistoryCreate {
	if s != nil {
		gshc.SetDeletedBy(*s)
	}
	return gshc
}

// SetVisibility sets the "visibility" field.
func (gshc *GroupSettingHistoryCreate) SetVisibility(e enums.Visibility) *GroupSettingHistoryCreate {
	gshc.mutation.SetVisibility(e)
	return gshc
}

// SetNillableVisibility sets the "visibility" field if the given value is not nil.
func (gshc *GroupSettingHistoryCreate) SetNillableVisibility(e *enums.Visibility) *GroupSettingHistoryCreate {
	if e != nil {
		gshc.SetVisibility(*e)
	}
	return gshc
}

// SetJoinPolicy sets the "join_policy" field.
func (gshc *GroupSettingHistoryCreate) SetJoinPolicy(ep enums.JoinPolicy) *GroupSettingHistoryCreate {
	gshc.mutation.SetJoinPolicy(ep)
	return gshc
}

// SetNillableJoinPolicy sets the "join_policy" field if the given value is not nil.
func (gshc *GroupSettingHistoryCreate) SetNillableJoinPolicy(ep *enums.JoinPolicy) *GroupSettingHistoryCreate {
	if ep != nil {
		gshc.SetJoinPolicy(*ep)
	}
	return gshc
}

// SetSyncToSlack sets the "sync_to_slack" field.
func (gshc *GroupSettingHistoryCreate) SetSyncToSlack(b bool) *GroupSettingHistoryCreate {
	gshc.mutation.SetSyncToSlack(b)
	return gshc
}

// SetNillableSyncToSlack sets the "sync_to_slack" field if the given value is not nil.
func (gshc *GroupSettingHistoryCreate) SetNillableSyncToSlack(b *bool) *GroupSettingHistoryCreate {
	if b != nil {
		gshc.SetSyncToSlack(*b)
	}
	return gshc
}

// SetSyncToGithub sets the "sync_to_github" field.
func (gshc *GroupSettingHistoryCreate) SetSyncToGithub(b bool) *GroupSettingHistoryCreate {
	gshc.mutation.SetSyncToGithub(b)
	return gshc
}

// SetNillableSyncToGithub sets the "sync_to_github" field if the given value is not nil.
func (gshc *GroupSettingHistoryCreate) SetNillableSyncToGithub(b *bool) *GroupSettingHistoryCreate {
	if b != nil {
		gshc.SetSyncToGithub(*b)
	}
	return gshc
}

// SetGroupID sets the "group_id" field.
func (gshc *GroupSettingHistoryCreate) SetGroupID(s string) *GroupSettingHistoryCreate {
	gshc.mutation.SetGroupID(s)
	return gshc
}

// SetNillableGroupID sets the "group_id" field if the given value is not nil.
func (gshc *GroupSettingHistoryCreate) SetNillableGroupID(s *string) *GroupSettingHistoryCreate {
	if s != nil {
		gshc.SetGroupID(*s)
	}
	return gshc
}

// SetID sets the "id" field.
func (gshc *GroupSettingHistoryCreate) SetID(s string) *GroupSettingHistoryCreate {
	gshc.mutation.SetID(s)
	return gshc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (gshc *GroupSettingHistoryCreate) SetNillableID(s *string) *GroupSettingHistoryCreate {
	if s != nil {
		gshc.SetID(*s)
	}
	return gshc
}

// Mutation returns the GroupSettingHistoryMutation object of the builder.
func (gshc *GroupSettingHistoryCreate) Mutation() *GroupSettingHistoryMutation {
	return gshc.mutation
}

// Save creates the GroupSettingHistory in the database.
func (gshc *GroupSettingHistoryCreate) Save(ctx context.Context) (*GroupSettingHistory, error) {
	if err := gshc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, gshc.sqlSave, gshc.mutation, gshc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (gshc *GroupSettingHistoryCreate) SaveX(ctx context.Context) *GroupSettingHistory {
	v, err := gshc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (gshc *GroupSettingHistoryCreate) Exec(ctx context.Context) error {
	_, err := gshc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (gshc *GroupSettingHistoryCreate) ExecX(ctx context.Context) {
	if err := gshc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (gshc *GroupSettingHistoryCreate) defaults() error {
	if _, ok := gshc.mutation.HistoryTime(); !ok {
		if groupsettinghistory.DefaultHistoryTime == nil {
			return fmt.Errorf("generated: uninitialized groupsettinghistory.DefaultHistoryTime (forgotten import generated/runtime?)")
		}
		v := groupsettinghistory.DefaultHistoryTime()
		gshc.mutation.SetHistoryTime(v)
	}
	if _, ok := gshc.mutation.CreatedAt(); !ok {
		if groupsettinghistory.DefaultCreatedAt == nil {
			return fmt.Errorf("generated: uninitialized groupsettinghistory.DefaultCreatedAt (forgotten import generated/runtime?)")
		}
		v := groupsettinghistory.DefaultCreatedAt()
		gshc.mutation.SetCreatedAt(v)
	}
	if _, ok := gshc.mutation.UpdatedAt(); !ok {
		if groupsettinghistory.DefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized groupsettinghistory.DefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := groupsettinghistory.DefaultUpdatedAt()
		gshc.mutation.SetUpdatedAt(v)
	}
	if _, ok := gshc.mutation.Visibility(); !ok {
		v := groupsettinghistory.DefaultVisibility
		gshc.mutation.SetVisibility(v)
	}
	if _, ok := gshc.mutation.JoinPolicy(); !ok {
		v := groupsettinghistory.DefaultJoinPolicy
		gshc.mutation.SetJoinPolicy(v)
	}
	if _, ok := gshc.mutation.SyncToSlack(); !ok {
		v := groupsettinghistory.DefaultSyncToSlack
		gshc.mutation.SetSyncToSlack(v)
	}
	if _, ok := gshc.mutation.SyncToGithub(); !ok {
		v := groupsettinghistory.DefaultSyncToGithub
		gshc.mutation.SetSyncToGithub(v)
	}
	if _, ok := gshc.mutation.ID(); !ok {
		if groupsettinghistory.DefaultID == nil {
			return fmt.Errorf("generated: uninitialized groupsettinghistory.DefaultID (forgotten import generated/runtime?)")
		}
		v := groupsettinghistory.DefaultID()
		gshc.mutation.SetID(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (gshc *GroupSettingHistoryCreate) check() error {
	if _, ok := gshc.mutation.HistoryTime(); !ok {
		return &ValidationError{Name: "history_time", err: errors.New(`generated: missing required field "GroupSettingHistory.history_time"`)}
	}
	if _, ok := gshc.mutation.Operation(); !ok {
		return &ValidationError{Name: "operation", err: errors.New(`generated: missing required field "GroupSettingHistory.operation"`)}
	}
	if v, ok := gshc.mutation.Operation(); ok {
		if err := groupsettinghistory.OperationValidator(v); err != nil {
			return &ValidationError{Name: "operation", err: fmt.Errorf(`generated: validator failed for field "GroupSettingHistory.operation": %w`, err)}
		}
	}
	if _, ok := gshc.mutation.Visibility(); !ok {
		return &ValidationError{Name: "visibility", err: errors.New(`generated: missing required field "GroupSettingHistory.visibility"`)}
	}
	if v, ok := gshc.mutation.Visibility(); ok {
		if err := groupsettinghistory.VisibilityValidator(v); err != nil {
			return &ValidationError{Name: "visibility", err: fmt.Errorf(`generated: validator failed for field "GroupSettingHistory.visibility": %w`, err)}
		}
	}
	if _, ok := gshc.mutation.JoinPolicy(); !ok {
		return &ValidationError{Name: "join_policy", err: errors.New(`generated: missing required field "GroupSettingHistory.join_policy"`)}
	}
	if v, ok := gshc.mutation.JoinPolicy(); ok {
		if err := groupsettinghistory.JoinPolicyValidator(v); err != nil {
			return &ValidationError{Name: "join_policy", err: fmt.Errorf(`generated: validator failed for field "GroupSettingHistory.join_policy": %w`, err)}
		}
	}
	return nil
}

func (gshc *GroupSettingHistoryCreate) sqlSave(ctx context.Context) (*GroupSettingHistory, error) {
	if err := gshc.check(); err != nil {
		return nil, err
	}
	_node, _spec := gshc.createSpec()
	if err := sqlgraph.CreateNode(ctx, gshc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected GroupSettingHistory.ID type: %T", _spec.ID.Value)
		}
	}
	gshc.mutation.id = &_node.ID
	gshc.mutation.done = true
	return _node, nil
}

func (gshc *GroupSettingHistoryCreate) createSpec() (*GroupSettingHistory, *sqlgraph.CreateSpec) {
	var (
		_node = &GroupSettingHistory{config: gshc.config}
		_spec = sqlgraph.NewCreateSpec(groupsettinghistory.Table, sqlgraph.NewFieldSpec(groupsettinghistory.FieldID, field.TypeString))
	)
	_spec.Schema = gshc.schemaConfig.GroupSettingHistory
	if id, ok := gshc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := gshc.mutation.HistoryTime(); ok {
		_spec.SetField(groupsettinghistory.FieldHistoryTime, field.TypeTime, value)
		_node.HistoryTime = value
	}
	if value, ok := gshc.mutation.Ref(); ok {
		_spec.SetField(groupsettinghistory.FieldRef, field.TypeString, value)
		_node.Ref = value
	}
	if value, ok := gshc.mutation.Operation(); ok {
		_spec.SetField(groupsettinghistory.FieldOperation, field.TypeEnum, value)
		_node.Operation = value
	}
	if value, ok := gshc.mutation.CreatedAt(); ok {
		_spec.SetField(groupsettinghistory.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := gshc.mutation.UpdatedAt(); ok {
		_spec.SetField(groupsettinghistory.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := gshc.mutation.CreatedBy(); ok {
		_spec.SetField(groupsettinghistory.FieldCreatedBy, field.TypeString, value)
		_node.CreatedBy = value
	}
	if value, ok := gshc.mutation.UpdatedBy(); ok {
		_spec.SetField(groupsettinghistory.FieldUpdatedBy, field.TypeString, value)
		_node.UpdatedBy = value
	}
	if value, ok := gshc.mutation.DeletedAt(); ok {
		_spec.SetField(groupsettinghistory.FieldDeletedAt, field.TypeTime, value)
		_node.DeletedAt = value
	}
	if value, ok := gshc.mutation.DeletedBy(); ok {
		_spec.SetField(groupsettinghistory.FieldDeletedBy, field.TypeString, value)
		_node.DeletedBy = value
	}
	if value, ok := gshc.mutation.Visibility(); ok {
		_spec.SetField(groupsettinghistory.FieldVisibility, field.TypeEnum, value)
		_node.Visibility = value
	}
	if value, ok := gshc.mutation.JoinPolicy(); ok {
		_spec.SetField(groupsettinghistory.FieldJoinPolicy, field.TypeEnum, value)
		_node.JoinPolicy = value
	}
	if value, ok := gshc.mutation.SyncToSlack(); ok {
		_spec.SetField(groupsettinghistory.FieldSyncToSlack, field.TypeBool, value)
		_node.SyncToSlack = value
	}
	if value, ok := gshc.mutation.SyncToGithub(); ok {
		_spec.SetField(groupsettinghistory.FieldSyncToGithub, field.TypeBool, value)
		_node.SyncToGithub = value
	}
	if value, ok := gshc.mutation.GroupID(); ok {
		_spec.SetField(groupsettinghistory.FieldGroupID, field.TypeString, value)
		_node.GroupID = value
	}
	return _node, _spec
}

// GroupSettingHistoryCreateBulk is the builder for creating many GroupSettingHistory entities in bulk.
type GroupSettingHistoryCreateBulk struct {
	config
	err      error
	builders []*GroupSettingHistoryCreate
}

// Save creates the GroupSettingHistory entities in the database.
func (gshcb *GroupSettingHistoryCreateBulk) Save(ctx context.Context) ([]*GroupSettingHistory, error) {
	if gshcb.err != nil {
		return nil, gshcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(gshcb.builders))
	nodes := make([]*GroupSettingHistory, len(gshcb.builders))
	mutators := make([]Mutator, len(gshcb.builders))
	for i := range gshcb.builders {
		func(i int, root context.Context) {
			builder := gshcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*GroupSettingHistoryMutation)
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
					_, err = mutators[i+1].Mutate(root, gshcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, gshcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, gshcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (gshcb *GroupSettingHistoryCreateBulk) SaveX(ctx context.Context) []*GroupSettingHistory {
	v, err := gshcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (gshcb *GroupSettingHistoryCreateBulk) Exec(ctx context.Context) error {
	_, err := gshcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (gshcb *GroupSettingHistoryCreateBulk) ExecX(ctx context.Context) {
	if err := gshcb.Exec(ctx); err != nil {
		panic(err)
	}
}
