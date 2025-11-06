//go:build cli

package group

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func groupCreateHook(spec *speccli.CreateSpec) speccli.CreatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		if client == nil {
			return true, speccli.OperationOutput{}, errors.New("client is required")
		}

		input, err := buildCreateGroupInput()
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		result, err := client.CreateGroup(ctx, input)
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

func buildCreateGroupInput() (openlaneclient.CreateGroupInput, error) {
	var input openlaneclient.CreateGroupInput

	input.Name = cmdpkg.Config.String("name")
	if input.Name == "" {
		return input, cmdpkg.NewRequiredFieldMissingError("name")
	}

	if displayName := cmdpkg.Config.String("display-name"); displayName != "" {
		input.DisplayName = &displayName
	}

	if description := cmdpkg.Config.String("description"); description != "" {
		input.Description = &description
	}

	if tags := cmdpkg.Config.Strings("tags"); len(tags) > 0 {
		input.Tags = tags
	}

	if cmdpkg.Config.Bool("private") {
		private := enums.VisibilityPrivate
		input.CreateGroupSettings = &openlaneclient.CreateGroupSettingInput{
			Visibility: &private,
		}
	}

	return input, nil
}
