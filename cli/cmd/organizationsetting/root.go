//go:build cli

package orgsetting

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

// cmd represents the base org setting command when called without any subcommands
var command = &cobra.Command{
	Use:     "organization-setting",
	Aliases: []string{"org-setting"},
	Short:   "the subcommands for working with the organization settings",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the output in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the output and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetOrganizationSettings:
		var nodes []*graphclient.GetOrganizationSettings_OrganizationSettings_Edges_Node

		for _, i := range v.OrganizationSettings.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetAllOrganizationSettings:
		var nodes []*graphclient.GetAllOrganizationSettings_OrganizationSettings_Edges_Node

		for _, i := range v.OrganizationSettings.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetOrganizationSettingByID:
		e = v.OrganizationSetting
	case *graphclient.UpdateOrganizationSetting:
		e = v.UpdateOrganizationSetting.OrganizationSetting
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.OrganizationSetting

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.OrganizationSetting
		err = json.Unmarshal(s, &in)
		cobra.CheckErr(err)

		list = append(list, in)
	}

	tableOutput(list)

	return nil
}

// jsonOutput prints the output in a JSON format
func jsonOutput(out any) error {
	s, err := json.Marshal(out)
	cobra.CheckErr(err)

	return cmd.JSONPrint(s)
}

// tableOutput prints the output in a table format
func tableOutput(out []graphclient.OrganizationSetting) {
	writer := tables.NewTableWriter(command.OutOrStdout(),
		"ID",
		"OrganizationName",
		"BillingContact",
		"BillingAddress",
		"BillingEmail",
		"BillingPhone",
		"GeoLocation",
		"TaxIdentifier",
		"Tags",
		"Domains",
		"AllowedEmailDomains",
		"AllowMatchingDomainsAutojoin",
		"IdentityProvider",
		"IdentityProviderEntityID",
		"IdentityProviderLoginEnforced",
		"OidcDiscoveryEndpoint",
		"ComplianceWebhookToken",
		"IdentityProviderClientSecret",
		"IdentityProviderClientID",
		"IdentityProviderMetadataEndpoint",
	)

	for _, i := range out {
		var billingContact, billingEmail, billingPhone, geoLocation, taxIdentifier, identityProvider, identityProviderEntityID, oidcDiscoveryEndpoint, complianceWebhookToken, identityProviderClientSecret, identityProviderClientID, identityProviderMetadataEndpoint string
		if i.BillingContact != nil {
			billingContact = *i.BillingContact
		}
		if i.BillingEmail != nil {
			billingEmail = *i.BillingEmail
		}
		if i.BillingPhone != nil {
			billingPhone = *i.BillingPhone
		}
		if i.GeoLocation != nil {
			geoLocation = i.GeoLocation.String()
		}
		if i.TaxIdentifier != nil {
			taxIdentifier = *i.TaxIdentifier
		}
		if i.IdentityProvider != nil {
			identityProvider = i.IdentityProvider.String()
		}
		if i.IdentityProviderEntityID != nil {
			identityProviderEntityID = *i.IdentityProviderEntityID
		}
		if i.OidcDiscoveryEndpoint != nil {
			oidcDiscoveryEndpoint = *i.OidcDiscoveryEndpoint
		}
		if i.ComplianceWebhookToken != nil {
			complianceWebhookToken = *i.ComplianceWebhookToken
		}
		if i.IdentityProviderClientSecret != nil {
			identityProviderClientSecret = *i.IdentityProviderClientSecret
		}
		if i.IdentityProviderClientID != nil {
			identityProviderClientID = *i.IdentityProviderClientID
		}
		if i.IdentityProviderMetadataEndpoint != nil {
			identityProviderMetadataEndpoint = *i.IdentityProviderMetadataEndpoint
		}

		var billingAddress string
		if i.BillingAddress != nil {
			billingAddress = i.BillingAddress.String()
		}

		var orgName string
		if i.Organization != nil {
			orgName = i.Organization.DisplayName
		}

		writer.AddRow(i.ID,
			orgName,
			billingContact,
			billingAddress,
			billingEmail,
			billingPhone,
			geoLocation,
			taxIdentifier,
			strings.Join(i.Tags, ", "),
			strings.Join(i.Domains, ", "),
			strings.Join(i.AllowedEmailDomains, ", "),
			i.AllowMatchingDomainsAutojoin,
			identityProvider,
			identityProviderEntityID,
			i.IdentityProviderLoginEnforced,
			oidcDiscoveryEndpoint,
			complianceWebhookToken,
			identityProviderClientSecret,
			identityProviderClientID,
			identityProviderMetadataEndpoint,
		)
	}

	writer.Render()
}
