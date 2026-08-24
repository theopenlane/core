package graphapihistory

import (
	"github.com/theopenlane/core/internal/graphapi/directives"
	gqlhistorygenerated "github.com/theopenlane/core/internal/graphapi/historygenerated"
)

// ImplementAllHistoryDirectives adds all active directives to the gqlgen config in the resolver
// setup for the history api
func ImplementAllHistoryDirectives(cfg *gqlhistorygenerated.Config) {
	cfg.Directives.Hidden = directives.HiddenDirective
	cfg.Directives.ReadOnly = directives.ReadOnlyDirective
	cfg.Directives.ExternalReadOnly = directives.ExternalReadOnlyDirective
	cfg.Directives.ExternalSource = directives.ExternalSourceDirective
}
