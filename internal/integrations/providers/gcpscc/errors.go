package gcpscc

import "errors"

var (
	// ErrSecurityCenterClientRequired indicates the security center client was not provided or is invalid
	ErrSecurityCenterClientRequired = errors.New("gcpscc: security center client required")
	// ErrBeginAuthNotSupported indicates BeginAuth is not supported for GCP SCC providers
	ErrBeginAuthNotSupported = errors.New("gcpscc: BeginAuth is not supported; supply metadata via credential schema")
	// ErrSubjectTokenRequired indicates the subject token is missing
	ErrSubjectTokenRequired = errors.New("gcpscc: subject token required")
	// ErrProjectIDRequired indicates the projectId is missing
	ErrProjectIDRequired = errors.New("gcpscc: projectId required")
	// ErrAudienceRequired indicates the audience is missing
	ErrAudienceRequired = errors.New("gcpscc: audience required")
	// ErrServiceAccountRequired indicates the service account email is missing
	ErrServiceAccountRequired = errors.New("gcpscc: service account email required")
	// ErrSourceIDRequired indicates the sourceId is missing
	ErrSourceIDRequired = errors.New("gcpscc: sourceId required")
	// ErrCredentialMetadataRequired indicates provider metadata is missing
	ErrCredentialMetadataRequired = errors.New("gcpscc: provider metadata required")
	// ErrAccessTokenMissing indicates the oauth token is missing from the credential payload
	ErrAccessTokenMissing = errors.New("gcpscc: oauth token missing from credential payload")
	// ErrServiceAccountKeyInvalid indicates the service account key is invalid
	ErrServiceAccountKeyInvalid = errors.New("gcpscc: service account key invalid")
	// ErrSTSInit indicates the STS client failed to initialize
	ErrSTSInit = errors.New("gcpscc: sts service init failed")
	// ErrWorkloadIdentityExchange indicates workload identity token exchange failed
	ErrWorkloadIdentityExchange = errors.New("gcpscc: workload identity token exchange failed")
	// ErrImpersonateServiceAccount indicates service account impersonation failed
	ErrImpersonateServiceAccount = errors.New("gcpscc: service account impersonation failed")
	// ErrImpersonatedTokenFetch indicates fetching the impersonated token failed
	ErrImpersonatedTokenFetch = errors.New("gcpscc: impersonated token fetch failed")
	// ErrSecurityCenterClientCreate indicates the security center client could not be created
	ErrSecurityCenterClientCreate = errors.New("gcpscc: security center client creation failed")
	// ErrMetadataDecode indicates provider metadata decoding failed
	ErrMetadataDecode = errors.New("gcpscc: provider metadata decode failed")
	// ErrSTSOptionsEncode indicates STS options encoding failed
	ErrSTSOptionsEncode = errors.New("gcpscc: sts options encode failed")
)
