package gala

import (
	"strings"

	"github.com/theopenlane/core/pkg/jsonx"
)

// uniqueKeySeparator delimits unique key segments
const uniqueKeySeparator = ":"

// Namespace is one node in the durable naming tree: a job kind at the root, topic families below
type Namespace struct {
	prefix string
	kind   string
}

// NewKind declares the job kind gala.<name> whose canonical topic prefix is <name>.
func NewKind(name string) Namespace {
	return Namespace{kind: "gala." + name, prefix: name + "."}
}

// Kind returns the job kind this namespace's topics dispatch under
func (n Namespace) Kind() string {
	return n.kind
}

// Queue returns the river queue name for the namespace's job kind
func (n Namespace) Queue() string {
	return strings.ReplaceAll(n.kind, ".", "_")
}

// Name mints a topic name in this namespace
func (n Namespace) Name(suffix string) TopicName {
	return TopicName(n.prefix + suffix)
}

// Key mints a colon-delimited insert-time dedup key under the namespace's prefix
func (n Namespace) Key(segments ...string) string {
	return strings.TrimSuffix(n.prefix, ".") + ":" + strings.Join(segments, ":")
}

// Child returns the namespace one segment deeper, keeping the parent's job kind
func (n Namespace) Child(segment string) Namespace {
	return Namespace{prefix: n.prefix + segment + ".", kind: n.kind}
}

// Prefixed returns the namespace with a segment prepended to its prefix, keeping the job kind
func (n Namespace) Prefixed(segment string) Namespace {
	return Namespace{prefix: segment + "." + n.prefix, kind: n.kind}
}

// At returns the namespace rehomed at the given prefix, keeping the job kind
func (n Namespace) At(prefix string) Namespace {
	return Namespace{prefix: prefix + ".", kind: n.kind}
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
func NamespacedTopic[T any](n Namespace, suffix string, opts ...TopicOption[T]) Topic[T] {
	topic := Topic[T]{Name: n.Name(suffix), Kind: n.kind}

	for _, opt := range opts {
		opt(&topic)
	}

	return topic
}

// NamespacedTopicFor constructs a typed topic named by the payload's JSON schema identifier
func NamespacedTopicFor[T any](n Namespace, opts ...TopicOption[T]) Topic[T] {
	return NamespacedTopic(n, jsonx.SchemaID(jsonx.SchemaFrom[T]()), opts...)
}

// root namespaces for every durable envelope family; each one is a river job kind and the
// canonical topic prefix its families derive from
var (
	// Mutation carries mutation event envelopes
	Mutation = NewKind("mutation")
	// Workflow carries workflow command envelopes
	Workflow = NewKind("workflow")
	// IntegrationRun carries one-shot integration operation envelopes
	IntegrationRun = NewKind("integration.run")
	// IntegrationReconcile carries recurring reconcile and scheduled cycle envelopes
	IntegrationReconcile = NewKind("integration.reconcile")
	// IntegrationIngest carries ingest persistence envelopes
	IntegrationIngest = NewKind("integration.ingest")
	// IntegrationWebhook carries inbound webhook envelopes
	IntegrationWebhook = NewKind("integration.webhook")
	// System carries startup and maintenance envelopes
	System = NewKind("system")
)

// JobKinds returns every root namespace a durable runtime registers by default
func JobKinds() []Namespace {
	return []Namespace{
		Mutation,
		Workflow,
		IntegrationRun,
		IntegrationReconcile,
		IntegrationIngest,
		IntegrationWebhook,
		System,
	}
}
