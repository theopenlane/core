// Code generated by ent, DO NOT EDIT.

package generated

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/core/internal/ent/generated/event"
)

// Event is the model entity for the Event schema.
type Event struct {
	config `json:"-"`
	// ID of the ent.
	ID string `json:"id,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// CreatedBy holds the value of the "created_by" field.
	CreatedBy string `json:"created_by,omitempty"`
	// UpdatedBy holds the value of the "updated_by" field.
	UpdatedBy string `json:"updated_by,omitempty"`
	// tags associated with the object
	Tags []string `json:"tags,omitempty"`
	// EventID holds the value of the "event_id" field.
	EventID string `json:"event_id,omitempty"`
	// CorrelationID holds the value of the "correlation_id" field.
	CorrelationID string `json:"correlation_id,omitempty"`
	// EventType holds the value of the "event_type" field.
	EventType string `json:"event_type,omitempty"`
	// Metadata holds the value of the "metadata" field.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the EventQuery when eager-loading is set.
	Edges        EventEdges `json:"edges"`
	selectValues sql.SelectValues
}

// EventEdges holds the relations/edges for other nodes in the graph.
type EventEdges struct {
	// User holds the value of the user edge.
	User []*User `json:"user,omitempty"`
	// Group holds the value of the group edge.
	Group []*Group `json:"group,omitempty"`
	// Integration holds the value of the integration edge.
	Integration []*Integration `json:"integration,omitempty"`
	// Organization holds the value of the organization edge.
	Organization []*Organization `json:"organization,omitempty"`
	// Invite holds the value of the invite edge.
	Invite []*Invite `json:"invite,omitempty"`
	// PersonalAccessToken holds the value of the personal_access_token edge.
	PersonalAccessToken []*PersonalAccessToken `json:"personal_access_token,omitempty"`
	// Hush holds the value of the hush edge.
	Hush []*Hush `json:"hush,omitempty"`
	// Orgmembership holds the value of the orgmembership edge.
	Orgmembership []*OrgMembership `json:"orgmembership,omitempty"`
	// Groupmembership holds the value of the groupmembership edge.
	Groupmembership []*GroupMembership `json:"groupmembership,omitempty"`
	// Subscriber holds the value of the subscriber edge.
	Subscriber []*Subscriber `json:"subscriber,omitempty"`
	// File holds the value of the file edge.
	File []*File `json:"file,omitempty"`
	// Orgsubscription holds the value of the orgsubscription edge.
	Orgsubscription []*OrgSubscription `json:"orgsubscription,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [12]bool
	// totalCount holds the count of the edges above.
	totalCount [12]map[string]int

	namedUser                map[string][]*User
	namedGroup               map[string][]*Group
	namedIntegration         map[string][]*Integration
	namedOrganization        map[string][]*Organization
	namedInvite              map[string][]*Invite
	namedPersonalAccessToken map[string][]*PersonalAccessToken
	namedHush                map[string][]*Hush
	namedOrgmembership       map[string][]*OrgMembership
	namedGroupmembership     map[string][]*GroupMembership
	namedSubscriber          map[string][]*Subscriber
	namedFile                map[string][]*File
	namedOrgsubscription     map[string][]*OrgSubscription
}

// UserOrErr returns the User value or an error if the edge
// was not loaded in eager-loading.
func (e EventEdges) UserOrErr() ([]*User, error) {
	if e.loadedTypes[0] {
		return e.User, nil
	}
	return nil, &NotLoadedError{edge: "user"}
}

// GroupOrErr returns the Group value or an error if the edge
// was not loaded in eager-loading.
func (e EventEdges) GroupOrErr() ([]*Group, error) {
	if e.loadedTypes[1] {
		return e.Group, nil
	}
	return nil, &NotLoadedError{edge: "group"}
}

// IntegrationOrErr returns the Integration value or an error if the edge
// was not loaded in eager-loading.
func (e EventEdges) IntegrationOrErr() ([]*Integration, error) {
	if e.loadedTypes[2] {
		return e.Integration, nil
	}
	return nil, &NotLoadedError{edge: "integration"}
}

// OrganizationOrErr returns the Organization value or an error if the edge
// was not loaded in eager-loading.
func (e EventEdges) OrganizationOrErr() ([]*Organization, error) {
	if e.loadedTypes[3] {
		return e.Organization, nil
	}
	return nil, &NotLoadedError{edge: "organization"}
}

