package azuresecuritycenter

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	// subscriptionScopeFormat is the ARM subscription scope format for Security Center requests
	subscriptionScopeFormat = "/subscriptions/%s"
)

// azureSecurityClient wraps the armsecurity clients with the installation subscription scope
type azureSecurityClient struct {
	// assessments is the ARM Security Center assessments client
	assessments *armsecurity.AssessmentsClient
	// subassessments is the ARM Security Center sub-assessments client
	subassessments *armsecurity.SubAssessmentsClient
	// subscriptionID is the Azure subscription identifier used to construct the ARM scope
	subscriptionID string
}

// scope returns the ARM subscription scope string for this installation
func (c *azureSecurityClient) scope() string {
	return fmt.Sprintf(subscriptionScopeFormat, c.subscriptionID)
}

// Client builds Azure Security Center clients for one installation
type Client struct{}

// Build constructs an Azure Security Center client using client credentials
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	var cred CredentialSchema
	if err := jsonx.UnmarshalIfPresent(req.Credential.ProviderData, &cred); err != nil {
		return nil, ErrCredentialInvalid
	}

	switch {
	case cred.TenantID == "":
		return nil, ErrTenantIDMissing
	case cred.ClientID == "":
		return nil, ErrClientIDMissing
	case cred.ClientSecret == "":
		return nil, ErrClientSecretMissing
	case cred.SubscriptionID == "":
		return nil, ErrSubscriptionIDMissing
	}

	azCred, err := azidentity.NewClientSecretCredential(cred.TenantID, cred.ClientID, cred.ClientSecret, nil)
	if err != nil {
		return nil, ErrCredentialBuildFailed
	}

	assessments, err := armsecurity.NewAssessmentsClient(azCred, nil)
	if err != nil {
		return nil, ErrAssessmentsClientBuildFailed
	}

	subassessments, err := armsecurity.NewSubAssessmentsClient(azCred, nil)
	if err != nil {
		return nil, ErrAssessmentsClientBuildFailed
	}

	return &azureSecurityClient{
		assessments:    assessments,
		subassessments: subassessments,
		subscriptionID: cred.SubscriptionID,
	}, nil
}
