# Openlane Workflow Examples

This directory contains example workflow definitions for Openlane. Workflows automate processes triggered by events in your Openlane organization.

## JSONSchema

All workflow definitions conform to the Openlane Workflow JSONSchema:
```
https://raw.githubusercontent.com/theopenlane/core/main/jsonschema/workflow-definition.schema.json
```

The schema is automatically generated from the Go types in `pkg/models/workflow.go`.

## Structure

A workflow definition consists of:

### Core Fields
- `name` (required): Human-readable workflow name
- `description`: Detailed description of what the workflow does
- `schemaType` (required): The schema this workflow applies to (e.g., Control, Evidence)
- `workflowKind` (required): Type of workflow (APPROVAL, LIFECYCLE, NOTIFICATION)
- `version`: Workflow definition version

### Triggers
Define when the workflow should execute:
```json
{
  "operation": "UPDATE",      // CREATE, UPDATE, DELETE
  "objectType": "Control",    // The schema type
  "fields": ["status"],       // Field changes that trigger evaluation
  "edges": ["controls"],      // Edge changes that trigger evaluation
  "selector": {                // Optional scoping
    "tagIds": ["tag-id"],
    "groupIds": ["group-id"],
    "objectTypes": ["Control"]
  },
  "expression": "..."         // Optional CEL expression
}
```

### Conditions
CEL expressions that must evaluate to true for the workflow to proceed:
```json
{
  "expression": "'status' in changed_fields",
  "description": "Only when status field changes"
}
```

#### CEL Variables (Base)
- `object`: the workflow object (typed entity)
- `user_id`: triggering user id (if available)
- `changed_fields`: list of changed field names
- `changed_edges`: list of changed edge names
- `added_ids`: map of edge name -> added IDs
- `removed_ids`: map of edge name -> removed IDs
- `event_type`: CREATE, UPDATE, DELETE

#### CEL Variables (Assignment Context - available in action When expressions)
- `assignments`: assignment summary object
  - `assignments.total`: total assignment count
  - `assignments.pending`: pending count
  - `assignments.approved`: approved count
  - `assignments.rejected`: rejected count
  - `assignments.by_action["action_key"]`: per-action summary with status, total, pending, approved, rejected, approver_ids
- `instance`: workflow instance context
  - `instance.id`: instance ID
  - `instance.state`: current state (RUNNING, PAUSED, etc.)
  - `instance.current_action_index`: current action index
- `initiator`: user ID who triggered the workflow

#### Common CEL Patterns
```cel
# Field was changed
'FIELD_NAME' in changed_fields

# Field equals value (use JSON field names, e.g. status, reference_framework)
object.status == "VALUE"

# Field changed to specific value
'status' in changed_fields && object.status == "VALUE"

# Edge added
'controls' in changed_edges && size(added_ids['controls']) > 0
```

### Actions
What happens when the workflow executes:
```json
{
  "key": "approval",
  "type": "REQUEST_APPROVAL",
  "description": "...",
  "params": {
    "targets": [
      {"type": "USER", "id": "user-id"},
      {"type": "GROUP", "id": "group-id"},
      {"type": "RESOLVER", "resolver_key": "CONTROL_OWNER"}
    ],
    "required": true,
    "label": "Approval label",
    "fields": ["status"]
  }
}
```

### Notification Actions
Notifications can be triggered conditionally based on assignment state:
```json
{
  "key": "notify_on_approval",
  "type": "NOTIFY",
  "when": "assignments.by_action[\"approval\"].status == \"APPROVED\"",
  "params": {
    "targets": [{"type": "RESOLVER", "resolver_key": "OBJECT_CREATOR"}],
    "title": "Request Approved",
    "body": "Your {{object_type}} change request has been approved",
    "channels": ["IN_APP"]
  }
}
```

#### Notification Action Notes
- `when` expressions are re-evaluated when assignment status changes
- Notifications with `when` expressions only fire once per workflow instance
- Available channels: `IN_APP`, `EMAIL`, `SLACK`
- Built-in resolver keys: `CONTROL_OWNER`, `CONTROL_AUDITOR`, `RESPONSIBLE_PARTY`, `POLICY_OWNER`, `POLICY_APPROVER`, `POLICY_DELEGATE`, `EVIDENCE_OWNER`, `OBJECT_CREATOR`
- There is no `INITIATOR` target resolver; use `OBJECT_CREATOR` or a static `USER`/`GROUP` target and reference `initiator` in `when` expressions if needed
- Template variables: `{{instance_id}}`, `{{object_id}}`, `{{object_type}}`, `{{action_key}}`

