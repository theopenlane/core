package enums

import (
	"errors"
	"io"
	"strings"
)

var (
	// ErrUsageTypeWrongType is returned when unmarshalling from a non-string value
	ErrUsageTypeWrongType = errors.New("wrong type for UsageType")
)

// UsageType is the type of resource usage
type UsageType string

var (
	UsageStorage  UsageType = "STORAGE"
	UsageRecords  UsageType = "RECORDS"
	UsageUsers    UsageType = "USERS"
	UsagePrograms UsageType = "PROGRAMS"
	UsageInvalid  UsageType = "INVALID"
)

func (UsageType) Values() (kinds []string) {
	for _, s := range []UsageType{UsageStorage, UsageRecords, UsageUsers, UsagePrograms} {
		kinds = append(kinds, string(s))
	}

	return
}

func (u UsageType) String() string { return string(u) }

// ToUsageType converts a string to a UsageType
func ToUsageType(s string) *UsageType {
	switch strings.ToUpper(s) {
	case UsageStorage.String():
		return &UsageStorage
	case UsageRecords.String():
		return &UsageRecords
	case UsageUsers.String():
		return &UsageUsers
	case UsagePrograms.String():
		return &UsagePrograms
	default:
		return &UsageInvalid
	}
}

func (u UsageType) MarshalGQL(w io.Writer) { _, _ = w.Write([]byte(`"` + u.String() + `"`)) }

func (u *UsageType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return ErrInvalidUsageType
	}

	*u = UsageType(str)

	return nil
}

// ErrInvalidUsageType is returned when the usage type is invalid
var ErrInvalidUsageType = errors.New("invalid usage type")
