# Integrations type system design

Status: revision 4. Parts 2–7 and 9 are reconciled with the code in the working tree, which is the source of truth; where revision 3 and the code disagreed, this revision follows the code. Appendix A keeps the revision 3 adversarial review log. Appendix B lists the implementation decisions taken after revision 3.

Scope: `internal/integrations` (types, registry, runtime, operations, providerkit, definitions), `internal/keystore`, `common/openapi`, `common/models`, `internal/ent/schema/integration.go`, `config/config.go`, and the REST/GraphQL/workflow/CLI consumers of installation data.

Conventions:
- `B:` means the branch working tree (`feat-integrationinstallationreconcile`, staged and unstaged).
- `M:` means `main`, read with `git show main:<path>`.
- `ii/` means `internal/integrations/` and `defs/` means `internal/integrations/definitions/`.
- Every file:line below was opened and checked. Estimates in Part 8 are labelled with their basis.
- Part 1 and Appendix A describe the branch as it stood before the redesign (the revision 3 baseline). Their `B:` file:line citations point at that state and no longer match the working tree. Text added in revision 4 cites file paths without line numbers.

---

## Part 1 — Analysis

### 1.1 Where the branch stands

The branch introduced `Input[T]`:
- a reflected layout and schema per stored type,
- an `Upgrader` hook detected on `*T`,
- a `Validator` hook,
- `Replaces`.

Those pieces are kept. What changes is what they are coupled to.

1. **A kind discriminator underlies every stored type.**
   - `InputKind` and its six constants are at B:ii/types/kind.go:4-19, and every `Input[T]` carries one (B:ii/types/input.go:33).
   - The switch is in `Definition.Inputs` (B:ii/types/definition.go:216-233).
   - The runtime branches on it at B:ii/runtime/inputs.go:79-96, 309-318, 339 and 377.
   - The registry selects and enumerates by kind at B:ii/registry/registry.go:114-121 and 338.
2. **Typed handles are erased into `InputRegistration`** (B:ii/types/definition.go:87-105). These are held in six differently shaped `Definition` fields (B:ii/types/definition.go:42-64), and `Definition.Inputs(kind)` flattens them back into one list.
3. **One literal storage type sits on two fields.**
   - `user_inputs` and `operation_inputs` are both `openapi.StoredDocuments` (B:internal/ent/schema/integration.go:108,115).
   - `installation_metadata` embeds the same `StoredDocument` (B:common/openapi/integration_models.go:67-73).
4. **Main's data is read from `config.clientConfig` by inferring each operation's section key from its name** (`lo.CamelCase(name)`, B:ii/runtime/inputs.go:341).
   - Where main's key differs from the type name, the branch declares a retired operation named after main's key (B:defs/cloudflare/builder.go:67; B:defs/githubapp/builder.go:100).
   - The same `Replaces` list also drives run renames, loop purges and health-key moves (B:ii/runtime/inputs.go:201-212, 254-265, 279-306).
   - So `findingSync` is also treated as a former operation name. Main's operation names were `FindingsSync` and `VulnerabilitySync`.
5. **A `stored` flag separates stored operation input from per-run payload config.** It is set through two constructor pairs (B:ii/types/ref.go:375-393). Both kinds share one registration shape, `InputRegistration.Schema` (B:ii/types/operation.go:114-118).
6. **Shared concepts are declared per definition and read by key.**
   - `primaryDirectory` appears in 7 user inputs and is read by key at B:ii/runtime/inputs.go:30,245.
   - `name` is read by key at B:ii/runtime/inputs.go:27,240.
   - `defaultMessaging` is read at B:internal/workflows/engine/integration_executor.go:311-320.
   - `primary` is read at B:internal/graphapi/internalpolicyextended_helpers.go:68-73.
7. **The version gate is entered from 8 call sites:** B:ii/runtime/credentials.go:66, dispatch.go:50, execution.go:70/299/385, health.go:270, upgrade_sweep.go:41, webhooks.go:198.
8. **Main's `installation_metadata.attributes` has no reader on the branch.**
   - Main stored `{attributes, display}` (M:common/openapi/integration_models.go:66-71).
   - The branch type is `{name, layout, data, display}` (B:common/openapi/integration_models.go:67-73).
9. **Adding an integration touches code outside its package:**
   - the builder list at B:defs/catalog/catalog.go:32-56,
   - a config field per definition at B:defs/catalog/types.go:20-54, wired at B:config/config.go:83,
   - per-integration selection at B:internal/graphapi/internalpolicyextended_helpers.go:29 and B:internal/graphapi/internalpolicyextended.resolvers.go:48-53.
10. **The working tree does not compile.**
    - `inputSchema` and `inputSchemas` sit inside a `/* */` block (B:internal/graphapi/integrationextended.resolvers.go:120-151) but are called at :70, :91 and :111.
    - This is inferred from undefined symbols; no build was run.

### 1.2 Files read

Read in full:
- **On the branch:** every non-test file in `ii/types`, `ii/registry`, `ii/runtime`, `ii/operations`, `ii/providerkit`, `ii/auth`, `internal/keystore`, and `defs/{awssecurityhub,cloudflare,githubapp,okta,scim,slack}`.
- **On main:** their counterparts, plus the main `UserInput` of every definition with stored operations.

Non-test `ii/auth` files are byte-identical to main. `auth/oauth_registration_test.go` differs.

Non-test line counts on the branch:

| Package | Lines |
|---|---|
| types | 2041 |
| registry | 1046 |
| runtime | 3682 |
| operations | 1332 |
| providerkit | 469 |
| keystore | 273 |

### 1.3 Inventory: exported types in `ii/types` (branch)

| Type | File:line | Fate |
|---|---|---|
| `InputKind` + 6 consts | kind.go:4-19 | removed |
| `Upgrader` | input.go:13 | kept, same signature; contract stated (§2.4) |
| `Validator`, `Identifier` | input.go:19, 25 | kept |
| `Input[T]` | input.go:31 | kept, same name; loses `kind`, embeds `Identity`, gains decode/conform methods |
| `DefinitionSpec` | definition.go:14 | kept |
| `Definition` | definition.go:38 | fields unexported, built by `Define` |
| `GalaListenerRegistration` | definition.go:68 | kept |
| `OperatorConfigRegistration` | definition.go:76 | erased view of `OperatorConfigRef[T]` |
| `UpgradeFunc`, `ValidateFunc` | definition.go:82, 85 | kept, internal to `Input[T]` |
| `InputRegistration` | definition.go:88 | removed |
| `CredentialRegistration`, `ConnectionRegistration` | definition.go:108, 125 | erased views produced by `declare` |
| `HealthCheckRegistration`, `MetaInfo` | definition.go:143, 151 | kept |
| `InstallationRequest` | installation_registration.go:11 | kept |
| `InstallationFunc`, `InstallationRegistration` | installation_registration.go:23, 26 | erased view of `InstallationRef[T]` |
| `RuntimeIntegrationRegistration` | runtime_integration.go:9 | erased view of `RuntimeConfigRef[T]`; loses embedded `InputRegistration` |
| aliases `IntegrationInstallationMetadata`/`Identity` | installation.go:6, 9 | kept |
| `CredentialSet`, `CredentialBinding(s)` | credential_binding.go:8-46 | kept |
| auth, client, execution, dispatch, runtime-services, disconnect, scope, mapping, errors, river types | various | kept |
| `DefinitionRef`, `CredentialSlotID`, `ClientID` | ref.go:21, 61, 241 | kept |
| nine typed handles | ref.go:95-693 | kept, same names; `Registration(base)` replaced by `Declaration` |
| `OperationSettings`, `OperationInput`, `OperationSettingsFrom` | operation.go:15-37 | `OperationSettings` → providerkit; `OperationInput` → `Configurable`; `OperationSettingsFrom` removed |
| `OperationRegistration` | operation.go:114 | erased view; loses embedded `InputRegistration` and `Stored`; keeps `Name`, `Description`, `Schema` as fields; gains `Settings`/`Config` closures |

### 1.4 Inventory: kind constants, switches, translators

| # | Site | What it does |
|---|---|---|
| K1 | B:ii/types/kind.go:6-19 | six `InputKind` constants |
| K2 | B:ii/types/input.go:59 | `kind != KindInstallation` gate |
| K3 | B:ii/types/definition.go:216-233 | `Inputs(kind)` switch |
| K4 | B:ii/types/definition.go:235-252 | `ResolveInput` including the single-registration fallback (:247) |
| K5 | B:ii/types/definition.go:158-170 | `optionalInput` (serves K3); `findNamed` (serves K4 and the name lookups at :174, 184, 189, 207) |
| K6 | B:ii/runtime/inputs.go:79-96 | `switch write.Kind` |
| K7 | B:ii/runtime/inputs.go:99, 104, 108 | `conformKind` once per kind |
| K8 | B:ii/runtime/inputs.go:309-318 | `inputSentinel(kind)` |
| K9 | B:ii/runtime/inputs.go:339, 377 | `kind == KindOperation` special cases |
| K10 | B:ii/runtime/inputs.go:321-351 | `storedOrConfig`/`configDocuments` |
| K11 | B:ii/runtime/inputs.go:404-418 | `mergeOperationSettings` |
| K12 | B:ii/runtime/credentials.go:155-166 | `bindingsFor` |
| K13 | B:internal/keystore/store.go:64-118 | hush rows ↔ `StoredDocuments` |
| K14 | B:ii/registry/registry.go:90-126, 129, 141 | `DefinitionSurface` → `SurfaceInput{Kind string}` |
| K15 | B:ii/registry/registry.go:338-358 | `inputKinds` loop |
| K16 | B:ii/registry/finalize.go:16-44 | per-field re-mapping |
| K17 | B:ii/registry/finalize.go:77-118 | `if operation.Stored` filterExpr wrapper |
| K18 | B:internal/httpserve/handlers/integration_config.go:131-157 | `inputWrites` |
| K19 | B:internal/httpserve/handlers/integration_config.go:137; integration_flow.go:69 | `def.Inputs(types.KindUserInput)` |
| K20 | B:internal/workflows/engine/integration_executor.go:296-307 | `Inputs(KindUserInput)[0]` |
| K21 | B:internal/graphapi/internalpolicyextended_helpers.go:63-73 | `Inputs(KindUserInput)[0]` |
| K22 | B:internal/graphapi/integrationextended.resolvers.go:70-151 | `inputSchema(s)` |
| K23 | B:ii/runtime/upgrade.go:15, 57 | provider_state `"credentialRef"` key read |
| K24 | B:ii/types/operation.go:153-158 | `DisabledFor(raw)` |
| K25 | B:ii/types/ref.go:375-393 | four operation constructors + `stored` |
| K26 | B:defs/cloudflare/types.go:18, 20 | one type as both operator config and runtime config |
| K27 | B:defs/catalog/catalog.go:32-56; catalog/types.go:20-54 | per-integration builder list and config fields |
| K28 | B:ii/runtime/inputs.go:443-450 | `metadataMirror` |
| K29 | B:internal/graphapi/internalpolicyextended.resolvers.go:48-53; internalpolicyextended_helpers.go:29 | switch / `DefinitionIDIn` over two definition IDs |

