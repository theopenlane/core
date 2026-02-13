# Workflow + Integrations TODO

## Core Engine Integration (In Progress)
- [x] Add integration dependencies to `WorkflowEngine` and wire them from server options.
- [x] Implement `executeIntegrationAction` in the workflow engine.
- [x] Move run/ingest logic from the legacy runner into the workflow engine.
- [x] Add integration action validation in `internal/graphapi/workflow_validators.go`.
- [x] Remove `RegisterIntegrationOperationListeners` usage and `cmd/serve.go` integration emitter setup.
- [x] Update `internal/httpserve/handlers/integration_operations.go` to call the engine instead of emitting directly.
- [x] Add dedicated integration event pool owned by the workflow engine (isolated workers).

## Notification Templates + Integrations
- [x] Extend `NotificationActionParams` to accept `template_id` and/or `template_key`.
- [x] Regenerate `jsonschema/workflow.definition.json` via `jsonschema/workflow_schema_generator.go` (manual patch; generator failed due to existing build errors).
- [x] Add validation for template references in `internal/graphapi/workflow_validators.go`.
- [x] Implement a template renderer that loads `NotificationTemplate` by ID/key, merges workflow context, validates JSON config, and renders title/body/blocks.
- [x] Add provider “send message” operations (e.g., `message.send`) for Slack and Teams.
- [x] Update provider OAuth scopes to allow message sending; document re-auth requirements.
- [x] Build a notification dispatcher that renders templates and calls `OperationManager.Run`.
- [x] Keep creating in-app `Notification` rows and attach `template_id` for traceability.
- [x] Wire template-based notifications into `executeNotification` with fallback to title/body.
- [ ] Optional: enqueue external sends via jobspec for async delivery and retries.
- [ ] Add tests for renderer, workflow template notifications, and provider operations.

## Follow-ups / Hardening
- [ ] Add durable workflow linkage fields to `IntegrationRun` (instance/action/object) for restart/resume.
- [ ] Implement integration run reconciliation on startup.
- [ ] Add chunked/fan-out ingestion tasks for large payloads.
- [ ] Add idempotency and retry policies for integration actions.
