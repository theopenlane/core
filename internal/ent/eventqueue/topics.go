package eventqueue

import (
	"strings"

	"github.com/theopenlane/core/pkg/gala"
)

const (
	workflowMutationTopicPrefix     = "workflow.mutation."
	notificationMutationTopicPrefix = "notification.mutation."
)

// MutationConcern identifies the eventing concern namespace for mutation topics
type MutationConcern string

const (
	// MutationConcernDirect is the default concern for direct mutation listeners
	MutationConcernDirect MutationConcern = "direct"
	// MutationConcernWorkflow is the concern for workflow mutation listeners
	MutationConcernWorkflow MutationConcern = "workflow"
	// MutationConcernNotification is the concern for notification mutation listeners
	MutationConcernNotification MutationConcern = "notification"
)

// MutationTopicName returns the mutation topic name for a concern + schema type pair
func MutationTopicName(concern MutationConcern, schemaType string) gala.TopicName {
	return prefixedMutationTopicName(mutationTopicPrefix(concern), schemaType)
}

// MutationTopic returns the typed mutation topic for a concern + schema type pair
func MutationTopic(concern MutationConcern, schemaType string) gala.Topic[MutationGalaPayload] {
	return mutationTopicFromName(MutationTopicName(concern, schemaType))
}

// prefixedMutationTopicName composes a topic name prefix with a normalized schema type
func prefixedMutationTopicName(prefix, schemaType string) gala.TopicName {
	schemaType = strings.TrimSpace(schemaType)
	if schemaType == "" {
		return ""
	}

	return gala.TopicName(prefix + schemaType)
}

// mutationTopicPrefix resolves the configured prefix for a mutation concern
func mutationTopicPrefix(concern MutationConcern) string {
	switch concern {
	case MutationConcernWorkflow:
		return workflowMutationTopicPrefix
	case MutationConcernNotification:
		return notificationMutationTopicPrefix
	default:
		return ""
	}
}

// mutationTopicFromName wraps a topic name in a typed mutation topic contract
func mutationTopicFromName(name gala.TopicName) gala.Topic[MutationGalaPayload] {
	return gala.Topic[MutationGalaPayload]{Name: name}
}
