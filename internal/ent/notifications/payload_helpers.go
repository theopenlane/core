package notifications

import (
	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/events"
)

// mutationProposedValue returns a proposed field value from mutation metadata.
func mutationProposedValue(payload *events.MutationPayload, field string) (any, bool) {
	return events.ProposedValue(payload, field)
}

// mutationProposedString returns a proposed field value as string when possible.
func mutationProposedString(payload *events.MutationPayload, field string) (string, bool) {
	return events.ProposedString(payload, field)
}

// mutationProposedAnySlice returns a proposed field value as a []any slice when possible.
func mutationProposedAnySlice(payload *events.MutationPayload, field string) ([]any, bool) {
	raw, ok := mutationProposedValue(payload, field)
	if !ok {
		return nil, false
	}

	switch value := raw.(type) {
	case nil:
		return nil, true
	case []any:
		return append([]any(nil), value...), true
	case []string:
		return lo.Map(value, func(entry string, _ int) any { return entry }), true
	default:
		return nil, false
	}
}

// isUpdateOperation reports whether the mutation operation is an update variant.
func isUpdateOperation(operation string) bool {
	switch operation {
	case "Update", "UpdateOne":
		return true
	default:
		return false
	}
}
