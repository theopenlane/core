package models

import (
	"database/sql/driver"
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

func (d *DateTime) Scan(value any) error {
	if value == nil {
		return nil
	}

	t, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("cannot scan type %T into DateTime", value)
	}

	*d = DateTime(t)
	return nil
}

func (d DateTime) Value() (driver.Value, error) {
	return time.Time(d), nil
}

func (d *DateTime) UnmarshalCSV(s string) error {
	if s == "" {
		*d = DateTime{}
		return nil
	}

	if t, err := time.Parse(dateLayout, s); err == nil {
		*d = DateTime(t)
		return nil
	}

	if t, err := time.Parse(isoDateLayout, s); err == nil {
		*d = DateTime(t)
		return nil
	}

	return fmt.Errorf("invalid date format: %q", s)
}

// UnmarshalGQL allows the DateTime to accept both "YYYY-MM-DD" and "YYYY-MM-DDTHH:MM:SSZ"
func (d *DateTime) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("date must be a string")
	}

	if t, err := time.Parse(dateLayout, str); err == nil {
		*d = DateTime(t)
		return nil
	}

	if t, err := time.Parse(isoDateLayout, str); err == nil {
		*d = DateTime(t)
		return nil
	}

	return fmt.Errorf("invalid date format, expected YYYY-MM-DD or full ISO8601")
}

// MarshalGQL writes the datetime as "YYYY-MM-DD"
func (d DateTime) MarshalGQL(w io.Writer) {
	t := time.Time(d)
	_, _ = io.WriteString(w, fmt.Sprintf("%q", t.Format(dateLayout)))
}

func ToModelsDateTime(f DateTime) DateTime {
	return DateTime(time.Time(f))
}

func ToModelsDateTimePtr(f *DateTime) *DateTime {
	if f == nil {
		return nil
	}

	t := DateTime(time.Time(*f))
	return &t
}
