package graphapi

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/gqlgen-plugins/graphutils"
)

// getDiscussionID extracts the discussion ID from the GraphQL context
// this is used for adding comments to discussions in various update operations
func getDiscussionIDFromUpdate(ctx context.Context) *string {
	parent := getParentObjectTypeFromContext(ctx)
	parentType := strings.ToLower(parent)

	// if this is a discussion related operation, we can get the id directly
	if strings.Contains(parentType, "discussion") {
		return graphutils.GetStringInputVariableByName(ctx, "id")
	}

	// else get it from the input, which should contain updateDiscussion
	inputKey := graphutils.GetInputFieldVariableName(ctx)
	dataInput := graphutils.GetMapInputVariableByName(ctx, inputKey)

	if dataInput == nil {
		return nil
	}

	d := *dataInput
	input, ok := d["updateDiscussion"].(map[string]any)
	if !ok {
		return nil
	}

	idVal, ok := input["id"].(string)
	if !ok || idVal == "" {
		return nil
	}

	return &idVal
}

// getParentObjectTypeFromContext retrieves the parent object type from the GraphQL context
// this is the name of the root resolver being executed
func getParentObjectTypeFromContext(ctx context.Context) string {
	rootFieldCtx := graphql.GetRootFieldContext(ctx)

	return rootFieldCtx.Object
}

// setParentObjectIDInInput sets the parent object ID in the CreateNoteInput based on the context of the mutation
// and the parent object type
func setParentObjectIDInInput(ctx context.Context, dataInput *generated.CreateNoteInput) error {
	parentID := graphutils.GetStringInputVariableByName(ctx, "id")
	parentOperation := getParentObjectTypeFromContext(ctx)

	if parentID == nil || parentOperation == "" {
		// if we don't have a parent ID or type, just return as is
		logx.FromContext(ctx).Debug().Msg("no parent ID or type found in context, skipping setting parent ID in note input")

		return nil
	}

	mapInput, err := convertToMap(ctx, *dataInput)
	if err != nil {
		return err
	}

	// stripe the "update" prefix from the operation to get the object type
	parentType := strings.TrimPrefix(parentOperation, "update")
	parentField := parentType + "ID"

	// set the parent ID in the input map
	mapInput[parentField] = *parentID

	// convert back to input struct
	if err := convertToInput(ctx, mapInput, dataInput); err != nil {
		return err
	}

	return nil
}

// convertToMap converts a generic input struct to a map[string]any
func convertToMap[T any](ctx context.Context, input T) (map[string]any, error) {
	mapInput, err := jsonx.ToMap(input)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to convert input to map to set parent ID")

		return nil, err
	}

	return mapInput, nil
}

// convertToInput converts a map[string]any to a generic input struct
func convertToInput[T any](ctx context.Context, mapInput map[string]any, output T) error {
	if err := jsonx.RoundTrip(mapInput, output); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to convert map input to struct to set parent ID")

		return err
	}

	return nil
}
