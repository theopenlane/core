package models

import (
	"database/sql/driver"
	"fmt"
	"io"

	"github.com/google/uuid"
)

// AAGUID is a custom type representing an authenticator attestation uuid.
type AAGUID []byte

func (a AAGUID) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}
	return []byte(a), nil
}

func (a *AAGUID) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*a = AAGUID(v)
		return nil
	case string:
		u, err := uuid.Parse(v)
		if err != nil {
			return fmt.Errorf("invalid UUID string: %w", err)
		}
		*a = AAGUID(u[:])
		return nil
	default:
		return fmt.Errorf("failed to scan AAGUID: unexpected type %T", value)
	}
}

func (a AAGUID) String() string {
	if len(a) != 16 {
		return "invalid AAGUID"
	}

	u, _ := uuid.FromBytes(a)
	return u.String()
}

func (a AAGUID) MarshalGQL(w io.Writer) {
	_, _ = io.WriteString(w, fmt.Sprintf("%q", a.String()))
}

func (a *AAGUID) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("AAGUID must be a string, got %T", v)
	}

	u, err := uuid.Parse(str)
	if err != nil {
		return fmt.Errorf("invalid AAGUID UUID format: %w", err)
	}

	*a = AAGUID(u[:])
	return nil
}

func ToAAGUID(b []byte) *AAGUID {
	a := AAGUID(b)
	return &a
}
