//go:build cli

package trustcenterdoc

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new trust center document",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("trust-center-id", "t", "", "trust center id for the document")
	createCmd.Flags().StringP("title", "n", "", "title of the document")
	createCmd.Flags().StringP("visibility", "v", "NOT_VISIBLE", "visibility of the document (NOT_VISIBLE, PROTECTED, PUBLICLY_VISIBLE)")
	createCmd.Flags().StringSliceP("tags", "g", []string{}, "tags associated with the document")
	createCmd.Flags().StringP("file", "f", "", "file to upload as the trust center document (required)")
	createCmd.Flags().BoolP("watermarking-enabled", "w", false, "whether watermarking is enabled for the document")
}

// createValidation validates the required fields for the command
func createValidation() (graphclient.CreateTrustCenterDocInput, *graphql.Upload, error) {
	input := graphclient.CreateTrustCenterDocInput{}

	title := cmd.Config.String("title")
	if title == "" {
		return input, nil, cmd.NewRequiredFieldMissingError("title")
	}

	filePath := cmd.Config.String("file")
	if filePath == "" {
		return input, nil, cmd.NewRequiredFieldMissingError("file")
	}

	input.Title = title

	trustCenterID := cmd.Config.String("trust-center-id")
	if trustCenterID != "" {
		input.TrustCenterID = &trustCenterID
	}

	visibility := cmd.Config.String("visibility")
	if visibility != "" {
		visibilityEnum := enums.ToTrustCenterDocumentVisibility(visibility)
		if visibilityEnum != nil {
			input.Visibility = visibilityEnum
		}
	}

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	watermarkingEnabled := cmd.Config.Bool("watermarking-enabled")
	if watermarkingEnabled {
		input.WatermarkingEnabled = &watermarkingEnabled
	}

	// Handle file upload
	u, err := storage.NewUploadFile(filePath)
	if err != nil {
		return input, nil, err
	}

	fileUpload := &graphql.Upload{
		File:        u.RawFile,
		Filename:    u.OriginalName,
		Size:        u.Size,
		ContentType: u.ContentType,
	}

	return input, fileUpload, nil
}

// create a new trust center document
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

	o, err := client.CreateTrustCenterDoc(ctx, input, *fileUpload)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
