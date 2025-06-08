package authmanager

import (
	"context"
	"testing"

	"github.com/theopenlane/core/internal/ent/generated"
	genprivacy "github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

func TestCreateClaimsWithOrg(t *testing.T) {
	user := &generated.User{ID: "u1", Edges: generated.UserEdges{Setting: &generated.UserSetting{Edges: generated.UserSettingEdges{DefaultOrg: &generated.Organization{ID: "org1"}}}}}
	c := createClaimsWithOrg(user, "")
	if c.OrgID != "org1" || c.UserID != "u1" {
		t.Fatalf("unexpected claims: %#v", c)
	}
	c2 := createClaimsWithOrg(user, "org2")
	if c2.OrgID != "org2" {
		t.Fatalf("target org not used")
	}
}

func TestSkipOrgValidation(t *testing.T) {
	ctx := context.Background()
	if skipOrgValidation(ctx) {
		t.Fatalf("expected false")
	}
	ctx = rule.WithInternalContext(ctx)
	if !skipOrgValidation(ctx) {
		t.Fatalf("expected true for internal request")
	}
	ctx = genprivacy.DecisionContext(context.Background(), genprivacy.Allow)
	if !skipOrgValidation(ctx) {
		t.Fatalf("expected true when allowed in privacy decision")
	}
}
