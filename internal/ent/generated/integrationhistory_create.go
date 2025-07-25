// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/integrationhistory"
	"github.com/theopenlane/entx/history"
)

// IntegrationHistoryCreate is the builder for creating a IntegrationHistory entity.
type IntegrationHistoryCreate struct {
	config
	mutation *IntegrationHistoryMutation
	hooks    []Hook
}

// SetHistoryTime sets the "history_time" field.
func (ihc *IntegrationHistoryCreate) SetHistoryTime(t time.Time) *IntegrationHistoryCreate {
	ihc.mutation.SetHistoryTime(t)
	return ihc
}

// SetNillableHistoryTime sets the "history_time" field if the given value is not nil.
func (ihc *IntegrationHistoryCreate) SetNillableHistoryTime(t *time.Time) *IntegrationHistoryCreate {
	if t != nil {
		ihc.SetHistoryTime(*t)
	}
	return ihc
}

// SetRef sets the "ref" field.
func (ihc *IntegrationHistoryCreate) SetRef(s string) *IntegrationHistoryCreate {
	ihc.mutation.SetRef(s)
	return ihc
}

// SetNillableRef sets the "ref" field if the given value is not nil.
func (ihc *IntegrationHistoryCreate) SetNillableRef(s *string) *IntegrationHistoryCreate {
	if s != nil {
		ihc.SetRef(*s)
	}
	return ihc
}

// SetOperation sets the "operation" field.
func (ihc *IntegrationHistoryCreate) SetOperation(ht history.OpType) *IntegrationHistoryCreate {
	ihc.mutation.SetOperation(ht)
	return ihc
}

