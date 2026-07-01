package ssoutils

import (
	"context"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
)

// LoadEnforcement loads the organization setting and, when a userID is provided, the subject's
// membership and email, and returns the EnforcementInput plus the loaded setting. It is the single
// db-aware source used by both the SSO handlers and the auth middleware to feed Evaluate, so the
// membership query is projected to only the fields the decision needs
func LoadEnforcement(ctx context.Context, db *ent.Client, orgID, userID, email string) (EnforcementInput, *ent.OrganizationSetting, error) {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	setting, err := db.OrganizationSetting.Query().
		Where(organizationsetting.OrganizationID(orgID)).
		Only(allowCtx)
	if err != nil {
		return EnforcementInput{}, nil, err
	}

	in := EnforcementInput{
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
		Select(orgmembership.FieldRole, orgmembership.FieldSSOExempt).
		Only(allowCtx)
	if mErr != nil {
		return EnforcementInput{}, nil, mErr
	}

	in.IsMember = true
	in.IsOwner = member.Role == enums.RoleOwner
	in.MemberExempt = member.SSOExempt

	if in.Email == "" {
		u, uErr := db.User.Query().Where(user.ID(userID)).Select(user.FieldEmail).Only(allowCtx)
		if uErr != nil {
			return EnforcementInput{}, nil, uErr
		}

		in.Email = u.Email
	}

	return in, setting, nil
}
