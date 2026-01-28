# Workflow Engine Architecture

This document explains how the workflow engine operates within the Openlane platform, including the flow of data through hooks, events, and the engine itself.

## Overview

The workflows capability is built on simple composable primitives that combine into complex behaviors:

| Primitive         | Purpose            | Example                        |
|-------------------|--------------------|--------------------------------|
| Trigger           | When to evaluate   | field modified, edge added     |
| Condition         | Whether to proceed | object.status == "DRAFT"       |
| When (per action) | Action-level gate  | assignments.approved >= 3      |
| Action            | What to do         | APPROVAL, NOTIFY, FIELD_UPDATE |
| Operators         | Composition        | field_modified AND is_tuesday  |

The engine evaluates triggers → checks conditions → executes actions (respecting their `When` expressions, e.g. blocking on approvals until criteria met).

Any meaningful sequential vs concurrent distinction around execution emerges from how you structure the definition, not from engine constraints:
- One approval action with multiple targets + quorum = concurrent
- Multiple approval actions = sequential with dependencies
- When expressions = conditional execution based on state

There is an action index as a part of the executor, but it is just tracking progression through the action list, with approvals creating natural "gates" that block until their criteria are satisfied. The engine doesn't impose ordering beyond that - the definition author controls the flow through how they structure actions and their conditions.

## High level Architecture

You can simplify all of the moving parts related to workflows into 2 buckets:
- Bucket 1: pre-commit interception, where we fork into a "proposed" change
- Bucket 2: post-commit events, where we're responding after a transaction

### Pre-Commit Interception

For workflows where we intercept the change before its applied :

```
User Request ──► Hook intercepts ──► Creates WorkflowProposal ──► Returns unchanged entity
                     │
                     ▼
            Proposal submitted ──► WorkflowInstance created ──► Assignments created
                                                                      │
                                                                      ▼
                                                              Users approve/reject
                                                                      │
                                                                      ▼
                                                              Quorum satisfied?
                                                                      │
                                          ┌───────────────────────────┴───────────┐
                                          │                                       │
                                          ▼                                       ▼
                                    YES: Apply                              NO: Fail
                                    proposal changes                        workflow
                                    to entity
```

### Post-Commit Events

For workflows that react to changes:

```
User Request ──► Mutation commits ──► soiree event emitted ──► handleWorkflowMutation
                                                                      │
                                                                      ▼
                                                              Find matching definitions
                                                                      │
                                                                      ▼
                                                              Trigger workflows
                                                              (NOTIFY, WEBHOOK, etc.)
```

## Key Components

### Hooks

| Hook | Purpose | Trigger |
|------|---------|---------|
| `HookWorkflowApprovalRouting` | Intercepts mutations, routes to proposal if approval required | UPDATE on workflowable schemas |
| `HookWorkflowProposalTriggerOnSubmit` | Starts workflow when proposal submitted | CREATE/UPDATE on WorkflowProposal |
| `HookWorkflowProposalInvalidateAssignments` | Invalidates approvals when proposal changes edited | UPDATE on WorkflowProposal |
| `HookWorkflowDefinitionPrefilter` | Derives trigger prefilter fields from definition JSON | CREATE/UPDATE on WorkflowDefinition |

### Approval Submission Modes

Workflow definitions with approval actions have an `approvalSubmissionMode` that controls when the approval process begins:

| Mode | Proposal Initial State | Behavior |
|------|----------------------|----------|
| `AUTO_SUBMIT` | `SUBMITTED` | Approval assignments are created immediately when the mutation is intercepted. Approvers are notified right away. This is the standard flow. |
| `MANUAL_SUBMIT` | `DRAFT` | The proposal is created in DRAFT state. Assignments are NOT created until the proposal is explicitly submitted. |

**Current Implementation Note:**
`AUTO_SUBMIT` is the primary supported mode. The `MANUAL_SUBMIT` mode creates proposals in DRAFT state, but the GraphQL mutation to submit draft proposals is not yet exposed (`WorkflowProposal` is an internal entity with `entgql.Skip(entgql.SkipAll)`). When MANUAL_SUBMIT support is completed, it will enable staging changes before requesting approval.

### Definition Prefiltering

