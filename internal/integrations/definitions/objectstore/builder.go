package objectstore

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// Builder returns the object storage definition builder advertising the supplied federation issuer and using
// cfg as the operator config carrying the AWS source identity and the runtime-only switch; when runtime.Provisioned() is true a
// RuntimeIntegration is included so a triggered system import can read the platform-owned bucket
func Builder(runtime *RuntimeConfig, federationIssuer string, cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		ingest := lo.Map(importSchemas, func(schema string, _ int) types.IngestContract {
			return types.IngestContract{Schema: schema}
		})

		def := types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Family:      "storage",
				DisplayName: "Object Storage",
				Description: "Read and write objects in your cloud storage bucket and import its JSON records.",
				Category:    "data",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/object-storage",
				Tags:        []string{"files", "import"},
				Active:      true,
				Visible:     !cfg.RuntimeOnly,
				RuntimeOnly: cfg.RuntimeOnly,
			},
			UserInput: &types.UserInputRegistration{
				Schema: jsonx.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         workloadIdentityCredential.ID(),
					Name:        "GCP Workload Identity Federation",
					Description: "Federated access to a Google Cloud Storage bucket with no stored keys.",
					Schema:      workloadIdentitySchema,
					Recommended: true,
				},
				{
					Ref:         serviceAccountCredential.ID(),
					Name:        "GCP Service Account Key",
					Description: "GCP service account key used to access a Google Cloud Storage bucket.",
					Schema:      serviceAccountSchema,
				},
				{
					Ref:         awsAssumeRoleCredential.ID(),
					Name:        "AWS IAM Role",
					Description: "Cross-account IAM role Openlane assumes to access an S3 bucket.",
					Schema:      awsAssumeRoleSchema,
				},
				{
					Ref:         awsAccessKeyCredential.ID(),
					Name:        "AWS Access Keys",
					Description: "Static IAM access keys used to access an S3 bucket.",
					Schema:      awsAccessKeySchema,
				},
				{
					Ref:         r2Credential.ID(),
					Name:        "Cloudflare R2",
					Description: "R2 API token keys used to access a Cloudflare R2 bucket.",
					Schema:      r2Schema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef: workloadIdentityCredential.ID(),
					Name:          "GCP Workload Identity Federation",
					Description:   "Configure Google Cloud Storage bucket access by trusting Openlane as an OIDC identity provider, so no service account key is ever stored.",
					Meta: map[string]types.MetaInfo{
						"Openlane Issuer URI": {
							Value:     federationIssuer,
							AllowCopy: true,
						},
					},
					CredentialRefs: []types.CredentialSlotID{workloadIdentityCredential.ID()},
					ClientRefs:     []types.ClientID{storageClient.ID()},
					HealthCheck: &types.HealthCheckRegistration{
						ClientRef: storageClient.ID(),
						Handle:    HealthCheck{}.Handle(),
					},
					Integration: installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: workloadIdentityCredential.ID(),
						Description:   "Removes the stored workload identity configuration from Openlane. If the workload identity pool is no longer needed, delete the pool and its provider from your Google Cloud project.",
					},
				},
				{
					CredentialRef:  serviceAccountCredential.ID(),
					Name:           "GCP Service Account",
					Description:    "Configure Google Cloud Storage bucket access using a GCP service account key.",
					CredentialRefs: []types.CredentialSlotID{serviceAccountCredential.ID()},
					ClientRefs:     []types.ClientID{storageClient.ID()},
					HealthCheck: &types.HealthCheckRegistration{
						ClientRef: storageClient.ID(),
						Handle:    HealthCheck{}.Handle(),
					},
					Integration: installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: serviceAccountCredential.ID(),
						Description:   "Removes the stored service account key from Openlane. If the GCP service account is no longer needed, delete it from your Google Cloud project.",
					},
				},
				{
					CredentialRef: awsAssumeRoleCredential.ID(),
					Name:          "AWS IAM Role",
					Description:   "Configure S3 bucket access using a cross-account IAM role that trusts Openlane with an external ID.",
					Meta: map[string]types.MetaInfo{
						"Openlane Principal ARN": {
							Value:     cfg.ARN,
							AllowCopy: true,
						},
					},
					CredentialRefs: []types.CredentialSlotID{awsAssumeRoleCredential.ID()},
					ClientRefs:     []types.ClientID{storageClient.ID()},
					HealthCheck: &types.HealthCheckRegistration{
						ClientRef: storageClient.ID(),
						Handle:    HealthCheck{}.Handle(),
					},
					Integration: installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: awsAssumeRoleCredential.ID(),
						Description:   "Removes the stored IAM role configuration from Openlane. If the cross-account IAM role is no longer needed, delete it from your AWS account.",
					},
				},
				{
					CredentialRef:  awsAccessKeyCredential.ID(),
					Name:           "AWS Access Keys",
					Description:    "Configure S3 bucket access using static IAM access keys.",
					CredentialRefs: []types.CredentialSlotID{awsAccessKeyCredential.ID()},
					ClientRefs:     []types.ClientID{storageClient.ID()},
					HealthCheck: &types.HealthCheckRegistration{
						ClientRef: storageClient.ID(),
						Handle:    HealthCheck{}.Handle(),
					},
					Integration: installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: awsAccessKeyCredential.ID(),
						Description:   "Removes the stored IAM access keys from Openlane. If the IAM user is no longer needed, delete it from your AWS account.",
					},
				},
				{
					CredentialRef:  r2Credential.ID(),
					Name:           "Cloudflare R2",
					Description:    "Configure Cloudflare R2 bucket access using an R2 API token's S3-compatible keys.",
					CredentialRefs: []types.CredentialSlotID{r2Credential.ID()},
					ClientRefs:     []types.ClientID{storageClient.ID()},
					HealthCheck: &types.HealthCheckRegistration{
						ClientRef: storageClient.ID(),
						Handle:    HealthCheck{}.Handle(),
					},
					Integration: installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: r2Credential.ID(),
						Description:   "Removes the stored R2 keys from Openlane. If the R2 API token is no longer needed, revoke it in your Cloudflare account.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            storageClient.ID(),
					CredentialRefs: credentialSlots,
					Description:    "Object storage client scoped to one bucket",
					Build:          clientBuilder{aws: cfg}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthCheckOperation.Name(),
					Description:  "List the bucket to verify the configured credentials can reach it",
					Topic:        DefinitionID.OperationTopic(healthCheckOperation.Name()),
					ClientRef:    storageClient.ID(),
					ConfigSchema: healthCheckSchema,
					Policy:       types.ExecutionPolicy{Inline: true},
					Handle:       HealthCheck{}.Handle(),
				},
				{
					Name:                importRecordsOperation.Name(),
					Description:         "Import the JSON records under a bucket prefix through the definition's mappings",
					Topic:               DefinitionID.OperationTopic(importRecordsOperation.Name()),
					ClientRef:           storageClient.ID(),
					ConfigSchema:        importRecordsSchema,
					Policy:              types.ExecutionPolicy{Reconcile: true},
					Ingest:              ingest,
					IngestHandle:        ImportRecords{}.IngestHandle(),
					SkipDefaultLookback: true,
					Disabled:            providerkit.DisabledWhen(func(u UserInput) bool { return u.Import.Disable }),
					ConfigResolver:      providerkit.ConfigFrom(func(u UserInput) ImportConfig { return u.Import }),
				},
				{
					Name:         writeObjectOperation.Name(),
					Description:  "Write a document into the bucket under a key",
					Topic:        DefinitionID.OperationTopic(writeObjectOperation.Name()),
					ClientRef:    storageClient.ID(),
					ConfigSchema: writeObjectSchema,
					Policy:       types.ExecutionPolicy{Inline: true},
					Handle:       WriteObject{}.Handle(),
				},
				{
					Name:                SystemImportOp.Name(),
					Description:         "Import platform-owned JSON records from the runtime bucket prefix as system-owned rows when triggered",
					Topic:               DefinitionID.OperationTopic(SystemImportOp.Name()),
					ClientRef:           storageClient.ID(),
					ConfigSchema:        systemImportSchema,
					Policy:              types.ExecutionPolicy{Inline: true},
					Internal:            true,
					CustomerSelectable:  lo.ToPtr(false),
					DisabledForAll:      !runtime.Provisioned(),
					Ingest:              ingest,
					IngestHandle:        SystemImport{}.IngestHandle(),
					SkipDefaultLookback: true,
				},
			},
			Mappings: []types.MappingRegistration{
				{
					Schema: entityops.SchemaEntity.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprEntity,
					},
				},
				{
					Schema:  entityops.SchemaEntity.Name,
					Variant: variantVendor,
					Spec: types.MappingOverride{
						FilterExpr: filterExprVendor,
						MapExpr:    mapExprEntity,
						Links: []types.LinkRule{
							{
								TargetSchema: entityops.SchemaEntityType.Name,
								Expression:   linkExprVendorEntityType,
							},
						},
					},
				},
			},
		}

		if runtime.Provisioned() {
			runtimeRef.SetConfig(runtime)

			marshaledConfig, err := runtimeRef.MarshalConfig()
			if err != nil {
				return types.Definition{}, fmt.Errorf("%w: %w", ErrRuntimeConfigDecode, err)
			}

			def.RuntimeIntegration = &types.RuntimeIntegrationRegistration{
				Ref:    runtimeRef.ID(),
				Schema: runtimeSchema,
				Config: marshaledConfig,
				Build:  runtimeClientBuilder(),
			}
		}

		return def, nil
	})
}
