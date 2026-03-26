package scim

import "errors"

var (
	// ErrResultEncode is returned when a SCIM result cannot be encoded
	ErrResultEncode = errors.New("scim: result encode failed")
	// ErrInvalidAttributes is returned when resource attributes are invalid
	ErrInvalidAttributes = errors.New("invalid resource attributes")
	// ErrDirectoryAccountNotFound is returned when a directory account cannot be found
	ErrDirectoryAccountNotFound = errors.New("directory account not found")
	// ErrDirectoryGroupNotFound is returned when a directory group cannot be found
	ErrDirectoryGroupNotFound = errors.New("directory group not found")
)
