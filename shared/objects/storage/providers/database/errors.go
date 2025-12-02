package database

import "errors"

var (
	// ErrMissingEntClient indicates that no ent client was available in context when required.
	ErrMissingEntClient = errors.New("database storage requires ent client in context")
	// ErrMissingFileIdentifier indicates neither file ID nor key was supplied for an operation.
	ErrMissingFileIdentifier = errors.New("file identifier required for database storage operation")
	// ErrTokenManagerRequired indicates presigned URL generation attempted without a token manager.
	ErrTokenManagerRequired = errors.New("token manager required for database presigned urls")
	// ErrFileNotFound is returned when the requested file does not exist or has no stored bytes.
	ErrFileNotFound = errors.New("file not found in database storage")
	// ErrDatabaseProviderRequiresProxyPresign indicates that the database storage provider requires proxy presigning to be enabled.
	ErrDatabaseProviderRequiresProxyPresign = errors.New("database storage provider requires proxy presign to be enabled")
)
