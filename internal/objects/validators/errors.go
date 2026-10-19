package validators

import "errors"

var (
	// ErrPasswordProtectedPDF is returned when an upload requires a password to open
	ErrPasswordProtectedPDF = errors.New("pdf is password protected, remove the password and upload again")
)
