package github

import "errors"

var (
	// ErrAPIRequest indicates a GitHub API request failed with a non-2xx status
	ErrAPIRequest = errors.New("github: api request failed")
	// ErrProviderMetadataRequired indicates provider metadata is required but not supplied
	ErrProviderMetadataRequired = errors.New("github: provider metadata required")
	// ErrAuthTypeMismatch indicates the provider spec specifies an incompatible auth type
	ErrAuthTypeMismatch = errors.New("github: auth type mismatch")
	// ErrBeginAuthNotSupported indicates BeginAuth is not supported for this provider
	ErrBeginAuthNotSupported = errors.New("github: BeginAuth is not supported; configure credentials via metadata")
	// ErrProviderNotInitialized indicates the provider instance is nil
	ErrProviderNotInitialized = errors.New("github: provider not initialized")
	// ErrOAuthTokenMissing indicates the OAuth token is not present in the credential payload
	ErrOAuthTokenMissing = errors.New("github: oauth token missing")
	// ErrAccessTokenEmpty indicates the access token field is empty
	ErrAccessTokenEmpty = errors.New("github: access token empty")
	// ErrAppIDMissing indicates the GitHub App ID is missing
	ErrAppIDMissing = errors.New("github: app id missing")
	// ErrInstallationIDMissing indicates the GitHub App installation ID is missing
	ErrInstallationIDMissing = errors.New("github: installation id missing")
	// ErrAppJWTMissing indicates the GitHub App JWT is missing
	ErrAppJWTMissing = errors.New("github: app jwt missing")
	// ErrPrivateKeyMissing indicates the GitHub App private key is missing
	ErrPrivateKeyMissing = errors.New("github: private key missing")
	// ErrPrivateKeyInvalid indicates the GitHub App private key could not be parsed
	ErrPrivateKeyInvalid = errors.New("github: private key invalid")
	// ErrJWTSigningFailed indicates the GitHub App JWT could not be signed
	ErrJWTSigningFailed = errors.New("github: jwt signing failed")
	// ErrInstallationTokenRequestFailed indicates the installation token request failed
	ErrInstallationTokenRequestFailed = errors.New("github: installation token request failed")
	// ErrInstallationTokenEmpty indicates the installation token response was empty
	ErrInstallationTokenEmpty = errors.New("github: installation token empty")
	// ErrWebhookSecretMissing indicates the GitHub App webhook secret is missing
	ErrWebhookSecretMissing = errors.New("github: webhook secret missing")
	// ErrAppPrivateKeyParse indicates the GitHub App private key could not be parsed
	ErrAppPrivateKeyParse = errors.New("github: app private key parse failed")
	// ErrAppJWTSign indicates signing the GitHub App JWT failed
	ErrAppJWTSign = errors.New("github: app jwt sign failed")
)
