// Package openlane wraps the Openlane API client for use inside the
// integrations CLI. Its sole responsibility is turning a resolved
// OpenlaneConfig into an authenticated *openlaneclient.Client; anything past
// that belongs in integration-specific subcommands.
package openlane
