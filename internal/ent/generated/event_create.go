// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/event"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/hush"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/invite"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/ent/generated/personalaccesstoken"
	"github.com/theopenlane/core/internal/ent/generated/subscriber"
	"github.com/theopenlane/core/internal/ent/generated/user"
)

// EventCreate is the builder for creating a Event entity.
type EventCreate struct {
	config
	mutation *EventMutation
	hooks    []Hook
}

// SetCreatedAt sets the "created_at" field.
func (ec *EventCreate) SetCreatedAt(t time.Time) *EventCreate {
	ec.mutation.SetCreatedAt(t)
	return ec
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (ec *EventCreate) SetNillableCreatedAt(t *time.Time) *EventCreate {
	if t != nil {
		ec.SetCreatedAt(*t)
	}
	return ec
}

// SetUpdatedAt sets the "updated_at" field.
func (ec *EventCreate) SetUpdatedAt(t time.Time) *EventCreate {
	ec.mutation.SetUpdatedAt(t)
	return ec
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (ec *EventCreate) SetNillableUpdatedAt(t *time.Time) *EventCreate {
	if t != nil {
		ec.SetUpdatedAt(*t)
	}
	return ec
}

// SetCreatedBy sets the "created_by" field.
func (ec *EventCreate) SetCreatedBy(s string) *EventCreate {
	ec.mutation.SetCreatedBy(s)
	return ec
}

// SetNillableCreatedBy sets the "created_by" field if the given value is not nil.
func (ec *EventCreate) SetNillableCreatedBy(s *string) *EventCreate {
	if s != nil {
		ec.SetCreatedBy(*s)
	}
	return ec
}

// SetUpdatedBy sets the "updated_by" field.
func (ec *EventCreate) SetUpdatedBy(s string) *EventCreate {
	ec.mutation.SetUpdatedBy(s)
	return ec
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (ec *EventCreate) SetNillableUpdatedBy(s *string) *EventCreate {
	if s != nil {
		ec.SetUpdatedBy(*s)
	}
	return ec
}

// SetTags sets the "tags" field.
func (ec *EventCreate) SetTags(s []string) *EventCreate {
	ec.mutation.SetTags(s)
	return ec
}

// SetEventID sets the "event_id" field.
func (ec *EventCreate) SetEventID(s string) *EventCreate {
	ec.mutation.SetEventID(s)
	return ec
}

// SetNillableEventID sets the "event_id" field if the given value is not nil.
func (ec *EventCreate) SetNillableEventID(s *string) *EventCreate {
	if s != nil {
		ec.SetEventID(*s)
	}
	return ec
}

// SetCorrelationID sets the "correlation_id" field.
func (ec *EventCreate) SetCorrelationID(s string) *EventCreate {
	ec.mutation.SetCorrelationID(s)
	return ec
}

// SetNillableCorrelationID sets the "correlation_id" field if the given value is not nil.
func (ec *EventCreate) SetNillableCorrelationID(s *string) *EventCreate {
	if s != nil {
		ec.SetCorrelationID(*s)
	}
	return ec
}

// SetEventType sets the "event_type" field.
func (ec *EventCreate) SetEventType(s string) *EventCreate {
	ec.mutation.SetEventType(s)
	return ec
}

// SetMetadata sets the "metadata" field.
func (ec *EventCreate) SetMetadata(m map[string]interface{}) *EventCreate {
	ec.mutation.SetMetadata(m)
	return ec
}

// SetID sets the "id" field.
func (ec *EventCreate) SetID(s string) *EventCreate {
	ec.mutation.SetID(s)
	return ec
}

// SetNillableID sets the "id" field if the given value is not nil.
func (ec *EventCreate) SetNillableID(s *string) *EventCreate {
	if s != nil {
		ec.SetID(*s)
	}
	return ec
}

// AddUserIDs adds the "user" edge to the User entity by IDs.
func (ec *EventCreate) AddUserIDs(ids ...string) *EventCreate {
	ec.mutation.AddUserIDs(ids...)
	return ec
}

// AddUser adds the "user" edges to the User entity.
func (ec *EventCreate) AddUser(u ...*User) *EventCreate {
	ids := make([]string, len(u))
	for i := range u {
		ids[i] = u[i].ID
	}
	return ec.AddUserIDs(ids...)
}

// AddGroupIDs adds the "group" edge to the Group entity by IDs.
func (ec *EventCreate) AddGroupIDs(ids ...string) *EventCreate {
	ec.mutation.AddGroupIDs(ids...)
	return ec
}

// AddGroup adds the "group" edges to the Group entity.
func (ec *EventCreate) AddGroup(g ...*Group) *EventCreate {
	ids := make([]string, len(g))
	for i := range g {
		ids[i] = g[i].ID
	}
	return ec.AddGroupIDs(ids...)
}

// AddIntegrationIDs adds the "integration" edge to the Integration entity by IDs.
func (ec *EventCreate) AddIntegrationIDs(ids ...string) *EventCreate {
	ec.mutation.AddIntegrationIDs(ids...)
	return ec
}

// AddIntegration adds the "integration" edges to the Integration entity.
func (ec *EventCreate) AddIntegration(i ...*Integration) *EventCreate {
	ids := make([]string, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return ec.AddIntegrationIDs(ids...)
}

// AddOrganizationIDs adds the "organization" edge to the Organization entity by IDs.
func (ec *EventCreate) AddOrganizationIDs(ids ...string) *EventCreate {
	ec.mutation.AddOrganizationIDs(ids...)
	return ec
}

// AddOrganization adds the "organization" edges to the Organization entity.
func (ec *EventCreate) AddOrganization(o ...*Organization) *EventCreate {
	ids := make([]string, len(o))
	for i := range o {
		ids[i] = o[i].ID
	}
	return ec.AddOrganizationIDs(ids...)
}

// AddInviteIDs adds the "invite" edge to the Invite entity by IDs.
func (ec *EventCreate) AddInviteIDs(ids ...string) *EventCreate {
	ec.mutation.AddInviteIDs(ids...)
	return ec
}

// AddInvite adds the "invite" edges to the Invite entity.
func (ec *EventCreate) AddInvite(i ...*Invite) *EventCreate {
	ids := make([]string, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return ec.AddInviteIDs(ids...)
}

// AddPersonalAccessTokenIDs adds the "personal_access_token" edge to the PersonalAccessToken entity by IDs.
func (ec *EventCreate) AddPersonalAccessTokenIDs(ids ...string) *EventCreate {
	ec.mutation.AddPersonalAccessTokenIDs(ids...)
	return ec
}

// AddPersonalAccessToken adds the "personal_access_token" edges to the PersonalAccessToken entity.
func (ec *EventCreate) AddPersonalAccessToken(p ...*PersonalAccessToken) *EventCreate {
	ids := make([]string, len(p))
	for i := range p {
		ids[i] = p[i].ID
	}
	return ec.AddPersonalAccessTokenIDs(ids...)
}

// AddHushIDs adds the "hush" edge to the Hush entity by IDs.
func (ec *EventCreate) AddHushIDs(ids ...string) *EventCreate {
	ec.mutation.AddHushIDs(ids...)
	return ec
}

// AddHush adds the "hush" edges to the Hush entity.
func (ec *EventCreate) AddHush(h ...*Hush) *EventCreate {
	ids := make([]string, len(h))
	for i := range h {
		ids[i] = h[i].ID
	}
	return ec.AddHushIDs(ids...)
}

// AddOrgmembershipIDs adds the "orgmembership" edge to the OrgMembership entity by IDs.
func (ec *EventCreate) AddOrgmembershipIDs(ids ...string) *EventCreate {
	ec.mutation.AddOrgmembershipIDs(ids...)
	return ec
}

// AddOrgmembership adds the "orgmembership" edges to the OrgMembership entity.
func (ec *EventCreate) AddOrgmembership(o ...*OrgMembership) *EventCreate {
	ids := make([]string, len(o))
	for i := range o {
		ids[i] = o[i].ID
	}
	return ec.AddOrgmembershipIDs(ids...)
}

// AddGroupmembershipIDs adds the "groupmembership" edge to the GroupMembership entity by IDs.
func (ec *EventCreate) AddGroupmembershipIDs(ids ...string) *EventCreate {
	ec.mutation.AddGroupmembershipIDs(ids...)
	return ec
}

// AddGroupmembership adds the "groupmembership" edges to the GroupMembership entity.
func (ec *EventCreate) AddGroupmembership(g ...*GroupMembership) *EventCreate {
	ids := make([]string, len(g))
	for i := range g {
		ids[i] = g[i].ID
	}
	return ec.AddGroupmembershipIDs(ids...)
}

// AddSubscriberIDs adds the "subscriber" edge to the Subscriber entity by IDs.
func (ec *EventCreate) AddSubscriberIDs(ids ...string) *EventCreate {
	ec.mutation.AddSubscriberIDs(ids...)
	return ec
}

// AddSubscriber adds the "subscriber" edges to the Subscriber entity.
func (ec *EventCreate) AddSubscriber(s ...*Subscriber) *EventCreate {
	ids := make([]string, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return ec.AddSubscriberIDs(ids...)
}

// AddFileIDs adds the "file" edge to the File entity by IDs.
func (ec *EventCreate) AddFileIDs(ids ...string) *EventCreate {
	ec.mutation.AddFileIDs(ids...)
	return ec
}

// AddFile adds the "file" edges to the File entity.
func (ec *EventCreate) AddFile(f ...*File) *EventCreate {
	ids := make([]string, len(f))
	for i := range f {
		ids[i] = f[i].ID
	}
	return ec.AddFileIDs(ids...)
}

// AddOrgsubscriptionIDs adds the "orgsubscription" edge to the OrgSubscription entity by IDs.
func (ec *EventCreate) AddOrgsubscriptionIDs(ids ...string) *EventCreate {
	ec.mutation.AddOrgsubscriptionIDs(ids...)
	return ec
}

// AddOrgsubscription adds the "orgsubscription" edges to the OrgSubscription entity.
func (ec *EventCreate) AddOrgsubscription(o ...*OrgSubscription) *EventCreate {
	ids := make([]string, len(o))
	for i := range o {
		ids[i] = o[i].ID
	}
	return ec.AddOrgsubscriptionIDs(ids...)
}

// Mutation returns the EventMutation object of the builder.
func (ec *EventCreate) Mutation() *EventMutation {
	return ec.mutation
}

// Save creates the Event in the database.
func (ec *EventCreate) Save(ctx context.Context) (*Event, error) {
	if err := ec.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, ec.sqlSave, ec.mutation, ec.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (ec *EventCreate) SaveX(ctx context.Context) *Event {
	v, err := ec.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ec *EventCreate) Exec(ctx context.Context) error {
	_, err := ec.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ec *EventCreate) ExecX(ctx context.Context) {
	if err := ec.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (ec *EventCreate) defaults() error {
	if _, ok := ec.mutation.CreatedAt(); !ok {
		if event.DefaultCreatedAt == nil {
			return fmt.Errorf("generated: uninitialized event.DefaultCreatedAt (forgotten import generated/runtime?)")
		}
		v := event.DefaultCreatedAt()
		ec.mutation.SetCreatedAt(v)
	}
	if _, ok := ec.mutation.UpdatedAt(); !ok {
		if event.DefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized event.DefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := event.DefaultUpdatedAt()
		ec.mutation.SetUpdatedAt(v)
	}
	if _, ok := ec.mutation.Tags(); !ok {
		v := event.DefaultTags
		ec.mutation.SetTags(v)
	}
	if _, ok := ec.mutation.ID(); !ok {
		if event.DefaultID == nil {
			return fmt.Errorf("generated: uninitialized event.DefaultID (forgotten import generated/runtime?)")
		}
		v := event.DefaultID()
		ec.mutation.SetID(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (ec *EventCreate) check() error {
	if _, ok := ec.mutation.EventType(); !ok {
		return &ValidationError{Name: "event_type", err: errors.New(`generated: missing required field "Event.event_type"`)}
	}
	return nil
}

func (ec *EventCreate) sqlSave(ctx context.Context) (*Event, error) {
	if err := ec.check(); err != nil {
		return nil, err
	}
	_node, _spec := ec.createSpec()
	if err := sqlgraph.CreateNode(ctx, ec.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected Event.ID type: %T", _spec.ID.Value)
		}
	}
	ec.mutation.id = &_node.ID
	ec.mutation.done = true
	return _node, nil
}

func (ec *EventCreate) createSpec() (*Event, *sqlgraph.CreateSpec) {
	var (
		_node = &Event{config: ec.config}
		_spec = sqlgraph.NewCreateSpec(event.Table, sqlgraph.NewFieldSpec(event.FieldID, field.TypeString))
	)
	_spec.Schema = ec.schemaConfig.Event
	if id, ok := ec.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := ec.mutation.CreatedAt(); ok {
		_spec.SetField(event.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := ec.mutation.UpdatedAt(); ok {
		_spec.SetField(event.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := ec.mutation.CreatedBy(); ok {
		_spec.SetField(event.FieldCreatedBy, field.TypeString, value)
		_node.CreatedBy = value
	}
	if value, ok := ec.mutation.UpdatedBy(); ok {
		_spec.SetField(event.FieldUpdatedBy, field.TypeString, value)
		_node.UpdatedBy = value
	}
	if value, ok := ec.mutation.Tags(); ok {
		_spec.SetField(event.FieldTags, field.TypeJSON, value)
		_node.Tags = value
	}
	if value, ok := ec.mutation.EventID(); ok {
		_spec.SetField(event.FieldEventID, field.TypeString, value)
		_node.EventID = value
	}
	if value, ok := ec.mutation.CorrelationID(); ok {
		_spec.SetField(event.FieldCorrelationID, field.TypeString, value)
		_node.CorrelationID = value
	}
	if value, ok := ec.mutation.EventType(); ok {
		_spec.SetField(event.FieldEventType, field.TypeString, value)
		_node.EventType = value
	}
	if value, ok := ec.mutation.Metadata(); ok {
		_spec.SetField(event.FieldMetadata, field.TypeJSON, value)
		_node.Metadata = value
	}
	if nodes := ec.mutation.UserIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   event.UserTable,
			Columns: event.UserPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeString),
			},
		}
		edge.Schema = ec.schemaConfig.UserEvents
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.GroupIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   event.GroupTable,
			Columns: event.GroupPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(group.FieldID, field.TypeString),
			},
		}
		edge.Schema = ec.schemaConfig.GroupEvents
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.IntegrationIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   event.IntegrationTable,
			Columns: event.IntegrationPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(integration.FieldID, field.TypeString),
			},
		}
		edge.Schema = ec.schemaConfig.IntegrationEvents
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.OrganizationIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   event.OrganizationTable,
			Columns: event.OrganizationPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(organization.FieldID, field.TypeString),
			},
		}
		edge.Schema = ec.schemaConfig.OrganizationEvents
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.InviteIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   event.InviteTable,
			Columns: event.InvitePrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(invite.FieldID, field.TypeString),
			},
		}
		edge.Schema = ec.schemaConfig.InviteEvents
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.PersonalAccessTokenIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   event.PersonalAccessTokenTable,
			Columns: event.PersonalAccessTokenPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(personalaccesstoken.FieldID, field.TypeString),
			},
		}
		edge.Schema = ec.schemaConfig.PersonalAccessTokenEvents
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.HushIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   event.HushTable,
			Columns: event.HushPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(hush.FieldID, field.TypeString),
			},
		}
		edge.Schema = ec.schemaConfig.HushEvents
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.OrgmembershipIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   event.OrgmembershipTable,
			Columns: event.OrgmembershipPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(orgmembership.FieldID, field.TypeString),
			},
		}
		edge.Schema = ec.schemaConfig.OrgMembershipEvents
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.GroupmembershipIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   event.GroupmembershipTable,
			Columns: event.GroupmembershipPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(groupmembership.FieldID, field.TypeString),
			},
		}
		edge.Schema = ec.schemaConfig.GroupMembershipEvents
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.SubscriberIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   event.SubscriberTable,
			Columns: event.SubscriberPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(subscriber.FieldID, field.TypeString),
			},
		}
		edge.Schema = ec.schemaConfig.SubscriberEvents
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.FileIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   event.FileTable,
			Columns: event.FilePrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(file.FieldID, field.TypeString),
			},
		}
		edge.Schema = ec.schemaConfig.FileEvents
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ec.mutation.OrgsubscriptionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   event.OrgsubscriptionTable,
			Columns: event.OrgsubscriptionPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(orgsubscription.FieldID, field.TypeString),
			},
		}
		edge.Schema = ec.schemaConfig.OrgSubscriptionEvents
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// EventCreateBulk is the builder for creating many Event entities in bulk.
type EventCreateBulk struct {
	config
	err      error
	builders []*EventCreate
}

// Save creates the Event entities in the database.
func (ecb *EventCreateBulk) Save(ctx context.Context) ([]*Event, error) {
	if ecb.err != nil {
		return nil, ecb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(ecb.builders))
	nodes := make([]*Event, len(ecb.builders))
	mutators := make([]Mutator, len(ecb.builders))
	for i := range ecb.builders {
		func(i int, root context.Context) {
			builder := ecb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*EventMutation)
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
					_, err = mutators[i+1].Mutate(root, ecb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ecb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, ecb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ecb *EventCreateBulk) SaveX(ctx context.Context) []*Event {
	v, err := ecb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ecb *EventCreateBulk) Exec(ctx context.Context) error {
	_, err := ecb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ecb *EventCreateBulk) ExecX(ctx context.Context) {
	if err := ecb.Exec(ctx); err != nil {
		panic(err)
	}
}
