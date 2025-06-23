package enums

import (
	"fmt"
	"io"
	"strings"
)

// TemplateRecipientStatus is a custom type representing the various states of TemplateRecipientStatus.
type TemplateRecipientStatus string

var (
	// TemplateRecipientStatusActive indicates the active.
	TemplateRecipientStatusActive TemplateRecipientStatus = "ACTIVE"
	// TemplateRecipientStatusExpired indicates the expired.
	TemplateRecipientStatusExpired TemplateRecipientStatus = "EXPIRED"
	// TemplateRecipientStatusUsed indicates the used.
	TemplateRecipientStatusUsed TemplateRecipientStatus = "USED"
	// TemplateRecipientStatusRevoked indicates the revoked.
	TemplateRecipientStatusRevoked TemplateRecipientStatus = "REVOKED"
	// TemplateRecipientStatusInvalid is used when an unknown or unsupported value is provided.
	TemplateRecipientStatusInvalid TemplateRecipientStatus = "TEMPLATERECIPIENTSTATUS_INVALID"
)

// Values returns a slice of strings representing all valid TemplateRecipientStatus values.
func (TemplateRecipientStatus) Values() []string {
	return []string{
		string(TemplateRecipientStatusActive),
		string(TemplateRecipientStatusExpired),
		string(TemplateRecipientStatusUsed),
		string(TemplateRecipientStatusRevoked),
	}
}

// String returns the string representation of the TemplateRecipientStatus value.
func (r TemplateRecipientStatus) String() string {
	return string(r)
}

// ToTemplateRecipientStatus converts a string to its corresponding TemplateRecipientStatus enum value.
func ToTemplateRecipientStatus(r string) *TemplateRecipientStatus {
	switch strings.ToUpper(r) {
	case TemplateRecipientStatusActive.String():
		return &TemplateRecipientStatusActive
	case TemplateRecipientStatusExpired.String():
		return &TemplateRecipientStatusExpired
	case TemplateRecipientStatusUsed.String():
		return &TemplateRecipientStatusUsed
	case TemplateRecipientStatusRevoked.String():
		return &TemplateRecipientStatusRevoked
	default:
		return &TemplateRecipientStatusInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r TemplateRecipientStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *TemplateRecipientStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for TemplateRecipientStatus, got: %T", v)  //nolint:err113
	}

	*r = TemplateRecipientStatus(str)

	return nil
}
