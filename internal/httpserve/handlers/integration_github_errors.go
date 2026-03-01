package handlers

import "errors"

var (
	// ErrGitHubWebhookEventHeaderMissing indicates the GitHub webhook event header is missing
	ErrGitHubWebhookEventHeaderMissing = errors.New("github webhook event header missing")
	// ErrGitHubWebhookSignatureInvalid indicates the GitHub webhook signature is invalid
	ErrGitHubWebhookSignatureInvalid = errors.New("github webhook signature invalid")
)
