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

// storedLayout is a stored payload's reflected schema with retired layouts and backfill
type storedLayout[T any] struct {
	// schema is the reflected JSON schema of T
	schema json.RawMessage
	// replacements are the retired layouts this layout takes over, keyed by retired name
	replacements map[string]replacement[T]
	// backfill completes a stored payload missing values
	backfill func(context.Context, InstallationRequest, *T) error
}

// newStoredLayout creates a stored layout with the schema reflected from T
func newStoredLayout[T any]() storedLayout[T] {
	return storedLayout[T]{schema: jsonx.SchemaFrom[T]()}
}

// replacing records a retired layout decoded as Old and reshaped through convert
func (l storedLayout[T]) replacing[Old any](name string, convert func(Old) T) storedLayout[T] {
	l.replacements = maps.Clone(l.replacements)
	if l.replacements == nil {
		l.replacements = map[string]replacement[T]{}
	}

	l.replacements[name] = replacement[T]{schema: jsonx.SchemaFrom[Old](), decode: func(payload json.RawMessage) (T, error) {
		previous, err := jsonx.Decode[Old](payload)
		if err != nil {
			var zero T

			return zero, err
		}

		return convert(previous), nil
	}}

	return l
}

// retired lists the retired layout names this layout takes over, sorted
func (l storedLayout[T]) retired() []string {
	return slices.Sorted(maps.Keys(l.replacements))
}

// convert reshapes a payload stored under the retired layout from into T
func (l storedLayout[T]) convert(from string, old json.RawMessage) (json.RawMessage, error) {
	retired, ok := l.replacements[from]
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

// Backfill completes a stored payload through the declared backfill, or returns it unchanged
func (l storedLayout[T]) Backfill(ctx context.Context, req InstallationRequest, payload json.RawMessage) (json.RawMessage, error) {
	if l.backfill == nil {
		return payload, nil
	}

	var value T

	if err := jsonx.UnmarshalIfPresent(payload, &value); err != nil {
		return nil, err
	}

	if err := l.backfill(ctx, req, &value); err != nil {
		return nil, err
	}

	return jsonx.ToRawMessage(value)
}

// =========
// Credentials
// =========

// CredentialSlotID is the non-generic durable identity for one credential slot used by a definition
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

	*r = NewCredentialSlotID(name)

	return nil
}

// CredentialRef is a typed handle for one credential slot, parameterized by its schema type
type CredentialRef[T any] struct {
	// id is the durable credential slot identity
	id CredentialSlotID
	// storedLayout is the credential type's reflected schema with retired slots and backfill
	storedLayout[T]
}

// NewCredentialRef creates a typed credential slot identity handle with the schema reflected from T
func NewCredentialRef[T any](name string) CredentialRef[T] {
	return CredentialRef[T]{id: NewCredentialSlotID(name), storedLayout: newStoredLayout[T]()}
}

// CredentialRefOf creates a typed credential slot handle named after T's reflected schema
func CredentialRefOf[T any]() CredentialRef[T] {
	layout := newStoredLayout[T]()

	return CredentialRef[T]{id: NewCredentialSlotID(jsonx.SchemaID(layout.schema)), storedLayout: layout}
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

	out, err := jsonx.Decode[T](cred.Data)

	return out, true, err
}

// Replacing declares that the slot takes over payloads stored under old
func (r CredentialRef[T]) Replacing[Old any](old CredentialRef[Old], convert func(Old) T) CredentialRef[T] {
	r.storedLayout = r.replacing(old.String(), convert)

	return r
}

// Backfilled declares how a stored payload missing values is completed
func (r CredentialRef[T]) Backfilled(fn func(context.Context, InstallationRequest, *T) error) CredentialRef[T] {
	r.backfill = fn

	return r
}

// Replaces lists the retired slots whose stored payloads this slot takes over, sorted
func (r CredentialRef[T]) Replaces() []CredentialSlotID {
	return lo.Map(r.retired(), func(name string, _ int) CredentialSlotID {
		return NewCredentialSlotID(name)
	})
}

// Convert reshapes a payload stored under one of the replaced slots into this slot's shape
func (r CredentialRef[T]) Convert(from CredentialSlotID, old json.RawMessage) (json.RawMessage, error) {
	return r.convert(from.String(), old)
}

// Registration projects the slot identity and lifecycle onto base
func (r CredentialRef[T]) Registration(base CredentialRegistration) CredentialRegistration {
	base.Ref = r.ID()
	base.StoredSchema = r.Schema()

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
	// name is the stable layout name a later layout retires this one by
	name string
	// storedLayout is the user input type's reflected schema with retired layouts and backfill
	storedLayout[T]
}

