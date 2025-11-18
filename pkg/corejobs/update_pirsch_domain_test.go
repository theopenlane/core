package corejobs_test

import (
	"context"
	"testing"

	"github.com/riverqueue/river"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/core/pkg/openlaneclient"

	olmocks "github.com/theopenlane/core/pkg/corejobs/internal/olclient/mocks"
	pirschmocks "github.com/theopenlane/core/pkg/corejobs/internal/pirsch/mocks"
)

func TestUpdatePirschDomainWorker(t *testing.T) {
	t.Parallel()

	trustCenterID := "trustcenter123"
	ownerID := "owner123"
	customDomainHostname := "trust.example.com"
	pirschDomainID := "pirsch123"

	testCases := []struct {
		name                       string
		trustCenterID              string
		expectedGetTrustCenterByID bool
		trustCenterResponse        *openlaneclient.GetTrustCenterByID
		trustCenterError           error
		expectUpdateHostname       bool
		updateHostnameError        error
		expectUpdateSubdomain      bool
		updateSubdomainError       error
		expectedError              string
	}{
		{
			name:                       "happy path",
			trustCenterID:              trustCenterID,
			expectedGetTrustCenterByID: true,
			trustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID:             trustCenterID,
					OwnerID:        &ownerID,
					PirschDomainID: &pirschDomainID,
					CustomDomain: &openlaneclient.GetTrustCenterByID_TrustCenter_CustomDomain{
						CnameRecord: customDomainHostname,
					},
				},
			},
			expectUpdateHostname:  true,
			expectUpdateSubdomain: true,
		},
		{
			name:          "missing trust center id",
			trustCenterID: "",
			expectedError: "trust_center_id is required for the update_pirsch_domain job",
		},
		{
			name:                       "trust center not found",
			trustCenterID:              trustCenterID,
			expectedGetTrustCenterByID: true,
			trustCenterError:           ErrTest,
			expectedError:              "test error",
		},
		{
			name:                       "trust center without pirsch domain id",
			trustCenterID:              trustCenterID,
			expectedGetTrustCenterByID: true,
			trustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID:      trustCenterID,
					OwnerID: &ownerID,
					CustomDomain: &openlaneclient.GetTrustCenterByID_TrustCenter_CustomDomain{
						CnameRecord: customDomainHostname,
					},
				},
			},
			expectedError: "trust center does not have a pirsch domain ID",
		},
		{
			name:                       "trust center without custom domain",
			trustCenterID:              trustCenterID,
			expectedGetTrustCenterByID: true,
			trustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID:             trustCenterID,
					OwnerID:        &ownerID,
					PirschDomainID: &pirschDomainID,
					CustomDomain:   nil,
				},
			},
			expectedError: "trust center does not have a custom domain",
		},
		{
			name:                       "invalid hostname format - no dot",
			trustCenterID:              trustCenterID,
			expectedGetTrustCenterByID: true,
			trustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID:             trustCenterID,
					OwnerID:        &ownerID,
					PirschDomainID: &pirschDomainID,
					CustomDomain: &openlaneclient.GetTrustCenterByID_TrustCenter_CustomDomain{
						CnameRecord: "invalidhostname",
					},
				},
			},
			expectedError: "invalid custom domain hostname",
		},
		{
			name:                       "update hostname fails",
			trustCenterID:              trustCenterID,
			expectedGetTrustCenterByID: true,
			trustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID:             trustCenterID,
					OwnerID:        &ownerID,
					PirschDomainID: &pirschDomainID,
					CustomDomain: &openlaneclient.GetTrustCenterByID_TrustCenter_CustomDomain{
						CnameRecord: customDomainHostname,
					},
				},
			},
			expectUpdateHostname: true,
			updateHostnameError:  ErrTest,
			expectedError:        "test error",
		},
		{
			name:                       "update subdomain fails",
			trustCenterID:              trustCenterID,
			expectedGetTrustCenterByID: true,
			trustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID:             trustCenterID,
					OwnerID:        &ownerID,
					PirschDomainID: &pirschDomainID,
					CustomDomain: &openlaneclient.GetTrustCenterByID_TrustCenter_CustomDomain{
						CnameRecord: customDomainHostname,
					},
				},
			},
			expectUpdateHostname:  true,
			expectUpdateSubdomain: true,
			updateSubdomainError:  ErrTest,
			expectedError:         "test error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			pirschMock := pirschmocks.NewMockClient(t)
			olMock := olmocks.NewMockOpenlaneGraphClient(t)

			if tc.expectedGetTrustCenterByID {
				olMock.EXPECT().GetTrustCenterByID(mock.Anything, tc.trustCenterID).Return(tc.trustCenterResponse, tc.trustCenterError)
			}

			if tc.expectUpdateHostname {
				pirschMock.EXPECT().UpdateHostname(mock.Anything, pirschDomainID, "example.com").Return(tc.updateHostnameError)
			}

			if tc.expectUpdateSubdomain {
				pirschMock.EXPECT().UpdateSubdomain(mock.Anything, pirschDomainID, "trust").Return(tc.updateSubdomainError)
			}

			worker := &corejobs.UpdatePirschDomainWorker{
				Config: corejobs.PirschDomainConfig{
					PirschClientID:     "test_client_id",
					PirschClientSecret: "test_client_secret",
				},
			}

			worker.WithPirschClient(pirschMock)
			worker.WithOpenlaneClient(olMock)

			err := worker.Work(ctx, &river.Job[corejobs.UpdatePirschDomainArgs]{Args: corejobs.UpdatePirschDomainArgs{
				TrustCenterID: tc.trustCenterID,
			}})

			if tc.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
			}

			if !tc.expectedGetTrustCenterByID {
				olMock.AssertNotCalled(t, "GetTrustCenterByID")
			}

			if !tc.expectUpdateHostname {
				pirschMock.AssertNotCalled(t, "UpdateHostname")
			}

			if !tc.expectUpdateSubdomain {
				pirschMock.AssertNotCalled(t, "UpdateSubdomain")
			}
		})
	}
}
