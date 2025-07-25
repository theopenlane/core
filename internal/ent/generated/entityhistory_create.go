// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/entityhistory"
	"github.com/theopenlane/entx/history"
)

// EntityHistoryCreate is the builder for creating a EntityHistory entity.
type EntityHistoryCreate struct {
	config
	mutation *EntityHistoryMutation
	hooks    []Hook
}

// SetHistoryTime sets the "history_time" field.
func (ehc *EntityHistoryCreate) SetHistoryTime(t time.Time) *EntityHistoryCreate {
	ehc.mutation.SetHistoryTime(t)
	return ehc
}

// SetNillableHistoryTime sets the "history_time" field if the given value is not nil.
func (ehc *EntityHistoryCreate) SetNillableHistoryTime(t *time.Time) *EntityHistoryCreate {
	if t != nil {
		ehc.SetHistoryTime(*t)
	}
	return ehc
}

// SetRef sets the "ref" field.
func (ehc *EntityHistoryCreate) SetRef(s string) *EntityHistoryCreate {
	ehc.mutation.SetRef(s)
	return ehc
}

// SetNillableRef sets the "ref" field if the given value is not nil.
func (ehc *EntityHistoryCreate) SetNillableRef(s *string) *EntityHistoryCreate {
	if s != nil {
		ehc.SetRef(*s)
	}
	return ehc
}

// SetOperation sets the "operation" field.
func (ehc *EntityHistoryCreate) SetOperation(ht history.OpType) *EntityHistoryCreate {
	ehc.mutation.SetOperation(ht)
	return ehc
}

