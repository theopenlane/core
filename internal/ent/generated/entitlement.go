// Code generated by ent, DO NOT EDIT.

package generated

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/core/internal/ent/generated/entitlement"
	"github.com/theopenlane/core/internal/ent/generated/entitlementplan"
	"github.com/theopenlane/core/internal/ent/generated/organization"
)

// Entitlement is the model entity for the Entitlement schema.
type Entitlement struct {
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
	// MappingID holds the value of the "mapping_id" field.
	MappingID string `json:"mapping_id,omitempty"`
	// tags associated with the object
	Tags []string `json:"tags,omitempty"`
	// DeletedAt holds the value of the "deleted_at" field.
	DeletedAt time.Time `json:"deleted_at,omitempty"`
	// DeletedBy holds the value of the "deleted_by" field.
	DeletedBy string `json:"deleted_by,omitempty"`
	// The organization id that owns the object
	OwnerID string `json:"owner_id,omitempty"`
	// the plan to which the entitlement belongs
	PlanID string `json:"plan_id,omitempty"`
	// the organization to which the entitlement belongs
	OrganizationID string `json:"organization_id,omitempty"`
	// used to store references to external systems, e.g. Stripe
	ExternalCustomerID string `json:"external_customer_id,omitempty"`
	// used to store references to external systems, e.g. Stripe
	ExternalSubscriptionID string `json:"external_subscription_id,omitempty"`
	// whether or not the customers entitlement expires - expires_at will show the time
	Expires bool `json:"expires,omitempty"`
	// the time at which a customer's entitlement will expire, e.g. they've cancelled but paid through the end of the month
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	// whether or not the customer has cancelled their entitlement - usually used in conjunction with expires and expires at
	Cancelled bool `json:"cancelled,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the EntitlementQuery when eager-loading is set.
	Edges        EntitlementEdges `json:"edges"`
	selectValues sql.SelectValues
}

// EntitlementEdges holds the relations/edges for other nodes in the graph.
type EntitlementEdges struct {
	// Owner holds the value of the owner edge.
	Owner *Organization `json:"owner,omitempty"`
	// Plan holds the value of the plan edge.
	Plan *EntitlementPlan `json:"plan,omitempty"`
	// Organization holds the value of the organization edge.
	Organization *Organization `json:"organization,omitempty"`
	// Events holds the value of the events edge.
	Events []*Event `json:"events,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [4]bool
	// totalCount holds the count of the edges above.
	totalCount [4]map[string]int

	namedEvents map[string][]*Event
}

// OwnerOrErr returns the Owner value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e EntitlementEdges) OwnerOrErr() (*Organization, error) {
	if e.Owner != nil {
		return e.Owner, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: organization.Label}
	}
	return nil, &NotLoadedError{edge: "owner"}
}

// PlanOrErr returns the Plan value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e EntitlementEdges) PlanOrErr() (*EntitlementPlan, error) {
	if e.Plan != nil {
		return e.Plan, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: entitlementplan.Label}
	}
	return nil, &NotLoadedError{edge: "plan"}
}

// OrganizationOrErr returns the Organization value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e EntitlementEdges) OrganizationOrErr() (*Organization, error) {
	if e.Organization != nil {
		return e.Organization, nil
	} else if e.loadedTypes[2] {
		return nil, &NotFoundError{label: organization.Label}
	}
	return nil, &NotLoadedError{edge: "organization"}
}

// EventsOrErr returns the Events value or an error if the edge
// was not loaded in eager-loading.
func (e EntitlementEdges) EventsOrErr() ([]*Event, error) {
	if e.loadedTypes[3] {
		return e.Events, nil
	}
	return nil, &NotLoadedError{edge: "events"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Entitlement) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case entitlement.FieldTags:
			values[i] = new([]byte)
		case entitlement.FieldExpires, entitlement.FieldCancelled:
			values[i] = new(sql.NullBool)
		case entitlement.FieldID, entitlement.FieldCreatedBy, entitlement.FieldUpdatedBy, entitlement.FieldMappingID, entitlement.FieldDeletedBy, entitlement.FieldOwnerID, entitlement.FieldPlanID, entitlement.FieldOrganizationID, entitlement.FieldExternalCustomerID, entitlement.FieldExternalSubscriptionID:
			values[i] = new(sql.NullString)
		case entitlement.FieldCreatedAt, entitlement.FieldUpdatedAt, entitlement.FieldDeletedAt, entitlement.FieldExpiresAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Entitlement fields.
func (e *Entitlement) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case entitlement.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				e.ID = value.String
			}
		case entitlement.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				e.CreatedAt = value.Time
			}
		case entitlement.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				e.UpdatedAt = value.Time
			}
		case entitlement.FieldCreatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_by", values[i])
			} else if value.Valid {
				e.CreatedBy = value.String
			}
		case entitlement.FieldUpdatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_by", values[i])
			} else if value.Valid {
				e.UpdatedBy = value.String
			}
		case entitlement.FieldMappingID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field mapping_id", values[i])
			} else if value.Valid {
				e.MappingID = value.String
			}
		case entitlement.FieldTags:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field tags", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &e.Tags); err != nil {
					return fmt.Errorf("unmarshal field tags: %w", err)
				}
			}
		case entitlement.FieldDeletedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_at", values[i])
			} else if value.Valid {
				e.DeletedAt = value.Time
			}
		case entitlement.FieldDeletedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_by", values[i])
			} else if value.Valid {
				e.DeletedBy = value.String
			}
		case entitlement.FieldOwnerID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field owner_id", values[i])
			} else if value.Valid {
				e.OwnerID = value.String
			}
		case entitlement.FieldPlanID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field plan_id", values[i])
			} else if value.Valid {
				e.PlanID = value.String
			}
		case entitlement.FieldOrganizationID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field organization_id", values[i])
			} else if value.Valid {
				e.OrganizationID = value.String
			}
		case entitlement.FieldExternalCustomerID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field external_customer_id", values[i])
			} else if value.Valid {
				e.ExternalCustomerID = value.String
			}
		case entitlement.FieldExternalSubscriptionID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field external_subscription_id", values[i])
			} else if value.Valid {
				e.ExternalSubscriptionID = value.String
			}
		case entitlement.FieldExpires:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field expires", values[i])
			} else if value.Valid {
				e.Expires = value.Bool
			}
		case entitlement.FieldExpiresAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field expires_at", values[i])
			} else if value.Valid {
				e.ExpiresAt = new(time.Time)
				*e.ExpiresAt = value.Time
			}
		case entitlement.FieldCancelled:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field cancelled", values[i])
			} else if value.Valid {
				e.Cancelled = value.Bool
			}
		default:
			e.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Entitlement.
