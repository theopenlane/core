//go:build cli

package trustcenternda

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func newSubmitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit",
		Short: "submit a trust center NDA response",
		RunE: func(c *cobra.Command, _ []string) error {
			out, err := submitTrustCenterNDA(c.Context())
			if err != nil {
				return err
			}

			return speccli.PrintJSON(out)
		},
	}

	cmd.Flags().StringP("template-id", "t", "", "template id for the NDA")
	cmd.Flags().StringP("response", "r", "", "path to JSON response payload")

	return cmd
}

func newSendEmailCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-email",
		Short: "send an NDA invitation email",
		RunE: func(c *cobra.Command, _ []string) error {
			out, err := sendTrustCenterNDAEmail(c.Context())
			if err != nil {
				return err
			}

			return speccli.PrintJSON(out)
		},
	}

	cmd.Flags().StringP("trust-center-id", "t", "", "trust center id for the NDA")
	cmd.Flags().StringP("email", "e", "", "email address to send the NDA to")

	return cmd
}

func trustCenterNdaCreateHook(spec *speccli.CreateSpec) speccli.CreatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		if client == nil {
			return true, speccli.OperationOutput{}, fmt.Errorf("client is required")
		}

		result, err := createTrustCenterNDA(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		out, wrapErr := speccli.WrapSingleResult(result, spec.ResultPath)
		if wrapErr != nil {
			return true, speccli.OperationOutput{}, wrapErr
		}

		return true, out, nil
	}
}

func trustCenterNdaUpdateHook(spec *speccli.UpdateSpec) speccli.UpdatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		if client == nil {
			return true, speccli.OperationOutput{}, fmt.Errorf("client is required")
		}

		result, err := updateTrustCenterNDA(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		out, wrapErr := speccli.WrapSingleResult(result, spec.ResultPath)
		if wrapErr != nil {
			return true, speccli.OperationOutput{}, wrapErr
		}

		return true, out, nil
	}
}

func createTrustCenterNDA(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.CreateTrustCenterNda, error) {
	input, upload, err := buildCreateInput()
	if err != nil {
		return nil, err
	}

	uploads := []*graphql.Upload{upload}
	return client.CreateTrustCenterNda(ctx, input, uploads)
}

func updateTrustCenterNDA(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.UpdateTrustCenterNda, error) {
	id := strings.TrimSpace(cmdpkg.Config.String("id"))
	if id == "" {
		return nil, cmdpkg.NewRequiredFieldMissingError("id")
	}

	path := strings.TrimSpace(cmdpkg.Config.String("nda-file"))
	var uploads []*graphql.Upload
	if path != "" {
		upload, err := speccli.UploadFromPath(path)
		if err != nil {
			return nil, err
		}
		if upload != nil {
			uploads = []*graphql.Upload{upload}
		}
	}

	return client.UpdateTrustCenterNda(ctx, id, uploads)
}

func submitTrustCenterNDA(ctx context.Context) (*openlaneclient.SubmitTrustCenterNDAResponse, error) {
	input, err := buildSubmitInput()
	if err != nil {
		return nil, err
	}

	client, cleanup, err := acquireClient(ctx)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	return client.SubmitTrustCenterNDAResponse(ctx, input)
}

func sendTrustCenterNDAEmail(ctx context.Context) (*openlaneclient.SendTrustCenterNDAEmail, error) {
	input, err := buildSendEmailInput()
	if err != nil {
		return nil, err
	}

	client, cleanup, err := acquireClient(ctx)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	return client.SendTrustCenterNDAEmail(ctx, input)
}

func buildCreateInput() (openlaneclient.CreateTrustCenterNDAInput, *graphql.Upload, error) {
	var input openlaneclient.CreateTrustCenterNDAInput

	trustCenterID := strings.TrimSpace(cmdpkg.Config.String("trust-center-id"))
	if trustCenterID == "" {
		return input, nil, cmdpkg.NewRequiredFieldMissingError("trust center id")
	}
	input.TrustCenterID = trustCenterID

	path := strings.TrimSpace(cmdpkg.Config.String("nda-file"))
	if path == "" {
		return input, nil, cmdpkg.NewRequiredFieldMissingError("nda file")
	}

	upload, err := speccli.UploadFromPath(path)
	if err != nil {
		return input, nil, err
	}

	return input, upload, nil
}

func buildSubmitInput() (openlaneclient.SubmitTrustCenterNDAResponseInput, error) {
	var input openlaneclient.SubmitTrustCenterNDAResponseInput

	templateID := strings.TrimSpace(cmdpkg.Config.String("template-id"))
	if templateID == "" {
		return input, cmdpkg.NewRequiredFieldMissingError("template id")
	}
	input.TemplateID = templateID

	responsePath := strings.TrimSpace(cmdpkg.Config.String("response"))
	if responsePath == "" {
		return input, cmdpkg.NewRequiredFieldMissingError("response")
	}

	data, err := os.ReadFile(responsePath)
	if err != nil {
		return input, err
	}

	var response map[string]any
	if err := json.Unmarshal(data, &response); err != nil {
		return input, err
	}

	input.Response = response

	return input, nil
}

func buildSendEmailInput() (openlaneclient.SendTrustCenterNDAInput, error) {
	var input openlaneclient.SendTrustCenterNDAInput

	trustCenterID := strings.TrimSpace(cmdpkg.Config.String("trust-center-id"))
	if trustCenterID == "" {
		return input, cmdpkg.NewRequiredFieldMissingError("trust center id")
	}
	input.TrustCenterID = trustCenterID

	email := strings.TrimSpace(cmdpkg.Config.String("email"))
	if email == "" {
		return input, cmdpkg.NewRequiredFieldMissingError("email")
	}
	input.Email = email

	return input, nil
}

func acquireClient(ctx context.Context) (*openlaneclient.OpenlaneClient, func(), error) {
	client, err := cmdpkg.TokenAuth(ctx, cmdpkg.Config)
	if err != nil || client == nil {
		client, err = cmdpkg.SetupClientWithAuth(ctx)
		if err != nil {
			return nil, func() {}, err
		}
		return client, func() { cmdpkg.StoreSessionCookies(client) }, nil
	}

	return client, func() {}, nil
}
