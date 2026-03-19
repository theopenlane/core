package types

import (
	"encoding/json"
	"fmt"

	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

// DefinitionRef is the durable identity for one registered definition
type DefinitionRef struct {
	id string
}

// NewDefinitionRef creates a definition identity handle
func NewDefinitionRef(id string) DefinitionRef {
	return DefinitionRef{id: id}
}

// ID returns the durable definition identifier
func (r DefinitionRef) ID() string {
	return r.id
}

// clientKey is an unexported zero-size struct used as a unique generic type parameter to prevent ClientID aliasing
type clientKey struct{ _ bool }

// ClientID is the opaque in-process identity for one registered client
type ClientID struct {
	key *clientKey `json:"-" yaml:"-"`
}

// Valid reports whether the client identity was initialized
func (id ClientID) Valid() bool {
	return id.key != nil
}

// String returns the in-process client identity string used for cache indexing
func (id ClientID) String() string {
	return fmt.Sprintf("%p", id.key)
}

// ClientRef is a typed handle for one registered client identity
type ClientRef[T any] struct {
	id ClientID `json:"-" yaml:"-"`
}

// NewClientRef creates a typed client identity handle
func NewClientRef[T any]() ClientRef[T] {
	return ClientRef[T]{
		id: ClientID{key: new(clientKey)},
	}
}

// ID returns the opaque client identity
func (r ClientRef[T]) ID() ClientID {
	return r.id
}

// Cast type-asserts a registered client instance to the typed client value
func (r ClientRef[T]) Cast(client any) (T, error) {
	c, ok := client.(T)
	if !ok {
		var zero T
		return zero, ErrClientCastFailed
	}
	return c, nil
}

type operationKey struct{ _ bool }

// OperationRef is a typed handle for one registered operation identity
type OperationRef[T any] struct {
	key  *operationKey `json:"-" yaml:"-"`
	name string
}

// NewOperationRef creates a typed operation identity handle
func NewOperationRef[T any](name string) OperationRef[T] {
	return OperationRef[T]{
		key:  new(operationKey),
		name: name,
	}
}

// Name returns the durable operation identifier
func (r OperationRef[T]) Name() string {
	return r.name
}

// Topic returns the canonical gala topic for one definition slug and operation
func (r OperationRef[T]) Topic(slug string) gala.TopicName {
	return gala.TopicName("integration." + slug + "." + r.name)
}

// UnmarshalConfig decodes a JSON operation config document into the typed config value
func (r OperationRef[T]) UnmarshalConfig(raw json.RawMessage) (T, error) {
	var out T
	if err := jsonx.UnmarshalIfPresent(raw, &out); err != nil {
		return out, err
	}
	return out, nil
}

type webhookKey struct{ _ bool }

// WebhookRef is a handle for one registered webhook contract identity
type WebhookRef struct {
	key  *webhookKey `json:"-" yaml:"-"`
	name string
}

// NewWebhookRef creates a webhook contract identity handle
func NewWebhookRef(name string) WebhookRef {
	return WebhookRef{
		key:  new(webhookKey),
		name: name,
	}
}

// Name returns the durable webhook identifier
func (r WebhookRef) Name() string {
	return r.name
}

type webhookEventKey struct{ _ bool }

// WebhookEventRef is a typed handle for one registered webhook event identity
type WebhookEventRef[T any] struct {
	key  *webhookEventKey `json:"-" yaml:"-"`
	name string
}

// NewWebhookEventRef creates a typed webhook event identity handle
func NewWebhookEventRef[T any](name string) WebhookEventRef[T] {
	return WebhookEventRef[T]{
		key:  new(webhookEventKey),
		name: name,
	}
}

// Name returns the durable webhook event identifier
func (r WebhookEventRef[T]) Name() string {
	return r.name
}

// Topic returns the canonical gala topic for one definition slug and webhook event
func (r WebhookEventRef[T]) Topic(slug string) gala.TopicName {
	return gala.TopicName("integration." + slug + ".webhook." + r.name)
}

// UnmarshalPayload decodes a JSON webhook event payload into the typed payload value
func (r WebhookEventRef[T]) UnmarshalPayload(raw json.RawMessage) (T, error) {
	var out T
	if err := jsonx.UnmarshalIfPresent(raw, &out); err != nil {
		return out, err
	}
	return out, nil
}
