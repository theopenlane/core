// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/assethistory"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/entx/history"
)

// AssetHistoryCreate is the builder for creating a AssetHistory entity.
type AssetHistoryCreate struct {
	config
	mutation *AssetHistoryMutation
	hooks    []Hook
}

// SetHistoryTime sets the "history_time" field.
func (ahc *AssetHistoryCreate) SetHistoryTime(t time.Time) *AssetHistoryCreate {
	ahc.mutation.SetHistoryTime(t)
	return ahc
}

// SetNillableHistoryTime sets the "history_time" field if the given value is not nil.
func (ahc *AssetHistoryCreate) SetNillableHistoryTime(t *time.Time) *AssetHistoryCreate {
	if t != nil {
		ahc.SetHistoryTime(*t)
	}
	return ahc
}

// SetRef sets the "ref" field.
func (ahc *AssetHistoryCreate) SetRef(s string) *AssetHistoryCreate {
	ahc.mutation.SetRef(s)
	return ahc
}

// SetNillableRef sets the "ref" field if the given value is not nil.
func (ahc *AssetHistoryCreate) SetNillableRef(s *string) *AssetHistoryCreate {
	if s != nil {
		ahc.SetRef(*s)
	}
	return ahc
}

// SetOperation sets the "operation" field.
func (ahc *AssetHistoryCreate) SetOperation(ht history.OpType) *AssetHistoryCreate {
	ahc.mutation.SetOperation(ht)
	return ahc
}

