package graphapi

import (
	"context"
	"errors"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/core/v2/internal/graphapi/model"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/objectstore"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/logx"
)

var (
	// ErrDemoDataUnavailable is returned when demo data cannot be managed on this deployment
	ErrDemoDataUnavailable = errors.New("demo data is not available right now, please try again later")

	// ErrDemoDataSeedFailed is returned when the seed could not be queued
	ErrDemoDataSeedFailed = errors.New("unable to add demo data, please try again")

	// ErrDemoDataDeleteFailed is returned when the removal could not be queued
	ErrDemoDataDeleteFailed = errors.New("unable to remove demo data, please try again")

	// ErrDemoDatasetNotFound is returned when the requested demo dataset does not exist
	ErrDemoDatasetNotFound = errors.New("that demo dataset does not exist")
)

// seedDemoData queues the demo dataset seed for the current organization when the caller has full access to it
func (r *mutationResolver) seedDemoData(ctx context.Context, input model.SeedDemoDataInput) (*model.SeedDemoDataPayload, error) {
	if !rule.HasCurrentOrgFullAccess(ctx) {
		return nil, generated.ErrPermissionDenied
	}

	rt, err := objectstoreRuntime(ctx, ErrDemoDataUnavailable)
	if err != nil {
		return nil, err
	}

	req, err := objectstore.SeedDemoRequest(ctx, objectstore.DemoDataset(input.Dataset), lo.FromPtr(input.Profile))
	if err != nil {
		return nil, demoError(ctx, err, ErrDemoDataSeedFailed)
	}

	if _, err := rt.Dispatch(ctx, req); err != nil {
		return nil, demoError(ctx, err, ErrDemoDataSeedFailed)
	}

	return &model.SeedDemoDataPayload{Accepted: true}, nil
}

// deleteDemoData queues the removal of a seeded demo dataset from the current organization when the caller has full access to it
func (r *mutationResolver) deleteDemoData(ctx context.Context, input model.DeleteDemoDataInput) (*model.DeleteDemoDataPayload, error) {
	if !rule.HasCurrentOrgFullAccess(ctx) {
		return nil, generated.ErrPermissionDenied
	}

	rt, err := objectstoreRuntime(ctx, ErrDemoDataUnavailable)
	if err != nil {
		return nil, err
	}

	req, err := objectstore.DeleteDemoRequest(ctx, objectstore.DemoDataset(input.Dataset), input.Keep)
	if err != nil {
		return nil, demoError(ctx, err, ErrDemoDataDeleteFailed)
	}

	if _, err := rt.Dispatch(ctx, req); err != nil {
		return nil, demoError(ctx, err, ErrDemoDataDeleteFailed)
	}

	return &model.DeleteDemoDataPayload{Accepted: true}, nil
}

// objectstoreRuntime returns the integrations runtime when it has an objectstore runtime client, so a deployment without the
// runtime bucket fails before queueing
func objectstoreRuntime(ctx context.Context, returnErr error) (*intruntime.Runtime, error) {
	rt, _, ok := runtimeClient(objectstore.DefinitionID.ID())
	if !ok {
		logx.FromContext(ctx).Error().Str("definition_id", objectstore.DefinitionID.ID()).Msg("integrations runtime or objectstore runtime client is not configured")
		return nil, returnErr
	}

	return rt, nil
}

// demoError maps an objectstore demo error to the message returned to the caller, logging causes the caller cannot act on
func demoError(ctx context.Context, err, failErr error) error {
	switch {
	case errors.Is(err, objectstore.ErrDemoDatasetInvalid):
		return ErrDemoDatasetNotFound
	case errors.Is(err, objectstore.ErrDemoProfileUnknown):
		return err
	default:
		logx.FromContext(ctx).Error().Err(err).Msg("demo data: failed to queue operation")
		return failErr
	}
}
