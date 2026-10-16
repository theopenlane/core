package objectstore

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

func TestBuilderRegisters(t *testing.T) {
	reg := registry.New()

	def, err := Builder(&RuntimeConfig{}, "", Config{})()
	assert.NilError(t, err)
	assert.Equal(t, DefinitionID.ID(), def.ID)
	assert.Equal(t, 3, len(def.Operations))
	assert.Equal(t, 2, len(def.Mappings))
	assert.Equal(t, 5, len(def.ConnectionList()))
	assert.Assert(t, def.RuntimeIntegration == nil)

	for _, connection := range def.ConnectionList() {
		assert.Equal(t, 1, len(connection.Clients), "connection %s must provide the storage client", connection.Credential.Name)
		assert.Assert(t, connection.Verify.Handle != nil, "connection %s must verify", connection.Credential.Name)
	}

	systemImport, found := lo.Find(def.Operations, func(op types.OperationRegistration) bool {
		return op.Name == SystemImportOp.Name()
	})
	assert.Assert(t, found, "expected the system import operation to be registered")
	assert.Assert(t, systemImport.DisabledForAll, "expected the system import to be disabled when the runtime is unprovisioned")
	assert.Assert(t, systemImport.Internal)
	assert.Assert(t, !systemImport.Stored, "the system import takes its overrides per trigger")

	importRecords, found := lo.Find(def.Operations, func(op types.OperationRegistration) bool {
		return op.Name == "ImportRecords"
	})
	assert.Assert(t, found, "expected the record import operation to be registered")
	assert.Assert(t, importRecords.Stored, "the record import keeps its prefix per installation")
	assert.Assert(t, importRecords.Policy.Reconcile)

	assert.NilError(t, reg.Register(def))
}

func TestBuilderHiddenButActive(t *testing.T) {
	def, err := Builder(&RuntimeConfig{}, "", Config{})()
	assert.NilError(t, err)
	assert.Assert(t, def.Active, "the runtime must be able to execute the definition")
	assert.Assert(t, !def.Visible, "organizations must not see or install the definition")
}

func TestSystemImportSpecOverrides(t *testing.T) {
	configured := ImportSpec{Prefix: "entities/", Schema: "Entity", Variant: "vendor"}

	assert.Equal(t, configured, SystemImport{}.spec(configured))
	assert.Equal(t, ImportSpec{Prefix: "customers/", Schema: "Entity", Variant: "vendor"}, SystemImport{Prefix: "customers/"}.spec(configured))
	assert.Equal(t, ImportSpec{Prefix: "entities/", Schema: "Entity", Variant: "other"}, SystemImport{Variant: "other"}.spec(configured))
}

func TestImportSpecRejectsUnsupportedSchema(t *testing.T) {
	cfg := ImportSpec{Prefix: "records/", Schema: "Nope"}

	sets, err := cfg.Run(context.Background(), nil)
	assert.Assert(t, errors.Is(err, ErrSchemaUnsupported))
	assert.Assert(t, sets == nil)
}

func TestImportSpecRejectsEmptyPrefix(t *testing.T) {
	sets, err := ImportSpec{Schema: "Entity"}.Run(context.Background(), nil)
	assert.Assert(t, errors.Is(err, ErrPrefixRequired))
	assert.Assert(t, sets == nil)
}

func TestValidateImportRecords(t *testing.T) {
	assert.NilError(t, validateImportRecords(context.Background(), types.InstallationRequest{}, &ImportRecords{ImportSpec: ImportSpec{Prefix: "records/", Schema: "Entity"}}))

	err := validateImportRecords(context.Background(), types.InstallationRequest{}, &ImportRecords{ImportSpec: ImportSpec{Prefix: "records/", Schema: "Nope"}})
	assert.Assert(t, errors.Is(err, ErrSchemaUnsupported))
}

func TestDecodeRecords(t *testing.T) {
	records, err := decodeRecords([]byte(` [{"a":1},{"b":2}] `))
	assert.NilError(t, err)
	assert.Equal(t, 2, len(records))
	assert.Equal(t, `{"a":1}`, string(records[0]))

	records, err = decodeRecords([]byte(`{"a":1}`))
	assert.NilError(t, err)
	assert.Equal(t, 1, len(records))
	assert.Equal(t, `{"a":1}`, string(records[0]))

	_, err = decodeRecords([]byte(`"not records"`))
	assert.Assert(t, err != nil)

	_, err = decodeRecords([]byte(`{"a":`))
	assert.Assert(t, err != nil)
}

func TestWriteObjectEncode(t *testing.T) {
	body, err := WriteObject{Content: `{"raw": true}`}.encode()
	assert.NilError(t, err)
	assert.Equal(t, `{"raw": true}`, string(body))

	body, err = WriteObject{Content: map[string]any{"count": 2}}.encode()
	assert.NilError(t, err)
	assert.Equal(t, `{"count":2}`, string(body))

	_, err = WriteObject{Content: make(chan int)}.encode()
	assert.Assert(t, errors.Is(err, ErrContentEncode))
}

