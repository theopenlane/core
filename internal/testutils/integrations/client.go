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

// tokenClient returns a client builder reading the token from the decoded credential
func tokenClient[T any](token func(T) string) func(context.Context, types.ConnectionRequest[T]) (*Client, error) {
	return func(_ context.Context, req types.ConnectionRequest[T]) (*Client, error) {
		value := token(req.Credential)
		if value == "" {
			return nil, ErrTokenMissing
		}

		return &Client{Token: value}, nil
	}
}
