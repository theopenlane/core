package objectstore

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

const (
	// defaultContentType is the MIME type used when the write operation does not specify one
	defaultContentType = "application/json"
	// vendorEntityTypeName is the entity type the Entity mapping links imported rows to
	vendorEntityTypeName = "vendor"
	// jsonObjectSuffix selects the objects under a prefix that hold JSON records
	jsonObjectSuffix = ".json"
	// processedDir is the directory under the import prefix that receives objects whose records imported without failure
	processedDir = "processed"
)

var (
	// DefinitionID is the stable identifier for the object storage integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0OBJECTSTORE00000000001")
	// installation is the typed installation metadata handle for the definition
	installation = types.InstallationOf[InstallationMetadata]()
	// workloadIdentity is the connection for GCS workload identity federation
	workloadIdentity = types.ConnectionOf[WorkloadIdentityCredentialSchema]()
	// serviceAccount is the connection for GCS service account key credentials
	serviceAccount = types.ConnectionOf[ServiceAccountCredentialSchema]()
	// awsAssumeRole is the connection for the cross-account IAM role reaching an S3 bucket
	awsAssumeRole = types.ConnectionOf[AWSAssumeRoleCredentialSchema]()
	// awsAccessKey is the connection for static IAM keys reaching an S3 bucket
	awsAccessKey = types.ConnectionOf[AWSAccessKeyCredentialSchema]()
	// r2 is the connection for Cloudflare R2 keys
	r2 = types.ConnectionOf[R2CredentialSchema]()
	// ImportRecordsOp is the operation ref for the record import shared by installations and the runtime
	ImportRecordsOp = types.OperationRefOf[ImportRecords]() //nolint:revive
)

// Client is the storage provider for one bucket with the import spec the runtime path reads
type Client struct {
	// Provider is the storage provider scoped to the bucket
	Provider storage.Provider
	// Import is the prefix and kind the runtime import reads when the operation config leaves them empty, zero for installations
	Import ImportSpec
}

// RuntimeConfig is the operator-provisioned configuration for the platform-owned bucket
type RuntimeConfig struct {
	// Providers holds the storage providers, exactly one enabled, holding the platform-owned bucket
	Providers storage.Providers `json:"providers" koanf:"providers" jsonschema:"description=Storage providers, exactly one enabled, holding the platform-owned bucket"`
	// Import selects the prefix and record kind the runtime import reads when the operation config leaves them empty
	Import ImportSpec `json:"import" koanf:"import" jsonschema:"description=Bucket prefix and record kind the runtime import reads when the operation config leaves them empty"`
}

// enabledProvider returns the single enabled provider, failing when none or more than one is enabled
func (c RuntimeConfig) enabledProvider() (storage.ProviderType, storage.ProviderCommon, error) {
	enabled := lo.PickBy(c.Providers.ByType(), func(_ storage.ProviderType, cfg storage.ProviderCommon) bool {
		return cfg.Enabled
	})

	switch len(enabled) {
	case 0:
		return "", storage.ProviderCommon{}, ErrRuntimeConfigInvalid
	case 1:
		providerType := lo.Keys(enabled)[0]

		return providerType, enabled[providerType], nil
	default:
		return "", storage.ProviderCommon{}, ErrRuntimeProviderAmbiguous
	}
}

// Provisioned reports whether the runtime config enables exactly one provider with a bucket, an import prefix and a record kind
func (c RuntimeConfig) Provisioned() bool {
	_, providerCfg, err := c.enabledProvider()

	return err == nil && providerCfg.Bucket != "" && c.Import.Prefix != ""
}

// BucketScope names the bucket an installation reads and writes
type BucketScope struct {
	// Bucket is the bucket the installation reads from and writes to
	Bucket string `json:"bucket" jsonschema:"required,title=Bucket,description=Bucket the integration reads from and writes to"`
}

// WorkloadIdentityCredentialSchema holds the workload identity federation inputs for one installation;
// Openlane signs a short-lived assertion with its own keys and exchanges it at Google STS so nothing here is secret
type WorkloadIdentityCredentialSchema struct {
	// ProjectNumber is the numeric project number of the GCP project hosting the openlane workload identity pool and provider
	ProjectNumber string `json:"projectNumber" jsonschema:"required,title=GCP Project Number,pattern=^[0-9]+$,description=Numeric project number of the GCP project hosting the openlane workload identity pool and provider"`
	// ServiceAccountEmail optionally impersonates a service account with access to the bucket
	ServiceAccountEmail string `json:"serviceAccountEmail,omitempty" jsonschema:"title=Service Account Email,description=Optional service account to impersonate when accessing the bucket"`
	// ProjectID is the Google Cloud project that owns the bucket
	ProjectID string `json:"projectId,omitempty" jsonschema:"title=Project ID,description=Google Cloud project that owns the bucket and is used for quota attribution"`
	// BucketScope names the bucket the installation reads and writes
	BucketScope
}

// ServiceAccountCredentialSchema holds the service account key credentials for one installation
type ServiceAccountCredentialSchema struct {
	// ServiceAccountKey is the service account key JSON used for direct credentials
	ServiceAccountKey string `json:"serviceAccountKey" jsonschema:"required,title=Service Account Key JSON,secret=true,description=Service Account JSON used to authenticate on your behalf to Google Cloud Storage"`
	// ProjectID is the Google Cloud project that owns the bucket
	ProjectID string `json:"projectId,omitempty" jsonschema:"title=Project ID,description=Google Cloud project that owns the bucket and is used for quota attribution"`
	// BucketScope names the bucket the installation reads and writes
	BucketScope
}

