//go:build cli

package cmd

import (
	"strings"

	"github.com/theopenlane/core/common/enums"
)

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

// ParseIDList parses a comma-separated list of IDs into a string slice
func ParseIDList(ids string) []string {
	var result []string

	for _, id := range strings.Split(ids, ",") {
		id = strings.TrimSpace(id)
		if id != "" {
			result = append(result, id)
		}
	}

	return result
}
