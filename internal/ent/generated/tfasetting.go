// Code generated by ent, DO NOT EDIT.

package generated

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/core/internal/ent/generated/tfasetting"
	"github.com/theopenlane/core/internal/ent/generated/user"
)

// TFASetting is the model entity for the TFASetting schema.
type TFASetting struct {
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
	// DeletedAt holds the value of the "deleted_at" field.
	DeletedAt time.Time `json:"deleted_at,omitempty"`
	// DeletedBy holds the value of the "deleted_by" field.
	DeletedBy string `json:"deleted_by,omitempty"`
	// The user id that owns the object
	OwnerID string `json:"owner_id,omitempty"`
	// TFA secret for the user
	TfaSecret *string `json:"tfa_secret,omitempty"`
	// specifies if the TFA device has been verified
	Verified bool `json:"verified,omitempty"`
	// recovery codes for 2fa
	RecoveryCodes []string `json:"recovery_codes,omitempty"`
	// specifies a user may complete authentication by verifying an OTP code delivered through SMS
	PhoneOtpAllowed bool `json:"phone_otp_allowed,omitempty"`
	// specifies a user may complete authentication by verifying an OTP code delivered through email
	EmailOtpAllowed bool `json:"email_otp_allowed,omitempty"`
	// specifies a user may complete authentication by verifying a TOTP code delivered through an authenticator app
	TotpAllowed bool `json:"totp_allowed,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the TFASettingQuery when eager-loading is set.
	Edges        TFASettingEdges `json:"edges"`
	selectValues sql.SelectValues
}

// TFASettingEdges holds the relations/edges for other nodes in the graph.
type TFASettingEdges struct {
	// Owner holds the value of the owner edge.
	Owner *User `json:"owner,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
	// totalCount holds the count of the edges above.
	totalCount [1]map[string]int
}

// OwnerOrErr returns the Owner value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e TFASettingEdges) OwnerOrErr() (*User, error) {
	if e.Owner != nil {
		return e.Owner, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: user.Label}
	}
	return nil, &NotLoadedError{edge: "owner"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*TFASetting) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case tfasetting.FieldRecoveryCodes:
			values[i] = new([]byte)
		case tfasetting.FieldVerified, tfasetting.FieldPhoneOtpAllowed, tfasetting.FieldEmailOtpAllowed, tfasetting.FieldTotpAllowed:
			values[i] = new(sql.NullBool)
		case tfasetting.FieldID, tfasetting.FieldCreatedBy, tfasetting.FieldUpdatedBy, tfasetting.FieldDeletedBy, tfasetting.FieldOwnerID, tfasetting.FieldTfaSecret:
			values[i] = new(sql.NullString)
		case tfasetting.FieldCreatedAt, tfasetting.FieldUpdatedAt, tfasetting.FieldDeletedAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the TFASetting fields.
func (ts *TFASetting) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case tfasetting.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				ts.ID = value.String
			}
		case tfasetting.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				ts.CreatedAt = value.Time
			}
		case tfasetting.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				ts.UpdatedAt = value.Time
			}
		case tfasetting.FieldCreatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_by", values[i])
			} else if value.Valid {
				ts.CreatedBy = value.String
			}
		case tfasetting.FieldUpdatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_by", values[i])
			} else if value.Valid {
				ts.UpdatedBy = value.String
			}
		case tfasetting.FieldDeletedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_at", values[i])
			} else if value.Valid {
				ts.DeletedAt = value.Time
			}
		case tfasetting.FieldDeletedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_by", values[i])
			} else if value.Valid {
				ts.DeletedBy = value.String
			}
		case tfasetting.FieldOwnerID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field owner_id", values[i])
			} else if value.Valid {
				ts.OwnerID = value.String
			}
		case tfasetting.FieldTfaSecret:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field tfa_secret", values[i])
			} else if value.Valid {
				ts.TfaSecret = new(string)
				*ts.TfaSecret = value.String
			}
		case tfasetting.FieldVerified:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field verified", values[i])
			} else if value.Valid {
				ts.Verified = value.Bool
			}
		case tfasetting.FieldRecoveryCodes:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field recovery_codes", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &ts.RecoveryCodes); err != nil {
					return fmt.Errorf("unmarshal field recovery_codes: %w", err)
				}
			}
		case tfasetting.FieldPhoneOtpAllowed:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field phone_otp_allowed", values[i])
			} else if value.Valid {
				ts.PhoneOtpAllowed = value.Bool
			}
		case tfasetting.FieldEmailOtpAllowed:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field email_otp_allowed", values[i])
			} else if value.Valid {
				ts.EmailOtpAllowed = value.Bool
			}
		case tfasetting.FieldTotpAllowed:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field totp_allowed", values[i])
			} else if value.Valid {
				ts.TotpAllowed = value.Bool
			}
		default:
			ts.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the TFASetting.
