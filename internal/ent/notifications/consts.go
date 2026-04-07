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
	standardURLPath  = "standards/%s"
	evidenceURLPath  = "evidence?id=%s"
	trustCenterNDA   = "trust-center/NDAs"
)

// getURLPathForObject constructs the URL path for a given object type and ID to be used in notifications, allowing users to navigate directly to the relevant page in the console
func getURLPathForObject(objectID, objectType string) string {
	switch objectType {
	case generated.TypeInternalPolicy:
		return fmt.Sprintf(policyURLPath, objectID)
	case generated.TypeProcedure:
		return fmt.Sprintf(procedureURLPath, objectID)
	case generated.TypeRisk:
		return fmt.Sprintf(riskURLPath, objectID)
	case generated.TypeTask:
		return fmt.Sprintf(taskURLPath, objectID)
	case generated.TypeControl:
		return fmt.Sprintf(controlURLPath, objectID)
	case generated.TypeStandard:
		return fmt.Sprintf(standardURLPath, objectID)
	case generated.TypeEvidence:
		return fmt.Sprintf(evidenceURLPath, objectID)
	case generated.TypeTrustCenterNDARequest:
		return trustCenterNDA
	}

	return ""
}
