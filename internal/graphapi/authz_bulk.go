package graphapi

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
)

// filterAuthorizedIDs checks which IDs the caller has access to for the given relation
// and returns only the authorized IDs. 
func (r *mutationResolver) filterAuthorizedIDs(ctx context.Context, ids []string, objectType fgax.Kind, relation string) []string {
	if len(ids) == 0 {
		return ids
	}

	if !generated.IsSelfAccessType(string(objectType)) {
		return ids
	}

	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		logx.FromContext(ctx).Error().Msg("unable to get caller from context for bulk access check")
		return nil
	}

	email := caller.SubjectEmail
	domain := email
	if strings.Contains(email, "@") {
		domain = strings.Split(email, "@")[1]
	}

	orgContext := &map[string]any{
		"email_domain": domain,
	}

	checks := make([]fgax.AccessCheck, 0, len(ids))
	for _, id := range ids {
		ac := fgax.AccessCheck{
			Relation:    relation,
			ObjectType:  objectType,
			ObjectID:    id,
			SubjectType: caller.SubjectType(),
			SubjectID:   caller.SubjectID,
		}
		if domain != "" {
			ac.Context = orgContext
		}
		checks = append(checks, ac)
	}

	allowedIDs, err := r.db.Authz.BatchCheckObjectAccess(ctx, checks)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("object_type", string(objectType)).Msg("error checking bulk access")
		return nil
	}

	if gCtx := graphql.GetFieldContext(ctx); gCtx != nil {
		gCtx.Args["ids"] = allowedIDs
	}

	return allowedIDs
}
