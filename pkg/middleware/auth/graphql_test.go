package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/ent/privacy/token"
)

func TestBlockNonTrustCenterAnonymous(t *testing.T) {
	next := func(c echox.Context) error {
		c.Response().WriteHeader(http.StatusOK)
		return nil
	}

	makeCtx := func(caller *auth.Caller) echox.Context {
		e := echox.New()
		req := httptest.NewRequest(http.MethodPost, "/query", nil)
		if caller != nil {
			req = req.WithContext(auth.WithCaller(req.Context(), caller))
		}
		return e.NewContext(req, httptest.NewRecorder())
	}

	mw := BlockNonTrustCenterAnonymous()(next)

	t.Run("questionnaire anon caller is blocked", func(t *testing.T) {
		caller := auth.NewQuestionnaireCaller("org1", "anon_questionnaire_abc", "Anon", "")
		c := makeCtx(caller)
		err := mw(c)
		assert.Check(t, err != nil)
		he, ok := err.(*echox.HTTPError)
		assert.Check(t, ok)
		assert.Check(t, he.Code == http.StatusUnauthorized)
	})

	t.Run("trust center anon caller is allowed", func(t *testing.T) {
		caller := auth.NewTrustCenterCaller("org1", "anon_tc_abc", "Anon", "")
		c := makeCtx(caller)
		err := mw(c)
		assert.NilError(t, err)
	})

	t.Run("regular authenticated user is allowed", func(t *testing.T) {
		caller := &auth.Caller{OrganizationID: "org1", SubjectID: "user1"}
		c := makeCtx(caller)
		err := mw(c)
		assert.NilError(t, err)
	})

	t.Run("privacy token with no caller passes middleware", func(t *testing.T) {
		e := echox.New()
		req := httptest.NewRequest(http.MethodPost, "/query", nil)
		req = req.WithContext(token.NewContextWithSignUpToken(req.Context(), "test@example.com"))
		c := e.NewContext(req, httptest.NewRecorder())
		err := mw(c)
		assert.NilError(t, err)
	})
}