#### Common Notification When Patterns
```cel
# Notify when approval action completes successfully
assignments.by_action["manager_approval"].status == "APPROVED"

# Notify when quorum is reached (2+ approvals)
assignments.by_action["team_review"].approved >= 2

# Notify when all approvals complete
assignments.pending == 0 && assignments.rejected == 0

# Notify on any rejection
assignments.rejected > 0

# Notify specific action rejection
assignments.by_action["compliance_review"].rejected > 0
```

### Approval Action Notes
- Approval actions must include at least one `fields` entry.
- `fields` must be workflow-eligible for the object type (see GraphQL `workflowMetadata`).
- Use JSON field names (snake_case), e.g. `status`, `category`, `reference_framework`.
- Approval actions cannot target `edges` yet.
- Only one approval action per distinct field set is allowed in a definition.

## Examples

### 1. Simple Approval (`control-approval.json`)
Single approver workflow triggered when specific control fields change.

**Use case**: Require compliance team approval when control status or category changes.

### 2. Multi-Step Approval (`multi-step-approval.json`)
Sequential approvals for different control fields.

**Use case**: Engineering review on control category updates, followed by compliance approval for status changes.

### 3. Conditional Approval (`conditional-approval.json`)
Workflow that only triggers for specific control categories or conditions.

**Use case**: Approval only required for Technical category controls when status changes.

### 4. Webhook Notification (`webhook-notification.json`)
Sends an external webhook (e.g., Slack) when a control reaches APPROVED status.

**Use case**: Notify external systems when important control milestones are met.

### 5. Evidence Review (`evidence-review-workflow.json`)
Approval workflow triggered when evidence is linked to a control (edge trigger).

**Use case**: Require a reviewer to approve newly linked evidence.

### 6. Approval with Notifications (`approval-with-notifications.json`)
Approval workflow with dynamic notifications based on approval outcomes.

**Use case**: Notify the object creator when their request is approved, and notify control owners when requests are rejected.

### 7. Multi-Approver with Quorum Notifications (`multi-approver-with-quorum-notifications.json`)
Multi-approver workflow with notifications at each milestone (first approval, quorum reached, rejection).

**Use case**: Track approval progress with notifications when 1 of 2 required approvals is received, when quorum is reached, or when the request is rejected.

## Using Workflows

### CLI (Recommended for System Templates)
```bash
# Create workflow from JSON file
openlane workflowdefinition create \
  --definition-file internal/workflows/examples/control-approval.json

# List all workflows
openlane workflowdefinition get

# List system templates
openlane workflowdefinition get --system-owned
```

### UI
1. Navigate to Workflows > Definitions > Create
2. Click "Upload JSON file"
3. Select a workflow JSON file
4. Review and customize
5. Click "Create workflow"

### GraphQL API
```graphql
mutation CreateWorkflow($input: CreateWorkflowDefinitionInput!) {
  createWorkflowDefinition(input: $input) {
    workflowDefinition {
      id
      name
      active
    }
  }
}
```

## Customization

To customize an example:

1. Copy the example file
2. Update placeholders (e.g., `<REPLACE_WITH_GROUP_ID>`)
3. Modify fields, conditions, or actions as needed
4. Import via CLI or UI

### Finding IDs

**Group IDs**:
```bash
openlane group get
```

**User IDs**:
```bash
openlane user get
```

## Generating the Schema

The JSONSchema is auto-generated from Go types:

```bash
cd jsonschema
task workflow-schema
```

This creates `jsonschema/workflow-definition.schema.json`.

## Contributing Workflows

To contribute a new workflow example:

1. Create a new `.json` file in this directory
2. Follow the schema structure
3. Add clear descriptions and comments
4. Include placeholders for user-specific IDs
5. Document the use case in this README
6. Submit a pull request

## Resources

- [Workflow Documentation](https://docs.openlane.io/workflows)
- [CEL Language Specification](https://github.com/google/cel-spec)
- [JSONSchema Specification](https://json-schema.org/)
