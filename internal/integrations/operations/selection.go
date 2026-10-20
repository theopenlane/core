package operations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// ResourceSelection holds the provider resources a user selected for an installation
type ResourceSelection struct {
	// Resources lists the selected provider resources, replacing any previous selection
	Resources []types.SelectedResource `json:"resources" jsonschema:"required,title=Resources"`
}

// SelectionStore saves the resources a user selected onto the installation field the definition reads them from
type SelectionStore func(ctx context.Context, request types.OperationRequest, resources []types.SelectedResource) error

// NewResourceSelectionOperation returns the inline operation ref that saves an installation's selected resources through store, accepting at most limit resources when limit is above zero
func NewResourceSelectionOperation(store SelectionStore, limit int) types.OperationRef[ResourceSelection] {
	return types.OperationPayloadOf[ResourceSelection]().
		HandlesRequest(resourceSelection{store: store, limit: limit}.run).
		Policy(types.ExecutionPolicy{Inline: true})
}

// resourceSelection binds a definition's store and selection limit to the shared selection handler
type resourceSelection struct {
	// store saves the selection onto the installation
	store SelectionStore
	// limit caps how many resources may be selected, zero for no cap
	limit int
}

// run validates the submitted resources and saves them through the definition's store
func (s resourceSelection) run(ctx context.Context, request types.OperationRequest, cfg ResourceSelection) (json.RawMessage, error) {
	resources := lo.UniqBy(lo.Filter(cfg.Resources, func(r types.SelectedResource, _ int) bool { return r.ID != "" }), selectedResourceID)

	if s.limit > 0 && len(resources) > s.limit {
		return nil, fmt.Errorf("%w: at most %d allowed", ErrSelectionLimitExceeded, s.limit)
	}

	if err := s.store(ctx, request, resources); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSelectionPersistFailed, err)
	}

	return providerkit.EncodeResult(ResourceSelection{Resources: resources}, ErrResultEncode)
}

// selectedResourceID keys a selected resource by its provider ID
func selectedResourceID(r types.SelectedResource) string {
	return r.ID
}
