// Package ssoenforcement holds the db-aware loader
package ssoenforcement

import (
	"context"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/user"
	sso "github.com/theopenlane/core/v2/pkg/ssoutils"
)

// LoadEnforcement loads the organization setting and, when a userID is provided, the subject's
// membership and email, and returns the EnforcementInput plus the loaded setting. It is the single
// db-aware source used by both the SSO handlers and the auth middleware to feed Evaluate, so the
// membership query is projected to only the fields the decision needs
func LoadEnforcement(ctx context.Context, db *ent.Client, orgID, userID, email string) (sso.EnforcementInput, *ent.OrganizationSetting, error) {
	// callers include unauthenticated sso checks with no caller, every lookup below is pinned to orgID or userID
	lookupCtx := auth.WithInternalReadCrossOrgContext(ctx)

	setting, err := db.OrganizationSetting.Query().
		Where(organizationsetting.OrganizationID(orgID)).
		Only(lookupCtx)
	if err != nil {
		return sso.EnforcementInput{}, nil, err
	}

	in := sso.EnforcementInput{
		SSOEnforced:   setting.IdentityProviderLoginEnforced,
		TFAEnforced:   setting.MultifactorAuthEnforced,
		ExemptDomains: setting.SSOExemptDomains,
		Email:         email,
	}

	if userID == "" {
		return in, setting, nil
	}

	// every caller passes a userID that is expected to be a member of orgID, so a missing membership is
	// an invariant violation, not a benign non-member; surface it like any other error
	member, mErr := db.OrgMembership.Query().
		Where(orgmembership.OrganizationID(orgID), orgmembership.UserID(userID)).
		Select(orgmembership.FieldRole, orgmembership.FieldSSOExempt, orgmembership.FieldTfaEnforced).
		Only(lookupCtx)
	if mErr != nil {
		return sso.EnforcementInput{}, nil, mErr
	}

	in.IsMember = true
	in.IsOwner = member.Role == enums.RoleOwner
	in.MemberExempt = member.SSOExempt
	in.MemberTFAEnforced = member.TfaEnforced

	if in.Email == "" {
		u, uErr := db.User.Query().Where(user.ID(userID)).Select(user.FieldEmail).Only(lookupCtx)
		if uErr != nil {
			return sso.EnforcementInput{}, nil, uErr
		}

		in.Email = u.Email
	}

	return in, setting, nil
}