### 1.5 Inventory: duplicates across definitions, and string-key reaches into data

| # | Concept | Sites |
|---|---|---|
| D1 | `type DirectorySync struct` | 12 definitions |
| D2 | `DisableGroupSync` | authentik, awssecurityhub, azureentraid, githubapp, keycloak, tailscale |
| D3 | `EnableGroupSync` (inverse) | okta |
| D4 | `types.OperationSettings` embeds | 26 sites |
| D5 | `PrimaryDirectory` in UserInput | authentik, azureentraid, googleworkspace, keycloak, okta, scim, zitadel |
| D6 | CEL `"installation.primary_directory"` | 12 `mappings.go` files |
| D7 | `DefaultMessaging` | slack, microsoftteams |
| D8 | `Primary` | googledrive, onedrive, `operations.UserInput` (B:ii/operations/ingest_document.go:13) |
| D9 | directory op chain | 12 `OperationOf[DirectorySync]()`; `DirectoryMappings` used 9× |
| D10 | required `Name` mirrored to installation name | scim |

| # | Site | Key |
|---|---|---|
| S1 | B:ii/runtime/inputs.go:27, 240 | `"name"` |
| S2 | B:ii/runtime/inputs.go:30, 245 | `"primaryDirectory"` |
| S3 | B:ii/runtime/inputs.go:341 | `lo.CamelCase(opName)` into `config.clientConfig` |
| S4 | B:ii/runtime/upgrade.go:15, 57 | `"credentialRef"` |
| S5 | B:ii/runtime/execution.go:88, 324, 468; dispatch.go:61; health.go:444; reconcile_loops.go:143; B:ii/operations/ingest.go:836 | `OperationInputs.Data(opName)` |
| S6 | B:internal/workflows/engine/integration_executor.go:313 | `defaultMessaging` |
| S7 | B:internal/graphapi/internalpolicyextended_helpers.go:68-73 | `primary` |
| S8 | B:defs/githubapp/webhook.go:308 | `sqljson.Path("display","externalId")` |
| S9 | B:ii/runtime/execution.go:134; health.go:175, 221; reconcile_loops.go:142 | `Health.UnhealthyOperations[opName]` |

### 1.6 Integration storage and main's production shape

| Field | M | B | Branch-only |
|---|---|---|---|
| `config` (schema :87) | `IntegrationConfig{ClientConfig "clientConfig"}` | same | no |
| `provider_state` (:101) | `IntegrationProviderState{Providers map[string]json.RawMessage}` | same | no |
| `installation_metadata` (:94) | `{attributes, display}` | `{name, layout, data, display}` | no |
| `user_inputs` (:108), `operation_inputs` (:115), `credential_ref` (:122) | — | present | yes |
| hush `credential_set` | `{data}` | adds `layout` | no |

**Branch-only columns have no production data.**
- `git show main:internal/ent/schema/integration.go | grep -cE '"(user_inputs|operation_inputs|credential_ref)"'` returns 0.
- `grep -rlE 'user_inputs|operation_inputs|credential_ref' db/` returns nothing.
- The branch migration pair adds `user_input` and `operation_config` as `jsonb NULL`: db/migrations/20261001161106_integrationoperationconfig.sql:2 and db/migrations-goose-postgres/20261001161120_integrationoperationconfig.sql:3. This revision targets exactly those two columns.

**Main writers:**

| Data | Main writer | Value / note |
|---|---|---|
| `config.clientConfig` | M:ii/runtime/credentials.go:180-182 | whole `UserInput` |
| `provider_state` | M:ii/runtime/credentials.go:515-524 | via `DefinitionProviderState` (M:ii/types/definition.go:162-206) |
| `installation_metadata` | M:ii/runtime/credentials.go:253-256 | value built at M:ii/types/ref.go:252-262 |
| hush | M:internal/keystore/store.go:112 | `secret_name` = slot name |

Main also mirrored `name` and `primaryDirectory` from the user input into their columns (M:ii/runtime/credentials.go:184-192).

**Main `UserInput` shapes for definitions with stored operations:**

| Shape | Definitions (main section keys → branch op names) |
|---|---|
| nested | awssecurityhub (`findingSync`→FindingSync, `directorySync`, `checkSync`, `assetSync`) |
| | cloudflare (`directorySync`, `assetSync`, `findingSync`→**FindingsSync**) |
| | gcpscc (`findingsSync`) |
| | githubapp (`findingSync`→**VulnerabilitySync**, `directorySync`, `repositorySync`) |
| | oci (`findingsSync`) |
| | slack (`defaultMessaging`, `directorySync`) |
| | tailscale (`directorySync`, `assetSync`) |
| flat (top-level `filterExpr` + op fields) | authentik, azureentraid, azuresecuritycenter, googledrive, googleworkspace, keycloak, okta, onedrive, scim, zitadel |
| flat, no stored op on branch | microsoftteams. Its `filterExpr` was inert on main, which has no ingest op for it. |

- `lo.CamelCase(<op name>)` equals main's key for every nested op except the two in bold. Both of those are stored under `findingSync`.
- No flat definition has a top-level key equal to `lo.CamelCase(<op name>)`.
- Main's resolver fell back to the top-level `filterExpr` for any op without a `ConfigResolver`, so the flat `filterExpr` applied to every op of the installation (M:ii/operations/ingest.go:875-899).
- Branch per-run config tag change: githubapp `max_repos` (M:defs/githubapp/operation_vulnerability_collect.go:19) became `maxRepos` (B:defs/githubapp/types.go:72). Per-run configs persist outside conform, in workflow action configs and in `integration_run.operation_config`.

---

## Part 2 — The construct set

### 2.1 Principles

1. **`Input[T]` stays the one generic that carries shared behaviour:** reflected layout and schema, identity, and the `Conform`/`Accept` pipeline, which detects `Upgrader`/`Validator` on `*T` by interface assertion on every call.
2. **The handles keep their names.** They register through `types.Define(spec, decls...)`, with no `Registration(base)`. Each thing is declared once, and every reference to it is validated by `Define`, which is the single validator of a definition's internal consistency.
3. **The `types` package owns all iteration over stored data.**
   - `Document` has unexported methods and is implemented by the four stored handles (`CredentialRef`, `UserInputRef`, `OperationRef`, `InstallationRef`).
   - The runtime calls exported functions in `types` (`Conform`, `SurfaceOf`, `Form`, `UserInputAs`), never `Document` methods.
   - There is no kind value, no kind marker method and no type switch over documents. Behaviour differs per document only through each handle's own method bodies, and `Definition` holds typed views (`credentials`, `userInput`, `installation`) written by the declaring handle.
4. **Each stored field has its own named type, embedding `models.Revision{Layout, Data}`,** except `installation_metadata`, whose payload key is fixed by production data.
   - Main's `config` column is never written again. It is the frozen source that main-era rows are read from once (§5.5), and it is dropped in Phase 4 (next release).
5. **One upgrade mechanism: the `Upgrader` method on the stored type.**
   - The common evolution is ConformToSchema plus, when needed, `UpgradeFrom`. Common cases: adding or removing fields, and filling new fields with data that didn't exist before.
   - A type rename is the escape hatch. The stored `layout` tells `UpgradeFrom` which type wrote the document, and `Replaces` keeps identity continuity.
6. **Conform is isolated, offline and order-safe.**
   - It never calls a provider API.
   - A failing document degrades only what it belongs to.
   - Every upgrade writes all conformed fields and the new version in one transaction.
   - A binary never writes to a row whose version is newer than its own. Versions are ULIDs, so they order by creation time (§5.4).

### 2.2 Storage types

These are the types in `common/models/revision.go`, `common/models/credentialset.go` and `common/openapi/integration_models.go`.

```go
// common/models/revision.go

// Revision is a stored document and the layout it conforms to; an empty Layout marks a document written before layouts were stamped
type Revision struct {
	// Layout is the reflected identifier of the type Data conforms to
	Layout string `json:"layout,omitempty"`
	// Data is the stored document
	Data json.RawMessage `json:"data,omitempty"`
}

// common/models/credentialset.go — same JSON as the pre-redesign CredentialSet

// CredentialSet is one hush credential row's stored document
type CredentialSet struct {
	Revision
}
```

```go
// common/openapi/integration_models.go

// IntegrationUserInput is the installation's user input document
type IntegrationUserInput struct {
	models.Revision
}

// IntegrationOperationConfig is one operation's stored per-installation config
type IntegrationOperationConfig struct {
	// Operation is the operation identity the config belongs to
	Operation string `json:"operation"`
	models.Revision
}

// IntegrationOperationConfigs is every stored operation config of an installation, sorted by operation
type IntegrationOperationConfigs []IntegrationOperationConfig

// IntegrationInstallationMetadata keeps main's attributes key, adds the layout stamp, and carries the derived identity
type IntegrationInstallationMetadata struct {
	// Layout is the reflected identifier of the installation metadata type
	Layout string `json:"layout,omitempty"`
	// Attributes is the installation metadata document
	Attributes json.RawMessage `json:"attributes,omitempty"`
	// Display is the normalized installation identity
	Display IntegrationInstallationIdentity `json:"display,omitzero"`
}
```

`common/openapi` already imports `common/models` (common/openapi/models.go), so embedding `models.Revision` creates no cycle.

`IntegrationProviderState` stays as main has it (M:common/openapi/integration_models.go:74-77). Its typed helpers are `Definition.ProviderState` and `Definition.WithProviderState` over `DefinitionProviderState{CredentialRef}` (ii/types/definition.go), restored from main; the branch-only `credential_ref` field and its key read (K23) are gone from the ent schema. `IntegrationConfig` stays only until the `config` column is dropped (Phase 4). `StoredDocument`/`StoredDocuments` are deleted from `common/openapi/integration_models.go` (Phase 3).

### 2.3 Identity and `Input[T]`

```go
// Identity is a declaration's stable name and the retired names it takes over
type Identity struct {
	name     string
	replaces []string
}
func (i Identity) Name() string
func (i Identity) Replaces() []string // sorted, deduplicated, nil when none

// Input is the reflected layout, schema, and lifecycle of one stored type T
type Input[T any] struct {
	Identity
	layout string
	schema json.RawMessage
}
func (i Input[T]) Layout() string
func (i Input[T]) Schema() json.RawMessage

// Conform decodes payload into T, applies T's upgrade with the layout it was stored under, conforms the re-encoding to the schema, validates it, and returns T with its payload
func (i Input[T]) Conform(ctx context.Context, req InstallationRequest, layout string, payload json.RawMessage) (T, json.RawMessage, error)

// Accept validates a submitted payload strictly against the schema, then conforms it under the current layout
func (i Input[T]) Accept(ctx context.Context, req InstallationRequest, payload json.RawMessage) (T, json.RawMessage, error)
```

- All of this is in ii/types/input.go. Retired names are added through each handle's `Replacing(...)`, which calls the unexported `Identity.replacing`; each handle's own source lookup walks `Name()` then `Replaces()`.
- `Identity` is also embedded by `WebhookRef`, which removes the second `replaces` implementation.
- `newInput[T](name)` reflects the schema with `jsonx.SchemaFrom[T]` and the layout with `jsonx.SchemaID`, and names the input after the layout when `name` is empty. It stores no hook closures: `Conform` asserts `Upgrader` on `&decoded` and `Validator` on `&out` each time it runs.

