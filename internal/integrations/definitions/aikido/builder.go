package aikido

import (
	"github.com/theopenlane/utils/jsonx"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/auth"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// Builder returns the Aikido definition builder with the operator OAuth app config applied
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			ID:          definitionID.ID(),
			Family:      "Aikido",
			DisplayName: "Aikido Security",
			Description: "Collect Aikido Security vulnerabilities, code, cloud and container findings, and the assets Aikido scans.",
			Category:    "security-posture",
			DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/aikido",
			Tags:        []string{"vulnerabilities", "findings", "assets"},
			Active:      true,
			Visible:     true,
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: jsonx.SchemaFrom[Config](),
			},
			Installation: installation.Registration(),
			Connections: []types.Connector{
				oauthConnection.
					Name("Aikido OAuth").
					Recommended().
					Description("Connect your Aikido workspace by authorizing the Openlane app.").
					Authenticates(auth.OAuthRegistration(auth.OAuthRegistrationOptions[aikidoOAuthCred]{
						Config: auth.OAuthConfig{ //nolint:gosec
							ClientID:     cfg.ClientID,
							ClientSecret: cfg.ClientSecret,
							AuthURL:      oauthEndpoint(oauthAppURL).AuthURL,
							TokenURL:     oauthEndpoint(oauthAppURL).TokenURL,
							RedirectURL:  cfg.RedirectURL,
						},
						Material:              oauthMaterial,
						EncodeCredentialError: ErrCredentialEncode,
					})).
					Provides(oauthClientBuilder(cfg)).
					Verified(verifyOAuth).
					Disconnects("Removes the stored OAuth credential from Openlane. To fully revoke access, remove the Openlane app from your Aikido workspace settings.", nil),
				aikidoConnection.
					Name("Aikido REST API Client").
					Description("Configure Aikido access using REST API client credentials created by a workspace admin under Settings > Integrations > REST API.").
					Provides(buildClient).
					Verified(verify).
					Disconnects("Removes the stored Aikido client credentials from Openlane. If they are no longer needed, revoke them from your Aikido REST API settings.", nil),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[AssetSync]().
					Ingests(runAssetSync).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Schedule(gala.NewFullFetchSchedule()).
					SkipDefaultLookback().
					Ingest(types.IngestContract{Schema: entityops.SchemaAsset.Name}).
					Permissions("Client credentials with the basics:read, repositories:read, clouds:read and domains:read scopes, plus virtual_machines:read and endpoint_protection:read to include virtual machines and devices").
					Description("Collect Aikido code repositories, containers, clouds, domains, subdomains, virtual machines and devices as assets").
					Registration(),
				types.OperationRefOf[IssueSync]().
					Ingests(runIssueSync).
					HealthCheck(probeAccess).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Schedule(gala.NewFullFetchSchedule()).
					SkipDefaultLookback().
					Ingest(
						types.IngestContract{Schema: entityops.SchemaVulnerability.Name},
						types.IngestContract{Schema: entityops.SchemaFinding.Name},
					).
					Permissions("Client credentials with the basics:read, issues:read and domains:read scopes, plus research:read for CVE details and pentests:read for pentest attack analyses").
					Description("Collect Aikido open source issues as vulnerabilities, and every other issue type and leaked credentials as findings").
					Registration(),
				types.OperationRefOf[AutofixSync]().
					Ingests(runAutofixSync).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Schedule(gala.NewFullFetchSchedule()).
					SkipDefaultLookback().
					Ingest(types.IngestContract{Schema: entityops.SchemaRemediation.Name}).
					Permissions("Client credentials with the basics:read and autofix:read scopes").
					Description("Collect Aikido autofix tasks as remediations linked to the issues they fix").
					Registration(),
				types.OperationRefOf[ComplianceSync]().
					Ingests(runComplianceSync).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Schedule(gala.NewFullFetchSchedule()).
					SkipDefaultLookback().
					Ingest(types.IngestContract{Schema: entityops.SchemaCheckResult.Name}).
					Permissions("Client credentials with the basics:read and reports:read scopes").
					Description("Collect Aikido SOC 2, ISO 27001, PCI DSS, GDPR, NIS2, DORA and CIS compliance report rules as check results linked to their controls").
					Registration(),
				types.OperationRefOf[ScanSync]().
					Ingests(runScanSync).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Schedule(gala.NewFullFetchSchedule()).
					SkipDefaultLookback().
					Ingest(types.IngestContract{Schema: entityops.SchemaScan.Name}).
					Permissions("Client credentials with the basics:read and pentests:read scopes").
					Description("Collect Aikido AI pentest and code analysis assessments as scans linked to the findings they produced").
					Registration(),
				types.OperationRefOf[providerkit.DirectorySync]().
					Ingests(runDirectorySync).
					Policy(types.ExecutionPolicy{Reconcile: true, Snapshot: true}).
					Schedule(gala.NewFullFetchSchedule()).
					SkipDefaultLookback().
					Ingest(providerkit.DirectoryIngestContracts()...).
					Permissions("Client credentials with the basics:read, users:read and teams:read scopes").
					Description("Sync Aikido workspace users as directory accounts and teams as groups with their members").
					Registration(),
			},
			Mappings: append([]types.MappingRegistration{
				providerkit.VulnerabilityMapping(mapExprVulnerability),
				providerkit.FindingMapping(mapExprFinding),
				providerkit.FindingVariantMapping(findingVariantDataleak, mapExprDataleakFinding),
				providerkit.RemediationMapping(mapExprRemediation),
				providerkit.CheckResultMapping(mapExprCheckResult),
				providerkit.ScanMapping(mapExprScan),
				providerkit.AssetMapping(assetVariantCodeRepository, mapExprCodeRepositoryAsset),
				providerkit.AssetMapping(assetVariantContainer, mapExprContainerAsset),
				providerkit.AssetMapping(assetVariantCloud, mapExprCloudAsset),
				providerkit.AssetMapping(assetVariantDomain, mapExprDomainAsset),
				providerkit.AssetMapping(assetVariantSubdomain, mapExprSubdomainAsset),
				providerkit.AssetMapping(assetVariantVirtualMachine, mapExprVirtualMachineAsset),
				providerkit.AssetMapping(assetVariantDevice, mapExprDeviceAsset),
			}, providerkit.DirectoryMappings(mapExprDirectoryAccount, mapExprDirectoryGroup, mapExprDirectoryMembership)...),
		}, nil
	})
}

// oauthMaterial maps completed OAuth flow material to the stored Aikido credential
func oauthMaterial(material auth.OAuthMaterial) (aikidoOAuthCred, error) {
	return aikidoOAuthCred{
		AccessToken:  material.AccessToken,
		RefreshToken: material.RefreshToken,
		Expiry:       material.Expiry,
	}, nil
}
