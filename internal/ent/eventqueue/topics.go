package eventqueue

import (
	"strings"

	"github.com/theopenlane/core/pkg/gala"
)

const (
	workflowMutationTopicPrefix     = "workflow.mutation."
	notificationMutationTopicPrefix = "notification.mutation."
)

// WorkflowMutationTopicName returns the workflow concern mutation topic name for a schema type.
func WorkflowMutationTopicName(schemaType string) gala.TopicName {
	return prefixedMutationTopicName(workflowMutationTopicPrefix, schemaType)
}

// NotificationMutationTopicName returns the notification concern mutation topic name for a schema type.
func NotificationMutationTopicName(schemaType string) gala.TopicName {
	return prefixedMutationTopicName(notificationMutationTopicPrefix, schemaType)
}

func prefixedMutationTopicName(prefix, schemaType string) gala.TopicName {
	schemaType = strings.TrimSpace(schemaType)
	if schemaType == "" {
		return ""
	}

	return gala.TopicName(prefix + schemaType)
}
