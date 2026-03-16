package types

import (
	"fmt"

	"github.com/theopenlane/core/pkg/gala"
)

// DefinitionRef is the durable identity for one registered definition.
type DefinitionRef struct {
	id string
}

// NewDefinitionRef creates a definition identity handle.
func NewDefinitionRef(id string) DefinitionRef {
	return DefinitionRef{id: id}
}

// ID returns the durable definition identifier.
func (r DefinitionRef) ID() string {
	return r.id
}

type clientKey struct{ _ bool }

// ClientID is the opaque in-process identity for one registered client.
type ClientID struct {
	key *clientKey `json:"-" yaml:"-"`
}

// Valid reports whether the client identity was initialized.
func (id ClientID) Valid() bool {
	return id.key != nil
}

// String returns the in-process client identity string used for cache indexing.
func (id ClientID) String() string {
	return fmt.Sprintf("%p", id.key)
}

// ClientRef is a typed handle for one registered client identity.
type ClientRef[T any] struct {
	id ClientID `json:"-" yaml:"-"`
}

// NewClientRef creates a typed client identity handle.
func NewClientRef[T any]() ClientRef[T] {
	return ClientRef[T]{
		id: ClientID{key: new(clientKey)},
	}
}

// ID returns the opaque client identity.
func (r ClientRef[T]) ID() ClientID {
	return r.id
}

type operationKey struct{ _ bool }

// OperationRef is a typed handle for one registered operation identity.
type OperationRef[T any] struct {
	key  *operationKey `json:"-" yaml:"-"`
	name string
}

// NewOperationRef creates a typed operation identity handle.
func NewOperationRef[T any](name string) OperationRef[T] {
	return OperationRef[T]{
		key:  new(operationKey),
		name: name,
	}
}

// Name returns the durable operation identifier.
func (r OperationRef[T]) Name() string {
	return r.name
}

// Topic returns the canonical gala topic for one definition slug and operation.
func (r OperationRef[T]) Topic(slug string) gala.TopicName {
	return gala.TopicName("integration." + slug + "." + r.name)
}
