package handlers

import (
	"net/http"

	"github.com/theopenlane/iam/sessions"
)

// clearAuthFlowCookies expires the named transient auth-flow cookies (SSO/OAuth state, support
// access handoff, integration auth, etc.) using the same session cookie config they were set with.
// Transient flow cookies are set from h.SessionConfig.CookieConfig, so they must be removed with the
// same config: a deletion cookie that omits the original Domain, Secure, or SameSite will not match
// and the browser will keep the cookie. Always route flow-cookie removal through this helper rather
// than building an ad-hoc CookieConfig so set and remove stay consistent.
func (h *Handler) clearAuthFlowCookies(w http.ResponseWriter, names ...string) {
	sessions.RemoveCookies(w, *h.SessionConfig.CookieConfig, names...)
}
