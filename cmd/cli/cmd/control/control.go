//go:build cli

package control

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func buildControlCreateInput() (openlaneclient.CreateControlInput, error) {
	var input openlaneclient.CreateControlInput

	input.RefCode = cmd.Config.String("ref-code")
	if input.RefCode == "" {
		return input, cmd.NewRequiredFieldMissingError("ref-code")
	}

	if description := cmd.Config.String("description"); description != "" {
		input.Description = &description
	}

	if source := cmd.Config.String("source"); source != "" {
		input.Source = enums.ToControlSource(source)
	}

	if category := cmd.Config.String("category"); category != "" {
		input.Category = &category
	}

	if categoryID := cmd.Config.String("category-id"); categoryID != "" {
		input.CategoryID = &categoryID
	}

	if subcategory := cmd.Config.String("subcategory"); subcategory != "" {
		input.Subcategory = &subcategory
	}

	input.Status = enums.ToControlStatus(cmd.Config.String("status"))

	if controlType := cmd.Config.String("control-type"); controlType != "" {
		input.ControlType = enums.ToControlType(controlType)
	}

	if mapped := cmd.Config.Strings("mapped-categories"); len(mapped) > 0 {
		input.MappedCategories = mapped
	}

	if frameworkID := cmd.Config.String("framework-id"); frameworkID != "" {
		input.StandardID = &frameworkID
	}

	if programs := cmd.Config.Strings("programs"); len(programs) > 0 {
		input.ProgramIDs = programs
	}

	if editors := cmd.Config.Strings("editors"); len(editors) > 0 {
		input.EditorIDs = editors
	}

	return input, nil
}

func createControl(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.CreateControl, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	input, err := buildControlCreateInput()
	if err != nil {
		return nil, err
	}

	return client.CreateControl(ctx, input)
}

func buildControlUpdateInput() (string, openlaneclient.UpdateControlInput, error) {
	id := cmd.Config.String("id")
	if id == "" {
		return "", openlaneclient.UpdateControlInput{}, cmd.NewRequiredFieldMissingError("id")
	}

	var input openlaneclient.UpdateControlInput

	if refCode := cmd.Config.String("ref-code"); refCode != "" {
		input.RefCode = &refCode
	}

	if description := cmd.Config.String("description"); description != "" {
		input.Description = &description
	}

	if source := cmd.Config.String("source"); source != "" {
		input.Source = enums.ToControlSource(source)
	}

	if category := cmd.Config.String("category"); category != "" {
		input.Category = &category
	}

	if categoryID := cmd.Config.String("category-id"); categoryID != "" {
		input.CategoryID = &categoryID
	}

	if subcategory := cmd.Config.String("subcategory"); subcategory != "" {
		input.Subcategory = &subcategory
	}

	if status := cmd.Config.String("status"); status != "" {
		input.Status = enums.ToControlStatus(status)
	}

	if controlType := cmd.Config.String("control-type"); controlType != "" {
		input.ControlType = enums.ToControlType(controlType)
	}

	if mapped := cmd.Config.Strings("mapped-categories"); len(mapped) > 0 {
		input.MappedCategories = mapped
	}

	if frameworkID := cmd.Config.String("framework-id"); frameworkID != "" {
		input.StandardID = &frameworkID
	}

	if addPrograms := cmd.Config.Strings("add-programs"); len(addPrograms) > 0 {
		input.AddProgramIDs = addPrograms
	}

	if removePrograms := cmd.Config.Strings("remove-programs"); len(removePrograms) > 0 {
		input.RemoveProgramIDs = removePrograms
	}

	if addEditors := cmd.Config.Strings("add-editors"); len(addEditors) > 0 {
		input.AddEditorIDs = addEditors
	}

	return id, input, nil
}

func updateControl(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.UpdateControl, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	id, input, err := buildControlUpdateInput()
	if err != nil {
		return nil, err
	}

	return client.UpdateControl(ctx, id, input)
}

func getControl(ctx context.Context, client *openlaneclient.OpenlaneClient) (any, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	id := cmd.Config.String("id")
	refCode := cmd.Config.String("ref-code")

	if id != "" {
		return client.GetControlByID(ctx, id)
	}

	if refCode != "" {
		filter := &openlaneclient.ControlWhereInput{RefCode: &refCode}
		return client.GetControls(ctx, cmd.First, cmd.Last, filter)
	}

	return client.GetAllControls(ctx)
}
