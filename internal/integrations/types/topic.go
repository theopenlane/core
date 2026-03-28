package types

import (
	"reflect"

	"github.com/theopenlane/core/pkg/gala"
)

// OperationTopic returns the canonical gala topic for one definition and operation
func OperationTopic(definitionID, name string) gala.TopicName {
	return gala.TopicName("integration." + definitionID + "." + name)
}

// WebhookEventTopic returns the canonical gala topic for one definition and webhook event
func WebhookEventTopic(definitionID, name string) gala.TopicName {
	return gala.TopicName("integration." + definitionID + ".webhook." + name)
}

// TopicFromType returns a canonical gala topic name derived from a Go type's name
func TopicFromType[T any]() gala.TopicName {
	return gala.TopicName("integration." + reflect.TypeFor[T]().Name())
}

// ListenerFromType returns a canonical gala listener name derived from a Go type's name
func ListenerFromType[T any]() string {
	return "integration." + reflect.TypeFor[T]().Name() + ".handler"
}