**Pipeline order inside `Conform`:**
1. plain decode into `T`
2. `UpgradeFrom` when `*T` implements `Upgrader` (receiver pre-decoded)
3. encode
4. `jsonx.ConformToSchema` on the encoding
5. `jsonx.Validate`
6. decode the conformed payload into a fresh `T`
7. `(*T).Validate` when `*T` implements `Validator`

`ConformToSchema` fills a missing required property only when the schema declares a `Default` (pkg/jsonx/schema_conform.go:72-91). A field without `omitempty` is always encoded, so it counts as present and its default never applies.

### 2.4 Hooks on stored types

```go
// Upgrader maps a stored document onto the receiver, which already holds the plain decode of stored
type Upgrader interface {
	// UpgradeFrom fills or reshapes fields from stored, which was persisted under layout from ("" for documents written before layouts were stamped)
	UpgradeFrom(ctx context.Context, req InstallationRequest, from string, stored json.RawMessage) error
}

// Validator checks a schema-valid document for constraints the schema cannot express
type Validator interface {
	Validate(ctx context.Context, req InstallationRequest) error
}

type Identifier interface {
	InstallationIdentity() IntegrationInstallationIdentity
}

// Configurable is implemented by operation configs persisted per installation; implementing it makes an operation stored
type Configurable interface {
	Validator
	Disabled() bool
	Filter() string
}
```

**`Upgrader` contract.** The signature is unchanged from B:ii/types/input.go:13-16.
- The receiver is pre-decoded.
- `UpgradeFrom` runs on every conform pass of the document, so it must be idempotent.
- `from` is the stored layout:
  - `""` for documents of a never-conformed installation, which only main wrote,
  - the current layout for `Accept`.
- Typical bodies are "map when the source key is present" and "fill when empty from `req`".

**`providerkit.OperationSettings{Disable, FilterExpr}`** (ii/providerkit/settings.go) implements `Configurable`. `Validate` compiles `FilterExpr` with `providerkit.ValidateExpr` and wraps a failure in `ErrOperationFilterExprInvalid`.
- Embedding it gives a config all three methods. That replaces K17, K24, `OperationSettingsFrom` and the `operationSettings()` marker.
- It lives in providerkit because providerkit imports types.
- **Storedness is the interface assertion and nothing else.** `OperationRef.stored()` is `any(new(Config)).(Configurable)`. There is no registration-time reflective check: the revision 3 check that rejected configs embedding `OperationSettings` without satisfying `Configurable` was removed together with `ErrOperationSettingsNotConfigurable` (Appendix B); `providerkit.OperationSettings` implements `Configurable` directly through its value-receiver `Disabled`, `Filter` and `Validate`. A config that makes `Validate` ambiguous through two same-depth embeddings is simply not stored.

### 2.5 State, documents and exported conform functions

```go
// InstallationState is the typed snapshot of what one installation stores
type InstallationState struct {
	Integration *generated.Integration
	Connection  CredentialSlotID
	Credentials CredentialBindings
	UserInput   openapi.IntegrationUserInput
	Operations  openapi.IntegrationOperationConfigs
	Metadata    openapi.IntegrationInstallationMetadata
}

// Declaration is anything a definition is built from
type Declaration interface {
	declare(*Definition) error
}

// Document is a declaration that persists a document; its methods are called only by this package's exported functions
type Document interface {
	Declaration
	conform(ctx context.Context, src Sources, next *InstallationState) error
	stage(ctx context.Context, req InstallationRequest, writes Writes, next *InstallationState) (bool, error)
	formInto(state InstallationState, set *FormSet) error
	surface(*Surface)
	failure(err error, selected []CredentialSlotID) Failure
}

// Sources is what a conform pass reads: the prior state and, for a never-conformed installation, main's frozen user input
type Sources struct {
	Request InstallationRequest
	Prior   InstallationState
	// Legacy is config.clientConfig when the installation's definition_version is empty, nil otherwise
	Legacy json.RawMessage
}

// Writes is a set of submitted documents
type Writes struct {
	UserInput  json.RawMessage
	Operations map[string]json.RawMessage
	Credential *CredentialWrite // {Slot CredentialSlotID; Data json.RawMessage}
}

// Failure is one document that failed to conform
type Failure interface {
	error
	// Aborts reports whether the failure aborts persisting the installation
	Aborts() bool
	// Apply records the failure's health consequence through marker
	Apply(ctx context.Context, marker HealthMarker) error
}

// HealthMarker records conform failures against installation health
type HealthMarker interface {
	MarkOperation(ctx context.Context, operation, reason string) error
	MarkInstallation(ctx context.Context, reason string) error
}

// Conform conforms every document of def from src, stages writes, and reports per-document failures; failed documents keep their prior value in next
func Conform(ctx context.Context, def Definition, src Sources, writes Writes) (InstallationState, []Failure, error)

// SurfaceOf projects def onto its committed surface
func SurfaceOf(def Definition) Surface

// Form returns the UI schema of each stored document with stored values as defaults
func Form(def Definition, state InstallationState) (FormSet, error)

// UserInputAs decodes the installation's user input and reports whether it implements I
func UserInputAs[I any](def Definition, state InstallationState) (I, bool, error)
```

These live in ii/types/{state,declare,conform,surface,form}.go.

**Document methods.** Each stored handle implements all five; there is no default implementation and no switch:

| Handle | `conform` | `stage` | `formInto` | `surface` | `failure` |
|---|---|---|---|---|---|
| `CredentialRef` | prior binding under its name, else under a `Replaces` name, moved onto the slot | submitted credential for its slot, through `Accept` | none | `Credentials` entry | `credentialFailure`, aborting when the slot belongs to the selected connection |
| `UserInputRef` | prior user input, else the whole `Legacy` document | submitted user input, through `Accept` | `FormSet.UserInput` | `UserInput` entry | `userInputFailure`, aborting |
| `OperationRef` | stored ops only: prior config under its name or a `Replaces` name, else the `Legacy` section (§5.5) | submitted config for a stored op, through `Accept` | stored ops only: key under `FormSet.Operations` | `Operations` entry | `operationFailure`, not aborting |
| `InstallationRef` | prior `Attributes` through `Input.Conform`; never calls `resolve` | nothing (metadata is derived, not submitted) | none | `Installation` entry | `installationFailure`, not aborting |

**`Failure` implementations** (ii/types/conform.go). Each is an unexported struct embedding the conform error and exposing `Aborts`, `Apply` and `Unwrap`:
- `operationFailure` — `Aborts() == false`; `Apply` calls `MarkOperation(name, err)`.
- `userInputFailure` — `Aborts() == true`; `Apply` calls `MarkInstallation`.
- `credentialFailure` — aborts, and `Apply` calls `MarkInstallation`, only when the slot belongs to the selected connection; otherwise a no-op.
- `installationFailure` — `Aborts() == false`; `Apply` records nothing, and the runtime logs it.

The interface carries `error`, `Aborts` and `Apply`; `Unwrap` is on each implementation, so `errors.Is` reaches the wrapped sentinel. The runtime never inspects which implementation it holds.

**`Conform` flow.** `next` starts as `src.Prior` (`next := src.Prior`), so anything no document rewrites is carried forward unchanged, including stored configs of operations the definition does not declare.
1. `checkWrites` rejects a write to a document the definition does not declare (`ErrInputNotFound`): a user input without a `UserInputRef`, an operation config for an undeclared or unstored op, a credential for an undeclared slot.
2. On a credential write, `selectConnection` fixes `next.Connection`: the current connection, which must declare the slot, else the connection the slot selects.
3. For each document in `def.documents`: `stage` first; if it staged, `conform` is skipped. A `stage` error aborts the whole call with that error. Otherwise `conform`; a `conform` error becomes `doc.failure(err, selected.CredentialRefs)` and the document keeps its prior value in `next`.

**Retired names.** `CredentialRef.conform` and `OperationRef.source` walk `Name()` then `Replaces()` and take the first stored entry (ii/types/ref.go). Every write onto the current name, conformed or staged, goes through the handle's `store`, which removes the entries under all of its `Replaces()` names from `next`: a value found only under a retired name moves onto the current name, and when the current name already holds a value that value wins and the retired entries are dropped. `persist` then deletes the hush rows no longer bound, so a pass interrupted between writing the current slot and deleting the retired one converges on the next pass.

**Order.** `def.documents` is in declaration order; `Define` imposes no grouping. Order does not matter: every `conform` reads only `src.Prior`, `src.Legacy` and `next.Connection`, and `next.Connection` is fixed before the loop. `InstallationRef.conform` stamps `Display.CredentialRef` from `next.Connection`.

**Persisting.** Writing state back to the installation is one runtime function, `persist` in ii/runtime/load.go. It is the single place that knows the storage columns, and it doesn't dispatch on anything.

**Erased runtime views.** `OperationRegistration` (ii/types/operation.go) keeps `Name`, `Description` and `Schema` as plain fields; `Schema` is serialised under the JSON key `configSchema`. It gains two closures built in `OperationRef.declare`:
- `Settings func(InstallationState) (Configurable, bool, error)`, nil for unstored ops,
- `Config func(InstallationState, json.RawMessage) (json.RawMessage, error)`, which returns the override when given, else the stored config, else `{}`.

These serve every runtime site in S5 without the generic handle. `OperationRegistration.UISchema`, `ClientRegistration.ConfigSchema` and `DisconnectRegistration.Schema` no longer exist.

**`Definition`** (ii/types/definition.go) has unexported fields:
- `documents` — every stored document, in declaration order,
- typed views written by the declaring handle: `credentials []CredentialRegistration` (once per slot), `userInput *userInputDeclaration` (catalog view plus a `decode` closure), `installation *InstallationRegistration`,
- `connections`, `clients`, `operations`, `webhooks`, `mappings`, `listeners`,
- `healthCheck`, `operator`, `runtime`.

Accessors:
- `UserInput()` (the `CatalogUserInput` view), `Installation()`, `Credentials()`, `CredentialRegistration(ref)`, `Connections()`, `Connection(slot)`, `Clients()`, `Client(id)`, `Operations()`, `Operation(name)`, `Webhooks()`, `Webhook(name)`, `ResolveWebhook(name)`, `Mappings()`, `Listeners()`, `HealthCheck()`, `Runtime()`, `Catalog()`, `ProviderState(state)`, `WithProviderState(state, next)`.
- There is no `Documents()` accessor. `UserInput()` and `Installation()` read the typed views; nothing asserts over `documents`.
- `Credentials()` sets `FormSchema` to the stored schema for every slot that no connection's `Auth` writes to, so credential forms come from the declared type.
- `Catalog()` (ii/types/catalog.go) is the provider-listing response. Its JSON is designed fresh, with no compatibility constraint. It lists customer-selectable operations only and sets `PrimaryDirectory` (§3.2).

### 2.6 Handles

All handles are in ii/types/ref.go, except `Connection` (ii/types/connection.go). The `OperationRef` fluent setters `Policy`, `Ingest`, `Permissions`, `Schedule`, `SkipDefaultLookback`, `RateLimit`, `Internal`, `CustomerSelectable`, `RequiresPaymentMethod`, `DisabledForAll` and the binders `Ingests`/`Handles`/`HandlesRequest` are kept. The full surface:

