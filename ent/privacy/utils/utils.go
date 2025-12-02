//revive:disable:var-naming
package utils

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/privacy"
)

// GenericMutation is an interface for getting a mutation ID and type
type GenericMutation interface {
	ID() (id string, exists bool)
	IDs(ctx context.Context) ([]string, error)
	Type() string
	Op() ent.Op
	Client() *generated.Client
}

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

	// if we don't have a mutation, return early with nil
	if m == nil {
		return nil
	}

	mut, ok := m.(MutationClient)
	if ok && mut.Client() != nil {
		return &mut.Client().Authz
	}

	return nil
}

// SliceToMap converts a slice of strings to a map for faster lookups
func SliceToMap(s []string) map[string]any {
	m := make(map[string]any)
	for _, t := range s {
		m[t] = struct{}{}
	}

	return m
}

// CheckContains checks if any of the elements in `e` are contained in `s`
func CheckContains(s []string, e []string) bool {
	m := SliceToMap(s)

	for _, item := range e {
		if _, exists := m[item]; exists {
			return true
		}
	}

	return false
}

// GetIntersection returns the results that are included in both slices
// removing any duplicates
func GetIntersection(s1 []string, s2 []string) []string {
	m := SliceToMap(s1)

	set := make(map[string]any)
	intersection := make([]string, 0)

	for _, item := range s2 {
		if _, exists := m[item]; exists {
			if _, seen := set[item]; !seen {
				set[item] = struct{}{}
				intersection = append(intersection, item)
			}
		}
	}

	return intersection
}

// ModulesEnabled checks if the modules feature is enabled for the given client
func ModulesEnabled(client *generated.Client) bool {
	if client == nil {
		return false
	}

	if client.EntConfig == nil {
		return false
	}

	return client.EntConfig.Modules.Enabled
}

// PaymentMethodCheckRequired checks if the config requires
// orgs to have a valid payment method in stripe
func PaymentMethodCheckRequired(client *generated.Client) bool {
	if client == nil {
		return false
	}

	if client.EntConfig == nil {
		return false
	}

	return client.EntConfig.Billing.RequirePaymentMethod
}
