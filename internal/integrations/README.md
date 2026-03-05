# Integrations Development Guide

This guide documents the minimum file changes required to:
1. Add a new integration provider.
2. Add a new ingest capability (object materialization) to an existing provider.

## Core Concepts

- Provider operations emit normalized `AlertEnvelope` payloads in operation `details`.
- Operation descriptors declare ingest behavior with `OperationDescriptor.Ingest []types.IngestContract`.
- Ingest contracts are schema-based (`Vulnerability`, `DirectoryAccount`, etc), not operation-name based.
- Provider default mappings are published via `types.MappingProvider` and resolved by the registry mapping catalog.
- Schema-scoped ingest topics/listener names are generated from ent schema annotations into `internal/ent/integrationgenerated`.

## Add a New Provider

1. Scaffold provider package.
- Command: `task new-provider NAME=<provider> AUTH=<oauth|apikey|custom>`
- Files created under: `internal/integrations/providers/<provider>/`

2. Add provider builder to catalog.
- Edit: `internal/integrations/providers/catalog/catalog.go`
- Register both OAuth and app-style builders as needed.

3. Add provider spec JSON.
- Add file: `internal/integrations/config/providers/<provider>.json`
- Include auth metadata, labels, and default scopes.

4. Implement provider operations.
- Edit: `internal/integrations/providers/<provider>/operations.go`
- Return `[]types.OperationDescriptor` with explicit `Name`, `Kind`, `Run`, and `ConfigSchema`.

5. Add operation ingest contracts where applicable.
- In each descriptor, set `Ingest` when the operation should materialize objects.
- Example:
```go
Ingest: []types.IngestContract{
  {
    Schema:         types.MappingSchemaVulnerability,
    EnsurePayloads: true,
  },
}
```

6. Publish provider default mappings (optional but recommended).
- Implement `DefaultMappings() []types.MappingRegistration` on provider type.
- Return schema + variant + mapping spec registrations.

7. Add tests.
- Provider operation tests.
- Mapping tests for any default mappings.
- Catalog conformance test should pass after registration.

## Add a New Ingest Capability to an Existing Provider

Use this when a provider should materialize an additional object type (for example, `DirectoryAccount`).

1. Decide whether you are reusing an existing schema or adding a new schema.
- Existing schema: just add provider mappings and operation contracts.
- New schema: add schema metadata + ingest handler (see "Add a New Schema").

2. Update operation descriptor contracts.
- Edit provider operation descriptor(s) in `operations.go`.
- Add schema contracts under `Ingest`.
- If operation output omits payloads by default, set `EnsurePayloads: true`.

3. Emit ingest envelopes in operation details.
- Single-schema operation: legacy `details.alerts` still works.
- Multi-schema operation: emit `details.ingest_batches`:
```json
{
  "ingest_batches": [
    {"schema": "Vulnerability", "envelopes": [...]},
    {"schema": "DirectoryAccount", "envelopes": [...]}
  ]
}
```

4. Add provider default mappings for the new schema.
- Edit provider mapping file (for example `ingest_mappings.go`).
- Return `types.MappingRegistration{Schema, Variant, Spec}` entries from `DefaultMappings()`.

5. If this capability is webhook-triggered, emit to the schema topic.
- Resolve topic with `ingest.IngestRequestedTopicForSchema(<schema>)`.
- Emit `ingest.RequestedPayload` with `Schema`, `Operation`, and `Envelopes`.

6. Validate with tests.
- Operation tests should assert envelope shape and schema routing.
- Ingest tests should cover mapping support and override behavior.

## Add a New Schema (Platform-Level)

Do this only when existing schemas cannot represent the object.

1. Annotate ent schema with `entx.IntegrationMappingSchema()` and regenerate (`task generate:ent`).
2. Confirm generated schema metadata in `internal/ent/integrationgenerated/integration_mapping_generated.go`:
- `IntegrationMappingSchemas` (CEL field map contract)
- `IntegrationIngestSchemas` (schema topic/listener contract)
3. Add schema constant in `internal/integrations/types` (`MappingSchema...`) for operation descriptors.
4. Implement ingest handler in `internal/integrations/ingest/`.
5. Register handler in `internal/integrations/ingest/dispatch.go`.
6. Ensure mappings resolve through registry mapping catalog.
7. Add end-to-end tests (operation -> ingest -> persistence).

## Example: Add Directory Account Import for GitHub

Goal: ingest users with access to GitHub resources as `DirectoryAccount`.

1. Add/extend GitHub operation to collect account access data.
- Edit: `internal/integrations/providers/github/operations.go`
- If operation should also continue vulnerabilities ingest, keep both contracts:
```go
Ingest: []types.IngestContract{
  {Schema: types.MappingSchemaVulnerability, EnsurePayloads: true},
  {Schema: types.MappingSchemaDirectoryAccount, EnsurePayloads: true},
}
```

2. Emit `ingest_batches` from operation details.
- Vulnerabilities go in `schema=Vulnerability` batch.
- Directory users go in `schema=DirectoryAccount` batch.

3. Add GitHub directory account default mappings.
- Edit: `internal/integrations/providers/github/ingest_mappings.go`
- Add `MappingRegistration` rows for `MappingSchemaDirectoryAccount`.
- Choose variants (for example `repo_collaborator`, `org_member`) if needed.

4. Keep provider mapping publisher generic.
- `DefaultMappings()` returns both vulnerability and directory-account registrations.

5. Test.
- Provider operation unit tests for batch output.
- Ingest mapping support tests.
- Workflow integration execution tests for multi-schema ingest.

## Validation Checklist

Run at minimum:

```bash
go test -tags test ./internal/integrations/... ./internal/workflows/engine
```

If adding a provider, also ensure catalog conformance tests pass (included in the command above).
