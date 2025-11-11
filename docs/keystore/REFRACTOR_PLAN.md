# Keystore Refactor & Type-System Plan

Path: `docs/keystore/REFRACTOR_PLAN.md`
Owner: Codex (GPT-5)
Last updated: 2025-11-10 (rev. 5)

---

## 1. Objectives

We are **not constrained by the existing keystore/keymaker implementation**. The plan is to design the ideal end state from scratch, then port/remove the current code once the new stack is proven. Objectives (derived from `internal/keystore/KEYSTORE_TODO.md`) now read:

1. **Provider architecture**: mirror Zitadel’s `internal/idp/provider.go` with a `Provider` interface plus per-provider modules and typed identifiers (no raw strings).
2. **Type reuse**: collapse every credential/token struct into a canonical type family (backed by `pkg/models.CredentialSet`) and propagate it across Keystore and Keymaker via generics/builders.
3. **Schema alignment**: use Ent edges to express org/integration/credential relationships (Integrations → Owner, Hush → Integrations). Query the graph with `WithOwner`, `WithSecrets`, etc., instead of copying IDs into metadata fields.
4. **Metadata removal**: stop persisting integration metadata as hush blobs; rely on structured JSON columns or new Ent fields when the schema truly needs more shape (most org/integration linkage already exists on the edges).
5. **Standardized grant flows**: lean on Zitadel’s RP + cloud/OIDC SDKs for refresh/mint operations instead of bespoke HTTP code.
6. **Typed provider lifecycle**: encode provider names, auth types, and capabilities as enums/interfaces; no `strings.ToLower` comparisons.
7. **Package separation**: maintain distinct Keystore (persistence/brokerage) and Keymaker (activation/UX) packages that share the same type system.
8. **Declarative handler alignment**: the new APIs must fully support the existing declarative OAuth/credential handlers (`internal/httpserve/handlers/*`). Provider schemas defined in JSON continue to drive the activation UX, and handlers interact with Keystore/Keymaker through typed interfaces.
9. **Client pooling**: leverage the `eddy` package for client pooling and credential reuse, offering worker/reconciler-friendly factories that plug into the new broker API.
10. **Error discipline**: define sentinel errors in `internal/integrations/errors.go` (or package-specific files) instead of ad hoc `fmt.Errorf`, and reuse them across layers.

> **Note**: Items 3 & 4 target long-lived credential data. OAuth activation state remains transient—keep that logic in signed state/cookies and reserve Ent/hush fields for durable credentials where the graph edges already provide the org/integration relationships (e.g., `Integration.WithOwner`, `Hush.WithIntegrations`).

## Common Directives

- Prefer sentinel errors (`ErrLoaderRequired`, `ErrCredentialNotFound`, etc.) over dynamic error strings.
- Use `samber/lo` helpers for slice/map operations (e.g., `lo.Map`, `lo.Filter`, `lo.Only`, `lo.Uniq`, `lo.Assign`) instead of hand-written loops; use `samber/mo` for option-like return values.
- Centralize sensitive operations in helpers (`helpers.CloneOAuthToken`, `helpers.CloneOIDCClaims`, `helpers.RandomState`) and reuse them rather than duplicating logic.
- Avoid string-literal keys/suffixes when Ent relationships already encode associations; rely on typed IDs and schema edges.
- Skip redundant nil/pointer guards in internal-facing code—callers are responsible for passing valid dependencies, and library errors should surface naturally.
- Keep GoDoc comments on exported items (no trailing periods) with field comments above declarations.
- Use Zitadel RP/OIDC types and published SDKs; only reach for `httpsling` when no client exists.
- Treat OAuth activation state as transient: use signed state/cookies (the existing handler pattern) and only persist long-lived credentials via hush + `CredentialSet` extensions (no Ent tables keyed by raw org/integration IDs for state storage).

---

## 2. Target Package Topology

