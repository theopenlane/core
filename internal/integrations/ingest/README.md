# Integrations Ingest

This package materializes integration operation envelopes into normalized database objects using schema-driven CEL mappings.

## Current Model

1. Operations declare ingest contracts in `types.OperationDescriptor.Ingest`.
2. Workflow execution routes each contract schema to an ingest handler.
3. Webhook ingest emits schema-scoped Gala topics generated from ent mapping schemas.
4. Handlers resolve mapping specs from:
- Integration mapping overrides.
- Provider default mappings (from registry mapping catalog).
5. CEL filter/map expressions are evaluated with standard mapping vars.
6. Mapped output is validated against schema metadata, then persisted.

## Handler Routing

Schema handlers are registered in `dispatch.go`:
- `MappingSchemaVulnerability` -> vulnerability ingest handler
- `MappingSchemaDirectoryAccount` -> directory account ingest handler

Webhook listeners are registered per schema topic using generated contracts from `internal/ent/integrationgenerated`.

## Operation Output

Supported operation detail formats:

1. Single-schema legacy output:
```json
{"alerts": [...]}
```

2. Multi-schema output:
```json
{
  "ingest_batches": [
    {"schema": "Vulnerability", "envelopes": [...]},
    {"schema": "DirectoryAccount", "envelopes": [...]}
  ]
}
```

## Provider Mappings

Providers publish default mappings by implementing `types.MappingProvider` and returning `[]types.MappingRegistration`.

For broader integration implementation steps, see `internal/integrations/README.md`.
