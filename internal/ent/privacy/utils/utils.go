package utils

import (
	"github.com/theopenlane/core/internal/ent/generated/privacy"
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
