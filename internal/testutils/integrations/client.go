//go:build test

package integrations

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
)

// Client is the functional test client built from the stored token credential
type Client struct {
	// Token is the resolved API token
	Token string
}

// buildClient constructs the client from the stored token credential
func buildClient(_ context.Context, req types.ClientBuildRequest) (any, error) {
	cred, ok, err := TokenCredential.Resolve(req.Credentials)
	if err != nil || !ok || cred.Token == "" {
		return nil, ErrTokenMissing
	}

	return &Client{Token: cred.Token}, nil
}
