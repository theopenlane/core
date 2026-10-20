package handlers

import (
	"github.com/google/uuid"
	"github.com/samber/lo"
	echo "github.com/theopenlane/echox"
)

// msftAssociatedApplications contains the associated azure applications with the domain
type msftAssociatedApplications struct {
	AssociatedApplications []application `json:"associatedApplications"`
}

// application contains the application id of a registered application
type application struct {
	// ApplicationID is the application of the registered application in Azure
	ApplicationID string `json:"applicationId"`
}

// MSFTIdentityWellKnownHandler returns the application ids for associated applications in
// order to populate the msft identity well-known handler
//
//	{
//	  "associatedApplications": [
//	    {
//	      "applicationId": "<uuid>"
//	    }
//	  ]
//	}
func (h *Handler) MSFTIdentityWellKnownHandler(ctx echo.Context) error {
	appIDs := []string{
		h.IntegrationsConfig.AzureEntraID.ApplicationID,
		h.IntegrationsConfig.AzureEntraID.Legacy.ApplicationID,
		h.IntegrationsConfig.AzureSecurityCenter.ApplicationID,
		h.IntegrationsConfig.OneDrive.ApplicationID,
		h.IntegrationsConfig.OneDrive.Legacy.ApplicationID,
		h.IntegrationsConfig.MicrosoftTeams.ApplicationID,
	}

	applications := lo.FilterMap(appIDs, func(appID string, _ int) (application, bool) {
		_, err := uuid.Parse(appID)

		return application{ApplicationID: appID}, err == nil
	})

	// ensure we do not have duplicate application ids
	applications = lo.Uniq(applications)

	return h.Success(ctx, msftAssociatedApplications{
		AssociatedApplications: applications,
	})
}
