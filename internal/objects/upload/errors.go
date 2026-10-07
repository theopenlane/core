package upload

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidProvenance is the parent of every error returned for an unacceptable provenance record
	ErrInvalidProvenance = errors.New("invalid file provenance")
	// ErrProvenanceNotSupported is returned when provenance is attached to an upload that is not evidence
	ErrProvenanceNotSupported = fmt.Errorf("%w: provenance can only be attached to evidence files", ErrInvalidProvenance)
	// ErrProvenanceTooLarge is returned when the provenance claims exceed the maximum encoded size
	ErrProvenanceTooLarge = fmt.Errorf("%w: claims are too large", ErrInvalidProvenance)
	// ErrProvenanceWithoutContent is returned when there are no received bytes to verify the claims against
	ErrProvenanceWithoutContent = fmt.Errorf("%w: the upload has no content to verify", ErrInvalidProvenance)
	// ErrInvalidProvenanceHash is returned when the claimed artifact_sha256 is not a hex-encoded SHA-256
	ErrInvalidProvenanceHash = fmt.Errorf("%w: artifact_sha256 must be a 64 character hex-encoded SHA-256", ErrInvalidProvenance)
	// ErrProvenanceHashMismatch is returned when the claimed artifact_sha256 differs from the received bytes
	ErrProvenanceHashMismatch = fmt.Errorf("%w: artifact_sha256 does not match the uploaded file", ErrInvalidProvenance)
)
