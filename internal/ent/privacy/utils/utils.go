package utils

import (
	"context"

	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/middleware/transaction"
)

// NewMutationPolicyWithoutNil is creating a new slice of `privacy.MutationPolicy` by
// removing any `nil` values from the input `source` slice. It iterates over each item in the source slice and appends it to the new slice only if it is not `nil` - the new slice is then returned
func NewMutationPolicyWithoutNil(source privacy.MutationPolicy) privacy.MutationPolicy {
	newSlice := make(privacy.MutationPolicy, 0, len(source))

	for _, item := range source {
		if item != nil {
			newSlice = append(newSlice, item)
		}
	}

	return newSlice
}

// MutationClient is an interface that can be implemented by a mutation to return the ent client
type MutationClient interface {
	Client() *generated.Client
}

// AuthzClientFromContext returns the authz client from the context if it exists
// this is useful when you need to get the client from the context directly
func AuthzClientFromContext(ctx context.Context) *fgax.Client {
	client := generated.FromContext(ctx)
	if client != nil {
		return &client.Authz
	}

	tx := transaction.FromContext(ctx)
	if tx != nil {
		return &tx.Authz
	}

	return nil
}

// AuthzClient returns the authz client from the context if it exists, otherwise it will
// attempt to get the client from the mutation if it implements the `MutationClient` interface
func AuthzClient(ctx context.Context, m generated.Mutation) *fgax.Client {
	client := AuthzClientFromContext(ctx)
	if client != nil {
		return client
	}

	mut, ok := m.(MutationClient)
	if ok && mut.Client() != nil {
		return &mut.Client().Authz
	}

	return nil
}
