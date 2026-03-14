package aws

import (
	"context"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/awssts"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TypeAWS identifies the consolidated AWS provider
const TypeAWS types.ProviderType = "aws"

// awsCredentialsSchema is the JSON Schema for AWS STS credentials.
var awsCredentialsSchema = []byte(`{"type":"object","additionalProperties":false,"required":["roleArn","externalId","homeRegion"],"allOf":[{"if":{"properties":{"accountScope":{"const":"specific"}},"required":["accountScope"]},"then":{"required":["accountIds"]}}],"properties":{"alias":{"type":"string","title":"Credential Alias","description":"Friendly identifier for this AWS integration."},"roleArn":{"type":"string","title":"IAM Role ARN","description":"Cross-account role Openlane should assume in the tenant environment."},"externalId":{"type":"string","title":"External ID","description":"External ID required in the tenant role trust policy to prevent confused deputy attacks."},"homeRegion":{"type":"string","title":"Security Hub Home Region","description":"Primary AWS region where Security Hub aggregation is managed.","default":"us-east-1"},"region":{"type":"string","title":"AWS Region (Legacy)","description":"Legacy alias for home region; use Security Hub Home Region when possible.","default":"us-east-1"},"linkedRegions":{"type":"array","title":"Linked Regions","description":"Optional list of regions to filter findings by region.","items":{"type":"string"}},"organizationId":{"type":"string","title":"AWS Organization ID","description":"Optional AWS Organizations identifier (for traceability and scoping)."},"accountScope":{"type":"string","title":"Account Scope","description":"Use all accessible accounts from the delegated admin role, or limit to specific account IDs.","default":"all","enum":["all","specific"]},"accountIds":{"type":"array","title":"Account IDs","description":"Required when Account Scope is set to specific.","items":{"type":"string"}},"sessionDuration":{"type":"string","title":"Session Duration","description":"Optional session duration (Go duration string, e.g. 1h30m)."},"sessionName":{"type":"string","title":"Session Name","description":"Optional session name override for STS AssumeRole calls."},"accessKeyId":{"type":"string","title":"Access Key ID","description":"Optional source credential key when Openlane cannot use runtime IAM credentials."},"secretAccessKey":{"type":"string","title":"Secret Access Key","description":"Optional source credential secret paired with Access Key ID.","secret":true},"sessionToken":{"type":"string","title":"Session Token","description":"Optional source session token when using temporary source credentials.","secret":true},"accountId":{"type":"string","title":"Account ID","description":"Optional AWS account identifier for reference."},"tags":{"type":"object","title":"Default Tags","description":"Optional key/value map added to generated integrations for traceability.","additionalProperties":{"type":"string"}}}}`)

// Builder returns the AWS provider builder
func Builder() providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeAWS,
		SpecFunc:     awsSpec,
		BuildFunc: func(ctx context.Context, s spec.ProviderSpec) (types.Provider, error) {
			return awssts.Builder(
				TypeAWS,
				awssts.WithOperations(awsOperations()),
				awssts.WithClientDescriptors(awsClientDescriptors()),
			).Build(ctx, s)
		},
	}
}

// awsSpec returns the static provider specification for the AWS provider.
func awsSpec() spec.ProviderSpec {
	return spec.ProviderSpec{
		Name:        "aws",
		DisplayName: "Amazon Web Services",
		Category:    "cloud",
		AuthType:    types.AuthKindAWSFederation,
		Active:      lo.ToPtr(true),
		Visible:     lo.ToPtr(true),
		LogoURL:     "",
		DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/aws/overview",
		Labels: map[string]string{
			"vendor": "aws",
		},
		CredentialsSchema: awsCredentialsSchema,
		Description:       "Collect AWS security and compliance findings and posture from Security Hub and Audit Manager data.",
	}
}