// InviteOrErr returns the Invite value or an error if the edge
// was not loaded in eager-loading.
func (e EventEdges) InviteOrErr() ([]*Invite, error) {
	if e.loadedTypes[4] {
		return e.Invite, nil
	}
	return nil, &NotLoadedError{edge: "invite"}
}

// PersonalAccessTokenOrErr returns the PersonalAccessToken value or an error if the edge
// was not loaded in eager-loading.
func (e EventEdges) PersonalAccessTokenOrErr() ([]*PersonalAccessToken, error) {
	if e.loadedTypes[5] {
		return e.PersonalAccessToken, nil
	}
	return nil, &NotLoadedError{edge: "personal_access_token"}
}

// HushOrErr returns the Hush value or an error if the edge
// was not loaded in eager-loading.
func (e EventEdges) HushOrErr() ([]*Hush, error) {
	if e.loadedTypes[6] {
		return e.Hush, nil
	}
	return nil, &NotLoadedError{edge: "hush"}
}

// OrgmembershipOrErr returns the Orgmembership value or an error if the edge
// was not loaded in eager-loading.
func (e EventEdges) OrgmembershipOrErr() ([]*OrgMembership, error) {
	if e.loadedTypes[7] {
		return e.Orgmembership, nil
	}
	return nil, &NotLoadedError{edge: "orgmembership"}
}

// GroupmembershipOrErr returns the Groupmembership value or an error if the edge
// was not loaded in eager-loading.
func (e EventEdges) GroupmembershipOrErr() ([]*GroupMembership, error) {
	if e.loadedTypes[8] {
		return e.Groupmembership, nil
	}
	return nil, &NotLoadedError{edge: "groupmembership"}
}

// SubscriberOrErr returns the Subscriber value or an error if the edge
// was not loaded in eager-loading.
func (e EventEdges) SubscriberOrErr() ([]*Subscriber, error) {
	if e.loadedTypes[9] {
		return e.Subscriber, nil
	}
	return nil, &NotLoadedError{edge: "subscriber"}
}

// FileOrErr returns the File value or an error if the edge
// was not loaded in eager-loading.
func (e EventEdges) FileOrErr() ([]*File, error) {
	if e.loadedTypes[10] {
		return e.File, nil
	}
	return nil, &NotLoadedError{edge: "file"}
}

// OrgsubscriptionOrErr returns the Orgsubscription value or an error if the edge
// was not loaded in eager-loading.
func (e EventEdges) OrgsubscriptionOrErr() ([]*OrgSubscription, error) {
	if e.loadedTypes[11] {
		return e.Orgsubscription, nil
	}
	return nil, &NotLoadedError{edge: "orgsubscription"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Event) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case event.FieldTags, event.FieldMetadata:
			values[i] = new([]byte)
		case event.FieldID, event.FieldCreatedBy, event.FieldUpdatedBy, event.FieldEventID, event.FieldCorrelationID, event.FieldEventType:
			values[i] = new(sql.NullString)
		case event.FieldCreatedAt, event.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Event fields.
func (e *Event) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case event.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				e.ID = value.String
			}
		case event.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				e.CreatedAt = value.Time
			}
		case event.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				e.UpdatedAt = value.Time
			}
		case event.FieldCreatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_by", values[i])
			} else if value.Valid {
				e.CreatedBy = value.String
			}
		case event.FieldUpdatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_by", values[i])
			} else if value.Valid {
				e.UpdatedBy = value.String
			}
		case event.FieldTags:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field tags", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &e.Tags); err != nil {
					return fmt.Errorf("unmarshal field tags: %w", err)
				}
			}
		case event.FieldEventID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field event_id", values[i])
			} else if value.Valid {
				e.EventID = value.String
			}
		case event.FieldCorrelationID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field correlation_id", values[i])
			} else if value.Valid {
				e.CorrelationID = value.String
			}
		case event.FieldEventType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field event_type", values[i])
			} else if value.Valid {
				e.EventType = value.String
			}
		case event.FieldMetadata:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field metadata", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &e.Metadata); err != nil {
					return fmt.Errorf("unmarshal field metadata: %w", err)
				}
			}
		default:
			e.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Event.
