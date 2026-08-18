package logx

import (
	"context"
	"testing"

	"github.com/theopenlane/iam/auth"
)

func TestWithCallerIdentity(t *testing.T) {
	ctx := auth.WithCaller(context.Background(), &auth.Caller{
		SubjectID:      "user-1",
		SubjectEmail:   "user@example.com",
		OrganizationID: "org-1",
		Capabilities:   auth.CapInternalOperation,
	})

	fields := FieldsFromContext(WithCallerIdentity(ctx))

	if fields["subject_id"] != "user-1" {
		t.Fatalf("expected subject_id field, got %v", fields["subject_id"])
	}

	if fields["subject_email"] != "user@example.com" {
		t.Fatalf("expected subject_email field, got %v", fields["subject_email"])
	}

	if fields["organization_id"] != "org-1" {
		t.Fatalf("expected organization_id field, got %v", fields["organization_id"])
	}

	if fields["capabilities"] != auth.CapInternalOperation {
		t.Fatalf("expected capabilities field, got %v", fields["capabilities"])
	}
}

func TestWithCallerIdentityWithoutCaller(t *testing.T) {
	ctx := context.Background()

	if got := WithCallerIdentity(ctx); got != ctx {
		t.Fatal("expected untouched context without a caller")
	}
}
