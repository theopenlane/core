package storage

import (
	"errors"
	"fmt"
)

var (
	// ErrParseFailed is matched by every ParseError, so callers can check for any document parse failure with errors.Is
	ErrParseFailed = errors.New("failed to parse document")
	// ErrProviderResolutionRequired is returned when provider resolution is required but not available
	ErrProviderResolutionRequired = errors.New("provider resolution required - use external orchestration layer")
)

// Document formats reported by ParseError
const (
	ParseFormatJSON = "JSON"
	ParseFormatYAML = "YAML"
	ParseFormatDOCX = "DOCX"
	ParseFormatPDF  = "PDF"
)

// ParseError is returned when a document cannot be parsed as its declared format
type ParseError struct {
	// Format is the document format that failed to parse, one of the ParseFormat constants
	Format string
	// Err is the underlying parser error
	Err error
}

// Error implements the error interface
func (e *ParseError) Error() string {
	return fmt.Sprintf("failed to parse %s: %v", e.Format, e.Err)
}

// Unwrap exposes ErrParseFailed and the underlying parser error to errors.Is and errors.As
func (e *ParseError) Unwrap() []error {
	return []error{ErrParseFailed, e.Err}
}

// newParseError wraps a parser error with the format that failed
func newParseError(format string, err error) error {
	return &ParseError{Format: format, Err: err}
}

type errorMsg string

func (e errorMsg) Error() string { return string(e) }

const (
	// ErrNoFilesUploaded is returned when no files were uploaded in the request
	ErrNoFilesUploaded = errorMsg("objects: no uploadable files found in request")
)
