//go:build test

package integrations

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// Client is the functional test client built from a stored token credential
type Client struct {
	// Token is the resolved API token
	Token string
}

// tokenClient returns a client builder reading the token from slot
func tokenClient[T any](slot types.CredentialRef[T], token func(T) string) func(context.Context, types.ClientBuildRequest) (*Client, error) {
	return func(_ context.Context, req types.ClientBuildRequest) (*Client, error) {
		cred, ok, err := slot.Resolve(req.Credentials)
		if err != nil || !ok || token(cred) == "" {
			return nil, ErrTokenMissing
		}

		return &Client{Token: token(cred)}, nil
	}
}
