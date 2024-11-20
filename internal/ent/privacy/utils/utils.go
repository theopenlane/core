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

// AuthzClientFromContext returns the authz client from the context if it exists
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
