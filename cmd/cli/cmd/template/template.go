//go:build cli

package templates

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func buildCreateTemplateInput() (openlaneclient.CreateTemplateInput, []*graphql.Upload, error) {
	var input openlaneclient.CreateTemplateInput

	input.Name = strings.TrimSpace(cmd.Config.String("name"))
	if input.Name == "" {
		return input, nil, cmd.NewRequiredFieldMissingError("name")
	}

	jsonConfig := strings.TrimSpace(cmd.Config.String("json-config"))
	if jsonConfig == "" {
		return input, nil, cmd.NewRequiredFieldMissingError("json-config")
	}

	configPayload, err := readJSONFile(jsonConfig)
	if err != nil {
		return input, nil, err
	}
	input.Jsonconfig = configPayload

	if description := strings.TrimSpace(cmd.Config.String("description")); description != "" {
		input.Description = &description
	}

	if schemaPath := strings.TrimSpace(cmd.Config.String("ui-schema")); schemaPath != "" {
		uiSchema, err := readJSONFile(schemaPath)
		if err != nil {
			return input, nil, err
		}
		input.Uischema = uiSchema
	}

	if templateType := strings.TrimSpace(cmd.Config.String("type")); templateType != "" {
		input.TemplateType = enums.ToDocumentType(templateType)
	}

	if templateKind := strings.TrimSpace(cmd.Config.String("kind")); templateKind != "" {
		input.Kind = enums.ToTemplateKind(templateKind)
	}

	if trustCenter := strings.TrimSpace(cmd.Config.String("trust-center-id")); trustCenter != "" {
		input.TrustCenterID = &trustCenter
	}

	return input, []*graphql.Upload{}, nil
}

func createTemplate(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.CreateTemplate, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	input, uploads, err := buildCreateTemplateInput()
	if err != nil {
		return nil, err
	}

	return client.CreateTemplate(ctx, input, uploads)
}

func getTemplates(ctx context.Context, client *openlaneclient.OpenlaneClient) (any, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	id := strings.TrimSpace(cmd.Config.String("id"))
	if id != "" {
		return client.GetTemplateByID(ctx, id)
	}

	return client.GetAllTemplates(ctx)
}

func readJSONFile(path string) (map[string]any, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, err
	}

	return data, nil
}