// NewUserInputRef creates a typed user input layout handle with the schema reflected from T
func NewUserInputRef[T any](name string) UserInputRef[T] {
	return UserInputRef[T]{name: name, storedLayout: newStoredLayout[T]()}
}

// Replacing declares that the layout takes over user input stored in old
func (r UserInputRef[T]) Replacing[Old any](old UserInputRef[Old], convert func(Old) T) UserInputRef[T] {
	r.storedLayout = r.replacing(old.name, convert)

	return r
}

// Backfilled declares how stored user input missing values is completed
func (r UserInputRef[T]) Backfilled(fn func(context.Context, InstallationRequest, *T) error) UserInputRef[T] {
	r.backfill = fn

	return r
}

// Replaces lists the retired layout names whose stored user input this layout takes over, sorted
func (r UserInputRef[T]) Replaces() []string {
	return r.retired()
}

// Convert reshapes stored user input through the first retired layout it matches
func (r UserInputRef[T]) Convert(old json.RawMessage) (json.RawMessage, error) {
	for _, retired := range r.retired() {
		converted, err := r.convert(retired, old)
		if err == nil {
			return converted, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrLayoutMismatch, r.name)
}

// Registration projects the reflected schema and declared lifecycle into a user input registration
func (r UserInputRef[T]) Registration() *UserInputRegistration {
	reg := &UserInputRegistration{Schema: jsonx.CloneRawMessage(r.schema)}

	if len(r.replacements) > 0 {
		reg.Replaces = r.Replaces()
		reg.Convert = r.Convert
	}

	if r.backfill != nil {
		reg.Backfill = r.Backfill
	}

	return reg
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

// Registration projects the client identity, credential slots, and build function onto base
func (r ClientRef[C]) Registration(build func(context.Context, ClientBuildRequest) (C, error), base ClientRegistration) ClientRegistration {
	base.Ref = r.ID()
	base.CredentialRefs = slices.Clone(r.slots)
	base.Build = func(ctx context.Context, req ClientBuildRequest) (any, error) {
		client, err := build(ctx, req)

		return client, err
	}

	return base
}

// HealthCheck binds fn as a definition health check that runs against this client
func (r ClientRef[C]) HealthCheck(fn func(context.Context, OperationRequest, C) (json.RawMessage, error)) *HealthCheckRegistration {
	return &HealthCheckRegistration{
		ClientRef: r.ID(),
		Handle: func(ctx context.Context, request OperationRequest) (json.RawMessage, error) {
			client, err := r.Cast(request.Client)
			if err != nil {
				return nil, err
			}

			return fn(ctx, request, client)
		},
	}
}

// CredentialHealthCheck binds fn as a health check receiving only credential bindings
func CredentialHealthCheck(fn func(context.Context, OperationRequest) (json.RawMessage, error)) *HealthCheckRegistration {
	return &HealthCheckRegistration{Handle: fn}
}

// =========
// OperationRef
// =========

// OperationRef is a typed handle for one operation identity, parameterized by its config type
type OperationRef[Cfg any] struct {
	// name is the stable operation name used for persistence and topic derivation
	name string
	// schema is the reflected JSON schema of the config type
	schema json.RawMessage
	// client is the registered client the operation runs against, invalid when the operation has none
	client ClientID
	// handle executes the operation when it does not produce ingest payloads
	handle OperationHandler
	// ingest executes the operation and returns typed payload sets for the ingest pipeline
	ingest IngestHandler
	// replaces lists the retired operation names this operation takes over
	replaces []string
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

// decodeConfig decodes an operation config payload into Cfg, treating absent as the zero value
func decodeConfig[Cfg any](raw json.RawMessage) (Cfg, error) {
	var cfg Cfg

	if err := jsonx.UnmarshalIfPresent(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("%w: %w", ErrOperationConfigInvalid, err)
	}

	return cfg, nil
}

// bindRequest casts the request client through client and decodes the request config into Cfg
func bindRequest[C, Cfg any](client ClientRef[C], request OperationRequest) (C, Cfg, error) {
	typed, err := client.Cast(request.Client)
	if err != nil {
		var cfg Cfg

		return typed, cfg, err
	}

	cfg, err := decodeConfig[Cfg](request.Config)

	return typed, cfg, err
}

// sortedUnique returns the names sorted with duplicates removed
func sortedUnique(names []string) []string {
	return slices.Compact(slices.Sorted(slices.Values(names)))
}

// Name returns the stable operation name
func (r OperationRef[Cfg]) Name() string {
	return r.name
}

// Schema returns a copy of the reflected JSON schema of the config type
func (r OperationRef[Cfg]) Schema() json.RawMessage {
	return jsonx.CloneRawMessage(r.schema)
}

// Ingests binds fn as the ingest handler, run against client with the decoded config
func (r OperationRef[Cfg]) Ingests[C any](client ClientRef[C], fn func(context.Context, OperationRequest, C, Cfg) ([]IngestPayloadSet, error)) OperationRef[Cfg] {
	r.client = client.ID()
	r.ingest = func(ctx context.Context, request OperationRequest) ([]IngestPayloadSet, error) {
		typed, cfg, err := bindRequest[C, Cfg](client, request)
		if err != nil {
			return nil, err
		}

		return fn(ctx, request, typed, cfg)
	}

	return r
}

// Handles binds fn as the handler, run against client with the decoded config
func (r OperationRef[Cfg]) Handles[C any](client ClientRef[C], fn func(context.Context, OperationRequest, C, Cfg) (json.RawMessage, error)) OperationRef[Cfg] {
	r.client = client.ID()
	r.handle = func(ctx context.Context, request OperationRequest) (json.RawMessage, error) {
		typed, cfg, err := bindRequest[C, Cfg](client, request)
		if err != nil {
			return nil, err
		}

		return fn(ctx, request, typed, cfg)
	}

	return r
}

// HandlesRequest binds fn as the handler, run without a client, with the decoded config
func (r OperationRef[Cfg]) HandlesRequest(fn func(context.Context, OperationRequest, Cfg) (json.RawMessage, error)) OperationRef[Cfg] {
	r.handle = func(ctx context.Context, request OperationRequest) (json.RawMessage, error) {
		cfg, err := decodeConfig[Cfg](request.Config)
		if err != nil {
			return nil, err
		}

		return fn(ctx, request, cfg)
	}

	return r
}

// Replacing declares that the operation takes over the recorded runs and health of old
func (r OperationRef[Cfg]) Replacing[Old any](old OperationRef[Old]) OperationRef[Cfg] {
	r.replaces = append(slices.Clone(r.replaces), old.Name())

	return r
}

// Registration projects the operation's name, topic, schema, client, and handler onto base
func (r OperationRef[Cfg]) Registration(definition DefinitionRef, base OperationRegistration) OperationRegistration {
	base.Name = r.name
	base.Topic = definition.OperationTopic(r.name)
	base.ConfigSchema = r.Schema()

	if r.client.Valid() {
		base.ClientRef = r.client
	}

	if r.handle != nil {
		base.Handle = r.handle
	}

	if r.ingest != nil {
		base.IngestHandle = r.ingest
	}

	if len(r.replaces) > 0 {
		base.Replaces = sortedUnique(r.replaces)
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

// Registration adapts the typed ref to the InstallationRegistration contract
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

// Registration projects the credential slot, disconnect flow, and enabled clients onto base
func (r ConnectionRef) Registration(base ConnectionRegistration) ConnectionRegistration {
	base.CredentialRef = r.credential
	base.CredentialRefs = []CredentialSlotID{r.credential}
	base.ClientRefs = slices.Clone(r.clients)

	if base.Disconnect != nil {
		disconnect := *base.Disconnect
		disconnect.CredentialRef = r.credential
		base.Disconnect = &disconnect
	}

	return base
}

// =========
// Webhooks
// =========

// WebhookRef is a handle for one registered webhook contract identity
type WebhookRef struct {
	// name is the stable webhook contract name used for persistence
	name string
	// replaces lists the retired contract names whose persisted webhook rows move onto this contract
	replaces []string
}

// NewWebhookRef creates a webhook contract identity handle
func NewWebhookRef(name string) WebhookRef {
	return WebhookRef{name: name}
}

// Name returns the stable webhook contract name
func (r WebhookRef) Name() string {
	return r.name
}

// Replacing declares that the contract takes over the persisted webhook rows of old
func (r WebhookRef) Replacing(old WebhookRef) WebhookRef {
	r.replaces = append(slices.Clone(r.replaces), old.Name())

	return r
}

// Registration projects the webhook contract name and retired names onto base
func (r WebhookRef) Registration(base WebhookRegistration) WebhookRegistration {
	base.Name = r.name

	if len(r.replaces) > 0 {
		base.Replaces = sortedUnique(r.replaces)
	}

	return base
}

// WebhookEventRef is a typed handle for one webhook event, parameterized by payload type
type WebhookEventRef[T any] struct {
	// name is the stable webhook event name used for persistence and topic derivation
	name string
}

// NewWebhookEventRef creates a typed webhook event identity handle
func NewWebhookEventRef[T any](name string) WebhookEventRef[T] {
	return WebhookEventRef[T]{name: name}
}

// WebhookEventRefOf creates a typed webhook event handle named after T's reflected schema
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
