package graphapi

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/samber/do/v2"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/pkg/gala"
)

var identityResolutionTestQueueSeq atomic.Uint64

// IdentityResolutionTestSetup contains the gala runtime for identity resolution integration tests
type IdentityResolutionTestSetup struct {
	Runtime *gala.Gala
}

// Teardown stops workers and releases connections
func (s *IdentityResolutionTestSetup) Teardown() {
	if s == nil || s.Runtime == nil {
		return
	}

	_ = s.Runtime.StopWorkers(context.Background())
	_ = s.Runtime.Close()
}

// SetupIdentityResolution creates a gala runtime with identity resolution listeners
// wired to the shared ent client for integration tests. Call Teardown on the returned
// setup when the test completes.
func SetupIdentityResolution(ctx context.Context, client *generated.Client, connectionURI string) (*IdentityResolutionTestSetup, error) {
	if client == nil {
		return nil, ErrClientRequired
	}

	if connectionURI == "" {
		return nil, ErrConnectionURIRequired
	}

	queueName := fmt.Sprintf("identity_resolution_test_%d", identityResolutionTestQueueSeq.Add(1))

	runtime, err := gala.NewGala(ctx, gala.Config{
		DispatchMode:      gala.DispatchModeDurable,
		ConnectionURI:     connectionURI,
		QueueName:         queueName,
		WorkerCount:       defaultWorkerCount,
		RunMigrations:     true,
		FetchCooldown:     time.Millisecond,
		FetchPollInterval: defaultFetchPoll,
	})
	if err != nil {
		return nil, err
	}

	client.Use(hooks.EmitGalaEventHook(func() *gala.Gala {
		return runtime
	}))

	if _, err := hooks.RegisterGalaIdentityResolutionListeners(runtime.Registry()); err != nil {
		_ = runtime.Close()

		return nil, err
	}

	do.ProvideValue(runtime.Injector(), client)

	if err := runtime.StartWorkers(ctx); err != nil {
		_ = runtime.Close()

		return nil, err
	}

	return &IdentityResolutionTestSetup{
		Runtime: runtime,
	}, nil
}

// WaitForIdentityHolderLink polls until the directory account has an identity_holder_id set or times out
func WaitForIdentityHolderLink(ctx context.Context, client *generated.Client, accountID string) (*generated.DirectoryAccount, error) {
	return WaitForIdentityHolderLinkWithTimeout(ctx, client, accountID, defaultPollTimeout)
}

// WaitForIdentityHolderLinkWithTimeout polls until the directory account has an identity_holder_id set or times out
func WaitForIdentityHolderLinkWithTimeout(ctx context.Context, client *generated.Client, accountID string, timeout time.Duration) (*generated.DirectoryAccount, error) {
	return pollUntil(ctx, timeout,
		func() (*generated.DirectoryAccount, error) {
			return client.DirectoryAccount.Query().Where(directoryaccount.IDEQ(accountID)).Only(ctx)
		},
		func(account *generated.DirectoryAccount) bool {
			return account.IdentityHolderID != nil && *account.IdentityHolderID != ""
		},
	)
}
