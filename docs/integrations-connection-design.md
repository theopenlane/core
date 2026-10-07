# Integration connections: collapsing credentials, connections, and clients

Status: agreed, checked claim by claim against core at `f7637bf47` and the console (`openlane-ui` at `cbdad4b57`), then reviewed adversarially; line numbers cite that core revision. Verification lives on the connection, operations receive only their client, disconnect is typed per connection, installation metadata is a conformed kind declared once per definition with a stored layout name, and the registry stamps topics so no ref takes a definition id.

## Outcome

A definition declares:

1. how a user supplies a credential: a form, or an auth flow that produces it
2. which clients that credential builds
3. how the credential is verified, and the installation identity that verification yields
4. what the operations use those clients for

Today those facts are spread over `CredentialRegistration`, `ConnectionRegistration`, `ClientRegistration`, `ClientRef.Using`, `AuthRegistration.CredentialRef`, `DisconnectRegistration.CredentialRef`, credential bindings, a definition-level health check, and a separate installation metadata resolver. The registry then validates that the copies agree, and every client builder, resolver, health check, and operation handler recovers its typed credential from untyped bindings.

## What the code shows

| Fact | Evidence |
|---|---|
| A connection always uses exactly its selecting slot | no definition sets `ConnectionRegistration.CredentialRefs`; finalize defaults it to `[CredentialRef]` (`registry/finalize.go:67-73`) |
| Multiple credentials mean alternative connection methods, never one connection with two slots | awssecurityhub, gcpscc, slack, zitadel: two credentials, two connections, one slot each; 15 definitions have one connection; scim and system have none |
| One client builder serves every method today, by probing which slot holds a credential | `gcpscc/client.go:58-98`, `awssecurityhub/aws.go:72-108`, `zitadel/client.go:70-111`, `slack/client.go:109-131` |
| The metadata resolver probes the same way, and branches per method | `awssecurityhub/installation.go:14-78` derives the account from the role ARN under assume-role and calls STS under static credentials |
| The slot name is written three times per connection, then validated for agreement | credential `Ref`, connection `CredentialRef`, disconnect `CredentialRef`; `validateConnections` (`registry/registry.go:643-664`) |
| `ClientRegistration.CredentialRefs` / `ClientRef.Using` only restate slots and are defaulted to all slots | `registry/finalize.go:58-64`; `.Using` appears only in test utilities |
| Every production client is named after its Go type | 22 `ClientRefOf[C]()` declarations across the definitions, no duplicate types within a definition; `ClientRefOf` strips pointers (`types/ref.go:292-299`), so `*X` and `X` share a name |
| Client builders read the credential, and three read more | `email/client.go:59-60` reads `req.Integration.UserInput.Data`; `slack/client.go:25` reads `req.Integration.Metadata`; `gcpscc/federation.go:31,54` reads `req.TokenManager` and `Integration.OwnerID`; none reads `req.Config`; `awssecurityhub` reads its builder's operator config `b.cfg` |
| Seven definitions decode the credential again inside operation handlers, for non-secret scope the client could carry | cloudflare (`operation_health.go:25`, `operation_findings.go:24`, `operation_asset_sync.go:20`, `operation_directory_sync.go:78`), gcpscc (`operation_findings.go:32`, `operation_health.go:25`), oci (`operation_findings.go:20`, `operation_health.go:25`), keycloak (`operation_directory_sync.go:15`, `operation_health.go:27`), zitadel (`operation_health.go:26`), oidclocal (`operation_claims.go:14`, `operation_health.go:13`), awssecurityhub (`operation_findings.go:33`); cloudflare already carries `Config.AccountID` on its client (`cloudflare/client.go:28-36`); googleworkspace's handler reads the stored `Attributes` instead (`googleworkspace/operation_directory_sync.go:25-32`) |
| No caller passes credentials into `ExecuteOperation` | `handlers/integration_operations.go:114`, `graphapi/internalpolicyextended.resolvers.go:71` both pass `nil` |
| No hook reads `DisconnectRequest.Connection` or `InstallationRequest.Connection` | set at `runtime/credentials.go:83`, `:209`, `:230`; no reader |
| The metadata resolver runs right after the connection health check, whose own result is discarded | `runtime/health.go:369-393`; `RefreshInstallationMetadata` follows the check at `credentials.go:141`→`:148` and `health.go:264`→`:291` |
| A resolver failure on the health path is logged, not classified | `health.go:291-293` logs the refresh error and the assessment still reports healthy; on the reconnect path it is an error (`credentials.go:151-154`) |
| The pre-compare refresh handles a provider-reported id change | `TestReconnectRefreshesChangedInstanceID` (`graphapi/eventstest/integrations_instanceid_backfill_test.go:217-237`) reconnects after the provider reports a new id and expects the stored id to follow |
| The health check result struct is per definition and never persisted | `HealthCheck` types in every `operation_health.go`, consumed only by `providerkit.EncodeResult` |
| `Registration(definitionID)` on operations and webhook events exists only to build the gala topic | `OperationRef.Registration` sets `Topic: definition.OperationTopic(name)` (`types/ref.go:641`); `WebhookEventRef.Registration` likewise (`ref.go:773`); the registry already receives `def.ID` when it indexes topics (`registry/registry.go:242-254`); about 59 production call sites plus test utilities and tests |
| Installation metadata is the one stored kind without a layout name or an upgrade hook | `IntegrationUserInput{Layout, Data}` vs `IntegrationInstallationMetadata{Attributes, Display}` (`common/openapi/integration_models.go:67-86`); `upgradeInstallation` conforms three kinds and never writes `installation_metadata` (`runtime/upgrade.go:63-151`) |
| Email has a connection and no installation metadata | `email/builder.go` declares no `Installation:`; its snapshot has no `installation` key; it takes the self-identity path (`credentials.go:217-220`) |
| Reconnect through a different method is rejected today | `resolveConnectionForCredential` returns `ErrCredentialNotDeclared` when the persisted connection does not declare the submitted slot (`credentials.go:389-409`; `credentials_test.go:165-195`) |
| Operation client resolution never reads provider state today | `resolveOperationClient` loads the client's slots and lets the builder probe (`execution.go:615-652`); a multi-connection installation with no persisted `CredentialRef` resolves under the new design through `resolvePersistedConnection`, which returns `ErrConnectionRequired` (`credentials.go:378-394`) |

