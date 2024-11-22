package schema

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

const (
	ownerFieldName = "owner_id"
)

// NewOrgOwnMixinWithRef creates a new OrgOwnedMixin with the given ref
// and sets the defaults
func NewOrgOwnMixinWithRef(ref string) ObjectOwnedMixin {
	return NewOrgOwnedMixin(
		ObjectOwnedMixin{
			Ref: ref,
		})
}

// NewOrgOwnedMixin creates a new OrgOwnedMixin with the given ObjectOwnedMixin
// and sets the Kind to ownerFieldName and the HookFunc to defaultOrgHookFunc
func NewOrgOwnedMixin(o ObjectOwnedMixin) ObjectOwnedMixin {
	o.FieldNames = []string{ownerFieldName}
	o.Kind = Organization.Type

	if o.HookFuncs == nil {
		o.HookFuncs = []HookFunc{defaultOrgHookFunc}
	}

	if o.InterceptorFunc == nil {
		o.InterceptorFunc = defaultOrgInterceptorFunc
	}

	return o
}

var defaultOrgHookFunc HookFunc = func(o ObjectOwnedMixin) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// skip the hook if the context has the token type
			// this is useful for tokens, where the user is not yet authenticated to
			// a particular organization yet and auth policy allows this
			for _, tokenType := range o.SkipTokenType {
				if rule.ContextHasPrivacyTokenOfType(ctx, tokenType) {
					return next.Mutate(ctx, m)
				}
			}

			// set owner on create mutation
			if m.Op() == ent.OpCreate {
				orgID, err := auth.GetOrganizationIDFromContext(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to get organization id from context: %w", err)
				}

				// set owner on mutation
				if err := m.SetField(ownerFieldName, orgID); err != nil {
					return nil, err
				}
			} else {
				orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to get organization id from context: %w", err)
				}

				// filter by owner on update and delete mutations
				mx, ok := m.(interface {
					SetOp(ent.Op)
					Client() *generated.Client
					WhereP(...func(*sql.Selector))
				})
				if !ok {
					return nil, ErrUnexpectedMutationType
				}

				o.P(mx, orgIDs)
			}

			return next.Mutate(ctx, m)
		})
	}
}

var defaultOrgInterceptorFunc InterceptorFunc = func(o ObjectOwnedMixin) ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		// skip the interceptor if the context has the token type
		// this is useful for tokens, where the user is not yet authenticated to
		// a particular organization yet
		for _, tokenType := range o.SkipTokenType {
			if rule.ContextHasPrivacyTokenOfType(ctx, tokenType) {
				return nil
			}
		}

		// check query context skips
		ctxQuery := ent.QueryFromContext(ctx)

		switch o.SkipInterceptor {
		case interceptors.SkipAll:
			return nil
		case interceptors.SkipOnlyQuery:
			{
				if ctxQuery.Op == "Only" {
					return nil
				}
			}
		}

		// add owner id(s) to the query
		orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
		if err != nil {
			return err
		}

		// sets the owner id on the query for the current organization
		o.P(q, orgIDs)

		return nil
	})
}
