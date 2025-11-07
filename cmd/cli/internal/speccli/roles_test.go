//go:build cli

package speccli

import (
	"testing"

	"github.com/theopenlane/core/pkg/enums"
)

func TestParseRole(t *testing.T) {
	role, err := ParseRole("admin")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if role != enums.RoleAdmin {
		t.Fatalf("expected admin role, got %s", role)
	}

	if _, err := ParseRole("unknown"); err == nil {
		t.Fatalf("expected error for invalid role")
	}
}

func TestParseInviteStatus(t *testing.T) {
	status, err := ParseInviteStatus("invitation_sent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != enums.InvitationSent {
		t.Fatalf("expected InvitationSent, got %s", status)
	}

	if _, err := ParseInviteStatus("foobar"); err == nil {
		t.Fatalf("expected error for invalid invite status")
	}
}
