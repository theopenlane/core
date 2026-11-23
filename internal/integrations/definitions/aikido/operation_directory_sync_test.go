package aikido

import (
	"context"
	"net/http"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
)

func TestCollectDirectory(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case pathUsers:
			assert.Check(t, is.Equal("1", r.URL.Query().Get("include_inactive")))

			if r.URL.Query().Get("filter_team_id") == "1" {
				writeJSON(w, `[{"id":5,"email":"wendy@example.com"}]`)

				return
			}

			writeJSON(w, `[{"id":5,"email":"wendy@example.com","active":1},{"id":6,"email":"sam@example.com","active":0}]`)
		case pathTeams:
			writeJSON(w, `[{"id":1,"name":"Frontend developers","active":true}]`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})

	payloads, err := CollectDirectory(context.Background(), client, false)
	assert.NilError(t, err)

	accounts := payloadSet(payloads, entityops.SchemaDirectoryAccount.Name)
	groups := payloadSet(payloads, entityops.SchemaDirectoryGroup.Name)
	memberships := payloadSet(payloads, entityops.SchemaDirectoryMembership.Name)

	assert.Check(t, is.Len(accounts, 2))
	assert.Assert(t, is.Len(groups, 1))
	assert.Assert(t, is.Len(memberships, 1))

	assert.Check(t, is.Contains(string(groups[0].Payload), `"member_count":1`))
	assert.Check(t, is.Equal("1:5", memberships[0].Resource))
	assert.Check(t, is.Equal(`{"team_id":"1","user_id":"5"}`, string(memberships[0].Payload)))
}

func TestCollectDirectorySkipsGroups(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Check(t, is.Equal(pathUsers, r.URL.Path))
		writeJSON(w, `[{"id":5,"email":"wendy@example.com","active":1}]`)
	})

	payloads, err := CollectDirectory(context.Background(), client, true)
	assert.NilError(t, err)

	assert.Check(t, is.Len(payloadSet(payloads, entityops.SchemaDirectoryAccount.Name), 1))
	assert.Check(t, is.Nil(payloadSet(payloads, entityops.SchemaDirectoryGroup.Name)))
}

func TestCollectDirectoryPropagatesTeamErrors(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == pathTeams {
			w.WriteHeader(http.StatusForbidden)

			return
		}

		writeJSON(w, `[]`)
	})

	_, err := CollectDirectory(context.Background(), client, false)
	assert.Check(t, is.ErrorIs(err, ErrDirectoryFetchFailed))
}