WorkflowDefinition persists derived trigger fields (`trigger_operations`, `trigger_fields`) for coarse SQL prefiltering. These are computed from `definition_json` by `HookWorkflowDefinitionPrefilter` and used by `FindMatchingDefinitions`
to reduce the candidate set before reading the full JSON and evaluating triggers.

Key properties:
- Intentionally lossy: `trigger_fields` is a union across triggers and does not preserve per-trigger pairing.
- Selectors and trigger expressions are not represented.
- If any trigger omits fields/edges, `trigger_fields` is cleared (nil) to avoid false negatives.
- The JSON definition remains the source of truth; prefiltering only excludes impossible matches.

Safe exclusion criteria:
- `eventType` not in `trigger_operations`.
- `trigger_fields` is non-empty and has no overlap with changed fields/edges.

### Event Listeners

| Listener | Event | Purpose |
|----------|-------|---------|
| `HandleWorkflowTriggered` | WorkflowTriggered | Process newly triggered instance, start first action |
| `HandleActionStarted` | ActionStarted | Execute the action (approval, notify, etc.) |
| `HandleActionCompleted` | ActionCompleted | Advance to next action or complete workflow |
| `HandleAssignmentCompleted` | AssignmentCompleted | Check quorum, resume workflow if satisfied |
| `HandleInstanceCompleted` | InstanceCompleted | Mark instance as completed/failed |

### Workflow Engine

| Component | Purpose |
|-----------|---------|
| `engine.go` | Core engine: FindMatchingDefinitions, TriggerWorkflow, ProcessAction |
| `executor.go` | Action executors: APPROVAL, NOTIFY, FIELD_UPDATE, WEBHOOK, etc |
| `evaluator.go` | CEL expression evaluation for conditions and When clauses |
| `resolver.go` | Target resolution (users, groups, custom resolvers) |
| `listeners.go` | Event handlers and workflow state management |

## Approval Invalidation

When a WorkflowProposal's changes are modified after approvals have been given:

1. `HookWorkflowProposalInvalidateAssignments` detects the change
1. All APPROVED assignments are reset to PENDING
1. Invalidation metadata is recorded (who, when, hash diff)
1. Affected users are notified
1. Re-approval is required

This implements GitHub-style "dismiss stale reviews" behavior.

## Adding Workflow Support to a Schema

1. Add the `WorkflowApprovalMixin` to your schema:

```go
func (Policy) Mixin() []ent.Mixin {
    return []ent.Mixin{
        // ... other mixins
        WorkflowApprovalMixin{},
    }
}
```

1. Register eligible fields in the workflow metadata (use the entx annotation)
1. Add edges from WorkflowObjectRef schema if they don't already exist
1. Run code generation / `task regenerate`, merge the output / changes
1. Create workflow definitions targeting your schema type and test it out

## Workflow Definition Structure

There are a few examples included in this package, and you can find the types in our models if you want to look at all the various fields, but here's a representative example of policy approvals:

```json
{
  "name": "Policy Approval Workflow",
  "schemaType": "Policy",
  "workflowKind": "APPROVAL",
  "approvalSubmissionMode": "AUTO_SUBMIT",
  "triggers": [
    {
      "operation": "UPDATE",
      "fields": ["description", "details"],
      "expression": "object.status == 'PUBLISHED'"
    }
  ],
  "conditions": [
    {
      "expression": "object.category != 'INTERNAL'"
    }
  ],
  "actions": [
    {
      "key": "manager_approval",
      "type": "APPROVAL",
      "params": {
        "targets": [{"type": "RESOLVER", "resolver_key": "object_owner"}],
        "required_count": 1
      }
    },
    {
      "key": "notify_team",
      "type": "NOTIFICATION",
      "when": "assignments.approved >= 1",
      "params": {
        "targets": [{"type": "GROUP", "id": "policy-team-group-id"}],
        "title": "Policy {{object.name}} approved",
        "body": "The policy has been approved and changes applied."
      }
    }
  ]
}
```

## Events and Event Topics

These are the “statuses” stored in workflow_events.event_type and payload.event_type. We only persist
business-facing snapshots plus emit failure tracking. Runtime flow still uses soiree topics
(WorkflowTriggered, ActionStarted, AssignmentCompleted, etc.), but not all topics are persisted.

