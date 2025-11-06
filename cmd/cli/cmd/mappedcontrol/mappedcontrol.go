//go:build cli

package mappedcontrol

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func fetchMappedControls(ctx context.Context, client *openlaneclient.OpenlaneClient) (any, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	id := cmd.Config.String("id")
	if id != "" {
		return client.GetMappedControlByID(ctx, id)
	}

	return client.GetAllMappedControls(ctx)
}

func buildCreateMappedControl() (openlaneclient.CreateMappedControlInput, error) {
	var input openlaneclient.CreateMappedControlInput

	if relation := strings.TrimSpace(cmd.Config.String("relation")); relation != "" {
		input.Relation = &relation
	}

	if confidence := cmd.Config.Int64("confidence"); confidence > 0 {
		input.Confidence = &confidence
	}

	if mappingType := strings.TrimSpace(cmd.Config.String("mapping-type")); mappingType != "" {
		input.MappingType = enums.ToMappingType(mappingType)
	}

	input.Source = enums.ToMappingSource(cmd.Config.String("source"))

	if from := cmd.Config.Strings("from-control-ids"); len(from) > 0 {
		input.FromControlIDs = from
	}

	if to := cmd.Config.Strings("to-control-ids"); len(to) > 0 {
		input.ToControlIDs = to
	}

	if fromSubs := cmd.Config.Strings("from-subcontrol-ids"); len(fromSubs) > 0 {
		input.FromSubcontrolIDs = fromSubs
	}

	if toSubs := cmd.Config.Strings("to-subcontrol-ids"); len(toSubs) > 0 {
		input.ToSubcontrolIDs = toSubs
	}

	return input, nil
}

func buildUpdateMappedControl() (string, openlaneclient.UpdateMappedControlInput, error) {
	id := cmd.Config.String("id")
	if id == "" {
		return "", openlaneclient.UpdateMappedControlInput{}, cmd.NewRequiredFieldMissingError("id")
	}

	var input openlaneclient.UpdateMappedControlInput

	if relation := strings.TrimSpace(cmd.Config.String("relation")); relation != "" {
		input.Relation = &relation
	}

	if confidence := cmd.Config.Int64("confidence"); confidence > 0 {
		input.Confidence = &confidence
	}

	if mappingType := strings.TrimSpace(cmd.Config.String("mapping-type")); mappingType != "" {
		input.MappingType = enums.ToMappingType(mappingType)
	}

	if source := strings.TrimSpace(cmd.Config.String("source")); source != "" {
		input.Source = enums.ToMappingSource(source)
	}

	if addFrom := cmd.Config.Strings("add-from-control-ids"); len(addFrom) > 0 {
		input.AddFromControlIDs = addFrom
	}

	if addTo := cmd.Config.Strings("add-to-control-ids"); len(addTo) > 0 {
		input.AddToControlIDs = addTo
	}

	if removeFrom := cmd.Config.Strings("remove-from-control-ids"); len(removeFrom) > 0 {
		input.RemoveFromControlIDs = removeFrom
	}

	if removeTo := cmd.Config.Strings("remove-to-control-ids"); len(removeTo) > 0 {
		input.RemoveToControlIDs = removeTo
	}

	if addFromSubs := cmd.Config.Strings("add-from-subcontrol-ids"); len(addFromSubs) > 0 {
		input.AddFromSubcontrolIDs = addFromSubs
	}

	if addToSubs := cmd.Config.Strings("add-to-subcontrol-ids"); len(addToSubs) > 0 {
		input.AddToSubcontrolIDs = addToSubs
	}

	if removeFromSubs := cmd.Config.Strings("remove-from-subcontrol-ids"); len(removeFromSubs) > 0 {
		input.RemoveFromSubcontrolIDs = removeFromSubs
	}

	if removeToSubs := cmd.Config.Strings("remove-to-subcontrol-ids"); len(removeToSubs) > 0 {
		input.RemoveToSubcontrolIDs = removeToSubs
	}

	return id, input, nil
}

func buildDeleteMappedControl() (string, error) {
	id := cmd.Config.String("id")
	if id == "" {
		return "", cmd.NewRequiredFieldMissingError("id")
	}
	return id, nil
}

func createMappedControl(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.CreateMappedControl, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	input, err := buildCreateMappedControl()
	if err != nil {
		return nil, err
	}

	return client.CreateMappedControl(ctx, input)
}

func updateMappedControl(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.UpdateMappedControl, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	id, input, err := buildUpdateMappedControl()
	if err != nil {
		return nil, err
	}

	return client.UpdateMappedControl(ctx, id, input)
}

func deleteMappedControl(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.DeleteMappedControl, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	id, err := buildDeleteMappedControl()
	if err != nil {
		return nil, err
	}

	return client.DeleteMappedControl(ctx, id)
}

func mappedControlRecords(result any) ([]map[string]any, error) {
	controls, err := extractMappedControls(result)
	if err != nil {
		return nil, err
	}

	records := make([]map[string]any, len(controls))
	for i, mc := range controls {
		to := collectRefCodes(mc.ToControls, mc.ToSubcontrols)
		from := collectRefCodes(mc.FromControls, mc.FromSubcontrols)

		confidence := "-"
		if mc.Confidence != nil {
			confidence = fmt.Sprintf("%d%%", *mc.Confidence)
		}

		relation := ""
		if mc.Relation != nil {
			relation = *mc.Relation
		}

		source := ""
		if mc.Source != nil {
			source = mc.Source.String()
		}

		records[i] = map[string]any{
			"id":          mc.ID,
			"to":          strings.Join(to, ", "),
			"from":        strings.Join(from, ", "),
			"relation":    relation,
			"confidence":  confidence,
			"mappingType": mc.MappingType.String(),
			"source":      source,
		}
	}

	return records, nil
}

func extractMappedControls(result any) ([]openlaneclient.MappedControl, error) {
	var nodes []any

	switch v := result.(type) {
	case *openlaneclient.CreateMappedControl:
		if v != nil {
			nodes = append(nodes, v.CreateMappedControl.MappedControl)
		}
	case *openlaneclient.UpdateMappedControl:
		if v != nil {
			nodes = append(nodes, v.UpdateMappedControl.MappedControl)
		}
	case *openlaneclient.GetMappedControlByID:
		if v != nil {
			nodes = append(nodes, v.MappedControl)
		}
	case *openlaneclient.GetAllMappedControls:
		for _, edge := range v.MappedControls.Edges {
			if edge != nil && edge.Node != nil {
				nodes = append(nodes, edge.Node)
			}
		}
	case *openlaneclient.GetMappedControls:
		for _, edge := range v.MappedControls.Edges {
			if edge != nil && edge.Node != nil {
				nodes = append(nodes, edge.Node)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported mapped control response type %T", result)
	}

	if len(nodes) == 0 {
		return []openlaneclient.MappedControl{}, nil
	}

	payload, err := json.Marshal(nodes)
	if err != nil {
		return nil, err
	}

	var out []openlaneclient.MappedControl
	if err := json.Unmarshal(payload, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func collectRefCodes(controls *openlaneclient.ControlConnection, subcontrols *openlaneclient.SubcontrolConnection) []string {
	refs := make([]string, 0)

	if controls != nil {
		for _, edge := range controls.Edges {
			if edge != nil && edge.Node != nil {
				refs = append(refs, edge.Node.RefCode)
			}
		}
	}

	if subcontrols != nil {
		for _, edge := range subcontrols.Edges {
			if edge != nil && edge.Node != nil {
				refs = append(refs, edge.Node.RefCode)
			}
		}
	}

	return refs
}
