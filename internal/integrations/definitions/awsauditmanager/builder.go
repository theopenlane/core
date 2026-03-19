package awsauditmanager

import (
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the AWS Audit Manager definition builder
func Builder() definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Family:      "aws",
				DisplayName: "AWS Audit Manager",
				Description: "Collect AWS Audit Manager assessment metadata for compliance posture checks using STS role assumption.",
				Category:    "compliance",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/aws/overview",
				Labels:      map[string]string{"vendor": "aws", "service": "audit-manager"},
				Active:      true,
				Visible:     false,
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			Credentials: &types.CredentialRegistration{
				Schema: providerkit.SchemaFrom[CredentialSchema](),
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         AuditManagerClient.ID(),
					Description: "AWS Audit Manager client",
					Build:       Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Validate Audit Manager access via GetAccountStatus; confirms the assumed role can reach Audit Manager and reports enrollment status",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   AuditManagerClient.ID(),
					Handle:      HealthCheck{}.Handle(),
				},
				{
					Name:         AssessmentsListOperation.Name(),
					Description:  "List Audit Manager assessments with their compliance type, status, and evidence counts for compliance posture reporting",
					Topic:        AssessmentsListOperation.Topic(Slug),
					ClientRef:    AuditManagerClient.ID(),
					ConfigSchema: providerkit.SchemaFrom[AssessmentsConfig](),
					Handle:       AssessmentsList{}.Handle(),
				},
			},
		}, nil
	})
}
