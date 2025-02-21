package graphutils

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
)

// CheckForRequestedField checks if the requested field is in the list of fields from the request
func CheckForRequestedField(ctx context.Context, fieldName string) bool {
	fields := GetPreloads(ctx)
	if fields == nil {
		return false
	}

	for _, f := range fields {
		// fields are in the format of "parent.parent.fieldName", e.g. "organization.orgSubscription.subscriptionURL"
		// so we check if the fieldName is in the string and not just equal to it
		if strings.Contains(f, fieldName) {
			return true
		}
	}

	return false
}

// GetPreloads returns the preloads for the current graphql operation
func GetPreloads(ctx context.Context) []string {
	// skip if the context is not a graphql operation context
	if ok := graphql.HasOperationContext(ctx); !ok {
		return nil
	}

	gCtx := graphql.GetOperationContext(ctx)
	if gCtx == nil {
		return nil
	}

	return getNestedPreloads(
		gCtx,
		graphql.CollectFieldsCtx(ctx, nil),
		"",
	)
}

// GetStringInputVariableByName returns the input variable by name for string variables (e.g. id)
func GetStringInputVariableByName(ctx context.Context, fieldName string) *string {
	val := getFieldByName(ctx, fieldName)
	if val == nil {
		return nil
	}

	switch val := val.(type) {
	case string:
		return &val
	case *string:
		return val
	}

	return nil
}

// GetMapInputVariableByName returns the input variable by name for map variables (e.g. input)
func GetMapInputVariableByName(ctx context.Context, fieldName string) *map[string]any {
	val := getFieldByName(ctx, fieldName)
	if val == nil {
		return nil
	}

	switch val := val.(type) {
	case map[string]any:
		return &val
	case *map[string]any:
		return val
	}

	return nil
}

// getNestedPreloads returns the nested preloads for the current graphql operation
func getNestedPreloads(ctx *graphql.OperationContext, fields []graphql.CollectedField, prefix string) (preloads []string) {
	for _, column := range fields {
		prefixColumn := getPreloadString(prefix, column.Name)
		preloads = append(preloads, prefixColumn)
		preloads = append(preloads, getNestedPreloads(ctx, graphql.CollectFields(ctx, column.Selections, nil), prefixColumn)...)
	}

	return
}

// getPreloadString returns the preload string for the given prefix and name
func getPreloadString(prefix, name string) string {
	if len(prefix) > 0 {
		return prefix + "." + name
	}

	return name
}

// getFieldByName returns the field value by name from the graphql request
// this returns a generic type, so it needs to be type asserted to the correct type
// by the exported functions
func getFieldByName(ctx context.Context, fieldName string) any {
	// skip if the context is not a graphql operation context
	if !graphql.HasOperationContext(ctx) {
		return nil
	}

	// get the root field context
	rootCtx := graphql.GetRootFieldContext(ctx)

	// determine the variable name used in the request for the field
	field := rootCtx.Field.Arguments.ForName(fieldName)
	if field == nil || field.Value == nil {
		return nil
	}

	variableName := field.Value.Raw

	opCtx := graphql.GetOperationContext(ctx)
	if opCtx == nil || opCtx.Variables == nil {
		return nil
	}

	val, ok := opCtx.Variables[variableName]
	if !ok {
		return nil
	}

	return val
}