## Design

### One type per connection method

```go
// Connection is one way a user connects a definition
type Connection struct {
	Name        string
	Description string
	Recommended bool
	Meta        map[string]MetaInfo
	// Credential is the stored credential layout; Credential.Name is the persisted slot and connection name
	Credential InputRegistration
	// Form is the user-facing credential form schema, empty when Auth obtains the credential
	Form       json.RawMessage
	Replaces   []string
	Auth       *AuthRegistration
	Disconnect *DisconnectRegistration
	// Clients builds each client this method provides, keyed by client name
	Clients map[string]ClientBuilderFunc
	// Verify probes the connection under the named client and derives the installation metadata
	Verify VerifyRegistration
}

// VerifyRegistration is the connection's verification
type VerifyRegistration struct {
	// ClientRef is the client the verification runs against; it must be one of Clients
	ClientRef string
	// Installation is the layout name of the metadata type the verification returns; it must equal Definition.Installation.Name
	Installation string
	// Handle runs the verification and returns the marshalled metadata
	Handle func(context.Context, ConnectionInput) (IntegrationInstallationMetadata, error)
}

// ConnectionInput is the erased input every connection closure receives; Client is set for verification, UserInput for disconnect
type ConnectionInput struct {
	Integration  *generated.Integration
	Credential   CredentialSet
	TokenManager *tokens.TokenManager
	Client       any
	UserInput    json.RawMessage
}

// ClientBuilderFunc builds one client from the erased input
type ClientBuilderFunc func(context.Context, ConnectionInput) (any, error)

// InstallationRegistration is the definition's installation metadata layout, declared once
type InstallationRegistration struct {
	// InputRegistration is the layout name, schema, upgrade, and validation of M
	InputRegistration
	// Identify recomputes the display identity from a stored metadata document, erased from M
	Identify func(json.RawMessage) (IntegrationInstallationIdentity, error)
}
```

`Definition.Connections []Connector` replaces `CredentialRegistrations`, `Connections`, `Clients`, and `HealthCheck`. `Definition.Installation` stays, declared through `InstallationOf[M]`, and is required whenever the definition has connections.

### Typed builder

```go
// ConnectionRef builds one Connection whose credential is T
type ConnectionRef[T any] struct{ connection Connection }

func ConnectionOf[T any]() ConnectionRef[T]             // named after T's schema id, as CredentialRefOf names slots today
func NewConnection[T any](name string) ConnectionRef[T] // literal name, as NewCredentialRef does today

func (r ConnectionRef[T]) Provides[C any](build func(context.Context, ConnectionRequest[T]) (C, error)) ConnectionRef[T]
func (r ConnectionRef[T]) Verified[C, M any](fn func(context.Context, ConnectionRequest[T], C) (M, error)) ConnectionRef[T]
func (r ConnectionRef[T]) Authenticates(flow AuthFlow[T]) ConnectionRef[T]
func NewAuthFlow[T any](start AuthStartFunc, complete AuthCompleteFunc) AuthFlow[T] // for hand-built flows: githubapp, azureentraid, the test fixture
func (r ConnectionRef[T]) Disconnects(description string, fn func(context.Context, DisconnectRequest[T]) (DisconnectResult, error)) ConnectionRef[T]
func (r ConnectionRef[T]) Name(string), Description(string), Recommended(), Meta(map[string]MetaInfo) ConnectionRef[T]
func (r ConnectionRef[T]) Replacing[Old any](ConnectionRef[Old]), Upgraded(...), Validated(...) ConnectionRef[T]

// InstallationRef declares the definition's installation metadata type M, its upgrade and validation hooks, once per definition
type InstallationRef[M any] struct{ registration InstallationRegistration }

func InstallationOf[M any]() InstallationRef[M] // named after M's schema id, like UserInputRefOf
func (r InstallationRef[M]) Upgraded(fn func(context.Context, InstallationRequest, string, json.RawMessage) (M, error)) InstallationRef[M]
func (r InstallationRef[M]) Validated(fn func(context.Context, InstallationRequest, *M) error) InstallationRef[M]
func (r InstallationRef[M]) Registration() *InstallationRegistration

// ConnectionRequest is what a client builder and the verification receive
type ConnectionRequest[T any] struct {
	Integration  *generated.Integration
	Credential   T
	TokenManager *tokens.TokenManager
}

// DisconnectRequest is what a connection's disconnect hook receives
type DisconnectRequest[T any] struct {
	Integration *generated.Integration
	Credential  T
	UserInput   json.RawMessage
}
```

- `ConnectionRef[T]` satisfies `Connector`, one unexported method returning `Connection`, so a definition lists refs directly and nothing calls `Registration()`.
- `OperationRef.Registration()` and `WebhookEventRef.Registration(base)` lose their `DefinitionRef` parameter. `registry.Register` stamps `Topic` on every operation and webhook event from `def.ID` through `NewDefinitionRef(def.ID).OperationTopic(name)` and `.WebhookEventTopic(name)` before indexing. The only definition-id reference left in a builder is `ID: definitionID.ID()`.
- `Provides` erases `build` into `ClientBuilderFunc` by decoding the stored credential into `T` once.
- `Verified` registers `C`'s name as `Verify.ClientRef`, records `M`'s schema id as `Verify.Installation`, and erases `fn` into `Verify.Handle`, which decodes `T`, casts the built client to `C`, marshals `M` into `Attributes`, and fills `Display` from `M.InstallationIdentity()` when `M` implements `InstallationIdentifiable`. `InstallationOf[M]` on the definition reflects the same `M` into the layout, carries the hooks, and erases `M.InstallationIdentity()` into `Identify` for the upgrade path. The registry refuses a connection without `Verified`, a connection whose `Verify.ClientRef` is not in its `Clients`, a definition with connections and no `Installation`, and a connection whose `Verify.Installation` differs from `Definition.Installation.Name`. The hooks are therefore declared once, which the registry can check; declared per connection they would be closures it cannot compare.
- **Identity rule, unchanged from today.** When `M` does not implement `InstallationIdentifiable`, `Verified` fills `Display.ExternalID` with the installation's own id, as the self-identity path sets it today for a definition with no installation registration (`credentials.go:217-220`); email is the one definition with a connection on that path, and `InstallationRegistration.Identifiable` records the distinction for the upgrade path, which uses the installation id instead of `Identify` when it is false. When `M` implements it and returns an empty `ExternalID`, the runtime fails the verification with `ErrInstallationInstanceIDRequired` (`credentials.go:231-233`); gcpscc's identity can be empty (`gcpscc/types.go:115-123`) and ingest requires a non-empty id.
- `Disconnects` erases `fn` the same way, decoding the stored credential into `T` for the hook. A connection with only a description (every definition except slack and githubapp) passes a nil hook.
- `AuthFlow[T]` is `AuthRegistration` tagged with `T`. `auth.OAuthRegistration` returns it, which keeps the compile-time link between the OAuth material mapper and the connection's credential type. Today that link comes from `OAuthRegistrationOptions.CredentialRef CredentialRef[T]` (`auth/oauth_registration.go:14`).

