// Code generated by ent, DO NOT EDIT.

package generated

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/pkg/enums"
)

// OrgMembership is the model entity for the OrgMembership schema.
type OrgMembership struct {
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
	// Role holds the value of the "role" field.
	Role enums.Role `json:"role,omitempty"`
	// OrganizationID holds the value of the "organization_id" field.
	OrganizationID string `json:"organization_id,omitempty"`
	// UserID holds the value of the "user_id" field.
	UserID string `json:"user_id,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the OrgMembershipQuery when eager-loading is set.
	Edges        OrgMembershipEdges `json:"edges"`
	selectValues sql.SelectValues
}

// OrgMembershipEdges holds the relations/edges for other nodes in the graph.
type OrgMembershipEdges struct {
	// Organization holds the value of the organization edge.
	Organization *Organization `json:"organization,omitempty"`
	// User holds the value of the user edge.
	User *User `json:"user,omitempty"`
	// Events holds the value of the events edge.
	Events []*Event `json:"events,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [3]bool
	// totalCount holds the count of the edges above.
	totalCount [3]map[string]int

	namedEvents map[string][]*Event
}

// OrganizationOrErr returns the Organization value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e OrgMembershipEdges) OrganizationOrErr() (*Organization, error) {
	if e.Organization != nil {
		return e.Organization, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: organization.Label}
	}
	return nil, &NotLoadedError{edge: "organization"}
}

// UserOrErr returns the User value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e OrgMembershipEdges) UserOrErr() (*User, error) {
	if e.User != nil {
		return e.User, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: user.Label}
	}
	return nil, &NotLoadedError{edge: "user"}
}

// EventsOrErr returns the Events value or an error if the edge
// was not loaded in eager-loading.
func (e OrgMembershipEdges) EventsOrErr() ([]*Event, error) {
	if e.loadedTypes[2] {
		return e.Events, nil
	}
	return nil, &NotLoadedError{edge: "events"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*OrgMembership) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case orgmembership.FieldID, orgmembership.FieldCreatedBy, orgmembership.FieldUpdatedBy, orgmembership.FieldRole, orgmembership.FieldOrganizationID, orgmembership.FieldUserID:
			values[i] = new(sql.NullString)
		case orgmembership.FieldCreatedAt, orgmembership.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the OrgMembership fields.
func (om *OrgMembership) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case orgmembership.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				om.ID = value.String
			}
		case orgmembership.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				om.CreatedAt = value.Time
			}
		case orgmembership.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				om.UpdatedAt = value.Time
			}
		case orgmembership.FieldCreatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_by", values[i])
			} else if value.Valid {
				om.CreatedBy = value.String
			}
		case orgmembership.FieldUpdatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_by", values[i])
			} else if value.Valid {
				om.UpdatedBy = value.String
			}
		case orgmembership.FieldRole:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field role", values[i])
			} else if value.Valid {
				om.Role = enums.Role(value.String)
			}
		case orgmembership.FieldOrganizationID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field organization_id", values[i])
			} else if value.Valid {
				om.OrganizationID = value.String
			}
		case orgmembership.FieldUserID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field user_id", values[i])
			} else if value.Valid {
				om.UserID = value.String
			}
		default:
			om.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the OrgMembership.
// This includes values selected through modifiers, order, etc.
func (om *OrgMembership) Value(name string) (ent.Value, error) {
	return om.selectValues.Get(name)
}

// QueryOrganization queries the "organization" edge of the OrgMembership entity.
func (om *OrgMembership) QueryOrganization() *OrganizationQuery {
	return NewOrgMembershipClient(om.config).QueryOrganization(om)
}

// QueryUser queries the "user" edge of the OrgMembership entity.
func (om *OrgMembership) QueryUser() *UserQuery {
	return NewOrgMembershipClient(om.config).QueryUser(om)
}

// QueryEvents queries the "events" edge of the OrgMembership entity.
func (om *OrgMembership) QueryEvents() *EventQuery {
	return NewOrgMembershipClient(om.config).QueryEvents(om)
}

// Update returns a builder for updating this OrgMembership.
// Note that you need to call OrgMembership.Unwrap() before calling this method if this OrgMembership
// was returned from a transaction, and the transaction was committed or rolled back.
func (om *OrgMembership) Update() *OrgMembershipUpdateOne {
	return NewOrgMembershipClient(om.config).UpdateOne(om)
}

// Unwrap unwraps the OrgMembership entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (om *OrgMembership) Unwrap() *OrgMembership {
	_tx, ok := om.config.driver.(*txDriver)
	if !ok {
		panic("generated: OrgMembership is not a transactional entity")
	}
	om.config.driver = _tx.drv
	return om
}

// String implements the fmt.Stringer.
func (om *OrgMembership) String() string {
	var builder strings.Builder
	builder.WriteString("OrgMembership(")
	builder.WriteString(fmt.Sprintf("id=%v, ", om.ID))
	builder.WriteString("created_at=")
	builder.WriteString(om.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(om.UpdatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("created_by=")
	builder.WriteString(om.CreatedBy)
	builder.WriteString(", ")
	builder.WriteString("updated_by=")
	builder.WriteString(om.UpdatedBy)
	builder.WriteString(", ")
	builder.WriteString("role=")
	builder.WriteString(fmt.Sprintf("%v", om.Role))
	builder.WriteString(", ")
	builder.WriteString("organization_id=")
	builder.WriteString(om.OrganizationID)
	builder.WriteString(", ")
	builder.WriteString("user_id=")
	builder.WriteString(om.UserID)
	builder.WriteByte(')')
	return builder.String()
}

// NamedEvents returns the Events named value or an error if the edge was not
// loaded in eager-loading with this name.
func (om *OrgMembership) NamedEvents(name string) ([]*Event, error) {
	if om.Edges.namedEvents == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := om.Edges.namedEvents[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (om *OrgMembership) appendNamedEvents(name string, edges ...*Event) {
	if om.Edges.namedEvents == nil {
		om.Edges.namedEvents = make(map[string][]*Event)
	}
	if len(edges) == 0 {
		om.Edges.namedEvents[name] = []*Event{}
	} else {
		om.Edges.namedEvents[name] = append(om.Edges.namedEvents[name], edges...)
	}
}

// OrgMemberships is a parsable slice of OrgMembership.
type OrgMemberships []*OrgMembership
