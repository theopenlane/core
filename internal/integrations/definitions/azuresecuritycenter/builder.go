package azuresecuritycenter

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/microsoft"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// Builder returns the Azure Security Center definition builder with the supplied operator config applied
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		template, err := cfg.Lighthouse.Template(lighthouseOffer)
		if err != nil {
			return types.Definition{}, err
		}

		return types.Definition{
			ID:          definitionID.ID(),
			Family:      "Azure",
			DisplayName: "Microsoft Defender for Cloud",
			Description: "Collect security assessment findings and vulnerability data from Microsoft Defender for Cloud across an Azure subscription.",
			Category:    "security-posture",
			DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/azure_security_center",
			Tags:        []string{"vulnerabilities", "findings", "checks"},
			Active:      true,
			Visible:     true,
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: jsonx.SchemaFrom[Config](),
			},
			UserInput:    userInput.Registration(),
			Installation: installation.Registration(),
			Connections: []types.Connector{
				lighthouseConnection.
					Name("Azure Lighthouse").
					Description("Delegate read-only Defender for Cloud access with Azure Lighthouse. In the Azure portal open Service providers, choose Add offer, Add via template, and deploy the template below on each subscription using the connection code from this form. Nothing is installed in your tenant.").
					Recommended().
					Meta(map[string]types.MetaInfo{
						"Managing Tenant ID":      {Value: cfg.Lighthouse.ManagingTenantID, AllowCopy: true},
						"Lighthouse ARM Template": {Value: template, AllowCopy: true},
					}).
					Provides(lighthouseClientBuilder(cfg)).
					Verified(verifyLighthouse).
					Disconnects("Removes the stored connection from Openlane and removes the Openlane Lighthouse delegation from each subscription delegated with this connection code. If removal fails, remove the Openlane Defender for Cloud offer from Service providers in the Azure portal.", disconnectLighthouse(cfg)),
				securityCenterConnection.
					Name("Azure Admin Consent").
					Description("Connect your Azure tenant using admin consent, then grant the Openlane app the read-only Security Reader role on your subscriptions.").
					Meta(map[string]types.MetaInfo{
						"Application ID": {Value: cfg.ApplicationID, AllowCopy: true},
					}).
					Authenticates(microsoft.AdminConsentFlow(cfg.AppConfig, newSecurityCenterCred)).
					Provides(clientBuilder(cfg)).
					Verified(verify).
					Disconnects("Removes the stored credential from Openlane. To fully revoke access, remove the Openlane app from your Azure enterprise applications and its subscription role assignments.", nil),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[AssessmentsCollect]().
					Ingests(runAssessmentsCollect).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaFinding.Name}).
					HealthCheck(probeAssessments).
					Permissions("SecurityReader").
					Description("Collect unhealthy security posture assessment findings for finding ingestion").
					Registration(),
				types.OperationRefOf[VulnerabilitiesCollect]().
					Ingests(runVulnerabilitiesCollect).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaVulnerability.Name}).
					HealthCheck(probeAssessments).
					Permissions("SecurityReader").
					Description("Collect software update and service upgrade recommendations with their CVEs for vulnerability ingestion").
					Registration(),
				types.OperationRefOf[ComplianceAssessmentsCollect]().
					Ingests(runComplianceAssessmentsCollect).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(
						types.IngestContract{Schema: entityops.SchemaCheckResult.Name},
						types.IngestContract{Schema: entityops.SchemaFinding.Name},
					).
					HealthCheck(probeAssessments).
					Permissions("SecurityReader").
					Description("Collect regulatory compliance assessments such as SOC 2 as checks linked to their controls, with failed assessments as findings").
					Registration(),
			},
			Mappings: []types.MappingRegistration{
				providerkit.VulnerabilityMapping(mapExprVulnerability),
				providerkit.FindingMapping(mapExprFinding),
				providerkit.CheckResultMapping(mapExprCheckResult),
			},
		}, nil
	})
}
