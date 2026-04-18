package emailbranding

import (
	"context"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/integrations/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new email branding configuration",
	RunE: func(c *cobra.Command, _ []string) error {
		return create(c.Context())
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "friendly name for this branding (required)")
	createCmd.Flags().String("brand-name", "", "brand name displayed in templates")
	createCmd.Flags().String("logo-url", "", "URL of the brand logo for emails")
	createCmd.Flags().String("primary-color", "", "primary brand color (hex, e.g. #1F2937)")
	createCmd.Flags().String("secondary-color", "", "secondary brand color")
	createCmd.Flags().String("background-color", "", "background color for emails")
	createCmd.Flags().String("text-color", "", "text color for emails")
	createCmd.Flags().String("button-color", "", "button background color")
	createCmd.Flags().String("button-text-color", "", "button text color")
	createCmd.Flags().String("link-color", "", "link color for emails")
	createCmd.Flags().String("font-family", "", "font family (enum value)")
	createCmd.Flags().Bool("is-default", false, "mark as the organization default branding")
	createCmd.Flags().StringSlice("tags", nil, "tags to attach to the branding record")
}

// buildCreateInput builds the CreateEmailBrandingInput from the loaded config
func buildCreateInput() (graphclient.CreateEmailBrandingInput, error) {
	var input graphclient.CreateEmailBrandingInput

	name := cmd.Config.String("name")
	if name == "" {
		return input, ErrNameRequired
	}

	input.Name = name
	input.BrandName = lo.EmptyableToPtr(cmd.Config.String("brand-name"))
	input.LogoRemoteURL = lo.EmptyableToPtr(cmd.Config.String("logo-url"))
	input.PrimaryColor = lo.EmptyableToPtr(cmd.Config.String("primary-color"))
	input.SecondaryColor = lo.EmptyableToPtr(cmd.Config.String("secondary-color"))
	input.BackgroundColor = lo.EmptyableToPtr(cmd.Config.String("background-color"))
	input.TextColor = lo.EmptyableToPtr(cmd.Config.String("text-color"))
	input.ButtonColor = lo.EmptyableToPtr(cmd.Config.String("button-color"))
	input.ButtonTextColor = lo.EmptyableToPtr(cmd.Config.String("button-text-color"))
	input.LinkColor = lo.EmptyableToPtr(cmd.Config.String("link-color"))

	if family := cmd.Config.String("font-family"); family != "" {
		f := enums.Font(family)
		input.FontFamily = &f
	}

	isDefault := cmd.Config.Bool("is-default")
	input.IsDefault = &isDefault

	if tags := cmd.Config.Strings("tags"); len(tags) > 0 {
		input.Tags = tags
	}

	return input, nil
}

// create executes the CreateEmailBranding mutation
func create(ctx context.Context) error {
	input, err := buildCreateInput()
	if err != nil {
		return err
	}

	client, err := cmd.ConnectClient(ctx)
	if err != nil {
		return err
	}

	resp, err := client.CreateEmailBranding(ctx, input)
	if err != nil {
		return err
	}

	b := resp.CreateEmailBranding.EmailBranding

	headers := []string{"ID", "Name", "BrandName", "PrimaryColor", "IsDefault"}
	rows := [][]string{{
		b.ID,
		b.Name,
		cmd.StrPtr(b.BrandName),
		cmd.StrPtr(b.PrimaryColor),
		cmd.BoolPtrStr(b.IsDefault),
	}}

	return cmd.RenderTable(resp, headers, rows)
}
