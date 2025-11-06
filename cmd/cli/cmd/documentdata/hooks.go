//go:build cli

package documentdata

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/spf13/cobra"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func documentDataCreateHook(spec *speccli.CreateSpec) speccli.CreatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		if client == nil {
			return true, speccli.OperationOutput{}, errors.New("client is required")
		}

		input, err := buildCreateDocumentDataInput()
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		result, err := client.CreateDocumentData(ctx, input)
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

func buildCreateDocumentDataInput() (openlaneclient.CreateDocumentDataInput, error) {
	var input openlaneclient.CreateDocumentDataInput

	input.TemplateID = cmdpkg.Config.String("template-id")
	if input.TemplateID == "" {
		return input, cmdpkg.NewRequiredFieldMissingError("template id")
	}

	dataPath := cmdpkg.Config.String("data")
	if dataPath == "" {
		return input, cmdpkg.NewRequiredFieldMissingError("data")
	}

	payload, err := os.ReadFile(dataPath)
	if err != nil {
		return input, err
	}

	var jsonData map[string]any
	if err := json.Unmarshal(payload, &jsonData); err != nil {
		return input, err
	}

	input.Data = jsonData

	return input, nil
}