```
internal/
└── integrations/
    ├── types/              // enums, credential payloads, builders, interfaces
    ├── providers/          // Provider interface + subpackages (github, slack, etc.)
    │   └── github/
    ├── registry/           // wiring for config + provider factories
    ├── config/             // provider catalog loading (JSON/YAML) + schema utils
    └── errors.go           // sentinel errors shared across integrations stack
```

- `internal/integrations/types` is the single source of truth for credential/token structs, provider enums, capability flags, and fluent builders.
- `internal/integrations/providers/<name>` contains provider-specific runtime logic (scopes, RP options, grant helpers).
- `internal/integrations/registry` loads config (from `config/providers/*.json` or Go-compiled specs) and returns typed providers.
- `internal/keystore` exposes persistence and broker APIs that operate strictly on `types` abstractions (legacy lives under `internal/keystore_old`).
- `internal/keymaker` orchestrates user flows (OAuth activation, verification) using the same types interface (legacy lives under `internal/keymaker_old`).

---

## 3. Type System Blueprint

Central package: `internal/integrations/types`

```go
package types

type ProviderType string

const (
    ProviderGitHub ProviderType = "github"
    ProviderSlack  ProviderType = "slack"
    // ...
)

type CredentialKind string

const (
    CredentialKindOAuthToken CredentialKind = "oauth_token"
    CredentialKindMetadata   CredentialKind = "integration_metadata"
)

type CredentialPayload struct {
    Provider   ProviderType
    OrgID      string
    IssuedAt   time.Time
    ExpiresAt  time.Time
    Credential models.CredentialSet
    Claims     map[string]string
    Scopes     []string
}

type Token[T any] struct {
    Credential CredentialPayload
    Raw        T
}

type Provider interface {
    Type() ProviderType
    BeginAuth(ctx context.Context, input AuthContext) (AuthSession, error)
    Mint(ctx context.Context, subject CredentialSubject) (CredentialPayload, error)
}

type AuthSession interface {
    RedirectURL() string
    Finish(ctx context.Context, code string) (CredentialPayload, error)
}
```

### Guiding principles

- **Single source of truth**: `CredentialPayload` (or its successor) is the sole struct persisted to storage, used by the broker, and exposed via APIs.
- **Generics for translation**: When provider-specific data is required (e.g., GitHub installation vs. Google WIF subject token) use `Token[TProviderClaims]` so we can maintain type-safety without defining new structs for each pipeline.
- **Builders/Fluent APIs**: introduce `CredentialBuilder` helpers to assemble payloads consistently, configured via functional options.
- **Library usage**: rely on `samber/lo` for collection helpers, `samber/mo` for optional values, Zitadel RP/OIDC packages for OAuth flows, `contextx` for per-request scoping, and `httpsling` only when a published SDK is unavailable.

---

## 4. API & Data Flow Overview

1. **Provider catalog**: JSON/YAML files (or Go maps) describe provider defaults. `registry.Loader` parses them into `types.ProviderConfig` structs.
2. **Provider runtime**: `providers/<name>` consumes the config + Zitadel RP to implement `types.Provider`.
3. **Keymaker**:
   - Presents activation UI using the catalog metadata.
   - Calls `Provider.BeginAuth`, handles redirects/callbacks, and receives a `CredentialPayload`.
   - Validates metadata via provider-defined `ValidateActivation` hooks.
4. **Keystore**:
   - Accepts `CredentialPayload` and persists it via Ent (using `hush.credential_set` + integration metadata).
   - Broker requests go through `Keystore.Broker`, which calls `Provider.Mint` (e.g., refresh token, STS exchange) and writes the result back using the same payload type.