```go
Installation: types.InstallationOf[InstallationMetadata]().Registration(),
Connections: []types.Connector{
	types.ConnectionOf[workloadIdentityCred]().Name("GCP Workload Identity Federation").Provides(sccClientFromWorkload).Verified(verifyWorkload),
	types.ConnectionOf[serviceAccountCred]().Name("GCP Service Account").Provides(sccClientFromServiceAccount).Verified(verifyServiceAccount),
},
```

Both `verify*` functions return `InstallationMetadata`; each reads its own `req.Credential.CollectionScope` and shares `resolveParents`. For the fifteen single-connection definitions the builder line is `ConnectionOf[cred]().Provides(build).Verified(check)`. Where the verification needs a client the operations do not use (azureentraid verifies through `azcore.TokenCredential` and operates through the Graph client; oci verifies through the Identity client and operates through Cloud Guard), the connection provides both and the registry's `Verify.ClientRef` check covers the second one. awssecurityhub's static-credential verification calls STS, which has no registered client today; its `Client` carries the `awssdk.Config` every service client is built from, so the verification builds the STS client from that.

### Operations receive only the client

`OperationRequest` loses `Credentials`. A handler receives `C`, decoded config, and the request; nothing in a handler decodes a credential. The seven definitions that decode credentials in handlers today move the non-secret scope those handlers read onto the client type the connection builds:

- cloudflare: already `CloudflareClient.Config.AccountID`; the four handlers read it instead of re-decoding
- gcpscc: `CollectionScope` on a wrapper around `*cloudscc.Client`
- oci: tenancy, compartment, and region on a wrapper around the Cloud Guard and Identity clients
- keycloak: realm and client credentials on a wrapper around `*gocloak.GoCloak`, since `LoginClient` needs them per call
- zitadel: domain on a wrapper around `*client.Client`
- oidclocal: the stored claims on a client type, since the definition has no client today
- awssecurityhub: account scope, account ids, and linked regions on the client, read by `buildFilters`. Today `buildFilters` requires the assume-role credential (`operation_findings.go:91-96`), so findings sync fails on the static-credential connection; with scope on the client the static connection collects with `AccountScope` all, which is a behaviour fix.

That is what azureentraid, slack, githubapp, okta, tailscale, googledrive, googleworkspace, onedrive, microsoftteams, authentik, azuresecuritycenter, and email already do. Keycloak's `ClientSecret` moves onto its client, which is acceptable: the pool is in-memory with a TTL and keyed by a credential digest, and cloudflare and slack already hold tokens on their clients. With this, `T` is decoded only inside the closures `Provides`, `Verified`, and `Disconnects` erase; no erased credential value crosses into definition code. `OperationRequest.Integration` stays: googleworkspace reads the stored `Attributes` from it, so conformed `Attributes` reach handlers. The upgrade path keeps `keystore.LoadAllCredentials`, because it may find retired slots.

### Names instead of typed identities

- **Connections and slots** are identified by `Connection.Credential.Name`, a `string`, like every other stored kind (`InputRegistration.Name`, `OperationRegistration.Replaces`).
- **Clients** are identified by the name derived from their Go type (the `ClientRefOf` naming).
  - `Provides[C]` registers the builder under that name; `Verified[C, M]` records it as `Verify.ClientRef`.
  - `OperationRef.Handles(fn)`, `Ingests(fn)`, and `HealthCheck(fn)` infer `C` from `fn` and set `OperationRegistration.ClientRef` (a `string`), so operations stop taking a client argument. An operation's probe runs under its own client, as `runOperationProbe` does today (`runtime/health.go:862-877`).
  - Registry validation checks that every connection provides every client an operation uses, and that each connection provides its own `Verify.ClientRef`. The first is today's behaviour, where one builder serves every method. A definition with no connections keeps today's rule: an operation naming a client is an error (`ErrClientNotFound`) unless the definition declares `RuntimeIntegration`, whose single runtime client serves every operation regardless of name (`execution.go:619-629`).

### Installation metadata comes from the connection's verification, not a resolver

- **Producer.** `Verified[C, M]` on each connection. `M` is the definition's installation metadata type, declared once as `Definition.Installation` through `InstallationOf[M]`; the registry requires every connection's `M` to match it.
- **What the connection verification is.** The identity call: the request that returns the account, tenant, workspace, or realm the credential is scoped to. That is what the instance-match guard and `installation_metadata` need before a credential can be saved. Scope checks belong to operation probes (`OperationRef.HealthCheck`), which run under the operation's own client and degrade only that operation; azureentraid's directory probe is the model (`azureentraid/builder.go:65`). Today's definition checks that are really scope probes, such as googleworkspace's `Users.List` and okta's `GetUser("me")`, move to the operation that needs the scope or are dropped where the identity call already proves access.
- **Result, not side effect.** `runConnectionHealthCheck` resolves the active connection, loads its one credential, builds `Verify.ClientRef` through the same client cache, runs `Verify.Handle`, and returns the metadata. It does not persist, because in `ReconcileCredential` the check runs before the instance-match guard and before the credential is saved (`credentials.go:141-172`).
  - `ReconcileCredential`: when a credential is already stored for the connection, run the verification under it first and persist the result, exactly what `RefreshInstallationMetadata` does today before the compare (`credentials.go:148-150`); this is what lets a provider-reported id change reconnect instead of failing the guard (`TestReconnectRefreshesChangedInstanceID`). Then run the verification under the submitted credential, compare its `Display.ExternalID` with the stored one, save the credential, persist the connection name, and `saveInstallationMetadata` with the result.
  - `RunHealthAssessment` and `updateInstallationInput` recovery: run the verification under the stored credential and `saveInstallationMetadata` directly, then clear unhealthy and assess operations. No guard is needed: the credential is the stored one. This replaces `RefreshInstallationMetadata`.
