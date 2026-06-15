package handlers

import (
	"net/http"

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
func (h *Handler) MSFTIdentityWellKnownHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	entraAppID := h.IntegrationsConfig.AzureEntraID.ApplicationID
	oneDriveAppID := h.IntegrationsConfig.OneDrive.ApplicationID
	teamsAppID := h.IntegrationsConfig.MicrosoftTeams.ApplicationID

	applications := []application{}

	if entraAppID != "" {
		if _, err := uuid.Parse(entraAppID); err == nil {
			applications = append(applications, application{
				ApplicationID: entraAppID,
			})
		}

	}

	if oneDriveAppID != "" {
		if _, err := uuid.Parse(oneDriveAppID); err == nil {
			applications = append(applications, application{
				ApplicationID: oneDriveAppID,
			})
		}

	}

	if teamsAppID != "" {
		if _, err := uuid.Parse(teamsAppID); err == nil {
			applications = append(applications, application{
				ApplicationID: teamsAppID,
			})
		}
	}

	// ensure we do not have duplicate application ids
	applications = lo.Uniq(applications)

	return ctx.JSON(http.StatusOK, msftAssociatedApplications{
		AssociatedApplications: applications,
	})
}
