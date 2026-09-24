package objectstore

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

func TestBuilderRegisters(t *testing.T) {
	reg := registry.New()

	def, err := Builder(&RuntimeConfig{}, "", Config{})()
	assert.NilError(t, err)
	assert.Equal(t, DefinitionID.ID(), def.ID)
	assert.Equal(t, 4, len(def.Operations))
	assert.Equal(t, 2, len(def.Mappings))
	assert.Equal(t, 5, len(def.CredentialRegistrations))
	assert.Equal(t, 5, len(def.Connections))
	assert.Equal(t, 1, len(def.Clients))
	assert.Equal(t, 5, len(def.Clients[0].CredentialRefs))
	assert.Assert(t, def.RuntimeIntegration == nil)

	systemImport, found := lo.Find(def.Operations, func(op types.OperationRegistration) bool {
		return op.Name == SystemImportOp.Name()
	})
	assert.Assert(t, found, "expected the system import operation to be registered")
	assert.Assert(t, systemImport.DisabledForAll, "expected the system import to be disabled when the runtime is unprovisioned")

	assert.NilError(t, reg.Register(def))
}

func TestBuilderRuntimeOnly(t *testing.T) {
	def, err := Builder(&RuntimeConfig{}, "", Config{RuntimeOnly: true})()
	assert.NilError(t, err)
	assert.Assert(t, def.RuntimeOnly)
	assert.Assert(t, !def.Visible)

	def, err = Builder(&RuntimeConfig{}, "", Config{})()
	assert.NilError(t, err)
	assert.Assert(t, !def.RuntimeOnly)
	assert.Assert(t, def.Visible)
}

func TestSystemImportSpecOverrides(t *testing.T) {
	configured := ImportRecords{Prefix: "entities/", Schema: "Entity", Variant: "vendor"}

	assert.Equal(t, configured, SystemImport{}.spec(configured))
	assert.Equal(t, ImportRecords{Prefix: "customers/", Schema: "Entity", Variant: "vendor"}, SystemImport{Prefix: "customers/"}.spec(configured))
	assert.Equal(t, ImportRecords{Prefix: "entities/", Schema: "Entity", Variant: "other"}, SystemImport{Variant: "other"}.spec(configured))
}

func TestImportRecordsRejectsUnsupportedSchema(t *testing.T) {
	cfg := ImportRecords{Prefix: "records/", Schema: "Nope"}

	sets, err := cfg.Run(context.Background(), nil)
	assert.Assert(t, errors.Is(err, ErrSchemaUnsupported))
	assert.Assert(t, sets == nil)
}