| WorkflowEventType | Written by (code) | Meaning / status | Readers / relies-on |
|---|---|---|---|
| WORKFLOW_TRIGGERED | HandleWorkflowTriggered (internal/workflows/engine/eventhandlers.go) | Snapshot of the trigger context (definition_id, object, trigger) | UI timeline; GraphQL filters |
| ASSIGNMENT_CREATED | recordAssignmentsCreated (internal/workflows/engine/emit.go) | Batch approval assignment creation (assignment_ids, target_user_ids, required_count) | UI timeline; GraphQL filters |
| ACTION_COMPLETED | HandleActionCompleted + HandleAssignmentCompleted | Action outcome snapshot. Details include success, skipped, error_message and (for approvals) counts + assignment ids | UI timeline; tests |
| WORKFLOW_COMPLETED | HandleInstanceCompleted | Final instance state (COMPLETED/FAILED) | UI timeline; tests |
| EMIT_FAILED | recordEmitFailure (internal/workflows/engine/emit.go) | Emit enqueue failure; stores EmitFailureDetails | Reconciler retries |
| EMIT_RECOVERED | Reconciler.updateWorkflowEvent | Emit retry succeeded | Tests / GraphQL filters only |
| EMIT_FAILED_TERMINAL | Reconciler.markTerminal | Emit retry exhausted; instance state set to Failed | Tests / GraphQL filters only |

Legacy event types like ACTION_STARTED, ACTION_FAILED, ACTION_SKIPPED, ASSIGNMENT_RESOLVED,
INSTANCE_PAUSED, and INSTANCE_RESUMED are no longer persisted; their outcomes are represented
via ACTION_COMPLETED details or are purely runtime concerns.

---

## Detailed Architecture Diagrams

The following mermaid diagrams illustrate the various flows and permutations supported by the workflow system.

### Pre-Commit Approval Flow (Detailed)

This diagram shows what happens when a user updates a field that requires approval.

```mermaid
sequenceDiagram
    autonumber
    participant User
    participant API as GraphQL API
    participant Hook as HookWorkflowApprovalRouting
    participant Engine as WorkflowEngine
    participant DB as Database
    participant Bus as Event Bus
    participant Listener as Event Listeners

    User->>API: UpdateControl(reference_id: "NEW-123")
    API->>Hook: Pre-commit hook fires

    Note over Hook: Check for workflow bypass context
    Hook->>Engine: FindMatchingDefinitions(Control, UPDATE, [reference_id])
    Engine->>DB: Query WorkflowDefinition (prefiltered)
    DB-->>Engine: Matching definitions with APPROVAL actions

    alt Has approval actions for these fields
        Hook->>DB: Create WorkflowProposal(state: DRAFT/SUBMITTED)
        Hook->>DB: Create WorkflowObjectRef(control_id)
        Hook->>DB: Create WorkflowInstance(state: RUNNING, proposal_id)

        Note over Hook: Strip workflow fields from mutation
        Hook-->>API: Return entity with ORIGINAL values

        alt Auto-Submit Mode
            Hook->>Bus: Emit WorkflowTriggered
            Bus->>Listener: HandleWorkflowTriggered
            Listener->>Engine: Process instance, create assignments
            Engine->>DB: Create WorkflowAssignment(s)
            Engine->>DB: Create WorkflowAssignmentTarget(s)
            Engine->>DB: Update instance state: PAUSED
        else Manual Submit Mode
            Note over Hook: Proposal stays in DRAFT
            Note over Hook: User must explicitly submit
        end
    else No approval actions
        Hook-->>API: Proceed with mutation normally
        API->>DB: Commit changes
        API->>Bus: Emit mutation event
    end
```

### Multi-Approver Quorum Flow

When a workflow requires multiple approvers with a quorum threshold.

