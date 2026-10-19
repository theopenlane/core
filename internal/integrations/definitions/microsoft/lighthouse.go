package microsoft

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"

	"github.com/theopenlane/core/v2/pkg/logx"
)

const (
	// SecurityReaderRoleID is the Azure built-in Security Reader role definition ID
	SecurityReaderRoleID = "39bc4728-0917-49c7-9d2c-d95423bc2eb4"
	// RegistrationAssignmentDeleteRoleID is the Azure built-in role that lets the managing tenant read and remove its own Lighthouse delegations
	RegistrationAssignmentDeleteRoleID = "91c1777a-f3dc-4fae-b103-61d183457e46"
	// registrationAssignmentsPathFormat is the ARM path listing a subscription's Lighthouse registration assignments with their definitions
	registrationAssignmentsPathFormat = "/subscriptions/%s/providers/Microsoft.ManagedServices/registrationAssignments?api-version=" + lighthouseAPIVersion + "&$expandRegistrationDefinition=true"
	// lighthouseTemplateSchema is the ARM schema for subscription-scope deployments
	lighthouseTemplateSchema = "https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json#"
	// lighthouseTemplateContentVersion is our version of the Lighthouse template, bumped when the template changes
	lighthouseTemplateContentVersion = "1.0.0.0"
	// lighthouseAPIVersion is the Microsoft.ManagedServices API version used for Lighthouse requests
	lighthouseAPIVersion = "2022-10-01"
	// registrationDefinitionType is the ARM resource type of a Lighthouse registration definition
	registrationDefinitionType = "Microsoft.ManagedServices/registrationDefinitions"
	// registrationAssignmentType is the ARM resource type of a Lighthouse registration assignment
	registrationAssignmentType = "Microsoft.ManagedServices/registrationAssignments"
	// connectionCodeParameter is the template parameter the customer fills with the code shown in Openlane
	connectionCodeParameter = "connectionCode"
)

// LighthouseConfig holds the managing tenant settings customers delegate Azure subscriptions to through Azure Lighthouse
type LighthouseConfig struct {
	// ManagingTenantID is the Openlane tenant that customer subscriptions are delegated to
	ManagingTenantID string `json:"managingtenantid" koanf:"managingtenantid"`
	// PrincipalID is the object ID of the Openlane app's service principal in the managing tenant
	PrincipalID string `json:"principalid" koanf:"principalid"`
}

// LighthouseOffer describes the delegation a Lighthouse template grants the managing tenant principal
type LighthouseOffer struct {
	// NamePrefix starts the offer name shown to the customer; the connection code completes it
	NamePrefix string
	// Description is the offer description shown to the customer
	Description string
	// PrincipalDisplayName labels the managing tenant principal in the customer's Service providers page
	PrincipalDisplayName string
	// RoleDefinitionIDs are the built-in roles granted to the principal on the delegated subscription
	RoleDefinitionIDs []string
}

// OfferName returns the offer name a template deployment with the connection code registers
func (o LighthouseOffer) OfferName(code string) string {
	return o.NamePrefix + " " + code
}

// armTemplate is a subscription-scope ARM deployment template
type armTemplate struct {
	// Schema is the template schema URL
	Schema string `json:"$schema"`
	// ContentVersion is the template version
	ContentVersion string `json:"contentVersion"`
	// Parameters are the values supplied at deployment
	Parameters map[string]armParameter `json:"parameters"`
	// Variables are the expressions derived from the parameters
	Variables map[string]string `json:"variables"`
	// Resources are the resources the template deploys
	Resources []armResource `json:"resources"`
}

// armParameter is one ARM template parameter
type armParameter struct {
	// Type is the parameter type
	Type string `json:"type"`
	// Metadata describes the parameter in the deployment form
	Metadata armMetadata `json:"metadata"`
}

// armMetadata describes an ARM template parameter
type armMetadata struct {
	// Description is shown next to the parameter in the deployment form
	Description string `json:"description"`
}

// armResource is one resource deployed by an ARM template
type armResource struct {
	// Type is the ARM resource type
	Type string `json:"type"`
	// APIVersion is the resource provider API version
	APIVersion string `json:"apiVersion"`
	// Name is the resource name expression
	Name string `json:"name"`
	// DependsOn lists the resources deployed first
	DependsOn []string `json:"dependsOn,omitempty"`
	// Properties are the resource properties
	Properties any `json:"properties"`
}

// lighthouseAuthorization is one principal and role pair in a registration definition
type lighthouseAuthorization struct {
	// PrincipalID is the managing tenant principal object ID
	PrincipalID string `json:"principalId"`
	// PrincipalIDDisplayName labels the principal for the customer
	PrincipalIDDisplayName string `json:"principalIdDisplayName"`
	// RoleDefinitionID is the built-in role granted to the principal
	RoleDefinitionID string `json:"roleDefinitionId"`
}

// registrationDefinitionProperties are the properties of a Lighthouse registration definition
type registrationDefinitionProperties struct {
	// RegistrationDefinitionName is the offer name
	RegistrationDefinitionName string `json:"registrationDefinitionName"`
	// Description is the offer description
	Description string `json:"description,omitempty"`
	// ManagedByTenantID is the managing tenant the subscription is delegated to
	ManagedByTenantID string `json:"managedByTenantId"`
	// ManageeTenantID is the customer tenant that owns the delegated subscription, set only on read
	ManageeTenantID string `json:"manageeTenantId,omitempty"`
	// Authorizations are the principals and roles the delegation grants
	Authorizations []lighthouseAuthorization `json:"authorizations,omitempty"`
}

// registrationDefinition is a Lighthouse registration definition
type registrationDefinition struct {
	// Properties are the definition properties
	Properties registrationDefinitionProperties `json:"properties"`
}

