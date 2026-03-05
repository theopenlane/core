package registry

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/types"
)

// ResolveOperation resolves one provider operation descriptor by name and/or kind.
func (r *Registry) ResolveOperation(provider types.ProviderType, operationName types.OperationName, operationKind types.OperationKind) (types.OperationDescriptor, error) {
	if provider == types.ProviderUnknown {
		return types.OperationDescriptor{}, ErrProviderTypeRequired
	}
	if operationName == "" && operationKind == "" {
		return types.OperationDescriptor{}, ErrOperationCriteriaRequired
	}

	descriptors := r.OperationDescriptors(provider)
	if len(descriptors) == 0 {
		return types.OperationDescriptor{}, ErrOperationNotRegistered
	}

	if operationName != "" {
		return resolveOperationByName(descriptors, operationName, operationKind)
	}

	matches := lo.Filter(descriptors, func(descriptor types.OperationDescriptor, _ int) bool {
		return descriptor.Kind == operationKind
	})
	switch len(matches) {
	case 0:
		return types.OperationDescriptor{}, ErrOperationNotRegistered
	case 1:
		return matches[0], nil
	default:
		return types.OperationDescriptor{}, ErrOperationDescriptorAmbiguous
	}
}

func resolveOperationByName(descriptors []types.OperationDescriptor, operationName types.OperationName, operationKind types.OperationKind) (types.OperationDescriptor, error) {
	matches := lo.Filter(descriptors, func(descriptor types.OperationDescriptor, _ int) bool {
		return descriptor.Name == operationName
	})
	switch len(matches) {
	case 0:
		return types.OperationDescriptor{}, ErrOperationNotRegistered
	case 1:
		if operationKind != "" && matches[0].Kind != operationKind {
			return types.OperationDescriptor{}, ErrOperationKindMismatch
		}

		return matches[0], nil
	default:
		return types.OperationDescriptor{}, ErrOperationDescriptorAmbiguous
	}
}
