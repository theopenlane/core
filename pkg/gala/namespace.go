package gala

import (
	"github.com/theopenlane/core/pkg/jsonx"
)

// TopicNamespace is a plain value binding a topic name prefix to the job kind its topics
// dispatch under; the runtime registry is the resolution authority once topics register
type TopicNamespace struct {
	prefix string
	kind   string
}

// NewTopicNamespace binds a designating prefix to a job kind
func NewTopicNamespace(prefix, kind string) TopicNamespace {
	return TopicNamespace{prefix: prefix, kind: kind}
}

// Name mints a topic name in this namespace
func (n TopicNamespace) Name(suffix string) TopicName {
	return TopicName(n.prefix + suffix)
}

// Prefix returns the namespace's designating prefix
func (n TopicNamespace) Prefix() string {
	return n.prefix
}

// Kind returns the job kind this namespace's topics dispatch under
func (n TopicNamespace) Kind() string {
	return n.kind
}

// TopicOption configures a constructed topic
type TopicOption[T any] func(*Topic[T])

// WithUniqueKey sets the topic's insert-time uniqueness derivation
func WithUniqueKey[T any](derive func(T) string) TopicOption[T] {
	return func(t *Topic[T]) {
		t.UniqueKey = derive
	}
}

// NamespacedTopic constructs a typed topic in the namespace, carrying its kind
func NamespacedTopic[T any](n TopicNamespace, suffix string, opts ...TopicOption[T]) Topic[T] {
	topic := Topic[T]{Name: n.Name(suffix), Kind: n.kind}

	for _, opt := range opts {
		opt(&topic)
	}

	return topic
}

// NamespacedTopicFor constructs a typed topic named by the payload's JSON schema identifier
func NamespacedTopicFor[T any](n TopicNamespace, opts ...TopicOption[T]) Topic[T] {
	return NamespacedTopic[T](n, jsonx.SchemaID(jsonx.SchemaFrom[T]()), opts...)
}
