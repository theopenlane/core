package gcpscc

import (
	cloudscc "cloud.google.com/go/securitycenter/apiv2"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID             = types.NewDefinitionRef("def_01K0GCPSCC00000000000000001")
	SCCClient                = types.NewClientRef[*cloudscc.Client]()
	HealthDefaultOperation   = types.NewOperationRef[HealthCheck]("health.default")
	FindingsCollectOperation = types.NewOperationRef[FindingsCollect]("findings.collect")
	SettingsScanOperation    = types.NewOperationRef[SettingsScan]("settings.scan")
)

const Slug = "gcp_scc"
