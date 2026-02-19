package hooks

import (
	"strings"

	"github.com/samber/do/v2"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/events"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/gala"
)

const mutationEntityIDProperty = "ID"

// mutationGalaTopic builds a mutation topic contract for an ent schema type.
func mutationGalaTopic(schemaType string) gala.Topic[eventqueue.MutationGalaPayload] {
	return gala.Topic[eventqueue.MutationGalaPayload]{Name: gala.TopicName(schemaType)}
}

// workflowMutationGalaTopic builds a workflow concern mutation topic contract for an ent schema type.
func workflowMutationGalaTopic(schemaType string) gala.Topic[eventqueue.MutationGalaPayload] {
	return gala.Topic[eventqueue.MutationGalaPayload]{Name: eventqueue.WorkflowMutationTopicName(schemaType)}
}

// mutationClientFromGala resolves the ent client from gala listener dependencies.
func mutationClientFromGala(ctx gala.HandlerContext) *entgen.Client {
	client, err := do.Invoke[*entgen.Client](ctx.Injector)
	if err != nil || client == nil {
		return nil
	}

	return client
}

// mutationEntityIDFromGala resolves the mutation entity ID from payload metadata or headers.
func mutationEntityIDFromGala(payload eventqueue.MutationGalaPayload, properties map[string]string) (string, bool) {
	entityID := strings.TrimSpace(payload.EntityID)
	if entityID != "" {
		return entityID, true
	}

	if len(properties) == 0 {
		return "", false
	}

	entityID = strings.TrimSpace(properties[mutationEntityIDProperty])
	if entityID == "" {
		return "", false
	}

	return entityID, true
}

// mutationFieldChanged reports whether a mutation payload indicates a change for a field.
func mutationFieldChanged(payload eventqueue.MutationGalaPayload, field string) bool {
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

// mutationStringValue returns a string proposed value when available.
func mutationStringValue(payload eventqueue.MutationGalaPayload, field string) (string, bool) {
	field = strings.TrimSpace(field)
	if field == "" || payload.ProposedChanges == nil {
		return "", false
	}

	raw, ok := payload.ProposedChanges[field]
	if !ok {
		return "", false
	}

	value, ok := events.ValueAsString(raw)
	if !ok {
		return "", false
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}

	return value, true
}

// mutationStringSliceValue converts a proposed value into a string slice when possible.
func mutationStringSliceValue(payload eventqueue.MutationGalaPayload, field string) []string {
	field = strings.TrimSpace(field)
	if field == "" || payload.ProposedChanges == nil {
		return nil
	}

	raw, ok := payload.ProposedChanges[field]
	if !ok || raw == nil {
		return nil
	}

	switch values := raw.(type) {
	case []string:
		return normalizedStringSlice(values)
	case []any:
		return normalizedStringSlice(lo.FilterMap(values, func(value any, _ int) (string, bool) {
			parsed, ok := events.ValueAsString(value)
			if !ok {
				return "", false
			}

			return parsed, true
		}))
	default:
		value, ok := events.ValueAsString(values)
		if !ok {
			return nil
		}

		return normalizedStringSlice([]string{value})
	}
}

func normalizedStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	out := lo.Uniq(lo.FilterMap(values, func(value string, _ int) (string, bool) {
		value = strings.TrimSpace(value)
		return value, value != ""
	}))
	if len(out) == 0 {
		return nil
	}

	return out
}
