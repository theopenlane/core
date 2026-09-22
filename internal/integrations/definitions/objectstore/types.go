package objectstore

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

const (
	// defaultContentType is the MIME type used when the write operation does not specify one
	defaultContentType = "application/json"
	// variantVendor is the mapping variant that links imported entities to the vendor entity type
	variantVendor = "vendor"
	// gcsScope is the OAuth scope requested for every Google Cloud Storage credential
	gcsScope = "https://www.googleapis.com/auth/devstorage.read_write"
	// jsonObjectSuffix selects the objects under a prefix that hold JSON records
	jsonObjectSuffix = ".json"
	// gcsScheme is the URI scheme for Google Cloud Storage buckets
	gcsScheme = "gs://"
	// s3Scheme is the URI scheme for Amazon S3 buckets
	s3Scheme = "s3://"
	// r2Scheme is the URI scheme for Cloudflare R2 buckets
	r2Scheme = "r2://"
)

var (
	// DefinitionID is the stable identifier for the object storage integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0OBJECTSTORE00000000001")
	// installation is the typed installation metadata handle for the definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// workloadIdentitySchema is the credential schema for GCS workload identity federation
	workloadIdentitySchema, workloadIdentityCredential = providerkit.CredentialSchema[WorkloadIdentityCredentialSchema]()
	// serviceAccountSchema is the credential schema for GCS service account key credentials
	serviceAccountSchema, serviceAccountCredential = providerkit.CredentialSchema[ServiceAccountCredentialSchema]()
	// awsAssumeRoleSchema is the credential schema for the cross-account IAM role reaching an S3 bucket
	awsAssumeRoleSchema, awsAssumeRoleCredential = providerkit.CredentialSchema[AWSAssumeRoleCredentialSchema]()
	// awsAccessKeySchema is the credential schema for static IAM keys reaching an S3 bucket
	awsAccessKeySchema, awsAccessKeyCredential = providerkit.CredentialSchema[AWSAccessKeyCredentialSchema]()
	// r2Schema is the credential schema for Cloudflare R2 keys
	r2Schema, r2Credential = providerkit.CredentialSchema[R2CredentialSchema]()
	// credentialSlots lists every credential slot in provider selection order
	credentialSlots = []types.CredentialSlotID{
		workloadIdentityCredential.ID(),
		serviceAccountCredential.ID(),
		awsAssumeRoleCredential.ID(),
		awsAccessKeyCredential.ID(),
		r2Credential.ID(),
	}
	// storageClient is the client ref for the storage client used by every operation
	storageClient = types.NewClientRef[*Client]()
	// runtimeSchema is the JSON schema and typed ref for the runtime storage config
	runtimeSchema, runtimeRef = providerkit.RuntimeSchema[RuntimeConfig]()
	// importRecordsSchema is the operation schema for the installation record import
	importRecordsSchema, importRecordsOperation = providerkit.OperationSchema[ImportRecords]()
	// systemImportSchema is the operation schema for the triggered system record import
	systemImportSchema, SystemImportOp = providerkit.OperationSchema[SystemImport]() //nolint:revive // co-initialized with schema
	// writeObjectSchema is the operation schema for the object write operation
	writeObjectSchema, writeObjectOperation = providerkit.OperationSchema[WriteObject]()
	// healthCheckSchema is the operation schema for the bucket health check
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// importSchemas lists the entityops schemas the record import operations can emit
	importSchemas = []string{entityops.SchemaEntity.Name}
)

// RuntimeConfig is the operator-provisioned configuration for the platform-owned bucket
type RuntimeConfig struct {
	// Providers holds the storage providers, exactly one enabled, holding the platform-owned bucket
	Providers storage.Providers `json:"providers" koanf:"providers" jsonschema:"description=Storage providers, exactly one enabled, holding the platform-owned bucket"`
	// Import selects the prefix and schema the system import reads unless a trigger overrides them
	Import ImportRecords `json:"import" koanf:"import" jsonschema:"description=Bucket prefix with the schema and mapping variant the system import reads unless a trigger overrides them"`
	// RunOnStartup runs the system import once when the server starts
	RunOnStartup bool `json:"runonstartup" koanf:"runonstartup" jsonschema:"description=Run the system import once when the server starts"`
}

