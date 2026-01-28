# Outstanding TODO's

- Allow for partial field applications when there are workflow eligible field modifications and ineligibile modifications that happen at the same time
high level:

  1. Add helper in internal/workflows/fields.go:
```
  // SeparateFieldsByEligibility splits fields into eligible (workflow-controlled) and
  // ineligible (pass-through) sets for a given schema type.
  func SeparateFieldsByEligibility(schemaType string, fields []string) (eligible, ineligible []string) {
      objectType := enums.ToWorkflowObjectType(schemaType)
      if objectType == nil {
          return nil, fields // Unknown type, all pass through
      }

      eligibleSet := EligibleWorkflowFields(*objectType)
      if len(eligibleSet) == 0 {
          return nil, fields // No eligible fields defined, all pass through
      }

      for _, field := range fields {
          if _, ok := eligibleSet[field]; ok {
              eligible = append(eligible, field)
          } else {
              ineligible = append(ineligible, field)
          }
      }
      return eligible, ineligible
  }
```
  2. Modify internal/ent/hooks/workflow_approval.go:
```
  // In HookWorkflowApprovalRouting, replace lines ~116-126 with:

  if len(matchingDefs) == 0 {
      return next.Mutate(ctx, m)
  }

  // Separate fields: eligible go to proposal, ineligible apply immediately
  eligibleFields, ineligibleFields := workflows.SeparateFieldsByEligibility(mut.Type(), allChangedFields)

  // Filter out system fields from ineligible (they'll be handled by the mutation naturally)
  ineligibleFields = filterNonSystemFields(ineligibleFields)

  // Apply ineligible fields immediately if any exist
  if len(ineligibleFields) > 0 {
      ineligibleChanges := workflows.BuildProposedChanges(mut, ineligibleFields)
      bypassCtx := workflows.AllowBypassContext(ctx)
      if err := workflows.ApplyObjectFieldUpdates(bypassCtx, client, *objectType, id, ineligibleChanges); err != nil {
          return nil, err
      }
  }

  // If no eligible fields remain after separation, we're done
  if len(eligibleFields) == 0 {
      // Reload entity to return updated state
      return workflows.LoadWorkflowObject(workflows.AllowContext(ctx), client, mut.Type(), id)
  }

  // Build proposed changes only from eligible fields
  proposedChanges := workflows.BuildProposedChanges(mut, eligibleFields)

  // Route eligible fields to proposals
  workflows.MarkSkipEventEmission(ctx)
  return routeMutationToProposals(ctx, client, mut, eligibleFields, proposedChanges, matchingDefs)
```
  3. Remove or repurpose validateWorkflowEligibleFields - no longer needed as a blocker.

  Key considerations:
  ┌───────────────┬─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
  │    Aspect     │                                                        Notes                                                        │
  ├───────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
  │ Recursion     │ AllowBypassContext prevents the hook from re-triggering                                                             │
  ├───────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
  │ Atomicity     │ Ineligible fields apply first, then proposal creation - if proposal fails, ineligible changes persist (acceptable?) │
  ├───────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
  │ Return value  │ Need to return entity reflecting ineligible changes applied                                                         │
  ├───────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
  │ System fields │ updated_at, updated_by etc. should flow through naturally, not be applied separately                                │
  └───────────────┴─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
- Ensure entire setup end to end works with Redis (including tests)
- Build out full support for `MANUAL_SUBMIT` mode (allow for someone to submit changes in `DRAFT` mode and make ongoing updates before submitting for approval rather than triggering approvals immediately)

┌───────────────────┬─────────────────────────────────────────────────────────────────────────────────┐
│       Area        │                                 Changes Needed                                  │
├───────────────────┼─────────────────────────────────────────────────────────────────────────────────┤
│ GraphQL Schema    │ Add queries, mutations, input/payload types                                     │
├───────────────────┼─────────────────────────────────────────────────────────────────────────────────┤
│ Resolvers         │ updateWorkflowProposalChanges, submitWorkflowProposal, withdrawWorkflowProposal │
├───────────────────┼─────────────────────────────────────────────────────────────────────────────────┤
│ Hooks             │ May need hook to handle withdrawal cleanup                                      │
├───────────────────┼─────────────────────────────────────────────────────────────────────────────────┤
│ Enums             │ Possibly add CANCELLED instance state                                           │
├───────────────────┼─────────────────────────────────────────────────────────────────────────────────┤
│ Authorization     │ Define who can view/edit/submit/withdraw proposals                              │
├───────────────────┼─────────────────────────────────────────────────────────────────────────────────┤
│ UI Feedback       │ Mechanism to inform user a draft was created                                    │
├───────────────────┼─────────────────────────────────────────────────────────────────────────────────┤
│ Object Extensions │ Way to see pending proposals on an object                                       │
└───────────────────┴─────────────────────────────────────────────────────────────────────────────────┘

