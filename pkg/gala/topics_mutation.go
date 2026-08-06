package gala

import (
	"strings"
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

const (
	// MutationPropertyEntityID is the standard mutation metadata key used for entity identifiers
	MutationPropertyEntityID = "ID"
	// MutationPropertyOperation is the mutation metadata key used for the operation type
	MutationPropertyOperation = "operation"
	// MutationPropertyMutationType is the mutation metadata key used for the ent schema type
	MutationPropertyMutationType = "mutation_type"
)

// MutationTopicName returns the mutation topic name for a concern + schema type pair
func MutationTopicName(concern MutationConcern, schemaType string) TopicName {
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

	return TopicName(prefix + schemaType)
}