- **One persistence point.** `saveInstallationMetadata` stamps both `Layout` from `def.Installation.Name` and `Display.CredentialRef` from the active connection's name on every write. Today only the refresh path stamps `CredentialRef` (`credentials.go:308`), and the GraphQL `Credentials` resolver falls back to the first registration when it is blank (`integrationextended.resolvers.go:64-67`), which would prefill the wrong form for a two-method definition.
- **Persistence shape.** Unchanged apart from `Layout`: `Attributes` is the marshalled `M`, `Display` comes from `M.InstallationIdentity()` plus the stamps above; the upgrade path recomputes it through `Installation.Identify`.
- **Error classification.** The connection verification keeps the connection check's semantics: a failure on the health path marks the installation errored (`health.go:276-285`), on reconnect it fails the request (`credentials.go:141-144`). What changes is that the identity call is the verification; today its failure after a passing check is only logged on the health path (`health.go:291-293`) because it was a second call. On the reconnect path it already fails the request (`credentials.go:151-154`), and reconnect is the only way a credential is accepted, so the identity call is already a hard requirement at connect time; a health assessment now holds the installation to the same bar. The check bodies otherwise change only to return `M` instead of an untyped payload: the per-definition `HealthCheck` result structs are deleted, since nothing persisted them; resolvers that return `ok=false` today (azureentraid with an empty tenant, oidclocal, keycloak, googleworkspace) return an explicit error, which is what `ok=false` already becomes at `credentials.go:231-233`; awssecurityhub's two resolver branches become the two connections' verifications.
- **Reconnect guard.** Unchanged in behaviour: refresh under the stored credential, verify the submitted one, compare, as above. A change in how `InstallationIdentity()` derives the id between versions is additionally covered by the upgrade path below, which recomputes `Display` on upgrade.
- **Method switching.** Unchanged: today a reconnect through a different method is rejected (`ErrCredentialNotDeclared`, `credentials.go:403-404`). `ReconcileCredential` keeps that rule as `ErrConnectionMismatch`, raised when the persisted connection name differs from the submitted one. Allowing a switch, with the retired slot deleted, is a product change outside this design.
- **Rows without a persisted connection.** The upgrade backfills `provider_state` for an installation whose state names no connection but whose Hush rows hold exactly one slot that maps to a connection, by name or through `Replaces`. A multi-connection installation with no state and no such slot keeps failing with `ErrConnectionRequired`, as it does today on every path except operation client resolution. Whether any such row exists in production is unverified.
- **Snapshots.** `surface.Installation` reads `M`'s schema from `Verify.Installation`: the same bytes as today's `InstallationRef[T]` schema for every definition except slack, whose `DefaultChannel` field is deleted, and email, which gains an installation schema. Both versions move; email's upgrade reshapes nothing.
- **`InstallationIdentifiable` stays.** It is the per-definition mapping from metadata to identity, and it is not removable without a cost:
  - Several mappings derive values rather than copy fields: `gcpscc/types.go:115-123` formats organization and project parents, and `keycloak/types.go:52-60` falls back between names.
  - Replacing it with uniform embedded fields renames keys in `Attributes`. The GraphQL `Credentials` resolver prefills the reconnect form from those keys (`graphapi/integrationextended.resolvers.go:77-88`, for example azureentraid `tenantId`).
  - The rename also adds properties to every metadata schema, which changes the definition snapshots and forces an upgrade of every installation of those definitions.

### Installation metadata is a conformed kind on the upgrade path

Today `upgradeInstallation` conforms credentials, user input, and operation input (`runtime/upgrade.go:63-100`) and never touches `installation_metadata` (`:128-151`); `InstallationRegistration` has no upgrade or validation hook (`types/installation_registration.go:28-33`). Installation metadata becomes the fourth kind, mirroring user input:

| | User input today | Installation metadata |
|---|---|---|
| Stored value | `IntegrationUserInput{Layout, Data}` | `IntegrationInstallationMetadata{Layout, Attributes, Display}` |
| Layout name | reflected type name of `UserInput` (`UserInputRefOf[T]`) | reflected type name of `M` (`Verified[C, M]`) |
| One layout per definition | `Definition.UserInput` | `Definition.Installation`, which every connection's `Verified[C, M]` must match |
| Written by | configure and auth-start handlers through `EnsureInstallation` | the verification result through `saveInstallationMetadata`, and the upgrade |
| Stamped on write | `mergeInput` sets `Layout = def.UserInput.Name` (`integration.go:283-286`) | `saveInstallationMetadata` sets `Layout = def.Installation.Name` |
| Derived mirror onto the record | `mirrorUserInput` copies name and primary directory (`integration.go:297-305`) | `Display` recomputed through `Installation.Identify` and merged into `metadata` |
| Upgrade | `userInputKind` resolves the stored `Layout` against the current name; `replaced` when they differ (`upgrade.go:354-367`) | `installationKind` does the same against `def.Installation.Name` |
| Hook and validation | `Upgraded` / `Validated` on `UserInputRefOf[T]`, run by `conformStored` | `Upgraded` / `Validated` on `InstallationOf[M]`, run by `conformStored` |
| Surface | `SurfaceSchema{Schema, Upgrade}` | same `SurfaceSchema{Schema, Upgrade}` |
| Empty document on an old row | `conformUserInput` skips a row with no layout and no data, else stores it under `""` and conforms onto the current name (`upgrade.go:317-339`) | same rule: a row with no `Layout` and no `Attributes` is skipped, so a nil document is not rewritten as `{}` with a blank `Display` |
| Definition without the kind | `userInputKind` returns ok false when `def.UserInput` is nil | `installationKind` returns ok false when the definition has no connections (scim, system), whose rows keep their self identity |

