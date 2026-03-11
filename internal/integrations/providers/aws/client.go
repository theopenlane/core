package aws

import (
	"context"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"

	"github.com/theopenlane/core/internal/integrations/providers/awskit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// clientConstructor is a function that creates a typed AWS client from an aws.Config
type clientConstructor[T any] func(cfg awssdk.Config) T

// buildAWSClient builds a typed AWS client from stored credentials using the provided constructor function
func buildAWSClient[T any](ctx context.Context, credential types.CredentialSet, constructor clientConstructor[T]) (T, awskit.Metadata, error) {
	var zero T

	meta, err := awsMetadataFromCredential(credential, awsDefaultSession)
	if err != nil {
		return zero, awskit.Metadata{}, err
	}

	cfg, err := awskit.BuildAWSConfig(ctx, meta.Region, awskit.CredentialsFromMetadata(meta), awskit.AssumeRole{
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
func resolveAWSClient[T any](ctx context.Context, input types.OperationInput, constructor clientConstructor[T]) (T, awskit.Metadata, error) {
	if client, ok := types.ClientInstanceAs[T](input.Client); ok {
		meta, err := awsMetadataFromCredential(input.Credential, awsDefaultSession)
		if err != nil {
			var zero T
			return zero, awskit.Metadata{}, err
		}

		return client, meta, nil
	}

	return buildAWSClient(ctx, input.Credential, constructor)
}
