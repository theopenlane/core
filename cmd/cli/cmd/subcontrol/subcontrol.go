//go:build cli

package subcontrol

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func fetchSubcontrols(ctx context.Context, client *openlaneclient.OpenlaneClient) (any, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	id := cmd.Config.String("id")
	if id != "" {
		return client.GetSubcontrolByID(ctx, id)
	}

	return client.GetAllSubcontrols(ctx)
}

func buildSubcontrolCreate() (openlaneclient.CreateSubcontrolInput, error) {
	var input openlaneclient.CreateSubcontrolInput

	input.RefCode = cmd.Config.String("ref-code")
	if input.RefCode == "" {
		return input, cmd.NewRequiredFieldMissingError("ref-code")
	}

	input.ControlID = cmd.Config.String("control")
	if input.ControlID == "" {
		return input, cmd.NewRequiredFieldMissingError("control")
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
	} else {
		input.Status = &enums.ControlStatusNotImplemented
	}

	if controlType := cmd.Config.String("control-type"); controlType != "" {
		input.ControlType = enums.ToControlType(controlType)
	}

	if mapped := cmd.Config.Strings("mapped-categories"); len(mapped) > 0 {
		input.MappedCategories = mapped
	}

	return input, nil
}

func createSubcontrol(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.CreateSubcontrol, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	input, err := buildSubcontrolCreate()
	if err != nil {
		return nil, err
	}

	return client.CreateSubcontrol(ctx, input)
}

func buildSubcontrolUpdate() (string, openlaneclient.UpdateSubcontrolInput, error) {
	id := cmd.Config.String("id")
	if id == "" {
		return "", openlaneclient.UpdateSubcontrolInput{}, cmd.NewRequiredFieldMissingError("subcontrol id")
	}

	var input openlaneclient.UpdateSubcontrolInput

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

	return id, input, nil
}

func updateSubcontrol(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.UpdateSubcontrol, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	id, input, err := buildSubcontrolUpdate()
	if err != nil {
		return nil, err
	}

	return client.UpdateSubcontrol(ctx, id, input)
}

func buildSubcontrolDelete() (string, error) {
	id := cmd.Config.String("id")
	if id == "" {
		return "", cmd.NewRequiredFieldMissingError("subcontrol id")
	}

	return id, nil
}

func deleteSubcontrol(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.DeleteSubcontrol, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	id, err := buildSubcontrolDelete()
	if err != nil {
		return nil, err
	}

	return client.DeleteSubcontrol(ctx, id)
}
