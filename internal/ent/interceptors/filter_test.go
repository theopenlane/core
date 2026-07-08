package interceptors

import (
	"context"
	"testing"

	"entgo.io/ent/dialect/sql"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/entx/history"

	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

type mockQuery struct {
	typ string
}

func (m *mockQuery) Type() string                  { return m.typ }
func (m *mockQuery) Limit(int)                     {}
func (m *mockQuery) Offset(int)                    {}
func (m *mockQuery) Unique(bool)                   {}
func (m *mockQuery) Order(...func(*sql.Selector))  {}
func (m *mockQuery) WhereP(...func(*sql.Selector)) {}

func orgQuery() intercept.Query { return &mockQuery{typ: "Organization"} }

func alwaysSkip(_ context.Context) bool  { return true }
func neverSkip(_ context.Context) bool   { return false }
func alwaysForce(_ context.Context) bool { return true }

func TestSkipFilter(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		q           intercept.Query
		forceFilter SkipperFunc
		customSkip  SkipperFunc
		want        bool
	}{
		{
			name:        "privacy allow context skips regardless of forceFilter",
			ctx:         privacy.DecisionContext(context.Background(), privacy.Allow),
			forceFilter: alwaysForce,
			want:        true,
		},
		{
			name:        "internal request context skips regardless of forceFilter",
			ctx:         rule.WithInternalContext(context.Background()),
			forceFilter: alwaysForce,
			want:        true,
		},
		{
			name:        "history request context skips regardless of forceFilter",
			ctx:         history.WithContext(context.Background()),
			forceFilter: alwaysForce,
			want:        true,
		},
		{
			name: "organization scope check returns allow when forceFilter is nil",
			ctx:  context.Background(),
			q:    orgQuery(),
			want: true,
		},
		{
			name:        "forceFilter prevents scope check from skipping for Organization",
			ctx:         context.Background(),
			q:           orgQuery(),
			forceFilter: alwaysForce,
			want:        false,
		},
		{
			name:        "forceFilter does not block customSkipperFunc",
			ctx:         context.Background(),
			q:           orgQuery(),
			forceFilter: alwaysForce,
			customSkip:  alwaysSkip,
			want:        true,
		},

		{
			name: "no skip reasons does not skip",
			ctx:  context.Background(),
			q:    &mockQuery{},
			want: false,
		},
		{
			name:        "forceFilter alone does not skip",
			ctx:         context.Background(),
			q:           &mockQuery{},
			forceFilter: alwaysForce,
			want:        false,
		},
		{
			name:       "customSkipperFunc returning true skips",
			ctx:        context.Background(),
			q:          &mockQuery{},
			customSkip: alwaysSkip,
			want:       true,
		},
		{
			name:       "customSkipperFunc returning false does not skip",
			ctx:        context.Background(),
			q:          &mockQuery{},
			customSkip: neverSkip,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			if tt.customSkip != nil {
				got = skipFilter(tt.ctx, tt.q, tt.forceFilter, tt.customSkip)
			} else {
				got = skipFilter(tt.ctx, tt.q, tt.forceFilter)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
