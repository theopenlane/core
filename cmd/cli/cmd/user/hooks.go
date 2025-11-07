//go:build cli

package user

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func updateUserHook(spec *speccli.UpdateSpec) speccli.UpdatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		id := strings.TrimSpace(cmdpkg.Config.String("id"))
		if id == "" {
			return true, speccli.OperationOutput{}, cmdpkg.NewRequiredFieldMissingError("user id")
		}

		input, avatarFile, err := buildUpdateUserInput()
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		result, err := client.UpdateUser(ctx, id, input, avatarFile)
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

func getUserHook(spec *speccli.GetSpec) speccli.GetPreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		id := strings.TrimSpace(cmdpkg.Config.String(spec.IDFlag.Name))
		if id != "" {
			result, err := client.GetUserByID(ctx, id)
			if err != nil {
				return true, speccli.OperationOutput{}, err
			}

			out, err := speccli.WrapSingleResult(result, spec.ResultPath)
			if err != nil {
				return true, speccli.OperationOutput{}, err
			}

			return true, out, nil
		}

		if cmdpkg.Config.Bool("self") {
			result, err := client.GetSelf(ctx)
			if err != nil {
				return true, speccli.OperationOutput{}, err
			}

			out, err := speccli.WrapSingleResult(result, []string{"self"})
			if err != nil {
				return true, speccli.OperationOutput{}, err
			}

			return true, out, nil
		}

		result, err := client.GetAllUsers(ctx)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		out, err := speccli.WrapListResult(result, spec.ListRoot)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		return true, out, nil
	}
}

func buildUpdateUserInput() (openlaneclient.UpdateUserInput, *graphql.Upload, error) {
	var input openlaneclient.UpdateUserInput

	if firstName := strings.TrimSpace(cmdpkg.Config.String("first-name")); firstName != "" {
		input.FirstName = &firstName
	}

	if lastName := strings.TrimSpace(cmdpkg.Config.String("last-name")); lastName != "" {
		input.LastName = &lastName
	}

	if displayName := strings.TrimSpace(cmdpkg.Config.String("display-name")); displayName != "" {
		input.DisplayName = &displayName
	}

	if email := strings.TrimSpace(cmdpkg.Config.String("email")); email != "" {
		input.Email = &email
	}

	avatarPath := strings.TrimSpace(cmdpkg.Config.String("avatar-file"))
	if avatarPath == "" {
		return input, nil, nil
	}

	upload, err := speccli.UploadFromPath(avatarPath)
	if err != nil {
		return input, nil, err
	}

	return input, upload, nil
}
