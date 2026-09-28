package types //nolint:revive

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

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

// replacement is one retired layout a typed ref takes over
type replacement[T any] struct {
	schema json.RawMessage
	decode func(json.RawMessage) (T, error)
}

// replacing records a retired layout under name, decoded through convert or directly into T when convert is nil
func replacing[T, Old any](replacements map[string]replacement[T], name string, convert func(Old) T) map[string]replacement[T] {
	next := maps.Clone(replacements)
	if next == nil {
		next = map[string]replacement[T]{}
	}

	next[name] = replacement[T]{schema: jsonx.SchemaFrom[Old](), decode: func(payload json.RawMessage) (T, error) {
		if convert == nil {
			return jsonx.Decode[T](payload)
		}

		previous, err := jsonx.Decode[Old](payload)
		if err != nil {
			var zero T

			return zero, err
		}

		return convert(previous), nil
	}}

	return next
}

// convertReplaced reshapes a payload stored under the retired layout from into T
func convertReplaced[T any](replacements map[string]replacement[T], from string, old json.RawMessage) (json.RawMessage, error) {
	retired, ok := replacements[from]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotReplaced, from)
	}

	result, err := jsonx.ValidateSchema(retired.schema, old)
	if err != nil {
		return nil, err
	}

	if !result.Valid() {
		return nil, fmt.Errorf("%w: %s: %s", ErrLayoutMismatch, from, strings.Join(jsonx.ValidationErrorStrings(result), "; "))
	}

	value, err := retired.decode(old)
	if err != nil {
		return nil, err
	}

	return jsonx.ToRawMessage(value)
}

// backfillPayload completes the payload through fn, returning it unchanged when fn is nil
func backfillPayload[T any](ctx context.Context, req InstallationRequest, fn func(context.Context, InstallationRequest, *T) error, payload json.RawMessage) (json.RawMessage, error) {
	if fn == nil {
		return payload, nil
	}

	var value T

	if err := jsonx.UnmarshalIfPresent(payload, &value); err != nil {
		return nil, err
	}

	if err := fn(ctx, req, &value); err != nil {
		return nil, err
	}

	return jsonx.ToRawMessage(value)
}

