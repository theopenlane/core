# GenCMD

This package is used to generate the base CRUD operations for a cli command for the `OpenlaneClient`

## Usage

The [Taskfile](Taskfile.yaml) includes two commands:

```
* cli:generate:                   generates a new cli cmd
* cli:generate:ro:                generates a new cli cmd with only the read cmds
```

To use the cli directly instead of the `Taskfile`:

```
go run pkg/gencmd/generate/main.g generate --help
generate is the command to generate the stub files for a given cli cmd

Usage:
   generate [flags]

Flags:
  -d, --dir string    root directory location to generate the files (default "cmd")
  -h, --help          help for generate
  -i, --interactive   interactive prompt, set to false to disable (default true)
  -n, --name string   name of the command to generate
  -r, --read-only     only generate the read only commands, no create, update or delete commands
 ```

### Generate All CRUD Operations

1. Run the generate command
    ```
    task cli:generate
    task: [cli:generate] go run gencmd/generate/main.go generate
    Name of the command (should be the singular version of the object): Contact
    ----> creating cli cmd for: Contact
    ----> executing template: create.tmpl
    ----> executing template: delete.tmpl
    ----> executing template: doc.tmpl
    ----> executing template: get.tmpl
    ----> executing template: root.tmpl
    ----> executing template: update.tmpl
    ```
1. This will create a set of cobra command files:
    ```
    ls -l cli/cmd/contact/
    total 48
    -rw-r--r--  1 sarahfunkhouser  staff  1261 Jun 27 13:30 create.go
    -rw-r--r--  1 sarahfunkhouser  staff  1086 Jun 27 13:30 delete.go
    -rw-r--r--  1 sarahfunkhouser  staff   102 Jun 27 13:30 doc.go
    -rw-r--r--  1 sarahfunkhouser  staff  1079 Jun 27 13:30 get.go
    -rw-r--r--  1 sarahfunkhouser  staff  2570 Jun 27 13:30 root.go
    -rw-r--r--  1 sarahfunkhouser  staff  1511 Jun 27 13:30 update.go
    ```
1. Add the new package to `cli/main.go`, in this case it would be:
    ```go
    	_ "github.com/theopenlane/core/cli/cmd/contact"
    ```
1. Add flags for the `Create` and `Update` commands for input
1. Add validation to the `createValidation()` and `updateValidation()` functions for the required input
1. Add fields for the table output in `root.go` in `tableOutput()` function, by default it only includes `ID`.
    ```go
    // tableOutput prints the plans in a table format
    func tableOutput(plans []graphclient.Contact) {
        writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Email", "PhoneNumber")

        for _, p := range plans {
            writer.AddRow(p.ID, p.Name, *p.Email, *p.PhoneNumber)
        }

        writer.Render()
    }
    ```

### Generate Read-Only Operations

A common use-case for some of the resolvers is a `read-only` set of functions, particularly in our `History` schemas. To generate the cmds with **only** the `get` functionality use the `--read-only` flag:

1. Run the read-only generate command
    ```
    task cli:generate:ro
    ```
1. The resulting commands will just be a `get` command
    ```
    task: [cli:generate:ro] go run gencmd/generate/main.go generate --read-only
    Name of the command (should be the singular version of the object): ContactHistory
    ----> creating cli cmd for: ContactHistory
    ----> executing template: doc.tmpl
    ----> executing template: get.tmpl
    ----> executing template: root.tmpl
    ```