package graphapi

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/internal/ent/csvgenerated"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/iam/auth"
)

// wrapGeneratedFn wraps a generated lookup or create function to get client and orgID from context.
func wrapGeneratedFn[F csvgenerated.CSVLookupFn | csvgenerated.CSVCreateFn](fn F) func(ctx context.Context, values []string) (map[string]string, error) {
	return func(ctx context.Context, values []string) (map[string]string, error) {
		client := withTransactionalMutation(ctx)

		caller, ok := auth.CallerFromContext(ctx)
		if !ok || caller == nil || caller.OrganizationID == "" {
			return nil, common.NewValidationError("organization id not found in context")
		}

		return (csvgenerated.CSVLookupFn)(fn)(ctx, client, caller.OrganizationID, values)
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
			Lookup:      wrapGeneratedFn(entry.Lookup),
			Normalize:   normalizeCSVReferenceKey,
		}

		if rr.CreateIfMissing && entry.Create != nil {
			rule.Create = wrapGeneratedFn(entry.Create)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}
