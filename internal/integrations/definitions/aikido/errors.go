package aikido

import "errors"

var (
	// ErrClientCredentialsMissing indicates the installation credential has no client ID or secret
	ErrClientCredentialsMissing = errors.New("aikido: client id and client secret are required")
	// ErrOAuthTokenMissing indicates the OAuth credential has no access token
	ErrOAuthTokenMissing = errors.New("aikido: oauth access token missing")
	// ErrCredentialEncode indicates the OAuth credential could not be encoded
	ErrCredentialEncode = errors.New("aikido: credential encode failed")
	// ErrRegionInvalid indicates the configured region is not a known Aikido region
	ErrRegionInvalid = errors.New("aikido: unknown region")
	// ErrClientBuild indicates the Aikido API client could not be constructed
	ErrClientBuild = errors.New("aikido: unable to build api client")
	// ErrUnauthorized indicates Aikido rejected the client credentials or the token lacks a required scope
	ErrUnauthorized = errors.New("aikido: client credentials were rejected or are missing a required read scope")
	// ErrRateLimited indicates Aikido kept rate limiting the request after retries
	ErrRateLimited = errors.New("aikido: rate limited by the api")
	// ErrAPIRequest indicates the Aikido API returned an unexpected response
	ErrAPIRequest = errors.New("aikido: unexpected api response")
	// ErrPageLimit indicates a listing reached the page cap, so its results would be incomplete
	ErrPageLimit = errors.New("aikido: reached the maximum page count before collecting every record")
	// ErrPayloadEncode indicates a record could not be encoded into an ingest envelope
	ErrPayloadEncode = errors.New("aikido: unable to encode payload")
	// ErrIssuesFetchFailed indicates issues could not be exported
	ErrIssuesFetchFailed = errors.New("aikido: unable to export issues")
	// ErrIssueGroupsFetchFailed indicates the issue groups carrying descriptions could not be listed
	ErrIssueGroupsFetchFailed = errors.New("aikido: unable to list issue groups")
	// ErrAutofixFetchFailed indicates the autofix history or the issues a task targets could not be listed
	ErrAutofixFetchFailed = errors.New("aikido: unable to list autofix tasks")
	// ErrDirectoryFetchFailed indicates the users, teams or team members could not be listed
	ErrDirectoryFetchFailed = errors.New("aikido: unable to list users and teams")
	// ErrComplianceFetchFailed indicates a compliance report could not be fetched
	ErrComplianceFetchFailed = errors.New("aikido: unable to fetch compliance reports")
	// ErrScansFetchFailed indicates the pentest or code analysis assessments could not be listed
	ErrScansFetchFailed = errors.New("aikido: unable to list assessments")
	// ErrAssetsFetchFailed indicates an asset inventory listing failed
	ErrAssetsFetchFailed = errors.New("aikido: unable to list assets")
	// ErrWorkspaceFetchFailed indicates the workspace details could not be retrieved
	ErrWorkspaceFetchFailed = errors.New("aikido: unable to fetch workspace details")
	// ErrWorkspaceIDMissing indicates the workspace response carried no ID for the installation identity
	ErrWorkspaceIDMissing = errors.New("aikido: workspace id missing")
)
