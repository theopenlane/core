# Integrations Ingest

This package is the provider-agnostic ingestion pipeline for security findings (currently vulnerability alerts). It converts provider alert payloads into normalized `Vulnerability` records using declarative mapping rules, rather than hard-coding provider-specific transforms in each integration.

**Why This Approach**
1. Decouples provider operations from persistence logic. Providers only emit alert envelopes; ingestion owns normalization and storage.
2. Enables customization without code changes. Mappings and overrides live in integration config so teams can tune field mapping or filtering per provider, alert type, or environment.
3. Scales to new providers. Adding a provider becomes “produce alert envelopes + ship default mappings,” not “write a new persistence pipeline.”
4. Enforces consistent validation. A single schema-driven validation step ensures required fields are present before persistence.

**Data Flow**
1. Provider operation collects raw alerts and wraps them in `types.AlertEnvelope` (alert type, resource, payload).
2. Ingest builds a mapping context with integration config, provider state, and operation inputs.
3. CEL filter expressions decide whether each envelope should be ingested.
4. CEL map expressions produce a normalized output map.
5. Output is validated against the schema and persisted via upsert.

**Key Concepts**
- `AlertEnvelope`: provider-agnostic wrapper for alert payloads and metadata.
- Mapping schemas: canonical field sets defined in `integrationgenerated.IntegrationMappingSchemas`.
- Mapping overrides: per-integration overrides that select a mapping by provider, schema, and alert type.
- Default mappings: built-in mappings (for example GitHub) in `defaults_github.go`.
- Retention policy: governs whether raw payloads are stored.

**Extensibility**
1. Add a new provider operation that emits alert envelopes.
2. Add a default mapping (or rely on integration config overrides).
3. Update or extend mapping schemas if new normalized fields are required.

This structure keeps ingestion consistent, testable, and configurable while avoiding one-off provider logic scattered across integrations.
