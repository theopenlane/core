package hooks

import (
	"context"
	"testing"

	"entgo.io/ent"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
)

func TestHookRequestor(t *testing.T) {
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
			invMut.SetOp(ent.OpCreate)

			var mockRequestorID string
			var mockOk bool

			nextMutator := ent.MutateFunc(func(_ context.Context, m ent.Mutation) (ent.Value, error) {
				if inv, ok := m.(*generated.InviteMutation); ok {
					mockRequestorID, mockOk = inv.RequestorID()
				}

				return nil, nil
			})

			hook := HookRequestor()
			_, err := hook(nextMutator).Mutate(tt.ctx, invMut)

			if tt.wantErr {
				require.Error(t, err)
				assert.False(t, mockOk)

				return
			}

			require.NoError(t, err)
			require.True(t, mockOk)

			assert.Equal(t, userID, mockRequestorID)
		})
	}
}
