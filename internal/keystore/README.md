## Keystore

`internal/keystore` hosts the declarative integration definitions and all of the
credential storage primitives. It includes:

- Provider specifications, schema generation helpers, and runtime wiring for
  OAuth/OIDC integrations.
- Persistence helpers for storing secrets + metadata plus the broker used to
  mint short-lived provider tokens.
- Shared helpers used by both the HTTP handlers and runtime components.

Runtime-facing packages (scheduler, reconciler, etc.) continue to live under
`internal/integrations`, but they now depend on this keystore layer rather than
the HTTP handler package.

### Provider Specifications

Provider metadata now lives in `internal/keystore/config/providers/<provider>.json`.
Each file mirrors the `ProviderSpec` struct (name, auth metadata, logo/docs
links, labels, etc.) and stays self-contained so new integrations can be added
without touching anything else. The same file also carries the
`credentialsSchema` section used by the UI; nothing additional needs to be
generated.

To add a new provider:

1. Create `internal/keystore/config/providers/<provider>.json` describing the
   auth strategy and defaults.
2. Point `integrationOauthProvider.providerSpecPath` (or leave it at the
   default) so the loader picks up the new definition.
3. The handler automatically exposes any `credentialsSchema` block you define,
   so make sure it’s present if the UI needs a manual form.