```go
// every handle and Connection
func (h X) Describe(text string) X

// CredentialRef — Input[T]; declaring it directly fails with ErrCredentialOutsideConnection
func CredentialOf[T any]() CredentialRef[T]
func NewCredential[T any](name string) CredentialRef[T]
func (r CredentialRef[T]) ID() CredentialSlotID
func (r CredentialRef[T]) Resolve(bindings CredentialBindings) (T, bool, error)
func (r CredentialRef[T]) Titled(title string) CredentialRef[T]
func (r CredentialRef[T]) Recommended() CredentialRef[T]
func (r CredentialRef[T]) Replacing(names ...string) CredentialRef[T]

// Connection — declares its credentials; the first selects the mode
func ConnectionOf[T any](cred CredentialRef[T]) Connection
func (c Connection) Named(name string) Connection
func (c Connection) Using[T any](cred CredentialRef[T]) Connection
func (c Connection) Meta(meta map[string]MetaInfo) Connection
func (c Connection) Auth(start AuthStartFunc, complete AuthCompleteFunc) Connection
func (c Connection) Disconnect(description string, fn DisconnectFunc) Connection // nil fn only describes the teardown

// UserInputRef — Input[T]; at most one per definition
func UserInputOf[T any]() UserInputRef[T]
func (r UserInputRef[T]) Resolve(installation *generated.Integration) (T, bool, error) // nil-safe; false when none stored

// OperationRef — Input[Config]; stored when *Config implements Configurable
func OperationOf[Config any]() OperationRef[Config]
func NewOperation[Config any](name string) OperationRef[Config]
func (r OperationRef[Config]) Probe(fn OperationHandler) OperationRef[Config]
func (r OperationRef[Config]) Replacing(names ...string) OperationRef[Config]
func (r OperationRef[Config]) IsCustomerSelectable() bool
func (r OperationRef[Config]) Config(state InstallationState, override json.RawMessage) (Config, error)

// InstallationRef — Input[T Identifier]
func InstallationOf[T Identifier](resolve func(context.Context, InstallationRequest) (T, bool, error)) InstallationRef[T]
func (r InstallationRef[T]) Resolve(installation *generated.Integration) (T, bool, error) // nil-safe; decodes Attributes

// ClientRef — declared once, referenced by operations and health checks
func NewClientRef[C any](name string) ClientRef[C]
func ClientRefOf[C any]() ClientRef[C]
func (r ClientRef[C]) ID() ClientID
func (r ClientRef[C]) Cast(client any) (C, error)
func (r ClientRef[C]) Builds(build func(context.Context, ClientBuildRequest) (C, error)) ClientRef[C]
func (r ClientRef[C]) Using[T any](cred CredentialRef[T]) ClientRef[C]

// OperatorConfigRef — Input[T]; registers OperatorConfigRegistration{Name, Layout, Description, Schema}; the value arrives as a Builder parameter from catalog.Config
func OperatorConfigOf[T any]() OperatorConfigRef[T]

// RuntimeConfigRef — Input[T] plus the provisioned value, marshalled into RuntimeIntegrationRegistration.Config at declare
func RuntimeConfigOf[T any](value *T, build func(context.Context, T) (any, error)) RuntimeConfigRef[T]

// WebhookRef — Identity + contract
func NewWebhookRef(name string) WebhookRef
// Replacing, Route, EndpointTemplate, Secret, Resolves, Verifies, Routes, Events: fluent; Routes is optional for webhooks served by their own handler (scim)

// WebhookEventRef — typed handler
func NewWebhookEventRef[T any](name string) WebhookEventRef[T]
func (r WebhookEventRef[T]) Name() string
func (r WebhookEventRef[T]) Handles(fn func(context.Context, WebhookHandleRequest, T) error) WebhookEventRef[T]
func (r WebhookEventRef[T]) Ingest(contracts ...IngestContract) WebhookEventRef[T]

// remaining (ii/types/declare.go)
func HealthCheck[C any](client ClientRef[C], fn func(context.Context, OperationRequest, C) (json.RawMessage, error)) Declaration
func CredentialHealthCheck(fn func(context.Context, OperationRequest) (json.RawMessage, error)) Declaration
func Listener(registration GalaListenerRegistration) Declaration
func Mappings(mappings ...MappingRegistration) Declaration
```

`UserInputRef.Resolve` and `InstallationRef.Resolve` read straight off the ent record (`installation.UserInput.Data`, `installation.InstallationMetadata.Attributes`) and return the zero `T` with `false` for a nil record or an empty document.