5. **Consumers** (other services) interact only with `Keystore` APIs that return typed credentials; they never touch provider-specific structs directly.
6. **Declarative handlers** (`internal/httpserve/handlers/integration_config.go`, OAuth activation endpoints, etc.) call into `keymaker` to start/finish flows, and convert user input using the generated JSON schemas. Handlers remain declarative by loading config metadata from `integrations/config`.
7. **Client pooling**: reconcilers/event buses obtain provider clients via the new Keystore client pool API that wraps `eddy`. Pools expose typed builders (e.g., `keystore.ClientPool[types.ProviderGitHub]`) which return ready-to-use SDK clients configured with credentials fetched from the broker.

---

## 5. Work Streams & Tasks (detailed checklist)

- ### 5.1 Shared Types & Errors (`internal/integrations/types`, `internal/integrations/errors.go`)
- [ ] Define `ProviderType`, `AuthKind`, `CredentialKind`, and capability flags (bitmask or struct).
- [x] Implement `CredentialPayload` that embeds upstream types (`oauth2.Token`, `oidc.IDTokenClaims`, provider-specific structs) instead of re-declaring fields. Persist non-OAuth credentials via typed sub-structs that still serialize through `models.CredentialSet`.
- [x] Provide builder helpers for setting the embedded upstream types (`WithOAuthToken(*oauth2.Token)`, `WithOIDCClaims(*oidc.IDTokenClaims)`), keeping scopes/expiries sourced from those types rather than duplicating them.
- [ ] Ensure org/integration context stays outside the payload (handled by store/broker signatures and Ent edges) to avoid redundant IDs.
- [x] Introduce redactable wrappers/summary helpers (e.g., `CredentialPayload.Redacted()`) so anything containing secrets provides masked copies for logging and telemetry; default logging must never emit raw tokens or credential sets.
- [ ] Create `AuthContext`, `AuthSession`, `CredentialSubject` interfaces mirroring Zitadel semantics but tailored to our declarative flows.
- [x] Standardize cross-type mapping helpers (e.g., `CopyOAuthToken`, `CopyOIDCClaims`) that internally leverage `samber/mo` for optional returns and `samber/lo`/`maps.Copy` for map merges so the rest of the codebase never reimplements ad-hoc field copies.
- [ ] Ensure every exported type/function has a GoDoc comment on its own line (no trailing periods) and field comments sit above the field definition to keep lint consistent.
- [x] Add sentinel errors (`ErrProviderNotFound`, `ErrCredentialExpired`, etc.) in `errors.go`; reuse them across packages.
- [ ] Document required library usage (Zitadel RP, contextx, minimal httpsling) within the package README.

### 5.2 Provider Packages & Registry (`internal/integrations/providers/*`, `internal/integrations/registry`)
- [ ] Build `providers/provider.go` defining the interface + shared structs (e.g., `ProviderConfig`).
- [ ] Implement per-provider packages (start with GitHub) that configure Zitadel RP, OAuth scopes, and grant helpers.
- [ ] Use `samber/lo` for config defaults/merges and `contextx` for request scoping.
- [x] Add `registry.Loader` that ingests JSON schemas, validates via gojsonschema, and instantiates providers.
- [x] Ensure declarative schema metadata (form fields, docs links) remains accessible to HTTP handlers.
- [x] Replace bespoke `${VAR}` interpolation helpers with koanf-backed loading so JSON specs keep env overrides without duplicating logic (FS loader now leans on koanf + decode hooks, `interpolate.go` deleted).
- [ ] Introduce `internal/integrations/helpers` for reusable helper functions (copying tokens/claims, redaction, metadata transforms) so HTTP handlers, brokers, and other packages can share implementations instead of duplicating logic.
- [ ] Create `internal/integrations/testutils` exposing reusable setup/teardown suites that provision integrations, hush secrets, and Ent fixtures so most tests can run with real types/structs rather than mocks (mocks are acceptable only where dependencies are impractical to reproduce).

