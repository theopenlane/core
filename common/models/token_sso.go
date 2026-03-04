package models

import (
	"io"
	"time"

	"github.com/theopenlane/utils/contextx"
)

// SSOAuthorizationMap tracks SSO verification timestamps per organization ID.
type SSOAuthorizationMap map[string]time.Time

var SSOAuthorizationsContextKey = contextx.NewKey[*SSOAuthorizationMap]()

// MarshalGQL implements the gqlgen Marshaler interface.
func (m SSOAuthorizationMap) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, m)
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (m *SSOAuthorizationMap) UnmarshalGQL(v any) error {
	return unmarshalGQLJSON(v, m)
}
