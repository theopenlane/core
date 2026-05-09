// Package config handles CLI configuration loading for the integrations CLI:
// env vars, config file, and flags merged into a single koanf namespace, plus
// the typed structs that subcommands read via Unmarshal.
package config
