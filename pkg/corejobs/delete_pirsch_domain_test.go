package corejobs_test

import (
	"context"
	"testing"

	"github.com/riverqueue/river"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/corejobs"

	pirschmocks "github.com/theopenlane/core/pkg/corejobs/internal/pirsch/mocks"
)

func TestDeletePirschDomainWorker(t *testing.T) {
	t.Parallel()

	pirschDomainID := "pirsch123"

	testCases := []struct {
		name              string
		pirschDomainID    string
		expectDeleteCall  bool
		deleteDomainError error
		expectedError     string
	}{
		{
			name:             "happy path",
			pirschDomainID:   pirschDomainID,
			expectDeleteCall: true,
		},
		{
			name:          "missing pirsch domain id",
			pirschDomainID: "",
			expectedError: "pirsch_domain_id is required for the delete_pirsch_domain job",
		},
		{
			name:              "delete domain fails",
			pirschDomainID:    pirschDomainID,
			expectDeleteCall:  true,
			deleteDomainError: ErrTest,
			expectedError:     "test error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			pirschMock := pirschmocks.NewMockClient(t)

			if tc.expectDeleteCall {
				pirschMock.EXPECT().DeleteDomain(mock.Anything, tc.pirschDomainID).Return(tc.deleteDomainError)
			}

			worker := &corejobs.DeletePirschDomainWorker{
				Config: corejobs.PirschDomainConfig{
					PirschClientID:     "test_client_id",
					PirschClientSecret: "test_client_secret",
				},
			}

			worker.WithPirschClient(pirschMock)

			err := worker.Work(ctx, &river.Job[corejobs.DeletePirschDomainArgs]{Args: corejobs.DeletePirschDomainArgs{
				PirschDomainID: tc.pirschDomainID,
			}})

			if tc.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
			}

			if !tc.expectDeleteCall {
				pirschMock.AssertNotCalled(t, "DeleteDomain")
			}
		})
	}
}
