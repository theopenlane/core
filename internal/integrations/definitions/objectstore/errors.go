package objectstore

import "errors"

var (
	// ErrPrefixRequired indicates no bucket prefix was configured for the record import
	ErrPrefixRequired = errors.New("objectstore: import prefix required")
	// ErrBucketRequired indicates no bucket was configured for the installation
	ErrBucketRequired = errors.New("objectstore: bucket required")
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
	// ErrVerifyFailed indicates the bucket could not be listed with the configured credentials
	ErrVerifyFailed = errors.New("objectstore: connection verification failed")
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
	// ErrEntityTypeSeed indicates the entity type the mapping links to could not be found or created
	ErrEntityTypeSeed = errors.New("objectstore: entity type seed failed")
	// ErrObjectMoveFailed indicates a processed object could not be moved into the processed directory
	ErrObjectMoveFailed = errors.New("objectstore: object move failed")
	// ErrProviderCannotCopy indicates the storage provider does not copy objects server side
	ErrProviderCannotCopy = errors.New("objectstore: storage provider cannot copy objects")
	// ErrRecordSchemaUnknown indicates a directory under the import prefix does not name an ingestable schema
	ErrRecordSchemaUnknown = errors.New("objectstore: directory does not name an import schema")
	// ErrDatasetOwnerRequired indicates the dataset import was dispatched without an organization
	ErrDatasetOwnerRequired = errors.New("objectstore: dataset import requires an organization")
	// ErrDemoDatasetInvalid indicates the demo dataset name is not a single lowercase directory name
	ErrDemoDatasetInvalid = errors.New("objectstore: demo dataset invalid")
	// ErrDatasetInvalid indicates a dataset path is not lowercase directory segments inside the bucket
	ErrDatasetInvalid = errors.New("objectstore: dataset invalid")
	// ErrDatasetProfileUnknown indicates the import named a profile the dataset manifest does not declare
	ErrDatasetProfileUnknown = errors.New("objectstore: dataset profile unknown")
	// ErrDatasetManifestInvalid indicates the dataset manifest could not be decoded
	ErrDatasetManifestInvalid = errors.New("objectstore: dataset manifest invalid")
	// ErrDatasetSchemaUnknown indicates the dataset manifest orders a schema that does not exist
	ErrDatasetSchemaUnknown = errors.New("objectstore: dataset manifest schema unknown")
	// ErrDatasetStandardClone indicates a dataset standard could not be cloned into the organization
	ErrDatasetStandardClone = errors.New("objectstore: dataset standard clone failed")
	// ErrDemoProfileUnknown indicates the seed named a profile that does not exist
	ErrDemoProfileUnknown = errors.New("objectstore: demo profile unknown")
	// ErrDatasetOrganizationLookup indicates the importing organization's name or domain could not be read
	ErrDatasetOrganizationLookup = errors.New("objectstore: dataset organization lookup failed")
)