### 5.3 Keystore v2 (`internal/integrations/keystore`)
- [x] Implement Ent store using `integration` + `hush` schemas, storing credentials in `credential_set`.
- [x] Provide broker methods (`Fetch`, `Mint`, `Record`) returning `CredentialPayload`. *(Broker + cache implemented; client pool wiring pending)*
- [ ] Integrate caching and client pooling using `eddy` (client factory functions typed per provider). *(cache done; eddy integration pending)*
- [ ] Expose mapper utilities (similar to `internal/entitlements/entmapping`) for Ent ↔ domain conversions.
- [ ] Use sentinel errors for all failure modes; add tests covering AWS/GCP/GitHub flows.

### 5.4 Keymaker v2 (`internal/keymaker`)
- [x] Rebuild OAuth activation flows to call provider `BeginAuth`/`AuthSession`.
- [ ] Wire declarative handlers (`internal/httpserve/handlers/integration_config.go`, OAuth endpoints) to the new package.
- [ ] Preserve JSON-schema-driven validation for metadata forms; integrate `samber/lo` for data shaping.
- [ ] Use Zitadel RP (`rp.RelyingParty`) + `contextx` to manage nonce/state, falling back to httpsling only when no SDK exists.
- [x] Emit `CredentialPayload` to Keystore once activation succeeds.
- Activation sessions are short-lived and rely entirely on the existing cookie/state helpers (no Redis required). Begin/complete handlers set/validate cookies the same way the legacy flow did, then hand off to `keymaker.Service` for typed persistence in hush/`CredentialSet`.
- In progress: the new `internal/keymaker.Service` now drives begin/complete flows against the provider registry + keystore; next deliverable will add declarative handler wiring, metadata validation, and session minting for refresh/client pooling hooks.

### 5.5 CredentialSet Extensions (`pkg/models/credential_set.go` + new subtypes)
- [ ] Extend `CredentialSet` or introduce specialized structs (OAuth, AWS, GCP) plus conversion helpers.
- [ ] Provide fluent builders and typed getters; ensure JSON marshalling encrypts sensitive fields.
- [ ] Update Keystore/Keymaker to rely exclusively on these types (no string maps).
- > Reminder: hush + integration schemas already encode organization relationships via their edges. When persisting credential data, always query via `Integration.WithSecrets`, `Hush.WithOwner`, etc., instead of duplicating IDs in metadata blobs.

### 5.6 Schema Modernization (`internal/ent/schema/*`)
- [ ] Add/adjust Ent fields for integration metadata (e.g., provider user info, scopes).
- [ ] Ensure `hush.credential_set` stores opaque `CredentialSetEnvelope`.
- [ ] Generate Ent code and create mapping helpers.

### 5.7 Grant Helpers & External SDKs
- [ ] Implement reusable grant utilities for OAuth refresh, Google STS, AWS STS, GitHub App tokens, Workload Identity Federation.
- [ ] Prefer official SDKs; use httpsling only when unavoidable.
- [ ] Surface functional options for customizing scopes/audiences.

### 5.8 Config Pipeline (`internal/integrations/config`)
- [x] Continue supporting JSON provider definitions; add schema versioning + environment interpolation.
- [x] Provide APIs for handlers to retrieve form metadata, docs links, etc.
- [ ] Ensure provider specs contain all OAuth endpoints, redirect URIs, and scope hints so handlers/keymaker can compose begin/complete flows dynamically (no hardcoded URLs).
- [ ] Mirror the legacy cookie/session pattern using declarative metadata only; config acts as the single source of truth for auth URLs, callback paths, and any provider-specific knobs surfaced to the UI.

### 5.9 Declarative HTTP Handlers (`internal/httpserve/handlers/*`)
- [ ] Update OAuth + config handlers to call the new Keymaker APIs.
- [ ] Ensure handlers produce/consume typed provider IDs (no strings).
- [ ] Add tests demonstrating end-to-end flow from handler -> keymaker -> keystore.
- Next up: inject `internal/keymaker.Service` (wired with `keystore.Store` + provider registry) into handler bootstrap, port `StartOAuthFlow/HandleOAuthCallback` to call it while preserving the cookie/state UX, and route the resulting `CredentialPayload` through the typed keystore path. Once that works, exercise the declarative provider config so auth URLs/callbacks are composed dynamically per provider.