**Validated by `Define`** (ii/types/declare.go and each handle's `declare`). `Define` is the single validator of a definition's internal consistency; the registry does not repeat these checks.

At declare time:
- `ClientRef`: an invalid (empty) client ID fails with `ErrClientIDRequired`; a client without `Builds` fails with `ErrClientBuildRequired`; a second declaration of the same client fails with `ErrDeclaredTwice`.
- `Connection`: a zero selecting slot fails with `ErrConnectionCredentialRequired`; a second connection on the same selecting slot fails with `ErrDuplicateIdentity`. Credentials are registered once per slot; the same slot with a different layout fails with `ErrDuplicateIdentity`.
- `CredentialRef` declared outside a connection fails with `ErrCredentialOutsideConnection`.
- User input, installation metadata, operator config, runtime config and health check may each appear once (`ErrDeclaredTwice`).
- `RuntimeConfigRef`: a nil build function fails with `ErrRuntimeBuildRequired`, mirroring `ClientRef`'s `ErrClientBuildRequired` (ii/types/ref.go, ii/types/errors.go).
- Duplicate operation names, webhook names, and event names within a webhook fail with `ErrDuplicateIdentity`.

After every declaration:
- A client with no `Using` is given every declared credential slot.
- `Replaces` on credentials, operations and webhooks never names a live identity (`ErrReplacedNameLive`).
- Every client an operation references is declared (`ErrClientNotDeclared`).
- Every credential slot a client uses is declared by a connection (`ErrClientCredentialNotDeclared`).
- A runtime config never shares its layout with the operator config (`ErrRuntimeConfigIsOperatorConfig`, K26).
- Health check: required when connections are declared; must have a handler; its client must be declared and must use every connection's selecting slot.

`Define` dedupes by identity, not by function equality: Go funcs are not comparable, which is why the client must be declared once explicitly.

**Handles depending on builder parameters** (`cfg`, `devMode`) are constructed inside the builder function, as they are today. Package-level handles stay package-level. Conditional declarations, such as slack's runtime config in dev mode, are appended conditionally to the `decls` slice.

### 2.7 Registration and operator config

The catalog stays as it is:
- the builder list in defs/catalog/catalog.go,
- `catalog.Config` with one typed field per definition config (defs/catalog/types.go), wired in config/config.go.

Builders keep receiving their typed config as parameters.

So adding an integration touches its own package plus one line in `Builders`, and one field in `catalog.Config` if it has operator or runtime config. Inside the definition package, everything is declared once (§2.6).

Moving operator config to a keyed map was rejected: it would trade the typed `catalog.Config` struct for removing those two lines.

**Registry** (ii/registry/registry.go). `Register` checks only what `Define` cannot see: a non-empty definition ID not already registered, mapping link rules, operation handler shape (`indexOperations`), webhook resolver and handler presence (`indexWebhooks`), and operation and webhook-event topic uniqueness across definitions (`indexUnique`). It no longer re-checks duplicate identities inside a definition. With `WithSnapshots` it also requires the committed snapshot (§5.4).

K29 (googledrive/onedrive selection) is a capability query:
1. A definition's user input embeds `providerkit.DocumentSource`, which implements `providerkit.DocumentPreference` (ii/providerkit/preferences.go).
2. `findPrimaryDriveIntegration` (internal/graphapi/internalpolicyextended_helpers.go) selects registered definitions with `types.UserInputAs[providerkit.DocumentPreference](def, types.InstallationState{})`, a type-only check because the empty state decodes to the zero value.
3. It queries connected installations of those definition IDs and reads each installation's preference with the same call over its `UserInput`.

### 2.8 Examples (inside builder functions)

**Okta: credential, metadata, stored op with a registered upgrade.**
Main's okta was flat: `{filterExpr, search, enableGroupSync, primaryDirectory}` (M:defs/okta/types.go:24-33). Main collected groups only when `enableGroupSync` was true (M:defs/okta/operation_directory_sync.go:39-43, 80-82).

The okta handles are package-level (defs/okta/types.go) because none depends on builder parameters.

```go
var (
	installation   = types.InstallationOf(resolveInstallationMetadata)
	oktaCredential = types.CredentialOf[CredentialSchema]()
	oktaClient     = types.ClientRefOf[*oktagosdk.APIClient]()
)

// DirectorySync configures collection of Okta directory users, groups, and memberships
type DirectorySync struct {
	providerkit.DirectorySync
	// Search is an optional Okta search expression applied server-side when listing users
	Search string `json:"search,omitempty" jsonschema:"title=User Search Expression"`
}

// UpgradeFrom maps Okta's former enableGroupSync onto disableGroupSync, keeping groups off for installations that never enabled them
func (d *DirectorySync) UpgradeFrom(_ context.Context, _ types.InstallationRequest, from string, stored json.RawMessage) error {
	if from != "" {
		return nil
	}

	enabled, _ := jsonx.DecodeObjectKey[bool](stored, "enableGroupSync")
	d.DisableGroupSync = !enabled

	return nil
}

func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Define(types.DefinitionSpec{ /* ... */ },
			types.ConnectionOf(oktaCredential.Titled("Okta Credential").Describe("...")).
				Named("Okta API Token").
				Describe("...").
				Disconnect("Removes the stored API token from Openlane. ...", nil),
			oktaClient.Builds(Client{}.Build).Describe("Okta API client"),
			installation,
			providerkit.DirectorySyncOperation[DirectorySync](oktaClient, runDirectorySync).Describe("..."),
			types.Mappings(providerkit.DirectoryMappings(mapExprDirectoryAccount, mapExprDirectoryGroup, mapExprDirectoryMembership)...),
			types.HealthCheck(oktaClient, checkHealth),
		)
	})
}
```

The `from != ""` early return is the one-time backfill guard described in §5.5.

**githubapp and cloudflare, the two keys that don't match.**

```go
// UpgradeFrom reads the settings main stored under the findingSync section
func (v *VulnerabilitySync) UpgradeFrom(_ context.Context, _ types.InstallationRequest, from string, stored json.RawMessage) error {
	if from != "" {
		return nil
	}

	section, ok := jsonx.DecodeObjectKey[json.RawMessage](stored, "findingSync")
	if !ok {
		return nil
	}

	return jsonx.UnmarshalIfPresent(section, v)
}
```

This is defs/githubapp/types.go. Cloudflare `FindingsSync` (defs/cloudflare/types.go) has the same body. The two fake `Replacing(types.NewOperationPayload[...]("findingSync"))` declarations are deleted. githubapp `MaxRepos` is back on main's tag `max_repos`, so persisted per-run configs keep decoding.

**Shared directory sync config.** cloudflare and googleworkspace embed `providerkit.DirectorySync` and honour `DisableGroupSync` in their directory sync runs (defs/cloudflare/operation_directory_sync.go, defs/googleworkspace/operation_directory_sync.go). slack, zitadel and scim keep local `DirectorySync` configs that embed only `providerkit.OperationSettings` (defs/slack/types.go, defs/zitadel/types.go, defs/scim/operation_directory_sync.go).

**githubapp webhook** (inside `Builder`, because `cfg` and `app` are builder locals, defs/githubapp/builder.go):

The identities are package-level in defs/githubapp/types.go (`InstallationEventsWebhook = types.NewWebhookRef("installation.events")`, `pingWebhookEvent = types.NewWebhookEventRef[githubWebhookEnvelope]("ping")`, ...); the builder binds handlers and the builder-local contract:

```go
webhook := InstallationEventsWebhook.
	Route("/github/app/webhook").
	Secret(func() string { return cfg.WebhookSecret }).
	Resolves(ResolveWebhookIntegration).
	Verifies(app.Verify).
	Routes(app.Event).
	Events(
		pingWebhookEvent.Handles(PingWebhook{}.Handle),
		dependabotAlertWebhookEvent.
			Ingest(types.IngestContract{Schema: entityops.SchemaVulnerability.Name}).
			Handles(DependabotAlertWebhook{}.Handle),
		/* ... */
	)
```

`app.Event` keeps its GitHub header switch and returns handle names (defs/githubapp/webhook.go).

**Slack: user input with a shared preference.** slack and microsoftteams embed `providerkit.Messaging`; googledrive and onedrive embed `providerkit.DocumentSource`.

```go
type UserInput struct {
	providerkit.Messaging
}
```

The workflow engine reads the preference with `types.UserInputAs[providerkit.MessagingPreference]` (internal/workflows/engine/integration_executor.go).

**Email.** The email definition declares its operations from its `Dispatcher` entries (`d.Declaration()`, defs/email/builder.go). Callers read catalog facts through `Dispatcher` accessors (`Name`, `OperationDescription`, `Schema`, `IsCustomerSelectable`, `ExamplePayload`, `BuilderUISchema`; defs/email/emailop.go) rather than through erased operation registrations.

**Microsoft Teams.** The teams definition declares no hand-written credential form schema; its slot is written by `Auth`, and every other slot's form comes from `Definition.Credentials()` (§2.5).

---

## Part 3 — Storage and shared definitions

### 3.1 End state

| Field | Type | Handle | vs main | vs branch |
|---|---|---|---|---|
| `user_input` | `IntegrationUserInput{Revision}` | `UserInputRef[T]` | new column | replaces `user_inputs` |
| `operation_config` | `IntegrationOperationConfigs` | `OperationRef[C]` | new column | replaces `operation_inputs` |
| `installation_metadata` | `{Layout, Attributes, Display}` | `InstallationRef[T]` | adds `layout` | back to `attributes` |
| `provider_state` | `IntegrationProviderState` | connection selection | unchanged | replaces `credential_ref` |
| hush `credential_set` | `CredentialSet{Revision}` | `CredentialRef[T]` | adds `layout` | same JSON |
| `config` | `IntegrationConfig` | none (frozen source) | no longer written; dropped in Phase 4 | — |
| `user_inputs`, `operation_inputs`, `credential_ref` | — | — | — | removed |

**Rollback compatibility.**
- `layout` is ignored by main's decoders: main's `CredentialSet` and `IntegrationInstallationMetadata` have no such field.
- Main reads `attributes`, `provider_state` and `config` unchanged.

### 3.2 Shared concepts

| Concept | Home | Replaces |
|---|---|---|
| `OperationSettings` | providerkit | D4, K17, K24 |
| `DirectorySync{OperationSettings; DisableGroupSync}` | providerkit | D1, D2; D3 via okta `UpgradeFrom` |
| `DirectorySyncOperation[C, Cl](client, fn) types.OperationRef[C]` — preconfigured `Ingests`, `Policy(Reconcile, Snapshot)`, `Ingest(DirectoryIngestContracts()...)`, `SkipDefaultLookback`, still fluent | providerkit | D9 |
| `Messaging{DefaultMessaging}` + `MessagingPreference` | providerkit | D7, S6 |
| `DocumentSource{Primary}` + `DocumentPreference` | providerkit | D8, S7, K29 |
| `ExprPrimaryDirectory` | providerkit | D6 |
| primary directory | `primary_directory` column, set by the user from a top-level `primaryDirectory` on configure and auth-start (internal/httpserve/handlers/integration_config.go, integration_flow.go), which reject it when `def.Catalog().PrimaryDirectory` is false. `Catalog.PrimaryDirectory` is true when any operation declares an ingest contract for `entityops.SchemaDirectoryAccount` (ii/types/catalog.go). That set is the 12 definitions whose mappings use `providerkit.ExprPrimaryDirectory`: authentik, awssecurityhub, azureentraid, cloudflare, githubapp, googleworkspace, keycloak, okta, scim, slack, tailscale, zitadel. | D5, S2 |
| installation name | scim's `UserInput` keeps its required `Name` field (defs/scim/types.go), because one org can run several SCIM installations. `persist` (ii/runtime/load.go) mirrors a non-empty top-level `name` from a submitted user input write to the `name` column. | — |

Scim keeps its own op (`HandlesRequest`, `Inline`, defs/scim/operation_directory_sync.go). Its config embeds `providerkit.OperationSettings`. Its op declares `DirectoryIngestContracts()`, so the primary-directory toggle is offered.

---

## Part 4 — Disposition of every inventory item

| Item | Replacement |
|---|---|
| K1–K5 | removed; `Define` collects documents; each handle looks up its own `Name()` then `Replaces()` |
| K6 | `types.Writes` → `types.Conform(..., writes)`; each document's `stage` picks up the write addressed to it |
| K7, K9 | `types.Conform` iterates documents; the section inference lives only in `OperationRef.source` (§5.5) |
| K8 | each handle wraps its own sentinel (`ErrCredentialInvalid`, `ErrUserInputInvalid`, `ErrOperationConfigInvalid`, `ErrInstallationMetadataInvalid`) and returns its own `Failure` implementation from `failure` |
| K10 | `Sources.Legacy` (§5.5) |
| K11 | removed |
| K12, K13 | keystore `Load`/`Save` on `CredentialBindings` |
| K14, K15 | `types.SurfaceOf` |
| K16 | removed (`finalize.go` deleted) |
| K17 | `OperationSettings.Validate` |
| K18–K21 | `types.Writes`, `def.UserInput()`, `types.UserInputAs` |
| K22 | `types.Form` |
| K23 | main's typed provider-state helpers restored |
| K24 | `OperationRegistration.Settings` → `Configurable.Disabled()` |
| K25 | `OperationOf`/`NewOperation`; storedness from `Configurable` |
| K26 | `RuntimeConfigRef` alone |
| K27 | kept by decision (§2.7) |
| K28 | kept as `metadataMirror` in ii/runtime/load.go, applied by `persist`; `Display.CredentialRef` is set from the connection in `InstallationRef.conform` and `deriveMetadata` |
| K29 | `DocumentPreference` capability query |
| D1–D9 | §3.2 |
| D10 | kept; scim's required `Name` |
| S1 | kept; the `name` key of a submitted user input is mirrored to the column |
| S2 | §3.2 |
| S3 | kept, confined to `OperationRef.source`, only for never-conformed installations (§5.5) |
| S4 | removed |
| S5 | `OperationRegistration.Settings`/`Config` |
| S6, S7 | `UserInputAs` |
| S8 | kept. It queries the platform struct `IntegrationInstallationIdentity.ExternalID` (common/openapi/integration_models.go), not definition data. |
| S9 | kept; renamed via `Replaces` |

---

## Part 5 — Upgrade/conform path

### 5.1 Main's data needs no SQL

| Data on a main row | Read by this design as |
|---|---|
| `config.clientConfig` | `Sources.Legacy`, the user input and op source while `definition_version == ""` |
| `installation_metadata` | `{Layout: "", Attributes, Display}` |
| hush rows | `CredentialSet{Layout: ""}` |
| `provider_state` | main's typed helper |

Main never writes `definition_version`: `git grep -nE 'SetDefinition(Version|Slug)' main -- internal ':!internal/ent/generated'` hits only a test. So every main row is a never-conformed installation.

### 5.2 Conform, step by step

The entry point is `r.load(ctx, installation, writes, input)` in ii/runtime/load.go, which delegates to `conformInstallation(ctx, installation, writes, input, false)`. `conformInstallation` is the only reader of stored integration data in the runtime; its last parameter, `rederive`, is set only by `RefreshInstallationMetadata` (see below).

**Step 1 — read.**
- `prior` is `storedState(installation)` (`UserInput`, `OperationConfig`, `InstallationMetadata`) plus:
  - `Credentials` from `keystore.Load(installation, names)`, where `names` is every declared slot and each slot's `Replaces` names; rows under any other `secret_name` are never returned (internal/keystore/store.go),
  - `Connection` from `storedSlot`: `def.ProviderState(installation.ProviderState).CredentialRef`, moved onto the declared slot that replaces it when it names a retired slot.
- `Sources.Legacy` = `installation.Config.ClientConfig` when `installation.DefinitionVersion == ""`, else nil (§5.5).

**Step 2 — gate.** `staged` is true when any write is present or `rederive` is set. Versions are compared by `versionOrder(stored, current) int` (`strings.Compare`), the single comparator used by the gate in `conformInstallation` and by the startup sweep's `staleInstallations` (ii/runtime/upgrade_sweep.go).

| Stored version vs this binary's | Action |
|---|---|
| equal, nothing staged | return `prior` |
| older, or empty (every main row) | continue to step 3 |
| equal, something staged | continue to step 3 |
| newer, nothing staged | return `prior` unconformed; nothing is written |
| newer, something staged | return `ErrInstallationVersionNewer`; nothing is written |

**Step 3 — `types.Conform`** (§2.5). Per document:
- **Credential, per declared slot.** Takes the binding under its name, or under a `Replaces` name, and runs `Input.Conform(layout)`. A failure on a slot of the selected connection aborts; any other slot's failure leaves that slot unchanged.
- **UserInputRef.** Takes `prior.UserInput` when non-empty; otherwise the whole `Legacy` document with `from = ""` (§5.5); otherwise nothing. Then `Input.Conform`.
- **OperationRef, stored ops only.** Takes the prior config under its name or a `Replaces` name; otherwise the `Legacy` source (§5.5); otherwise the op stays absent and reads as the zero `Config`.
- **InstallationRef.** Conforms non-empty `prior.Metadata.Attributes` through `Input.Conform`. It never calls `resolve`. It sets `Display` from `T.InstallationIdentity()`, plus `CredentialRef` from `next.Connection`.
- **Failures.** A failed document keeps its prior value in `next` and is reported as a `Failure`.

For a definition with no connections, `load` sets `Display.ExternalID` to the installation ID.

**Step 4 — abort or derive.**
- Every failure is logged. If any `Failure.Aborts()` (user input, or a credential of the selected connection), `failInstallation` applies only the aborting failures through the health marker and returns the error wrapped by `types.Unhealthy`, or unwrapped when the status is pending or disabled. Nothing is persisted, so `definition_version` keeps its old value.
- On a credential write, `checkCredential` runs the health check over the selected connection's bindings and computes the external ID the new metadata must match.
- When `writes.Credential != nil || rederive`, `deriveMetadata` calls the installation's `resolve` (or self-identifies when there are no connections). A credential write fails with `ErrInstallationInstanceMismatch` when the external ID changes.

**Step 5 — persist, in one transaction** (`persist`).
- `keystore.Save` the conformed bindings, and `keystore.Delete` the prior rows whose slot is no longer bound (the ones moved off a `Replaces` name). Undeclared hush rows are never loaded and never touched; users can link arbitrary secrets via `CreateHushInput.integrationIDs`.
- Rename `integration_run.operation_name` and `integration_webhook.name` from `Replaces` names to the live name.
- Update `user_input`, `operation_config`, `installation_metadata`, `metadata` (`metadataMirror`), `health` (`retiredHealth` moves marks off retired op names and drops marks for undeclared ops), `definition_version`, plus `provider_state` when the connection changed and `name` when a submitted user input carries one.
- `definition_version` is set to this binary's version in the same transaction as every conformed field. An upgrade is never persisted without its version, and the version never moves without the fields.
- A non-aborting failure writes the failed document's prior value back unchanged.
- `config` is never written.

After the transaction, when the version changed, `finishUpgrade` purges reconcile loops under retired op names, reconciles the installation's webhooks, and resets reconcile loops if any op was renamed.

**Step 6 — outcomes** (`conformOutcomes`). The runtime's `conformMarker` implements `types.HealthMarker`; every non-aborting failure is applied through `Failure.Apply`:
- `MarkOperation` calls `MarkOperationUnhealthy` with a reason prefixed by `ErrInstallationUpgradeFailed`. Reconcile skips the op because it already checks `UnhealthyOperations`.
- `MarkInstallation` calls `MarkIntegrationUnhealthy`, except for pending or disabled installations.
- Marks with that prefix are cleared when the document now conforms: `ClearOperationUnhealthy` for each op not marked in this pass, and `ClearIntegrationUnhealthy` when the installation is errored with that prefix.
- Example: googledrive rows without `folderId` (required by the FolderSync config) degrade only FolderSync. That matches main, which failed only that run (M:defs/googledrive/operation_folder_sync.go:38-42).

**Step 7 — steady state.**
- Step 2 short-circuits until the definition changes.
- After a change, every document conforms again. The backfill `UpgradeFrom` bodies return early because `from` is non-empty, and the encodings are byte-identical, so only `definition_version` changes.
- `Legacy` is nil once `definition_version` is set, so inference never repeats.

**Disconnect** (ii/runtime/credentials.go) calls `load` and logs and proceeds when it fails, then reads the connection from `provider_state` through `storedSlot`.

**`RefreshInstallationMetadata`** (ii/runtime/credentials.go) calls `conformInstallation(ctx, installation, types.Writes{}, nil, true)`. `rederive` counts as staged and triggers `deriveMetadata` inside the conform-and-persist path, so the newer-version check is the single `versionOrder` gate in step 2.

### 5.3 Ongoing evolution

| Change | Author action | Mechanism |
|---|---|---|
| Add a field | add it | version changes; conform pass; `ConformToSchema` defaults (omitempty + default only) |
| Add a field populated from data we didn't have | add it; `UpgradeFrom` fills it when empty from `req` | runs once per installation on the next pass |
| Remove a field | delete it | decode ignores it, conform strips it |
| Rename a JSON tag | change the tag; `UpgradeFrom` copies the old key when present | `stored` keeps the old bytes |
| Rename an operation | `NewOperation[C]("New").Replacing("Old")` | §5.2 step 5 moves config, runs, health and webhooks; `finishUpgrade` moves loops |
| Rename the Go type (escape hatch) | keep the identity with `NewOperation[C]("Old")`; branch on `from == "OldType"` only if shapes differ | the `layout` stamp |

### 5.4 Versions and rolling deploys

**Before.** The version was a hash of the definition surface, compared by equality. A hash says whether two versions differ, not which is newer.

**Now: a ULID per committed snapshot.**
- `registry.Snapshot` (ii/registry/snapshot.go) is `{hash, version, surface}`: `Hash` is `SurfaceHash(surface)` (ii/registry/version.go), `Version` is a ULID, and `Surface` is `types.SurfaceOf(def)`.
- The snapshot generator (jsonschema/integration_schema_generator.go, run by `task config:generate` through the jsonschema Taskfile's `integration-schema` task as `go run -tags test jsonschema/integration_schema_generator.go`) registers `catalog.Builders(catalog.Config{}, "", false)`, then per definition:
  - runs `GateSurfaceChange(committed, surface)` only when the committed snapshot's `Hash != ""`, so a missing snapshot or one in a pre-hash shape is written ungated,
  - keeps the committed `Version` while the hash is unchanged, and mints `ulid.Make()` when it differs,
  - writes the snapshot to internal/integrations/registry/surfaces/<definition id>.json.
- The same run then writes the test fixture snapshots, one registry per fixture, to internal/testutils/integrations/surfaces/<fixture>/surfaces/<definition id>.json for `base`, `mockhttp`, `v1`, `v2`, `v3`, with the same gate and version rule. `v1`, `v2` and `v3` are marked `versioned: true` and are written in that order; once one of them mints a version, every later one re-mints too, so their ULIDs ascend.
- ULID strings are a millisecond timestamp in Crockford base32 with an uppercase alphabet, so later versions compare greater as plain strings, and `""` (main rows) compares less than any ULID.
- The snapshots are embedded as `registry.Surfaces` (`//go:embed surfaces/*.json`).
- `WithSnapshots(fsys)` makes `Register` require `surfaces/<id>.json` in `fsys` and its `Hash` to equal the definition's current hash (`ErrSnapshotStale` otherwise), and records its `Version`; `Registry.Version(id)` returns it. Without `WithSnapshots` the version is `""`.
- The runtime (ii/runtime/runtime.go) passes `WithSnapshots(registry.Surfaces)` only when it builds the registry from the catalog builders, i.e. when `Config` supplies neither a `Registry` nor `DefinitionBuilders`. A binary built from the catalog therefore cannot run a changed definition under a stale version.
- Test definitions read their fixture snapshots through package `surfaces` (internal/testutils/integrations/surfaces, build tag `test`), which embeds `{base,mockhttp,v1,v2,v3}/surfaces/*.json` and exposes one filesystem per fixture (`Base`, `MockHTTP`, `V1`, `V2`, `V3`) for registries built with `WithSnapshots`.
- New installations are created with `definition_version = Registry().Version(def.ID)` (ii/runtime/integration.go).
- `Registry.Fingerprint()` hashes the sorted versions and keys the startup seed job.

**Effect.**
- An old pod sees a newer stored ULID and does not write.
- A new pod sees an older ULID, or an empty one, and upgrades.
- No history list is needed.

**Re-baselining.** The 21 committed snapshots in internal/integrations/registry/surfaces are still in the pre-hash shape (no `hash`, `version` holding a hex hash, credentials keyed `ref`). Because their `Hash` is empty, the generator rewrites them ungated; until it is run, `Register` with `WithSnapshots` fails with `ErrSnapshotStale`. The fixture snapshot files under internal/testutils/integrations/surfaces do not exist yet, so package `surfaces` does not compile until the generator writes them. Snapshot generation has not been run; it is user-run. Branch-local rows holding lowercase hex hashes do not order meaningfully against ULIDs, so branch databases need `definition_version` cleared once. This does not affect production: main never wrote the field.

**Call sites.** `load` replaces the 8 `ensureCurrentVersion` call sites and also covers `BuildClientForIntegration` and `EnsureWebhook`. The startup sweep (`UpgradeInstallations`, called from internal/httpserve/serveropts/integration_base.go) loads every installation in a pending, connected, degraded, errored or disabled status whose version orders before the binary's. `Reconcile` is `load` with writes. `Accept` passes the current layout as `from`, so the guarded `UpgradeFrom` bodies never touch submissions.

### 5.5 One-time backfill from main

This is the only intermediary construct in the design. Everything else reads and writes the end-state storage of §3.1.

**What it is.** Three pieces, all reachable only for an installation whose `definition_version` is `""`:
1. **`Sources.Legacy`** (ii/types/conform.go). `load` sets it to `config.clientConfig` while `installation.DefinitionVersion == ""`, and to nil otherwise (ii/runtime/load.go). Two document methods read it:
   - `UserInputRef.conform` takes the whole `Legacy` document, with `from = ""`, when no user input is stored.
   - `OperationRef.source`, for stored ops with no stored config under the name or a `Replaces` name, infers the section: `Legacy[lo.CamelCase(name)]` when that key holds a non-empty object (`legacySection`), else the whole `Legacy` document (an empty `{}` included), with `from = ""` (ii/types/ref.go). The plain decode then carries every field whose tag matches.
2. **The `from == ""` guards** in exactly three `UpgradeFrom` bodies, each of which returns immediately when `from != ""`:
   - githubapp `VulnerabilitySync` (defs/githubapp/types.go) and cloudflare `FindingsSync` (defs/cloudflare/types.go) decode main's `findingSync` section, which `lo.CamelCase` of their op names does not reach.
   - okta `DirectorySync` (defs/okta/types.go) sets `DisableGroupSync = !enableGroupSync`, including on empty configs, so installations that never enabled groups keep them off.
3. **The `config` column**, which nothing writes and only step 1 reads.

**When it stops being reachable.** `Legacy` is non-nil only for a row with `definition_version == ""`, and a successful `load` persists the binary's version in the same transaction as the conformed fields. So the backfill is unreachable once every row has been conformed:
- the startup sweep loads every installation in all five statuses (pending, connected, degraded, errored, disabled) whose version orders before the binary's, which includes every `""` row present at startup;
- rows that old (main) pods create during the rollout carry `""` and can appear after the sweep has run; any `load` call site conforms them lazily on first access, and the next startup sweep catches the rest;
- installations created by new pods carry a version from creation and never read `Legacy`;
- a row whose conform aborts (user input, or a credential of the selected connection) is not persisted and stays at `""`, so it remains a `Legacy` reader until it conforms.

**Removal.** Phase 4 removes it together with the `config` column: `Sources.Legacy`, the legacy branches of `UserInputRef.conform` and `OperationRef.source` with `legacySection`, the three `UpgradeFrom` methods (whose bodies act only when `from == ""`), and `IntegrationConfig`.

---

## Part 6 — Validation

**Writes:** `Input.Accept` runs strict schema validation, so unknown properties are rejected. Then it runs `Conform` with the current layout.

**Stored documents:** `Input.Conform`, whose steps are listed in §2.3.

**Definition-specific validation:** a `Validate` method on the stored type. Shared validation arrives by embedding. A type that shadows `Validate` calls the embedded one; nothing checks this at registration (§2.4).

**Sentinels** (ii/types/errors.go): `ErrCredentialInvalid`, `ErrUserInputInvalid`, `ErrOperationConfigInvalid`, `ErrInstallationMetadataInvalid`, plus the `Define` sentinels listed in §2.6.

**Per-run dispatch config:** validated by the operation's `Schema` through `operations.ValidateInput` at dispatch (ii/runtime/dispatch.go) and inline execute (ii/runtime/execution.go), unchanged.

**Definition consistency:** `Define` only (§2.6).

**Registry:**
- Keeps definition ID checks, handler-shape checks, topic uniqueness across definitions, link-rule validation and the runtime client build (§2.7).
- With `WithSnapshots`, requires the committed snapshot hash to match (§5.4).
- `GateSurfaceChange` (ii/registry/surface_diff.go) refuses a committed credential, operation or webhook identity that the next surface neither declares nor replaces, and a committed user input or installation the next surface no longer declares. Its only caller is the snapshot generator, which calls it only when the committed `Hash` is non-empty and mints a new ULID version when the hash changes.

---

## Part 7 — Phase plan and generation boundaries

### How generation runs here

- **`task generate`** runs onboarding, core (ent then gqlgen), catalog, fga, then openapi (Taskfile.yaml:81-102).
- **`entc` type-checks the `codegen`-tagged closure of `internal/ent/schema`.** That closure includes all integrations packages, all definitions plus catalog, runtime, keystore and workflows. `go list -tags=codegen -deps ./internal/ent/schema | grep -c 'integrations/definitions/'` gives 23.
- **Migrations are produced by the user's generation runs.** They are not planned steps of this design; no phase below creates, edits or deletes a migration file.
- **The surface snapshots come from `task config:generate`** (jsonschema/integration_schema_generator.go, §5.4).
- **entx `GenQuery` prunes removed fields from `internal/graphapi/query/*.graphql` during the ent step.**
- **The rule:** before every `task generate`, the whole non-test module compiles against the generated code of that moment.
- All generation is user-run.

### Status

Phases 1 and 2 are present in the working tree. The revision 3 precondition (restoring `inputSchema`/`inputSchemas` in internal/graphapi/integrationextended.resolvers.go) is obsolete: the resolvers call `types.Form`. `RefreshInstallationMetadata` is routed through `conformInstallation` (§5.2). Phase 3's generated-field removal and the `StoredDocument`/`StoredDocuments` deletion are done; its remaining item is the user-run snapshot generation.

### Phase 1 — common and additive schema (before regen #1)

**common:**
- add `models.Revision`
- `CredentialSet` embeds it (same JSON)
- add `IntegrationUserInput`, `IntegrationOperationConfig(s)`
- reshape `IntegrationInstallationMetadata` to `{Layout, Attributes, Display}`

Generated code reaches the metadata only through `.Display.ExternalID`, so no generated signature changes.

**Hand-written metadata users that change in this phase:**
- B:internal/httpserve/handlers/integration_config.go:95
- B:internal/graphapi/integrationextended.resolvers.go:70
- B:ii/types/ref.go:599 and :617. Decode `Attributes` directly; `Input.decode` looks up by `Name` (B:ii/types/input.go:105), which the new shape lacks.
- B:defs/githubapp/installation.go:37
- tests:
  - B:internal/graphapi/eventstest/integrations_lift_test.go:88-90, 131-135
  - B:internal/httpserve/handlers/integration_github_webhook_flow_test.go:50
  - B:defs/githubapp/installation_test.go:18
  - B:defs/googleworkspace/operations_test.go:24

**ent schema:** add `user_input` (`openapi.IntegrationUserInput`) and `operation_config` (`openapi.IntegrationOperationConfigs`) with `SkipType|SkipWhereInput`.

→ **Regen #1:** `task generate` (user-run). Only additions, so the tree compiles.

### Phase 2 — the type system and every consumer (after regen #1)

1. **ii/types:**
   - remove `kind.go`, `InputRegistration` and its lookups, `Registration(base)`, and the payload constructors
   - add `Identity`, `declare.go` (`Define`, `Declaration`, `Document`), `state.go`, `conform.go` (`Conform`, `Sources`, `Writes`, `Failure`, `HealthMarker`), `surface.go` (`SurfaceOf`), `form.go`, `connection.go`, `catalog.go`
   - restore the provider-state helpers
   - remove `OperationRegistration.UISchema`, `ClientRegistration.ConfigSchema` and `DisconnectRegistration.Schema`
2. **ii/providerkit:** `OperationSettings`, `DirectorySync`, `DirectorySyncOperation`, `Messaging`, `DocumentSource`, `ExprPrimaryDirectory`.
3. **ii/registry:**
   - delete `finalize.go`
   - drop the duplicate-identity checks `Define` owns
   - add `Snapshot{hash, version, surface}`, embedded `Surfaces`, `WithSnapshots` with the hash check in `Register`, `Version`/`Fingerprint`, and the extended gate
4. **internal/keystore:** `Load` filtered to declared slots and their `Replaces`; `Save`/`Delete` on `CredentialBindings`.
5. **ii/runtime:**
   - `load`/`persist`/outcome mapping replaces `inputs.go`, `upgrade.go` and the 8 call sites, with `versionOrder` as the single comparator
   - `WithSnapshots(registry.Surfaces)` only when building from catalog builders
   - `Disconnect` calls `load`, proceeds on failure, and reads `provider_state`
   - `RefreshInstallationMetadata` routed through `conformInstallation(..., rederive = true)`
   - every reader uses `UserInput`, `OperationConfig`, `InstallationMetadata.Attributes` and `ProviderState`
   - nothing references `UserInputs`, `OperationInputs`, `CredentialRef` or writes `Config`
6. **ii/operations:** ingest filter via `OperationRegistration.Settings`; delete `operations.UserInput`.
7. **defs/\* (21 builders):**
   - builders return `types.Define(...)`
   - shared types
   - `PrimaryDirectory` removed from the 7 user inputs; scim keeps `Name`
   - `UpgradeFrom` backfill guards on githubapp `VulnerabilitySync`, cloudflare `FindingsSync` and okta `DirectorySync` (§5.5)
   - cloudflare and googleworkspace on `providerkit.DirectorySync`, honouring `DisableGroupSync`; slack, zitadel and scim keep local `OperationSettings`-only configs
   - githubapp `max_repos` tag restored
   - cloudflare single runtime config
   - email operations declared from `Dispatcher` entries; teams declares no custom credential form schema
   - every `Disconnect` call passes `(description, fn)`
   - the catalog files are unchanged
8. **Consumers:**
   - handlers (`Writes`, top-level `primaryDirectory`, `def.Catalog()`)
   - graphapi (`types.Form`, `UserInputAs`, the K29 capability query, internal/graphapi/workflow_metadata_extensions.go to accessors)
   - workflows (`UserInputAs`)
   - internal/keymaker/service.go to `def.Connection(slot)`
   - the internal CLI (`ii/cli/cmd/integration/configure.go`) and the separate `cli` workspace module (request shapes)
   - CLI stub templates (`ii/cli/stub/*.tmpl`)
   - `internal/testutils/integrations` (non-test; builds `Definition` literals)
9. **Tests:** every test using removed APIs or branch-only fields. These are enumerated by `git grep -n -E 'UserInputs|OperationInputs|SetCredentialRef|StoredDocuments?\b|Inputs\(types\.Kind|\.Registration\(types\.' -- ':!docs' ':!internal/ent/generated' ':!internal/graphapi/generated' ':!internal/ent/entityops'`.

### Phase 3 — targeted removal of the branch-only generated fields

Done: `user_inputs`, `operation_inputs` and the Integration `credential_ref` are removed from internal/ent/schema/integration.go and from the generated code (internal/ent/generated/integration.go, mutation.go, entql.go, integration/integration.go, integration/where.go, migrate/schema.go) and internal/ent/entityops/entity_registry.go.

`StoredDocument`/`StoredDocuments` are deleted from common/openapi/integration_models.go.

Remaining:
- generate the snapshots by running `task config:generate` (user-run, not yet run). The committed `Hash` is empty, so the generator rewrites all 21 files under internal/integrations/registry/surfaces ungated with `{hash, version, surface}` and fresh ULIDs, and writes the fixture snapshot files that the `go:embed` patterns in internal/testutils/integrations/surfaces name (`{base,mockhttp,v1,v2,v3}/surfaces/*.json`), which are not in the working tree yet (§5.4).

Any migration change comes from the user's generation run, not from a step here.

### Phase 4 — next release

- `task drop:field -- Integration.config`.
- Remove the one-time backfill (§5.5): `Sources.Legacy`, the legacy branches of `UserInputRef.conform` and `OperationRef.source` with `legacySection`, the three `UpgradeFrom` methods, and `IntegrationConfig`.

---

## Part 8 — Net code estimate (non-test)

These are the revision 3 estimates, kept for reference and not re-measured against the implementation. Measured bases are marked M; judged sizes are marked E.

| Package | Removed | Added | Net |
|---|---|---|---|
| ii/types | ≈360 (M): kind.go 19; definition.go InputRegistration/lookups/fields ≈94; ref.go `Registration`/`Connection`/payload constructors ≈186 with comments; operation.go settings ≈30; input.go registration/replacing ≈22; misc ≈10 | ≈1,100 (E): Define/Definition/accessors/Catalog ≈300; Conform/Sources/Writes/Failure/Surface/Form iteration ≈200; four handles × document methods ≈280; Connection ≈90; Input methods ≈90; Webhook fluent ≈100; provider-state restore ≈45 (M: M:definition.go:162-206) | ≈+740 |
| ii/registry | ≈265 (M): finalize.go 119; registry.go Surface types and projection 36-154 ≈119; validateInputs 337-363 ≈27 | ≈150 (E): embedded surfaces, `{hash, version}` check, ULID comparison, extended gate | ≈−115 |
| ii/runtime | ≈600 (M): inputs.go 450; upgrade.go 65; call-site blocks ≈40; credential_ref paths ≈40 | ≈330 (E): load/persist/outcomes; rename logic moved (≈85, M) | ≈−270 |
| ii/operations | ≈45 (M) | ≈10 (E) | ≈−35 |
| ii/providerkit | ≈20 (M) | ≈150 (E) | ≈+130 |
| internal/keystore | ≈40 (M) | ≈30 (E) | ≈−10 |
| defs (21) + catalog | not audited per file | not audited per file | directionally negative; unverified |
| handlers / graphapi / workflows | ≈90 (M) | ≈50 (E) | ≈−40 |
| common | ≈40 (M) | ≈40 (E) | ≈0 |
| ent schema | ≈20 (M) | ≈14 (E) | ≈−6 |

Outside the definitions, net is roughly +390 lines. The growth is the explicit document machinery and failure isolation that the branch did not have. The definitions shrink, but by an amount not measured here.

---

## Part 9 — Decisions

1. **Version and fields move together.** Every upgrade writes all conformed fields and the new version in one transaction (§5.2 step 5). No rollback runbook is part of the design.
2. **Primary directory.** The user sets it. The catalog offers it only on definitions with an operation that ingests `DirectoryAccount`: the 12 listed in §3.2.
3. **Operator config.** The catalog and `catalog.Config` stay as they are (§2.7).
4. **Rolling deploys.** Versions are ULIDs from committed `{hash, version, surface}` snapshots. A binary never writes to a row whose version is newer than its own (§5.4).
5. **SCIM name.** It stays a required user input field, because one org can run several SCIM installations. A submitted user input's `name` is mirrored to the `name` column.
6. **Single validator.** `Define` validates a definition's internal consistency; the registry checks only cross-definition and registry-level facts (§2.6, §2.7).
7. **No kind dispatch.** Documents differ only through their own methods and the typed `Definition` views; failures differ only through the `Failure` interface (§2.5).
8. **One intermediary.** The one-time backfill from main (§5.5) is the only transitional construct, and Phase 4 removes it with the `config` column.
9. **Snapshot enforcement.** The runtime enforces committed snapshots only when it builds from the catalog builders; the generator gates only against snapshots that carry a hash (§5.4).

---

## Appendix A — Adversarial review log

Every finding was re-checked by hand before it was accepted. Notation:
- **V** — verified by opening the cited code
- **R** — rejected
- **Disp.** — the disposition in this revision

### A.1 Data path

| # | Finding | Check | Disp. |
|---|---|---|---|
| 1 | googledrive rows without `folderId` would fail conform and error the whole installation | V: B:defs/googledrive/operation_folder_sync.go:25 `required`; main tolerated it (M:...:38-42); schema_conform.go:72-91 fills only declared defaults | per-document failure isolation (§5.2 step 5) |
| 2 | `resolve` (a network call) in conform errors installations on provider outages | V: B:defs/okta/installation.go:26; upgrade.go:23-35 | conform is offline; resolve only on credential write / refresh |
| 3 | Okta empty `clientConfig` turns group sync on | V: M:okta/operation_directory_sync.go:39-43, 80-82 | inference keyed on `definition_version == ""` incl. `{}`; guarded `UpgradeFrom` |
| 4 | Rollback after stripping or clearing `config` loses main's op settings | V: 7 definitions without `UserInputOf` (awssecurityhub, azuresecuritycenter, cloudflare, gcpscc, githubapp, oci, tailscale) | `config` never written; new `user_input` column; drop in Phase 4 |
| 5 | Roll-forward ignores edits made on main during rollback | V: main never sets `definition_version` | not designed for; versions and fields move in one transaction (Part 9.1) |
| 6 | Equality gate ping-pongs across rolling deploys | V: B:ii/runtime/upgrade.go:19 | ULID versions; newer rows are not written (§5.4) |
| 7 | Deleting slots absent from `next` would delete user-linked hush rows | V: B:internal/keystore/store.go:64-83, 223-245 unfiltered; schema.graphql:14744 | `Load` filtered; delete only moved `Replaces` |
| 8 | Disconnect loses its fallback when the upgrade fails | V: B:ii/runtime/credentials.go:66-70 | Disconnect reads `provider_state` directly |
| 9 | Required defaults never apply to non-omitempty fields | V: schema_conform.go:72-76 | stated in §2.3 |
| 10 | `max_repos` → `maxRepos` breaks persisted per-run configs | V: M:...:19, B:githubapp/types.go:72 | restore `max_repos` |
| 11 | Branch-era connection choices are lost | V; non-production only | accepted |
| 12 | `Display.CredentialRef` dropped | V: B:ii/runtime/inputs.go:437 | set from connection |
| 13 | "Only `definition_version` written" was false | V | step 6 reworded; no resolve in conform |
| S1 | Accept with `from == ""` would apply okta's legacy default to new submissions | V by construction | `Accept` passes the current layout |

### A.2 Type system

| # | Finding | Check | Disp. |
|---|---|---|---|
| 1 | Runtime can't call unexported `Document` methods | V (Go visibility) | exported `types.Conform/Surface/Form/UserInputAs` |
| 2 | Erased op views can't reach typed config | V: S5 sites | `OperationRegistration.Settings`/`Config` closures; `Name`/`Description`/`Schema` kept |
| 3 | Storage types repeat `Layout` | V | `models.Revision` embedded; metadata deviation stated |
| 4 | Adding an integration touches catalog and config | V: catalog.go:32-56, types.go:20-54, config.go:83; K29 | catalog kept by decision (§2.7); K29 via capability query |
| 5 | Implicit client install can't be single-sourced; package-var examples don't compile | V | explicit one-time client declaration; examples inside builders |
| 6 | No `Description` setter, no per-op probe; email reads the erased view | V: azureentraid/builder.go:63; email/system_emails.go:758-769 | `Describe`, `Probe`, `IsCustomerSelectable` |
| 7 | Accessors incomplete; testutils builds literals | V | accessor list extended; Phase 2 item 9 |
| 8 | "Webhook requires Routes" breaks scim; scim doesn't fit the preset | V: scim/builder.go:24-28 | `Routes` optional; eligibility from ingest contract |
| 9 | State member claim contradicted §5.2; `Credentials[name]` on a slice | V | order groups stated; bindings API used |
| 10 | providerkit can't implement `Declaration` | V | preset returns `types.OperationRef[C]` |
| 11 | Ambiguous `Validate` silently un-stores an op | V (snippet) | `Define` reflective check |
| 12 | K27 (githubapp literals) was wrong | V: webhook.go:158-174 returns `.Name()` | removed from inventory |
| 13 | `UserInputAs` needs a decode hook | V | lives in types, uses internal method |
| 14 | Duplicate storage inside `Definition` | V | single `documents` store with assertion accessors |
| 15 | `Replaces` implemented twice | V: ref.go:630-652 | shared `Identity` |
| S1 | `lo.CamelCase` heuristic vs per-op upgrade | judgment | kept per stated direction; confined to never-conformed rows |

### A.3 Phases

| # | Finding | Check | Disp. |
|---|---|---|---|
| 1 | Phase 1 missed integration_config.go:95 and tests; listed files that don't change | V (grep) | list corrected |
| 2 | Phase 2 missed testutils, keymaker, workflow metadata, CLI and stubs, tests | V | Phase 2 items 9–10 |
| 3 | Regen doesn't generate migrations; deleting the migration pair breaks atlas.sum | V: create_migrations.go:43; atlas.sum tails | migration pair kept; single `atlas:create` in Phase 3 |
| 4 | Committed snapshots in HEAD shape fail the gate | V: python read of okta snapshot keys `ref`, `userInput` | re-baseline in Phase 3 |
| 5 | GraphQL `credentialRef` pruned by `GenQuery` | V (code read, entx genquery) | stated |
| 6 | CLI and request shapes omitted | V | Phase 2 item 9 |
| 7 | Part 8 bases overstated | V: measured ref.go 124 lines code (186 with comments) | Part 8 recomputed |

### A.4 Citation passes

Each correction below was re-checked and applied:
- K19 lines are :137/:69.
- K5 attribution corrected.
- `RuntimeIntegrationRegistration` row added.
- The installation-metadata DB writer is M:credentials.go:253-256.
- Main filter resolver is M:ingest.go:875-899.
- `CredentialSet` spans 11-16.
- Main `IntegrationProviderState` is at M:74-77.
- `Reconcile` is at 109-152.
- Health rules are at 365-395.
- The rename logic excludes the row update (201-229, 253-306).
- Part 8 sizes were recomputed.
- The auth byte-identity claim is limited to non-test files.

---

## Appendix B — Implementation decisions after revision 3

Each entry was checked against the working tree. Where an entry reverses a revision 3 statement, Parts 2–7 now follow the entry.

### B.1 Types

| # | Decision | Where |
|---|---|---|
| 1 | `Connection.Disconnect(description string, fn DisconnectFunc)`; a nil `fn` only describes the teardown | ii/types/connection.go |
| 2 | `UserInputRef.Resolve` and `InstallationRef.Resolve` take `*generated.Integration`, are nil-safe, and return `false` for an empty document | ii/types/ref.go |
| 3 | `Failure` is an interface (`error`, `Aborts`, `Apply`) applied through `HealthMarker`; the four implementations each carry `Unwrap`; the runtime has no switch over failure kinds | ii/types/conform.go, ii/runtime/load.go |
| 4 | `Document` methods are `conform`, `stage`, `formInto`, `surface`, `failure`; `Definition` keeps typed views (`credentials`, `userInput *userInputDeclaration`, `installation`) written at declare time; no kind markers, no type switches, no `Documents()` accessor | ii/types/declare.go, ii/types/definition.go |
| 5 | `Define` rejects clients without `Builds`, runtime configs without a build function (`ErrRuntimeBuildRequired`; the registry no longer checks it), invalid (empty) client IDs, connections with an empty selecting slot, and client credential slots no connection declares | ii/types/ref.go, ii/types/connection.go, ii/types/declare.go |
| 6 | The reflective registration check on configs embedding `OperationSettings` was removed; storedness is the `Configurable` assertion alone | ii/types/ref.go (`stored`) |
| 7 | `OperationRegistration.UISchema`, `ClientRegistration.ConfigSchema` and `DisconnectRegistration.Schema` were removed; `OperationRegistration.Schema` serialises as `configSchema` | ii/types/operation.go, client.go, disconnect.go |

### B.2 Registry and versions

| # | Decision | Where |
|---|---|---|
| 8 | Registry duplicate-identity checks removed; `Define` is the single validator | ii/registry/registry.go |
| 9 | `catalog` and `catalog.Config` kept as they are | defs/catalog |
| 10 | ULID versions in committed `{hash, version, surface}` snapshots, embedded from internal/integrations/registry/surfaces; `WithSnapshots` used only when the runtime builds from the catalog builders | ii/registry/snapshot.go, ii/runtime/runtime.go |
| 11 | The snapshot generator gates only when the committed `Hash != ""` | jsonschema/integration_schema_generator.go |
| 12 | Test definitions get committed fixture snapshots in internal/testutils/integrations/surfaces | internal/testutils/integrations/surfaces/surfaces.go |
| 13 | `RefreshInstallationMetadata` routed through `conformInstallation` with `rederive = true`; a single `versionOrder` comparator serves the gate and the upgrade sweep (§5.2) | ii/runtime/load.go, credentials.go, upgrade_sweep.go |

### B.3 Definitions and shared concepts

| # | Decision | Where |
|---|---|---|
| 14 | The primary directory is offered when any operation ingests `DirectoryAccount` | ii/types/catalog.go |
| 15 | scim `Name` stays required and is mirrored to the `name` column | defs/scim/types.go, ii/runtime/load.go |
| 16 | cloudflare and googleworkspace use `providerkit.DirectorySync` and honour `DisableGroupSync` | defs/cloudflare, defs/googleworkspace |
| 17 | slack, zitadel and scim keep local `OperationSettings`-only directory sync configs | defs/slack/types.go, defs/zitadel/types.go, defs/scim/operation_directory_sync.go |
| 18 | email exposes catalog facts through `Dispatcher` accessors | defs/email/emailop.go |
| 19 | teams' custom credential `FormSchema` dropped | defs/microsoftteams/builder.go |
| 20 | The one-time backfill from main (`Sources.Legacy`, its section inference, and the three `from == ""` guards) is the only intermediary construct; Phase 4 removes it | §5.5 |
| 21 | Phase 3 removed the `user_inputs`/`operation_inputs`/`credential_ref` fields from the ent schema and generated code; migrations come from the user's generation and are not planned steps | Part 7 |
