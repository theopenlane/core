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
	schemaType = strings.TrimSpace(schemaType)
	if schemaType == "" {
		return ""
	}

	prefix := ""
	switch concern {
	case MutationConcernWorkflow:
		prefix = workflowMutationTopicPrefix
	case MutationConcernNotification:
		prefix = notificationMutationTopicPrefix
	}

	return gala.TopicName(prefix + schemaType)
}

// MutationTopic returns the typed mutation topic for a concern + schema type pair
func MutationTopic(concern MutationConcern, schemaType string) gala.Topic[MutationGalaPayload] {
	return gala.Topic[MutationGalaPayload]{Name: MutationTopicName(concern, schemaType)}
}