### 5.10 Client Pool Consumers
- [ ] Offer `keystore.ClientPool` builders that wrap `eddy` for worker/reconciler access.
- [ ] Document usage for event buses and background jobs, ensuring thread-safe credential refresh.
- [ ] Provide examples for fetching provider clients (GitHub/GCP/AWS) through the pool.

### 5.11 Legacy Deletion
- [ ] Remove old `internal/keystore` and `internal/keymaker` once parity achieved.
- [ ] Update imports across repo, run gofmt/go test.

---

## 6. Sequencing

1. **Types & enums** (5.1, 5.5): lock down the shared models first.
2. **Provider scaffolding & registry** (5.2, 5.7, 5.8): stand up the new runtime, focusing on one provider (GitHub) as a reference implementation. *(registry + generic OAuth provider scaffolded; provider catalog + builder wiring next, followed by GitHub-specific logic)*
3. **Keystore v2** (5.3, 5.6, 5.10): wire persistence/broker/client pools to the new types.
4. **Keymaker v2 + handler updates** (5.4, 5.9): implement OAuth activation flows with Zitadel RP, update HTTP handlers.
5. **Roll out additional providers** (repeat 5.2/5.7 per provider).
6. **Decommission legacy code** (5.9) once tests and handlers consume the new packages.

---

## 7. Open Questions

1. **CredentialSet scope**: expand the existing struct vs. create provider-specific subtypes (e.g., `credentialset.OAuth`).
2. **Config source of truth**: keep provider JSON files or move to Go-based configs for stronger typing?
3. **Testing harness**: do we need contract tests for every provider to ensure the new builder API stays consistent?
4. **Package naming**: should the top-level path remain `internal/integrations` or live under `pkg/integrations` for external reuse?
5. **Client pool ergonomics**: what is the preferred API shape for reconcilers (e.g., context-aware `Get(ctx, ProviderGitHub)` vs. typed factories)?

---

## 8. Tracking & Next Steps

- Use this document as the canonical tracking file; update the table as tasks complete.
- Consider adding checkboxes per task once implementation begins.
- When coding starts, cross-link PRs and migration scripts here for quick reference.
- **Progress 2025-11-10**:
  - Koanf now handles `${VAR}` expansion inside provider specs; the bespoke interpolation helpers and tests were removed, and loader coverage was extended to nested schema/default maps.
- **Immediate next steps**:
  - ✅ Swap `internal/integrations/config` over to koanf-based interpolation so embedded JSON specs expand `${VAR}` without the custom helper file.
  - ✅ Remove `interpolate.go` and any dead tests once koanf loading is wired, keeping the same error semantics (`ErrEnvVarNotDefined`, unsupported schema versions, etc.).
  - ☐ Re-run handler tests to confirm the new loader keeps OAuth flows functional.
  - ☐ Follow up on a real koanf provider (raw bytes or `fs.Provider`) plus watchers so we stop double-decoding JSON and can support YAML/provider reloads.
- **Declarative compatibility**: `internal/httpserve/handlers` should import `internal/keymaker` for activation flows. Handlers continue to reference provider JSON schemas for validation; they are not aware of storage details.
- **eddy integration**: `integrations/keystore` exposes client pools built with `github.com/theopenlane/core/pkg/eddy` (or equivalent) so reconciler/event-bus code can fetch typed clients safely.
- **Error/style guidance**: use sentinel errors from `integrations/errors.go`, `samber/lo` + `samber/mo` for functional helpers, `contextx` for request scoping, and `httpsling` sparingly (prefer published SDKs + Zitadel RP). Functional options + fluent APIs should configure builders/providers.
- **Ent mappers**: if conversion between Ent types and domain structs is required, follow the pattern from `internal/entitlements/entmapping` to keep mapping logic centralized.