// This includes values selected through modifiers, order, etc.
func (e *Event) Value(name string) (ent.Value, error) {
	return e.selectValues.Get(name)
}

// QueryUser queries the "user" edge of the Event entity.
func (e *Event) QueryUser() *UserQuery {
	return NewEventClient(e.config).QueryUser(e)
}

// QueryGroup queries the "group" edge of the Event entity.
func (e *Event) QueryGroup() *GroupQuery {
	return NewEventClient(e.config).QueryGroup(e)
}

// QueryIntegration queries the "integration" edge of the Event entity.
func (e *Event) QueryIntegration() *IntegrationQuery {
	return NewEventClient(e.config).QueryIntegration(e)
}

// QueryOrganization queries the "organization" edge of the Event entity.
func (e *Event) QueryOrganization() *OrganizationQuery {
	return NewEventClient(e.config).QueryOrganization(e)
}

// QueryInvite queries the "invite" edge of the Event entity.
func (e *Event) QueryInvite() *InviteQuery {
	return NewEventClient(e.config).QueryInvite(e)
}

// QueryPersonalAccessToken queries the "personal_access_token" edge of the Event entity.
func (e *Event) QueryPersonalAccessToken() *PersonalAccessTokenQuery {
	return NewEventClient(e.config).QueryPersonalAccessToken(e)
}

// QueryHush queries the "hush" edge of the Event entity.
func (e *Event) QueryHush() *HushQuery {
	return NewEventClient(e.config).QueryHush(e)
}

// QueryOrgmembership queries the "orgmembership" edge of the Event entity.
func (e *Event) QueryOrgmembership() *OrgMembershipQuery {
	return NewEventClient(e.config).QueryOrgmembership(e)
}

// QueryGroupmembership queries the "groupmembership" edge of the Event entity.
func (e *Event) QueryGroupmembership() *GroupMembershipQuery {
	return NewEventClient(e.config).QueryGroupmembership(e)
}

// QuerySubscriber queries the "subscriber" edge of the Event entity.
func (e *Event) QuerySubscriber() *SubscriberQuery {
	return NewEventClient(e.config).QuerySubscriber(e)
}

// QueryFile queries the "file" edge of the Event entity.
func (e *Event) QueryFile() *FileQuery {
	return NewEventClient(e.config).QueryFile(e)
}

// QueryOrgsubscription queries the "orgsubscription" edge of the Event entity.
func (e *Event) QueryOrgsubscription() *OrgSubscriptionQuery {
	return NewEventClient(e.config).QueryOrgsubscription(e)
}

// Update returns a builder for updating this Event.
// Note that you need to call Event.Unwrap() before calling this method if this Event
// was returned from a transaction, and the transaction was committed or rolled back.
func (e *Event) Update() *EventUpdateOne {
	return NewEventClient(e.config).UpdateOne(e)
}

// Unwrap unwraps the Event entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (e *Event) Unwrap() *Event {
	_tx, ok := e.config.driver.(*txDriver)
	if !ok {
		panic("generated: Event is not a transactional entity")
	}
	e.config.driver = _tx.drv
	return e
}

