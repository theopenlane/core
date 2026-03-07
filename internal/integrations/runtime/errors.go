package runtime

import "errors"

var (
	// ErrDBClientRequired indicates the database client dependency is missing.
	ErrDBClientRequired = errors.New("integrations runtime: db client required")
	// ErrGitHubAppProviderNotFound indicates the GitHub App provider spec was not found in the registry.
	ErrGitHubAppProviderNotFound = errors.New("integrations runtime: github app provider config not found")
)
