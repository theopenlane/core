package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

// DateTime is a custom GraphQL scalar that converts to/from time.Time
type DateTime time.Time

const (
	dateLayout    = "2006-01-02"
	isoDateLayout = time.RFC3339
)

var (
	ErrUnsupportedDateTimeType = errors.New("unsupported time format")
	ErrInvalidTimeType         = errors.New("invalid date format, expected YYYY-MM-DD or full ISO8601")
)

// Ensure DateTime implements the Valuer, Scanner, and Marshal interfaces
var _ driver.Valuer = (*DateTime)(nil)
var _ sql.Scanner = (*DateTime)(nil)
var _ encoding.TextMarshaler = (*DateTime)(nil)
var _ encoding.TextUnmarshaler = (*DateTime)(nil)
var _ json.Marshaler = DateTime{}
var _ json.Unmarshaler = (*DateTime)(nil)

// Scan implements the sql.Scanner interface for DateTime
func (d *DateTime) Scan(value interface{}) error {
	if value == nil {
		value = time.Time{} // Handle nil value as zero time
	}

	switch v := value.(type) {
	case time.Time:
		*d = DateTime(v)
		return nil
	default:
		return ErrUnsupportedDateTimeType
	}
}

// Value implements the driver.Valuer interface for DateTime
func (d DateTime) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}

	return time.Time(d), nil
}

// IsZero checks if the DateTime is zero (equivalent to time.Time.IsZero)
func (d DateTime) IsZero() bool {
	t := time.Time(d)

	return t.IsZero()
}

// UnmarshalCSV allows the DateTime to accept both "YYYY-MM-DD" and "YYYY-MM-DDTHH:MM:SSZ"
func (d *DateTime) UnmarshalCSV(s string) error {
	if s == "" {
		*d = DateTime{}
		return nil
	}

	if t, err := time.Parse(isoDateLayout, s); err == nil {
		*d = DateTime(t)
		return nil
	}

	if t, err := time.Parse(dateLayout, s); err == nil {
		*d = DateTime(t)
		return nil
	}

	return ErrUnsupportedDateTimeType
}

// UnmarshalGQL allows the DateTime to accept both "YYYY-MM-DD" and "YYYY-MM-DDTHH:MM:SSZ"
func (d *DateTime) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return ErrUnsupportedDateTimeType
	}

	if str == "" {
		*d = DateTime{}
		return nil
	}

	if t, err := time.Parse(isoDateLayout, str); err == nil {
		*d = DateTime(t)
		return nil
	}

	if t, err := time.Parse(dateLayout, str); err == nil {
		*d = DateTime(t)
		return nil
	}

	return ErrInvalidTimeType
}

// MarshalGQL writes the datetime as "YYYY-MM-DD"
func (d DateTime) MarshalGQL(w io.Writer) {
	t := time.Time(d)
	if t.IsZero() {
		_, _ = io.WriteString(w, `""`)
		return
	}

	formatted := fmt.Sprintf("%q", t.Format(isoDateLayout))
	_, _ = io.WriteString(w, formatted)
}

// UnmarshalText parses the DateTime from a byte slice
// this function is used by the cursor pagination to correctly parse the date from the cursor string
func (d *DateTime) UnmarshalText(b []byte) error {
	if len(b) == 0 {
		*d = DateTime{}
		return nil
	}

	s := string(b)

	t, err := time.Parse(isoDateLayout, s)
	if err == nil {
		*d = DateTime(t)
		return nil
	}

	t, err = time.Parse(dateLayout, s)
	if err == nil {
		*d = DateTime(t)
		return nil
	}

	return ErrUnsupportedDateTimeType
}

// MarshalText formats the DateTime as "YYYY-MM-DD" for text representation
// this function is used by the cursor pagination to correctly format the date into the cursor string
func (d DateTime) MarshalText() ([]byte, error) {
	if d.IsZero() {
		return nil, nil
	}

	t := time.Time(d)

	return []byte(t.Format(isoDateLayout)), nil
}

// UnmarshalJSON parses the DateTime from a JSON string
// it accepts both "YYYY-MM-DD" and "YYYY-MM-DDTHH:MM:SSZ" formats
// and returns an error if the format is invalid
func (d *DateTime) UnmarshalJSON(b []byte) error {
	if string(b) == "" {
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	t, err := time.Parse(isoDateLayout, s)
	if err == nil {
		*d = DateTime(t)
		return nil
	}

	t, err = time.Parse(dateLayout, s)
	if err == nil {
		*d = DateTime(t)
		return nil
	}

	return ErrUnsupportedDateTimeType
}

// MarshalJSON formats the DateTime as a JSON string
func (d DateTime) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte(""), nil
	}

	t := time.Time(d)
	s := t.Format(isoDateLayout)

	return json.Marshal(s)
}

// String formats the given datetime into a human readable version
func (d DateTime) String() string {
	t := time.Time(d)
	if t.IsZero() {
		return ""
	}

	formatted := t.Format(isoDateLayout)

	return formatted
}

// ToDateTime converts a string to a DateTime pointer.
// It accepts both "YYYY-MM-DD" and "YYYY-MM-DDTHH:MM:SSZ" formats.
// Returns an error if the string is empty or in an invalid format.
func ToDateTime(s string) (*DateTime, error) {
	if s == "" {
		return nil, ErrInvalidTimeType
	}

	if t, err := time.Parse(isoDateLayout, s); err == nil {
		dt := DateTime(t)
		return &dt, nil
	}

	if t, err := time.Parse(dateLayout, s); err == nil {
		dt := DateTime(t)
		return &dt, nil
	}

	return nil, ErrInvalidTimeType
}
