package types //nolint:revive

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// keyID produces a distinct pointer, giving every package-level ref variable a unique
// in-process identity without requiring per-type backing structs
type keyID struct{ _ bool }

// namedRef is the shared base for ref types that carry a pointer-based in-process identity
// and a stable string name used for persistence and topic derivation
type namedRef struct {
	// key is the pointer-based in-process identity for the ref
	key *keyID `json:"-" yaml:"-"`
	// name is the stable string name used for persistence and topic derivation
	name string
}

// newNamedRef creates a named ref with a fresh in-process identity
func newNamedRef(name string) namedRef {
	return namedRef{key: new(keyID), name: name}
}

// Name returns the stable identifier for the ref
func (r namedRef) Name() string {
	return r.name
}

// =========
// Definitions
// This is the only entity in here that uses plain string because its string identity is the canonical ID
// =========

// DefinitionRef is the durable identity for one registered definition
type DefinitionRef struct {
	// id is the durable definition identifier
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

// OperationTopics returns the topic namespace for this definition's operations
func (r DefinitionRef) OperationTopics() gala.Namespace {
	return gala.IntegrationRun.Child(r.id)
}

// WebhookEventTopics returns the topic namespace for this definition's webhook events
func (r DefinitionRef) WebhookEventTopics() gala.Namespace {
	return gala.IntegrationWebhook.Child(r.id)
}

// OperationTopic returns the canonical gala topic for one definition operation
func (r DefinitionRef) OperationTopic(name string) gala.TopicName {
	return r.OperationTopics().Name(name)
}

// WebhookEventTopic returns the canonical gala topic for one definition webhook event
func (r DefinitionRef) WebhookEventTopic(name string) gala.TopicName {
	return r.WebhookEventTopics().Name(name)
}

// =========
// Credentials
// =========

// CredentialSlotID is the non-generic durable identity for one credential slot used by a definition
// It is used in registration structs, bindings, and persistence where the credential schema type is not needed
type CredentialSlotID struct {
	// name is the stable credential slot name used for persistence and equality comparisons
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

// CredentialSlot is the typed identity a credential registration points at; only CredentialRef satisfies it
type CredentialSlot interface {
	// ID returns the non-generic credential slot identity
	ID() CredentialSlotID
	// Schema returns the reflected JSON schema of the stored credential type
	Schema() json.RawMessage
	// Replaces lists the retired slots whose stored payloads this slot takes over
	Replaces() []CredentialSlotID
	// Convert reshapes a payload stored under one of the replaced slots into this slot's shape
	Convert(from CredentialSlotID, old json.RawMessage) (json.RawMessage, error)
	// Backfills reports whether the slot declares how to complete a stored payload missing values
	Backfills() bool
	// Backfill derives values a stored payload lacks from the live installation
	Backfill(ctx context.Context, req InstallationRequest, payload json.RawMessage) (json.RawMessage, error)
}

// CredentialRef is a typed handle for one credential slot, parameterized by the credential schema type
type CredentialRef[T any] struct {
	// id is the non-generic credential slot identity derived from the type name
	id CredentialSlotID
	// schema is the reflected JSON schema of T
	schema json.RawMessage
	// replacements decode payloads stored under retired slots into T
	replacements map[CredentialSlotID]func(json.RawMessage) (T, error)
	// backfill derives values a decoded payload lacks from the live installation
	backfill func(context.Context, InstallationRequest, *T) error
}

// NewCredentialRef reflects T once and creates the typed credential slot handle named after it
func NewCredentialRef[T any]() CredentialRef[T] {
	schema := jsonx.SchemaFrom[T]()

	return CredentialRef[T]{id: NewCredentialSlotID(jsonx.SchemaID(schema)), schema: schema}
}

// Replacing declares that ref takes over payloads stored under old, converted with convert or decoded directly into T when convert is nil
func Replacing[T, Old any](ref CredentialRef[T], old CredentialRef[Old], convert func(Old) T) CredentialRef[T] {
	if ref.replacements == nil {
		ref.replacements = map[CredentialSlotID]func(json.RawMessage) (T, error){}
	}

	ref.replacements[old.ID()] = func(payload json.RawMessage) (T, error) {
		if convert == nil {
			return jsonx.Decode[T](payload)
		}

		previous, err := jsonx.Decode[Old](payload)
		if err != nil {
			var zero T

			return zero, err
		}

		return convert(previous), nil
	}

	return ref
}

// Backfilled declares how a stored payload missing values is completed from the live installation
func (r CredentialRef[T]) Backfilled(fn func(context.Context, InstallationRequest, *T) error) CredentialRef[T] {
	r.backfill = fn

	return r
}

// Replaces lists the retired slots whose stored payloads this slot takes over
func (r CredentialRef[T]) Replaces() []CredentialSlotID {
	return lo.Keys(r.replacements)
}

// Convert reshapes a payload stored under one of the replaced slots into this slot's shape
func (r CredentialRef[T]) Convert(from CredentialSlotID, old json.RawMessage) (json.RawMessage, error) {
	decode, ok := r.replacements[from]
	if !ok {
		return nil, fmt.Errorf("%w: %s does not replace %s", ErrCredentialNotReplaced, r.id, from)
	}

	value, err := decode(old)
	if err != nil {
		return nil, err
	}

	return jsonx.ToRawMessage(value)
}

// Backfills reports whether the slot declares how to complete a stored payload missing values
func (r CredentialRef[T]) Backfills() bool {
	return r.backfill != nil
}

// Backfill derives values a stored payload lacks from the live installation, returning it unchanged when none is declared
func (r CredentialRef[T]) Backfill(ctx context.Context, req InstallationRequest, payload json.RawMessage) (json.RawMessage, error) {
	if r.backfill == nil {
		return payload, nil
	}

	var value T
	if err := jsonx.UnmarshalIfPresent(payload, &value); err != nil {
		return nil, err
	}

	if err := r.backfill(ctx, req, &value); err != nil {
		return nil, err
	}

	return jsonx.ToRawMessage(value)
}

// ID returns the non-generic credential slot identity
func (r CredentialRef[T]) ID() CredentialSlotID {
	return r.id
}

// Schema returns the reflected JSON schema of the stored credential type
func (r CredentialRef[T]) Schema() json.RawMessage {
	return jsonx.CloneRawMessage(r.schema)
}

// MarshalJSON encodes the ref as its stable slot name
func (r CredentialRef[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.id)
}

// String returns the stable credential name used for persistence and equality comparisons
func (r CredentialRef[T]) String() string {
	return r.id.String()
}

// Resolve decodes the credential bound to this slot from the supplied bindings
func (r CredentialRef[T]) Resolve(bindings CredentialBindings) (T, bool, error) {
	cred, ok := bindings.Resolve(r.id)
	if !ok {
		var zero T
		return zero, false, nil
	}

	var out T

	if err := json.Unmarshal(cred.Data, &out); err != nil {
		return out, true, err
	}

	return out, true, nil
}

// =========
// Clients
// =========

// ClientID is the opaque in-process identity for one registered client
type ClientID struct {
	// key is the pointer-based in-process identity for the client
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
	// id is the opaque client identity
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
	namedRef
}

// NewOperationRef creates a typed operation identity handle
func NewOperationRef[T any](name string) OperationRef[T] {
	return OperationRef[T]{namedRef: newNamedRef(name)}
}

// UnmarshalConfig decodes a JSON operation config document into the typed config value
func (r OperationRef[T]) UnmarshalConfig(raw json.RawMessage) (T, error) {
	var out T

	return out, jsonx.UnmarshalIfPresent(raw, &out)
}

// =========
// Installations
// =========

// InstallationRef is a typed handle for one definition's installation metadata derivation
type InstallationRef[T any] struct {
	// key is the pointer-based in-process identity for the installation ref
	key *keyID
	// fn is the typed resolve function that derives installation metadata
	fn func(ctx context.Context, req InstallationRequest) (T, bool, error)
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

	meta := IntegrationInstallationMetadata{Attributes: raw}

	if identifiable, ok := any(typed).(InstallationIdentifiable); ok {
		meta.Display = identifiable.InstallationIdentity()
	}

	return meta, true, nil
}

// Registration adapts the typed ref to the InstallationRegistration contract for use in a connection builder
func (r InstallationRef[T]) Registration() *InstallationRegistration {
	return &InstallationRegistration{Resolve: r.Resolve}
}

// =========
// Webhooks
// =========

// WebhookRef is a handle for one registered webhook contract identity
type WebhookRef struct {
	namedRef
}

// NewWebhookRef creates a webhook contract identity handle
func NewWebhookRef(name string) WebhookRef {
	return WebhookRef{namedRef: newNamedRef(name)}
}

// WebhookEventRef is a typed handle for one registered webhook event identity
type WebhookEventRef[T any] struct {
	namedRef
}

// NewWebhookEventRef creates a typed webhook event identity handle
func NewWebhookEventRef[T any](name string) WebhookEventRef[T] {
	return WebhookEventRef[T]{namedRef: newNamedRef(name)}
}

// UnmarshalPayload decodes a JSON webhook event payload into the typed payload value
func (r WebhookEventRef[T]) UnmarshalPayload(raw json.RawMessage) (T, error) {
	var out T
	return out, jsonx.UnmarshalIfPresent(raw, &out)
}
