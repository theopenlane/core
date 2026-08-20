package graphapi

import (
	"github.com/theopenlane/core/internal/graphapi/directives"
	gqlgenerated "github.com/theopenlane/core/internal/graphapi/generated"
)

// ImplementAllDirectives adds all active directives to the gqlgen config in the resolver setup
func ImplementAllDirectives(cfg *gqlgenerated.Config) {
	cfg.Directives.Hidden = directives.HiddenDirective
	cfg.Directives.ReadOnly = directives.ReadOnlyDirective
	cfg.Directives.ExternalReadOnly = directives.ExternalReadOnlyDirective
	cfg.Directives.ExternalSource = directives.ExternalSourceDirective
}
