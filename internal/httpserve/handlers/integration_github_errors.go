package handlers

import "errors"

var (
	// ErrGitHubAppSlugRequired indicates the GitHub App slug is missing
	ErrGitHubAppSlugRequired = errors.New("github app slug required")
	// ErrGitHubAppIDRequired indicates the GitHub App ID is missing
	ErrGitHubAppIDRequired = errors.New("github app id required")
	// ErrGitHubAppPrivateKeyRequired indicates the GitHub App private key is missing
	ErrGitHubAppPrivateKeyRequired = errors.New("github app private key required")
	// ErrGitHubAppWebhookSecretRequired indicates the GitHub App webhook secret is missing
	ErrGitHubAppWebhookSecretRequired = errors.New("github app webhook secret required")
	// ErrGitHubWebhookEventHeaderMissing indicates the GitHub webhook event header is missing
	ErrGitHubWebhookEventHeaderMissing = errors.New("github webhook event header missing")
	// ErrGitHubWebhookSignatureInvalid indicates the GitHub webhook signature is invalid
	ErrGitHubWebhookSignatureInvalid = errors.New("github webhook signature invalid")
)
