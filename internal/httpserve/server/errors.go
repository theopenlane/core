package server

import (
	"errors"
)

var (
	// ErrFailedToGetFilePath is returned when runtime.Caller fails to get the current file path
	ErrFailedToGetFilePath = errors.New("failed to get current file path")
	// ErrSCIMSpecNotFound is returned when the SCIM spec file is not found
	ErrSCIMSpecNotFound = errors.New("SCIM spec file not found")
	// ErrMissingSpecInstances is returned when analyzed handlers reference model types absent from the generated instances file
	ErrMissingSpecInstances = errors.New("model types missing from generated spec instances, run task generate:openapi to refresh them")
)
