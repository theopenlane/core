//go:build cli

package trustcenterwatermarkconfig

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects/storage"
	openlaneclient "github.com/theopenlane/go-client"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new trust center watermark config",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("trust-center-id", "i", "", "trust center id for the watermark config")
	createCmd.Flags().StringP("text", "t", "", "watermark text")
	createCmd.Flags().StringP("logo-file", "f", "", "logo file to upload")
	createCmd.Flags().Float64P("font-size", "s", 48.0, "font size of the watermark text")
	createCmd.Flags().Float64P("opacity", "o", 0.3, "opacity of the watermark text")
	createCmd.Flags().Float64P("rotation", "r", 45.0, "rotation of the watermark text")
	createCmd.Flags().StringP("color", "c", "", "color of the watermark text")
	createCmd.Flags().StringP("font", "n", "", "font of the watermark text")
}

// createValidation validates the required fields for the command
func createValidation() (*openlaneclient.CreateTrustCenterWatermarkConfigInput, *graphql.Upload, error) {
	input := &openlaneclient.CreateTrustCenterWatermarkConfigInput{}
	trustCenterID := cmd.Config.String("trust-center-id")
	if trustCenterID == "" {
		return nil, nil, cmd.NewRequiredFieldMissingError("trust center id")
	}
	input.TrustCenterID = &trustCenterID

	text := cmd.Config.String("text")
	if text != "" {
		input.Text = &text
	}

	fontSize := cmd.Config.Float64("font-size")
	if fontSize != 0 {
		input.FontSize = &fontSize
	}

	opacity := cmd.Config.Float64("opacity")
	if opacity != 0 {
		input.Opacity = &opacity
	}

	rotation := cmd.Config.Float64("rotation")
	if rotation != 0 {
		input.Rotation = &rotation
	}

	color := cmd.Config.String("color")
	if color != "" {
		input.Color = &color
	}

	font := cmd.Config.String("font")
	if font != "" {
		input.Font = enums.ToFont(font)
	}

	var fileUpload *graphql.Upload
	logoFile := cmd.Config.String("logo-file")
	if logoFile != "" {
		upload, err := storage.NewUploadFile(logoFile)
		if err != nil {
			return nil, nil, err
		}
		fileUpload = &graphql.Upload{
			File:        upload.RawFile,
			Filename:    upload.OriginalName,
			Size:        upload.Size,
			ContentType: upload.ContentType,
		}
	}

	return input, fileUpload, nil
}

// create a new trust center watermark config
func create(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, fileUpload, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateTrustCenterWatermarkConfig(ctx, *input, fileUpload)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
