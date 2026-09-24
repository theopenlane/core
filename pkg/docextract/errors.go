package docextract

import "errors"

var (
	// ErrTooShort is returned when the document has fewer pages than the profile allows
	ErrTooShort = errors.New("the uploaded file has too few pages to be the expected kind of document")
	// ErrNotMatched is returned when a required marker is missing
	ErrNotMatched = errors.New("the uploaded file does not appear to be the expected kind of document")
	// ErrMissingSections is returned when the required markers are present but too few strong markers are
	ErrMissingSections = errors.New("the uploaded file is missing the sections expected in this kind of document")
	// ErrUnreadable is returned when the pdf cannot be opened
	ErrUnreadable = errors.New("the uploaded file could not be read as a pdf")
	// ErrNoText is returned when the pdf has no readable text layer, e.g. a scanned document
	ErrNoText = errors.New("the uploaded file has no readable text, scanned or image-only pdfs are not supported")
	// ErrNoProfile is returned when the kind has no validation profile registered
	ErrNoProfile = errors.New("docextract: no validation profile for document kind")
	// ErrSystemInstructionRequired is returned when an extraction runs without a system instruction configured
	ErrSystemInstructionRequired = errors.New("docextract: system instruction prompt is required")
	// ErrMaxAttemptsReached is returned when generation keeps being interrupted past the attempt budget
	ErrMaxAttemptsReached = errors.New("docextract: max generation attempts reached")
	// ErrPromptTemplateInvalid is returned when a configured prompt template fails to parse or render
	ErrPromptTemplateInvalid = errors.New("docextract: prompt template invalid")
	// ErrCredentialsInvalid is returned when the configured service account key cannot be parsed
	ErrCredentialsInvalid = errors.New("docextract: credentials invalid")
	// ErrCacheMissing is returned when the shared document cache named in a request no longer exists
	ErrCacheMissing = errors.New("docextract: document cache missing")
)

// ValidationError carries the sentinel plus the user-facing reason built from the profile
type ValidationError struct {
	Err    error
	Reason string
}

func (e *ValidationError) Error() string { return e.Reason }

func (e *ValidationError) Unwrap() error { return e.Err }

// ValidationReason returns the user-facing reason for a validation failure, or the error text otherwise
func ValidationReason(err error) string {
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		return validationErr.Reason
	}

	return err.Error()
}
