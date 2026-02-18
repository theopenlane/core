package aws

import (
	"context"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

// ClientConstructor is a function that creates a typed AWS client from an aws.Config
type ClientConstructor[T any] func(cfg awssdk.Config) T

// buildAWSClient builds a typed AWS client from stored credentials using the provided constructor function
func buildAWSClient[T any](ctx context.Context, payload types.CredentialPayload, constructor ClientConstructor[T]) (T, auth.AWSMetadata, error) {
	var zero T

	meta, err := awsMetadataFromPayload(payload, awsDefaultSession)
	if err != nil {
		return zero, auth.AWSMetadata{}, err
	}

	cfg, err := auth.BuildAWSConfig(ctx, meta.Region, auth.AWSCredentialsFromPayload(payload), auth.AWSAssumeRole{
		RoleARN:         meta.RoleARN,
		ExternalID:      meta.ExternalID,
		SessionName:     meta.SessionName,
		SessionDuration: meta.SessionDuration,
	})
	if err != nil {
		return zero, meta, err
	}

	return constructor(cfg), meta, nil
}

// resolveAWSClient returns a pooled client when supplied or builds one on demand
func resolveAWSClient[T any](ctx context.Context, input types.OperationInput, constructor ClientConstructor[T]) (T, auth.AWSMetadata, error) {
	if client, ok := input.Client.(T); ok {
		meta, err := awsMetadataFromPayload(input.Credential, awsDefaultSession)
		if err != nil {
			var zero T
			return zero, auth.AWSMetadata{}, err
		}

		return client, meta, nil
	}

	return buildAWSClient(ctx, input.Credential, constructor)
}
