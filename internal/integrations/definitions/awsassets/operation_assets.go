package awsassets

import (
	"context"
	"encoding/json"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/theopenlane/core/internal/integrations/definitions/awskit"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// AssetCollect collects AWS asset inventory
type AssetCollect struct {
	// AccountID is the AWS account identifier
	AccountID string `json:"accountId"`
	// Region is the AWS region used for the session
	Region string `json:"region"`
	// Message describes the collection readiness state
	Message string `json:"message"`
}

// Handle adapts asset collection to the generic operation registration boundary
func (a AssetCollect) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return a.Run(ctx, request.Credential, c)
	}
}

// Run collects AWS asset inventory
func (AssetCollect) Run(ctx context.Context, credential types.CredentialSet, c *sts.Client) (json.RawMessage, error) {
	meta, err := awskit.MetadataFromProviderData(credential.ProviderData, awsDefaultSessionName)
	if err != nil {
		return nil, ErrCredentialMetadataInvalid
	}

	resp, err := c.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, ErrIdentityVerificationFailed
	}

	return providerkit.EncodeResult(AssetCollect{
		AccountID: awssdk.ToString(resp.Account),
		Region:    meta.Region,
		Message:   "aws asset collection ready",
	}, ErrResultEncode)
}
