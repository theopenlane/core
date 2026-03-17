package runtime

import (
	"errors"
	"fmt"
	"testing"

	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/registry"
)

func TestNormalizeDispatchError(t *testing.T) {
	t.Parallel()

	unexpectedErr := errors.New("boom")

	tests := []struct {
		name        string
		err         error
		expectedErr error
	}{
		{
			name:        "nil",
			err:         nil,
			expectedErr: nil,
		},
		{
			name:        "definition not found",
			err:         registry.ErrDefinitionNotFound,
			expectedErr: ErrDefinitionNotFound,
		},
		{
			name:        "wrapped operation not found",
			err:         fmt.Errorf("wrapped: %w", registry.ErrOperationNotFound),
			expectedErr: ErrOperationNotFound,
		},
		{
			name:        "dispatch input invalid",
			err:         operations.ErrDispatchInputInvalid,
			expectedErr: operations.ErrDispatchInputInvalid,
		},
		{
			name:        "unexpected error",
			err:         unexpectedErr,
			expectedErr: unexpectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeDispatchError(tt.err)
			switch {
			case tt.expectedErr == nil:
				if got != nil {
					t.Fatalf("expected nil error, got %v", got)
				}
			case !errors.Is(got, tt.expectedErr):
				t.Fatalf("expected %v, got %v", tt.expectedErr, got)
			}
		})
	}
}