// SetCreatedAt sets the "created_at" field.
func (ahc *AssetHistoryCreate) SetCreatedAt(t time.Time) *AssetHistoryCreate {
	ahc.mutation.SetCreatedAt(t)
	return ahc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (ahc *AssetHistoryCreate) SetNillableCreatedAt(t *time.Time) *AssetHistoryCreate {
	if t != nil {
		ahc.SetCreatedAt(*t)
	}
	return ahc
}

// SetUpdatedAt sets the "updated_at" field.
func (ahc *AssetHistoryCreate) SetUpdatedAt(t time.Time) *AssetHistoryCreate {
	ahc.mutation.SetUpdatedAt(t)
	return ahc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (ahc *AssetHistoryCreate) SetNillableUpdatedAt(t *time.Time) *AssetHistoryCreate {
	if t != nil {
		ahc.SetUpdatedAt(*t)
	}
	return ahc
}

// SetCreatedBy sets the "created_by" field.
func (ahc *AssetHistoryCreate) SetCreatedBy(s string) *AssetHistoryCreate {
	ahc.mutation.SetCreatedBy(s)
	return ahc
}

// SetNillableCreatedBy sets the "created_by" field if the given value is not nil.
func (ahc *AssetHistoryCreate) SetNillableCreatedBy(s *string) *AssetHistoryCreate {
	if s != nil {
		ahc.SetCreatedBy(*s)
	}
	return ahc
}

// SetUpdatedBy sets the "updated_by" field.
func (ahc *AssetHistoryCreate) SetUpdatedBy(s string) *AssetHistoryCreate {
	ahc.mutation.SetUpdatedBy(s)
	return ahc
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (ahc *AssetHistoryCreate) SetNillableUpdatedBy(s *string) *AssetHistoryCreate {
	if s != nil {
		ahc.SetUpdatedBy(*s)
	}
	return ahc
}

// SetDeletedAt sets the "deleted_at" field.
func (ahc *AssetHistoryCreate) SetDeletedAt(t time.Time) *AssetHistoryCreate {
	ahc.mutation.SetDeletedAt(t)
	return ahc
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (ahc *AssetHistoryCreate) SetNillableDeletedAt(t *time.Time) *AssetHistoryCreate {
	if t != nil {
		ahc.SetDeletedAt(*t)
	}
	return ahc
}

// SetDeletedBy sets the "deleted_by" field.
func (ahc *AssetHistoryCreate) SetDeletedBy(s string) *AssetHistoryCreate {
	ahc.mutation.SetDeletedBy(s)
	return ahc
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (ahc *AssetHistoryCreate) SetNillableDeletedBy(s *string) *AssetHistoryCreate {
	if s != nil {
		ahc.SetDeletedBy(*s)
	}
	return ahc
}

// SetTags sets the "tags" field.
func (ahc *AssetHistoryCreate) SetTags(s []string) *AssetHistoryCreate {
	ahc.mutation.SetTags(s)
	return ahc
}

// SetOwnerID sets the "owner_id" field.
func (ahc *AssetHistoryCreate) SetOwnerID(s string) *AssetHistoryCreate {
	ahc.mutation.SetOwnerID(s)
	return ahc
}

// SetNillableOwnerID sets the "owner_id" field if the given value is not nil.
func (ahc *AssetHistoryCreate) SetNillableOwnerID(s *string) *AssetHistoryCreate {
	if s != nil {
		ahc.SetOwnerID(*s)
	}
	return ahc
}

// SetAssetType sets the "asset_type" field.
func (ahc *AssetHistoryCreate) SetAssetType(et enums.AssetType) *AssetHistoryCreate {
	ahc.mutation.SetAssetType(et)
	return ahc
}

// SetNillableAssetType sets the "asset_type" field if the given value is not nil.
func (ahc *AssetHistoryCreate) SetNillableAssetType(et *enums.AssetType) *AssetHistoryCreate {
	if et != nil {
		ahc.SetAssetType(*et)
	}
	return ahc
}

// SetName sets the "name" field.
func (ahc *AssetHistoryCreate) SetName(s string) *AssetHistoryCreate {
	ahc.mutation.SetName(s)
	return ahc
}

// SetDescription sets the "description" field.
func (ahc *AssetHistoryCreate) SetDescription(s string) *AssetHistoryCreate {
	ahc.mutation.SetDescription(s)
	return ahc
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (ahc *AssetHistoryCreate) SetNillableDescription(s *string) *AssetHistoryCreate {
	if s != nil {
		ahc.SetDescription(*s)
	}
	return ahc
}

// SetIdentifier sets the "identifier" field.
func (ahc *AssetHistoryCreate) SetIdentifier(s string) *AssetHistoryCreate {
	ahc.mutation.SetIdentifier(s)
	return ahc
}

// SetNillableIdentifier sets the "identifier" field if the given value is not nil.
func (ahc *AssetHistoryCreate) SetNillableIdentifier(s *string) *AssetHistoryCreate {
	if s != nil {
		ahc.SetIdentifier(*s)
	}
	return ahc
}

// SetWebsite sets the "website" field.
func (ahc *AssetHistoryCreate) SetWebsite(s string) *AssetHistoryCreate {
	ahc.mutation.SetWebsite(s)
	return ahc
}

// SetNillableWebsite sets the "website" field if the given value is not nil.
func (ahc *AssetHistoryCreate) SetNillableWebsite(s *string) *AssetHistoryCreate {
	if s != nil {
		ahc.SetWebsite(*s)
	}
	return ahc
}

// SetCpe sets the "cpe" field.
func (ahc *AssetHistoryCreate) SetCpe(s string) *AssetHistoryCreate {
	ahc.mutation.SetCpe(s)
	return ahc
}

// SetNillableCpe sets the "cpe" field if the given value is not nil.
func (ahc *AssetHistoryCreate) SetNillableCpe(s *string) *AssetHistoryCreate {
	if s != nil {
		ahc.SetCpe(*s)
	}
	return ahc
}

// SetCategories sets the "categories" field.
func (ahc *AssetHistoryCreate) SetCategories(s []string) *AssetHistoryCreate {
	ahc.mutation.SetCategories(s)
	return ahc
}

// SetID sets the "id" field.
func (ahc *AssetHistoryCreate) SetID(s string) *AssetHistoryCreate {
	ahc.mutation.SetID(s)
	return ahc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (ahc *AssetHistoryCreate) SetNillableID(s *string) *AssetHistoryCreate {
	if s != nil {
		ahc.SetID(*s)
	}
	return ahc
}

// Mutation returns the AssetHistoryMutation object of the builder.
func (ahc *AssetHistoryCreate) Mutation() *AssetHistoryMutation {
	return ahc.mutation
}

// Save creates the AssetHistory in the database.
func (ahc *AssetHistoryCreate) Save(ctx context.Context) (*AssetHistory, error) {
	if err := ahc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, ahc.sqlSave, ahc.mutation, ahc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (ahc *AssetHistoryCreate) SaveX(ctx context.Context) *AssetHistory {
	v, err := ahc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ahc *AssetHistoryCreate) Exec(ctx context.Context) error {
	_, err := ahc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ahc *AssetHistoryCreate) ExecX(ctx context.Context) {
	if err := ahc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (ahc *AssetHistoryCreate) defaults() error {
	if _, ok := ahc.mutation.HistoryTime(); !ok {
		if assethistory.DefaultHistoryTime == nil {
			return fmt.Errorf("generated: uninitialized assethistory.DefaultHistoryTime (forgotten import generated/runtime?)")
		}
		v := assethistory.DefaultHistoryTime()
		ahc.mutation.SetHistoryTime(v)
	}
	if _, ok := ahc.mutation.CreatedAt(); !ok {
		if assethistory.DefaultCreatedAt == nil {
			return fmt.Errorf("generated: uninitialized assethistory.DefaultCreatedAt (forgotten import generated/runtime?)")
		}
		v := assethistory.DefaultCreatedAt()
		ahc.mutation.SetCreatedAt(v)
	}
	if _, ok := ahc.mutation.UpdatedAt(); !ok {
		if assethistory.DefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized assethistory.DefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := assethistory.DefaultUpdatedAt()
		ahc.mutation.SetUpdatedAt(v)
	}
	if _, ok := ahc.mutation.Tags(); !ok {
		v := assethistory.DefaultTags
		ahc.mutation.SetTags(v)
	}
	if _, ok := ahc.mutation.AssetType(); !ok {
		v := assethistory.DefaultAssetType
		ahc.mutation.SetAssetType(v)
	}
	if _, ok := ahc.mutation.ID(); !ok {
		if assethistory.DefaultID == nil {
			return fmt.Errorf("generated: uninitialized assethistory.DefaultID (forgotten import generated/runtime?)")
		}
		v := assethistory.DefaultID()
		ahc.mutation.SetID(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (ahc *AssetHistoryCreate) check() error {
	if _, ok := ahc.mutation.HistoryTime(); !ok {
		return &ValidationError{Name: "history_time", err: errors.New(`generated: missing required field "AssetHistory.history_time"`)}
	}
	if _, ok := ahc.mutation.Operation(); !ok {
		return &ValidationError{Name: "operation", err: errors.New(`generated: missing required field "AssetHistory.operation"`)}
	}
	if v, ok := ahc.mutation.Operation(); ok {
		if err := assethistory.OperationValidator(v); err != nil {
			return &ValidationError{Name: "operation", err: fmt.Errorf(`generated: validator failed for field "AssetHistory.operation": %w`, err)}
		}
	}
	if _, ok := ahc.mutation.AssetType(); !ok {
		return &ValidationError{Name: "asset_type", err: errors.New(`generated: missing required field "AssetHistory.asset_type"`)}
	}
	if v, ok := ahc.mutation.AssetType(); ok {
		if err := assethistory.AssetTypeValidator(v); err != nil {
			return &ValidationError{Name: "asset_type", err: fmt.Errorf(`generated: validator failed for field "AssetHistory.asset_type": %w`, err)}
		}
	}
	if _, ok := ahc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`generated: missing required field "AssetHistory.name"`)}
	}
	return nil
}

func (ahc *AssetHistoryCreate) sqlSave(ctx context.Context) (*AssetHistory, error) {
	if err := ahc.check(); err != nil {
		return nil, err
	}
	_node, _spec := ahc.createSpec()
	if err := sqlgraph.CreateNode(ctx, ahc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected AssetHistory.ID type: %T", _spec.ID.Value)
		}
	}
	ahc.mutation.id = &_node.ID
	ahc.mutation.done = true
	return _node, nil
}

func (ahc *AssetHistoryCreate) createSpec() (*AssetHistory, *sqlgraph.CreateSpec) {
	var (
		_node = &AssetHistory{config: ahc.config}
		_spec = sqlgraph.NewCreateSpec(assethistory.Table, sqlgraph.NewFieldSpec(assethistory.FieldID, field.TypeString))
	)
	_spec.Schema = ahc.schemaConfig.AssetHistory
	if id, ok := ahc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := ahc.mutation.HistoryTime(); ok {
		_spec.SetField(assethistory.FieldHistoryTime, field.TypeTime, value)
		_node.HistoryTime = value
	}
	if value, ok := ahc.mutation.Ref(); ok {
		_spec.SetField(assethistory.FieldRef, field.TypeString, value)
		_node.Ref = value
	}
	if value, ok := ahc.mutation.Operation(); ok {
		_spec.SetField(assethistory.FieldOperation, field.TypeEnum, value)
		_node.Operation = value
	}
	if value, ok := ahc.mutation.CreatedAt(); ok {
		_spec.SetField(assethistory.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := ahc.mutation.UpdatedAt(); ok {
		_spec.SetField(assethistory.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := ahc.mutation.CreatedBy(); ok {
		_spec.SetField(assethistory.FieldCreatedBy, field.TypeString, value)
		_node.CreatedBy = value
	}
	if value, ok := ahc.mutation.UpdatedBy(); ok {
		_spec.SetField(assethistory.FieldUpdatedBy, field.TypeString, value)
		_node.UpdatedBy = value
	}
	if value, ok := ahc.mutation.DeletedAt(); ok {
		_spec.SetField(assethistory.FieldDeletedAt, field.TypeTime, value)
		_node.DeletedAt = value
	}
	if value, ok := ahc.mutation.DeletedBy(); ok {
		_spec.SetField(assethistory.FieldDeletedBy, field.TypeString, value)
		_node.DeletedBy = value
	}
	if value, ok := ahc.mutation.Tags(); ok {
		_spec.SetField(assethistory.FieldTags, field.TypeJSON, value)
		_node.Tags = value
	}
	if value, ok := ahc.mutation.OwnerID(); ok {
		_spec.SetField(assethistory.FieldOwnerID, field.TypeString, value)
		_node.OwnerID = value
	}
	if value, ok := ahc.mutation.AssetType(); ok {
		_spec.SetField(assethistory.FieldAssetType, field.TypeEnum, value)
		_node.AssetType = value
	}
	if value, ok := ahc.mutation.Name(); ok {
		_spec.SetField(assethistory.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := ahc.mutation.Description(); ok {
		_spec.SetField(assethistory.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	if value, ok := ahc.mutation.Identifier(); ok {
		_spec.SetField(assethistory.FieldIdentifier, field.TypeString, value)
		_node.Identifier = value
	}
	if value, ok := ahc.mutation.Website(); ok {
		_spec.SetField(assethistory.FieldWebsite, field.TypeString, value)
		_node.Website = value
	}
	if value, ok := ahc.mutation.Cpe(); ok {
		_spec.SetField(assethistory.FieldCpe, field.TypeString, value)
		_node.Cpe = value
	}
	if value, ok := ahc.mutation.Categories(); ok {
		_spec.SetField(assethistory.FieldCategories, field.TypeJSON, value)
		_node.Categories = value
	}
	return _node, _spec
}

// AssetHistoryCreateBulk is the builder for creating many AssetHistory entities in bulk.
type AssetHistoryCreateBulk struct {
	config
	err      error
	builders []*AssetHistoryCreate
}

// Save creates the AssetHistory entities in the database.
func (ahcb *AssetHistoryCreateBulk) Save(ctx context.Context) ([]*AssetHistory, error) {
	if ahcb.err != nil {
		return nil, ahcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(ahcb.builders))
	nodes := make([]*AssetHistory, len(ahcb.builders))
	mutators := make([]Mutator, len(ahcb.builders))
	for i := range ahcb.builders {
		func(i int, root context.Context) {
			builder := ahcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*AssetHistoryMutation)
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
					_, err = mutators[i+1].Mutate(root, ahcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ahcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, ahcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ahcb *AssetHistoryCreateBulk) SaveX(ctx context.Context) []*AssetHistory {
	v, err := ahcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ahcb *AssetHistoryCreateBulk) Exec(ctx context.Context) error {
	_, err := ahcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ahcb *AssetHistoryCreateBulk) ExecX(ctx context.Context) {
	if err := ahcb.Exec(ctx); err != nil {
		panic(err)
	}
}