```mermaid
sequenceDiagram
    autonumber
    participant User1 as Approver 1
    participant User2 as Approver 2
    participant User3 as Approver 3
    participant API as GraphQL API
    participant Engine as WorkflowEngine
    participant DB as Database

    Note over DB: Proposal requires 2 of 3 approvers<br/>(required_count: 2)

    rect rgb(240, 248, 255)
        Note right of DB: Initial State: 3 PENDING assignments
    end

    User1->>API: Approve assignment
    API->>Engine: CompleteAssignment(status: APPROVED)
    Engine->>DB: Update assignment status: APPROVED
    Engine->>Engine: Check quorum: 1 approved, 2 pending

    rect rgb(255, 255, 224)
        Note right of Engine: Quorum NOT met (1 < 2)<br/>Workflow remains PAUSED
    end

    User2->>API: Approve assignment
    API->>Engine: CompleteAssignment(status: APPROVED)
    Engine->>DB: Update assignment status: APPROVED
    Engine->>Engine: Check quorum: 2 approved, 1 pending

    rect rgb(224, 255, 224)
        Note right of Engine: Quorum MET (2 >= 2)<br/>Apply proposal changes
    end

    Engine->>DB: Apply proposal to Control
    Engine->>DB: Update proposal state: APPLIED
    Engine->>DB: Update instance state: COMPLETED

    rect rgb(245, 245, 245)
        Note right of DB: User3's assignment remains PENDING<br/>(late approval will be no-op)
    end

    User3->>API: Approve assignment (late)
    API->>Engine: CompleteAssignment(status: APPROVED)
    Engine->>DB: Update assignment status: APPROVED
    Note over Engine: Workflow already completed<br/>No additional effects
```

### Multiple Concurrent Workflow Instances

When a single mutation triggers multiple workflow definitions.

```mermaid
flowchart TB
    subgraph "Single Mutation"
        Mutation["UpdateControl(reference_id, status)"]
    end

    subgraph "Matching Definitions"
        Def1["Definition 1<br/>Fields: [reference_id]<br/>Approvers: Team A"]
        Def2["Definition 2<br/>Fields: [status]<br/>Approvers: Team B"]
        Def3["Definition 3<br/>Fields: [reference_id, status]<br/>Approvers: Compliance"]
    end

    subgraph "Created Instances"
        Inst1["Instance 1<br/>Domain: reference_id<br/>State: PAUSED"]
        Inst2["Instance 2<br/>Domain: status<br/>State: PAUSED"]
        Inst3["Instance 3<br/>Domain: reference_id,status<br/>State: PAUSED"]
    end

    subgraph "Proposals"
        Prop1["Proposal 1<br/>Changes: {reference_id: NEW}"]
        Prop2["Proposal 2<br/>Changes: {status: APPROVED}"]
        Prop3["Proposal 3<br/>Changes: {reference_id: NEW, status: APPROVED}"]
    end

    subgraph "Assignments"
        Assign1A["Assignment: Team A Member 1"]
        Assign1B["Assignment: Team A Member 2"]
        Assign2A["Assignment: Team B Member 1"]
        Assign3A["Assignment: Compliance Officer"]
    end

    Mutation --> Def1
    Mutation --> Def2
    Mutation --> Def3

    Def1 --> Inst1
    Def2 --> Inst2
    Def3 --> Inst3

    Inst1 --> Prop1
    Inst2 --> Prop2
    Inst3 --> Prop3

    Inst1 --> Assign1A
    Inst1 --> Assign1B
    Inst2 --> Assign2A
    Inst3 --> Assign3A
```

### Domain-Based Approval Isolation

Different field sets create separate approval domains, allowing concurrent workflows on the same object.

```mermaid
flowchart TB
    subgraph "Control Object"
        Control["Control CTL-001"]
        Field1["reference_id: REF-OLD"]
        Field2["status: DRAFT"]
        Field3["category: Technical"]
    end

    subgraph "Domain 1: Reference ID Changes"
        D1Def["Definition: Ref ID Approval<br/>Fields: [reference_id]"]
        D1Prop["Proposal 1<br/>domain_key: Control:reference_id<br/>changes: {reference_id: REF-NEW}"]
        D1Inst["Instance 1<br/>State: PAUSED"]
        D1Assign["Approver: Finance Team"]
    end

    subgraph "Domain 2: Status Changes"
        D2Def["Definition: Status Approval<br/>Fields: [status]"]
        D2Prop["Proposal 2<br/>domain_key: Control:status<br/>changes: {status: APPROVED}"]
        D2Inst["Instance 2<br/>State: PAUSED"]
        D2Assign["Approver: Compliance Team"]
    end

    Control --> D1Def
    Control --> D2Def

    D1Def --> D1Inst --> D1Prop
    D1Inst --> D1Assign

    D2Def --> D2Inst --> D2Prop
    D2Inst --> D2Assign

    Note1["These workflows proceed independently.<br/>Approving one does not affect the other."]
```

