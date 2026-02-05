package awssecurityhub

import "errors"

var (
	ErrMetadataMissing = errors.New("aws security hub: provider metadata missing")
	ErrRoleARNMissing  = errors.New("aws security hub: roleArn missing")
	ErrRegionMissing   = errors.New("aws security hub: region missing")
)
