package scim

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Spec is the provider spec for the SCIM 2.0 push-based integration
var Spec = spec.ProviderSpec{
	Name:        "scim",
	DisplayName: "SCIM 2.0",
	Category:    "identity",
	AuthType:    types.AuthKindNone,
	Active:      lo.ToPtr(true),
	Visible:     lo.ToPtr(true),
	DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/scim/overview",
	Tags:        []string{"scim", "provisioning", "identity"},
	Labels:      map[string]string{"protocol": "scim2"},
}
