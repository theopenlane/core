package awsauditmanager

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID             = types.NewDefinitionRef("def_01K0AWSAUDITM0000000000001")
	HealthDefaultOperation   = types.NewOperationRef[struct{}]("health.default")
	AssessmentsListOperation = types.NewOperationRef[struct{}]("assessments.list")
)

const Slug = "aws_audit_manager"

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label        string `json:"label,omitempty"       jsonschema:"title=Installation Label"`
	AssessmentID string `json:"assessmentId,omitempty" jsonschema:"title=Assessment ID,description=Optional assessment ID to scope collection to a single assessment."`
}

// credential holds the AWS STS role and optional static key material for one Audit Manager installation
// Fields are named to match awskit.awsProviderData JSON tags so MetadataFromProviderData decodes them correctly
// Fields from the v1 credential schema that were never implemented (systemsManagerDocument,
// configAggregatorName, controlTowerRegions) are intentionally omitted
type credential struct {
	RoleARN         string `json:"roleArn"                   jsonschema:"required,title=IAM Role ARN,description=Cross-account role Openlane should assume in the tenant environment."`
	ExternalID      string `json:"externalId"                jsonschema:"required,title=External ID,description=External ID required in the tenant role trust policy."`
	HomeRegion      string `json:"homeRegion"                jsonschema:"required,title=Home Region,description=AWS region where Audit Manager data is managed (e.g. us-east-1)."`
	AccountID       string `json:"accountId,omitempty"       jsonschema:"title=Account ID,description=AWS account ID for reference."`
	SessionName     string `json:"sessionName,omitempty"     jsonschema:"title=Session Name,description=Optional STS session name override."`
	SessionDuration string `json:"sessionDuration,omitempty" jsonschema:"title=Session Duration,description=Optional STS session duration (e.g. 1h)."`
	AccessKeyID     string `json:"accessKeyId,omitempty"     jsonschema:"title=Access Key ID,description=Optional static source credential when runtime IAM is unavailable."`
	SecretAccessKey string `json:"secretAccessKey,omitempty" jsonschema:"title=Secret Access Key"`
	SessionToken    string `json:"sessionToken,omitempty"    jsonschema:"title=Session Token"`
}

// Builder returns the AWS Audit Manager definition builder
func Builder() definition.Builder {
	return definition.Builder(func(_ context.Context) (types.Definition, error) {
		clientRef := types.NewClientRef[any]()

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Version:     "v1",
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
				Schema: providerkit.SchemaFrom[userInput](),
			},
			Credentials: &types.CredentialRegistration{
				Schema: providerkit.SchemaFrom[credential](),
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         clientRef.ID(),
					Description: "AWS Audit Manager client",
					Build:       buildAuditManagerClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Validate Audit Manager access via GetAccountStatus; confirms the assumed role can reach Audit Manager and reports enrollment status",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   clientRef.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
				{
					Name:         AssessmentsListOperation.Name(),
					Description:  "List Audit Manager assessments with their compliance type, status, and evidence counts for compliance posture reporting",
					Topic:        AssessmentsListOperation.Topic(Slug),
					ClientRef:    clientRef.ID(),
					ConfigSchema: providerkit.SchemaFrom[assessmentsConfig](),
					Policy:       types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Handle:       runAssessmentsListOperation,
				},
			},
		}, nil
	})
}
