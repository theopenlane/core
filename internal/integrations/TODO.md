# Integrations TODO

## Completed
1. Replaced `OperationInput.Client any` with `types.ClientInstance` and provider adapter decode boundaries
2. Reduced keystore pooled client boundary from raw `any` return types to `types.ClientInstance`
3. Replaced `OperationRequest.Config map[string]any` with JSON document boundary (`json.RawMessage`) and `jsonx` codec decode at operation execution boundaries
4. Reduced workflow integration execution map boundaries for `config` and `scope_payload` to JSON document boundaries with `jsonx` codec decode
5. Replaced integration/workflow operation result success/failure detail map literals with typed envelopes encoded through `jsonx` in the operation helpers

## Remaining Contract Hardening
1. None

## Boundary Rule
1. Keep `any` only for generic type parameters and JSON codec entry points
2. Keep `map[string]any` only at explicit JSON schema/CEL evaluation edges
3. Do not introduce new runtime domain contracts that expose raw `any`

## No-deviation Guardrail
1. Do not add new files/types solely for decode convenience in boundary hardening work
2. Do not add wrapper helpers that preserve the same `map[string]any` seam under a different function name
3. Prefer in-place strict assertions and existing helper reuse (`jsonx`, `mapx`, `gala`)
