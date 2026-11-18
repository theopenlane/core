package interceptors

import (
	"context"
	"testing"

	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/intercept"
)

func TestInterceptorAnonTrustCenter(t *testing.T) {
	trustCenterID := ulids.New().String()

	anonCtx := auth.WithCaller(context.Background(), auth.NewTrustCenterCaller(ulids.New().String(), ulids.New().String(), "Anonymous User", ""))
	anonCtx = auth.ActiveTrustCenterIDKey.Set(anonCtx, trustCenterID)

	testCases := []struct {
		name           string
		ctx            context.Context
		scope          AnonTrustCenterScope
		expectedErr    string
		expectedFilter bool
	}{
		{
			name:        "anonymous trust center caller is denied without a scope",
			ctx:         anonCtx,
			expectedErr: "anonymous trust center access not allowed",
		},
		{
			name:           "anonymous trust center caller is scoped to the active trust center",
			ctx:            anonCtx,
			scope:          AnonScopeTrustCenterID,
			expectedFilter: true,
		},
		{
			name:  "authenticated caller is not scoped",
			ctx:   auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String()),
			scope: AnonScopeTrustCenterID,
		},
		{
			name: "authenticated caller is not denied without a scope",
			ctx:  auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String()),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			traverse, ok := InterceptorAnonTrustCenter(tc.scope).(intercept.TraverseFunc)
			assert.Assert(t, ok)

			query := &mockQuery{typ: generated.TypeTrustCenterDoc}

			err := traverse(tc.ctx, query)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Len(query.predicates, 0))

				return
			}

			assert.NilError(t, err)

			if !tc.expectedFilter {
				assert.Check(t, is.Len(query.predicates, 0))

				return
			}

			selector := sql.Select("*").From(sql.Table("trust_center_docs"))
			for _, p := range query.predicates {
				p(selector)
			}

			stmt, args := selector.Query()
			assert.Check(t, is.Contains(stmt, "trust_center_id"))
			assert.Check(t, is.Contains(args, any(trustCenterID)))
		})
	}
}
