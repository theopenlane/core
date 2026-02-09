package github

import "errors"

var (
	// ErrAPIRequest indicates a GitHub API request failed with a non-2xx status
	ErrAPIRequest = errors.New("github: api request failed")
	// ErrOAuthTokenMissing indicates the OAuth token is not present in the credential payload
	ErrOAuthTokenMissing = errors.New("github: oauth token missing")
	// ErrAccessTokenEmpty indicates the access token field is empty
	ErrAccessTokenEmpty = errors.New("github: access token empty")
	// ErrAuthTypeMismatch indicates the provider spec specifies an incompatible auth type
	ErrAuthTypeMismatch = errors.New("github: auth type mismatch")
	// ErrBeginAuthNotSupported indicates BeginAuth is not supported for GitHub App providers
	ErrBeginAuthNotSupported = errors.New("github: BeginAuth is not supported for GitHub App providers")
	// ErrProviderNotInitialized indicates the provider instance is nil
	ErrProviderNotInitialized = errors.New("github: provider not initialized")
	// ErrAppIDMissing indicates the GitHub App ID is missing
	ErrAppIDMissing = errors.New("github: app id missing")
	// ErrInstallationIDMissing indicates the GitHub App installation ID is missing
	ErrInstallationIDMissing = errors.New("github: installation id missing")
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
)
