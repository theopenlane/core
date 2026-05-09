//go:build examples

// Package integration exposes generic integration subcommands:
// listing available provider definitions, configuring a provider
// (create or update credentials / installation metadata), and running
// provider operations. The commands are provider-agnostic — the
// integration-specific payload lives in the --body and --user-input
// JSON flags.
package integration
