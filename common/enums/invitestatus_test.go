package enums_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/common/enums"
)

func TestToInviteStatus(t *testing.T) {
	testCases := []struct {
		name     string
		status   string
		expected enums.InviteStatus
	}{
		{
			name:     "invitation sent",
			status:   "invitation_sent",
			expected: enums.InvitationSent,
		},
		{
			name:     "invitation accepted",
			status:   "invitation_accepted",
			expected: enums.InvitationAccepted,
		},
		{
			name:     "invitation expired, all caps",
			status:   "INVITATION_EXPIRED",
			expected: enums.InvitationExpired,
		},
		{
			name:     "invitation approval required",
			status:   "approval_required",
			expected: enums.ApprovalRequired,
		},
		{
			name:     "invalid",
			status:   "invite_sent",
			expected: enums.InviteInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := enums.ToInviteStatus(tc.status)
			assert.Equal(t, tc.expected, *res)
		})
	}
}
