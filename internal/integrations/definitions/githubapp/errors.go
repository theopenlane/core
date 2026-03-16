package githubapp

import "errors"

var (
	// ErrAPIRequest indicates a GitHub API request failed
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
	// ErrWebhookSignatureMissing indicates the GitHub webhook signature header was not sent
	ErrWebhookSignatureMissing = errors.New("githubapp: webhook signature missing")
	// ErrWebhookBodyRead indicates the GitHub webhook body could not be read
	ErrWebhookBodyRead = errors.New("githubapp: webhook body read failed")
	// ErrWebhookSignatureMismatch indicates the GitHub webhook signature did not match
	ErrWebhookSignatureMismatch = errors.New("githubapp: webhook signature mismatch")
	// ErrAppSlugMissing indicates the GitHub App slug is missing from operator config
	ErrAppSlugMissing = errors.New("githubapp: app slug missing")
	// ErrClientType indicates the provided client is not a supported type
	ErrClientType = errors.New("githubapp: unexpected client type")
	// ErrIngestPayloadEncode indicates a collected GitHub payload could not be serialized for ingest
	ErrIngestPayloadEncode = errors.New("githubapp: ingest payload encode failed")
)