### Post-Commit Notification Flow

For workflows that react to changes (no approval required).

```mermaid
sequenceDiagram
    autonumber
    participant User
    participant API as GraphQL API
    participant DB as Database
    participant Bus as Event Bus
    participant Listener as HandleWorkflowMutation
    participant Engine as WorkflowEngine
    participant Webhook as External Webhook

    User->>API: UpdateControl(status: APPROVED)
    API->>DB: Commit changes
    API->>Bus: Emit mutation event

    Bus->>Listener: HandleWorkflowMutation(payload)
    Listener->>Engine: FindMatchingDefinitions(Control, UPDATE, [status])
    Engine-->>Listener: [NotificationWorkflowDef]

    loop For each matching definition
        Listener->>Engine: TriggerWorkflow(def, object)
        Engine->>DB: Create WorkflowInstance(state: RUNNING)

        Note over Engine: Process actions sequentially

        Engine->>Engine: Execute NOTIFICATION action
        Engine->>DB: Record notification sent

        Engine->>Engine: Execute WEBHOOK action
        Engine->>Webhook: POST /webhook (payload)
        Webhook-->>Engine: 200 OK

        Engine->>DB: Update instance state: COMPLETED
    end
```

### Parallel Approval Actions

When a workflow has multiple independent approval actions that execute concurrently.

```mermaid
sequenceDiagram
    autonumber
    participant Hook as Trigger Hook
    participant Engine as WorkflowEngine
    participant DB as Database
    participant User1 as Legal Approver
    participant User2 as Finance Approver

    Hook->>Engine: TriggerWorkflow(multi-approval-def)
    Engine->>DB: Create WorkflowInstance

    Note over Engine: Detect multiple approval actions<br/>with independent "when" conditions

    par Create all approval assignments concurrently
        Engine->>DB: Create Assignment (Legal Review)
        Engine->>DB: Create AssignmentTarget (User1)
    and
        Engine->>DB: Create Assignment (Finance Review)
        Engine->>DB: Create AssignmentTarget (User2)
    end

    Engine->>DB: Store parallel_approval_keys in instance context
    Engine->>DB: Update instance state: PAUSED

    Note over DB: Both approvals must complete<br/>before workflow can proceed

    User1->>Engine: Approve (Legal Review)
    Engine->>DB: Update assignment: APPROVED
    Engine->>Engine: Check: All parallel approvals satisfied?
    Note over Engine: NO - Finance still pending

    User2->>Engine: Approve (Finance Review)
    Engine->>DB: Update assignment: APPROVED
    Engine->>Engine: Check: All parallel approvals satisfied?
    Note over Engine: YES - Both approved

    Engine->>DB: Apply proposal changes
    Engine->>DB: Update instance state: COMPLETED
```

### Approval Invalidation Flow

When a proposal is edited after approvals have been given.

```mermaid
sequenceDiagram
    autonumber
    participant Editor as Proposal Editor
    participant Hook as HookWorkflowProposalInvalidateAssignments
    participant DB as Database
    participant Approver as Previous Approver

    Note over DB: Proposal State: SUBMITTED<br/>Assignment 1: APPROVED (hash: abc123)<br/>Assignment 2: PENDING

    Editor->>DB: Update proposal changes
    DB->>Hook: Pre-commit hook fires

    Hook->>Hook: Detect state == SUBMITTED
    Hook->>Hook: Compare old hash vs new hash
    Note over Hook: Hash changed: abc123 != def456

    Hook->>DB: Query APPROVED assignments

    loop For each APPROVED assignment
        Hook->>DB: Reset status to PENDING
        Hook->>DB: Record invalidation metadata
        Note over DB: {reason: "proposal changes edited",<br/>previous_hash: abc123,<br/>invalidated_at: now}
    end

    Hook->>DB: Commit proposal update

    Note over Approver: Notification: "Your approval<br/>has been invalidated due to<br/>proposal changes. Please re-review."
```

