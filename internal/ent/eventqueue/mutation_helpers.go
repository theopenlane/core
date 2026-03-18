package eventqueue

import (
	"strings"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/mutations"
)

const (
	// MutationPropertyEntityID is the standard mutation metadata key used for entity identifiers
	MutationPropertyEntityID = "ID"
	// MutationPropertyOperation is the mutation metadata key used for the operation type
	MutationPropertyOperation = "operation"
	// MutationPropertyMutationType is the mutation metadata key used for the ent schema type
	MutationPropertyMutationType = "mutation_type"
	// SoftDeleteOne is a synthetic operation used for soft-delete hooks
	SoftDeleteOne = "SoftDeleteOne"
)

// MutationEntityID resolves the mutation entity ID from payload metadata or headers
func MutationEntityID(payload MutationGalaPayload, properties map[string]string) (string, bool) {
	entityID := strings.TrimSpace(payload.EntityID)
	if entityID != "" {
		return entityID, true
	}

	if len(properties) == 0 {
		return "", false
	}

	entityID = strings.TrimSpace(properties[MutationPropertyEntityID])
	if entityID == "" {
		return "", false
	}

	return entityID, true
}

// MutationFieldChanged reports whether a payload indicates a field changed
func MutationFieldChanged(payload MutationGalaPayload, field string) bool {
	field = strings.TrimSpace(field)
	if field == "" {
		return false
	}

	if payload.ProposedChanges != nil {
		if _, ok := payload.ProposedChanges[field]; ok {
			return true
		}
	}

	return lo.SomeBy(payload.ChangedFields, func(changed string) bool {
		return strings.TrimSpace(changed) == field
	})
}

// MutationValue returns a proposed mutation value for a field
func MutationValue(payload MutationGalaPayload, field string) (any, bool) {
	field = strings.TrimSpace(field)
	if field == "" || payload.ProposedChanges == nil {
		return nil, false
	}

	value, ok := payload.ProposedChanges[field]
	if !ok {
		return nil, false
	}

	return value, true
}

// MutationStringValue returns a proposed string mutation value for a field
func MutationStringValue(payload MutationGalaPayload, field string) (string, bool) {
	raw, ok := MutationValue(payload, field)
	if !ok {
		return "", false
	}

	value, ok := ValueAsString(raw)
	if !ok {
		return "", false
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}

	return value, true
}

// MutationStringSliceValue returns a proposed string-slice mutation value for a field
func MutationStringSliceValue(payload MutationGalaPayload, field string) []string {
	raw, ok := MutationValue(payload, field)
	if !ok || raw == nil {
		return nil
	}

	switch values := raw.(type) {
	case []string:
		return mutations.NormalizeStrings(values)
	case []any:
		return mutations.NormalizeStrings(lo.FilterMap(values, func(value any, _ int) (string, bool) {
			parsed, ok := ValueAsString(value)
			if !ok {
				return "", false
			}

			return parsed, true
		}))
	default:
		value, ok := ValueAsString(values)
		if !ok {
			return nil
		}

		return mutations.NormalizeStrings([]string{value})
	}
}

// MutationStringFromProperties returns a trimmed string value from envelope properties
func MutationStringFromProperties(properties map[string]string, field string) string {
	field = strings.TrimSpace(field)
	if field == "" || len(properties) == 0 {
		return ""
	}

	return strings.TrimSpace(properties[field])
}

// MutationStringValueOrProperty returns the proposed string value for a field with property fallback
func MutationStringValueOrProperty(payload MutationGalaPayload, properties map[string]string, field string) string {
	if value, ok := MutationStringValue(payload, field); ok {
		return value
	}

	return MutationStringFromProperties(properties, field)
}

// MutationStringValuePreferPayload returns property fallback only when the field is absent from proposed changes
func MutationStringValuePreferPayload(payload MutationGalaPayload, properties map[string]string, field string) string {
	raw, exists := MutationValue(payload, field)
	if !exists {
		return MutationStringFromProperties(properties, field)
	}

	value, ok := ValueAsString(raw)
	if !ok {
		return ""
	}

	return strings.TrimSpace(value)
}
