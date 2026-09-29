package interceptors

import (
	"context"
	"testing"

	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/entconfig"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/ent/generated/intercept"
)

func TestInterceptorTrustCenterControlAnonymousFilter(t *testing.T) {
	testCases := []struct {
		name           string
		modulesEnabled bool
	}{
		{
			name:           "modules enabled",
			modulesEnabled: true,
		},
		{
			name:           "modules disabled",
			modulesEnabled: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			orgID := ulids.New().String()

			client := generated.NewClient()
			client.EntConfig = &entconfig.Config{Modules: entconfig.Modules{Enabled: tc.modulesEnabled}}

			ctx := generated.NewContext(context.Background(), client)
			ctx = auth.WithCaller(ctx, auth.NewTrustCenterCaller(orgID, ulids.New().String(), "Anonymous User", ""))
			ctx = auth.ActiveTrustCenterIDKey.Set(ctx, ulids.New().String())

			traverse, ok := InterceptorTrustCenterControl().(intercept.TraverseFunc)
			assert.Assert(t, ok)

			query := &mockQuery{typ: generated.TypeControl}
			assert.NilError(t, traverse(ctx, query))

			selector := sql.Select("*").From(sql.Table(control.Table))
			for _, p := range query.predicates {
				p(selector)
			}

			stmt, args := selector.Query()
			assert.Check(t, is.Contains(stmt, control.FieldIsTrustCenterControl))
			assert.Check(t, is.Contains(stmt, control.FieldTrustCenterVisibility))
			assert.Check(t, is.Contains(stmt, control.FieldOwnerID))
			assert.Check(t, is.Contains(args, any(orgID)))
		})
	}
}