### Concurrent Approval Race Condition Handling

How the engine handles concurrent approval submissions.

```mermaid
sequenceDiagram
    autonumber
    participant User1 as Approver 1
    participant User2 as Approver 2
    participant Engine as WorkflowEngine
    participant DB as Database

    Note over DB: Workflow requires 2 approvals<br/>Both users click "Approve" simultaneously

    par Concurrent approval requests
        User1->>Engine: CompleteAssignment(assignment1, APPROVED)
    and
        User2->>Engine: CompleteAssignment(assignment2, APPROVED)
    end

    Note over Engine: Race condition: both check quorum

    Engine->>DB: Update assignment1: APPROVED
    Engine->>DB: Update assignment2: APPROVED

    par Both check quorum satisfaction
        Engine->>Engine: Check quorum (Thread 1)
        Note over Engine: 2 approved >= 2 required<br/>Attempt to apply proposal
    and
        Engine->>Engine: Check quorum (Thread 2)
        Note over Engine: 2 approved >= 2 required<br/>Attempt to apply proposal
    end

    Note over Engine: Idempotent completion logic

    Engine->>DB: Apply proposal (Thread 1 wins)
    Engine->>DB: Update instance: COMPLETED

    Note over Engine: Thread 2 sees instance already COMPLETED<br/>No duplicate application

    Engine->>DB: Record single INSTANCE_COMPLETED event
```

### Event Emission and Recovery

How the system handles event emission failures and recovery.

```mermaid
sequenceDiagram
    autonumber
    participant Engine as WorkflowEngine
    participant Bus as Event Bus
    participant DB as Database
    participant Reconciler as Reconciler

    Engine->>Bus: Emit(WorkflowTriggered)

    alt Emit succeeds
        Bus-->>Engine: Success
        Note over Engine: Continue normally
    else Emit fails (queue unavailable)
        Bus-->>Engine: Error: enqueue failed
        Engine->>DB: Record WorkflowEvent<br/>type: EMIT_FAILED<br/>details: {topic, payload, attempts: 1}
        Note over Engine: Workflow instance created<br/>but event not published
    end

    Note over Reconciler: Periodic reconciliation job

    Reconciler->>DB: Query EMIT_FAILED events
    DB-->>Reconciler: [failed_event_1, ...]

    loop For each failed event
        Reconciler->>Bus: Retry emit(payload)

        alt Retry succeeds
            Bus-->>Reconciler: Success
            Reconciler->>DB: Record EMIT_RECOVERED event
            Reconciler->>DB: Delete EMIT_FAILED event
        else Retry fails
            alt attempts < max_attempts
                Reconciler->>DB: Update EMIT_FAILED<br/>attempts: attempts + 1
            else attempts >= max_attempts
                Reconciler->>DB: Record EMIT_FAILED_TERMINAL
                Reconciler->>DB: Update instance state: FAILED
            end
        end
    end
```


### Field Eligibility and Proposal Routing

How the system determines which fields can be workflow-controlled.

```mermaid
flowchart TB
    subgraph "Schema Definition"
        Schema["Control Schema<br/>with WorkflowApprovalMixin"]
        Fields["Fields:<br/>- reference_id (workflow_eligible: true)<br/>- status (workflow_eligible: true)<br/>- description (workflow_eligible: false)<br/>- created_at (workflow_eligible: false)"]
    end

    subgraph "Mutation Request"
        Mutation["UpdateControl(<br/>  reference_id: NEW,<br/>  description: Updated<br/>)"]
    end

    subgraph "Hook Analysis"
        Check1{All changed fields<br/>workflow eligible?}
        Check2[Eligible: reference_id]
        Check3[Ineligible: description]
    end

    subgraph "Routing Decision"
        Route1["ERROR: ErrWorkflowIneligibleField<br/>Cannot mix eligible and ineligible<br/>fields in same mutation"]
        Route2["Route to WorkflowProposal<br/>Changes: {reference_id: NEW}"]
        Route3["Apply directly<br/>(no workflow)"]
    end

    Schema --> Fields
    Mutation --> Check1
    Check1 -->|Mixed| Route1
    Check1 -->|All eligible| Check2 --> Route2
    Check1 -->|All ineligible| Check3 --> Route3
```
