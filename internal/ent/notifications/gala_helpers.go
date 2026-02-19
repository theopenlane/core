package notifications

import (
	"strings"

	"entgo.io/ent"
	"github.com/samber/do/v2"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/gala"
)

const mutationEntityIDProperty = "ID"

func mutationTopic(schemaType string) gala.Topic[eventqueue.MutationGalaPayload] {
	return gala.Topic[eventqueue.MutationGalaPayload]{Name: eventqueue.NotificationMutationTopicName(schemaType)}
}

func clientFromHandler(ctx gala.HandlerContext) (*generated.Client, bool) {
	client, err := do.Invoke[*generated.Client](ctx.Injector)
	if err != nil || client == nil {
		return nil, false
	}

	return client, true
}

func mutationEntityID(payload eventqueue.MutationGalaPayload, properties map[string]string) (string, bool) {
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

	for _, changed := range payload.ChangedFields {
		if strings.TrimSpace(changed) == field {
			return true
		}
	}

	return false
}

func mutationValue(payload eventqueue.MutationGalaPayload, field string) (any, bool) {
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

func mutationStringValue(payload eventqueue.MutationGalaPayload, field string) (string, bool) {
	value, ok := mutationValue(payload, field)
	if !ok {
		return "", false
	}

	asString, ok := events.ValueAsString(value)
	if !ok {
		return "", false
	}

	asString = strings.TrimSpace(asString)
	if asString == "" {
		return "", false
	}

	return asString, true
}

func mutationStringFromProperties(properties map[string]string, field string) string {
	field = strings.TrimSpace(field)
	if field == "" || len(properties) == 0 {
		return ""
	}

	return strings.TrimSpace(properties[field])
}

func isUpdateOperation(operation string) bool {
	switch strings.TrimSpace(operation) {
	case ent.OpUpdate.String(), ent.OpUpdateOne.String():
		return true
	default:
		return false
	}
}
