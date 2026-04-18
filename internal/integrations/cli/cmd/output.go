package cmd

import (
	"github.com/theopenlane/core/internal/integrations/cli/config"
)

// OutputFormat returns the resolved output format (json or table)
func OutputFormat() string {
	return Config.String("output.format")
}

// RenderTable prints the supplied headers + rows when the output format is
// table; otherwise marshals out as JSON. Subcommands pass a payload suited to
// each format so neither has to preserve the other's state.
func RenderTable(out any, headers []string, rows [][]string) error {
	if OutputFormat() == JSONOutput {
		return config.JSONOutput(out)
	}

	return config.TableOutput(headers, rows)
}

// RenderJSON prints out as indented JSON regardless of the configured format.
// Use for payloads that don't have a meaningful table projection.
func RenderJSON(out any) error {
	return config.JSONOutput(out)
}

// BoolStr renders a bool as "true"/"false" for table cells
func BoolStr(b bool) string {
	if b {
		return "true"
	}

	return "false"
}

// StrPtr dereferences a *string for table cells; nil becomes ""
func StrPtr(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

// BoolPtrStr dereferences a *bool for table cells; nil becomes ""
func BoolPtrStr(b *bool) string {
	if b == nil {
		return ""
	}

	return BoolStr(*b)
}