// sortedNames returns the retired layout names in sorted order
func sortedNames[T any](replacements map[string]replacement[T]) []string {
	names := lo.Keys(replacements)

	slices.Sort(names)

	return names
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

// CredentialRef is a typed handle for one credential slot, parameterized by the credential schema type
type CredentialRef[T any] struct {
	// id is the durable credential slot identity
	id CredentialSlotID
	// schema is the reflected JSON schema of the credential type
	schema json.RawMessage
	// replacements are the retired slot layouts this slot takes over, keyed by retired slot name
	replacements map[string]replacement[T]
	// backfill completes a stored payload missing values
	backfill func(context.Context, InstallationRequest, *T) error
}

// NewCredentialRef creates a typed credential slot identity handle with the schema reflected from T
func NewCredentialRef[T any](name string) CredentialRef[T] {
	return CredentialRef[T]{id: NewCredentialSlotID(name), schema: jsonx.SchemaFrom[T]()}
}

// CredentialRefOf creates a typed credential slot identity handle named after the reflected schema of T
func CredentialRefOf[T any]() CredentialRef[T] {
	schema := jsonx.SchemaFrom[T]()

	return CredentialRef[T]{id: NewCredentialSlotID(jsonx.SchemaID(schema)), schema: schema}
}

// ID returns the non-generic credential slot identity
func (r CredentialRef[T]) ID() CredentialSlotID {
	return r.id
}

// String returns the stable credential name used for persistence and equality comparisons
func (r CredentialRef[T]) String() string {
	return r.id.String()
}

// Schema returns a copy of the reflected JSON schema of the credential type
func (r CredentialRef[T]) Schema() json.RawMessage {
	return jsonx.CloneRawMessage(r.schema)
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

// Replacing declares that the slot takes over payloads stored under old
func (r CredentialRef[T]) Replacing[Old any](old CredentialRef[Old], convert func(Old) T) CredentialRef[T] {
	r.replacements = replacing(r.replacements, old.String(), convert)

	return r
}

// Backfilled declares how a stored payload missing values is completed
func (r CredentialRef[T]) Backfilled(fn func(context.Context, InstallationRequest, *T) error) CredentialRef[T] {
	r.backfill = fn

	return r
}

// Replaces lists the retired slots whose stored payloads this slot takes over, sorted
func (r CredentialRef[T]) Replaces() []CredentialSlotID {
	return lo.Map(sortedNames(r.replacements), func(name string, _ int) CredentialSlotID {
		return NewCredentialSlotID(name)
	})
}

// Convert reshapes a payload stored under one of the replaced slots into this slot's shape
func (r CredentialRef[T]) Convert(from CredentialSlotID, old json.RawMessage) (json.RawMessage, error) {
	return convertReplaced(r.replacements, from.String(), old)
}

// Backfill completes a stored payload missing values through the declared backfill
func (r CredentialRef[T]) Backfill(ctx context.Context, req InstallationRequest, payload json.RawMessage) (json.RawMessage, error) {
	return backfillPayload(ctx, req, r.backfill, payload)
}

// Registration projects the slot identity and declared lifecycle onto base, leaving Schema and the descriptive fields as authored
func (r CredentialRef[T]) Registration(base CredentialRegistration) CredentialRegistration {
	base.Ref = r.ID()

	if len(r.replacements) > 0 {
		base.Replaces = r.Replaces()
		base.Convert = r.Convert
	}

	if r.backfill != nil {
		base.Backfill = r.Backfill
	}

	return base
}

// =========
// User input
// =========

// UserInputRef is a typed handle for one definition's installation-scoped user input layout
type UserInputRef[T any] struct {
	name         string
	replacements map[string]replacement[T]
	backfill     func(context.Context, InstallationRequest, *T) error
}

// NewUserInputRef creates a typed user input layout handle
func NewUserInputRef[T any](name string) UserInputRef[T] {
	return UserInputRef[T]{name: name}
}

// Name returns the stable layout name
func (r UserInputRef[T]) Name() string {
	return r.name
}

// Replacing declares that the layout takes over user input stored in old
func (r UserInputRef[T]) Replacing[Old any](old UserInputRef[Old], convert func(Old) T) UserInputRef[T] {
	r.replacements = replacing(r.replacements, old.name, convert)

	return r
}

// Backfilled declares how stored user input missing values is completed
func (r UserInputRef[T]) Backfilled(fn func(context.Context, InstallationRequest, *T) error) UserInputRef[T] {
	r.backfill = fn

	return r
}

// Replaces lists the retired layout names whose stored user input this layout takes over, sorted
func (r UserInputRef[T]) Replaces() []string {
	return sortedNames(r.replacements)
}

// Convert reshapes stored user input through the first retired layout it matches
func (r UserInputRef[T]) Convert(old json.RawMessage) (json.RawMessage, error) {
	for _, retired := range sortedNames(r.replacements) {
		converted, err := convertReplaced(r.replacements, retired, old)
		if err == nil {
			return converted, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrLayoutMismatch, r.name)
}

// Backfill completes stored user input missing values through the declared backfill
func (r UserInputRef[T]) Backfill(ctx context.Context, req InstallationRequest, payload json.RawMessage) (json.RawMessage, error) {
	return backfillPayload(ctx, req, r.backfill, payload)
}

// =========
// Clients
// =========

// ClientID is the in-process identity of one registered client, unique within a definition
type ClientID struct {
	// name is the client name used for registry indexing and client caching
	name string
}

// NewClientID creates a client identity from its name
func NewClientID(name string) ClientID {
	return ClientID{name: name}
}

// Valid reports whether the client identity was initialized
func (id ClientID) Valid() bool {
	return id.name != ""
}

// String returns the client name used for cache indexing
func (id ClientID) String() string {
	return id.name
}

// ClientRef is a typed handle for one registered client identity
type ClientRef[C any] struct {
	id    ClientID
	slots []CredentialSlotID
}

// NewClientRef creates a typed client identity handle with a literal name
func NewClientRef[C any](name string) ClientRef[C] {
	return ClientRef[C]{id: NewClientID(name)}
}

// ClientRefOf creates a typed client identity handle named after the client type
func ClientRefOf[C any]() ClientRef[C] {
	clientType := reflect.TypeFor[C]()

	for clientType.Kind() == reflect.Pointer {
		clientType = clientType.Elem()
	}

	return NewClientRef[C](clientType.String())
}

// ID returns the opaque client identity
func (r ClientRef[C]) ID() ClientID {
	return r.id
}

// Cast type-asserts a registered client instance to the typed client value
func (r ClientRef[C]) Cast(client any) (C, error) {
	c, ok := client.(C)
	if !ok {
		var zero C
		return zero, ErrClientCastFailed
	}
	return c, nil
}

// Using declares that the client is built from the given credential slot
func (r ClientRef[C]) Using[T any](cred CredentialRef[T]) ClientRef[C] {
	r.slots = append(slices.Clone(r.slots), cred.ID())

	return r
}

// Registration projects the client identity, its credential slots, and the typed build function onto base
func (r ClientRef[C]) Registration(build func(context.Context, ClientBuildRequest) (C, error), base ClientRegistration) ClientRegistration {
	base.Ref = r.ID()
	base.CredentialRefs = slices.Clone(r.slots)
	base.Build = func(ctx context.Context, req ClientBuildRequest) (any, error) {
		client, err := build(ctx, req)

		return client, err
	}

	return base
}

// =========
// OperationRef
// =========

// OperationRef is a typed handle for one registered operation identity, parameterized by the operation config type
type OperationRef[Cfg any] struct {
	// name is the stable operation name used for persistence and topic derivation
	name string
	// schema is the reflected JSON schema of the config type
	schema json.RawMessage
	// client is the registered client the operation runs against, invalid when the operation has none
	client ClientID
}

// NewOperationRef creates a typed operation identity handle with the schema reflected from Cfg
func NewOperationRef[Cfg any](name string) OperationRef[Cfg] {
	return OperationRef[Cfg]{name: name, schema: jsonx.SchemaFrom[Cfg]()}
}

// OperationRefOf creates a typed operation identity handle named after the reflected schema of Cfg
func OperationRefOf[Cfg any]() OperationRef[Cfg] {
	schema := jsonx.SchemaFrom[Cfg]()

	return OperationRef[Cfg]{name: jsonx.SchemaID(schema), schema: schema}
}

// Name returns the stable operation name
func (r OperationRef[Cfg]) Name() string {
	return r.name
}

// Schema returns a copy of the reflected JSON schema of the config type
func (r OperationRef[Cfg]) Schema() json.RawMessage {
	return jsonx.CloneRawMessage(r.schema)
}

// UnmarshalConfig decodes a JSON operation config document into the typed config value
func (r OperationRef[Cfg]) UnmarshalConfig(raw json.RawMessage) (Cfg, error) {
	var out Cfg

	return out, jsonx.UnmarshalIfPresent(raw, &out)
}

// Using declares that the operation runs against the given client
func (r OperationRef[Cfg]) Using[C any](client ClientRef[C]) OperationRef[Cfg] {
	r.client = client.ID()

	return r
}

// ConfigFrom returns an OperationRegistration.ConfigResolver that decodes the installation's user input
// as U, extracts this operation's config through fn, and re-encodes it as JSON
func (r OperationRef[Cfg]) ConfigFrom[U any](fn func(U) Cfg) func(json.RawMessage) json.RawMessage {
	return func(userInput json.RawMessage) json.RawMessage {
		var input U
		if err := json.Unmarshal(userInput, &input); err != nil {
			return nil
		}

		out, err := json.Marshal(fn(input))
		if err != nil {
			return nil
		}

		return out
	}
}

// Registration projects the operation name, topic, config schema, and declared client onto base
func (r OperationRef[Cfg]) Registration(definition DefinitionRef, base OperationRegistration) OperationRegistration {
	base.Name = r.name
	base.Topic = definition.OperationTopic(r.name)
	base.ConfigSchema = r.Schema()

	if r.client.Valid() {
		base.ClientRef = r.client
	}

	return base
}

// =========
// Installations
// =========

// InstallationRef is a typed handle for one definition's installation metadata derivation
type InstallationRef[T any] struct {
	schema json.RawMessage
	fn     func(ctx context.Context, req InstallationRequest) (T, bool, error)
}

// NewInstallationRef creates a typed installation metadata handle
func NewInstallationRef[T any](fn func(ctx context.Context, req InstallationRequest) (T, bool, error)) InstallationRef[T] {
	return InstallationRef[T]{schema: jsonx.SchemaFrom[T](), fn: fn}
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
	return &InstallationRegistration{Resolve: r.Resolve, Schema: jsonx.CloneRawMessage(r.schema)}
}

// =========
// Connections
// =========

// ConnectionRef ties one connection mode's primary credential slot to the clients it initializes
type ConnectionRef struct {
	// credential is the credential slot that selects the connection mode
	credential CredentialSlotID
	// clients lists the clients the connection mode initializes
	clients []ClientID
}

// NewConnectionRef creates a connection mode handle selected by the given credential slot
func NewConnectionRef[T any](cred CredentialRef[T]) ConnectionRef {
	return ConnectionRef{credential: cred.ID()}
}

// Enables declares that the connection mode initializes the given client
func (r ConnectionRef) Enables[C any](client ClientRef[C]) ConnectionRef {
	r.clients = append(slices.Clone(r.clients), client.ID())

	return r
}

// Registration projects the selecting credential slot and enabled clients onto base, leaving CredentialRefs as authored
func (r ConnectionRef) Registration(base ConnectionRegistration) ConnectionRegistration {
	base.CredentialRef = r.credential
	base.ClientRefs = slices.Clone(r.clients)

	return base
}

// =========
// Webhooks
// =========

// WebhookRef is a handle for one registered webhook contract identity
type WebhookRef struct {
	// name is the stable webhook contract name used for persistence
	name string
}

// NewWebhookRef creates a webhook contract identity handle
func NewWebhookRef(name string) WebhookRef {
	return WebhookRef{name: name}
}

// Name returns the stable webhook contract name
func (r WebhookRef) Name() string {
	return r.name
}

// Registration projects the webhook contract name onto base
func (r WebhookRef) Registration(base WebhookRegistration) WebhookRegistration {
	base.Name = r.name

	return base
}

// WebhookEventRef is a typed handle for one registered webhook event identity, parameterized by the payload type
type WebhookEventRef[T any] struct {
	// name is the stable webhook event name used for persistence and topic derivation
	name string
}

// NewWebhookEventRef creates a typed webhook event identity handle
func NewWebhookEventRef[T any](name string) WebhookEventRef[T] {
	return WebhookEventRef[T]{name: name}
}

// WebhookEventRefOf creates a typed webhook event identity handle named after the reflected schema of T
func WebhookEventRefOf[T any]() WebhookEventRef[T] {
	return WebhookEventRef[T]{name: jsonx.SchemaID(jsonx.SchemaFrom[T]())}
}

// Name returns the stable webhook event name
func (r WebhookEventRef[T]) Name() string {
	return r.name
}

// UnmarshalPayload decodes a JSON webhook event payload into the typed payload value
func (r WebhookEventRef[T]) UnmarshalPayload(raw json.RawMessage) (T, error) {
	var out T
	return out, jsonx.UnmarshalIfPresent(raw, &out)
}

// Registration projects the event name and its topic under definition onto base
func (r WebhookEventRef[T]) Registration(definition DefinitionRef, base WebhookEventRegistration) WebhookEventRegistration {
	base.Name = r.name
	base.Topic = definition.WebhookEventTopic(r.name)

	return base
}