// String implements the fmt.Stringer.
func (e *Event) String() string {
	var builder strings.Builder
	builder.WriteString("Event(")
	builder.WriteString(fmt.Sprintf("id=%v, ", e.ID))
	builder.WriteString("created_at=")
	builder.WriteString(e.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(e.UpdatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("created_by=")
	builder.WriteString(e.CreatedBy)
	builder.WriteString(", ")
	builder.WriteString("updated_by=")
	builder.WriteString(e.UpdatedBy)
	builder.WriteString(", ")
	builder.WriteString("tags=")
	builder.WriteString(fmt.Sprintf("%v", e.Tags))
	builder.WriteString(", ")
	builder.WriteString("event_id=")
	builder.WriteString(e.EventID)
	builder.WriteString(", ")
	builder.WriteString("correlation_id=")
	builder.WriteString(e.CorrelationID)
	builder.WriteString(", ")
	builder.WriteString("event_type=")
	builder.WriteString(e.EventType)
	builder.WriteString(", ")
	builder.WriteString("metadata=")
	builder.WriteString(fmt.Sprintf("%v", e.Metadata))
	builder.WriteByte(')')
	return builder.String()
}

// NamedUser returns the User named value or an error if the edge was not
// loaded in eager-loading with this name.
func (e *Event) NamedUser(name string) ([]*User, error) {
	if e.Edges.namedUser == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := e.Edges.namedUser[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (e *Event) appendNamedUser(name string, edges ...*User) {
	if e.Edges.namedUser == nil {
		e.Edges.namedUser = make(map[string][]*User)
	}
	if len(edges) == 0 {
		e.Edges.namedUser[name] = []*User{}
	} else {
		e.Edges.namedUser[name] = append(e.Edges.namedUser[name], edges...)
	}
}

// NamedGroup returns the Group named value or an error if the edge was not
// loaded in eager-loading with this name.
func (e *Event) NamedGroup(name string) ([]*Group, error) {
	if e.Edges.namedGroup == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := e.Edges.namedGroup[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (e *Event) appendNamedGroup(name string, edges ...*Group) {
	if e.Edges.namedGroup == nil {
		e.Edges.namedGroup = make(map[string][]*Group)
	}
	if len(edges) == 0 {
		e.Edges.namedGroup[name] = []*Group{}
	} else {
		e.Edges.namedGroup[name] = append(e.Edges.namedGroup[name], edges...)
	}
}

// NamedIntegration returns the Integration named value or an error if the edge was not
// loaded in eager-loading with this name.
func (e *Event) NamedIntegration(name string) ([]*Integration, error) {
	if e.Edges.namedIntegration == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := e.Edges.namedIntegration[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (e *Event) appendNamedIntegration(name string, edges ...*Integration) {
	if e.Edges.namedIntegration == nil {
		e.Edges.namedIntegration = make(map[string][]*Integration)
	}
	if len(edges) == 0 {
		e.Edges.namedIntegration[name] = []*Integration{}
	} else {
		e.Edges.namedIntegration[name] = append(e.Edges.namedIntegration[name], edges...)
	}
}

// NamedOrganization returns the Organization named value or an error if the edge was not
// loaded in eager-loading with this name.
func (e *Event) NamedOrganization(name string) ([]*Organization, error) {
	if e.Edges.namedOrganization == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := e.Edges.namedOrganization[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (e *Event) appendNamedOrganization(name string, edges ...*Organization) {
	if e.Edges.namedOrganization == nil {
		e.Edges.namedOrganization = make(map[string][]*Organization)
	}
	if len(edges) == 0 {
		e.Edges.namedOrganization[name] = []*Organization{}
	} else {
		e.Edges.namedOrganization[name] = append(e.Edges.namedOrganization[name], edges...)
	}
}

// NamedInvite returns the Invite named value or an error if the edge was not
// loaded in eager-loading with this name.
func (e *Event) NamedInvite(name string) ([]*Invite, error) {
	if e.Edges.namedInvite == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := e.Edges.namedInvite[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (e *Event) appendNamedInvite(name string, edges ...*Invite) {
	if e.Edges.namedInvite == nil {
		e.Edges.namedInvite = make(map[string][]*Invite)
	}
	if len(edges) == 0 {
		e.Edges.namedInvite[name] = []*Invite{}
	} else {
		e.Edges.namedInvite[name] = append(e.Edges.namedInvite[name], edges...)
	}
}

// NamedPersonalAccessToken returns the PersonalAccessToken named value or an error if the edge was not
// loaded in eager-loading with this name.
func (e *Event) NamedPersonalAccessToken(name string) ([]*PersonalAccessToken, error) {
	if e.Edges.namedPersonalAccessToken == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := e.Edges.namedPersonalAccessToken[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (e *Event) appendNamedPersonalAccessToken(name string, edges ...*PersonalAccessToken) {
	if e.Edges.namedPersonalAccessToken == nil {
		e.Edges.namedPersonalAccessToken = make(map[string][]*PersonalAccessToken)
	}
	if len(edges) == 0 {
		e.Edges.namedPersonalAccessToken[name] = []*PersonalAccessToken{}
	} else {
		e.Edges.namedPersonalAccessToken[name] = append(e.Edges.namedPersonalAccessToken[name], edges...)
	}
}

// NamedHush returns the Hush named value or an error if the edge was not
// loaded in eager-loading with this name.
func (e *Event) NamedHush(name string) ([]*Hush, error) {
	if e.Edges.namedHush == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := e.Edges.namedHush[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (e *Event) appendNamedHush(name string, edges ...*Hush) {
	if e.Edges.namedHush == nil {
		e.Edges.namedHush = make(map[string][]*Hush)
	}
	if len(edges) == 0 {
		e.Edges.namedHush[name] = []*Hush{}
	} else {
		e.Edges.namedHush[name] = append(e.Edges.namedHush[name], edges...)
	}
}

// NamedOrgmembership returns the Orgmembership named value or an error if the edge was not
// loaded in eager-loading with this name.
func (e *Event) NamedOrgmembership(name string) ([]*OrgMembership, error) {
	if e.Edges.namedOrgmembership == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := e.Edges.namedOrgmembership[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (e *Event) appendNamedOrgmembership(name string, edges ...*OrgMembership) {
	if e.Edges.namedOrgmembership == nil {
		e.Edges.namedOrgmembership = make(map[string][]*OrgMembership)
	}
	if len(edges) == 0 {
		e.Edges.namedOrgmembership[name] = []*OrgMembership{}
	} else {
		e.Edges.namedOrgmembership[name] = append(e.Edges.namedOrgmembership[name], edges...)
	}
}

// NamedGroupmembership returns the Groupmembership named value or an error if the edge was not
// loaded in eager-loading with this name.
func (e *Event) NamedGroupmembership(name string) ([]*GroupMembership, error) {
	if e.Edges.namedGroupmembership == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := e.Edges.namedGroupmembership[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (e *Event) appendNamedGroupmembership(name string, edges ...*GroupMembership) {
	if e.Edges.namedGroupmembership == nil {
		e.Edges.namedGroupmembership = make(map[string][]*GroupMembership)
	}
	if len(edges) == 0 {
		e.Edges.namedGroupmembership[name] = []*GroupMembership{}
	} else {
		e.Edges.namedGroupmembership[name] = append(e.Edges.namedGroupmembership[name], edges...)
	}
}

// NamedSubscriber returns the Subscriber named value or an error if the edge was not
// loaded in eager-loading with this name.
func (e *Event) NamedSubscriber(name string) ([]*Subscriber, error) {
	if e.Edges.namedSubscriber == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := e.Edges.namedSubscriber[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (e *Event) appendNamedSubscriber(name string, edges ...*Subscriber) {
	if e.Edges.namedSubscriber == nil {
		e.Edges.namedSubscriber = make(map[string][]*Subscriber)
	}
	if len(edges) == 0 {
		e.Edges.namedSubscriber[name] = []*Subscriber{}
	} else {
		e.Edges.namedSubscriber[name] = append(e.Edges.namedSubscriber[name], edges...)
	}
}

// NamedFile returns the File named value or an error if the edge was not
// loaded in eager-loading with this name.
func (e *Event) NamedFile(name string) ([]*File, error) {
	if e.Edges.namedFile == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := e.Edges.namedFile[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (e *Event) appendNamedFile(name string, edges ...*File) {
	if e.Edges.namedFile == nil {
		e.Edges.namedFile = make(map[string][]*File)
	}
	if len(edges) == 0 {
		e.Edges.namedFile[name] = []*File{}
	} else {
		e.Edges.namedFile[name] = append(e.Edges.namedFile[name], edges...)
	}
}

// NamedOrgsubscription returns the Orgsubscription named value or an error if the edge was not
// loaded in eager-loading with this name.
func (e *Event) NamedOrgsubscription(name string) ([]*OrgSubscription, error) {
	if e.Edges.namedOrgsubscription == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := e.Edges.namedOrgsubscription[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (e *Event) appendNamedOrgsubscription(name string, edges ...*OrgSubscription) {
	if e.Edges.namedOrgsubscription == nil {
		e.Edges.namedOrgsubscription = make(map[string][]*OrgSubscription)
	}
	if len(edges) == 0 {
		e.Edges.namedOrgsubscription[name] = []*OrgSubscription{}
	} else {
		e.Edges.namedOrgsubscription[name] = append(e.Edges.namedOrgsubscription[name], edges...)
	}
}

// Events is a parsable slice of Event.
type Events []*Event
