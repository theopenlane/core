package olauth

import (
	"context"
	"strings"
	"testing"

	"github.com/theopenlane/ent/generated"
	genprivacy "github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/ent/privacy/rule"
)

func TestCreateClaimsWithOrg(t *testing.T) {
	user := &generated.User{ID: "u1", Edges: generated.UserEdges{Setting: &generated.UserSetting{Edges: generated.UserSettingEdges{DefaultOrg: &generated.Organization{ID: "org1"}}}}}
	c := createClaimsWithOrg(context.Background(), user, "")
	if c.OrgID != "org1" || c.UserID != "u1" {
		t.Fatalf("unexpected claims: %#v", c)
	}
	c2 := createClaimsWithOrg(context.Background(), user, "org2")
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

func TestAnonymousPrefixes(t *testing.T) {
	// Verify that the prefix constants are correct
	if !strings.HasPrefix(AnonTrustcenterJWTPrefix, "anon_") {
		t.Errorf("AnonTrustcenterJWTPrefix should start with 'anon_', got: %s", AnonTrustcenterJWTPrefix)
	}

	if !strings.HasPrefix(AnonQuestionnaireJWTPrefix, "anon_") {
		t.Errorf("AnonQuestionnaireJWTPrefix should start with 'anon_', got: %s", AnonQuestionnaireJWTPrefix)
	}

	// Verify they are different
	if AnonTrustcenterJWTPrefix == AnonQuestionnaireJWTPrefix {
		t.Error("AnonTrustcenterJWTPrefix and AnonQuestionnaireJWTPrefix should be different")
	}

	// Verify expected values
	if AnonTrustcenterJWTPrefix != "anon_trustcenter_" {
		t.Errorf("Expected AnonTrustcenterJWTPrefix to be 'anon_trustcenter_', got: %s", AnonTrustcenterJWTPrefix)
	}

	if AnonQuestionnaireJWTPrefix != "anon_questionnaire_" {
		t.Errorf("Expected AnonQuestionnaireJWTPrefix to be 'anon_questionnaire_', got: %s", AnonQuestionnaireJWTPrefix)
	}
}
