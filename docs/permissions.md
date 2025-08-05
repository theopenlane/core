# Permissions Breakdown

This document outlines the permissions model for various object types within the system. It details who can view or edit each object, based on roles, group membership, or inheritance from parent objects. Use this as a reference to understand access levels and responsibilities for each resource.

## Mapping table based on the fga model giving more clear descriptions of access
| Object Type              | View Access                                       | Edit Access                                           |
| ------------------------ | ------------------------------------------------- | ----------------------------------------------------- |
| `organization`           | all members can view                              | admins or owners, or inherited from parent org        |
| `group`                  | all members can view, unless private              | group admins or inherited from parent org             |
| `control`                | all members can view, unless blocked              | org owner, delegate, group editor, or program access  |
| `subcontrol`             | all members can view, unless blocked              | owner, delegate, group editor, or from parent control |
| `internal_policy`        | all members can view, unless blocked              | admin, approver, delegate, or group editor            |
| `procedure`              | all members can view, unless blocked              | admin, approver, delegate, or group editor            |
| `file`                   | inherited via parent                              | inherited from parent (user, evidence, etc)           |
| `program`                | restricted — must be member or assigned           | program admin or group editor                         |
| `risk`                   | restricted — group viewer, editor, or from parent | stakeholder, delegate, or group editor                |
| `narrative`              | restricted — group viewer, editor, or from parent | group editor or org owner                             |
| `note`                   | inherited from parent                             | owner or org owner                                    |
| `evidence`               | inherited from parent                             | inherited from parent                                 |
| `task`                   | restricted — assigner, assignee, or parent        | assigner, assignee, or org owner                      |
| `template`               | all members can view                              | group editor or inherited from parent                 |
| `document_data`          | all members can view                              | group editor or inherited from parent                 |
| `contact`                | all members can view, unless blocked              | group editor or inherited from parent                 |
| `control_objective`      | all members who can view parent control           | group editor or inherited from parent                 |
| `control_implementation` | all members who can view parent control           | group editor or inherited from parent                 |
| `mapped_control`         | all members can view                              | group editor or inherited from parent                 |
| `action_plan`            | restricted — group viewer, editor, or parent      | group editor or inherited from parent                 |
| `job_runner`             | all members can view                              | editor or inherited from parent                       |
| `job_template`           | all members can view                              | editor or inherited from parent                       |
| `scheduled_job`          | all members can view                              | editor or inherited from parent                       |
| `standard`               | all members can view                              | editor or inherited from parent                       |
| `trust_center`           | all members can view                              | editor or inherited from parent                       |
| `trust_center_setting`   | all members can view                              | editor or inherited from parent                       |
| `export`                 | restricted — must be service or admin             | system admin                                          |
