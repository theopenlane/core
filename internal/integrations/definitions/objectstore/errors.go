package objectstore

import "errors"

var (
	// ErrOperationConfigInvalid indicates operation config could not be decoded
	ErrOperationConfigInvalid = errors.New("objectstore: operation config invalid")
	// ErrSchemaUnsupported indicates the requested import schema is not one this definition can emit
	ErrSchemaUnsupported = errors.New("objectstore: import schema unsupported")
	// ErrPrefixRequired indicates no bucket prefix was configured for the record import
	ErrPrefixRequired = errors.New("objectstore: import prefix required")
	// ErrBucketRequired indicates no bucket was configured for the installation
	ErrBucketRequired = errors.New("objectstore: bucket required")
	// ErrCredentialMetadataRequired indicates no credential metadata was provided
	ErrCredentialMetadataRequired = errors.New("objectstore: credential metadata required")
	// ErrMetadataDecode indicates credential metadata could not be decoded
	ErrMetadataDecode = errors.New("objectstore: failed to decode credential metadata")
	// ErrProjectNumberRequired indicates no workload identity pool project number was provided
	ErrProjectNumberRequired = errors.New("objectstore: workload identity project number required")
	// ErrServiceAccountKeyInvalid indicates the service account key JSON is invalid
	ErrServiceAccountKeyInvalid = errors.New("objectstore: service account key invalid")
	// ErrClientCreate indicates the storage client could not be created
	ErrClientCreate = errors.New("objectstore: storage client creation failed")
	// ErrRuntimeConfigDecode indicates the runtime config could not be decoded
	ErrRuntimeConfigDecode = errors.New("objectstore: runtime config decode failed")
	// ErrRuntimeConfigInvalid indicates the runtime config has no enabled provider, bucket or import prefix
	ErrRuntimeConfigInvalid = errors.New("objectstore: runtime config invalid")
	// ErrRuntimeProviderAmbiguous indicates the runtime config enables more than one provider
	ErrRuntimeProviderAmbiguous = errors.New("objectstore: runtime config enables more than one provider")
	// ErrHealthCheckFailed indicates the bucket could not be listed with the configured credentials
	ErrHealthCheckFailed = errors.New("objectstore: health check failed")
	// ErrListObjectsFailed indicates the objects under the import prefix could not be listed
	ErrListObjectsFailed = errors.New("objectstore: list objects failed")
	// ErrObjectDownloadFailed indicates an object's bytes could not be read from the bucket
	ErrObjectDownloadFailed = errors.New("objectstore: object download failed")
	// ErrObjectContentInvalid indicates an object is neither a JSON array of records nor a single JSON record
	ErrObjectContentInvalid = errors.New("objectstore: object content is not json records")
	// ErrContentEncode indicates the document to write could not be JSON encoded
	ErrContentEncode = errors.New("objectstore: content encode failed")
	// ErrObjectUploadFailed indicates the document could not be written to the bucket
	ErrObjectUploadFailed = errors.New("objectstore: object upload failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("objectstore: result encode failed")
)
