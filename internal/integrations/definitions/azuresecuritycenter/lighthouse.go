package azuresecuritycenter

import "github.com/theopenlane/core/v2/internal/integrations/definitions/microsoft"

// lighthouseOffer is the read-only Defender for Cloud delegation the Lighthouse template grants Openlane
var lighthouseOffer = microsoft.LighthouseOffer{
	NamePrefix:           "Openlane Defender for Cloud",
	Description:          "Read-only Microsoft Defender for Cloud access for Openlane",
	PrincipalDisplayName: "Openlane",
	RoleDefinitionIDs:    []string{microsoft.SecurityReaderRoleID, microsoft.RegistrationAssignmentDeleteRoleID},
}
