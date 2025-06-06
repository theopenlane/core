// Code generated by ent, DO NOT EDIT.

package generated

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/core/internal/ent/generated/orgmembershiphistory"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/entx/history"
)

// OrgMembershipHistory is the model entity for the OrgMembershipHistory schema.
type OrgMembershipHistory struct {
	config `json:"-"`
	// ID of the ent.
	ID string `json:"id,omitempty"`
	// HistoryTime holds the value of the "history_time" field.
	HistoryTime time.Time `json:"history_time,omitempty"`
	// Ref holds the value of the "ref" field.
	Ref string `json:"ref,omitempty"`
	// Operation holds the value of the "operation" field.
	Operation history.OpType `json:"operation,omitempty"`
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
	UserID       string `json:"user_id,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*OrgMembershipHistory) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case orgmembershiphistory.FieldOperation:
			values[i] = new(history.OpType)
		case orgmembershiphistory.FieldID, orgmembershiphistory.FieldRef, orgmembershiphistory.FieldCreatedBy, orgmembershiphistory.FieldUpdatedBy, orgmembershiphistory.FieldRole, orgmembershiphistory.FieldOrganizationID, orgmembershiphistory.FieldUserID:
			values[i] = new(sql.NullString)
		case orgmembershiphistory.FieldHistoryTime, orgmembershiphistory.FieldCreatedAt, orgmembershiphistory.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the OrgMembershipHistory fields.
func (omh *OrgMembershipHistory) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case orgmembershiphistory.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				omh.ID = value.String
			}
		case orgmembershiphistory.FieldHistoryTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field history_time", values[i])
			} else if value.Valid {
				omh.HistoryTime = value.Time
			}
		case orgmembershiphistory.FieldRef:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field ref", values[i])
			} else if value.Valid {
				omh.Ref = value.String
			}
		case orgmembershiphistory.FieldOperation:
			if value, ok := values[i].(*history.OpType); !ok {
				return fmt.Errorf("unexpected type %T for field operation", values[i])
			} else if value != nil {
				omh.Operation = *value
			}
		case orgmembershiphistory.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				omh.CreatedAt = value.Time
			}
		case orgmembershiphistory.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				omh.UpdatedAt = value.Time
			}
		case orgmembershiphistory.FieldCreatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_by", values[i])
			} else if value.Valid {
				omh.CreatedBy = value.String
			}
		case orgmembershiphistory.FieldUpdatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_by", values[i])
			} else if value.Valid {
				omh.UpdatedBy = value.String
			}
		case orgmembershiphistory.FieldRole:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field role", values[i])
			} else if value.Valid {
				omh.Role = enums.Role(value.String)
			}
		case orgmembershiphistory.FieldOrganizationID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field organization_id", values[i])
			} else if value.Valid {
				omh.OrganizationID = value.String
			}
		case orgmembershiphistory.FieldUserID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field user_id", values[i])
			} else if value.Valid {
				omh.UserID = value.String
			}
		default:
			omh.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the OrgMembershipHistory.
// This includes values selected through modifiers, order, etc.
func (omh *OrgMembershipHistory) Value(name string) (ent.Value, error) {
	return omh.selectValues.Get(name)
}

// Update returns a builder for updating this OrgMembershipHistory.
// Note that you need to call OrgMembershipHistory.Unwrap() before calling this method if this OrgMembershipHistory
// was returned from a transaction, and the transaction was committed or rolled back.
func (omh *OrgMembershipHistory) Update() *OrgMembershipHistoryUpdateOne {
	return NewOrgMembershipHistoryClient(omh.config).UpdateOne(omh)
}

// Unwrap unwraps the OrgMembershipHistory entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (omh *OrgMembershipHistory) Unwrap() *OrgMembershipHistory {
	_tx, ok := omh.config.driver.(*txDriver)
	if !ok {
		panic("generated: OrgMembershipHistory is not a transactional entity")
	}
	omh.config.driver = _tx.drv
	return omh
}

// String implements the fmt.Stringer.
func (omh *OrgMembershipHistory) String() string {
	var builder strings.Builder
	builder.WriteString("OrgMembershipHistory(")
	builder.WriteString(fmt.Sprintf("id=%v, ", omh.ID))
	builder.WriteString("history_time=")
	builder.WriteString(omh.HistoryTime.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("ref=")
	builder.WriteString(omh.Ref)
	builder.WriteString(", ")
	builder.WriteString("operation=")
	builder.WriteString(fmt.Sprintf("%v", omh.Operation))
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(omh.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(omh.UpdatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("created_by=")
	builder.WriteString(omh.CreatedBy)
	builder.WriteString(", ")
	builder.WriteString("updated_by=")
	builder.WriteString(omh.UpdatedBy)
	builder.WriteString(", ")
	builder.WriteString("role=")
	builder.WriteString(fmt.Sprintf("%v", omh.Role))
	builder.WriteString(", ")
	builder.WriteString("organization_id=")
	builder.WriteString(omh.OrganizationID)
	builder.WriteString(", ")
	builder.WriteString("user_id=")
	builder.WriteString(omh.UserID)
	builder.WriteByte(')')
	return builder.String()
}

// OrgMembershipHistories is a parsable slice of OrgMembershipHistory.
type OrgMembershipHistories []*OrgMembershipHistory
