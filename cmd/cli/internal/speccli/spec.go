//go:build cli

package speccli

import "reflect"

// ValueKind describes the expected CLI value type for a field.
type ValueKind int

const (
	// ValueString expects single string input.
	ValueString ValueKind = iota
	// ValueStringSlice expects a string slice input.
	ValueStringSlice
	// ValueBool expects a boolean input.
	ValueBool
	// ValueDuration expects a time.Duration input.
	ValueDuration
	// ValueInt expects an integer input.
	ValueInt
)

// FlagSpec defines metadata for a cobra flag.
type FlagSpec struct {
	// Name is the canonical CLI flag name.
	Name string
	// Shorthand is the optional single-letter alias.
	Shorthand string
	// Usage is the help text shown in CLI help output.
	Usage string
	// Required indicates the flag must be provided.
	Required bool
	// Default is the string representation of the default value.
	Default string
}

// FieldSpec defines how a CLI flag maps to an input struct field.
type FieldSpec struct {
	// Flag describes the cobra flag backing this field.
	Flag FlagSpec
	// Kind indicates the expected CLI value type.
	Kind ValueKind
	// Field is the struct field name in the GraphQL input.
	Field string
	// Parser converts the raw CLI value to a typed Go value.
	Parser ValueParser
}

// ValueParser converts a raw input into a typed value.
type ValueParser func(any) (any, error)

// ColumnSpec describes a column in table output.
type ColumnSpec struct {
	// Header is the column label rendered in tabular output.
	Header string
	// Path is the JSON path to the value relative to the record root.
	Path []string
	// Format is an optional formatter ID applied to the value.
	Format string
}

// ListSpec models the list command behaviour.
type ListSpec struct {
	// Method is the Openlane client method invoked for list operations.
	Method string
	// Root is the key holding list edges in the GraphQL response.
	Root string
	// FilterMethod is the client method that supports where filters.
	FilterMethod string
	// Where describes filterable fields for list operations.
	Where *WhereSpec
}

// WhereSpec describes how to build a where input for list filters.
type WhereSpec struct {
	// Type is the Go type of the GraphQL where input struct.
	Type reflect.Type
	// Fields enumerates the CLI flags that feed the where input.
	Fields []FieldSpec
}

// GetSpec models the get command behaviour.
type GetSpec struct {
	// Method is the Openlane client getter name.
	Method string
	// IDFlag is the CLI flag used to capture the resource identifier.
	IDFlag FlagSpec
	// ResultPath points to the node of interest within the GraphQL response.
	ResultPath []string
	// FallbackList toggles list execution when no ID is provided.
	FallbackList bool
	// Where enables filtered get operations when an ID is not supplied.
	Where *WhereSpec
	// ListRoot is the GraphQL key used when the GET operation returns edges.
	ListRoot string
	// Flags describes additional CLI flags consumed by the GET command.
	Flags []FieldSpec
	// PreHook runs before default GET handling, allowing custom logic.
	PreHook GetPreHook
}

// CreateSpec models the create command behaviour.
type CreateSpec struct {
	Method     string        // Method is the Openlane mutation invoked for create.
	InputType  reflect.Type  // InputType is the Go struct used for the GraphQL mutation input.
	Fields     []FieldSpec   // Fields describes the CLI flags wired into the input struct.
	ResultPath []string      // ResultPath selects the created node from the mutation response.
	PreHook    CreatePreHook // PreHook runs before default create handling.
}

// UpdateSpec models the update command behaviour.
type UpdateSpec struct {
	Method     string        // Method is the Openlane mutation invoked for update.
	IDFlag     FlagSpec      // IDFlag captures the identifier for the resource being updated.
	InputType  reflect.Type  // InputType is the Go struct used for the update input.
	Fields     []FieldSpec   // Fields describes CLI flags applied to the update payload.
	ResultPath []string      // ResultPath selects the updated node from the mutation response.
	PreHook    UpdatePreHook // PreHook runs before default update handling.
}

// PrimarySpec models behaviour when invoking the resource command directly without subcommands.
type PrimarySpec struct {
	// Flags lists CLI flags supported when the resource command is invoked
	// directly without subcommands.
	Flags []FieldSpec
	// PreHook executes when the primary command runs.
	PreHook PrimaryPreHook
}

// DeleteSpec models the delete command behaviour.
type DeleteSpec struct {
	Method      string        // Method is the Openlane mutation invoked for delete.
	IDFlag      FlagSpec      // IDFlag captures the identifier to delete.
	ResultPath  []string      // ResultPath points to the delete mutation payload.
	ResultField string        // ResultField is the key pulled into rendered output.
	PreHook     DeletePreHook // PreHook runs before default delete handling.
}

// CommandSpec defines the full lifecycle of a resource command.
type CommandSpec struct {
	// Name is the canonical resource identifier.
	Name string
	// Use is the cobra command use line.
	Use string
	// Short is the short description shown in help.
	Short string
	// Aliases holds optional command aliases.
	Aliases []string
	// List configures the list subcommand.
	List *ListSpec
	// Get configures the get subcommand.
	Get *GetSpec
	// Create configures the create subcommand.
	Create *CreateSpec
	// Update configures the update subcommand.
	Update *UpdateSpec
	// Delete configures the delete subcommand.
	Delete *DeleteSpec
	// Columns describe list/get table output.
	Columns []ColumnSpec
	// DeleteColumns describe delete table output.
	DeleteColumns []ColumnSpec
	// Primary configures top-level command behavior.
	Primary *PrimarySpec
}
