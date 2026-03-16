# Integrations

`internal/integrations` is the package family that defines, registers, executes, and ingests provider integrations.

It covers:

- provider definition contracts
- installation-scoped credential storage
- installation-scoped client construction and pooling
- operation dispatch and execution
- webhook handling
- payload mapping and ingest

## Package Map

- `definition/`
  Builds and registers definition manifests
- `definitions/`
  Built-in provider implementations
- `registry/`
  In-memory definition, client, operation, and webhook lookup
- `runtime/`
  Main facade used by handlers and workers
- `operations/`
  Dispatch, run tracking, execution, and ingest
- `providerkit/`
  Shared helpers for schemas, OAuth, CEL, envelopes, results, and HTTP clients
- `types/`
  Shared contracts used across the runtime

Related packages:

- `internal/keymaker`
  Auth start/callback orchestration and credential persistence after auth completion
- `internal/keystore`
  Credential persistence and installation-scoped client pooling

## Core Concepts

### Definition

A definition is the unit of integration behavior. It can declare:

- catalog metadata
- installation-scoped `UserInput`
- credentials schema
- auth behavior
- one or more clients
- one or more operations
- mappings for ingest
- webhook contracts

### Installation

An installation is a persisted `Integration` record plus its associated state:

- installation-scoped user input is stored on `integration.Config.ClientConfig`
- credentials are stored separately in keystore-backed hush secrets
- operations execute against one installation

### Client

A client is a provider SDK object or authenticated HTTP wrapper built for one installation.

In runtime terms, clients are installation-scoped today because the keystore pools them by:

- installation ID
- client ref
- credential content
- operation config payload

Definitions usually receive that built client as `request.Client` and then type-assert it with a `FromAny` helper.

### Operation

An operation is a named executable unit such as:

- `health.default`
- `directory.sync`
- `vulnerabilities.collect`

Operations can be inline or queued, and can optionally emit ingest payload sets.

### Ingest

Ingest turns provider payloads into normalized records by:

- finding the correct mapping for schema and variant
- applying installation-level and mapping-level CEL filters
- applying a CEL map expression
- upserting the normalized record

## End-To-End Flow

```mermaid
sequenceDiagram
    actor User
    participant Provider as External Provider
    participant API as HTTP Handlers
    participant DB as Ent
    participant Registry
    participant Runtime
    participant Keymaker
    participant Keystore
    participant Gala
    participant Def as Definition
    participant Ops as Ingest Runtime

    rect rgb(245, 245, 245)
        Note over User,Ops: Installation activation
        User->>API: Start install / configure integration
        API->>Registry: Resolve definition
        API->>DB: Create or load installation
        API->>DB: Persist UserInput to Config.ClientConfig

        alt Direct credential install
            User->>API: Submit credential material
            API->>Runtime: Optional validation operation\n(often health.default)
            Runtime->>Registry: Resolve operation + client ref
            Runtime->>Keystore: BuildClient(installation, clientRef, credential, config=nil)
            alt Pooled client hit
                Keystore-->>Runtime: Cached client
            else Pooled client miss
                Keystore->>Def: Client.Build(ClientBuildRequest)
                Def-->>Keystore: Initialized client
            end
            Runtime->>Def: Handle(OperationRequest)
            Def-->>Runtime: Validation result
            API->>Keystore: SaveCredential(installation, credential)
        else OAuth install
            API->>Runtime: BeginAuth(definitionID, installationID)
            Runtime->>Keymaker: BeginAuth(...)
            Keymaker->>Def: Auth.Start / generic OAuth start
            Def-->>Keymaker: Auth URL + callback state
            API-->>User: Redirect to provider
            User->>Provider: Complete consent
            Provider-->>API: OAuth callback
            API->>Runtime: CompleteAuth(state, callback input)
            Runtime->>Keymaker: CompleteAuth(...)
            Keymaker->>Def: Auth.Complete / generic OAuth complete
            Def-->>Keymaker: CredentialSet
            Keymaker->>Keystore: SaveInstallationCredential(installationID, credential)
        end

        Keystore->>Keystore: Persist credential
        Keystore->>Keystore: Invalidate installation client cache entries
        API->>DB: Mark installation connected
        API->>Runtime: SyncWebhooks(installation)
    end

    rect rgb(245, 245, 245)
        Note over User,Ops: Operation execution
        User->>API: Run operation
        alt Inline execution
            API->>Runtime: LoadCredential + ExecuteOperation
        else Queued execution
            API->>Runtime: Dispatch(request)
            Runtime->>Gala: Emit operation envelope
            Gala-->>Runtime: HandleOperation(envelope)
            Runtime->>DB: Load installation
            Runtime->>Keystore: LoadCredential(installation)
        end

        Runtime->>Registry: Resolve client ref
        Runtime->>Keystore: BuildClient(...)
        alt Cached client hit
            Keystore-->>Runtime: Cached client
        else Cache miss
            Keystore->>Def: Client.Build(ClientBuildRequest)
            Def-->>Keystore: Initialized client
        end
        Runtime->>Def: Handle(OperationRequest{Integration, Credential, Client, Config})
        Def-->>Runtime: Operation response
    end

    rect rgb(245, 245, 245)
        Note over User,Ops: Ingest
        alt Operation emits ingest payloads
            Runtime->>Ops: ProcessIngest(operation, response)
            Ops->>Registry: Resolve mapping
            Ops->>Ops: Read installation filterExpr from ClientConfig
            Ops->>Ops: Apply CEL filter + map expressions
            Ops->>DB: Upsert normalized records
        end
    end

    rect rgb(245, 245, 245)
        Note over User,Ops: Webhooks
        Provider->>API: Deliver webhook event
        API->>Runtime: DispatchWebhookEvent
        Runtime->>Gala: Emit normalized webhook envelope
        Gala-->>Runtime: HandleWebhookEvent
        Runtime->>Def: Webhook Handle(request)
        alt Webhook emits ingest payloads
            Def->>Ops: Ingest(payloadSets)
            Ops->>DB: Persist normalized records
        else Webhook dispatches follow-up operation
            Def->>Runtime: DispatchOperation(name, config)
            Runtime->>Gala: Queue operation
        end
    end
```

