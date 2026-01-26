package notifications

import (
	"fmt"

	"github.com/theopenlane/core/internal/ent/generated"
)

const (
	policyURLPath    = "policies/%s/view"
	procedureURLPath = "procedures/%s/view"
	riskURLPath      = "risks/%s"
	taskURLPath      = "tasks?id=%s"
	controlURLPath   = "controls/%s"
	evidenceURLPath  = "evidence?id=%s"
	trustCenterNDA   = "trust-center/NDAs"
)

// getURLPathForObject constructs the URL path for a given object type and ID
func getURLPathForObject(base, objectID, objectType string) string {
	switch objectType {
	case generated.TypeInternalPolicy:
		return base + fmt.Sprintf(policyURLPath, objectID)
	case generated.TypeProcedure:
		return base + fmt.Sprintf(procedureURLPath, objectID)
	case generated.TypeRisk:
		return base + fmt.Sprintf(riskURLPath, objectID)
	case generated.TypeTask:
		return base + fmt.Sprintf(taskURLPath, objectID)
	case generated.TypeControl:
		return base + fmt.Sprintf(controlURLPath, objectID)
	case generated.TypeEvidence:
		return base + fmt.Sprintf(evidenceURLPath, objectID)
	case generated.TypeTrustCenterNDARequest:
		return base + trustCenterNDA
	}

	return ""
}