// This includes values selected through modifiers, order, etc.
func (e *Entitlement) Value(name string) (ent.Value, error) {
	return e.selectValues.Get(name)
}

// QueryOwner queries the "owner" edge of the Entitlement entity.
func (e *Entitlement) QueryOwner() *OrganizationQuery {
	return NewEntitlementClient(e.config).QueryOwner(e)
}

// QueryPlan queries the "plan" edge of the Entitlement entity.
func (e *Entitlement) QueryPlan() *EntitlementPlanQuery {
	return NewEntitlementClient(e.config).QueryPlan(e)
}

// QueryOrganization queries the "organization" edge of the Entitlement entity.
func (e *Entitlement) QueryOrganization() *OrganizationQuery {
	return NewEntitlementClient(e.config).QueryOrganization(e)
}

// QueryEvents queries the "events" edge of the Entitlement entity.
func (e *Entitlement) QueryEvents() *EventQuery {
	return NewEntitlementClient(e.config).QueryEvents(e)
}

// Update returns a builder for updating this Entitlement.
// Note that you need to call Entitlement.Unwrap() before calling this method if this Entitlement
// was returned from a transaction, and the transaction was committed or rolled back.
func (e *Entitlement) Update() *EntitlementUpdateOne {
	return NewEntitlementClient(e.config).UpdateOne(e)
}

// Unwrap unwraps the Entitlement entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (e *Entitlement) Unwrap() *Entitlement {
	_tx, ok := e.config.driver.(*txDriver)
	if !ok {
		panic("generated: Entitlement is not a transactional entity")
	}
	e.config.driver = _tx.drv
	return e
}

// String implements the fmt.Stringer.
func (e *Entitlement) String() string {
	var builder strings.Builder
	builder.WriteString("Entitlement(")
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
	builder.WriteString("mapping_id=")
	builder.WriteString(e.MappingID)
	builder.WriteString(", ")
	builder.WriteString("tags=")
	builder.WriteString(fmt.Sprintf("%v", e.Tags))
	builder.WriteString(", ")
	builder.WriteString("deleted_at=")
	builder.WriteString(e.DeletedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("deleted_by=")
	builder.WriteString(e.DeletedBy)
	builder.WriteString(", ")
	builder.WriteString("owner_id=")
	builder.WriteString(e.OwnerID)
	builder.WriteString(", ")
	builder.WriteString("plan_id=")
	builder.WriteString(e.PlanID)
	builder.WriteString(", ")
	builder.WriteString("organization_id=")
	builder.WriteString(e.OrganizationID)
	builder.WriteString(", ")
	builder.WriteString("external_customer_id=")
	builder.WriteString(e.ExternalCustomerID)
	builder.WriteString(", ")
	builder.WriteString("external_subscription_id=")
	builder.WriteString(e.ExternalSubscriptionID)
	builder.WriteString(", ")
	builder.WriteString("expires=")
	builder.WriteString(fmt.Sprintf("%v", e.Expires))
	builder.WriteString(", ")
	if v := e.ExpiresAt; v != nil {
		builder.WriteString("expires_at=")
		builder.WriteString(v.Format(time.ANSIC))
	}
	builder.WriteString(", ")
	builder.WriteString("cancelled=")
	builder.WriteString(fmt.Sprintf("%v", e.Cancelled))
	builder.WriteByte(')')
	return builder.String()
}

// NamedEvents returns the Events named value or an error if the edge was not
// loaded in eager-loading with this name.
func (e *Entitlement) NamedEvents(name string) ([]*Event, error) {
	if e.Edges.namedEvents == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := e.Edges.namedEvents[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (e *Entitlement) appendNamedEvents(name string, edges ...*Event) {
	if e.Edges.namedEvents == nil {
		e.Edges.namedEvents = make(map[string][]*Event)
	}
	if len(edges) == 0 {
		e.Edges.namedEvents[name] = []*Event{}
	} else {
		e.Edges.namedEvents[name] = append(e.Edges.namedEvents[name], edges...)
	}
}

// Entitlements is a parsable slice of Entitlement.
type Entitlements []*Entitlement