- `IntegrationInstallationMetadata` gains `Layout string` with `omitempty`. It is a JSON column, so no migration; the ent field is skipped in GraphQL (`ent/schema/integration.go:94-99`), REST exposes only `Attributes` (`handlers/integration_config.go:70`), and the console reads identity from `integration.metadata`, not this column.
- **Stamping is lazy.** A row is stamped on its next verification or upgrade. Only slack and email get a version bump from this design, so most rows keep an empty `Layout` until a health assessment, a reconnect, or a later version bump of their definition. An unstamped row is conformed exactly as an unstamped user input row is (`upgrade.go:317-339`, `:354-367`): it resolves to the current layout with `from` equal to the empty string, so a later rename of `M` hands those rows an empty `from`; the hook's default branch must therefore recognise the pre-rename shape from the document, as the credential hooks in `testutils/integrations/versions.go` do in their default case. That is the only content-based case, and it ends once every row has been stamped.
- `upgradeInstallation` adds an `installationKind` whose document is `Attributes` under the stored `Layout`, conforms it with `conformDocuments`, and writes `Layout`, `Attributes`, and a `Display` recomputed through `Installation.Identify` plus the `CredentialRef` from provider state, in the same transaction. No provider call: the document is reshaped from what is stored, as the other three kinds are. A conformance failure fails the upgrade and marks the installation unhealthy, as it does for the other three kinds (`upgrade.go:46-53`); the kinds are not treated differently.
- **No lost update.** The other kinds are conformed from the in-memory record before the transaction and overwritten inside it (`upgrade.go:74-100`, `:143-150`). Metadata is the one kind another path writes concurrently (a reconnect or health save), so the transaction re-reads `installation_metadata` after it claims the version and conforms that copy. Conformance makes no provider call, so the re-read costs one query.
- `Surface.Installation.Upgrade` reports whether the hook is declared, like `SurfaceCredential` and `SurfaceOperation`.

A change to how `InstallationIdentity()` derives `ExternalID` is therefore shipped as a metadata type change with an upgrade hook, which bumps the version and rewrites `Display` on upgrade, so reconnect after upgrade compares like against like. Renaming `M` works like renaming a credential type: `Replacing` on the ref, a hook keyed on `from`, the retired layout name dropped once the replacement is stored.

## Storage and API

One additive JSON key; no other persisted shape or API contract changes.

| Item | Status | Evidence |
|---|---|---|
| credential slot names in `Hush` rows | unchanged | connection names are today's slot names |
| `provider_state` | unchanged | `DefinitionProviderState.CredentialRef` becomes `string`; `CredentialSlotID` already marshals as its name string, so the JSON is identical |
| `installation_metadata` (`Attributes`, `Display`) | gains `layout`, `omitempty` | the verification result fills the same fields and stamps the layout and `CredentialRef`; the upgrade path rewrites them under a declared hook; no migration for a JSON column |
| `Display.CredentialRef` | kept, stamped on every write | the console reads it from `integration.metadata` (`openlane-ui` `installed-integration-card.tsx:43-50`, `document-sync-prompt-dialog.tsx:34`, `primary-directory-prompt-dialog.tsx:48`) |
| definition snapshots | unchanged except slack and email | the surface emits the same credential, connection, and installation-schema JSON; slack's installation schema loses `defaultChannel`; email gains one; topics are not in the surface |
| provider listing (`GET` providers) | unchanged for every key the console reads | the console reads `spec.*`, `operatorConfig.schema`, `userInput.schema`, `operations[].{name,description,requiredPermissions,disabledForAll,configSchema}`, `webhooks`, `credentialRegistrations[].{ref,name,description,recommended,schema}`, and `connections[].{credentialRef,name,description,meta,auth}` where only `auth != null` is tested (`openlane-ui` `lib/integrations/types.ts:76-153`, `utils.ts:44-63`, `flow.ts:31,85`). The response keeps the `Definition` JSON shape with `credentialRegistrations` and `connections` projected from `Connection`; `meta` keeps its untagged `Value` and `AllowCopy` keys; `auth` is emitted as an object when the connection has a flow, since every `AuthRegistration` field becomes `json:"-"`. `clients`, `installation`, `mappings`, and `runtimeIntegration` leave the response; nothing reads them. The OpenAPI spec is regenerated from the response type (`httpserve/specs`), which is the user's generation step. |
| `ConfigureIntegrationRequest.CredentialRef`, `IntegrationAuthStartRequest.CredentialRef` | unchanged | the value is the connection name |

## Runtime resolution

