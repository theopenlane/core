package githubapp

import "errors"

var (
	// ErrAPIRequest indicates a GitHub API request failed
	ErrAPIRequest = errors.New("githubapp: api request failed")
	// ErrAccessTokenMissing indicates the access token is not present in the credential
	ErrAccessTokenMissing = errors.New("githubapp: access token missing")
	// ErrCredentialDecode indicates the credential data could not be decoded
	ErrCredentialDecode = errors.New("githubapp: credential decode failed")
	// ErrCredentialEncode indicates the credential data could not be encoded
	ErrCredentialEncode = errors.New("githubapp: credential encode failed")
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
	// ErrAuthStateEncode indicates auth state could not be serialized
	ErrAuthStateEncode = errors.New("githubapp: auth state encode failed")
	// ErrAuthStateDecode indicates auth state could not be decoded
	ErrAuthStateDecode = errors.New("githubapp: auth state decode failed")
	// ErrAuthStateMismatch indicates the callback state did not match the saved auth state
	ErrAuthStateMismatch = errors.New("githubapp: auth state mismatch")
	// ErrAuthStateGenerate indicates the CSRF state token could not be generated
	ErrAuthStateGenerate = errors.New("githubapp: auth state generate failed")
	// ErrInstallationMetadataEncode indicates installation metadata could not be encoded
	ErrInstallationMetadataEncode = errors.New("githubapp: installation metadata encode failed")
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
