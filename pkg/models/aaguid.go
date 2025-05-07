package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
)

var ErrUnsupportedDataType = errors.New("unsupported aaguid format")

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
			return ErrUnsupportedDataType
		}

		*a = AAGUID(u[:])

		return nil
	default:
		return ErrUnsupportedDataType
	}
}

func (a AAGUID) String() string {
	u, err := uuid.FromBytes(a)
	if err != nil {
		return ""
	}

	return u.String()
}

func (a AAGUID) MarshalGQL(w io.Writer) {
	_, _ = io.WriteString(w, fmt.Sprintf("%q", a.String()))
}

func (a *AAGUID) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return ErrUnsupportedDataType
	}

	u, err := uuid.Parse(str)
	if err != nil {
		return ErrUnsupportedDataType
	}

	*a = AAGUID(u[:])

	return nil
}

func ToAAGUID(b []byte) *AAGUID {
	a := AAGUID(b)
	return &a
}
