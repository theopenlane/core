package interceptors

import (
	"entgo.io/ent/dialect/sql"

	"github.com/theopenlane/core/v2/internal/ent/generated/note"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenterfaq"
)

// AnonScopeNote limits anonymous trust center callers to the posts and faq notes of the active trust center
func AnonScopeNote(trustCenterID string) func(*sql.Selector) {
	return note.Or(
		note.TrustCenterID(trustCenterID),
		note.HasTrustCenterFaqsWith(trustcenterfaq.TrustCenterID(trustCenterID)),
	)
}
