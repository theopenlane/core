package oauthserver

import (
	"context"
	"testing"

	"github.com/theopenlane/iam/fgax"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
)

func TestAccessScopes(t *testing.T) {
	tests := []struct {
		name  string
		level enums.OAuthAccessLevel
		want  []string
	}{
		{name: "read", level: enums.OAuthAccessLevelRead, want: []string{ScopeRead}},
		{name: "write", level: enums.OAuthAccessLevelWrite, want: []string{ScopeRead, ScopeWrite}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Check(t, is.DeepEqual(AccessScopes(tc.level), tc.want))
		})
	}
}

func TestMaxAccess(t *testing.T) {
	tests := []struct {
		name      string
		requested []string
		want      enums.OAuthAccessLevel
	}{
		{name: "no requested scopes", want: enums.OAuthAccessLevelWrite},
		{name: "read scope", requested: []string{ScopeRead}, want: enums.OAuthAccessLevelRead},
		{name: "object write scope", requested: []string{ScopeRead, "control:write"}, want: enums.OAuthAccessLevelWrite},
		{name: "object read scope", requested: []string{ScopeUser, "control:read"}, want: enums.OAuthAccessLevelRead},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Check(t, is.Equal(MaxAccess(tc.requested), tc.want))
		})
	}
}

func TestObjectScopes(t *testing.T) {
	tests := []struct {
		name   string
		scopes []string
		want   []string
	}{
		{name: "mixed", scopes: []string{ScopeUser, "control:read", ScopeWrite, "risk:delete"}, want: []string{"control:read", "risk:delete"}},
		{name: "personal only", scopes: []string{ScopeRead}, want: []string{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Check(t, is.DeepEqual(ObjectScopes(tc.scopes), tc.want))
		})
	}
}

func TestInvalidObjectScopes(t *testing.T) {
	tests := []struct {
		name   string
		scopes []string
		want   []string
	}{
		{name: "all valid", scopes: []string{"control:read", "risk:delete"}, want: []string{}},
		{name: "unknown verb", scopes: []string{"control:read", "control:admin"}, want: []string{"control:admin"}},
		{name: "unknown object", scopes: []string{"bogus:read"}, want: []string{"bogus:read"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Check(t, is.DeepEqual(InvalidObjectScopes(tc.scopes), tc.want))
		})
	}
}

func TestScopeSetAllows(t *testing.T) {
	tests := []struct {
		name     string
		scopes   []string
		object   string
		relation string
		want     bool
	}{
		{name: "read scope reads everything", scopes: []string{ScopeRead}, object: "risk", relation: fgax.CanView, want: true},
		{name: "read scope does not write", scopes: []string{ScopeRead}, object: "risk", relation: fgax.CanEdit, want: false},
		{name: "write scope deletes", scopes: []string{ScopeWrite}, object: "risk", relation: fgax.CanDelete, want: true},
		{name: "object read", scopes: []string{"control:read"}, object: "control", relation: fgax.CanView, want: true},
		{name: "object write implies read", scopes: []string{"control:write"}, object: "control", relation: fgax.CanView, want: true},
		{name: "object write does not imply delete", scopes: []string{"control:write"}, object: "control", relation: fgax.CanDelete, want: false},
		{name: "object delete implies write", scopes: []string{"control:delete"}, object: "control", relation: fgax.CanEdit, want: true},
		{name: "other object", scopes: []string{"control:delete"}, object: "risk", relation: fgax.CanView, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Check(t, is.Equal(NewScopeSet(tc.scopes).Allows(tc.object, tc.relation), tc.want))
		})
	}
}

func TestValidatePersonalGrant(t *testing.T) {
	mcp := []string{ScopeRead, "control:write", "entity:write"}

	tests := []struct {
		name      string
		requested []string
		granted   []string
		want      enums.OAuthAccessLevel
		wantErr   error
	}{
		{name: "everything requested", requested: mcp, granted: mcp, want: enums.OAuthAccessLevelWrite},
		{name: "read only", requested: mcp, granted: []string{ScopeRead}, want: enums.OAuthAccessLevelRead},
		{name: "object reads and write", requested: mcp, granted: []string{"control:read", "evidence:read", "control:write"}, want: enums.OAuthAccessLevelWrite},
		{name: "unrequested object write", requested: mcp, granted: []string{ScopeRead, "risk:write"}, wantErr: ErrScopesNotRequested},
		{name: "unrequested delete", requested: mcp, granted: []string{ScopeRead, "control:delete"}, wantErr: ErrScopesNotRequested},
		{name: "unrequested full write", requested: mcp, granted: []string{ScopeWrite}, wantErr: ErrScopesNotRequested},
		{name: "full write covers object delete", requested: []string{ScopeRead, ScopeWrite}, granted: []string{"risk:delete"}, want: enums.OAuthAccessLevelWrite},
		{name: "object write covers object read", requested: []string{"control:write"}, granted: []string{"control:read"}, want: enums.OAuthAccessLevelRead},
		{name: "nothing requested", granted: []string{ScopeWrite}, want: enums.OAuthAccessLevelWrite},
		{name: "nothing granted", requested: mcp, wantErr: ErrNoScopes},
		{name: "user scope only", requested: []string{ScopeUser, "control:read"}, granted: []string{ScopeUser}, wantErr: ErrNoScopes},
		{name: "user scope with object read", requested: []string{ScopeUser, "control:read", "control:write"}, granted: []string{ScopeUser, "control:read"}, want: enums.OAuthAccessLevelRead},
		{name: "user scope with unrequested object", requested: []string{ScopeUser, "control:read"}, granted: []string{ScopeUser, "evidence:read"}, wantErr: ErrScopesNotRequested},
		{name: "invalid verb", requested: mcp, granted: []string{"control:admin"}, wantErr: ErrInvalidPersonalScope},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			access, err := ValidatePersonalGrant(tc.requested, tc.granted)
			if tc.wantErr != nil {
				assert.Check(t, is.ErrorIs(err, tc.wantErr))
				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.Equal(access, tc.want))
		})
	}
}

