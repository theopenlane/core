// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/entity"
	"github.com/theopenlane/core/internal/ent/generated/note"
	"github.com/theopenlane/core/internal/ent/generated/organization"
)

// NoteCreate is the builder for creating a Note entity.
type NoteCreate struct {
	config
	mutation *NoteMutation
	hooks    []Hook
}

// SetCreatedAt sets the "created_at" field.
func (nc *NoteCreate) SetCreatedAt(t time.Time) *NoteCreate {
	nc.mutation.SetCreatedAt(t)
	return nc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (nc *NoteCreate) SetNillableCreatedAt(t *time.Time) *NoteCreate {
	if t != nil {
		nc.SetCreatedAt(*t)
	}
	return nc
}

// SetUpdatedAt sets the "updated_at" field.
func (nc *NoteCreate) SetUpdatedAt(t time.Time) *NoteCreate {
	nc.mutation.SetUpdatedAt(t)
	return nc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (nc *NoteCreate) SetNillableUpdatedAt(t *time.Time) *NoteCreate {
	if t != nil {
		nc.SetUpdatedAt(*t)
	}
	return nc
}

// SetCreatedBy sets the "created_by" field.
func (nc *NoteCreate) SetCreatedBy(s string) *NoteCreate {
	nc.mutation.SetCreatedBy(s)
	return nc
}

// SetNillableCreatedBy sets the "created_by" field if the given value is not nil.
func (nc *NoteCreate) SetNillableCreatedBy(s *string) *NoteCreate {
	if s != nil {
		nc.SetCreatedBy(*s)
	}
	return nc
}

// SetUpdatedBy sets the "updated_by" field.
func (nc *NoteCreate) SetUpdatedBy(s string) *NoteCreate {
	nc.mutation.SetUpdatedBy(s)
	return nc
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (nc *NoteCreate) SetNillableUpdatedBy(s *string) *NoteCreate {
	if s != nil {
		nc.SetUpdatedBy(*s)
	}
	return nc
}

// SetMappingID sets the "mapping_id" field.
func (nc *NoteCreate) SetMappingID(s string) *NoteCreate {
	nc.mutation.SetMappingID(s)
	return nc
}

// SetNillableMappingID sets the "mapping_id" field if the given value is not nil.
func (nc *NoteCreate) SetNillableMappingID(s *string) *NoteCreate {
	if s != nil {
		nc.SetMappingID(*s)
	}
	return nc
}

// SetDeletedAt sets the "deleted_at" field.
func (nc *NoteCreate) SetDeletedAt(t time.Time) *NoteCreate {
	nc.mutation.SetDeletedAt(t)
	return nc
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (nc *NoteCreate) SetNillableDeletedAt(t *time.Time) *NoteCreate {
	if t != nil {
		nc.SetDeletedAt(*t)
	}
	return nc
}

// SetDeletedBy sets the "deleted_by" field.
func (nc *NoteCreate) SetDeletedBy(s string) *NoteCreate {
	nc.mutation.SetDeletedBy(s)
	return nc
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (nc *NoteCreate) SetNillableDeletedBy(s *string) *NoteCreate {
	if s != nil {
		nc.SetDeletedBy(*s)
	}
	return nc
}

// SetTags sets the "tags" field.
func (nc *NoteCreate) SetTags(s []string) *NoteCreate {
	nc.mutation.SetTags(s)
	return nc
}

// SetOwnerID sets the "owner_id" field.
func (nc *NoteCreate) SetOwnerID(s string) *NoteCreate {
	nc.mutation.SetOwnerID(s)
	return nc
}

// SetNillableOwnerID sets the "owner_id" field if the given value is not nil.
func (nc *NoteCreate) SetNillableOwnerID(s *string) *NoteCreate {
	if s != nil {
		nc.SetOwnerID(*s)
	}
	return nc
}

// SetText sets the "text" field.
func (nc *NoteCreate) SetText(s string) *NoteCreate {
	nc.mutation.SetText(s)
	return nc
}

// SetID sets the "id" field.
func (nc *NoteCreate) SetID(s string) *NoteCreate {
	nc.mutation.SetID(s)
	return nc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (nc *NoteCreate) SetNillableID(s *string) *NoteCreate {
	if s != nil {
		nc.SetID(*s)
	}
	return nc
}

// SetOwner sets the "owner" edge to the Organization entity.
func (nc *NoteCreate) SetOwner(o *Organization) *NoteCreate {
	return nc.SetOwnerID(o.ID)
}

// SetEntityID sets the "entity" edge to the Entity entity by ID.
func (nc *NoteCreate) SetEntityID(id string) *NoteCreate {
	nc.mutation.SetEntityID(id)
	return nc
}

// SetNillableEntityID sets the "entity" edge to the Entity entity by ID if the given value is not nil.
func (nc *NoteCreate) SetNillableEntityID(id *string) *NoteCreate {
	if id != nil {
		nc = nc.SetEntityID(*id)
	}
	return nc
}

// SetEntity sets the "entity" edge to the Entity entity.
func (nc *NoteCreate) SetEntity(e *Entity) *NoteCreate {
	return nc.SetEntityID(e.ID)
}

// Mutation returns the NoteMutation object of the builder.
func (nc *NoteCreate) Mutation() *NoteMutation {
	return nc.mutation
}

// Save creates the Note in the database.
func (nc *NoteCreate) Save(ctx context.Context) (*Note, error) {
	if err := nc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, nc.sqlSave, nc.mutation, nc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (nc *NoteCreate) SaveX(ctx context.Context) *Note {
	v, err := nc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (nc *NoteCreate) Exec(ctx context.Context) error {
	_, err := nc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (nc *NoteCreate) ExecX(ctx context.Context) {
	if err := nc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (nc *NoteCreate) defaults() error {
	if _, ok := nc.mutation.CreatedAt(); !ok {
		if note.DefaultCreatedAt == nil {
			return fmt.Errorf("generated: uninitialized note.DefaultCreatedAt (forgotten import generated/runtime?)")
		}
		v := note.DefaultCreatedAt()
		nc.mutation.SetCreatedAt(v)
	}
	if _, ok := nc.mutation.UpdatedAt(); !ok {
		if note.DefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized note.DefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := note.DefaultUpdatedAt()
		nc.mutation.SetUpdatedAt(v)
	}
	if _, ok := nc.mutation.MappingID(); !ok {
		if note.DefaultMappingID == nil {
			return fmt.Errorf("generated: uninitialized note.DefaultMappingID (forgotten import generated/runtime?)")
		}
		v := note.DefaultMappingID()
		nc.mutation.SetMappingID(v)
	}
	if _, ok := nc.mutation.Tags(); !ok {
		v := note.DefaultTags
		nc.mutation.SetTags(v)
	}
	if _, ok := nc.mutation.ID(); !ok {
		if note.DefaultID == nil {
			return fmt.Errorf("generated: uninitialized note.DefaultID (forgotten import generated/runtime?)")
		}
		v := note.DefaultID()
		nc.mutation.SetID(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (nc *NoteCreate) check() error {
	if _, ok := nc.mutation.MappingID(); !ok {
		return &ValidationError{Name: "mapping_id", err: errors.New(`generated: missing required field "Note.mapping_id"`)}
	}
	if v, ok := nc.mutation.OwnerID(); ok {
		if err := note.OwnerIDValidator(v); err != nil {
			return &ValidationError{Name: "owner_id", err: fmt.Errorf(`generated: validator failed for field "Note.owner_id": %w`, err)}
		}
	}
	if _, ok := nc.mutation.Text(); !ok {
		return &ValidationError{Name: "text", err: errors.New(`generated: missing required field "Note.text"`)}
	}
	if v, ok := nc.mutation.Text(); ok {
		if err := note.TextValidator(v); err != nil {
			return &ValidationError{Name: "text", err: fmt.Errorf(`generated: validator failed for field "Note.text": %w`, err)}
		}
	}
	return nil
}

func (nc *NoteCreate) sqlSave(ctx context.Context) (*Note, error) {
	if err := nc.check(); err != nil {
		return nil, err
	}
	_node, _spec := nc.createSpec()
	if err := sqlgraph.CreateNode(ctx, nc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected Note.ID type: %T", _spec.ID.Value)
		}
	}
	nc.mutation.id = &_node.ID
	nc.mutation.done = true
	return _node, nil
}

func (nc *NoteCreate) createSpec() (*Note, *sqlgraph.CreateSpec) {
	var (
		_node = &Note{config: nc.config}
		_spec = sqlgraph.NewCreateSpec(note.Table, sqlgraph.NewFieldSpec(note.FieldID, field.TypeString))
	)
	_spec.Schema = nc.schemaConfig.Note
	if id, ok := nc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := nc.mutation.CreatedAt(); ok {
		_spec.SetField(note.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := nc.mutation.UpdatedAt(); ok {
		_spec.SetField(note.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := nc.mutation.CreatedBy(); ok {
		_spec.SetField(note.FieldCreatedBy, field.TypeString, value)
		_node.CreatedBy = value
	}
	if value, ok := nc.mutation.UpdatedBy(); ok {
		_spec.SetField(note.FieldUpdatedBy, field.TypeString, value)
		_node.UpdatedBy = value
	}
	if value, ok := nc.mutation.MappingID(); ok {
		_spec.SetField(note.FieldMappingID, field.TypeString, value)
		_node.MappingID = value
	}
	if value, ok := nc.mutation.DeletedAt(); ok {
		_spec.SetField(note.FieldDeletedAt, field.TypeTime, value)
		_node.DeletedAt = value
	}
	if value, ok := nc.mutation.DeletedBy(); ok {
		_spec.SetField(note.FieldDeletedBy, field.TypeString, value)
		_node.DeletedBy = value
	}
	if value, ok := nc.mutation.Tags(); ok {
		_spec.SetField(note.FieldTags, field.TypeJSON, value)
		_node.Tags = value
	}
	if value, ok := nc.mutation.Text(); ok {
		_spec.SetField(note.FieldText, field.TypeString, value)
		_node.Text = value
	}
	if nodes := nc.mutation.OwnerIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   note.OwnerTable,
			Columns: []string{note.OwnerColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(organization.FieldID, field.TypeString),
			},
		}
		edge.Schema = nc.schemaConfig.Note
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.OwnerID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := nc.mutation.EntityIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   note.EntityTable,
			Columns: []string{note.EntityColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(entity.FieldID, field.TypeString),
			},
		}
		edge.Schema = nc.schemaConfig.Note
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.entity_notes = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// NoteCreateBulk is the builder for creating many Note entities in bulk.
type NoteCreateBulk struct {
	config
	err      error
	builders []*NoteCreate
}

// Save creates the Note entities in the database.
func (ncb *NoteCreateBulk) Save(ctx context.Context) ([]*Note, error) {
	if ncb.err != nil {
		return nil, ncb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(ncb.builders))
	nodes := make([]*Note, len(ncb.builders))
	mutators := make([]Mutator, len(ncb.builders))
	for i := range ncb.builders {
		func(i int, root context.Context) {
			builder := ncb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*NoteMutation)
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
					_, err = mutators[i+1].Mutate(root, ncb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ncb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, ncb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ncb *NoteCreateBulk) SaveX(ctx context.Context) []*Note {
	v, err := ncb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ncb *NoteCreateBulk) Exec(ctx context.Context) error {
	_, err := ncb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ncb *NoteCreateBulk) ExecX(ctx context.Context) {
	if err := ncb.Exec(ctx); err != nil {
		panic(err)
	}
}