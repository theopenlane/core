//go:build cli

package speccli

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
)

type testInput struct {
	Name    string
	Tags    []string
	Enabled bool
	Timeout time.Duration
	Count   int64
}

func resetKoanf() {
	cmdpkg.Config = koanf.New(".")
}

func TestBuildInputAssignsSupportedKinds(t *testing.T) {
	resetKoanf()
	cmdpkg.Config.Set("name", "example")
	cmdpkg.Config.Set("tags", []string{"alpha", "beta"})
	cmdpkg.Config.Set("enabled", true)
	cmdpkg.Config.Set("timeout", "1s")
	cmdpkg.Config.Set("count", 5)

	cmd := &cobra.Command{Use: "test"}
	fields := []FieldSpec{
		{Flag: FlagSpec{Name: "name", Required: true}, Kind: ValueString, Field: "Name"},
		{Flag: FlagSpec{Name: "tags"}, Kind: ValueStringSlice, Field: "Tags"},
		{Flag: FlagSpec{Name: "enabled"}, Kind: ValueBool, Field: "Enabled"},
		{Flag: FlagSpec{Name: "timeout"}, Kind: ValueDuration, Field: "Timeout"},
		{Flag: FlagSpec{Name: "count"}, Kind: ValueInt, Field: "Count"},
	}

	registerFieldFlags(cmd, fields)

	payload, err := buildInput(fields, reflect.TypeOf(testInput{}), cmd)
	if err != nil {
		t.Fatalf("buildInput returned error: %v", err)
	}

	input, ok := payload.(testInput)
	if !ok {
		t.Fatalf("buildInput returned unexpected type %T", payload)
	}

	if input.Name != "example" {
		t.Fatalf("expected name to be example, got %s", input.Name)
	}
	if len(input.Tags) != 2 || input.Tags[0] != "alpha" {
		t.Fatalf("unexpected tags: %#v", input.Tags)
	}
	if !input.Enabled {
		t.Fatalf("expected enabled to be true")
	}
	if input.Timeout != time.Second {
		t.Fatalf("expected timeout to be 1s, got %s", input.Timeout)
	}
	if input.Count != 5 {
		t.Fatalf("expected count to be 5, got %d", input.Count)
	}
}

func TestBuildInputMissingRequiredField(t *testing.T) {
	resetKoanf()

	cmd := &cobra.Command{Use: "test"}
	fields := []FieldSpec{
		{Flag: FlagSpec{Name: "name", Required: true}, Kind: ValueString, Field: "Name"},
	}
	registerFieldFlags(cmd, fields)

	_, err := buildInput(fields, reflect.TypeOf(testInput{}), cmd)
	if err == nil || !strings.Contains(err.Error(), "name is required") {
		t.Fatalf("expected required field error, got %v", err)
	}
}

func TestRegisterFieldFlagsAppliesDefaults(t *testing.T) {
	cmd := &cobra.Command{Use: "defaults"}
	fields := []FieldSpec{
		{Flag: FlagSpec{Name: "count", Default: "42"}, Kind: ValueInt},
		{Flag: FlagSpec{Name: "timeout", Default: "2s"}, Kind: ValueDuration},
		{Flag: FlagSpec{Name: "enabled", Default: "true"}, Kind: ValueBool},
		{Flag: FlagSpec{Name: "names", Default: "foo,bar"}, Kind: ValueStringSlice},
	}

	registerFieldFlags(cmd, fields)

	if val, err := cmd.Flags().GetInt64("count"); err != nil || val != 42 {
		t.Fatalf("expected default count 42, got %d (err=%v)", val, err)
	}
	if val, err := cmd.Flags().GetDuration("timeout"); err != nil || val != 2*time.Second {
		t.Fatalf("expected default timeout 2s, got %s (err=%v)", val, err)
	}
	if val, err := cmd.Flags().GetBool("enabled"); err != nil || val != true {
		t.Fatalf("expected default enabled true, got %t (err=%v)", val, err)
	}
	if val, err := cmd.Flags().GetStringSlice("names"); err != nil || len(val) != 2 || val[0] != "foo" {
		t.Fatalf("unexpected default names %#v (err=%v)", val, err)
	}
}

func TestRenderOutputJSONAndTable(t *testing.T) {
	defer func(prev string) { cmdpkg.OutputFormat = prev }(cmdpkg.OutputFormat)

	cols := []ColumnSpec{{Header: "ID", Path: []string{"id"}}}
	out := OperationOutput{
		Raw:     map[string]any{"id": "abc"},
		Records: []map[string]any{{"id": "abc"}},
	}

	cmd := &cobra.Command{Use: "render"}
	tableBuf := &bytes.Buffer{}
	cmd.SetOut(tableBuf)

	cmdpkg.OutputFormat = cmdpkg.TableOutput
	if err := renderOutput(cmd, out, cols); err != nil {
		t.Fatalf("renderOutput table returned error: %v", err)
	}
	if tableBuf.Len() == 0 {
		t.Fatalf("expected table output, buffer empty")
	}

	jsonBuf := &bytes.Buffer{}
	prev := cmdpkg.RootCmd.OutOrStdout()
	cmdpkg.RootCmd.SetOut(jsonBuf)
	defer cmdpkg.RootCmd.SetOut(prev)

	cmdpkg.OutputFormat = cmdpkg.JSONOutput
	if err := renderOutput(cmd, out, cols); err != nil {
		t.Fatalf("renderOutput json returned error: %v", err)
	}
	if !strings.Contains(jsonBuf.String(), "abc") {
		t.Fatalf("expected json output to contain record, got %s", jsonBuf.String())
	}
}