// AWSAssumeRoleCredentialSchema holds the cross-account role Openlane assumes to reach an S3 bucket
type AWSAssumeRoleCredentialSchema struct {
	// RoleARN is the cross-account IAM role Openlane assumes in the customer account
	RoleARN string `json:"roleArn" jsonschema:"required,title=IAM Role ARN,description=Cross-account role Openlane assumes in your account,secret=true"`
	// ExternalID is the external ID required in the role trust policy
	ExternalID string `json:"externalId" jsonschema:"required,title=External ID,description=External ID required in the role trust policy" jsonschema_extras:"generate=true"`
	// Region is the AWS region of the bucket
	Region string `json:"region" jsonschema:"required,title=Region,description=AWS region of the bucket (e.g. us-east-1)"`
	// SessionName is an optional STS session name override
	SessionName string `json:"sessionName,omitempty" jsonschema:"title=Session Name,description=Optional STS session name override"`
	// BucketScope names the bucket the installation reads and writes
	BucketScope
}

// AWSAccessKeyCredentialSchema holds static IAM keys for an S3 bucket
type AWSAccessKeyCredentialSchema struct {
	// AccessKeyID is the IAM access key ID
	AccessKeyID string `json:"accessKeyId" jsonschema:"required,title=Access Key ID"`
	// SecretAccessKey is the IAM secret access key
	SecretAccessKey string `json:"secretAccessKey" jsonschema:"required,title=Secret Access Key,secret=true"`
	// Region is the AWS region of the bucket
	Region string `json:"region" jsonschema:"required,title=Region,description=AWS region of the bucket (e.g. us-east-1)"`
	// BucketScope names the bucket the installation reads and writes
	BucketScope
}

// R2CredentialSchema holds Cloudflare R2 S3-compatible keys for a bucket
type R2CredentialSchema struct {
	// AccountID is the Cloudflare account that owns the bucket
	AccountID string `json:"accountId" jsonschema:"required,title=Account ID,description=Cloudflare account ID that owns the bucket"`
	// AccessKeyID is the R2 API token access key ID
	AccessKeyID string `json:"accessKeyId" jsonschema:"required,title=Access Key ID"`
	// SecretAccessKey is the R2 API token secret access key
	SecretAccessKey string `json:"secretAccessKey" jsonschema:"required,title=Secret Access Key,secret=true"`
	// BucketScope names the bucket the installation reads and writes
	BucketScope
}

// InstallationMetadata holds the stable provider, bucket and identity attributes for one installation
type InstallationMetadata struct {
	// Provider is the storage provider holding the bucket
	Provider storage.ProviderType `json:"provider,omitempty" jsonschema:"title=Provider"`
	// Bucket is the bucket the installation is scoped to
	Bucket string `json:"bucket,omitempty" jsonschema:"title=Bucket"`
	// Region is the AWS region of the bucket when the provider is s3
	Region string `json:"region,omitempty" jsonschema:"title=Region"`
	// ProjectID is the Google Cloud project that owns the bucket when the provider is gcs
	ProjectID string `json:"projectId,omitempty" jsonschema:"title=Project ID"`
	// ServiceAccountEmail is the service account the installation acts as when available
	ServiceAccountEmail string `json:"serviceAccountEmail,omitempty" jsonschema:"title=Service Account Email"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{ExternalID: storage.Scheme(m.Provider) + m.Bucket}
}

// ImportSpec selects the objects under a bucket prefix
type ImportSpec struct {
	// Prefix is the object key prefix whose JSON files hold arrays of records
	Prefix string `json:"prefix" koanf:"prefix" jsonschema:"required,title=Prefix,description=Object key prefix (directory) whose JSON files hold arrays of records"`
	// MoveProcessed moves each object under the processed directory of the prefix once its records import without failure
	MoveProcessed bool `json:"moveProcessed,omitempty" koanf:"moveprocessed" jsonschema:"title=Move Processed,description=Move each object into the processed directory under the prefix once its records import without failure"`
}

// ImportRecordsResult reports the record import across every object under the prefix
type ImportRecordsResult struct {
	operations.IngestResult
	// Objects is the number of JSON objects read under the prefix
	Objects int
	// Moved is the number of objects moved into the processed directory
	Moved int
}

// ImportRecords is the stored input of the record import
type ImportRecords struct {
	types.OperationSettings
	// ImportSpec selects the prefix to import
	ImportSpec
}

// WriteObject writes a document to the bucket under a key
type WriteObject struct {
	// Key is the object key to write the document under
	Key string `json:"key" jsonschema:"required,title=Key,description=Object key to write the document under"`
	// ContentType is the MIME type recorded for the object
	ContentType string `json:"contentType,omitempty" jsonschema:"title=Content Type,description=MIME type of the object and defaults to application/json"`
	// Content is the document to write; a string is written verbatim and any other value is JSON encoded
	Content any `json:"content" jsonschema:"required,title=Content,description=Document to write; strings are written verbatim and other values are JSON encoded"`
}

// WriteObjectResult describes the object written by the write operation
type WriteObjectResult struct {
	// Bucket is the bucket the object was written to
	Bucket string `json:"bucket"`
	// Key is the object key the document was written under
	Key string `json:"key"`
	// ContentType is the MIME type recorded for the object
	ContentType string `json:"contentType"`
	// Size is the number of bytes written
	Size int64 `json:"size"`
	// URI is the provider URI of the written object
	URI string `json:"uri"`
}