// SetCreatedAt sets the "created_at" field.
func (ehc *EntityHistoryCreate) SetCreatedAt(t time.Time) *EntityHistoryCreate {
	ehc.mutation.SetCreatedAt(t)
	return ehc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (ehc *EntityHistoryCreate) SetNillableCreatedAt(t *time.Time) *EntityHistoryCreate {
	if t != nil {
		ehc.SetCreatedAt(*t)
	}
	return ehc
}

// SetUpdatedAt sets the "updated_at" field.
func (ehc *EntityHistoryCreate) SetUpdatedAt(t time.Time) *EntityHistoryCreate {
	ehc.mutation.SetUpdatedAt(t)
	return ehc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (ehc *EntityHistoryCreate) SetNillableUpdatedAt(t *time.Time) *EntityHistoryCreate {
	if t != nil {
		ehc.SetUpdatedAt(*t)
	}
	return ehc
}

// SetCreatedBy sets the "created_by" field.
func (ehc *EntityHistoryCreate) SetCreatedBy(s string) *EntityHistoryCreate {
	ehc.mutation.SetCreatedBy(s)
	return ehc
}

// SetNillableCreatedBy sets the "created_by" field if the given value is not nil.
func (ehc *EntityHistoryCreate) SetNillableCreatedBy(s *string) *EntityHistoryCreate {
	if s != nil {
		ehc.SetCreatedBy(*s)
	}
	return ehc
}

// SetUpdatedBy sets the "updated_by" field.
func (ehc *EntityHistoryCreate) SetUpdatedBy(s string) *EntityHistoryCreate {
	ehc.mutation.SetUpdatedBy(s)
	return ehc
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (ehc *EntityHistoryCreate) SetNillableUpdatedBy(s *string) *EntityHistoryCreate {
	if s != nil {
		ehc.SetUpdatedBy(*s)
	}
	return ehc
}

// SetDeletedAt sets the "deleted_at" field.
func (ehc *EntityHistoryCreate) SetDeletedAt(t time.Time) *EntityHistoryCreate {
	ehc.mutation.SetDeletedAt(t)
	return ehc
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (ehc *EntityHistoryCreate) SetNillableDeletedAt(t *time.Time) *EntityHistoryCreate {
	if t != nil {
		ehc.SetDeletedAt(*t)
	}
	return ehc
}

// SetDeletedBy sets the "deleted_by" field.
func (ehc *EntityHistoryCreate) SetDeletedBy(s string) *EntityHistoryCreate {
	ehc.mutation.SetDeletedBy(s)
	return ehc
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (ehc *EntityHistoryCreate) SetNillableDeletedBy(s *string) *EntityHistoryCreate {
	if s != nil {
		ehc.SetDeletedBy(*s)
	}
	return ehc
}

// SetTags sets the "tags" field.
func (ehc *EntityHistoryCreate) SetTags(s []string) *EntityHistoryCreate {
	ehc.mutation.SetTags(s)
	return ehc
}

// SetOwnerID sets the "owner_id" field.
func (ehc *EntityHistoryCreate) SetOwnerID(s string) *EntityHistoryCreate {
	ehc.mutation.SetOwnerID(s)
	return ehc
}

// SetNillableOwnerID sets the "owner_id" field if the given value is not nil.
func (ehc *EntityHistoryCreate) SetNillableOwnerID(s *string) *EntityHistoryCreate {
	if s != nil {
		ehc.SetOwnerID(*s)
	}
	return ehc
}

// SetName sets the "name" field.
func (ehc *EntityHistoryCreate) SetName(s string) *EntityHistoryCreate {
	ehc.mutation.SetName(s)
	return ehc
}

// SetNillableName sets the "name" field if the given value is not nil.
func (ehc *EntityHistoryCreate) SetNillableName(s *string) *EntityHistoryCreate {
	if s != nil {
		ehc.SetName(*s)
	}
	return ehc
}

// SetDisplayName sets the "display_name" field.
func (ehc *EntityHistoryCreate) SetDisplayName(s string) *EntityHistoryCreate {
	ehc.mutation.SetDisplayName(s)
	return ehc
}

// SetNillableDisplayName sets the "display_name" field if the given value is not nil.
func (ehc *EntityHistoryCreate) SetNillableDisplayName(s *string) *EntityHistoryCreate {
	if s != nil {
		ehc.SetDisplayName(*s)
	}
	return ehc
}

// SetDescription sets the "description" field.
func (ehc *EntityHistoryCreate) SetDescription(s string) *EntityHistoryCreate {
	ehc.mutation.SetDescription(s)
	return ehc
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (ehc *EntityHistoryCreate) SetNillableDescription(s *string) *EntityHistoryCreate {
	if s != nil {
		ehc.SetDescription(*s)
	}
	return ehc
}

// SetDomains sets the "domains" field.
func (ehc *EntityHistoryCreate) SetDomains(s []string) *EntityHistoryCreate {
	ehc.mutation.SetDomains(s)
	return ehc
}

// SetEntityTypeID sets the "entity_type_id" field.
func (ehc *EntityHistoryCreate) SetEntityTypeID(s string) *EntityHistoryCreate {
	ehc.mutation.SetEntityTypeID(s)
	return ehc
}

// SetNillableEntityTypeID sets the "entity_type_id" field if the given value is not nil.
func (ehc *EntityHistoryCreate) SetNillableEntityTypeID(s *string) *EntityHistoryCreate {
	if s != nil {
		ehc.SetEntityTypeID(*s)
	}
	return ehc
}

// SetStatus sets the "status" field.
func (ehc *EntityHistoryCreate) SetStatus(s string) *EntityHistoryCreate {
	ehc.mutation.SetStatus(s)
	return ehc
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (ehc *EntityHistoryCreate) SetNillableStatus(s *string) *EntityHistoryCreate {
	if s != nil {
		ehc.SetStatus(*s)
	}
	return ehc
}

// SetID sets the "id" field.
func (ehc *EntityHistoryCreate) SetID(s string) *EntityHistoryCreate {
	ehc.mutation.SetID(s)
	return ehc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (ehc *EntityHistoryCreate) SetNillableID(s *string) *EntityHistoryCreate {
	if s != nil {
		ehc.SetID(*s)
	}
	return ehc
}

// Mutation returns the EntityHistoryMutation object of the builder.
func (ehc *EntityHistoryCreate) Mutation() *EntityHistoryMutation {
	return ehc.mutation
}

// Save creates the EntityHistory in the database.
func (ehc *EntityHistoryCreate) Save(ctx context.Context) (*EntityHistory, error) {
	if err := ehc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, ehc.sqlSave, ehc.mutation, ehc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (ehc *EntityHistoryCreate) SaveX(ctx context.Context) *EntityHistory {
	v, err := ehc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ehc *EntityHistoryCreate) Exec(ctx context.Context) error {
	_, err := ehc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ehc *EntityHistoryCreate) ExecX(ctx context.Context) {
	if err := ehc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (ehc *EntityHistoryCreate) defaults() error {
	if _, ok := ehc.mutation.HistoryTime(); !ok {
		if entityhistory.DefaultHistoryTime == nil {
			return fmt.Errorf("generated: uninitialized entityhistory.DefaultHistoryTime (forgotten import generated/runtime?)")
		}
		v := entityhistory.DefaultHistoryTime()
		ehc.mutation.SetHistoryTime(v)
	}
	if _, ok := ehc.mutation.CreatedAt(); !ok {
		if entityhistory.DefaultCreatedAt == nil {
			return fmt.Errorf("generated: uninitialized entityhistory.DefaultCreatedAt (forgotten import generated/runtime?)")
		}
		v := entityhistory.DefaultCreatedAt()
		ehc.mutation.SetCreatedAt(v)
	}
	if _, ok := ehc.mutation.UpdatedAt(); !ok {
		if entityhistory.DefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized entityhistory.DefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := entityhistory.DefaultUpdatedAt()
		ehc.mutation.SetUpdatedAt(v)
	}
	if _, ok := ehc.mutation.Tags(); !ok {
		v := entityhistory.DefaultTags
		ehc.mutation.SetTags(v)
	}
	if _, ok := ehc.mutation.Status(); !ok {
		v := entityhistory.DefaultStatus
		ehc.mutation.SetStatus(v)
	}
	if _, ok := ehc.mutation.ID(); !ok {
		if entityhistory.DefaultID == nil {
			return fmt.Errorf("generated: uninitialized entityhistory.DefaultID (forgotten import generated/runtime?)")
		}
		v := entityhistory.DefaultID()
		ehc.mutation.SetID(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (ehc *EntityHistoryCreate) check() error {
	if _, ok := ehc.mutation.HistoryTime(); !ok {
		return &ValidationError{Name: "history_time", err: errors.New(`generated: missing required field "EntityHistory.history_time"`)}
	}
	if _, ok := ehc.mutation.Operation(); !ok {
		return &ValidationError{Name: "operation", err: errors.New(`generated: missing required field "EntityHistory.operation"`)}
	}
	if v, ok := ehc.mutation.Operation(); ok {
		if err := entityhistory.OperationValidator(v); err != nil {
			return &ValidationError{Name: "operation", err: fmt.Errorf(`generated: validator failed for field "EntityHistory.operation": %w`, err)}
		}
	}
	return nil
}

func (ehc *EntityHistoryCreate) sqlSave(ctx context.Context) (*EntityHistory, error) {
	if err := ehc.check(); err != nil {
		return nil, err
	}
	_node, _spec := ehc.createSpec()
	if err := sqlgraph.CreateNode(ctx, ehc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected EntityHistory.ID type: %T", _spec.ID.Value)
		}
	}
	ehc.mutation.id = &_node.ID
	ehc.mutation.done = true
	return _node, nil
}

func (ehc *EntityHistoryCreate) createSpec() (*EntityHistory, *sqlgraph.CreateSpec) {
	var (
		_node = &EntityHistory{config: ehc.config}
		_spec = sqlgraph.NewCreateSpec(entityhistory.Table, sqlgraph.NewFieldSpec(entityhistory.FieldID, field.TypeString))
	)
	_spec.Schema = ehc.schemaConfig.EntityHistory
	if id, ok := ehc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := ehc.mutation.HistoryTime(); ok {
		_spec.SetField(entityhistory.FieldHistoryTime, field.TypeTime, value)
		_node.HistoryTime = value
	}
	if value, ok := ehc.mutation.Ref(); ok {
		_spec.SetField(entityhistory.FieldRef, field.TypeString, value)
		_node.Ref = value
	}
	if value, ok := ehc.mutation.Operation(); ok {
		_spec.SetField(entityhistory.FieldOperation, field.TypeEnum, value)
		_node.Operation = value
	}
	if value, ok := ehc.mutation.CreatedAt(); ok {
		_spec.SetField(entityhistory.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := ehc.mutation.UpdatedAt(); ok {
		_spec.SetField(entityhistory.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := ehc.mutation.CreatedBy(); ok {
		_spec.SetField(entityhistory.FieldCreatedBy, field.TypeString, value)
		_node.CreatedBy = value
	}
	if value, ok := ehc.mutation.UpdatedBy(); ok {
		_spec.SetField(entityhistory.FieldUpdatedBy, field.TypeString, value)
		_node.UpdatedBy = value
	}
	if value, ok := ehc.mutation.DeletedAt(); ok {
		_spec.SetField(entityhistory.FieldDeletedAt, field.TypeTime, value)
		_node.DeletedAt = value
	}
	if value, ok := ehc.mutation.DeletedBy(); ok {
		_spec.SetField(entityhistory.FieldDeletedBy, field.TypeString, value)
		_node.DeletedBy = value
	}
	if value, ok := ehc.mutation.Tags(); ok {
		_spec.SetField(entityhistory.FieldTags, field.TypeJSON, value)
		_node.Tags = value
	}
	if value, ok := ehc.mutation.OwnerID(); ok {
		_spec.SetField(entityhistory.FieldOwnerID, field.TypeString, value)
		_node.OwnerID = value
	}
	if value, ok := ehc.mutation.Name(); ok {
		_spec.SetField(entityhistory.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := ehc.mutation.DisplayName(); ok {
		_spec.SetField(entityhistory.FieldDisplayName, field.TypeString, value)
		_node.DisplayName = value
	}
	if value, ok := ehc.mutation.Description(); ok {
		_spec.SetField(entityhistory.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	if value, ok := ehc.mutation.Domains(); ok {
		_spec.SetField(entityhistory.FieldDomains, field.TypeJSON, value)
		_node.Domains = value
	}
	if value, ok := ehc.mutation.EntityTypeID(); ok {
		_spec.SetField(entityhistory.FieldEntityTypeID, field.TypeString, value)
		_node.EntityTypeID = value
	}
	if value, ok := ehc.mutation.Status(); ok {
		_spec.SetField(entityhistory.FieldStatus, field.TypeString, value)
		_node.Status = value
	}
	return _node, _spec
}

// EntityHistoryCreateBulk is the builder for creating many EntityHistory entities in bulk.
type EntityHistoryCreateBulk struct {
	config
	err      error
	builders []*EntityHistoryCreate
}

// Save creates the EntityHistory entities in the database.
func (ehcb *EntityHistoryCreateBulk) Save(ctx context.Context) ([]*EntityHistory, error) {
	if ehcb.err != nil {
		return nil, ehcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(ehcb.builders))
	nodes := make([]*EntityHistory, len(ehcb.builders))
	mutators := make([]Mutator, len(ehcb.builders))
	for i := range ehcb.builders {
		func(i int, root context.Context) {
			builder := ehcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*EntityHistoryMutation)
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
					_, err = mutators[i+1].Mutate(root, ehcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ehcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, ehcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ehcb *EntityHistoryCreateBulk) SaveX(ctx context.Context) []*EntityHistory {
	v, err := ehcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ehcb *EntityHistoryCreateBulk) Exec(ctx context.Context) error {
	_, err := ehcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ehcb *EntityHistoryCreateBulk) ExecX(ctx context.Context) {
	if err := ehcb.Exec(ctx); err != nil {
		panic(err)
	}
}
