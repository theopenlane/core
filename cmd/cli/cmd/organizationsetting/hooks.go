//go:build cli

package orgsetting

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func updateOrganizationSettingHook(spec *speccli.UpdateSpec) speccli.UpdatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		id := strings.TrimSpace(cmdpkg.Config.String("id"))
		if id == "" {
			fetchedID, err := fetchSingleOrganizationSettingID(ctx, client)
			if err != nil {
				return true, speccli.OperationOutput{}, err
			}
			id = fetchedID
		}

		input, err := buildUpdateOrganizationSettingInput()
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		result, err := client.UpdateOrganizationSetting(ctx, id, input)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		out, err := speccli.WrapSingleResult(result, spec.ResultPath)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		return true, out, nil
	}
}

func fetchSingleOrganizationSettingID(ctx context.Context, client *openlaneclient.OpenlaneClient) (string, error) {
	settings, err := client.GetAllOrganizationSettings(ctx)
	if err != nil {
		return "", err
	}

	edges := settings.OrganizationSettings.Edges
	if len(edges) != 1 {
		return "", cmdpkg.NewRequiredFieldMissingError("id")
	}

	return edges[0].Node.ID, nil
}

func buildUpdateOrganizationSettingInput() (openlaneclient.UpdateOrganizationSettingInput, error) {
	var input openlaneclient.UpdateOrganizationSettingInput

	if billingContact := strings.TrimSpace(cmdpkg.Config.String("billing-contact")); billingContact != "" {
		input.BillingContact = &billingContact
	}

	if billingEmail := strings.TrimSpace(cmdpkg.Config.String("billing-email")); billingEmail != "" {
		input.BillingEmail = &billingEmail
	}

	if billingPhone := strings.TrimSpace(cmdpkg.Config.String("billing-phone")); billingPhone != "" {
		input.BillingPhone = &billingPhone
	}

	if taxID := strings.TrimSpace(cmdpkg.Config.String("tax-identifier")); taxID != "" {
		input.TaxIdentifier = &taxID
	}

	if tags := cmdpkg.Config.Strings("tags"); len(tags) > 0 {
		input.Tags = tags
	}

	if domains := cmdpkg.Config.Strings("domains"); len(domains) > 0 {
		input.Domains = domains
	}

	allowedAppend := cmdpkg.Config.Strings("append-allowed-domains")
	if len(allowedAppend) > 0 {
		input.AllowedEmailDomains = allowedAppend
	}

	if clearAllowed := cmdpkg.Config.Bool("clear-allowed-domains"); clearAllowed {
		input.ClearAllowedEmailDomains = &clearAllowed
	}

	if provider := strings.TrimSpace(cmdpkg.Config.String("identity-provider")); provider != "" {
		parsed := enums.ToSSOProvider(provider)
		if parsed != nil && parsed.String() != enums.SSOProviderInvalid.String() {
			input.IdentityProvider = parsed
		}
	}

	if clientID := strings.TrimSpace(cmdpkg.Config.String("client-id")); clientID != "" {
		input.IdentityProviderClientID = &clientID
	}

	if clientSecret := strings.TrimSpace(cmdpkg.Config.String("client-secret")); clientSecret != "" {
		input.IdentityProviderClientSecret = &clientSecret
	}

	if metadataURL := strings.TrimSpace(cmdpkg.Config.String("metadata-url")); metadataURL != "" {
		input.IdentityProviderMetadataEndpoint = &metadataURL
	}

	if discoveryURL := strings.TrimSpace(cmdpkg.Config.String("discovery-url")); discoveryURL != "" {
		input.OidcDiscoveryEndpoint = &discoveryURL
	}

	if enforce := cmdpkg.Config.Bool("enforce-sso"); enforce {
		input.IdentityProviderLoginEnforced = &enforce
	}

	if allowAuto := cmdpkg.Config.Bool("allow-matching-domains-autojoin"); allowAuto {
		input.AllowMatchingDomainsAutojoin = &allowAuto
	}

	if setAllowed := cmdpkg.Config.Strings("set-allowed-email-domains"); len(setAllowed) > 0 {
		input.AllowedEmailDomains = setAllowed
		input.ClearAllowedEmailDomains = nil
	}

	return input, nil
}