// This includes values selected through modifiers, order, etc.
func (ts *TFASetting) Value(name string) (ent.Value, error) {
	return ts.selectValues.Get(name)
}

// QueryOwner queries the "owner" edge of the TFASetting entity.
func (ts *TFASetting) QueryOwner() *UserQuery {
	return NewTFASettingClient(ts.config).QueryOwner(ts)
}

// Update returns a builder for updating this TFASetting.
// Note that you need to call TFASetting.Unwrap() before calling this method if this TFASetting
// was returned from a transaction, and the transaction was committed or rolled back.
func (ts *TFASetting) Update() *TFASettingUpdateOne {
	return NewTFASettingClient(ts.config).UpdateOne(ts)
}

// Unwrap unwraps the TFASetting entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (ts *TFASetting) Unwrap() *TFASetting {
	_tx, ok := ts.config.driver.(*txDriver)
	if !ok {
		panic("generated: TFASetting is not a transactional entity")
	}
	ts.config.driver = _tx.drv
	return ts
}

// String implements the fmt.Stringer.
func (ts *TFASetting) String() string {
	var builder strings.Builder
	builder.WriteString("TFASetting(")
	builder.WriteString(fmt.Sprintf("id=%v, ", ts.ID))
	builder.WriteString("created_at=")
	builder.WriteString(ts.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(ts.UpdatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("created_by=")
	builder.WriteString(ts.CreatedBy)
	builder.WriteString(", ")
	builder.WriteString("updated_by=")
	builder.WriteString(ts.UpdatedBy)
	builder.WriteString(", ")
	builder.WriteString("deleted_at=")
	builder.WriteString(ts.DeletedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("deleted_by=")
	builder.WriteString(ts.DeletedBy)
	builder.WriteString(", ")
	builder.WriteString("owner_id=")
	builder.WriteString(ts.OwnerID)
	builder.WriteString(", ")
	if v := ts.TfaSecret; v != nil {
		builder.WriteString("tfa_secret=")
		builder.WriteString(*v)
	}
	builder.WriteString(", ")
	builder.WriteString("verified=")
	builder.WriteString(fmt.Sprintf("%v", ts.Verified))
	builder.WriteString(", ")
	builder.WriteString("recovery_codes=")
	builder.WriteString(fmt.Sprintf("%v", ts.RecoveryCodes))
	builder.WriteString(", ")
	builder.WriteString("phone_otp_allowed=")
	builder.WriteString(fmt.Sprintf("%v", ts.PhoneOtpAllowed))
	builder.WriteString(", ")
	builder.WriteString("email_otp_allowed=")
	builder.WriteString(fmt.Sprintf("%v", ts.EmailOtpAllowed))
	builder.WriteString(", ")
	builder.WriteString("totp_allowed=")
	builder.WriteString(fmt.Sprintf("%v", ts.TotpAllowed))
	builder.WriteByte(')')
	return builder.String()
}

// TFASettings is a parsable slice of TFASetting.
type TFASettings []*TFASetting
