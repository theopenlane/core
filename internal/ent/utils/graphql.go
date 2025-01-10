package utils

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
)

// CheckForRequestedField checks if the requested field is in the list of fields
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
