package corejobs_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/riverqueue/river"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/eddy"

	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/openlaneclient"

	olmocks "github.com/theopenlane/core/pkg/corejobs/internal/olclient/mocks"
)

// createMinimalStorageService creates a minimal storage service for testing
func createMinimalStorageService() *objects.Service {
	pool := eddy.NewClientPool[storage.Provider](time.Minute)
	clientService := eddy.NewClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions](pool)
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	return objects.NewService(objects.Config{
		Resolver:      resolver,
		ClientService: clientService,
	})
}

func TestCacheTrustCenterDataWorker_NoStorageService(t *testing.T) {
	ctx := context.Background()
	olMock := olmocks.NewMockOpenlaneGraphClient(t)

	worker := &corejobs.CacheTrustCenterDataWorker{}
	worker.WithOpenlaneClient(olMock)

	err := worker.Work(ctx, &river.Job[corejobs.CacheTrustCenterDataArgs]{})

	require.Error(t, err)
	require.Contains(t, err.Error(), "storage service is required")
}

func TestCacheTrustCenterDataWorker_OpenlaneClientError(t *testing.T) {
	ctx := context.Background()
	olMock := olmocks.NewMockOpenlaneGraphClient(t)

	olMock.EXPECT().GetTrustCentersCacheData(
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(nil, errors.New("api error"))

	worker := &corejobs.CacheTrustCenterDataWorker{}
	worker.WithOpenlaneClient(olMock)
	worker.WithStorageService(createMinimalStorageService())

	err := worker.Work(ctx, &river.Job[corejobs.CacheTrustCenterDataArgs]{})

	require.Error(t, err)
	require.Contains(t, err.Error(), "api error")
}

func TestCacheTrustCenterDataWorker_NoTrustCenters(t *testing.T) {
	ctx := context.Background()
	olMock := olmocks.NewMockOpenlaneGraphClient(t)

	response := &openlaneclient.GetTrustCentersCacheData{
		TrustCenters: openlaneclient.GetTrustCentersCacheData_TrustCenters{
			Edges: nil,
			PageInfo: openlaneclient.GetTrustCentersCacheData_TrustCenters_PageInfo{
				HasNextPage: false,
			},
		},
	}

	olMock.EXPECT().GetTrustCentersCacheData(
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(response, nil)

	worker := &corejobs.CacheTrustCenterDataWorker{}
	worker.WithOpenlaneClient(olMock)
	worker.WithStorageService(createMinimalStorageService())

	err := worker.Work(ctx, &river.Job[corejobs.CacheTrustCenterDataArgs]{})

	require.NoError(t, err)
}

func TestCacheTrustCenterDataWorker_TrustCenterWithoutCustomDomain(t *testing.T) {
	ctx := context.Background()
	olMock := olmocks.NewMockOpenlaneGraphClient(t)

	response := &openlaneclient.GetTrustCentersCacheData{
		TrustCenters: openlaneclient.GetTrustCentersCacheData_TrustCenters{
			Edges: []*openlaneclient.GetTrustCentersCacheData_TrustCenters_Edges{
				{
					Node: &openlaneclient.GetTrustCentersCacheData_TrustCenters_Edges_Node{
						CustomDomain: nil,
						Setting:      &openlaneclient.TrustCenterSettingFields{},
					},
				},
			},
			PageInfo: openlaneclient.GetTrustCentersCacheData_TrustCenters_PageInfo{
				HasNextPage: false,
			},
		},
	}

	olMock.EXPECT().GetTrustCentersCacheData(
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(response, nil)

	worker := &corejobs.CacheTrustCenterDataWorker{}
	worker.WithOpenlaneClient(olMock)
	worker.WithStorageService(createMinimalStorageService())

	err := worker.Work(ctx, &river.Job[corejobs.CacheTrustCenterDataArgs]{})

	require.NoError(t, err)
}

func TestCacheTrustCenterDataWorker_EmptyCustomDomainCname(t *testing.T) {
	ctx := context.Background()
	olMock := olmocks.NewMockOpenlaneGraphClient(t)

	response := &openlaneclient.GetTrustCentersCacheData{
		TrustCenters: openlaneclient.GetTrustCentersCacheData_TrustCenters{
			Edges: []*openlaneclient.GetTrustCentersCacheData_TrustCenters_Edges{
				{
					Node: &openlaneclient.GetTrustCentersCacheData_TrustCenters_Edges_Node{
						CustomDomain: &openlaneclient.GetTrustCentersCacheData_TrustCenters_Edges_Node_CustomDomain{
							CnameRecord: "",
						},
						Setting: &openlaneclient.TrustCenterSettingFields{},
					},
				},
			},
			PageInfo: openlaneclient.GetTrustCentersCacheData_TrustCenters_PageInfo{
				HasNextPage: false,
			},
		},
	}

	olMock.EXPECT().GetTrustCentersCacheData(
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(response, nil)

	worker := &corejobs.CacheTrustCenterDataWorker{}
	worker.WithOpenlaneClient(olMock)
	worker.WithStorageService(createMinimalStorageService())

	err := worker.Work(ctx, &river.Job[corejobs.CacheTrustCenterDataArgs]{})

	require.NoError(t, err)
}

func TestCacheTrustCenterDataWorker_NoOrganizationID(t *testing.T) {
	ctx := context.Background()
	olMock := olmocks.NewMockOpenlaneGraphClient(t)

	response := &openlaneclient.GetTrustCentersCacheData{
		TrustCenters: openlaneclient.GetTrustCentersCacheData_TrustCenters{
			Edges: []*openlaneclient.GetTrustCentersCacheData_TrustCenters_Edges{
				{
					Node: &openlaneclient.GetTrustCentersCacheData_TrustCenters_Edges_Node{
						ID: "trust-center-1",
						CustomDomain: &openlaneclient.GetTrustCentersCacheData_TrustCenters_Edges_Node_CustomDomain{
							CnameRecord: "example.com",
						},
						Setting: &openlaneclient.TrustCenterSettingFields{},
						OwnerID: nil, // No organization ID
					},
				},
			},
			PageInfo: openlaneclient.GetTrustCentersCacheData_TrustCenters_PageInfo{
				HasNextPage: false,
			},
		},
	}

	olMock.EXPECT().GetTrustCentersCacheData(
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(response, nil)

	worker := &corejobs.CacheTrustCenterDataWorker{}
	worker.WithOpenlaneClient(olMock)
	worker.WithStorageService(createMinimalStorageService())

	err := worker.Work(ctx, &river.Job[corejobs.CacheTrustCenterDataArgs]{})

	require.NoError(t, err)
}
