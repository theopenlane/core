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
	// ErrWebhookEventMissing indicates the GitHub webhook event header was not sent
	ErrWebhookEventMissing = errors.New("githubapp: webhook event missing")
	// ErrWebhookSignatureMissing indicates the GitHub webhook signature header was not sent
	ErrWebhookSignatureMissing = errors.New("githubapp: webhook signature missing")
	// ErrWebhookSignatureMismatch indicates the GitHub webhook signature did not match
	ErrWebhookSignatureMismatch = errors.New("githubapp: webhook signature mismatch")
	// ErrAppSlugMissing indicates the GitHub App slug is missing from operator config
	ErrAppSlugMissing = errors.New("githubapp: app slug missing")
	// ErrClientType indicates the provided client is not a supported type
	ErrClientType = errors.New("githubapp: unexpected client type")
	// ErrOperationConfigInvalid indicates operation config could not be decoded
	ErrOperationConfigInvalid = errors.New("githubapp: operation config invalid")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("githubapp: result encode failed")
	// ErrAuthStartInputInvalid indicates auth start input could not be decoded
	ErrAuthStartInputInvalid = errors.New("githubapp: auth start input invalid")
	// ErrAuthCompleteInputInvalid indicates auth completion input could not be decoded
	ErrAuthCompleteInputInvalid = errors.New("githubapp: auth complete input invalid")
	// ErrAuthProviderDataEncode indicates provider data could not be serialized
	ErrAuthProviderDataEncode = errors.New("githubapp: auth provider data encode failed")
	// ErrInstallationMetadataDecode indicates installation metadata could not be decoded from credential data
	ErrInstallationMetadataDecode = errors.New("githubapp: installation metadata decode failed")
	// ErrIngestPayloadEncode indicates a collected GitHub payload could not be serialized for ingest
	ErrIngestPayloadEncode = errors.New("githubapp: ingest payload encode failed")
	// ErrWebhookPayloadInvalid indicates the webhook payload could not be decoded
	ErrWebhookPayloadInvalid = errors.New("githubapp: webhook payload invalid")
	// ErrWebhookMetadataEncode indicates webhook metadata could not be encoded
	ErrWebhookMetadataEncode = errors.New("githubapp: webhook metadata encode failed")
	// ErrWebhookPersistFailed indicates webhook side effects could not be persisted
	ErrWebhookPersistFailed = errors.New("githubapp: webhook persist failed")
	// ErrWebhookIngestFailed indicates webhook ingest failed
	ErrWebhookIngestFailed = errors.New("githubapp: webhook ingest failed")
)
