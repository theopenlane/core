package awssecurityhub

import (
	"github.com/aws/aws-sdk-go-v2/service/securityhub"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID                    = types.NewDefinitionRef("def_01K0AWSSECHUB0000000000001")
	SecurityHubClient               = types.NewClientRef[*securityhub.Client]()
	HealthDefaultOperation          = types.NewOperationRef[HealthCheck]("health.default")
	VulnerabilitiesCollectOperation = types.NewOperationRef[VulnerabilitiesCollect]("vulnerabilities.collect")
)

// Slug is the unique identifier for the AWS Security Hub integration
const Slug = "aws_security_hub"
