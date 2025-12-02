package hooks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/ent/generated"
)

func TestSetRequestor(t *testing.T) {
	// setup valid user
	userID := ulids.New().String()
	userCtx := auth.NewTestContextWithValidUser(userID)

	tests := []struct {
		name    string
		ctx     context.Context
		want    string
		wantErr bool
	}{
		{
			name:    "happy path, user found in context",
			ctx:     userCtx,
			want:    userID,
			wantErr: false,
		},
		{
			name:    "no user in context",
			ctx:     context.Background(),
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invMut := &generated.InviteMutation{}

			got, err := setRequestor(tt.ctx, invMut)

			requestor, ok := got.RequestorID()

			if tt.wantErr {
				require.Error(t, err)
				assert.False(t, ok)

				return
			}

			require.NoError(t, err)
			require.True(t, ok)

			assert.Equal(t, userID, requestor)
		})
	}
}
