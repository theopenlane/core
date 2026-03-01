# Integrations Refactor Plan (Scope Lock)

## Purpose
Keep the integrations refactor on a strict, explicit path and prevent unapproved design drift

## Locked Decisions
1. Resolve execution targets from the registered provider operation registry using explicit criteria
2. Use operation descriptors (`name`, `kind`) instead of string heuristics (`contains`, loose matching)
3. Keep materialization contract-driven through mappings/CEL and shared upsert flows
4. Avoid provider-specific intermediate model trees for ingestion/materialization
5. Preserve provider payload values as-is; avoid internal string normalization/manipulation
6. Keep errors static/sentinel where used for control flow
7. Persist provider state as provider-keyed generic JSON maps, not provider-specific typed root fields
8. Use a single extensible workflow metadata field (`extensions`) for integration metadata evolution to avoid repeated GraphQL shape churn

## Current In Scope
1. Workflow integration action criteria:
- `integration_id` or `provider`
- `operation_name` or `operation_kind`
2. Explicit target resolution via `internal/integrations/targetresolver`
3. Scope expression gating for integration actions via `internal/integrations/scope`
4. Runtime policy routing from explicit metadata (`run_type`, `operation_kind`)
5. Notification message targets supporting both user targets and direct channel targets
6. Google Workspace directory sync operation emitting envelopes for shared directory-account ingest
7. OAuth callback path/config correctness for providers currently being touched

## Out Of Scope
1. New provider-specific state mutation branches in centralized store logic
2. Provider-name substring routing/classification
3. New bespoke ingestion models per provider when mappings/CEL can represent the transform
4. GraphQL schema expansion unless strictly required for generated object shape
5. Extra operation families/domains not yet backed by active provider implementations

## Change Control
1. Any new behavior outside “Current In Scope” must be added here first
2. If a change does not map to a listed scope item, do not implement it
3. Keep deletion-first refactors: remove replaced code paths before adding new replacements
4. Follow helper reuse directives in `internal/integrations/HELPER_REUSE_GUIDE.md`

## Implementation Phases
1. Cleanup and remove conflicting/brittle staged paths
2. Land resolver-based operation selection and queueing criteria
3. Land workflow integration scope gating and validation
4. Land notification message-target execution path
5. Land Google Workspace directory envelope sync and shared ingest use
6. Final pass to remove stale fallbacks/legacy aliases and confirm no dead additions

## Pre-merge Checklist
1. No `strings.Contains` operation classification remains
2. No newly introduced provider-specific switch logic in shared core paths
3. No new intermediary provider data model trees for ingest
4. All added errors are used
5. All added behavior maps to “Current In Scope”
6. Provider state reads/writes use `provider_state.providers.<provider>.*` and shared merge helpers

## Contract Hardening Inventory (2026-03-01)
Primary audit scope: `common/integrations`, `internal/integrations`, `internal/workflows`, `internal/httpserve/handlers`, `internal/keystore`, `internal/keymaker`

Highest-priority contract boundaries:
1. `common/integrations/types/operation.go` (`OperationInput.Client`, operation config/details maps)
2. `common/integrations/types/provider.go` (`ClientBuilderFunc`, provider/client config maps)
3. `internal/integrations/ingest/*` request boundaries (`ProviderState`, operation config map flow)
4. `internal/workflows/engine/integration_executor.go` request/config/scope map boundaries
5. `internal/workflows/engine/notification_templates.go` rendered template payload boundaries
6. `internal/keystore/client_manager.go` pooled client boundary (`ClientPool[any, map[string]any]`)
7. `internal/httpserve/handlers/integration_operations.go` request/response JSON document boundaries

## Paused Point And Next Steps
Last paused point before this pass: map/`any` boundary review was raised after the metrics hardening work in `integration_executor`

Completed in this pass:
1. Tightened ingest request contracts from `ProviderState any` to `state.IntegrationProviderState`
2. Tightened provider-state merge contract to `MergeProviderData(provider string, patch map[string]any)`
3. Removed redundant provider-state patch decode error
4. Tightened Slack operation config contract by removing `TrimmedString`/`any` fields in favor of concrete types
5. Tightened rendered notification template blocks from `any` to `[]map[string]any` with explicit decode
6. Tightened workflow schema type registry from `Value any` to `Type reflect.Type`
7. Tightened workflow object node contract from `any` to `generated.Noder` with explicit typed decode in helper loader
8. Tightened operation/client execution boundary by replacing raw client `any` contracts with `types.ClientInstance` across `common/integrations`, providers, and `internal/keystore`
9. Removed `OperationResult.Details map[string]any` success payloads in integration/workflow runtime paths in favor of typed detail envelopes encoded via `jsonx`
10. Removed normalization trims in touched activation/session/template paths in this session (`activation/service.go`, `keymaker/session_store.go`, `operation_templates.go`)

Remaining TODO hardening items:
1. None for the current integrations/workflows scope
2. Keep `any` only for generic helper constraints and JSON codec entry points, not runtime domain contracts

## Continuation Update (2026-03-01)
Completed in this continuation pass:
1. Replaced `common/integrations/types.OperationRequest.Config` contract from `map[string]any` to `json.RawMessage`
2. Refactored `internal/keystore/operations_manager.go` to decode operation request config documents through `jsonx.ToMap` at execution/client boundaries
3. Refactored `internal/workflows/engine/integration_executor.go` integration queue/action runtime config boundaries (`config`, `scope_payload`) to JSON documents (`json.RawMessage`) decoded via `jsonx`
4. Updated workflow/manual integration operation callsites (`notification_templates`, `integration_operations`) to emit config documents via `jsonx.RoundTrip`

## Session Close Checkpoint (2026-03-01)
1. Net retained changes in this session are limited to:
- `common/integrations/operations/operation_templates.go`
- `internal/integrations/activation/service.go`
- `internal/keymaker/session_store.go`
2. Constraint lock for next session:
- Do not add new intermediary model/type files for boundary hardening work
- Do not add new local wrapper/helper layers that preserve `map[string]any` seams
- Only land in-place changes that reduce existing dynamic boundary surfaces
