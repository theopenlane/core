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

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing trust center watermark config",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "trust center watermark config id to update")
	updateCmd.Flags().StringP("text", "t", "", "watermark text")
	updateCmd.Flags().StringP("logo-file", "f", "", "logo file to upload")
	updateCmd.Flags().Float64P("font-size", "s", 48.0, "font size of the watermark text")
	updateCmd.Flags().Float64P("opacity", "o", 0.3, "opacity of the watermark text")
	updateCmd.Flags().Float64P("rotation", "r", 45.0, "rotation of the watermark text")
	updateCmd.Flags().StringP("color", "c", "", "color of the watermark text")
	updateCmd.Flags().StringP("font", "n", "", "font of the watermark text")
}

// updateValidation validates the required fields for the command
func updateValidation() (string, *openlaneclient.UpdateTrustCenterWatermarkConfigInput, *graphql.Upload, error) {
	id := cmd.Config.String("id")
	if id == "" {
		return "", nil, nil, cmd.NewRequiredFieldMissingError("id")
	}

	input := &openlaneclient.UpdateTrustCenterWatermarkConfigInput{}

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
			return id, input, nil, err
		}
		fileUpload = &graphql.Upload{
			File:        upload.RawFile,
			Filename:    upload.OriginalName,
			Size:        upload.Size,
			ContentType: upload.ContentType,
		}
	}

	return id, input, fileUpload, nil
}

// update an existing trust center watermark config in the platform
func update(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, fileUpload, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateTrustCenterWatermarkConfig(ctx, id, *input, fileUpload)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