| Path | Today | After |
|---|---|---|
| Active connection | provider state `CredentialRef`, mapped to the connection with that `CredentialRef` (`credentials.go:354`) | the same stored name, mapped to the `Connection` with that name |
| Operation client | `Registry().Client(def, op.ClientRef)`, load that client's slots, `keystore.BuildClient` (`execution.go:615`) | active connection's `Clients[op.ClientRef]`; `LoadCredential` of the one slot; same `BuildClient` cache, keyed by installation, connection, client name, and credential digest, without config |
| Operation request | `Credentials` bindings plus `Client any` | `Client any` only; `ExecuteOperation` loses its `credentials` parameter |
| Credential save | resolve the connection for the slot, reject a different method, validate against the credential form schema (`credentials.go:130`), run the definition health check, refresh under the stored credential, run the resolver, compare, save | the connection name selects the connection and must match the persisted one; validate against `Form` plus `Credential.Validate`; verify under the stored credential and persist; verify under the submitted one; compare; save credential, state, and the verification's metadata |
| Health assessment | definition health check, then `RefreshInstallationMetadata` runs the resolver again (`health.go:264`→`:291`) | the connection's verification once; its result is persisted |
| Auth complete | `connection.Auth.CredentialRef` names the slot (`keymaker/service.go:192`) | the connection's name |
| Disconnect | `DisconnectRegistration.CredentialRef`; hook receives untyped bindings | the connection's name; hook receives `DisconnectRequest[T]` |
| Installation metadata | `RefreshInstallationMetadata` runs `def.Installation.Resolve` after the health check | the verification returns it; callers persist through the existing `saveInstallationMetadata` |
| Upgrade | `credentialKind` reads `CredentialRegistration.Stored`; metadata untouched | reads `Connection.Credential`; `installationKind` re-reads the row in the transaction, resolves the stored `Layout`, conforms `Attributes`, recomputes `Display`, stamps `Layout`; backfills `provider_state` from a lone stored slot |
| Topics | `OperationRef.Registration(definitionID)` and `WebhookEventRef.Registration(definitionID, base)` build them (`ref.go:641`, `:773`) | `registry.Register` stamps them from `def.ID` right after `finalizeDefinition`, before validation and `compileDefinition`, so the indexed entries and `Registry().Definition` carry them; cross-definition topic collisions become impossible and their tests go |
| GraphQL `Credentials` resolver | finds the `CredentialRegistration` whose ref matches `Display.CredentialRef` (`integrationextended.resolvers.go:56-61`) | finds the `Connection` by that name and uses its `Form` |
| Workflow metadata credential entries | built from `def.CredentialRegistrations` (`graphapi/workflow_metadata_extensions.go:124-125`) | built from connections |
| Reconcile loop seeding | `BuildClientForIntegration(ctx, installation, op.ClientRef)` (`reconcile_loops.go:50-51`) | active connection's `Clients[op.ClientRef]` through the same cache |
| Installation creation | pending status keyed on `len(CredentialRegistrations)`, self identity on `len(Connections)` (`integration.go:166,170`) | both keyed on `len(Connections)` |
| keymaker auth state | `AuthState.CredentialRef` is a `CredentialSlotID` stored untagged in Redis as its name string (`keymaker/session_store.go:17-32`, `redis_store.go:49`) | a `string` under the same field name, so in-flight sessions decode across a deploy |

## Type audit

### `internal/integrations/types`

| Type / member | Fate | Why |
|---|---|---|
| `ID[K]`, `credentialSlotKind`, `clientKind`, `CredentialSlotID`, `ClientID`, `NewCredentialSlotID`, `NewClientID` | deleted | slots are connection names and clients are type-derived names, both plain `string` like operation names |
| `CredentialRef[T]` | replaced by `ConnectionRef[T]` | the credential type is the connection's type parameter |
| `CredentialRegistration` | folded into `Connection` | |
| `ConnectionRegistration` | replaced by `Connection` | `CredentialRef` and `CredentialRefs` restated the slot |
| `ClientRegistration`, `ClientRef[C]`, `ClientRef.Using`, `ClientRef.Cast` | deleted | builders live in `Connection.Clients`; handlers receive `C` directly |
| `ClientBuildRequest` | replaced by `ConnectionRequest[T]` on the typed side and `ConnectionInput` on the erased side; `Config` deleted | no builder reads it |
| `ClientBuilderFunc` | kept as the erased builder, taking `ConnectionInput` | |
| `HealthCheckRegistration`, `ClientRef[C].HealthCheck(fn)`, `CredentialHealthCheck(fn)`, `Definition.HealthCheck` | deleted | verification is `ConnectionRef[T].Verified[C, M]`, erased into `Connection.Verify` |
| `VerifyRegistration`, `ConnectionInput` | new | the erased verification and the one erased input shared by build, verify, and disconnect |
| `AuthFlow[T]`, `NewAuthFlow[T]` | new | typed flow; the constructor serves githubapp, azureentraid, and the test fixture, which build `AuthRegistration` by hand |
| `OperationRef.HealthCheck(check *HealthCheckRegistration)` | becomes `HealthCheck(fn func(context.Context, OperationRequest, C) error)` | the probe infers `C` like `Handles`; it never used the registration's `ClientRef` |
| `AuthRegistration.CredentialRef` | deleted | the connection names the slot |
| `AuthCompleteResult.InstallationInput`, `InstallationRequest.Input`, `ReconcileCredential`'s `installationInput` | deleted | only githubapp sets the input (`githubapp/auth.go:117`), with the installation id and org name `mintCredential` already stores in the credential (`githubapp/auth.go:197-202`); slack reads it (`slack/installation.go:30`) but its OAuth flow never sets it |
| `DisconnectRegistration.CredentialRef` | deleted | the connection names the slot |
| `DisconnectRegistration.Disconnect`, `DisconnectFunc`, `DisconnectRequest` | `Disconnect` becomes the erased hook; `DisconnectRequest[T]` is what the typed hook receives | githubapp reads `cred.InstallationID` (`githubapp/auth.go:123`), slack reads only metadata; the other definitions carry a description and no hook |
| `DisconnectRequest.Connection`, `InstallationRequest.Connection` | deleted | never read |
| `OperationRequest.Credentials` | deleted | handlers receive only the client |
| `CredentialBinding`, `CredentialBindings`, `.Resolve`, `.With` | deleted | no erased credential crosses into definition code |
| `InstallationFunc`, `NewInstallationRef(fn)`, `InstallationRef.Resolve` | deleted | the verification produces the metadata |
| `InstallationRef[M]`, `InstallationOf[M]`, `Definition.Installation` | kept; the ref loses its resolver and gains `Upgraded` and `Validated` | the metadata type and its hooks are declared once per definition, like `UserInput` |
| `InstallationRegistration` | embeds `InputRegistration` and gains `Identify` | installation metadata is a conformed kind; `Identify` lets the upgrade recompute `Display` from stored bytes |
| `InstallationRequest` | kept for upgrade and validation hooks; `Credentials` becomes the stored slots as `map[string]CredentialSet` | the upgrade path still sees every stored slot |
| `InstallationIdentifiable` | kept | see the metadata section |
| `IntegrationInstallationMetadata` | gains `Layout string` | the layout name every other conformed kind already stores |
| `IntegrationInstallationIdentity`, `DefinitionProviderState`, `IntegrationProviderState` | kept | persisted shapes are unchanged |
| `OperationRef.Registration(definition DefinitionRef)`, `WebhookEventRef.Registration(definition DefinitionRef, base)` | lose the `DefinitionRef` parameter and no longer set `Topic` | the registry stamps topics from `def.ID` |
| `Definition.CredentialRegistrations`, `Definition.Clients` | deleted | `Definition.Connections` |
| `Definition.CredentialRegistration`, `.ConnectionRegistration`, `.ResolveCredential` | replaced by connection lookup by name, with `Replaces` through `resolveReplaced` | |
| `InputRegistration`, `UserInputRef[T]`, `OperationRef[Config]`, `WebhookRef`, `WebhookEventRef[T]`, `DefinitionRef`, `RuntimeIntegrationRegistration` | kept | outside the connection model |

