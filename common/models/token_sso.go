package models

import (
	"context"
	"io"
	"time"

	"github.com/theopenlane/utils/contextx"
)

// SSOAuthorizationMap tracks SSO verification timestamps per organization ID.
type SSOAuthorizationMap map[string]time.Time

// MarshalGQL implements the gqlgen Marshaler interface.
func (m SSOAuthorizationMap) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, m)
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (m *SSOAuthorizationMap) UnmarshalGQL(v any) error {
	return unmarshalGQLJSON(v, m)
}

// WithSSOAuthorizations stores the SSOAuthorizations in the context
func WithSSOAuthorizations(ctx context.Context, auth *SSOAuthorizationMap) context.Context {
	return contextx.With(ctx, auth)
}

// SSOAuthorizationsFromContext retrieves SSOAuthorizations from the context
func SSOAuthorizationsFromContext(ctx context.Context) (*SSOAuthorizationMap, bool) {
	return contextx.From[*SSOAuthorizationMap](ctx)
}
