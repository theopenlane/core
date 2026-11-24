package aikido

import (
	"context"
	"net/http"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
)

func TestCollectScans(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case pathPentests:
			writeJSON(w, `[{"id":"f19e449d-2bf0-8d37-ab9a-3bcbc748157d","name":"Production web app pentest","status":"completed","created_at":1760000000}]`)
		case pathCodeAudits:
			writeJSON(w, `[{"id":"a29e449d-2bf0-8d37-ab9a-3bcbc748157d","name":"Main repository audit","status":"running","created_at":1760000000}]`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})

	payloads, err := CollectScans(context.Background(), client)
	assert.NilError(t, err)

	envelopes := payloadSet(payloads, entityops.SchemaScan.Name)
	assert.Assert(t, is.Len(envelopes, 2))

	assert.Check(t, is.Equal("f19e449d-2bf0-8d37-ab9a-3bcbc748157d", envelopes[0].Resource))
	assert.Check(t, is.Contains(string(envelopes[0].Payload), `"kind":"ai_pentest"`))
	assert.Check(t, is.Contains(string(envelopes[1].Payload), `"kind":"code_audit"`))
}

func TestCollectScansPropagatesErrors(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	_, err := CollectScans(context.Background(), client)
	assert.Check(t, is.ErrorIs(err, ErrScansFetchFailed))
}
