# Integration definition stub templates

Most of our integrations follow a pretty standard structure in terms of the builder, client, type registration, health check, etc., and to avoid creating all of those from scratch every time, this can be used to create a basic scaffold

```sh
task generate:integration -- oci "Oracle Cloud Infrastructure"
task generate:integration -- oci "Oracle Cloud Infrastructure" Oracle identity
```

Each `<file>.go.tmpl` becomes `../../definitions/<name>/<file>.go`. Substitution is literal text replacement
on `{{ token }}` placeholders — using sed with bash (and not go templating) so the spacing has to match exactly:

| Token | Source | Default |
|---|---|---|
| `{{ name }}` | 1st arg | required |
| `{{ display }}` | 2nd arg | name |
| `{{ family }}` | 3rd arg | display |
| `{{ category }}` | 4th arg | `security-posture` |

Existing files are never overwritten; the task skips any file already present, so re-running
after adding an operation only fills in what is missing.

## What it definitely does not do

Wiring into `internal/integrations/definitions/catalog/catalog.go` is manual: add the import
and one `<name>.Builder(),` line to `Builders()`. Two lines, and the compiler tells you if you
forget.

All other sdk / client registration, functionality, etc. Pure scaffold.