func TestValidateOrganizationGrant(t *testing.T) {
	requested := []string{"control:read", "evidence:write"}

	tests := []struct {
		name      string
		requested []string
		granted   []string
		wantErr   error
	}{
		{name: "subset of requested", requested: requested, granted: []string{"control:read"}},
		{name: "nothing requested", granted: []string{"risk:delete"}},
		{name: "unrequested scope", requested: requested, granted: []string{"control:read", "risk:read"}, wantErr: ErrScopesNotRequested},
		{name: "unknown scope", requested: requested, granted: []string{"control:admin"}, wantErr: ErrInvalidObjectScope},
		{name: "nothing granted", requested: requested, wantErr: ErrNoScopes},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateOrganizationGrant(tc.requested, tc.granted)
			if tc.wantErr != nil {
				assert.Check(t, is.ErrorIs(err, tc.wantErr))
				return
			}

			assert.NilError(t, err)
		})
	}
}

func TestClassifyScopes(t *testing.T) {
	tests := []struct {
		name        string
		scopes      []string
		wantKind    enums.OAuthGrantKind
		wantDecided bool
		wantErr     error
	}{
		{name: "no scopes"},
		{name: "read", scopes: []string{ScopeRead}, wantKind: enums.OAuthGrantKindUser, wantDecided: true},
		{name: "read and write", scopes: []string{ScopeRead, ScopeWrite}, wantKind: enums.OAuthGrantKindUser, wantDecided: true},
		{name: "object scopes only", scopes: []string{"control:read", "evidence:write"}, wantKind: enums.OAuthGrantKindOrganization, wantDecided: true},
		{name: "read with object writes", scopes: []string{ScopeRead, "control:write", "risk:delete"}, wantKind: enums.OAuthGrantKindUser, wantDecided: true},
		{name: "user with object read", scopes: []string{ScopeUser, "control:read"}, wantKind: enums.OAuthGrantKindUser, wantDecided: true},
		{name: "personal with invalid verb", scopes: []string{ScopeRead, "control:admin"}, wantErr: ErrInvalidPersonalScope},
		{name: "organization with unknown scope", scopes: []string{"control:read", "bogus:read"}, wantErr: ErrInvalidObjectScope},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kind, decided, err := ClassifyScopes(tc.scopes)
			if tc.wantErr != nil {
				assert.Check(t, is.ErrorIs(err, tc.wantErr))
				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.Equal(kind, tc.wantKind))
			assert.Check(t, is.Equal(decided, tc.wantDecided))
		})
	}
}

func TestCallerGrantReadLimited(t *testing.T) {
	tests := []struct {
		name  string
		grant CallerGrant
		want  bool
	}{
		{name: "read scope", grant: CallerGrant{Kind: enums.OAuthGrantKindUser, Access: enums.OAuthAccessLevelWrite, Scopes: []string{ScopeRead, "control:write"}}, want: false},
		{name: "object reads", grant: CallerGrant{Kind: enums.OAuthGrantKindUser, Access: enums.OAuthAccessLevelWrite, Scopes: []string{"control:read", "risk:delete"}}, want: true},
		{name: "organization install", grant: CallerGrant{Kind: enums.OAuthGrantKindOrganization}, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Check(t, is.Equal(tc.grant.ReadLimited(), tc.want))
		})
	}
}

func TestCallerGrantWriteLimited(t *testing.T) {
	tests := []struct {
		name  string
		grant CallerGrant
		want  bool
	}{
		{name: "object writes", grant: CallerGrant{Kind: enums.OAuthGrantKindUser, Access: enums.OAuthAccessLevelWrite, Scopes: []string{ScopeRead, "control:write"}}, want: true},
		{name: "full write", grant: CallerGrant{Kind: enums.OAuthGrantKindUser, Access: enums.OAuthAccessLevelWrite, Scopes: []string{ScopeRead, ScopeWrite}}, want: false},
		{name: "read only", grant: CallerGrant{Kind: enums.OAuthGrantKindUser, Access: enums.OAuthAccessLevelRead, Scopes: []string{ScopeRead}}, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Check(t, is.Equal(tc.grant.WriteLimited(), tc.want))
		})
	}
}

func TestCallerGrantFromContext(t *testing.T) {
	grant := CallerGrant{ID: "g1", Access: enums.OAuthAccessLevelRead}

	tests := []struct {
		name   string
		ctx    context.Context
		want   CallerGrant
		wantOK bool
	}{
		{name: "empty context", ctx: context.Background()},
		{name: "grant set", ctx: WithCallerGrant(context.Background(), grant), want: grant, wantOK: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := CallerGrantFromContext(tc.ctx)
			assert.Check(t, is.Equal(ok, tc.wantOK))
			assert.Check(t, is.DeepEqual(got, tc.want))
		})
	}
}
