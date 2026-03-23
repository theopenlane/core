package types

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

// keyID produces a distinct pointer, giving every package-level ref variable a unique
// in-process identity without requiring per-type backing structs
type keyID struct{ _ bool }

// =========
// Definitions
// This is the only entity in here that uses plain string because it's string identity is the canonical ID and we store it as a key used to perform lookups
// =========

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

// =========
// Credentials
// =========

// CredentialSlotID is the non-generic durable identity for one credential slot used by a definition.
// It is used in registration structs, bindings, and persistence where the credential schema type is not needed
type CredentialSlotID struct {
	name string
}

// NewCredentialSlotID creates a credential slot identity handle with a stable name for persistence
func NewCredentialSlotID(name string) CredentialSlotID {
	return CredentialSlotID{name: name}
}

// String returns the stable credential name used for persistence and equality comparisons
func (r CredentialSlotID) String() string {
	return r.name
}

// MarshalJSON encodes the credential slot ID as its stable name string
func (r CredentialSlotID) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.name)
}

// UnmarshalJSON decodes a credential slot ID from its stable name string
func (r *CredentialSlotID) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}

	if name == "" {
		*r = CredentialSlotID{}
		return nil
	}

	*r = NewCredentialSlotID(name)

	return nil
}

// CredentialRef is a typed handle for one credential slot, parameterized by the credential schema type
type CredentialRef[T any] struct {
	id CredentialSlotID
}

// NewCredentialRef creates a typed credential slot identity handle
func NewCredentialRef[T any](name string) CredentialRef[T] {
	return CredentialRef[T]{id: NewCredentialSlotID(name)}
}

// ID returns the non-generic credential slot identity
func (r CredentialRef[T]) ID() CredentialSlotID {
	return r.id
}

// String returns the stable credential name used for persistence and equality comparisons
func (r CredentialRef[T]) String() string {
	return r.id.String()
}

// Resolve decodes the credential bound to this slot from the supplied bindings
func (r CredentialRef[T]) Resolve(bindings CredentialBindings) (T, bool, error) {
	for _, b := range bindings {
		if b.Ref.String() == r.id.String() {
			var out T
			if err := json.Unmarshal(b.Credential.Data, &out); err != nil {
				return out, true, err
			}

			return out, true, nil
		}
	}

	var zero T

	return zero, false, nil
}

// =========
// Clients
// =========

// ClientID is the opaque in-process identity for one registered client
type ClientID struct {
	key *keyID `json:"-" yaml:"-"`
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
		id: ClientID{key: new(keyID)},
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

// =========
// OperationRef
// =========

// OperationRef is a typed handle for one registered operation identity
type OperationRef[T any] struct {
	key  *keyID `json:"-" yaml:"-"`
	name string
}

// NewOperationRef creates a typed operation identity handle
func NewOperationRef[T any](name string) OperationRef[T] {
	return OperationRef[T]{
		key:  new(keyID),
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

// =========
// Webhooks
// =========

// WebhookRef is a handle for one registered webhook contract identity
type WebhookRef struct {
	key  *keyID `json:"-" yaml:"-"`
	name string
}

// NewWebhookRef creates a webhook contract identity handle
func NewWebhookRef(name string) WebhookRef {
	return WebhookRef{
		key:  new(keyID),
		name: name,
	}
}

// Name returns the durable webhook identifier
func (r WebhookRef) Name() string {
	return r.name
}

// =========
// Installations
// =========

// InstallationRef is a typed handle for one definition's installation metadata derivation
type InstallationRef[T any] struct {
	key *keyID
	fn  func(ctx context.Context, req InstallationRequest) (T, bool, error)
}

// NewInstallationRef creates a typed installation metadata handle
func NewInstallationRef[T any](fn func(ctx context.Context, req InstallationRequest) (T, bool, error)) InstallationRef[T] {
	return InstallationRef[T]{key: new(keyID), fn: fn}
}

// Resolve derives and marshals installation metadata for one installation
func (r InstallationRef[T]) Resolve(ctx context.Context, req InstallationRequest) (IntegrationInstallationMetadata, bool, error) {
	typed, ok, err := r.fn(ctx, req)
	if err != nil || !ok {
		return IntegrationInstallationMetadata{}, ok, err
	}

	raw, err := jsonx.ToRawMessage(typed)
	if err != nil {
		return IntegrationInstallationMetadata{}, false, err
	}

	return IntegrationInstallationMetadata{Attributes: raw}, true, nil
}

// Registration adapts the typed ref to the InstallationRegistration contract for use in a connection builder
func (r InstallationRef[T]) Registration() *InstallationRegistration {
	return &InstallationRegistration{Resolve: r.Resolve}
}

// =========
// Webhooks
// =========

// WebhookEventRef is a typed handle for one registered webhook event identity
type WebhookEventRef[T any] struct {
	key  *keyID `json:"-" yaml:"-"`
	name string
}

// NewWebhookEventRef creates a typed webhook event identity handle
func NewWebhookEventRef[T any](name string) WebhookEventRef[T] {
	return WebhookEventRef[T]{
		key:  new(keyID),
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
