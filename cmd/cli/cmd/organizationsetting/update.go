//go:build cli

package orgsetting

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
	"github.com/theopenlane/shared/enums"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing organization setting",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "org setting id to update")
	updateCmd.Flags().StringSliceP("domains", "d", []string{}, "domains associated with the org")
	updateCmd.Flags().StringP("billing-contact", "c", "", "billing contact for the org")
	updateCmd.Flags().StringP("billing-email", "e", "", "billing email for the org")
	updateCmd.Flags().StringP("billing-phone", "p", "", "billing phone for the org")
	updateCmd.Flags().StringP("tax-identifier", "x", "", "tax identifier for the org")
	updateCmd.Flags().StringSliceP("tags", "t", []string{}, "tags associated with the org")
	updateCmd.Flags().StringSliceP("append-allowed-domains", "a", []string{}, "emails domains allowed to access the org")
	updateCmd.Flags().BoolP("clear-allowed-domains", "l", false, "clear emails domains allowed to access the org")
	updateCmd.Flags().String("identity-provider", "", "SSO provider type")
	updateCmd.Flags().String("client-id", "", "OIDC client ID")
	updateCmd.Flags().String("client-secret", "", "OIDC client secret")
	updateCmd.Flags().String("metadata-url", "", "SSO metadata URL")
	updateCmd.Flags().String("discovery-url", "", "OIDC discovery URL")
	updateCmd.Flags().Bool("enforce-sso", false, "enforce SSO login")
	updateCmd.Flags().Bool("allow-matching-domains-autojoin", false, "allow users with matching email domains to automatically join the organization")
	updateCmd.Flags().StringSlice("set-allowed-email-domains", []string{}, "set allowed email domains for auto-join (replaces existing list)")
}

// updateValidation validates the input flags provided by the user
func updateValidation() (id string, input openlaneclient.UpdateOrganizationSettingInput, err error) {
	id = cmd.Config.String("id")

	billingContact := cmd.Config.String("billing-contact")
	if billingContact != "" {
		input.BillingContact = &billingContact
	}

	billingEmail := cmd.Config.String("billing-email")
	if billingEmail != "" {
		input.BillingEmail = &billingEmail
	}

	billingPhone := cmd.Config.String("billing-phone")
	if billingPhone != "" {
		input.BillingPhone = &billingPhone
	}

	taxIdentifier := cmd.Config.String("tax-identifier")
	if taxIdentifier != "" {
		input.TaxIdentifier = &taxIdentifier
	}

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	domains := cmd.Config.Strings("domains")
	if len(domains) > 0 {
		input.Domains = domains
	}

	allowedDomains := cmd.Config.Strings("append-allowed-domains")
	clearDomains := cmd.Config.Bool("clear-allowed-domains")

	if len(allowedDomains) > 0 {
		input.AllowedEmailDomains = allowedDomains
	} else if clearDomains {
		input.ClearAllowedEmailDomains = &clearDomains
	}

	provider := cmd.Config.String("identity-provider")
	if provider != "" {
		input.IdentityProvider = enums.ToSSOProvider(provider)
	}

	clientID := cmd.Config.String("client-id")
	if clientID != "" {
		input.IdentityProviderClientID = &clientID
	}

	clientSecret := cmd.Config.String("client-secret")
	if clientSecret != "" {
		input.IdentityProviderClientSecret = &clientSecret
	}

	metadataURL := cmd.Config.String("metadata-url")
	if metadataURL != "" {
		input.IdentityProviderMetadataEndpoint = &metadataURL
	}

	discoveryURL := cmd.Config.String("discovery-url")
	if discoveryURL != "" {
		input.OidcDiscoveryEndpoint = &discoveryURL
	}

	enforce := cmd.Config.Bool("enforce-sso")
	if enforce {
		input.IdentityProviderLoginEnforced = &enforce
	}

	allowAutoJoin := cmd.Config.Bool("allow-matching-domains-autojoin")
	if allowAutoJoin {
		input.AllowMatchingDomainsAutojoin = &allowAutoJoin
	}

	setAllowedEmailDomains := cmd.Config.Strings("set-allowed-email-domains")
	if len(setAllowedEmailDomains) > 0 {
		input.AllowedEmailDomains = setAllowedEmailDomains
	}

	return id, input, nil
}

// update an organization setting in the platform
func update(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	if id == "" {
		settings, err := client.GetAllOrganizationSettings(ctx)
		cobra.CheckErr(err)

		if len(settings.OrganizationSettings.Edges) == 1 {
			id = settings.OrganizationSettings.Edges[0].Node.ID
		} else {
			return cmd.NewRequiredFieldMissingError("id")
		}
	}

	o, err := client.UpdateOrganizationSetting(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
