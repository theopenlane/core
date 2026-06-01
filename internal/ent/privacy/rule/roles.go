package rule

import (
	"strings"

	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/enums"
)

// InviteRelationForRole returns the organization relation required to assign a role.
func InviteRelationForRole(role enums.Role) string {
	switch strings.ToLower(role.String()) {
	case fgax.MemberRelation:
		return inviteMemberRelation
	case fgax.AdminRelation:
		return inviteAdminRelation
	case fgax.SuperAdminRelation:
		return inviteSuperAdminRelation
	case fgax.AuditorRelation:
		return inviteAuditors
	default:
		return inviteSuperAdminRelation
	}
}
