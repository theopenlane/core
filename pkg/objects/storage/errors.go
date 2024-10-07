package storage

import "errors"

var (
	ErrProvideValidS3Bucket = errors.New("please provide a valid s3 bucket")
)
