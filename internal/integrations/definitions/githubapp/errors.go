package githubapp

import "errors"

var (
	// ErrAPIRequest indicates a GitHub API request failed with a non-2xx status
	ErrAPIRequest = errors.New("githubapp: api request failed")
	// ErrOAuthTokenMissing indicates the OAuth access token is not present in the credential
	ErrOAuthTokenMissing = errors.New("githubapp: oauth token missing")
	// ErrAppIDMissing indicates the GitHub App ID is missing from operator config
	ErrAppIDMissing = errors.New("githubapp: app id missing")
	// ErrInstallationIDMissing indicates the GitHub App installation ID is missing
	ErrInstallationIDMissing = errors.New("githubapp: installation id missing")
	// ErrPrivateKeyMissing indicates the GitHub App private key is missing from operator config
	ErrPrivateKeyMissing = errors.New("githubapp: private key missing")
	// ErrPrivateKeyInvalid indicates the GitHub App private key could not be parsed
	ErrPrivateKeyInvalid = errors.New("githubapp: private key invalid")
	// ErrJWTSigningFailed indicates the GitHub App JWT could not be signed
	ErrJWTSigningFailed = errors.New("githubapp: jwt signing failed")
	// ErrInstallationTokenRequestFailed indicates the installation token exchange failed
	ErrInstallationTokenRequestFailed = errors.New("githubapp: installation token request failed")
	// ErrInstallationTokenEmpty indicates the installation token response was empty
	ErrInstallationTokenEmpty = errors.New("githubapp: installation token empty")
	// ErrWebhookSecretMissing indicates the GitHub App webhook secret is missing from operator config
	ErrWebhookSecretMissing = errors.New("githubapp: webhook secret missing")
	// ErrRepositoryInvalid indicates a repository identifier is not in owner/repo format
	ErrRepositoryInvalid = errors.New("githubapp: repository identifier must be owner/repo")
	// ErrOrganizationRequired indicates an organization login is required for the operation
	ErrOrganizationRequired = errors.New("githubapp: organization required")
	// ErrAppSlugMissing indicates the GitHub App slug is missing from operator config
	ErrAppSlugMissing = errors.New("githubapp: app slug missing")
	// ErrClientType indicates the provided client is not a supported type
	ErrClientType = errors.New("githubapp: unexpected client type")
)
