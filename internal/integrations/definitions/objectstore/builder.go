package objectstore

import (
	"fmt"

	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// Builder returns the object storage definition builder advertising the supplied federation issuer and using
// cfg as the operator config carrying the AWS source identity; the definition registers hidden so organizations
// cannot install it, and when runtime.Provisioned() is true a RuntimeIntegration is included so the scheduled
// record import cycle reads the platform-owned bucket
func Builder(runtime *RuntimeConfig, federationIssuer string, cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		def := types.Definition{
			ID:          DefinitionID.ID(),
			Family:      "storage",
			DisplayName: "Object Storage",
			Description: "Read and write objects in your cloud storage bucket and import its JSON records.",
			Category:    "data",
			DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/object-storage",
			Tags:        []string{"files", "import"},
			Active:      true,
			Visible:     false,
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: jsonx.SchemaFrom[Config](),
			},
			Installation: installation.Registration(),
			Connections: []types.Connector{
				workloadIdentity.
					Name("GCP Workload Identity Federation").
					Description("Configure Google Cloud Storage bucket access by trusting Openlane as an OIDC identity provider, so no service account key is ever stored.").
					Recommended().
					Meta(map[string]types.MetaInfo{
						"Openlane Issuer URI": {
							Value:     federationIssuer,
							AllowCopy: true,
						},
					}).
					Provides(workloadIdentityClient).
					Verified(verifyWorkloadIdentity).
					Disconnects("Removes the stored workload identity configuration from Openlane. If the workload identity pool is no longer needed, delete the pool and its provider from your Google Cloud project.", nil),
				serviceAccount.
					Name("GCP Service Account").
					Description("Configure Google Cloud Storage bucket access using a GCP service account key.").
					Provides(serviceAccountClient).
					Verified(verifyServiceAccount).
					Disconnects("Removes the stored service account key from Openlane. If the GCP service account is no longer needed, delete it from your Google Cloud project.", nil),
				awsAssumeRole.
					Name("AWS IAM Role").
					Description("Configure S3 bucket access using a cross-account IAM role that trusts Openlane with an external ID.").
					Meta(map[string]types.MetaInfo{
						"Openlane Principal ARN": {
							Value:     cfg.ARN,
							AllowCopy: true,
						},
					}).
					Provides(assumeRoleClient(cfg)).
					Verified(verifyAssumeRole).
					Disconnects("Removes the stored IAM role configuration from Openlane. If the cross-account IAM role is no longer needed, delete it from your AWS account.", nil),
				awsAccessKey.
					Name("AWS Access Keys").
					Description("Configure S3 bucket access using static IAM access keys.").
					Provides(accessKeyClient).
					Verified(verifyAccessKey).
					Disconnects("Removes the stored IAM access keys from Openlane. If the IAM user is no longer needed, delete it from your AWS account.", nil),
				r2.
					Name("Cloudflare R2").
					Description("Configure Cloudflare R2 bucket access using an R2 API token's S3-compatible keys.").
					Provides(r2Client).
					Verified(verifyR2).
					Disconnects("Removes the stored R2 keys from Openlane. If the R2 API token is no longer needed, revoke it in your Cloudflare account.", nil),
			},
			Operations: []types.OperationRegistration{
				ImportRecordsOp.
					Handles(runImportRecords).
					Validated(validateImportRecords).
					Policy(importPolicy).
					Ingest(importContracts...).
					SkipDefaultLookback().
					Description("Import the JSON records under a bucket prefix object by object, moving each imported object aside when configured").
					Registration(),
				types.OperationPayloadOf[WriteObject]().
					Handles(runWriteObject).
					Policy(types.ExecutionPolicy{Inline: true}).
					Description("Write a document into the bucket under a key").
					Registration(),
			},
			Mappings: []types.MappingRegistration{entityMapping},
		}

		if runtime.Provisioned() {
			config, err := jsonx.ToRawMessage(runtime)
			if err != nil {
				return types.Definition{}, fmt.Errorf("%w: %w", ErrRuntimeConfigDecode, err)
			}

			def.RuntimeIntegration = &types.RuntimeIntegrationRegistration{
				Schema: jsonx.SchemaFrom[RuntimeConfig](),
				Config: config,
				Build:  runtimeClientBuilder(),
			}
		}

		return def, nil
	})
}
