package awssecurityhub

import "errors"

var (
	// ErrMetadataMissing indicates provider metadata was not supplied
	ErrMetadataMissing = errors.New("awssecurityhub: provider metadata missing")
	// ErrRoleARNMissing indicates the roleArn field is missing from metadata
	ErrRoleARNMissing = errors.New("awssecurityhub: roleArn missing")
	// ErrRegionMissing indicates the region field is missing from metadata
	ErrRegionMissing = errors.New("awssecurityhub: region missing")
)
