# taskrules

This package defines **suggested-task rules**: rules that automatically create tasks for an
organization in response to an entity being created or updated (onboarding answers submitted, an
organization created, a notification raised, and so on).

A rule has two halves that are joined by a shared **rule ID**:

1. **The rule spec** (Go) — *when* the rule fires and *what value* it fires against.
2. **The task template** (YAML) — *what task* gets created when it fires.

```
┌─ Go: entx.TaskRuleSpec ─┐         ┌─ YAML: templates/*.yaml ─┐
│ RuleID:     "framework" │◀── ID ─▶│ framework:               │
│ EachElement/Expression  │  match  │   title: ...             │
│ Trigger                 │         │   details: ...           │
└─────────────────────────┘         │   metadata: ...          │
                                     └──────────────────────────┘
```

The rule spec is attached to an ent schema field/schema via an `entx` annotation. When that field
changes and the trigger/expression matches, the engine looks up the template by rule ID
([templates.go](templates.go) `Lookup`), renders it, and creates the task. If a rule fires but no
template is registered for its ID, task creation fails with `ErrMissingTaskTemplate` — the two
halves must always stay in sync.

## Files

| File | Holds |
|---|---|
| one `*.go` file per source entity | rule specs (`[]entx.TaskRuleSpec`) and their rule-ID constants; add a new file as new source entities gain rules |
| `constants.go` | the `Source` values a template can declare |
| `templates/*.yaml` | task templates keyed by rule ID |
| `templates.go` | loads and validates the YAML at init, exposes `Lookup` |

The template YAML files are embedded into the binary (`//go:embed templates/*.yaml`) and parsed
once at package init

## How a rule fires

`entx.TaskRuleSpec` has two mutually exclusive firing modes:

- **`Expression`** — a CEL boolean evaluated against the field value. Fires **at most one** task
  when it evaluates true. Example:
  ```go
  {RuleID: RuleInviteTeam, Expression: "!has(value.personal_org) || value.personal_org == false", Trigger: entx.TaskRuleOnCreateOnly}
  ```
- **`EachElement`** — a CEL expression that evaluates to a **list**. Fires **one task per element**,
  with each element bound as `{{.value}}` (and a resolved `{{.label}}`) in the template. Example:
  ```go
  {RuleID: RuleFramework, EachElement: "value.frameworks", Trigger: entx.TaskRuleOnCreateOnly}
  ```

`Trigger` controls when it is evaluated — `TaskRuleOnCreateOnly` fires only when the entity is
first created (used everywhere here, since these fields are set once).

## Template fields

Each entry under `rules:` in a YAML file maps to a `taskrules.Template`
([templates.go](templates.go)):

| YAML key | Meaning |
|---|---|
| `title` | task title (Go-template string) |
| `details` | task body (Go-template string); keep short when `metadata.link` is set, since the user is sent to the link instead of opening the task |
| `priority` | task priority, used for ordering, higher numbers shown first |
| `taskKindName` | the kind of task to create |
| `source` | where the task came from; defaults to `openlane_recommendations` (see `constants.go`) |
| `metadata` | free-form map; the frontend recognizes the keys below |

### `metadata` keys the frontend understands

- **`link`** — if present, clicking the task does **not** open it; the user is navigated to this
  URL instead. Because of that, keep `details` short for linked tasks.
- **`docsLink`** — the URL behind the task's **View Docs / Docs** button.
- **`references`** — a list of `{name, url}` shown in the task view as additional reading.

### Full example

Every field a template can set, in one entry:

```yaml
rules:
  import-existing-controls:                  # rule ID — must match a rule spec constant
    title: Import existing controls          # required; Go-template string
    details: |                               # task body; Go-template string
      Bring in the controls you already have via CSV import.
    priority: 50                             # task priority
    taskKindName: Control Implementation     # the kind of task to create
    source: openlane_onboarding              # optional; defaults to openlane_recommendations
    metadata:                                # optional; frontend-recognized keys below
      link: /controls/import?onboarding=true # if set, task opens this URL instead of the task view
      docsLink: https://docs.theopenlane.io/docs/platform/compliance-management/controls/import
      references:                            # additional reading shown in the task view
        - name: "Getting started: From Gaps to Controls"
          url: https://docs.theopenlane.io/docs/platform/grc-fundamentals/audit/gapanalysis
        - name: Import Fields for Controls
          url: https://docs.theopenlane.io/docs/platform/compliance-management/controls/import#fields
```

`title`, `details`, `priority`, and `taskKindName` are the core task fields; `source` and
`metadata` are optional. The task's `Key` (its stable identifier) is generated automatically from
the rule ID (and the element value for `EachElement` rules)

## Templating

`title`, `details`, and every string value in `metadata` are executed as
[Go `text/template`](https://pkg.go.dev/text/template) strings when the task is rendered. Available
data:

- **`{{.value}}`** — the current element (EachElement rules only; empty for Expression rules).
- **`{{.label}}`** — a human-readable label for `.value`, resolved by an optional per-rule resolver
  (see `taskLabelResolvers` in `internal/ent/hooks/listeners_task_rules.go`); falls back to
  `.value`.
- **the firing entity's own scalar fields** — e.g. `{{.id}}`, `{{.body}}`, `{{.auditor_email}}`.

Missing keys render as empty (`missingkey=zero`), so conditional links are safe:

```yaml
link: '{{if eq .value "soc2"}}/programs/create/soc2{{else}}/programs/create/framework-based?framework={{.label}}{{end}}'
```

## Adding a new rule

1. **Pick a rule ID** and add a constant in the rule-spec `*.go` file for that source entity,
   creating a new file if the entity does not have one yet. Rule IDs are unique across *all* YAML files.

2. **Add the rule spec** to the appropriate `[]entx.TaskRuleSpec` slice, choosing `Expression`
   (fire once) or `EachElement` (fire per list item) and a `Trigger`.

3. **Wire the slice to a schema** if it is not already, via an `entx` annotation on the field or
   schema it should watch:
   ```go
   // on a field:
   Annotations(entx.FieldTaskRule(taskrules.OnboardingComplianceRules...))
   // or on a schema:
   entx.SchemaTaskRule(taskrules.OrganizationSuggestedRules...)
   ```

4. **Add the task template** under `rules:` in the matching `templates/*.yaml`, keyed by the exact
   rule ID from step 1. Fill in `title`, `details`, `priority`, `taskKindName`, and any `metadata`.

5. **(Optional) Add a label resolver** if an `EachElement` rule needs `{{.label}}` to differ from
   the raw value — register one in `taskLabelResolvers`
   (`internal/ent/hooks/listeners_task_rules.go`).

6. **Regenerate any dependent code and run the tests.** A rule spec with no matching template (or a
   template with no spec) will surface as a missing-template error when the rule fires.
