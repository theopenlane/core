package okta

import (
	oktagosdk "github.com/okta/okta-sdk-golang/v5/okta"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID             = types.NewDefinitionRef("def_01K0OKTA0000000000000000001")
	OktaClient               = types.NewClientRef[*oktagosdk.APIClient]()
	HealthDefaultOperation   = types.NewOperationRef[HealthCheck]("health.default")
	PoliciesCollectOperation = types.NewOperationRef[PoliciesCollect]("policies.collect")
)

const Slug = "okta"
