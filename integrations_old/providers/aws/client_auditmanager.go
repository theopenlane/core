package aws

import (
	"context"
	"encoding/json"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers/awskit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// ClientAWSAuditManager identifies the AWS Audit Manager client descriptor
	ClientAWSAuditManager types.ClientName = "auditmanager"
)

// awsAuditManagerClientDescriptors returns the AWS Audit Manager client descriptors for pooling
func awsAuditManagerClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeAWS, ClientAWSAuditManager, "AWS Audit Manager client", pooledAuditManagerClient)
}

// pooledAuditManagerClient builds the AWS Audit Manager client for pooling, discarding metadata
func pooledAuditManagerClient(ctx context.Context, credential types.CredentialSet, _ json.RawMessage) (types.ClientInstance, error) {
	client, _, err := buildAuditManagerClient(ctx, credential)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	return types.NewClientInstance(client), nil
}

// newAuditManagerSDKClient wraps auditmanager.NewFromConfig for use with generic helpers
func newAuditManagerSDKClient(cfg awssdk.Config) *auditmanager.Client {
	return auditmanager.NewFromConfig(cfg)
}

// buildAuditManagerClient constructs an Audit Manager client from the stored credential
func buildAuditManagerClient(ctx context.Context, credential types.CredentialSet) (*auditmanager.Client, awskit.Metadata, error) {
	return buildAWSClient(ctx, credential, newAuditManagerSDKClient)
}

// resolveAuditManagerClient returns a pooled client when available or builds one on demand
func resolveAuditManagerClient(ctx context.Context, input types.OperationInput) (*auditmanager.Client, awskit.Metadata, error) {
	return resolveAWSClient(ctx, input, newAuditManagerSDKClient)
}
