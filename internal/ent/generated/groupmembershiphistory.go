// Code generated by ent, DO NOT EDIT.

package generated

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/core/internal/ent/generated/groupmembershiphistory"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/entx/history"
)

// GroupMembershipHistory is the model entity for the GroupMembershipHistory schema.
type GroupMembershipHistory struct {
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
	// GroupID holds the value of the "group_id" field.
	GroupID string `json:"group_id,omitempty"`
	// UserID holds the value of the "user_id" field.
	UserID       string `json:"user_id,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*GroupMembershipHistory) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case groupmembershiphistory.FieldOperation:
			values[i] = new(history.OpType)
		case groupmembershiphistory.FieldID, groupmembershiphistory.FieldRef, groupmembershiphistory.FieldCreatedBy, groupmembershiphistory.FieldUpdatedBy, groupmembershiphistory.FieldRole, groupmembershiphistory.FieldGroupID, groupmembershiphistory.FieldUserID:
			values[i] = new(sql.NullString)
		case groupmembershiphistory.FieldHistoryTime, groupmembershiphistory.FieldCreatedAt, groupmembershiphistory.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the GroupMembershipHistory fields.
func (gmh *GroupMembershipHistory) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case groupmembershiphistory.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				gmh.ID = value.String
			}
		case groupmembershiphistory.FieldHistoryTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field history_time", values[i])
			} else if value.Valid {
				gmh.HistoryTime = value.Time
			}
		case groupmembershiphistory.FieldRef:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field ref", values[i])
			} else if value.Valid {
				gmh.Ref = value.String
			}
		case groupmembershiphistory.FieldOperation:
			if value, ok := values[i].(*history.OpType); !ok {
				return fmt.Errorf("unexpected type %T for field operation", values[i])
			} else if value != nil {
				gmh.Operation = *value
			}
		case groupmembershiphistory.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				gmh.CreatedAt = value.Time
			}
		case groupmembershiphistory.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				gmh.UpdatedAt = value.Time
			}
		case groupmembershiphistory.FieldCreatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_by", values[i])
			} else if value.Valid {
				gmh.CreatedBy = value.String
			}
		case groupmembershiphistory.FieldUpdatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_by", values[i])
			} else if value.Valid {
				gmh.UpdatedBy = value.String
			}
		case groupmembershiphistory.FieldRole:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field role", values[i])
			} else if value.Valid {
				gmh.Role = enums.Role(value.String)
			}
		case groupmembershiphistory.FieldGroupID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field group_id", values[i])
			} else if value.Valid {
				gmh.GroupID = value.String
			}
		case groupmembershiphistory.FieldUserID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field user_id", values[i])
			} else if value.Valid {
				gmh.UserID = value.String
			}
		default:
			gmh.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the GroupMembershipHistory.
// This includes values selected through modifiers, order, etc.
func (gmh *GroupMembershipHistory) Value(name string) (ent.Value, error) {
	return gmh.selectValues.Get(name)
}

// Update returns a builder for updating this GroupMembershipHistory.
// Note that you need to call GroupMembershipHistory.Unwrap() before calling this method if this GroupMembershipHistory
// was returned from a transaction, and the transaction was committed or rolled back.
func (gmh *GroupMembershipHistory) Update() *GroupMembershipHistoryUpdateOne {
	return NewGroupMembershipHistoryClient(gmh.config).UpdateOne(gmh)
}

// Unwrap unwraps the GroupMembershipHistory entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (gmh *GroupMembershipHistory) Unwrap() *GroupMembershipHistory {
	_tx, ok := gmh.config.driver.(*txDriver)
	if !ok {
		panic("generated: GroupMembershipHistory is not a transactional entity")
	}
	gmh.config.driver = _tx.drv
	return gmh
}

// String implements the fmt.Stringer.
func (gmh *GroupMembershipHistory) String() string {
	var builder strings.Builder
	builder.WriteString("GroupMembershipHistory(")
	builder.WriteString(fmt.Sprintf("id=%v, ", gmh.ID))
	builder.WriteString("history_time=")
	builder.WriteString(gmh.HistoryTime.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("ref=")
	builder.WriteString(gmh.Ref)
	builder.WriteString(", ")
	builder.WriteString("operation=")
	builder.WriteString(fmt.Sprintf("%v", gmh.Operation))
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(gmh.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(gmh.UpdatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("created_by=")
	builder.WriteString(gmh.CreatedBy)
	builder.WriteString(", ")
	builder.WriteString("updated_by=")
	builder.WriteString(gmh.UpdatedBy)
	builder.WriteString(", ")
	builder.WriteString("role=")
	builder.WriteString(fmt.Sprintf("%v", gmh.Role))
	builder.WriteString(", ")
	builder.WriteString("group_id=")
	builder.WriteString(gmh.GroupID)
	builder.WriteString(", ")
	builder.WriteString("user_id=")
	builder.WriteString(gmh.UserID)
	builder.WriteByte(')')
	return builder.String()
}

// GroupMembershipHistories is a parsable slice of GroupMembershipHistory.
type GroupMembershipHistories []*GroupMembershipHistory