func TestClientBuildAccessKeys(t *testing.T) {
	client, err := accessKeyClient(context.Background(), types.ConnectionRequest[AWSAccessKeyCredentialSchema]{
		Credential: AWSAccessKeyCredentialSchema{
			AccessKeyID:     "k",
			SecretAccessKey: "s",
			Region:          "us-east-1",
			BucketScope:     BucketScope{Bucket: "records"},
		},
	})
	assert.NilError(t, err)
	assert.Equal(t, storage.S3Provider, client.Provider.ProviderType())
	assert.Equal(t, ImportSpec{}, client.Import, "an installation client carries no runtime import spec")
}

func TestClientBuildR2(t *testing.T) {
	client, err := r2Client(context.Background(), types.ConnectionRequest[R2CredentialSchema]{
		Credential: R2CredentialSchema{
			AccountID:       "acct",
			AccessKeyID:     "k",
			SecretAccessKey: "s",
			BucketScope:     BucketScope{Bucket: "records"},
		},
	})
	assert.NilError(t, err)
	assert.Equal(t, storage.R2Provider, client.Provider.ProviderType())
}

func TestRuntimeClientBuilderDisk(t *testing.T) {
	cfg := RuntimeConfig{
		Providers: storage.Providers{Disk: storage.DiskConfig{ProviderCommon: storage.ProviderCommon{Enabled: true, Bucket: t.TempDir()}}},
		Import:    ImportSpec{Prefix: "records/", Schema: "Entity", Variant: "vendor"},
	}

	config, err := json.Marshal(cfg)
	assert.NilError(t, err)

	built, err := runtimeClientBuilder()(context.Background(), config)
	assert.NilError(t, err)

	client, ok := built.(*Client)
	assert.Assert(t, ok, "the runtime client must be the storage client every operation binds")
	assert.Equal(t, storage.DiskProvider, client.Provider.ProviderType())
	assert.Equal(t, cfg.Import, client.Import)
}

func TestRuntimeClientBuilderRejectsNoEnabledProvider(t *testing.T) {
	config, err := json.Marshal(RuntimeConfig{
		Providers: storage.Providers{Disk: storage.DiskConfig{ProviderCommon: storage.ProviderCommon{Bucket: "records"}}},
		Import:    ImportSpec{Prefix: "records/", Schema: "Entity"},
	})
	assert.NilError(t, err)

	_, err = runtimeClientBuilder()(context.Background(), config)
	assert.Assert(t, errors.Is(err, ErrRuntimeConfigInvalid))
}

func TestRuntimeClientBuilderRejectsAmbiguousProviders(t *testing.T) {
	config, err := json.Marshal(RuntimeConfig{
		Providers: storage.Providers{
			S3:   storage.S3Config{ProviderCommon: storage.ProviderCommon{Enabled: true, Bucket: "records"}},
			Disk: storage.DiskConfig{ProviderCommon: storage.ProviderCommon{Enabled: true, Bucket: "records"}},
		},
		Import: ImportSpec{Prefix: "records/", Schema: "Entity"},
	})
	assert.NilError(t, err)

	_, err = runtimeClientBuilder()(context.Background(), config)
	assert.Assert(t, errors.Is(err, ErrRuntimeProviderAmbiguous))
}

func TestInstallationIdentityCarriesProviderScheme(t *testing.T) {
	assert.Equal(t, "s3://records", InstallationMetadata{Provider: storage.S3Provider, Bucket: "records"}.InstallationIdentity().ExternalID)
	assert.Equal(t, "gs://records", InstallationMetadata{Provider: storage.GCSProvider, Bucket: "records"}.InstallationIdentity().ExternalID)
	assert.Equal(t, "r2://records", InstallationMetadata{Provider: storage.R2Provider, Bucket: "records"}.InstallationIdentity().ExternalID)
}

func TestLoadAWSConfig(t *testing.T) {
	cfg, err := loadAWSConfig(context.Background(), "us-east-1", storage.AccessKeyCredentials{AccessKeyID: "k", SecretAccessKey: "s"})
	assert.NilError(t, err)
	assert.Equal(t, "us-east-1", cfg.Region)

	creds, err := cfg.Credentials.Retrieve(context.Background())
	assert.NilError(t, err)
	assert.Equal(t, "k", creds.AccessKeyID)
	assert.Equal(t, "s", creds.SecretAccessKey)

	cfg, err = loadAWSConfig(context.Background(), "eu-west-1", storage.AccessKeyCredentials{})
	assert.NilError(t, err)
	assert.Equal(t, "eu-west-1", cfg.Region)
}

func TestClientBuildServiceAccountRejectsInvalidKey(t *testing.T) {
	_, err := serviceAccountClient(context.Background(), types.ConnectionRequest[ServiceAccountCredentialSchema]{
		Credential: ServiceAccountCredentialSchema{
			ServiceAccountKey: `{"type":"authorized_user"}`,
			BucketScope:       BucketScope{Bucket: "records"},
		},
	})
	assert.Assert(t, errors.Is(err, ErrServiceAccountKeyInvalid))
}

func TestServiceAccountEmail(t *testing.T) {
	assert.Equal(t, "svc@example.com", serviceAccountEmail(`{"client_email":"svc@example.com"}`))
	assert.Equal(t, "svc@example.com", serviceAccountEmail(`"{\"client_email\":\"svc@example.com\"}"`))
	assert.Equal(t, "", serviceAccountEmail(""))
}