// registrationAssignmentProperties are the properties of a Lighthouse registration assignment
type registrationAssignmentProperties struct {
	// RegistrationDefinitionID is the definition the assignment applies
	RegistrationDefinitionID string `json:"registrationDefinitionId"`
	// RegistrationDefinition is the expanded definition, set only on read
	RegistrationDefinition *registrationDefinition `json:"registrationDefinition,omitempty"`
}

// registrationAssignment is a Lighthouse registration assignment
type registrationAssignment struct {
	// ID is the fully qualified ARM path of the assignment, set only on read
	ID string `json:"id,omitempty"`
	// Properties are the assignment properties
	Properties registrationAssignmentProperties `json:"properties"`
}

// registrationAssignmentPage is one page of a subscription's registration assignments
type registrationAssignmentPage struct {
	// Value is the assignments on this page
	Value []registrationAssignment `json:"value"`
	// NextLink is the absolute URL of the next page, empty on the last page
	NextLink string `json:"nextLink"`
}

// Template returns the subscription-scope ARM template that delegates a subscription to the managing tenant principal for the offer
func (c LighthouseConfig) Template(offer LighthouseOffer) (string, error) {
	authorizations := make([]lighthouseAuthorization, 0, len(offer.RoleDefinitionIDs))
	for _, roleID := range offer.RoleDefinitionIDs {
		authorizations = append(authorizations, lighthouseAuthorization{
			PrincipalID:            c.PrincipalID,
			PrincipalIDDisplayName: offer.PrincipalDisplayName,
			RoleDefinitionID:       roleID,
		})
	}

	definitionID := "[resourceId('" + registrationDefinitionType + "/', variables('mspRegistrationName'))]"

	template := armTemplate{
		Schema:         lighthouseTemplateSchema,
		ContentVersion: lighthouseTemplateContentVersion,
		Parameters: map[string]armParameter{
			connectionCodeParameter: {Type: "string", Metadata: armMetadata{Description: "Connection code shown in Openlane"}},
		},
		Variables: map[string]string{
			"mspOfferName":        "[concat('" + offer.NamePrefix + " ', parameters('" + connectionCodeParameter + "'))]",
			"mspRegistrationName": "[guid(variables('mspOfferName'))]",
		},
		Resources: []armResource{
			{
				Type:       registrationDefinitionType,
				APIVersion: lighthouseAPIVersion,
				Name:       "[variables('mspRegistrationName')]",
				Properties: registrationDefinitionProperties{
					RegistrationDefinitionName: "[variables('mspOfferName')]",
					Description:                offer.Description,
					ManagedByTenantID:          c.ManagingTenantID,
					Authorizations:             authorizations,
				},
			},
			{
				Type:       registrationAssignmentType,
				APIVersion: lighthouseAPIVersion,
				Name:       "[variables('mspRegistrationName')]",
				DependsOn:  []string{definitionID},
				Properties: registrationAssignmentProperties{RegistrationDefinitionID: definitionID},
			},
		},
	}

	out, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return "", ErrLighthouseTemplateBuild
	}

	return string(out), nil
}

// Delegation is one Lighthouse registration assignment on a subscription
type Delegation struct {
	// AssignmentID is the fully qualified ARM path of the registration assignment
	AssignmentID string
	// OfferName is the registration definition name
	OfferName string
}

// FindDelegation returns the delegation whose offer name matches name, ignoring case
func FindDelegation(delegations []Delegation, name string) (Delegation, bool) {
	for _, d := range delegations {
		if strings.EqualFold(d.OfferName, name) {
			return d, true
		}
	}

	return Delegation{}, false
}

// ListDelegations returns the subscription's Lighthouse delegations to managedByTenantID
func (c *ARMClient) ListDelegations(ctx context.Context, subscriptionID, managedByTenantID string) ([]Delegation, error) {
	var delegations []Delegation

	next := c.client.Endpoint() + fmt.Sprintf(registrationAssignmentsPathFormat, subscriptionID)
	for next != "" {
		page, err := getARM[registrationAssignmentPage](ctx, c.client.Pipeline(), next, ErrRegistrationAssignmentListFailed)
		if err != nil {
			return nil, err
		}

		for _, assignment := range page.Value {
			definition := assignment.Properties.RegistrationDefinition
			if definition != nil && strings.EqualFold(definition.Properties.ManagedByTenantID, managedByTenantID) {
				delegations = append(delegations, Delegation{AssignmentID: assignment.ID, OfferName: definition.Properties.RegistrationDefinitionName})
			}
		}

		next = page.NextLink
	}

	return delegations, nil
}

// DeleteDelegation removes a Lighthouse registration assignment, which requires the Registration Assignment Delete role
func (c *ARMClient) DeleteDelegation(ctx context.Context, assignmentID string) error {
	requestURL := c.client.Endpoint() + assignmentID + "?api-version=" + lighthouseAPIVersion

	req, err := runtime.NewRequest(ctx, http.MethodDelete, requestURL)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("url", requestURL).Msg("failed to build registration assignment delete request")
		return ErrRegistrationAssignmentDeleteFailed
	}

	resp, err := c.client.Pipeline().Do(req)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("url", requestURL).Msg("registration assignment delete request failed")
		return ErrRegistrationAssignmentDeleteFailed
	}

	defer resp.Body.Close()

	if !runtime.HasStatusCode(resp, http.StatusOK, http.StatusAccepted, http.StatusNoContent) {
		logx.FromContext(ctx).Error().Err(runtime.NewResponseError(resp)).Str("url", requestURL).Msg("registration assignment delete returned unexpected status")
		return ErrRegistrationAssignmentDeleteFailed
	}

	return nil
}
