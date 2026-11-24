package aikido

import (
	"context"
	"net/http"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
)

func TestCollectAutofixes(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case pathAutofixHistory:
			writeJSON(w, `[{"id":12345,"type":"dependency","status":"pr_merged","autofix_target":{"code_repo_id":1,"title":"lodash → 4.17.21"}},{"id":12346,"type":"sast","status":"preview_available"},{"id":12347,"type":"sast","status":"pr_closed"}]`)
		case pathAutofixHistory + "/12345/issues":
			writeJSON(w, `{"targeted_issues":[{"id":1},{"id":2}]}`)
		case pathAutofixHistory + "/12346/issues", pathAutofixHistory + "/12347/issues":
			t.Errorf("issues requested for an autofix that was not merged: %s", r.URL.Path)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})

	payloads, err := CollectAutofixes(context.Background(), client)
	assert.NilError(t, err)

	envelopes := payloadSet(payloads, entityops.SchemaRemediation.Name)
	assert.Assert(t, is.Len(envelopes, 1))

	assert.Check(t, is.Equal("lodash → 4.17.21", envelopes[0].Resource))
	assert.Check(t, is.Contains(string(envelopes[0].Payload), `"targeted_issue_ids":["1","2"]`))
}

func TestCollectAutofixesPropagatesIssueErrors(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == pathAutofixHistory {
			writeJSON(w, `[{"id":7,"status":"pr_merged"}]`)

			return
		}

		w.WriteHeader(http.StatusNotFound)
	})

	_, err := CollectAutofixes(context.Background(), client)
	assert.Check(t, is.ErrorIs(err, ErrAutofixFetchFailed))
}