- Make webhook template / replacement configurable input
- Add email notification support and sending
- Add circuit breaker for external calls in actions to prevent worker saturation
- Add ability (or expose additional schemas / requests) which would allow for easy visual indicators of what field(s) would require approval to modify and which wouldn't
- Add ability (or expose additional schemas / requests) which would be able to show the approval flow or hierarchy related to object modification before a workflow instance actually exists
- Refactor existing "job" and "scheduled job" to tie into this framework and allow for scheduling, remote job processing, etc.
- Add Delegation & Escalation
┌────────────────────────┬──────────────────────────────────────────────────────────┐
│        Feature         │                       Description                        │
├────────────────────────┼──────────────────────────────────────────────────────────┤
│ Delegation             │ Approver delegates to another user (vacation, expertise) │
├────────────────────────┼──────────────────────────────────────────────────────────┤
│ SLA-based escalation   │ Auto-escalate if no response within N hours/days         │
├────────────────────────┼──────────────────────────────────────────────────────────┤
│ Escalation chains      │ Define fallback approvers if primary is unavailable      │
├────────────────────────┼──────────────────────────────────────────────────────────┤
│ Out-of-office handling │ Automatic delegation when user is marked OOO             │
└────────────────────────┴──────────────────────────────────────────────────────────┘
- Add more robust Rejection flows
┌───────────────────┬────────────────────────────────────────────────────────────────────┐
│      Feature      │                            Description                             │
├───────────────────┼────────────────────────────────────────────────────────────────────┤
│ Rejection reasons │ Required or optional text explaining rejection                     │
├───────────────────┼────────────────────────────────────────────────────────────────────┤
│ Revise & resubmit │ Allow proposer to modify and resubmit without starting fresh       │
├───────────────────┼────────────────────────────────────────────────────────────────────┤
│ Rejection routing │ Different actions based on rejection (notify manager, create task) │
└───────────────────┴────────────────────────────────────────────────────────────────────┘
- Reminder and notification enhancements
┌────────────────────────┬──────────────────────────────────────────┐
│        Feature         │               Description                │
├────────────────────────┼──────────────────────────────────────────┤
│ Reminder notifications │ Periodic reminders for pending approvals │
├────────────────────────┼──────────────────────────────────────────┤
│ Digest notifications   │ Daily/weekly summary of pending items    │
├────────────────────────┼──────────────────────────────────────────┤
│ Channel preferences    │ Email, Slack, Teams, in-app              │
├────────────────────────┼──────────────────────────────────────────┤
│ Customizable templates │ User-defined notification content        │
└────────────────────────┴──────────────────────────────────────────┘
- More robust workflow administration capabilities
┌────────────────────────┬────────────────────────────────────────────┐
│        Feature         │                Description                 │
├────────────────────────┼────────────────────────────────────────────┤
│ Force complete/cancel  │ Admin ability to terminate stuck workflows │
├────────────────────────┼────────────────────────────────────────────┤
│ Reassign approvals     │ Move pending assignment to different user  │
├────────────────────────┼────────────────────────────────────────────┤
│ Bulk operations        │ Act on multiple instances at once          │
└────────────────────────┼────────────────────────────────────────────┘
- Workflow versioning
┌────────────────────┬─────────────────────────────────────────────────────────────────┐
│      Feature       │                           Description                           │
├────────────────────┼─────────────────────────────────────────────────────────────────┤
│ Version tracking   │ Track definition changes over time                              │
├────────────────────┼─────────────────────────────────────────────────────────────────┤
│ In-flight behavior │ Define what happens to active instances when definition changes │
├────────────────────┼─────────────────────────────────────────────────────────────────┤
│ Rollback           │ Revert to previous definition version                           │
├────────────────────┼─────────────────────────────────────────────────────────────────┤
│ A/B testing        │ Run two versions simultaneously                                 │
└────────────────────┴─────────────────────────────────────────────────────────────────┘
- Integration expansion
┌───────────────────────────┬────────────────────────────────────────────┐
│          Feature          │                Description                 │
├───────────────────────────┼────────────────────────────────────────────┤
│ Slack/Teams actions       │ Post to channels, request approval via bot │
├───────────────────────────┼────────────────────────────────────────────┤
│ External approval systems │ Integrate with ServiceNow, Jira, etc.      │
├───────────────────────────┼────────────────────────────────────────────┤
│ Custom action plugins     │ Extensible action type system              │
└───────────────────────────┴────────────────────────────────────────────┘