## Definition Authoring

Each built-in provider lives under `internal/integrations/definitions/<provider>/`.

The default goal is:

- keep provider-specific behavior self-contained in one package
- keep shared runtime concerns in `runtime/`, `operations/`, and `providerkit/`
- make the package shape predictable across providers

### Recommended Default Files

Create these by default for a new provider package:

```text
definitions/<provider>/
  builder.go
  client.go
  operations.go
  errors.go
```

Add these when needed:

```text
  config.go        # operator config, shared UserInput, shared config types
  auth.go          # OAuth or custom auth helpers
  mappings.go      # default ingest mappings
  webhook.go       # inbound webhook handlers and contracts
  schema.go        # JSONSchemaExtend hooks
  *_test.go        # schema, client, auth, operation, or ingest tests
```

### Recommended Default Types

Start with these types by default:

- `UserInput`
  Installation-scoped user configuration
- `CredentialSchema`
  Persisted credential/provider data shape for direct-credential providers
- `Client`
  The provider client builder and `FromAny` caster
- one struct per operation
  Example: `HealthCheck`, `FindingsCollect`, `DirectorySync`
- one config struct per configurable operation
  Example: `FindingsConfig`, `DirectorySyncConfig`

Add these only when the provider needs them:

- `Config`
  Operator-owned config supplied at runtime startup
- auth state / callback payload structs
- webhook payload wrapper structs

### Definition Conventions

- `builder.go` should be the registration entrypoint
- `UserInput` should live close to the builder unless it is reused heavily
- operation handlers should be thin adapters around typed `Run(...)` methods
- `Client.Build(...)` should build the provider client from `ClientBuildRequest`
- `Client.FromAny(...)` should only type-assert the already-built pooled client
- mapping and ingest logic should prefer shared `providerkit` helpers

One subtle but important rule:

- if a client builder starts depending on more installation state than `credential` and `config`, make sure the keystore cache key and invalidation model still match that behavior

## Copy/Paste Scaffold

The following templates are intentionally minimal. They are meant to be copied into a new definition package and then find-and-replaced.

Replace these placeholders first:

