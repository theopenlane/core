package disk

import (
	"errors"
)

var (
	// ErrDiskCheckExists is returned when file existence check fails
	ErrDiskCheckExists = errors.New("failed to check if file exists")
	// ErrInvalidFolderPath is returned when an invalid folder path is provided
	ErrInvalidFolderPath = errors.New("invalid folder path provided")
	// ErrMissingLocalURL is returned when no local URL is configured for presigned links
	ErrMissingLocalURL = errors.New("missing local URL in disk storage options")
)
