package aws

import "errors"

var (
	// ErrMetadataMissing indicates provider metadata was not supplied.
	ErrMetadataMissing = errors.New("aws: provider metadata missing")
	// ErrRoleARNMissing indicates the roleArn field is missing from metadata.
	ErrRoleARNMissing = errors.New("aws: roleArn missing")
	// ErrRegionMissing indicates the region field is missing from metadata.
	ErrRegionMissing = errors.New("aws: region missing")
)
