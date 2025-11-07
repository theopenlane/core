//go:build cli

package speccli

import (
	"strings"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
)

// ParseRole maps a CLI string to an enum role or returns ErrInvalidRole.
func ParseRole(role string) (enums.Role, error) {
	r := enums.ToRole(role)
	if r == nil {
		return enums.RoleInvalid, cmdpkg.ErrInvalidRole
	}

	if strings.EqualFold(r.String(), enums.RoleInvalid.String()) {
		return enums.RoleInvalid, cmdpkg.ErrInvalidRole
	}

	return *r, nil
}

// ParseInviteStatus resolves a string into an invite status or returns ErrInvalidInviteStatus.
func ParseInviteStatus(status string) (enums.InviteStatus, error) {
	s := enums.ToInviteStatus(status)
	if s == nil {
		return enums.InviteInvalid, cmdpkg.ErrInvalidInviteStatus
	}

	if strings.EqualFold(s.String(), enums.InviteInvalid.String()) {
		return enums.InviteInvalid, cmdpkg.ErrInvalidInviteStatus
	}

	return *s, nil
}
