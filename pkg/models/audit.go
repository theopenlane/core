package models

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

// Change represents a change in an entity's field.
type Change struct {
	// FieldName is the name of the field that changed.
	FieldName string
	// Old is the old value of the field.
	Old any
	// New is the new value of the field.
	New any
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (c Change) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, c)
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (c *Change) UnmarshalGQL(v interface{}) error {
	return unmarshalGQLJSON(v, c)
}

// Properties by which AuditLog connections can be ordered.
type AuditLogOrderField string

const (
	AuditLogOrderFieldHistoryTime AuditLogOrderField = "history_time"
)

// AllAuditLogOrderField contains all valid AuditLogOrderField values.
var AllAuditLogOrderField = []AuditLogOrderField{
	AuditLogOrderFieldHistoryTime,
}

// IsValid checks if the AuditLogOrderField is valid.
func (e AuditLogOrderField) IsValid() bool {
	return e == AuditLogOrderFieldHistoryTime
}

// String returns the string representation of the AuditLogOrderField.
func (e AuditLogOrderField) String() string {
	return string(e)
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen.
func (e *AuditLogOrderField) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings") //nolint:err113
	}

	*e = AuditLogOrderField(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid AuditLogOrderField", str) //nolint:err113
	}

	return nil
}

// MarshalGQL implements the Marshaler interface for gqlgen.
func (e AuditLogOrderField) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

// UnmarshalJSON implements the json.Unmarshaler interface for AuditLogOrderField.
func (e *AuditLogOrderField) UnmarshalJSON(b []byte) error {
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}

	return e.UnmarshalGQL(s)
}

// MarshalJSON implements the json.Marshaler interface for AuditLogOrderField.
func (e AuditLogOrderField) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	e.MarshalGQL(&buf)

	return buf.Bytes(), nil
}
