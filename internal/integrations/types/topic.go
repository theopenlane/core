package types

import "github.com/theopenlane/core/pkg/gala"

// OperationTopic returns the canonical gala topic for one definition and operation
func OperationTopic(definitionID, name string) gala.TopicName {
	return gala.TopicName("integration." + definitionID + "." + name)
}

// WebhookEventTopic returns the canonical gala topic for one definition and webhook event
func WebhookEventTopic(definitionID, name string) gala.TopicName {
	return gala.TopicName("integration." + definitionID + ".webhook." + name)
}
