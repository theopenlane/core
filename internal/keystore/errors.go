package keystore

import "errors"

var (
	errAPIKeyMintingUnsupported         = errors.New("apikey integrations do not support token minting")
	errNoCredentialsProduced            = errors.New("no credentials produced")
	errMissingGitHubAppCredentials      = errors.New("missing github app credentials (private_key, installation_id, app_id)")
	errWorkloadIdentityAudienceMissing  = errors.New("workload identity audience missing")
	errTargetServiceAccountEmailMissing = errors.New("target service account email missing")
	errInvalidGitHubAppPrivateKey       = errors.New("invalid github app private key")
	errPrivateKeyNotRSA                 = errors.New("private key is not RSA")
	errUnsupportedAuthType              = errors.New("unsupported auth type")
	errGitHubInstallationTokenExchange  = errors.New("github installation token exchange failed")
	errSTSTokenExchange                 = errors.New("sts token exchange failed")
	errAWSFederationConfigurationMissing = errors.New("aws federation configuration missing required parameters")
	errAWSWebIdentityTokenMissing        = errors.New("aws federation web identity token missing")

	errNoUserInfoSpec         = errors.New("user info spec missing")
	errMissingUserInfoSpec    = errors.New("user info request requires auth header definition")
	errUserInfoHTTPFailure    = errors.New("user info request failed")
	errSecondaryEmailFailure  = errors.New("secondary email lookup failed")
)
