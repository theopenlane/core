package awskit

import "errors"

var (
	// ErrMetadataDecode indicates AWS provider metadata could not be decoded
	ErrMetadataDecode = errors.New("awskit: metadata decode failed")
	// ErrAWSConfigBuildFailed indicates the AWS SDK config could not be constructed
	ErrAWSConfigBuildFailed = errors.New("awskit: aws config build failed")
)