- `yourprovider`
- `Your Provider`
- `YOUR_PROVIDER`
- `YourClient`
- `CollectDefaultOperation`

### `builder.go`

```go
package yourprovider

import (
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID            = types.NewDefinitionRef("def_REPLACE_ME")
	ProviderClient          = types.NewClientRef[any]()
	HealthDefaultOperation  = types.NewOperationRef[struct{}]("health.default")
	CollectDefaultOperation = types.NewOperationRef[struct{}]("collect.default")
)

const Slug = "yourprovider"

type UserInput struct {
	Label      string `json:"label,omitempty" jsonschema:"title=Installation Label"`
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
}

type CredentialSchema struct {
	APIToken string `json:"apiToken,omitempty" jsonschema:"title=API Token"`
}

func Builder() definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Version:     "v1",
				DisplayName: "Your Provider",
				Category:    "REPLACE_ME",
				Active:      true,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			Credentials: &types.CredentialRegistration{
				Schema: providerkit.SchemaFrom[CredentialSchema](),
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         ProviderClient.ID(),
					Description: "Primary Your Provider client",
					Build:       Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:      HealthDefaultOperation.Name(),
					Topic:     HealthDefaultOperation.Topic(Slug),
					ClientRef: ProviderClient.ID(),
					Policy:    types.ExecutionPolicy{Idempotent: true},
					Handle:    HealthCheck{}.Handle(Client{}),
				},
				{
					Name:      CollectDefaultOperation.Name(),
					Topic:     CollectDefaultOperation.Topic(Slug),
					ClientRef: ProviderClient.ID(),
					Policy:    types.ExecutionPolicy{Idempotent: true},
					Handle:    Collect{}.Handle(Client{}),
				},
			},
		}, nil
	})
}
```

### `client.go`

```go
package yourprovider

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
)

type YourClient struct{}

type Client struct{}

func (Client) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	_ = ctx
	_ = req

	if len(req.Credential.ProviderData) == 0 {
		return nil, ErrCredentialRequired
	}

	var client YourClient

	return &client, nil
}

func (Client) FromAny(value any) (*YourClient, error) {
	client, ok := value.(*YourClient)
	if !ok {
		return nil, ErrClientType
	}

	return client, nil
}
```

### `operations.go`

```go
package yourprovider

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

type HealthCheck struct{}

type Collect struct{}

func (h HealthCheck) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return h.Run(ctx, request.Credential, c)
	}
}

func (HealthCheck) Run(ctx context.Context, credential types.CredentialSet, client *YourClient) (json.RawMessage, error) {
	_ = ctx
	_ = credential
	_ = client

	return providerkit.EncodeResult(map[string]any{"ok": true}, ErrResultEncode)
}

func (c Collect) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		typedClient, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return c.Run(ctx, request.Credential, typedClient)
	}
}

func (Collect) Run(ctx context.Context, credential types.CredentialSet, client *YourClient) (json.RawMessage, error) {
	_ = ctx
	_ = credential
	_ = client

	return providerkit.EncodeResult(map[string]any{"items": []any{}}, ErrResultEncode)
}
```

### `errors.go`

```go
package yourprovider

import "errors"

var (
	ErrClientType         = errors.New("yourprovider: client type invalid")
	ErrCredentialRequired = errors.New("yourprovider: credential required")
	ErrResultEncode       = errors.New("yourprovider: result encode failed")
)
```

### Optional `mappings.go`

Add this when the provider emits ingest payloads:

```go
package yourprovider

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/types"
)

func mappings() []types.MappingRegistration {
	return []types.MappingRegistration{
		{
			Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    "{}",
			},
		},
	}
}
```

## Practical Defaults

If you are starting a new provider and want the smallest sensible shape:

1. Create `builder.go`, `client.go`, `operations.go`, and `errors.go`
2. Add `UserInput` with at least `label` and `filterExpr`
3. Add a `health.default` operation first
4. Register exactly one client unless there is a real need for more
5. Add `mappings.go` only if the provider emits ingest payloads
6. Add `auth.go` only if the provider needs OAuth or another install flow
7. Add `webhook.go` only if the provider receives inbound events

That shape keeps most providers consistent while leaving room for more complex auth, ingest, or webhook behavior when it is actually needed
