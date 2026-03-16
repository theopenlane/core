package awsassets

import (
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID           = types.NewDefinitionRef("def_01K0AWSAST00000000000000001")
	AWSAssetsClient        = types.NewClientRef[*sts.Client]()
	HealthDefaultOperation = types.NewOperationRef[HealthCheck]("health.default")
	AssetCollectOperation  = types.NewOperationRef[AssetCollect]("asset.collect")
)

// Slug is the unique identifier for the AWS Assets integration
const Slug = "aws_assets"
