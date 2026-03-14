package awsauditmanager

import "errors"

var (
	// ErrMetadataMissing indicates provider metadata was not supplied
	ErrMetadataMissing = errors.New("awsauditmanager: provider metadata missing")
	// ErrRoleARNMissing indicates the roleArn field is missing from metadata
	ErrRoleARNMissing = errors.New("awsauditmanager: roleArn missing")
	// ErrRegionMissing indicates the region field is missing from metadata
	ErrRegionMissing = errors.New("awsauditmanager: region missing")
)