// enabledProvider returns the single enabled provider, failing when none or more than one is enabled
func (c RuntimeConfig) enabledProvider() (storage.ProviderType, storage.ProviderConfigs, error) {
	enabled := lo.PickBy(c.Providers.ByType(), func(_ storage.ProviderType, cfg storage.ProviderConfigs) bool {
		return cfg.Enabled
	})

	switch len(enabled) {
	case 0:
		return "", storage.ProviderConfigs{}, ErrRuntimeConfigInvalid
	case 1:
		providerType := lo.Keys(enabled)[0]

		return providerType, enabled[providerType], nil
	default:
		return "", storage.ProviderConfigs{}, ErrRuntimeProviderAmbiguous
	}
}

// Provisioned reports whether the runtime config enables exactly one provider with a bucket and an import prefix
func (c RuntimeConfig) Provisioned() bool {
	_, providerCfg, err := c.enabledProvider()

	return err == nil && providerCfg.Bucket != "" && c.Import.Prefix != ""
}

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// Import includes the configuration for the record import operation
	Import ImportConfig `json:"import" jsonschema:"title=Record Import"`
}

// ImportConfig holds installation-specific configuration for the record import operation
type ImportConfig struct {
	// ImportRecords selects the prefix and schema the installation imports
	ImportRecords
	// Disable is used to disable the scheduled record import for the installation
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable the scheduled import of records from the bucket prefix"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting,example=Example: payload.status == 'ACTIVE'"`
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
	Provider string `json:"provider,omitempty" jsonschema:"title=Provider"`
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
	return types.IntegrationInstallationIdentity{ExternalID: providerScheme(storagetypes.ProviderType(m.Provider)) + m.Bucket}
}

// providerScheme returns the bucket URI scheme for a storage provider
func providerScheme(provider storagetypes.ProviderType) string {
	switch provider {
	case storage.GCSProvider:
		return gcsScheme
	case storage.S3Provider:
		return s3Scheme
	case storage.R2Provider:
		return r2Scheme
	default:
		return ""
	}
}

// ImportRecords selects the objects under a bucket prefix to map and upsert
type ImportRecords struct {
	// Prefix is the object key prefix whose JSON files hold arrays of records
	Prefix string `json:"prefix" koanf:"prefix" jsonschema:"required,title=Prefix,description=Object key prefix (directory) whose JSON files hold arrays of records"`
	// Schema is the internal model the records map to
	Schema string `json:"schema" koanf:"schema" jsonschema:"required,title=Schema,description=Internal model the records are mapped and upserted into (e.g. Entity)"`
	// Variant selects a mapping variant shipped by the definition, such as vendor to link entities to the vendor entity type
	Variant string `json:"variant,omitempty" koanf:"variant" jsonschema:"title=Variant,description=Mapping variant to apply such as vendor to link imported entities to the vendor entity type"`
}

// SystemImport is the config for one triggered system import; every field is optional and overrides the runtime config's import spec
type SystemImport struct {
	// Prefix overrides the object key prefix to read
	Prefix string `json:"prefix,omitempty" jsonschema:"title=Prefix,description=Object key prefix to read instead of the configured one"`
	// Schema overrides the internal model the records map to
	Schema string `json:"schema,omitempty" jsonschema:"title=Schema,description=Internal model to map the records into instead of the configured one"`
	// Variant overrides the mapping variant to apply
	Variant string `json:"variant,omitempty" jsonschema:"title=Variant,description=Mapping variant to apply instead of the configured one"`
}

// spec merges the trigger's overrides over the configured import spec
func (s SystemImport) spec(configured ImportRecords) ImportRecords {
	return ImportRecords{
		Prefix:  lo.CoalesceOrEmpty(s.Prefix, configured.Prefix),
		Schema:  lo.CoalesceOrEmpty(s.Schema, configured.Schema),
		Variant: lo.CoalesceOrEmpty(s.Variant, configured.Variant),
	}
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

// HealthCheck holds the result of a bucket health check
type HealthCheck struct {
	// Provider is the storage provider whose bucket was listed
	Provider string `json:"provider"`
}
