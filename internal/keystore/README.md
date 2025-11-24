# Keystore Overview

The `keystore` package owns credential persistence, secure refresh, and the reuse of provider‑specific SDK clients and operations. It is built from a handful of composable building blocks that each focus on a single responsibility.

## Data & Credential Flow

1. **Store (`store.go`)** persists provider credentials for an org inside integrations + hush secrets (Ent).
2. **Broker (`broker.go`)** wraps the store, fetching cached payloads or minting fresh ones through the provider registry. It implements the `CredentialSource` interface consumed by higher layers.
3. **ClientPool (`client_pool.go`)** uses eddy’s caching primitives to build/reuse short‑lived provider SDK clients for a single provider/client pairing.
4. **ClientPoolManager (`client_manager.go`)** maintains a registry of pools driven by provider‑published `ClientDescriptor`s so integrations can self‑describe which clients they expose.
5. **OperationManager (`operations_manager.go`)** executes provider operations described by `OperationDescriptor`s, resolving credentials via the broker and optional clients via the manager.

## Components

### Store
- Validates org/provider inputs, ensures an integration record exists, and persists credential payloads as hush secrets (`SaveCredential`, `LoadCredential`, `DeleteIntegration`).
- Converts between `types.CredentialPayload` and the `models.CredentialSet` envelope for storage, cloning timestamps and provider metadata.
- Acts as the durable source of truth for all credentials referenced by the broker.

### Broker (CredentialSource)
- Implements `Get` by checking an in‑memory cache keyed by org + provider, falling back to the store when missing or expired.
- Implements `Mint` by calling the registered provider’s `Mint` handler with the current stored payload, persisting the result back through the store, and refreshing the cache.
- Provides the `CredentialSource` interface that `ClientPool`, `ClientPoolManager`, and `OperationManager` depend on so they can stay agnostic of storage details.

### ClientPool
- Parameterized by concrete client (`T`) and config (`Config`) types to preserve compile‑time safety inside first‑party packages.
- Accepts a `ClientBuilder` (usually an integration SDK adapter) and reuses instances via `eddy.ClientService`, deduplicating builds per org/provider/version.
- Handles credential refresh logic: force refresh, automatic refresh when token expiry nears (`refreshSkew`), and eviction when new credentials are minted.
- Exposes `ClientRequestOption`s so callers can inject per‑call config or force a refresh of cached credentials/clients.

### ClientPoolManager
- Lets providers publish `ClientDescriptor`s (provider, logical client name, build func). Each descriptor gets its own underlying `ClientPool[any, map[string]any]`.
- Performs descriptor validation (`descriptorKey`), clones config maps defensively, and keeps a copy of descriptors for discovery via `Descriptors()`.
- Gives dynamic consumers a single entry point (`Get`) where they specify org, provider, and client name without wiring concrete builders ahead of time.

### OperationManager
- Registers `OperationDescriptor`s (provider + operation name + run func) and exposes `Run` for callers.
- Resolves credentials through the shared `CredentialSource` (typically the broker) and, if the descriptor declares a client dependency, asks the injected `ClientPoolManager` to supply it.
- Supports per‑operation config maps, force‑mint (`req.Force`), and optional client force refresh (`req.ClientForce`).
- Returns structured `types.OperationResult` objects so integrations/CLI surfaces can report status and payloads consistently.

## Descriptor Types

- **`types.ClientDescriptor`**: provider identifier, logical client name, and a build function `func(context.Context, types.CredentialPayload, map[string]any) (any, error)`. Used by `ClientPoolManager`.
- **`types.OperationDescriptor`**: provider identifier, operation name, optional client name dependency, and a run function that receives `types.OperationInput`. Used by `OperationManager`.
- Helper constructors such as `FlattenDescriptors` / `FlattenOperationDescriptors` convert provider‑keyed maps into the slices accepted by the managers.

## Extending the System

1. **Add credential persistence**: ensure the provider’s onboarding flow calls `Store.SaveCredential` (usually via the broker).
2. **Register provider** with the integrations registry so the broker can look it up for minting.
3. **Expose clients** by returning `types.ClientDescriptor`s from the provider package; wire them into a `ClientPoolManager` via `NewClientPoolManager`.
4. **Expose operations** by returning `types.OperationDescriptor`s; pass them to `NewOperationManager`, supplying `WithOperationClients` if an operation needs reusable clients.
5. **Callers** use the manager APIs (`ClientPoolManager.Get`, `OperationManager.Run`) without needing to understand how credentials are persisted or clients are cached.

Together these components keep the keystore boundary clean: persistence lives in the store, refresh/caching in the broker and eddy pools, and dynamic runtime behaviors (clients + operations) hanging off descriptors so new providers can plug in without changing core code.