func TestResolveImportConfig(t *testing.T) {
	configured := ImportRecords{Prefix: "vendors/", Schema: "Entity"}

	cfg, err := resolveImportConfig(json.RawMessage(`{"prefix":"records/","schema":"Entity","variant":"vendor","disable":false}`), configured)
	assert.NilError(t, err)
	assert.Equal(t, ImportRecords{Prefix: "records/", Schema: "Entity", Variant: "vendor"}, cfg)

	cfg, err = resolveImportConfig(nil, configured)
	assert.NilError(t, err)
	assert.Equal(t, configured, cfg)

	_, err = resolveImportConfig(nil, ImportRecords{})
	assert.Assert(t, errors.Is(err, ErrPrefixRequired))

	_, err = resolveImportConfig(json.RawMessage(`{"prefix":`), configured)
	assert.Assert(t, errors.Is(err, ErrOperationConfigInvalid))
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

// installationWithImport is an installation whose user input carries the import spec the client is expected to carry
var installationWithImport = &generated.Integration{
	Config: openapi.IntegrationConfig{
		ClientConfig: json.RawMessage(`{"import":{"prefix":"vendors/","schema":"Entity","variant":"vendor"}}`),
	},
}

func TestClientBuildAccessKeys(t *testing.T) {
	raw, err := json.Marshal(AWSAccessKeyCredentialSchema{
		AccessKeyID:     "k",
		SecretAccessKey: "s",
		Region:          "us-east-1",
		BucketScope:     BucketScope{Bucket: "records"},
	})
	assert.NilError(t, err)

	built, err := clientBuilder{}.Build(context.Background(), types.ClientBuildRequest{
		Credentials: types.CredentialBindings{
			{Ref: awsAccessKeyCredential.ID(), Credential: types.CredentialSet{Data: raw}},
		},
		Integration: installationWithImport,
	})
	assert.NilError(t, err)

	client, err := storageClient.Cast(built)
	assert.NilError(t, err)
	assert.Equal(t, storage.S3Provider, client.Provider.ProviderType())
	assert.Equal(t, ImportRecords{Prefix: "vendors/", Schema: "Entity", Variant: "vendor"}, client.Import)
}

func TestClientBuildR2(t *testing.T) {
	raw, err := json.Marshal(R2CredentialSchema{
		AccountID:       "acct",
		AccessKeyID:     "k",
		SecretAccessKey: "s",
		BucketScope:     BucketScope{Bucket: "records"},
	})
	assert.NilError(t, err)

	built, err := clientBuilder{}.Build(context.Background(), types.ClientBuildRequest{
		Credentials: types.CredentialBindings{
			{Ref: r2Credential.ID(), Credential: types.CredentialSet{Data: raw}},
		},
		Integration: installationWithImport,
	})
	assert.NilError(t, err)

	client, err := storageClient.Cast(built)
	assert.NilError(t, err)
	assert.Equal(t, storage.R2Provider, client.Provider.ProviderType())
	assert.Equal(t, ImportRecords{Prefix: "vendors/", Schema: "Entity", Variant: "vendor"}, client.Import)
}

func TestClientBuildRequiresCredential(t *testing.T) {
	_, err := clientBuilder{}.Build(context.Background(), types.ClientBuildRequest{})
	assert.Assert(t, errors.Is(err, ErrCredentialMetadataRequired))
}

func TestRuntimeClientBuilderDisk(t *testing.T) {
	cfg := RuntimeConfig{
		Providers: storage.Providers{Disk: storage.DiskConfig{ProviderCommon: storage.ProviderCommon{Enabled: true, Bucket: t.TempDir()}}},
		Import:    ImportRecords{Prefix: "records/", Schema: "Entity", Variant: "vendor"},
	}

	config, err := json.Marshal(cfg)
	assert.NilError(t, err)

	built, err := runtimeClientBuilder()(context.Background(), config)
	assert.NilError(t, err)

	client, err := storageClient.Cast(built)
	assert.NilError(t, err)
	assert.Equal(t, storage.DiskProvider, client.Provider.ProviderType())
	assert.Equal(t, cfg.Import, client.Import)
}

func TestRuntimeClientBuilderRejectsNoEnabledProvider(t *testing.T) {
	config, err := json.Marshal(RuntimeConfig{
		Providers: storage.Providers{Disk: storage.DiskConfig{ProviderCommon: storage.ProviderCommon{Bucket: "records"}}},
		Import:    ImportRecords{Prefix: "records/", Schema: "Entity"},
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
		Import: ImportRecords{Prefix: "records/", Schema: "Entity"},
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
	raw, err := json.Marshal(ServiceAccountCredentialSchema{
		ServiceAccountKey: `{"type":"authorized_user"}`,
		BucketScope:       BucketScope{Bucket: "records"},
	})
	assert.NilError(t, err)

	_, err = clientBuilder{}.Build(context.Background(), types.ClientBuildRequest{
		Credentials: types.CredentialBindings{
			{Ref: serviceAccountCredential.ID(), Credential: types.CredentialSet{Data: raw}},
		},
		Integration: installationWithImport,
	})
	assert.Assert(t, errors.Is(err, ErrServiceAccountKeyInvalid))
}
