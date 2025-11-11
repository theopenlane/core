package keystore

import (
	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/types"
)

type (
	AuthType            = types.AuthKind
	WorkloadIdentitySpec = config.WorkloadIdentitySpec
	GitHubAppSpec        = config.GitHubAppSpec
	PersistenceSpec      = config.PersistenceSpec
)
