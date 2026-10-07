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

// reflectedInput creates an input registration with the schema reflected from T, named name or after the reflected schema when name is empty
func reflectedInput[T any](name string) InputRegistration {
	schema := jsonx.SchemaFrom[T]()

	return InputRegistration{Name: lo.CoalesceOrEmpty(name, jsonx.SchemaID(schema)), Schema: schema}
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
// Connections
// =========

// ConnectionRef builds one Connection whose credential is T
type ConnectionRef[T any] struct {
	// connection is the connection under construction
	connection Connection
	// authManaged reports whether an auth flow obtains the credential, so no form is shown
	authManaged bool
}

// ConnectionOf creates a connection handle named after T's reflected schema
func ConnectionOf[T any]() ConnectionRef[T] {
	return ConnectionRef[T]{connection: Connection{Credential: reflectedInput[T]("")}}
}

// NewConnection creates a connection handle with a literal name
func NewConnection[T any](name string) ConnectionRef[T] {
	return ConnectionRef[T]{connection: Connection{Credential: reflectedInput[T](name)}}
}

// Name declares the user-facing connection name
func (r ConnectionRef[T]) Name(name string) ConnectionRef[T] {
	r.connection.Name = name

	return r
}

// Description declares what the connection does
func (r ConnectionRef[T]) Description(description string) ConnectionRef[T] {
	r.connection.Description = description

	return r
}

// Recommended marks the connection as the recommended method
func (r ConnectionRef[T]) Recommended() ConnectionRef[T] {
	r.connection.Recommended = true

	return r
}

// Meta declares the additional data the user might need to set up the connection
func (r ConnectionRef[T]) Meta(meta map[string]MetaInfo) ConnectionRef[T] {
	r.connection.Meta = maps.Clone(meta)

	return r
}

// Replacing declares that the connection takes over the payloads stored under old
func (r ConnectionRef[T]) Replacing[Old any](old ConnectionRef[Old]) ConnectionRef[T] {
	r.connection.Replaces = append(slices.Clone(r.connection.Replaces), old.connection.Credential.Name)

	return r
}

// Upgraded declares how a stored credential payload is reshaped from the connection it was persisted under into T
func (r ConnectionRef[T]) Upgraded(fn func(context.Context, InstallationRequest, string, json.RawMessage) (T, error)) ConnectionRef[T] {
	r.connection.Credential.Upgrade = upgraded(fn)

	return r
}

// Validated declares a semantic check run on a schema-valid credential payload decoded as T
func (r ConnectionRef[T]) Validated(fn func(context.Context, InstallationRequest, *T) error) ConnectionRef[T] {
	r.connection.Credential.Validate = validated(fn)

	return r
}

// Authenticates declares the auth flow that obtains the credential
func (r ConnectionRef[T]) Authenticates(flow AuthFlow[T]) ConnectionRef[T] {
	registration := flow.registration
	r.connection.Auth = &registration
	r.authManaged = true

	return r
}

// Disconnects declares the teardown description and the optional hook receiving the decoded credential
func (r ConnectionRef[T]) Disconnects(description string, fn func(context.Context, DisconnectRequest[T]) (DisconnectResult, error)) ConnectionRef[T] {
	registration := &DisconnectRegistration{Description: description}

	if fn != nil {
		registration.Disconnect = func(ctx context.Context, input ConnectionInput) (DisconnectResult, error) {
			credential, err := decodeCredential[T](input)
			if err != nil {
				return DisconnectResult{}, err
			}

			return fn(ctx, DisconnectRequest[T]{Integration: input.Integration, Credential: credential, UserInput: input.UserInput})
		}
	}

	r.connection.Disconnect = registration

	return r
}

// Provides declares the client the connection builds from its decoded credential
func (r ConnectionRef[T]) Provides[C any](build func(context.Context, ConnectionRequest[T]) (C, error)) ConnectionRef[T] {
	builder := func(ctx context.Context, input ConnectionInput) (any, error) {
		request, err := connectionRequest[T](input)
		if err != nil {
			return nil, err
		}

		return build(ctx, request)
	}

	r.connection.Clients = lo.Assign(r.connection.Clients, map[string]ClientBuilderFunc{clientName[C](): builder})

	return r
}

// Verified declares the verification run against client C that returns the installation metadata M
func (r ConnectionRef[T]) Verified[C, M any](fn func(context.Context, ConnectionRequest[T], C) (M, error)) ConnectionRef[T] {
	r.connection.Verify = VerifyRegistration{
		ClientRef:    clientName[C](),
		Installation: jsonx.SchemaID(jsonx.SchemaFrom[M]()),
		Handle: func(ctx context.Context, input ConnectionInput) (IntegrationInstallationMetadata, error) {
			request, err := connectionRequest[T](input)
			if err != nil {
				return IntegrationInstallationMetadata{}, err
			}

			client, err := castClient[C](input.Client)
			if err != nil {
				return IntegrationInstallationMetadata{}, err
			}

			metadata, err := fn(ctx, request, client)
			if err != nil {
				return IntegrationInstallationMetadata{}, err
			}

			attributes, err := jsonx.ToRawMessage(metadata)
			if err != nil {
				return IntegrationInstallationMetadata{}, err
			}

			result := IntegrationInstallationMetadata{Attributes: attributes, Display: IntegrationInstallationIdentity{ExternalID: input.Integration.ID}}

			if identifiable, ok := any(metadata).(InstallationIdentifiable); ok {
				result.Display = identifiable.InstallationIdentity()
			}

			return result, nil
		},
	}

	return r
}

// Connection returns the connection with its replaced names sorted, its form derived, and its slices and maps copied
func (r ConnectionRef[T]) Connection() Connection {
	connection := r.connection
	connection.Credential = connection.Credential.Clone()
	connection.Meta = maps.Clone(connection.Meta)
	connection.Clients = maps.Clone(connection.Clients)
	connection.Replaces = nil
	connection.Form = nil

	if len(r.connection.Replaces) > 0 {
		connection.Replaces = sortedUnique(r.connection.Replaces, strings.Compare)
	}

	if !r.authManaged {
		connection.Form = jsonx.CloneRawMessage(connection.Credential.Schema)
	}

	return connection
}

// connectionRequest decodes the credential of the erased input into the typed connection request
func connectionRequest[T any](input ConnectionInput) (ConnectionRequest[T], error) {
	credential, err := decodeCredential[T](input)
	if err != nil {
		return ConnectionRequest[T]{}, err
	}

	return ConnectionRequest[T]{Integration: input.Integration, Credential: credential, TokenManager: input.TokenManager}, nil
}

// decodeCredential decodes the stored credential as T, treating an absent payload as the zero value
func decodeCredential[T any](input ConnectionInput) (T, error) {
	var credential T

	err := jsonx.UnmarshalIfPresent(input.Credential.Data, &credential)

	return credential, err
}

// =========
// User input
// =========

// UserInputRef is a typed handle for one definition's installation-scoped user input layout
type UserInputRef[T any] struct {
	// input is the layout name reflected from T, its schema, and its declared upgrade and validation
	input InputRegistration
}

// UserInputRefOf creates a typed user input layout handle named after T's reflected schema
func UserInputRefOf[T any]() UserInputRef[T] {
	return UserInputRef[T]{input: reflectedInput[T]("")}
}

// Name returns the stable layout name the stored document is keyed by
func (r UserInputRef[T]) Name() string {
	return r.input.Name
}

// Upgraded declares how stored user input is reshaped from the layout it was persisted under into T
func (r UserInputRef[T]) Upgraded(fn func(context.Context, InstallationRequest, string, json.RawMessage) (T, error)) UserInputRef[T] {
	r.input.Upgrade = upgraded(fn)

	return r
}

// Validated declares a semantic check run on schema-valid user input decoded as T
func (r UserInputRef[T]) Validated(fn func(context.Context, InstallationRequest, *T) error) UserInputRef[T] {
	r.input.Validate = validated(fn)

	return r
}

// Registration returns the layout name, reflected schema, declared upgrade, and validation as an input registration
func (r UserInputRef[T]) Registration() *InputRegistration {
	input := r.input.Clone()

	return &input
}

// =========
// Clients
// =========

// clientName returns the client name derived from C, with pointers stripped so *X and X share a name
func clientName[C any]() string {
	clientType := reflect.TypeFor[C]()

	for clientType.Kind() == reflect.Pointer {
		clientType = clientType.Elem()
	}

	return clientType.String()
}

// castClient type-asserts a built client instance to the typed client value
func castClient[C any](client any) (C, error) {
	typed, ok := client.(C)
	if !ok {
		var zero C

		return zero, ErrClientCastFailed
	}

	return typed, nil
}

// =========
// OperationRef
// =========

// OperationRef is a typed handle for one operation identity, parameterized by its config type
type OperationRef[Config any] struct {
	// input is the stable operation name used for persistence and topic derivation, the config type's reflected schema, and its declared upgrade and validation
	input InputRegistration
	// description describes what the operation does
	description string
	// client is the name of the client the operation runs against, empty when the operation has none
	client string
	// healthCheck probes the operation's prerequisites under its client
	healthCheck OperationHandler
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
	return OperationRef[Config]{input: reflectedInput[Config](name), stored: true}
}

// OperationRefOf creates a stored-input operation handle named after the reflected schema of Config; Config embeds
// OperationSettings and its reflected schema is the per-installation input document stored under the operation name
func OperationRefOf[Config OperationInput]() OperationRef[Config] {
	return OperationRef[Config]{input: reflectedInput[Config](""), stored: true}
}

// NewOperationPayload creates a payload operation handle with the given name; Config is the payload a caller
// supplies on each dispatch, nothing is stored per installation, and Upgraded and Validated are ignored
func NewOperationPayload[Config any](name string) OperationRef[Config] {
	return OperationRef[Config]{input: reflectedInput[Config](name)}
}

// OperationPayloadOf creates a payload operation handle named after the reflected schema of Config; Config is the
// payload a caller supplies on each dispatch, nothing is stored per installation, and Upgraded and Validated are ignored
func OperationPayloadOf[Config any]() OperationRef[Config] {
	return OperationRef[Config]{input: reflectedInput[Config]("")}
}

// decodeConfig decodes an operation config payload into Config, treating absent as the zero value
func decodeConfig[Config any](raw json.RawMessage) (Config, error) {
	var cfg Config

	if err := jsonx.UnmarshalIfPresent(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("%w: %w", ErrOperationConfigInvalid, err)
	}

	return cfg, nil
}

// sortedUnique returns the items sorted by compare with duplicates removed
func sortedUnique[T comparable](items []T, compare func(a, b T) int) []T {
	return slices.Compact(slices.SortedFunc(slices.Values(items), compare))
}

// Name returns the stable operation name
func (r OperationRef[Config]) Name() string {
	return r.input.Name
}

// Schema returns a copy of the reflected JSON schema of the config type
func (r OperationRef[Config]) Schema() json.RawMessage {
	return jsonx.CloneRawMessage(r.input.Schema)
}

// Description declares what the operation does
func (r OperationRef[Config]) Description(description string) OperationRef[Config] {
	r.description = description

	return r
}

// HealthCheck binds fn as the probe of the operation's prerequisites, run under the client C
func (r OperationRef[Config]) HealthCheck[C any](fn func(context.Context, OperationRequest, C) error) OperationRef[Config] {
	r.client = clientName[C]()
	r.healthCheck = func(ctx context.Context, request OperationRequest) (json.RawMessage, error) {
		typed, err := castClient[C](request.Client)
		if err != nil {
			return nil, err
		}

		return nil, fn(ctx, request, typed)
	}

	return r
}

// Ingests binds fn as the ingest handler, run against client C with the decoded config
func (r OperationRef[Config]) Ingests[C any](fn func(context.Context, OperationRequest, C, Config) ([]IngestPayloadSet, error)) OperationRef[Config] {
	r.client = clientName[C]()
	r.ingest = func(ctx context.Context, request OperationRequest) ([]IngestPayloadSet, error) {
		typed, err := castClient[C](request.Client)
		if err != nil {
			return nil, err
		}

		cfg, err := decodeConfig[Config](request.Config)
		if err != nil {
			return nil, err
		}

		return fn(ctx, request, typed, cfg)
	}

	return r
}

// Handles binds fn as the handler, run against client C with the decoded config
func (r OperationRef[Config]) Handles[C any](fn func(context.Context, OperationRequest, C, Config) (json.RawMessage, error)) OperationRef[Config] {
	r.client = clientName[C]()
	r.handle = func(ctx context.Context, request OperationRequest) (json.RawMessage, error) {
		typed, err := castClient[C](request.Client)
		if err != nil {
			return nil, err
		}

		cfg, err := decodeConfig[Config](request.Config)
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
	typed := upgraded(fn)

	r.input.Upgrade = func(ctx context.Context, req InstallationRequest, from string, stored json.RawMessage) (json.RawMessage, error) {
		document, err := typed(ctx, req, from, stored)
		if err != nil {
			return nil, err
		}

		settings, err := OperationSettingsFrom(stored)
		if err != nil {
			return nil, err
		}

		patch, err := jsonx.ToRawMap(settings)
		if err != nil {
			return nil, err
		}

		merged, _, err := jsonx.MergeObjectMap(document, patch)

		return merged, err
	}

	return r
}

// Validated declares a semantic check run on a schema-valid stored input document decoded as Config; ignored on a payload operation
func (r OperationRef[Config]) Validated(fn func(context.Context, InstallationRequest, *Config) error) OperationRef[Config] {
	r.input.Validate = validated(fn)

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

// Registration builds the operation registration from the operation's identity, handler, stored input, and declared behavior
func (r OperationRef[Config]) Registration() OperationRegistration {
	registration := OperationRegistration{
		Name:                  r.input.Name,
		Description:           r.description,
		RequiredPermissions:   slices.Clone(r.permissions),
		ClientRef:             r.client,
		CustomerSelectable:    r.customerSelectable,
		Internal:              r.internal,
		RequiresPaymentMethod: r.requiresPaymentMethod,
		Policy:                r.policy,
		RateLimit:             r.rateLimit,
		Ingest:                slices.Clone(r.contracts),
		HealthCheck:           r.healthCheck,
		Handle:                r.handle,
		IngestHandle:          r.ingest,
		DisabledForAll:        r.disabledForAll,
		Input:                 r.input.Clone(),
		Stored:                r.stored,
		Schedule:              r.schedule,
		SkipDefaultLookback:   r.skipDefaultLookback,
	}

	if len(r.replaces) > 0 {
		registration.Replaces = sortedUnique(r.replaces, strings.Compare)
	}

	return registration
}

// =========
// Installations
// =========

// InstallationRef is a typed handle for one definition's installation metadata layout
type InstallationRef[M any] struct {
	// input is the layout name reflected from M, its schema, and its declared upgrade and validation
	input InputRegistration
}

// InstallationOf creates an installation metadata layout handle named after M's reflected schema
func InstallationOf[M any]() InstallationRef[M] {
	return InstallationRef[M]{input: reflectedInput[M]("")}
}

// Upgraded declares how stored installation metadata is reshaped from the layout it was persisted under into M
func (r InstallationRef[M]) Upgraded(fn func(context.Context, InstallationRequest, string, json.RawMessage) (M, error)) InstallationRef[M] {
	r.input.Upgrade = upgraded(fn)

	return r
}

// Validated declares a semantic check run on schema-valid installation metadata decoded as M
func (r InstallationRef[M]) Validated(fn func(context.Context, InstallationRequest, *M) error) InstallationRef[M] {
	r.input.Validate = validated(fn)

	return r
}

// Registration returns the layout and its identity derivation as an installation registration
func (r InstallationRef[M]) Registration() *InstallationRegistration {
	_, identifiable := any(lo.Empty[M]()).(InstallationIdentifiable)

	return &InstallationRegistration{
		InputRegistration: r.input.Clone(),
		Identifiable:      identifiable,
		Identify: func(stored json.RawMessage) (IntegrationInstallationIdentity, error) {
			metadata, err := jsonx.Decode[M](stored)
			if err != nil {
				return IntegrationInstallationIdentity{}, err
			}

			if identifiable, ok := any(metadata).(InstallationIdentifiable); ok {
				return identifiable.InstallationIdentity(), nil
			}

			return IntegrationInstallationIdentity{}, nil
		},
	}
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
		base.Replaces = sortedUnique(r.replaces, strings.Compare)
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

// Registration projects the event name onto base
func (r WebhookEventRef[T]) Registration(base WebhookEventRegistration) WebhookEventRegistration {
	base.Name = r.name

	return base
}
