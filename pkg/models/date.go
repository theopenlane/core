package models

import (
	"database/sql/driver"
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

func (d *DateTime) Scan(value any) error {
	if value == nil {
		return nil
	}

	t, ok := value.(time.Time)
	if !ok {
		return ErrUnsupportedDateTimeType
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

	return ErrUnsupportedDateTimeType
}

// UnmarshalGQL allows the DateTime to accept both "YYYY-MM-DD" and "YYYY-MM-DDTHH:MM:SSZ"
func (d *DateTime) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return ErrUnsupportedDateTimeType
	}

	if t, err := time.Parse(dateLayout, str); err == nil {
		*d = DateTime(t)
		return nil
	}

	if t, err := time.Parse(isoDateLayout, str); err == nil {
		*d = DateTime(t)
		return nil
	}

	return ErrInvalidTimeType
}

// MarshalGQL writes the datetime as "YYYY-MM-DD"
func (d DateTime) MarshalGQL(w io.Writer) {
	t := time.Time(d)
	_, _ = io.WriteString(w, fmt.Sprintf("%q", t.Format(dateLayout)))
}

func (d DateTime) String() string {
	return time.Time(d).Format(isoDateLayout)
}

func ToDateTime(s string) (*DateTime, error) {
	if s == "" {
		return nil, nil
	}

	if t, err := time.Parse(dateLayout, s); err == nil {
		dt := DateTime(t)
		return &dt, nil
	}

	if t, err := time.Parse(isoDateLayout, s); err == nil {
		dt := DateTime(t)
		return &dt, nil
	}

	return nil, ErrInvalidTimeType
}