### `internal/integrations/registry`

| Member | Fate |
|---|---|
| `indexClients`, `validateConnections`, `validateHealthCheck`, `Registry().Client` | deleted; they only check that the duplicated declarations agree |
| `ErrCredentialRefNotDeclared`, `ErrConnectionCredentialRefRequired`, `ErrConnectionCredentialRefNotDeclared`, `ErrConnectionAuthCredentialRefNotDeclared`, `ErrConnectionDisconnectCredentialRefNotDeclared`, `ErrClientRequired`, `ErrHealthCheckRequired`, `ErrHealthCheckClientCredentialMissing`, `ErrConnectionClientRefNotDeclared`, `ErrConnectionHealthCheckHandlerRequired` | deleted with those checks |
| `finalizeClient`, `finalizeConnection`, `finalizeCredential` | deleted; `Connection` is built complete |
| `ErrConnectionClientNotProvided` | new, a connection does not provide a client an operation uses |
| `ErrConnectionVerifyRequired` | new, a connection declares no verification |
| `ErrConnectionVerifyClientNotProvided` | new, a connection's `Verify.ClientRef` is not among its `Clients` |
| `ErrInstallationRequired` | new, a definition with connections declares no `Installation` |
| `ErrInstallationSchemaMismatch` | new, a connection's `Verified[C, M]` names a layout other than `Definition.Installation` |
| `ErrClientNotFound` | kept for an operation naming a client in a definition with no connections and no `RuntimeIntegration` |
| `DefinitionSurface` | credentials and connections from `Definition.Connections`; installation schema and upgrade flag from `Definition.Installation`, as today plus the flag |
| `Register` | stamps `Topic` on each operation and webhook event from `def.ID` before indexing them |

### `internal/integrations/runtime`, `keystore`, `keymaker`, `auth`, `handlers`

| Member | Fate |
|---|---|
| `resolveConnectionForCredential`, `ErrCredentialNotDeclared` | replaced by a name comparison in `ReconcileCredential` and `ErrConnectionMismatch`; the behaviour is unchanged |
| `RefreshInstallationMetadata`, `resolveConnectionIdentity` | deleted; `runConnectionHealthCheck` returns the verification's metadata and callers persist it through the existing `saveInstallationMetadata`; the pre-compare refresh in `ReconcileCredential` becomes a verification under the stored credential |
| `runConnectionHealthCheck` | runs the active connection's `Verify` under its client and returns `IntegrationInstallationMetadata` |
| `runOperationProbe` | unchanged in behaviour; the probe is the typed `OperationRef.HealthCheck` |
| `ExecuteOperation`, `executeOperationInline`, `executeResolvedOperation` `credentials` parameter | deleted; every caller passes `nil` |
| `upgradeInstallation` | gains `installationKind`; re-reads the row after claiming the version, writes `Layout`, conformed `Attributes`, and recomputed `Display` in the transaction; backfills `provider_state` from a lone stored slot |
| `saveInstallationMetadata` | stamps `Layout` from `def.Installation.Name` and `Display.CredentialRef` from the active connection name |
| `EnsureInstallation` | pending status and self identity both keyed on `len(def.Connections)` |
| `keystore.LoadCredentials(refs)` | replaced by the existing `LoadCredential(name)` |
| `keystore.BuildClient` | takes the connection name, client name, builder, and credential; cache key drops config |
| keymaker `AuthState.CredentialRef`, `BeginRequest.CredentialRef`, `CompleteResult.CredentialRef`, `AuthCompleteHookFunc` | `string` connection name |
| `auth.OAuthRegistrationOptions.CredentialRef` | deleted; `OAuthRegistration` returns `AuthFlow[T]` |
| `auth.FederatedTokenSource(ctx, req ClientBuildRequest, spec)` | takes the token manager and organization id directly |
| `handlers.ListIntegrationProviders` | projects connections into the existing response arrays |
| every definition's `installation.go` resolver and `HealthCheck` result struct, slack's `InstallationInput` type and the `DefaultChannel` field it fed into slack's metadata type | deleted; slack's client keeps reading `defaultChannel` from the integration `metadata` map (`slack/client.go:25-38`) through a local type, since it decoded into the metadata type today; nothing writes that key on the customer path, so the field's removal is behaviour-neutral |
| `internal/testutils/integrations`, `slack/testing.go` | the shared fixture has three connections and one client bound to one slot, and `UnresolvableOp` exercises a connection that cannot serve an operation; under the coverage rule that is a registration error, so the fixture gives every connection a client and the unresolvable case moves to a missing credential row; `mockhttp.go` builds its client with `.Using`; `slack/testing.go` mutates `def.Clients[i].Build` (`:130-134`) and is rewritten against `Connection.Clients` |

## Resolved decisions

- **Verification per connection.** A definition-level check receiving the active connection's credential would have to branch on which connection is active, which is the probing this design removes from client builders; awssecurityhub's resolver branches that way today. Per connection, the client, the credential type, and the metadata producer are fixed by construction. Operation probes stay on the operation.
- **The connection verification is the identity call.** Scope checks are operation probes. Its failure keeps the connection check's classification on every path; the health path previously logged a failed second call, and no longer has one.
- **Metadata type and hooks are declared once, on `Definition.Installation`.** `UpgradeFunc` and `ValidateFunc` are closures the registry cannot compare across connections. `InstallationOf[M]` is the same shape as `UserInputRefOf[T]`; a connection declares only what differs per method, the client and the verification, and the registry checks its `M` against the definition's.
- **The pre-compare refresh stays.** It is a verification under the stored credential, and it is what lets a provider-reported id change reconnect. Dropping it was reviewed and rejected.
- **Method switching stays rejected, and `Display.CredentialRef` is stamped on every write.** Both preserve today's behaviour; the second fixes the health path, which would otherwise blank the ref and misdirect the reconnect form.
- **Operations receive only the client.** The seven handlers that decode credentials read non-secret scope the client can carry; cloudflare already carries it. Disconnect hooks receive `DisconnectRequest[T]`.
- **Installation metadata is conformed on upgrade, with a stored layout name.** It is the one stored kind the upgrade path skipped; adding it closes the gap a derivation change opens and removes the need for the pre-compare refresh. The `Layout` key makes its hook receive the real `from` like user input and credentials do, at the cost of one additive JSON key.
- **Refs do not take a definition id.** `Registration(definitionID)` existed only to build the topic; the registry has `def.ID` and stamps topics itself. Going further to a no-call `[]Operation` interface touches 21 readers of `def.Operations` and is left as a follow-on.

