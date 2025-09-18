package disk

import (
	"errors"
)

var (
	// ErrDiskPathRequired is returned when disk base path is not provided
	ErrDiskPathRequired = errors.New("missing required disk credential: base_path")
	// ErrDiskBasePathRequired is returned when disk base path is required
	ErrDiskBasePathRequired = errors.New("disk base path is required")
	// ErrDiskCreateBasePath is returned when base path creation fails
	ErrDiskCreateBasePath = errors.New("failed to create base path")
	// ErrDiskCreateDirectory is returned when directory creation fails
	ErrDiskCreateDirectory = errors.New("failed to create directory")
	// ErrDiskCreateFile is returned when file creation fails
	ErrDiskCreateFile = errors.New("failed to create file")
	// ErrDiskWriteFile is returned when file writing fails
	ErrDiskWriteFile = errors.New("failed to write file")
	// ErrDiskFileNotFound is returned when file is not found
	ErrDiskFileNotFound = errors.New("file not found")
	// ErrDiskStatFile is returned when file stat fails
	ErrDiskStatFile = errors.New("failed to stat file")
	// ErrDiskReadFile is returned when file reading fails
	ErrDiskReadFile = errors.New("failed to read file")
	// ErrDiskDeleteFile is returned when file deletion fails
	ErrDiskDeleteFile = errors.New("failed to delete file")
	// ErrDiskCheckExists is returned when file existence check fails
	ErrDiskCheckExists = errors.New("failed to check if file exists")
	// ErrDiskReadDirectory is returned when directory reading fails
	ErrDiskReadDirectory = errors.New("failed to read base directory")
	// ErrInvalidFolderPath is returned when an invalid folder path is provided
	ErrInvalidFolderPath = errors.New("invalid folder path provided")
	// ErrMissingLocalURL = errors.New("missing local URL in disk storage options"
	ErrMissingLocalURL = errors.New("missing local URL in disk storage options")
)
