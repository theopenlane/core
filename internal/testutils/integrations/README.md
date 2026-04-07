# Integration Test Console

The integration test UI at [http://localhost:3004](http://localhost:3004) can drive end-to-end auth + callback testing through a local Dex-backed OIDC integration definition.

## Local Dex-backed OIDC Flow

1. Enable the local definition in `config/.config.yaml`:

```yaml
integrations:
  oidclocal:
    enabled: true
```

The operator config defaults are already aligned with the dedicated Dex compose file:

- discovery URL: `http://localhost:5557/dex`
- client ID: `local-core-oidc`
- client secret: `local-core-oidc-secret`
- redirect URL: `http://localhost:17608/v1/integrations/auth/callback`

Start the API server:

```sh
task run-dev
```

Start the local Dex container:

```sh
task compose:dex:integrations
```

Seed the basics:

```sh
task cli:user:all
```

Open [http://localhost:3004](http://localhost:3004), choose `Local OIDC (Dex)`, and start auth.

Use the Dex test user:

- email: `integration-test@theopenlane.io`
- password: `password`

After the callback completes, run the inline `HealthCheck` or `ClaimsInspect` operations from the test UI to verify the stored credential and returned ID token claims.