## Change set

- `common/openapi`: `IntegrationInstallationMetadata.Layout`.
- `types`: `Connection`, `VerifyRegistration`, `VerifyRequest`, `ConnectionRef[T]`, `ConnectionRequest[T]`, `DisconnectRequest[T]`, `Connector`, `AuthFlow[T]`, `NewAuthFlow[T]`, `InstallationOf[M]` with hooks and `Identify`, typed operation probe; `OperationRef.Registration()` and `WebhookEventRef.Registration(base)` without the definition; delete the types in the audit.
- `auth`: `OAuthRegistration` returns `AuthFlow[T]`; `FederatedTokenSource` takes the manager and organization id.
- `registry`: build and validate connections, require a verification per connection whose client the connection provides, one installation layout per definition, stamp topics, surface from connections, delete the agreement checks.
- `runtime`: connection lookup by name with `ErrConnectionMismatch`, `resolveOperationClient`, `ReconcileCredential` with the refresh as a stored-credential verification, `runConnectionHealthCheck` returning metadata, `RunHealthAssessment`, `Disconnect`, `credentialKind`, `installationKind` with the in-transaction re-read and the provider-state backfill, `saveInstallationMetadata` stamping layout and connection name, `EnsureInstallation`, `BuildClientForIntegration`, `ExecuteOperation` signature.
- `keystore`, `keymaker`: string connection names, single-credential load, config-free cache key.
- `graphapi`: the `Credentials` resolver and workflow metadata read connections.
- `handlers`: provider listing projection.
- Every definition's `builder.go` (connections, and the `Registration(definitionID)` calls lose their argument), `client.go`, and health check; the `operation_*.go` files that read credentials move their scope reads onto the client; the `installation.go` files and `HealthCheck` result structs are deleted.
- `internal/testutils/integrations` fixtures, `slack/testing.go`, and the tests that construct these registrations; tests asserting a topic on an unregistered registration assert it through a registry instead.
- Regeneration, which is the user's step: the surface snapshots (slack and email move; the committed files also predate the `hash` key) and the OpenAPI spec for the provider listing.

## Status of the implementation

Every item in the change set is applied in the working tree, unvalidated: no build, lint, or test has been run. Remaining steps are the user's:

- `task jsonschema:integration-schema` (the `integration-schema` task in `jsonschema/Taskfile.yaml`) to rewrite the surface snapshots; the committed files predate the `hash` key, and slack and email change.
- The OpenAPI spec regeneration for the provider listing, whose response type is now `handlers.IntegrationProvider`.
- `task go:lint` and the targeted `task go:test -- -tags test` runs over `internal/integrations/...`, `internal/keystore`, `internal/keymaker`, `internal/httpserve/handlers`, and `internal/graphapi/...`.

Known gaps in test coverage after the port: `ErrConnectionMismatch` and the provider-state backfill have no unit test (both sit behind the keystore and a transaction); the four `installation_test.go` cases per definition that needed a provider call are gone with the resolvers, and `verify` is covered only where it needs no network. Decisions taken during the port that the document did not spell out: keycloak's token method is `ClientToken` so it does not shadow the SDK's `Login`; googleworkspace's `Users.List` and okta's `GetUser("me")` became operation probes on their directory syncs; googledrive's `Files.List`, authentik's `CoreUsersMeRetrieve`, and tailscale's plain user list were dropped as redundant with the identity call; the erased credential decode treats an absent payload as the zero value so a disconnect without a credential row falls back to stored metadata as before; a materialized `Connection` is itself a `Connector`, which is how `slack/testing.go` swaps client builders.

## Checked and cut from earlier drafts

- **Typed `CredentialSlotID` / `ClientID`.** Redundant with connection names and type-derived client names.
- **Changing `installation_metadata` storage beyond the layout key.** The verification result fits the existing fields; only the layout name is added, and only because every other conformed kind stores one.
- **A no-call `[]Operation` interface for operations and webhook events.** Dropping the definition-id argument gets the oddity out of the builder for about eight lines; the interface form costs a registry accessor plus 21 call sites, and is a follow-on.
- **Removing `InstallationIdentifiable`.** It costs reconnect-form prefill and forces a snapshot version bump; see the metadata section.
- **Moving the active connection out of `provider_state` or `display.credentialRef`.** Both are read where they are, and the JSON does not change.
- **A console contract change for the provider listing.** A projection keeps the response the console reads.
- **Skipping operations a method does not support.** Today one builder serves every method, so the validation rule preserves current behaviour.
- **A definition-level health check with a `ConnectionCredential` on the request.** It re-introduces per-method branching inside the check and keeps an erased credential type on every operation request.
- **Health checks per operation only.** Operation probes stay; the connection's verification is what produces the installation identity, and it has to run before the credential is saved.
- **Dropping the pre-compare refresh.** It is not backfill-only: a provider-reported id change, an outdated installation reconnecting before its upgrade, and a rolling deploy all depend on it, and a test asserts the first.
- **Allowing method switching.** Today's rejection stays; widening it is a product change with a retired-slot cleanup the design would have to specify.
- **Making `Verified` optional for email.** A trivial `M` with self identity costs one version bump and keeps the rule uniform.
- **A separate `CredentialCarrier` interface.** Nothing outside the connection's own closures needs the credential.
- **Re-deriving metadata during upgrade.** It would put a provider call inside the upgrade transaction on every job bootstrap; conforming the stored document matches the other three kinds.
