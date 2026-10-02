package types //nolint:revive

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"

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

// storedLayout is a stored document's reflected schema, the upgrade that moves older documents onto it, and its semantic validation
type storedLayout[T any] struct {
	// schema is the reflected JSON schema of T
	schema json.RawMessage
	// upgrade reshapes a stored document from the layout it was persisted under into T, nil when none is declared
	upgrade UpgradeFunc
	// validate checks a schema-valid document for constraints the schema cannot express, nil when none is declared
	validate ValidateFunc
}

// newStoredLayout creates a stored layout with the schema reflected from T
func newStoredLayout[T any]() storedLayout[T] {
	return storedLayout[T]{schema: jsonx.SchemaFrom[T]()}
}

// validated binds a typed validation closure as a ValidateFunc, decoding the payload into T first
func validated[T any](fn func(context.Context, InstallationRequest, *T) error) ValidateFunc {
	return func(ctx context.Context, req InstallationRequest, payload json.RawMessage) error {
		var value T

		if err := jsonx.UnmarshalIfPresent(payload, &value); err != nil {
			return err
		}

		return fn(ctx, req, &value)
	}
}

// upgraded binds a typed upgrade closure as an UpgradeFunc, marshalling its result
func upgraded[T any](fn func(context.Context, InstallationRequest, string, json.RawMessage) (T, error)) UpgradeFunc {
	return func(ctx context.Context, req InstallationRequest, from string, stored json.RawMessage) (json.RawMessage, error) {
		value, err := fn(ctx, req, from, stored)
		if err != nil {
			return nil, err
		}

		return jsonx.ToRawMessage(value)
	}
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
	// storedLayout is the credential type's reflected schema and declared upgrade
	storedLayout[T]
	// replaces lists the retired slots whose stored payloads move onto this slot
	replaces []CredentialSlotID
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

// Replacing declares that the slot takes over the payloads stored under old
func (r CredentialRef[T]) Replacing[Old any](old CredentialRef[Old]) CredentialRef[T] {
	r.replaces = append(slices.Clone(r.replaces), old.ID())

	return r
}

// Upgraded declares how a stored payload is reshaped from the slot it was persisted under into T
func (r CredentialRef[T]) Upgraded(fn func(context.Context, InstallationRequest, string, json.RawMessage) (T, error)) CredentialRef[T] {
	r.upgrade = upgraded(fn)

	return r
}

// Validated declares a semantic check run on a schema-valid credential payload decoded as T
func (r CredentialRef[T]) Validated(fn func(context.Context, InstallationRequest, *T) error) CredentialRef[T] {
	r.validate = validated(fn)

	return r
}

// Replaces lists the retired slots whose stored payloads this slot takes over, sorted
func (r CredentialRef[T]) Replaces() []CredentialSlotID {
	return lo.Map(sortedUnique(lo.Map(r.replaces, func(slot CredentialSlotID, _ int) string { return slot.String() })), func(name string, _ int) CredentialSlotID {
		return NewCredentialSlotID(name)
	})
}

// Registration projects the slot identity and lifecycle onto base
func (r CredentialRef[T]) Registration(base CredentialRegistration) CredentialRegistration {
	base.Ref = r.ID()
	base.StoredSchema = r.Schema()
	base.Upgrade = r.upgrade
	base.Validate = r.validate
	base.Replaces = nil

	if len(r.replaces) > 0 {
		base.Replaces = r.Replaces()
	}

	return base
}

// =========
// User input
// =========

// UserInputRef is a typed handle for one definition's installation-scoped user input layout
type UserInputRef[T any] struct {
	// name is the stable layout name the stored document is keyed by, reflected from T
	name string
	// storedLayout is the user input type's reflected schema and declared upgrade
	storedLayout[T]
}

// UserInputRefOf creates a typed user input layout handle named after T's reflected schema
func UserInputRefOf[T any]() UserInputRef[T] {
	layout := newStoredLayout[T]()

	return UserInputRef[T]{name: jsonx.SchemaID(layout.schema), storedLayout: layout}
}

// Name returns the stable layout name the stored document is keyed by
func (r UserInputRef[T]) Name() string {
	return r.name
}

// Upgraded declares how stored user input is reshaped from the layout it was persisted under into T
func (r UserInputRef[T]) Upgraded(fn func(context.Context, InstallationRequest, string, json.RawMessage) (T, error)) UserInputRef[T] {
	r.upgrade = upgraded(fn)

	return r
}

// Validated declares a semantic check run on schema-valid user input decoded as T
func (r UserInputRef[T]) Validated(fn func(context.Context, InstallationRequest, *T) error) UserInputRef[T] {
	r.validate = validated(fn)

	return r
}

// Registration projects the layout name, reflected schema, declared upgrade, and validation into an input registration
func (r UserInputRef[T]) Registration() *InputRegistration {
	return &InputRegistration{Name: r.name, Schema: jsonx.CloneRawMessage(r.schema), Upgrade: r.upgrade, Validate: r.validate}
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
type OperationRef[Config any] struct {
	// name is the stable operation name used for persistence and topic derivation
	name string
	// storedLayout is the config type's reflected schema and declared upgrade
	storedLayout[Config]
	// client is the registered client the operation runs against, invalid when the operation has none
	client ClientID
	// handle executes the operation when it does not produce ingest payloads
	handle OperationHandler
	// ingest executes the operation and returns typed payload sets for the ingest pipeline
	ingest IngestHandler
	// replaces lists the retired operation names this operation takes over
	replaces []string
	// policy controls synchronous execution behavior for the operation
	policy ExecutionPolicy
	// contracts declares the normalized schemas emitted by the operation
	contracts []IngestContract
	// permissions lists scopes or permissions needed to retrieve data for the operation
	permissions []string
	// schedule overrides the default adaptive schedule for reconcile or scheduled cycles
	schedule *gala.Schedule
	// skipDefaultLookback disables the runtime's default lookback window on initial runs
	skipDefaultLookback bool
	// rateLimit bounds how often the operation may run per organization
	rateLimit *RateLimitPolicy
	// internal marks the operation as reachable only through its own listener or saga machinery
	internal bool
	// customerSelectable controls whether the operation is exposed in customer-facing surfaces
	customerSelectable *bool
	// requiresPaymentMethod gates direct invocation on the org having a payment method on file
	requiresPaymentMethod bool
	// disabledForAll marks the operation unavailable for every installation
	disabledForAll bool
	// stored reports whether the operation keeps a per-installation input document under its name
	stored bool
}

// NewOperationRef creates a stored-input operation handle with the given name; Config embeds OperationSettings and
// its reflected schema is the per-installation input document stored under the operation name
func NewOperationRef[Config OperationInput](name string) OperationRef[Config] {
	return OperationRef[Config]{name: name, storedLayout: newStoredLayout[Config](), stored: true}
}

// OperationRefOf creates a stored-input operation handle named after the reflected schema of Config; Config embeds
// OperationSettings and its reflected schema is the per-installation input document stored under the operation name
func OperationRefOf[Config OperationInput]() OperationRef[Config] {
	layout := newStoredLayout[Config]()

	return OperationRef[Config]{name: jsonx.SchemaID(layout.schema), storedLayout: layout, stored: true}
}

// NewOperationPayload creates a payload operation handle with the given name; Config is the payload a caller
// supplies on each dispatch, nothing is stored per installation, and Upgraded and Validated are ignored
func NewOperationPayload[Config any](name string) OperationRef[Config] {
	return OperationRef[Config]{name: name, storedLayout: newStoredLayout[Config]()}
}

// OperationPayloadOf creates a payload operation handle named after the reflected schema of Config; Config is the
// payload a caller supplies on each dispatch, nothing is stored per installation, and Upgraded and Validated are ignored
func OperationPayloadOf[Config any]() OperationRef[Config] {
	layout := newStoredLayout[Config]()

	return OperationRef[Config]{name: jsonx.SchemaID(layout.schema), storedLayout: layout}
}

// decodeConfig decodes an operation config payload into Config, treating absent as the zero value
func decodeConfig[Config any](raw json.RawMessage) (Config, error) {
	var cfg Config

	if err := jsonx.UnmarshalIfPresent(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("%w: %w", ErrOperationConfigInvalid, err)
	}

	return cfg, nil
}

// bindRequest casts the request client through client and decodes the request config into Config
func bindRequest[C, Config any](client ClientRef[C], request OperationRequest) (C, Config, error) {
	typed, err := client.Cast(request.Client)
	if err != nil {
		var cfg Config

		return typed, cfg, err
	}

	cfg, err := decodeConfig[Config](request.Config)

	return typed, cfg, err
}

// sortedUnique returns the names sorted with duplicates removed
func sortedUnique(names []string) []string {
	return slices.Compact(slices.Sorted(slices.Values(names)))
}

// Name returns the stable operation name
func (r OperationRef[Config]) Name() string {
	return r.name
}

// Schema returns a copy of the reflected JSON schema of the config type
func (r OperationRef[Config]) Schema() json.RawMessage {
	return jsonx.CloneRawMessage(r.schema)
}

// Ingests binds fn as the ingest handler, run against client with the decoded config
func (r OperationRef[Config]) Ingests[C any](client ClientRef[C], fn func(context.Context, OperationRequest, C, Config) ([]IngestPayloadSet, error)) OperationRef[Config] {
	r.client = client.ID()
	r.ingest = func(ctx context.Context, request OperationRequest) ([]IngestPayloadSet, error) {
		typed, cfg, err := bindRequest[C, Config](client, request)
		if err != nil {
			return nil, err
		}

		return fn(ctx, request, typed, cfg)
	}

	return r
}

// Handles binds fn as the handler, run against client with the decoded config
func (r OperationRef[Config]) Handles[C any](client ClientRef[C], fn func(context.Context, OperationRequest, C, Config) (json.RawMessage, error)) OperationRef[Config] {
	r.client = client.ID()
	r.handle = func(ctx context.Context, request OperationRequest) (json.RawMessage, error) {
		typed, cfg, err := bindRequest[C, Config](client, request)
		if err != nil {
			return nil, err
		}

		return fn(ctx, request, typed, cfg)
	}

	return r
}

// HandlesRequest binds fn as the handler, run without a client, with the decoded config
func (r OperationRef[Config]) HandlesRequest(fn func(context.Context, OperationRequest, Config) (json.RawMessage, error)) OperationRef[Config] {
	r.handle = func(ctx context.Context, request OperationRequest) (json.RawMessage, error) {
		cfg, err := decodeConfig[Config](request.Config)
		if err != nil {
			return nil, err
		}

		return fn(ctx, request, cfg)
	}

	return r
}

// Replacing declares that the operation takes over the stored input, recorded runs, and health of old
func (r OperationRef[Config]) Replacing[Old any](old OperationRef[Old]) OperationRef[Config] {
	r.replaces = append(slices.Clone(r.replaces), old.Name())

	return r
}

// Upgraded declares how a stored input document is reshaped from the operation it was persisted under into Config; ignored on a payload operation
func (r OperationRef[Config]) Upgraded(fn func(context.Context, InstallationRequest, string, json.RawMessage) (Config, error)) OperationRef[Config] {
	r.upgrade = upgraded(fn)

	return r
}

// Validated declares a semantic check run on a schema-valid stored input document decoded as Config; ignored on a payload operation
func (r OperationRef[Config]) Validated(fn func(context.Context, InstallationRequest, *Config) error) OperationRef[Config] {
	r.validate = validated(fn)

	return r
}

// Policy declares the execution policy of the operation
func (r OperationRef[Config]) Policy(policy ExecutionPolicy) OperationRef[Config] {
	r.policy = policy

	return r
}

// Ingest declares the normalized schemas emitted by the operation
func (r OperationRef[Config]) Ingest(contracts ...IngestContract) OperationRef[Config] {
	r.contracts = append(slices.Clone(r.contracts), contracts...)

	return r
}

// Permissions declares the scopes or permissions needed to retrieve data for the operation
func (r OperationRef[Config]) Permissions(permissions ...string) OperationRef[Config] {
	r.permissions = append(slices.Clone(r.permissions), permissions...)

	return r
}

// Schedule overrides the default adaptive schedule for reconcile or scheduled cycles
func (r OperationRef[Config]) Schedule(schedule *gala.Schedule) OperationRef[Config] {
	r.schedule = schedule

	return r
}

// SkipDefaultLookback disables the runtime's default lookback window on initial runs
func (r OperationRef[Config]) SkipDefaultLookback() OperationRef[Config] {
	r.skipDefaultLookback = true

	return r
}

// RateLimit bounds how often the operation may run per organization
func (r OperationRef[Config]) RateLimit(policy RateLimitPolicy) OperationRef[Config] {
	r.rateLimit = &policy

	return r
}

// Internal marks the operation as reachable only through its own listener or saga machinery
func (r OperationRef[Config]) Internal() OperationRef[Config] {
	r.internal = true

	return r
}

// CustomerSelectable controls whether the operation is exposed in customer-facing surfaces
func (r OperationRef[Config]) CustomerSelectable(selectable bool) OperationRef[Config] {
	r.customerSelectable = &selectable

	return r
}

// RequiresPaymentMethod gates direct invocation on the org having a payment method on file
func (r OperationRef[Config]) RequiresPaymentMethod() OperationRef[Config] {
	r.requiresPaymentMethod = true

	return r
}

// DisabledForAll marks the operation unavailable for every installation when disabled is true
func (r OperationRef[Config]) DisabledForAll(disabled bool) OperationRef[Config] {
	r.disabledForAll = disabled

	return r
}

// Registration projects the operation's identity, handler, stored input, and declared behavior onto base
func (r OperationRef[Config]) Registration(definition DefinitionRef, base OperationRegistration) OperationRegistration {
	base.Name = r.name
	base.Topic = definition.OperationTopic(r.name)
	base.ConfigSchema = r.Schema()
	base.Policy = r.policy
	base.Ingest = slices.Clone(r.contracts)
	base.RequiredPermissions = slices.Clone(r.permissions)
	base.Schedule = r.schedule
	base.SkipDefaultLookback = r.skipDefaultLookback
	base.RateLimit = r.rateLimit
	base.Internal = r.internal
	base.CustomerSelectable = r.customerSelectable
	base.RequiresPaymentMethod = r.requiresPaymentMethod
	base.DisabledForAll = r.disabledForAll

	if r.stored {
		base.Input = &InputRegistration{Name: r.name, Schema: r.Schema(), Upgrade: r.upgrade, Validate: r.validate}
	}

	if r.client.Valid() {
		base.ClientRef = r.client
	}

	if r.handle != nil {
		base.Handle = r.handle
	}

	if r.ingest != nil {
		base.IngestHandle = r.ingest
	}

	base.Replaces = nil

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
	base.Replaces = nil

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
