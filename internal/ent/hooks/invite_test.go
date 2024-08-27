package hooks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/middleware/echocontext"
)

func TestSetRequestor(t *testing.T) {
	// setup valid user
	userID := ulids.New().String()
	userCtx, err := auth.NewTestEchoContextWithValidUser(userID)
	require.NoError(t, err)

	// create parent context
	ctx := context.WithValue(userCtx.Request().Context(), echocontext.EchoContextKey, userCtx)

	tests := []struct {
		name    string
		ctx     context.Context
		want    string
		wantErr bool
	}{
		{
			name:    "happy path, user found in context",
			ctx:     ctx,
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
			invMut.Logger = *zap.NewNop().Sugar()

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
