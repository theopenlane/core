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

## Email Template Builder preview

`email-template-preview.html` is a representative mockup of the customer-facing
configurable email template editor. It drives the real GraphQL API end to end so it
reflects exactly what the production UI would receive and render:

- queries `emailTemplateCatalog` and builds the form **dynamically from `configSchema`**
  (field types/widgets come from the reflected JSON Schema — `format=color` → color
  picker, `format=uri` → URL input, arrays → repeatable lists), honoring the small
  `uiSchema` hint (multi-line body paragraphs);
- exposes a variable picker sourced from `variables` (`{{ .firstName }}` etc.) that
  inserts tokens into the focused text field;
- renders a **live HTML preview** on every edit via `previewEmailTemplate(key, defaults)`,
  with unfilled fields falling back to the server's demo values;
- includes an "API inspector" panel showing the raw `configSchema` / `uiSchema` /
  `variables` and the `defaults` payload being submitted — useful for articulating to the
  product UI team exactly what the API provides and expects.

### Run it

```sh
task run-dev                 # API server on :17608 (with the email integration in dev mode)
task cli:user:all            # seed the mitb@theopenlane.io user the page logs in as
task oauth-test-ui           # nginx serving this directory on :3004
```

Open [http://localhost:3004/email-template-preview.html](http://localhost:3004/email-template-preview.html).

> Note: the form opens empty while the preview shows demo data — that demonstrates the
> demo-fallback behavior. To open the form pre-filled with the example values instead, the
> catalog entry would need to expose the example payload as structured data (a small
> additive `exampleValues: Map` field); today only the rendered `htmlPreview` is exposed.
