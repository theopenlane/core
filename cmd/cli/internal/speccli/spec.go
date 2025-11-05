//go:build cli

package speccli

import "reflect"

// ValueKind describes the expected CLI value type for a field.
type ValueKind int

const (
	ValueString ValueKind = iota
	ValueStringSlice
	ValueBool
	ValueDuration
)

// FlagSpec defines metadata for a cobra flag.
type FlagSpec struct {
	Name      string
	Shorthand string
	Usage     string
	Required  bool
	Default   string
}

// FieldSpec defines how a CLI flag maps to an input struct field.
type FieldSpec struct {
	Flag   FlagSpec
	Kind   ValueKind
	Field  string
	Parser ValueParser
}

// ValueParser converts a raw input into a typed value.
type ValueParser func(any) (any, error)

// ColumnSpec describes a column in table output.
type ColumnSpec struct {
	Header string
	Path   []string
	Format string
}

// ListSpec models the list command behaviour.
type ListSpec struct {
	Method       string
	Root         string
	FilterMethod string
	Where        *WhereSpec
}

// WhereSpec describes how to build a where input for list filters.
type WhereSpec struct {
	Type   reflect.Type
	Fields []FieldSpec
}

// GetSpec models the get command behaviour.
type GetSpec struct {
	Method       string
	IDFlag       FlagSpec
	ResultPath   []string
	FallbackList bool
	Where        *WhereSpec
	ListRoot     string
}

// CreateSpec models the create command behaviour.
type CreateSpec struct {
	Method     string
	InputType  reflect.Type
	Fields     []FieldSpec
	ResultPath []string
	PreHook    CreatePreHook
}

// UpdateSpec models the update command behaviour.
type UpdateSpec struct {
	Method     string
	IDFlag     FlagSpec
	InputType  reflect.Type
	Fields     []FieldSpec
	ResultPath []string
}

// DeleteSpec models the delete command behaviour.
type DeleteSpec struct {
	Method      string
	IDFlag      FlagSpec
	ResultPath  []string
	ResultField string
}

// CommandSpec defines the full lifecycle of a resource command.
type CommandSpec struct {
	Name          string
	Use           string
	Short         string
	Aliases       []string
	List          *ListSpec
	Get           *GetSpec
	Create        *CreateSpec
	Update        *UpdateSpec
	Delete        *DeleteSpec
	Columns       []ColumnSpec
	DeleteColumns []ColumnSpec
}
