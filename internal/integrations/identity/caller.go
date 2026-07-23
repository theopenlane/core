package identity

import (
	"context"

	"github.com/theopenlane/iam/auth"

	ent "github.com/theopenlane/core/internal/ent/generated"
)

// Config contains the virtual integration user values
type Config struct {
	// SubjectID is the static user ID used for db records/objects created by an integration.
	SubjectID string `json:"subjectid" koanf:"subjectid" default:"01JNTGACTR0000000000000000"`
	// Email is the email used for db objects created by an integration
	Email string `json:"email" koanf:"email" default:"integrations@theopenlane.io"`
	// DisplayName is the display name used for db objects created by an integration
	DisplayName string `json:"displayname" koanf:"displayname" default:"Openlane Integrations"`
}

// NewIntegrationCaller returns the virtual service user for an integration installation.
func NewIntegrationCaller(integration *ent.Integration, cfg Config) *auth.Caller {
	if integration == nil {
		return nil
	}

	var orgIDs []string
	if integration.OwnerID != "" {
		orgIDs = []string{integration.OwnerID}
	}

	return &auth.Caller{
		SubjectID:          cfg.SubjectID,
		SubjectName:        cfg.DisplayName,
		SubjectEmail:       cfg.Email,
		OrganizationID:     integration.OwnerID,
		OrganizationIDs:    orgIDs,
		AuthenticationType: auth.APITokenAuthentication,
		Capabilities:       auth.CapIntegrationActor,
	}
}

// WithIntegrationCaller stores the integration user in the context
func WithIntegrationCaller(ctx context.Context, integration *ent.Integration, cfg Config) context.Context {
	caller := NewIntegrationCaller(integration, cfg)
	if caller == nil {
		return ctx
	}

	return auth.WithCaller(ctx, caller)
}
