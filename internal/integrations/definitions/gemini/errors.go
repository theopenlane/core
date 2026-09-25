package gemini

import "errors"

var (
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("gemini: result encode failed")
	// ErrOperationConfigInvalid indicates operation configuration could not be decoded
	ErrOperationConfigInvalid = errors.New("gemini: operation config invalid")
	// ErrRuntimeConfigDecode indicates the runtime config could not be decoded
	ErrRuntimeConfigDecode = errors.New("gemini: runtime config decode failed")
	// ErrRuntimeConfigInvalid indicates the runtime config is missing required fields
	ErrRuntimeConfigInvalid = errors.New("gemini: runtime config invalid")
	// ErrRuntimeOnly indicates the client was requested for an installation instead of the runtime path
	ErrRuntimeOnly = errors.New("gemini: client is only available through the runtime config")
	// ErrOrganizationRequired indicates the request carried no organization
	ErrOrganizationRequired = errors.New("gemini: organization id required")
	// ErrReportFileMissing indicates the scan has no pdf file attached to parse
	ErrReportFileMissing = errors.New("gemini: scan has no pdf file attached")
	// ErrObjectManagerRequired indicates the runtime has no object storage to read the report from
	ErrObjectManagerRequired = errors.New("gemini: object storage not configured")
	// ErrNoPartsConfigured indicates no soc2 section prompts are configured so nothing can be parsed
	ErrNoPartsConfigured = errors.New("gemini: no report sections have a configured prompt")
	// ErrSectionTooThin indicates a parsed section returned far fewer items than a report should contain
	ErrSectionTooThin = errors.New("gemini: section returned too few items")
	// ErrPartMaxAttemptsReached indicates a section gave up after exhausting its parse attempts
	ErrPartMaxAttemptsReached = errors.New("gemini: max part attempts reached")
)

// These sentinels are wrapped into the reason stored on a failed scan and rendered into the
// notification the submitter reads, so their text is user-facing and carries no package prefix
var (
	// ErrRequestedPartsUnavailable indicates the scan asked only for sections with no configured prompt
	ErrRequestedPartsUnavailable = errors.New("none of the requested report sections are available")
	// ErrUnknownRequestedPart is returned when a scan asks for a section the SOC 2 kind does not know
	ErrUnknownRequestedPart = errors.New("unknown report section requested")
	// ErrSectionFailed is wrapped into a section's failure reason
	ErrSectionFailed = errors.New("could not be extracted")
	// ErrResultsIncomplete is wrapped into a section's warning when some batches failed
	ErrResultsIncomplete = errors.New("results may be incomplete")
)