// SetCreatedAt sets the "created_at" field.
func (ihc *IntegrationHistoryCreate) SetCreatedAt(t time.Time) *IntegrationHistoryCreate {
	ihc.mutation.SetCreatedAt(t)
	return ihc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (ihc *IntegrationHistoryCreate) SetNillableCreatedAt(t *time.Time) *IntegrationHistoryCreate {
	if t != nil {
		ihc.SetCreatedAt(*t)
	}
	return ihc
}

// SetUpdatedAt sets the "updated_at" field.
func (ihc *IntegrationHistoryCreate) SetUpdatedAt(t time.Time) *IntegrationHistoryCreate {
	ihc.mutation.SetUpdatedAt(t)
	return ihc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (ihc *IntegrationHistoryCreate) SetNillableUpdatedAt(t *time.Time) *IntegrationHistoryCreate {
	if t != nil {
		ihc.SetUpdatedAt(*t)
	}
	return ihc
}

// SetCreatedBy sets the "created_by" field.
func (ihc *IntegrationHistoryCreate) SetCreatedBy(s string) *IntegrationHistoryCreate {
	ihc.mutation.SetCreatedBy(s)
	return ihc
}

// SetNillableCreatedBy sets the "created_by" field if the given value is not nil.
func (ihc *IntegrationHistoryCreate) SetNillableCreatedBy(s *string) *IntegrationHistoryCreate {
	if s != nil {
		ihc.SetCreatedBy(*s)
	}
	return ihc
}

// SetUpdatedBy sets the "updated_by" field.
func (ihc *IntegrationHistoryCreate) SetUpdatedBy(s string) *IntegrationHistoryCreate {
	ihc.mutation.SetUpdatedBy(s)
	return ihc
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (ihc *IntegrationHistoryCreate) SetNillableUpdatedBy(s *string) *IntegrationHistoryCreate {
	if s != nil {
		ihc.SetUpdatedBy(*s)
	}
	return ihc
}

// SetDeletedAt sets the "deleted_at" field.
func (ihc *IntegrationHistoryCreate) SetDeletedAt(t time.Time) *IntegrationHistoryCreate {
	ihc.mutation.SetDeletedAt(t)
	return ihc
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (ihc *IntegrationHistoryCreate) SetNillableDeletedAt(t *time.Time) *IntegrationHistoryCreate {
	if t != nil {
		ihc.SetDeletedAt(*t)
	}
	return ihc
}

// SetDeletedBy sets the "deleted_by" field.
func (ihc *IntegrationHistoryCreate) SetDeletedBy(s string) *IntegrationHistoryCreate {
	ihc.mutation.SetDeletedBy(s)
	return ihc
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (ihc *IntegrationHistoryCreate) SetNillableDeletedBy(s *string) *IntegrationHistoryCreate {
	if s != nil {
		ihc.SetDeletedBy(*s)
	}
	return ihc
}

// SetTags sets the "tags" field.
func (ihc *IntegrationHistoryCreate) SetTags(s []string) *IntegrationHistoryCreate {
	ihc.mutation.SetTags(s)
	return ihc
}

// SetOwnerID sets the "owner_id" field.
func (ihc *IntegrationHistoryCreate) SetOwnerID(s string) *IntegrationHistoryCreate {
	ihc.mutation.SetOwnerID(s)
	return ihc
}

// SetNillableOwnerID sets the "owner_id" field if the given value is not nil.
func (ihc *IntegrationHistoryCreate) SetNillableOwnerID(s *string) *IntegrationHistoryCreate {
	if s != nil {
		ihc.SetOwnerID(*s)
	}
	return ihc
}

// SetName sets the "name" field.
func (ihc *IntegrationHistoryCreate) SetName(s string) *IntegrationHistoryCreate {
	ihc.mutation.SetName(s)
	return ihc
}

// SetDescription sets the "description" field.
func (ihc *IntegrationHistoryCreate) SetDescription(s string) *IntegrationHistoryCreate {
	ihc.mutation.SetDescription(s)
	return ihc
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (ihc *IntegrationHistoryCreate) SetNillableDescription(s *string) *IntegrationHistoryCreate {
	if s != nil {
		ihc.SetDescription(*s)
	}
	return ihc
}

// SetKind sets the "kind" field.
func (ihc *IntegrationHistoryCreate) SetKind(s string) *IntegrationHistoryCreate {
	ihc.mutation.SetKind(s)
	return ihc
}

// SetNillableKind sets the "kind" field if the given value is not nil.
func (ihc *IntegrationHistoryCreate) SetNillableKind(s *string) *IntegrationHistoryCreate {
	if s != nil {
		ihc.SetKind(*s)
	}
	return ihc
}

// SetID sets the "id" field.
func (ihc *IntegrationHistoryCreate) SetID(s string) *IntegrationHistoryCreate {
	ihc.mutation.SetID(s)
	return ihc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (ihc *IntegrationHistoryCreate) SetNillableID(s *string) *IntegrationHistoryCreate {
	if s != nil {
		ihc.SetID(*s)
	}
	return ihc
}

// Mutation returns the IntegrationHistoryMutation object of the builder.
func (ihc *IntegrationHistoryCreate) Mutation() *IntegrationHistoryMutation {
	return ihc.mutation
}

// Save creates the IntegrationHistory in the database.
func (ihc *IntegrationHistoryCreate) Save(ctx context.Context) (*IntegrationHistory, error) {
	if err := ihc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, ihc.sqlSave, ihc.mutation, ihc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (ihc *IntegrationHistoryCreate) SaveX(ctx context.Context) *IntegrationHistory {
	v, err := ihc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ihc *IntegrationHistoryCreate) Exec(ctx context.Context) error {
	_, err := ihc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ihc *IntegrationHistoryCreate) ExecX(ctx context.Context) {
	if err := ihc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (ihc *IntegrationHistoryCreate) defaults() error {
	if _, ok := ihc.mutation.HistoryTime(); !ok {
		if integrationhistory.DefaultHistoryTime == nil {
			return fmt.Errorf("generated: uninitialized integrationhistory.DefaultHistoryTime (forgotten import generated/runtime?)")
		}
		v := integrationhistory.DefaultHistoryTime()
		ihc.mutation.SetHistoryTime(v)
	}
	if _, ok := ihc.mutation.CreatedAt(); !ok {
		if integrationhistory.DefaultCreatedAt == nil {
			return fmt.Errorf("generated: uninitialized integrationhistory.DefaultCreatedAt (forgotten import generated/runtime?)")
		}
		v := integrationhistory.DefaultCreatedAt()
		ihc.mutation.SetCreatedAt(v)
	}
	if _, ok := ihc.mutation.UpdatedAt(); !ok {
		if integrationhistory.DefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized integrationhistory.DefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := integrationhistory.DefaultUpdatedAt()
		ihc.mutation.SetUpdatedAt(v)
	}
	if _, ok := ihc.mutation.Tags(); !ok {
		v := integrationhistory.DefaultTags
		ihc.mutation.SetTags(v)
	}
	if _, ok := ihc.mutation.ID(); !ok {
		if integrationhistory.DefaultID == nil {
			return fmt.Errorf("generated: uninitialized integrationhistory.DefaultID (forgotten import generated/runtime?)")
		}
		v := integrationhistory.DefaultID()
		ihc.mutation.SetID(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (ihc *IntegrationHistoryCreate) check() error {
	if _, ok := ihc.mutation.HistoryTime(); !ok {
		return &ValidationError{Name: "history_time", err: errors.New(`generated: missing required field "IntegrationHistory.history_time"`)}
	}
	if _, ok := ihc.mutation.Operation(); !ok {
		return &ValidationError{Name: "operation", err: errors.New(`generated: missing required field "IntegrationHistory.operation"`)}
	}
	if v, ok := ihc.mutation.Operation(); ok {
		if err := integrationhistory.OperationValidator(v); err != nil {
			return &ValidationError{Name: "operation", err: fmt.Errorf(`generated: validator failed for field "IntegrationHistory.operation": %w`, err)}
		}
	}
	if _, ok := ihc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`generated: missing required field "IntegrationHistory.name"`)}
	}
	return nil
}

func (ihc *IntegrationHistoryCreate) sqlSave(ctx context.Context) (*IntegrationHistory, error) {
	if err := ihc.check(); err != nil {
		return nil, err
	}
	_node, _spec := ihc.createSpec()
	if err := sqlgraph.CreateNode(ctx, ihc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected IntegrationHistory.ID type: %T", _spec.ID.Value)
		}
	}
	ihc.mutation.id = &_node.ID
	ihc.mutation.done = true
	return _node, nil
}

func (ihc *IntegrationHistoryCreate) createSpec() (*IntegrationHistory, *sqlgraph.CreateSpec) {
	var (
		_node = &IntegrationHistory{config: ihc.config}
		_spec = sqlgraph.NewCreateSpec(integrationhistory.Table, sqlgraph.NewFieldSpec(integrationhistory.FieldID, field.TypeString))
	)
	_spec.Schema = ihc.schemaConfig.IntegrationHistory
	if id, ok := ihc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := ihc.mutation.HistoryTime(); ok {
		_spec.SetField(integrationhistory.FieldHistoryTime, field.TypeTime, value)
		_node.HistoryTime = value
	}
	if value, ok := ihc.mutation.Ref(); ok {
		_spec.SetField(integrationhistory.FieldRef, field.TypeString, value)
		_node.Ref = value
	}
	if value, ok := ihc.mutation.Operation(); ok {
		_spec.SetField(integrationhistory.FieldOperation, field.TypeEnum, value)
		_node.Operation = value
	}
	if value, ok := ihc.mutation.CreatedAt(); ok {
		_spec.SetField(integrationhistory.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := ihc.mutation.UpdatedAt(); ok {
		_spec.SetField(integrationhistory.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := ihc.mutation.CreatedBy(); ok {
		_spec.SetField(integrationhistory.FieldCreatedBy, field.TypeString, value)
		_node.CreatedBy = value
	}
	if value, ok := ihc.mutation.UpdatedBy(); ok {
		_spec.SetField(integrationhistory.FieldUpdatedBy, field.TypeString, value)
		_node.UpdatedBy = value
	}
	if value, ok := ihc.mutation.DeletedAt(); ok {
		_spec.SetField(integrationhistory.FieldDeletedAt, field.TypeTime, value)
		_node.DeletedAt = value
	}
	if value, ok := ihc.mutation.DeletedBy(); ok {
		_spec.SetField(integrationhistory.FieldDeletedBy, field.TypeString, value)
		_node.DeletedBy = value
	}
	if value, ok := ihc.mutation.Tags(); ok {
		_spec.SetField(integrationhistory.FieldTags, field.TypeJSON, value)
		_node.Tags = value
	}
	if value, ok := ihc.mutation.OwnerID(); ok {
		_spec.SetField(integrationhistory.FieldOwnerID, field.TypeString, value)
		_node.OwnerID = value
	}
	if value, ok := ihc.mutation.Name(); ok {
		_spec.SetField(integrationhistory.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := ihc.mutation.Description(); ok {
		_spec.SetField(integrationhistory.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	if value, ok := ihc.mutation.Kind(); ok {
		_spec.SetField(integrationhistory.FieldKind, field.TypeString, value)
		_node.Kind = value
	}
	return _node, _spec
}

// IntegrationHistoryCreateBulk is the builder for creating many IntegrationHistory entities in bulk.
type IntegrationHistoryCreateBulk struct {
	config
	err      error
	builders []*IntegrationHistoryCreate
}

// Save creates the IntegrationHistory entities in the database.
func (ihcb *IntegrationHistoryCreateBulk) Save(ctx context.Context) ([]*IntegrationHistory, error) {
	if ihcb.err != nil {
		return nil, ihcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(ihcb.builders))
	nodes := make([]*IntegrationHistory, len(ihcb.builders))
	mutators := make([]Mutator, len(ihcb.builders))
	for i := range ihcb.builders {
		func(i int, root context.Context) {
			builder := ihcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*IntegrationHistoryMutation)
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
					_, err = mutators[i+1].Mutate(root, ihcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ihcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, ihcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ihcb *IntegrationHistoryCreateBulk) SaveX(ctx context.Context) []*IntegrationHistory {
	v, err := ihcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ihcb *IntegrationHistoryCreateBulk) Exec(ctx context.Context) error {
	_, err := ihcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ihcb *IntegrationHistoryCreateBulk) ExecX(ctx context.Context) {
	if err := ihcb.Exec(ctx); err != nil {
		panic(err)
	}
}
