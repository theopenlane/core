// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/apitoken"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/pkg/models"
)

// APITokenCreate is the builder for creating a APIToken entity.
type APITokenCreate struct {
	config
	mutation *APITokenMutation
	hooks    []Hook
}

// SetCreatedAt sets the "created_at" field.
func (atc *APITokenCreate) SetCreatedAt(t time.Time) *APITokenCreate {
	atc.mutation.SetCreatedAt(t)
	return atc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (atc *APITokenCreate) SetNillableCreatedAt(t *time.Time) *APITokenCreate {
	if t != nil {
		atc.SetCreatedAt(*t)
	}
	return atc
}

// SetUpdatedAt sets the "updated_at" field.
func (atc *APITokenCreate) SetUpdatedAt(t time.Time) *APITokenCreate {
	atc.mutation.SetUpdatedAt(t)
	return atc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (atc *APITokenCreate) SetNillableUpdatedAt(t *time.Time) *APITokenCreate {
	if t != nil {
		atc.SetUpdatedAt(*t)
	}
	return atc
}

// SetCreatedBy sets the "created_by" field.
func (atc *APITokenCreate) SetCreatedBy(s string) *APITokenCreate {
	atc.mutation.SetCreatedBy(s)
	return atc
}

// SetNillableCreatedBy sets the "created_by" field if the given value is not nil.
func (atc *APITokenCreate) SetNillableCreatedBy(s *string) *APITokenCreate {
	if s != nil {
		atc.SetCreatedBy(*s)
	}
	return atc
}

// SetUpdatedBy sets the "updated_by" field.
func (atc *APITokenCreate) SetUpdatedBy(s string) *APITokenCreate {
	atc.mutation.SetUpdatedBy(s)
	return atc
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (atc *APITokenCreate) SetNillableUpdatedBy(s *string) *APITokenCreate {
	if s != nil {
		atc.SetUpdatedBy(*s)
	}
	return atc
}

// SetDeletedAt sets the "deleted_at" field.
func (atc *APITokenCreate) SetDeletedAt(t time.Time) *APITokenCreate {
	atc.mutation.SetDeletedAt(t)
	return atc
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (atc *APITokenCreate) SetNillableDeletedAt(t *time.Time) *APITokenCreate {
	if t != nil {
		atc.SetDeletedAt(*t)
	}
	return atc
}

// SetDeletedBy sets the "deleted_by" field.
func (atc *APITokenCreate) SetDeletedBy(s string) *APITokenCreate {
	atc.mutation.SetDeletedBy(s)
	return atc
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (atc *APITokenCreate) SetNillableDeletedBy(s *string) *APITokenCreate {
	if s != nil {
		atc.SetDeletedBy(*s)
	}
	return atc
}

// SetTags sets the "tags" field.
func (atc *APITokenCreate) SetTags(s []string) *APITokenCreate {
	atc.mutation.SetTags(s)
	return atc
}

// SetOwnerID sets the "owner_id" field.
func (atc *APITokenCreate) SetOwnerID(s string) *APITokenCreate {
	atc.mutation.SetOwnerID(s)
	return atc
}

// SetNillableOwnerID sets the "owner_id" field if the given value is not nil.
func (atc *APITokenCreate) SetNillableOwnerID(s *string) *APITokenCreate {
	if s != nil {
		atc.SetOwnerID(*s)
	}
	return atc
}

// SetName sets the "name" field.
func (atc *APITokenCreate) SetName(s string) *APITokenCreate {
	atc.mutation.SetName(s)
	return atc
}

// SetToken sets the "token" field.
func (atc *APITokenCreate) SetToken(s string) *APITokenCreate {
	atc.mutation.SetToken(s)
	return atc
}

// SetNillableToken sets the "token" field if the given value is not nil.
func (atc *APITokenCreate) SetNillableToken(s *string) *APITokenCreate {
	if s != nil {
		atc.SetToken(*s)
	}
	return atc
}

// SetExpiresAt sets the "expires_at" field.
func (atc *APITokenCreate) SetExpiresAt(t time.Time) *APITokenCreate {
	atc.mutation.SetExpiresAt(t)
	return atc
}

// SetNillableExpiresAt sets the "expires_at" field if the given value is not nil.
func (atc *APITokenCreate) SetNillableExpiresAt(t *time.Time) *APITokenCreate {
	if t != nil {
		atc.SetExpiresAt(*t)
	}
	return atc
}

// SetDescription sets the "description" field.
func (atc *APITokenCreate) SetDescription(s string) *APITokenCreate {
	atc.mutation.SetDescription(s)
	return atc
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (atc *APITokenCreate) SetNillableDescription(s *string) *APITokenCreate {
	if s != nil {
		atc.SetDescription(*s)
	}
	return atc
}

// SetScopes sets the "scopes" field.
func (atc *APITokenCreate) SetScopes(s []string) *APITokenCreate {
	atc.mutation.SetScopes(s)
	return atc
}

// SetLastUsedAt sets the "last_used_at" field.
func (atc *APITokenCreate) SetLastUsedAt(t time.Time) *APITokenCreate {
	atc.mutation.SetLastUsedAt(t)
	return atc
}

// SetNillableLastUsedAt sets the "last_used_at" field if the given value is not nil.
func (atc *APITokenCreate) SetNillableLastUsedAt(t *time.Time) *APITokenCreate {
	if t != nil {
		atc.SetLastUsedAt(*t)
	}
	return atc
}

// SetIsActive sets the "is_active" field.
func (atc *APITokenCreate) SetIsActive(b bool) *APITokenCreate {
	atc.mutation.SetIsActive(b)
	return atc
}

// SetNillableIsActive sets the "is_active" field if the given value is not nil.
func (atc *APITokenCreate) SetNillableIsActive(b *bool) *APITokenCreate {
	if b != nil {
		atc.SetIsActive(*b)
	}
	return atc
}

// SetRevokedReason sets the "revoked_reason" field.
func (atc *APITokenCreate) SetRevokedReason(s string) *APITokenCreate {
	atc.mutation.SetRevokedReason(s)
	return atc
}

// SetNillableRevokedReason sets the "revoked_reason" field if the given value is not nil.
func (atc *APITokenCreate) SetNillableRevokedReason(s *string) *APITokenCreate {
	if s != nil {
		atc.SetRevokedReason(*s)
	}
	return atc
}

// SetRevokedBy sets the "revoked_by" field.
func (atc *APITokenCreate) SetRevokedBy(s string) *APITokenCreate {
	atc.mutation.SetRevokedBy(s)
	return atc
}

// SetNillableRevokedBy sets the "revoked_by" field if the given value is not nil.
func (atc *APITokenCreate) SetNillableRevokedBy(s *string) *APITokenCreate {
	if s != nil {
		atc.SetRevokedBy(*s)
	}
	return atc
}

// SetRevokedAt sets the "revoked_at" field.
func (atc *APITokenCreate) SetRevokedAt(t time.Time) *APITokenCreate {
	atc.mutation.SetRevokedAt(t)
	return atc
}

// SetNillableRevokedAt sets the "revoked_at" field if the given value is not nil.
func (atc *APITokenCreate) SetNillableRevokedAt(t *time.Time) *APITokenCreate {
	if t != nil {
		atc.SetRevokedAt(*t)
	}
	return atc
}

// SetSSOAuthorizations sets the "sso_authorizations" field.
func (atc *APITokenCreate) SetSSOAuthorizations(mam models.SSOAuthorizationMap) *APITokenCreate {
	atc.mutation.SetSSOAuthorizations(mam)
	return atc
}

// SetID sets the "id" field.
func (atc *APITokenCreate) SetID(s string) *APITokenCreate {
	atc.mutation.SetID(s)
	return atc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (atc *APITokenCreate) SetNillableID(s *string) *APITokenCreate {
	if s != nil {
		atc.SetID(*s)
	}
	return atc
}

// SetOwner sets the "owner" edge to the Organization entity.
func (atc *APITokenCreate) SetOwner(o *Organization) *APITokenCreate {
	return atc.SetOwnerID(o.ID)
}

// Mutation returns the APITokenMutation object of the builder.
func (atc *APITokenCreate) Mutation() *APITokenMutation {
	return atc.mutation
}

// Save creates the APIToken in the database.
func (atc *APITokenCreate) Save(ctx context.Context) (*APIToken, error) {
	if err := atc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, atc.sqlSave, atc.mutation, atc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (atc *APITokenCreate) SaveX(ctx context.Context) *APIToken {
	v, err := atc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (atc *APITokenCreate) Exec(ctx context.Context) error {
	_, err := atc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (atc *APITokenCreate) ExecX(ctx context.Context) {
	if err := atc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (atc *APITokenCreate) defaults() error {
	if _, ok := atc.mutation.CreatedAt(); !ok {
		if apitoken.DefaultCreatedAt == nil {
			return fmt.Errorf("generated: uninitialized apitoken.DefaultCreatedAt (forgotten import generated/runtime?)")
		}
		v := apitoken.DefaultCreatedAt()
		atc.mutation.SetCreatedAt(v)
	}
	if _, ok := atc.mutation.UpdatedAt(); !ok {
		if apitoken.DefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized apitoken.DefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := apitoken.DefaultUpdatedAt()
		atc.mutation.SetUpdatedAt(v)
	}
	if _, ok := atc.mutation.Tags(); !ok {
		v := apitoken.DefaultTags
		atc.mutation.SetTags(v)
	}
	if _, ok := atc.mutation.Token(); !ok {
		if apitoken.DefaultToken == nil {
			return fmt.Errorf("generated: uninitialized apitoken.DefaultToken (forgotten import generated/runtime?)")
		}
		v := apitoken.DefaultToken()
		atc.mutation.SetToken(v)
	}
	if _, ok := atc.mutation.IsActive(); !ok {
		v := apitoken.DefaultIsActive
		atc.mutation.SetIsActive(v)
	}
	if _, ok := atc.mutation.ID(); !ok {
		if apitoken.DefaultID == nil {
			return fmt.Errorf("generated: uninitialized apitoken.DefaultID (forgotten import generated/runtime?)")
		}
		v := apitoken.DefaultID()
		atc.mutation.SetID(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (atc *APITokenCreate) check() error {
	if v, ok := atc.mutation.OwnerID(); ok {
		if err := apitoken.OwnerIDValidator(v); err != nil {
			return &ValidationError{Name: "owner_id", err: fmt.Errorf(`generated: validator failed for field "APIToken.owner_id": %w`, err)}
		}
	}
	if _, ok := atc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`generated: missing required field "APIToken.name"`)}
	}
	if v, ok := atc.mutation.Name(); ok {
		if err := apitoken.NameValidator(v); err != nil {
			return &ValidationError{Name: "name", err: fmt.Errorf(`generated: validator failed for field "APIToken.name": %w`, err)}
		}
	}
	if _, ok := atc.mutation.Token(); !ok {
		return &ValidationError{Name: "token", err: errors.New(`generated: missing required field "APIToken.token"`)}
	}
	return nil
}

func (atc *APITokenCreate) sqlSave(ctx context.Context) (*APIToken, error) {
	if err := atc.check(); err != nil {
		return nil, err
	}
	_node, _spec := atc.createSpec()
	if err := sqlgraph.CreateNode(ctx, atc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected APIToken.ID type: %T", _spec.ID.Value)
		}
	}
	atc.mutation.id = &_node.ID
	atc.mutation.done = true
	return _node, nil
}

func (atc *APITokenCreate) createSpec() (*APIToken, *sqlgraph.CreateSpec) {
	var (
		_node = &APIToken{config: atc.config}
		_spec = sqlgraph.NewCreateSpec(apitoken.Table, sqlgraph.NewFieldSpec(apitoken.FieldID, field.TypeString))
	)
	_spec.Schema = atc.schemaConfig.APIToken
	if id, ok := atc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := atc.mutation.CreatedAt(); ok {
		_spec.SetField(apitoken.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := atc.mutation.UpdatedAt(); ok {
		_spec.SetField(apitoken.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := atc.mutation.CreatedBy(); ok {
		_spec.SetField(apitoken.FieldCreatedBy, field.TypeString, value)
		_node.CreatedBy = value
	}
	if value, ok := atc.mutation.UpdatedBy(); ok {
		_spec.SetField(apitoken.FieldUpdatedBy, field.TypeString, value)
		_node.UpdatedBy = value
	}
	if value, ok := atc.mutation.DeletedAt(); ok {
		_spec.SetField(apitoken.FieldDeletedAt, field.TypeTime, value)
		_node.DeletedAt = value
	}
	if value, ok := atc.mutation.DeletedBy(); ok {
		_spec.SetField(apitoken.FieldDeletedBy, field.TypeString, value)
		_node.DeletedBy = value
	}
	if value, ok := atc.mutation.Tags(); ok {
		_spec.SetField(apitoken.FieldTags, field.TypeJSON, value)
		_node.Tags = value
	}
	if value, ok := atc.mutation.Name(); ok {
		_spec.SetField(apitoken.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := atc.mutation.Token(); ok {
		_spec.SetField(apitoken.FieldToken, field.TypeString, value)
		_node.Token = value
	}
	if value, ok := atc.mutation.ExpiresAt(); ok {
		_spec.SetField(apitoken.FieldExpiresAt, field.TypeTime, value)
		_node.ExpiresAt = &value
	}
	if value, ok := atc.mutation.Description(); ok {
		_spec.SetField(apitoken.FieldDescription, field.TypeString, value)
		_node.Description = &value
	}
	if value, ok := atc.mutation.Scopes(); ok {
		_spec.SetField(apitoken.FieldScopes, field.TypeJSON, value)
		_node.Scopes = value
	}
	if value, ok := atc.mutation.LastUsedAt(); ok {
		_spec.SetField(apitoken.FieldLastUsedAt, field.TypeTime, value)
		_node.LastUsedAt = &value
	}
	if value, ok := atc.mutation.IsActive(); ok {
		_spec.SetField(apitoken.FieldIsActive, field.TypeBool, value)
		_node.IsActive = value
	}
	if value, ok := atc.mutation.RevokedReason(); ok {
		_spec.SetField(apitoken.FieldRevokedReason, field.TypeString, value)
		_node.RevokedReason = &value
	}
	if value, ok := atc.mutation.RevokedBy(); ok {
		_spec.SetField(apitoken.FieldRevokedBy, field.TypeString, value)
		_node.RevokedBy = &value
	}
	if value, ok := atc.mutation.RevokedAt(); ok {
		_spec.SetField(apitoken.FieldRevokedAt, field.TypeTime, value)
		_node.RevokedAt = &value
	}
	if value, ok := atc.mutation.SSOAuthorizations(); ok {
		_spec.SetField(apitoken.FieldSSOAuthorizations, field.TypeJSON, value)
		_node.SSOAuthorizations = value
	}
	if nodes := atc.mutation.OwnerIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   apitoken.OwnerTable,
			Columns: []string{apitoken.OwnerColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(organization.FieldID, field.TypeString),
			},
		}
		edge.Schema = atc.schemaConfig.APIToken
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.OwnerID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// APITokenCreateBulk is the builder for creating many APIToken entities in bulk.
type APITokenCreateBulk struct {
	config
	err      error
	builders []*APITokenCreate
}

// Save creates the APIToken entities in the database.
func (atcb *APITokenCreateBulk) Save(ctx context.Context) ([]*APIToken, error) {
	if atcb.err != nil {
		return nil, atcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(atcb.builders))
	nodes := make([]*APIToken, len(atcb.builders))
	mutators := make([]Mutator, len(atcb.builders))
	for i := range atcb.builders {
		func(i int, root context.Context) {
			builder := atcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*APITokenMutation)
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
					_, err = mutators[i+1].Mutate(root, atcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, atcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, atcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (atcb *APITokenCreateBulk) SaveX(ctx context.Context) []*APIToken {
	v, err := atcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (atcb *APITokenCreateBulk) Exec(ctx context.Context) error {
	_, err := atcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (atcb *APITokenCreateBulk) ExecX(ctx context.Context) {
	if err := atcb.Exec(ctx); err != nil {
		panic(err)
	}
}
