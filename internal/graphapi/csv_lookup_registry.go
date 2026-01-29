package graphapi

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/internal/ent/csvgenerated"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/iam/auth"
)

// csvLookupFn is the function signature for CSV reference lookups used internally.
type csvLookupFn func(ctx context.Context, values []string) (map[string]string, error)

// csvCreateFn is the function signature for CSV reference auto-creation used internally.
type csvCreateFn func(ctx context.Context, values []string) (map[string]string, error)

// wrapGeneratedLookup wraps a generated lookup function to get client and orgID from context.
func wrapGeneratedLookup(fn csvgenerated.CSVLookupFn) csvLookupFn {
	return func(ctx context.Context, values []string) (map[string]string, error) {
		client := withTransactionalMutation(ctx)

		orgID, err := auth.GetOrganizationIDFromContext(ctx)
		if err != nil {
			return nil, common.NewValidationError("organization id not found in context")
		}

		return fn(ctx, client, orgID, values)
	}
}

// wrapGeneratedCreate wraps a generated create function to get client and orgID from context.
func wrapGeneratedCreate(fn csvgenerated.CSVCreateFn) csvCreateFn {
	return func(ctx context.Context, values []string) (map[string]string, error) {
		client := withTransactionalMutation(ctx)

		orgID, err := auth.GetOrganizationIDFromContext(ctx)
		if err != nil {
			return nil, common.NewValidationError("organization id not found in context")
		}

		return fn(ctx, client, orgID, values)
	}
}

// BuildCSVReferenceRulesFromRegistry creates CSVReferenceRule objects from the generated registry.
func BuildCSVReferenceRulesFromRegistry(schemaName string) ([]CSVReferenceRule, error) {
	registryRules := csvgenerated.GetCSVReferenceRules(schemaName)
	if len(registryRules) == 0 {
		return nil, nil
	}

	rules := make([]CSVReferenceRule, 0, len(registryRules))

	for _, rr := range registryRules {
		entry, ok := csvgenerated.GetCSVLookupEntry(rr.TargetEntity, rr.MatchField)
		if !ok {
			return nil, common.NewValidationError(
				fmt.Sprintf("no lookup registered for %s.%s", rr.TargetEntity, rr.MatchField),
			)
		}

		if rr.CreateIfMissing && entry.Create == nil {
			return nil, common.NewValidationError(
				fmt.Sprintf("create not supported for %s.%s", rr.TargetEntity, rr.MatchField),
			)
		}

		rule := CSVReferenceRule{
			SourceField: rr.SourceColumn,
			TargetField: rr.TargetField,
			Lookup:      wrapGeneratedLookup(entry.Lookup),
			Normalize:   normalizeCSVReferenceKey,
		}

		if rr.CreateIfMissing && entry.Create != nil {
			rule.Create = wrapGeneratedCreate(entry.Create)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}
