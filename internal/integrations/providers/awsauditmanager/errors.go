package awsauditmanager

import "errors"

var (
	// ErrMetadataMissing indicates provider metadata was not supplied
	ErrMetadataMissing = errors.New("aws audit manager: provider metadata missing")
	// ErrRoleARNMissing indicates the roleArn field is missing from metadata
	ErrRoleARNMissing = errors.New("aws audit manager: roleArn missing")
	// ErrRegionMissing indicates the region field is missing from metadata
	ErrRegionMissing = errors.New("aws audit manager: region missing")
)
