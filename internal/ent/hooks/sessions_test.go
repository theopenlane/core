package hooks

import (
	"context"
	"testing"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/httpserve/authmanager"
)

func TestUpdateUserAuthSessionSkipsNonInteractiveCallers(t *testing.T) {
	userID := ulids.New().String()
	orgID := ulids.New().String()

	tests := []struct {
		name   string
		caller *auth.Caller
	}{
		{
			name: "api token",
			caller: &auth.Caller{
				SubjectID:          ulids.New().String(),
				OrganizationID:     orgID,
				AuthenticationType: auth.APITokenAuthentication,
			},
		},
		{
			name: "personal access token",
			caller: &auth.Caller{
				SubjectID:          userID,
				OrganizationIDs:    []string{orgID},
				AuthenticationType: auth.PATAuthentication,
			},
		},
		{
			name: "impersonation session",
			caller: &auth.Caller{
				SubjectID:          userID,
				OrganizationID:     orgID,
				AuthenticationType: auth.JWTAuthentication,
				Impersonation:      &auth.ImpersonationContext{ImpersonatorID: ulids.New().String(), TargetUserID: userID},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := auth.WithCaller(context.Background(), tc.caller)

			got, err := updateUserAuthSession(ctx, &authmanager.Client{}, ulids.New().String())
			assert.NilError(t, err)
			assert.Check(t, got == ctx)
		})
	}
}
