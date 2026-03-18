# Integrations Ingest

This package is a provider-agnostic ingestion pipeline using the CEL expression library rather than intermediary structs or taking the upstream data types and loading directly into our data types. It converts provider alert payloads into normalized `Vulnerability` records using declarative mapping rules, rather than hard-coding provider-specific transforms in each integration.

**Why This Approach?**
1. Decouples provider operations from persistence logic. Providers only emit alert envelopes; ingestion owns normalization and storage.
1. Enables customization without code changes. Mappings and overrides live in integration config so teams can tune field mapping or filtering per provider, alert type, or environment.
1. Scales to new providers. Adding a provider becomes “produce alert envelopes + ship default mappings,” not “write a new persistence pipeline.” (ideally...)
1. Enforces consistent validation. A single schema-driven validation step ensures required fields are present before persistence.

**Data Flow**
1. Provider operation collects raw alerts and wraps them in `types.AlertEnvelope` (alert type, resource, payload)
1. Ingest builds a mapping context with integration config, provider state, and operation inputs
1. CEL filter expressions decide whether each envelope should be ingested
1. CEL map expressions produce a normalized output map
1. Output is validated against the schema and persisted via upsert

**Key Concepts**
- `AlertEnvelope`: provider-agnostic wrapper for alert payloads and metadata
- Mapping schemas: canonical field sets defined in `integrationgenerated.IntegrationMappingSchemas`
- Mapping overrides: per-integration overrides that select a mapping by provider, schema, and alert type
- Default mappings: built-in mappings (for example GitHub) in `defaults_github.go`
- Retention policy: governs whether raw payloads are stored (intent would be to store via object storage provider)

This structure keeps ingestion consistent, testable, and configurable while avoiding one-off provider logic scattered across integrations. Hopefully.
