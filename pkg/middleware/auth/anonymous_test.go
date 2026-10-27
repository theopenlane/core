package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"
)

func okHandler(c echox.Context) error {
	c.Response().WriteHeader(http.StatusOK)
	return nil
}

func makeCtx(caller *auth.Caller) echox.Context {
	e := echox.New()
	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	if caller != nil {
		req = req.WithContext(auth.WithCaller(req.Context(), caller))
	}
	return e.NewContext(req, httptest.NewRecorder())
}

func TestBlockAnonymous(t *testing.T) {
	questionnaireCaller := auth.NewQuestionnaireCaller("org1", "anon_questionnaire_abc", "Anon", "")
	trustCenterCaller := auth.NewTrustCenterCaller("org1", "anon_tc_abc", "Anon", "")
	userCaller := &auth.Caller{OrganizationID: "org1", SubjectID: "user1"}

	testCases := []struct {
		name        string
		allowed     []auth.Capability
		caller      *auth.Caller
		wantBlocked bool
	}{
		{name: "questionnaire anon caller blocked by default", caller: questionnaireCaller, wantBlocked: true},
		{name: "trust center anon caller blocked by default", caller: trustCenterCaller, wantBlocked: true},
		{name: "regular user allowed by default", caller: userCaller},
		{name: "no caller passes", caller: nil},
		{name: "questionnaire anon caller allowed when opted in", allowed: []auth.Capability{auth.CapQuestionnaireAnonymous}, caller: questionnaireCaller},
		{name: "trust center anon caller blocked on questionnaire route", allowed: []auth.Capability{auth.CapQuestionnaireAnonymous}, caller: trustCenterCaller, wantBlocked: true},
		{name: "regular user allowed on questionnaire route", allowed: []auth.Capability{auth.CapQuestionnaireAnonymous}, caller: userCaller},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := BlockAnonymous(tc.allowed...)(okHandler)(makeCtx(tc.caller))
			if !tc.wantBlocked {
				assert.NilError(t, err)
				return
			}

			assert.Assert(t, err != nil)
			he, ok := err.(*echox.HTTPError)
			assert.Check(t, ok)
			assert.Check(t, he.Code == http.StatusUnauthorized)
		})
	}
}
