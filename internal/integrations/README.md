# Integrations Execution Model

This package uses explicit provider descriptors and workflow-driven criteria to execute integration commands and materialize provider data

## Command Execution

1. Providers register operation descriptors through their builder (`types.OperationDescriptor`)
2. Callers submit execution criteria (`integration_id` or `provider`, plus `operation_name` or `operation_kind`)
3. `targetresolver` resolves the installed integration record and exact descriptor from the shared registry
4. Queue policy is selected from explicit metadata (`run_type` and operation descriptor `kind`)

## Workflow Integration Actions

Workflow `INTEGRATION` action params support:
- `integration_id`
- `provider`
- `operation_name`
- `operation_kind`
- `scope_expression`, `scope_payload`, `scope_resource`
- `config`

`scope_expression` is evaluated with CEL before queueing the run using:
- `payload`
- `resource`
- `provider`
- `operation`
- `config`
- `integration_config`
- `provider_state`
- `org_id`
- `integration_id`

`provider_state` uses provider-keyed state maps:
- `provider_state.providers.<provider>.<field>`
- example: `provider_state.providers.githubapp.installationId != ""`

## Workflow Metadata Surface

`workflowMetadata.extensions.integrations` is the extensible payload for integration action configuration metadata
It includes action contract selectors, scope variable names, run types, and provider operation descriptors
This is the stable expansion point so future integration metadata additions avoid repeated GraphQL schema changes

## Materialization

Materialization remains contract-driven through mapping schemas and CEL:
- providers emit generic envelopes (`types.AlertEnvelope`)
- ingest resolves mapping overrides/defaults
- CEL filter + map expressions produce normalized fields
- persistence uses generated ent inputs and upsert paths

Do not introduce provider-specific intermediary model trees for materialization when schema + CEL mapping can express the transform

## Adding a New Operation

1. Add the operation descriptor to the provider (`Name`, `Kind`, `Run`, optional schemas)
2. Ensure provider builder returns sanitized descriptors
3. Trigger execution using explicit criteria through workflow actions or internal queue requests
4. If the operation emits materializable payloads, emit envelopes and add mapping defaults/overrides
5. Reuse shared ingest/upsert paths before adding provider-specific persistence logic

## Guardrails

- Do not classify operations with `strings.Contains` or other name heuristics
- Do not trim or mutate provider payload values during internal processing
- Prefer static sentinel errors for predictable routing and handling
- Reuse existing shared helpers documented in `internal/integrations/HELPER_REUSE_GUIDE.md` before adding new utility methods
