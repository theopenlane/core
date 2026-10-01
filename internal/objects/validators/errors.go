package validators

import "errors"

var (
	// ErrPasswordProtectedNDA is returned when an NDA upload requires a password to open
	ErrPasswordProtectedNDA = errors.New("nda pdf is password protected, remove the password and upload again")
)
