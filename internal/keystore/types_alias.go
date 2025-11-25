package keystore

import (
	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/types"
)

type (
	AuthType                   = types.AuthKind
	GoogleWorkloadIdentitySpec = config.GoogleWorkloadIdentitySpec
	GitHubAppSpec              = config.GitHubAppSpec
	PersistenceSpec            = config.PersistenceSpec
	ClientDescriptor           = types.ClientDescriptor
	ClientName                 = types.ClientName
	OperationDescriptor        = types.OperationDescriptor
	OperationName              = types.OperationName
	OperationRequest           = types.OperationRequest
	OperationResult            = types.OperationResult
)
