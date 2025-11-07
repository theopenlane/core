//go:build cli

package organization

import (
	"context"
	"fmt"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/iam/tokens"
)

func createOrganizationHook(spec *speccli.CreateSpec) speccli.CreatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		input, avatarFile, err := buildCreateOrganizationInput()
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		result, err := client.CreateOrganization(ctx, input, avatarFile)
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

func updateOrganizationHook(spec *speccli.UpdateSpec) speccli.UpdatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		id := strings.TrimSpace(cmdpkg.Config.String("id"))
		if id == "" {
			return true, speccli.OperationOutput{}, cmdpkg.NewRequiredFieldMissingError("organization id")
		}

		input, avatarFile, err := buildUpdateOrganizationInput()
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		result, err := client.UpdateOrganization(ctx, id, input, avatarFile)
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

func getOrganizationHook(spec *speccli.GetSpec) speccli.GetPreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		if cmdpkg.Config.Bool("current-only") {
			token, err := client.Config().Credentials.AccessToken()
			if err != nil {
				return true, speccli.OperationOutput{}, err
			}

			claims, err := tokens.ParseUnverifiedTokenClaims(token)
			if err != nil {
				return true, speccli.OperationOutput{}, err
			}

			if strings.TrimSpace(claims.OrgID) == "" {
				log.Error().Msg("no organization ID found in the token claims, cannot get current organization")
				return true, speccli.OperationOutput{}, fmt.Errorf("current token is not scoped to an organization")
			}

			result, err := client.GetOrganizationByID(ctx, claims.OrgID)
			if err != nil {
				return true, speccli.OperationOutput{}, err
			}

			out, err := speccli.WrapSingleResult(result, spec.ResultPath)
			if err != nil {
				return true, speccli.OperationOutput{}, err
			}

			return true, out, nil
		}

		id := strings.TrimSpace(cmdpkg.Config.String(spec.IDFlag.Name))
		if id != "" {
			result, err := client.GetOrganizationByID(ctx, id)
			if err != nil {
				return true, speccli.OperationOutput{}, err
			}

			out, err := speccli.WrapSingleResult(result, spec.ResultPath)
			if err != nil {
				return true, speccli.OperationOutput{}, err
			}

			return true, out, nil
		}

		includePersonal := cmdpkg.Config.Bool("include-personal-orgs")

		var (
			result any
			err    error
		)

		if includePersonal {
			result, err = client.GetAllOrganizations(ctx)
		} else {
			personal := false
			where := &openlaneclient.OrganizationWhereInput{PersonalOrg: &personal}
			result, err = client.GetOrganizations(ctx, cmdpkg.First, cmdpkg.Last, where)
		}

		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		root := spec.ListRoot
		if root == "" {
			root = "organizations"
		}

		out, err := speccli.WrapListResult(result, root)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		return true, out, nil
	}
}

func buildCreateOrganizationInput() (openlaneclient.CreateOrganizationInput, *graphql.Upload, error) {
	var input openlaneclient.CreateOrganizationInput

	name := strings.TrimSpace(cmdpkg.Config.String("name"))
	if name == "" {
		return input, nil, cmdpkg.NewRequiredFieldMissingError("organization name")
	}
	input.Name = name

	if displayName := strings.TrimSpace(cmdpkg.Config.String("display-name")); displayName != "" {
		input.DisplayName = &displayName
	}

	if description := strings.TrimSpace(cmdpkg.Config.String("description")); description != "" {
		input.Description = &description
	}

	if parent := strings.TrimSpace(cmdpkg.Config.String("parent-org-id")); parent != "" {
		input.ParentID = &parent
	}

	if tags := cmdpkg.Config.Strings("tags"); len(tags) > 0 {
		input.Tags = tags
	}

	if dedicatedDB := cmdpkg.Config.Bool("dedicated-db"); dedicatedDB {
		input.DedicatedDb = &dedicatedDB
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

func buildUpdateOrganizationInput() (openlaneclient.UpdateOrganizationInput, *graphql.Upload, error) {
	var input openlaneclient.UpdateOrganizationInput

	if name := strings.TrimSpace(cmdpkg.Config.String("name")); name != "" {
		input.Name = &name
	}

	if displayName := strings.TrimSpace(cmdpkg.Config.String("display-name")); displayName != "" {
		input.DisplayName = &displayName
	}

	if description := strings.TrimSpace(cmdpkg.Config.String("description")); description != "" {
		input.Description = &description
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
