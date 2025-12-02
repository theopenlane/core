//go:build cli

package cmd

import "github.com/theopenlane/shared/enums"

// GetRoleEnum returns the Role if valid, otherwise returns an error
func GetRoleEnum(role string) (enums.Role, error) {
	r := enums.ToRole(role)

	if r.String() == enums.RoleInvalid.String() {
		return *r, ErrInvalidRole
	}

	return *r, nil
}

// GetInviteStatusEnum returns the invitation status if valid, otherwise returns an error
func GetInviteStatusEnum(status string) (enums.InviteStatus, error) {
	r := enums.ToInviteStatus(status)

	if r.String() == enums.InviteInvalid.String() {
		return *r, ErrInvalidInviteStatus
	}

	return *r, nil
}
