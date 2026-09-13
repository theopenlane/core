package githubapp

import (
	"context"
	"github.com/samber/lo"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/httpsling"
)

// graphQLQueryMarker values identify which directory sync query a GraphQL request body carries
const (
	graphQLViewerRepositoriesMarker = "repositories("
	graphQLExternalIdentitiesMarker = "externalIdentities("
	graphQLMembersWithRoleMarker    = "membersWithRole("
	graphQLTeamsMarker              = "teams("
)

// graphQL response bodies used to drive the directory sync handler
const (
	viewerOrganizationsResponse    = `{"data":{"viewer":{"repositories":{"nodes":[{"owner":{"login":"acme","__typename":"Organization"}}],"pageInfo":{"endCursor":"","hasNextPage":false}}}}}`
	externalIdentitiesResponse     = `{"data":{"organization":{"samlIdentityProvider":null}}}`
	organizationMembersResponse    = `{"data":{"organization":{"membersWithRole":{"nodes":[{"databaseId":1,"login":"octocat","name":"Octo Cat","email":"octo@example.com","avatarUrl":"","organizationVerifiedDomainEmails":[]}],"pageInfo":{"endCursor":"","hasNextPage":false}}}}}`
	organizationTeamsResponse      = `{"data":{"organization":{"teams":{"nodes":[{"databaseId":2,"name":"core","slug":"core","description":"","privacy":"CLOSED","members":{"edges":[{"node":{"databaseId":1,"login":"octocat"},"role":"MEMBER"}],"pageInfo":{"endCursor":"","hasNextPage":false}}}],"pageInfo":{"endCursor":"","hasNextPage":false}}}}}`
	organizationTeamsErrorResponse = `{"errors":[{"message":"teams unavailable"}]}`
)

// newDirectorySyncGraphQLServer serves canned GraphQL responses for each directory sync query, returning the given teams response
func newDirectorySyncGraphQLServer(t *testing.T, teamsResponse string) GraphQLClient {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		assert.NilError(t, err)

		query := string(body)

		w.Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

		switch {
		case strings.Contains(query, graphQLTeamsMarker):
			_, _ = w.Write([]byte(teamsResponse))
		case strings.Contains(query, graphQLMembersWithRoleMarker):
			_, _ = w.Write([]byte(organizationMembersResponse))
		case strings.Contains(query, graphQLExternalIdentitiesMarker):
			_, _ = w.Write([]byte(externalIdentitiesResponse))
		case strings.Contains(query, graphQLViewerRepositoriesMarker):
			_, _ = w.Write([]byte(viewerOrganizationsResponse))
		default:
			http.NotFound(w, req)
		}
	}))
	t.Cleanup(server.Close)

	client, err := newGraphQLClient(server.Client(), server.URL)
	assert.NilError(t, err)

	return client
}

// TestDirectorySyncSnapshotComplete verifies account, group, and membership payload sets carry the expected completeness flags
func TestDirectorySyncSnapshotComplete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		teamsResponse string
		cfg           DirectorySync
		want          map[string]bool
	}{
		{
			name:          "teams query succeeds marks all sets complete",
			teamsResponse: organizationTeamsResponse,
			want: map[string]bool{
				entityops.SchemaDirectoryAccount.Name:    true,
				entityops.SchemaDirectoryGroup.Name:      true,
				entityops.SchemaDirectoryMembership.Name: true,
			},
		},
		{
			name:          "teams query fails marks group and membership sets incomplete",
			teamsResponse: organizationTeamsErrorResponse,
			want: map[string]bool{
				entityops.SchemaDirectoryAccount.Name:    true,
				entityops.SchemaDirectoryGroup.Name:      false,
				entityops.SchemaDirectoryMembership.Name: false,
			},
		},
		{
			name:          "group sync disabled emits only a complete account set",
			teamsResponse: organizationTeamsErrorResponse,
			cfg:           DirectorySync{DisableGroupSync: true},
			want: map[string]bool{
				entityops.SchemaDirectoryAccount.Name: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := newDirectorySyncGraphQLServer(t, tc.teamsResponse)

			sets, err := tc.cfg.Run(context.Background(), client)
			assert.NilError(t, err)
			assert.DeepEqual(t, lo.SliceToMap(sets, func(set types.IngestPayloadSet) (string, bool) { return set.Schema, set.SnapshotComplete }), tc.want)
		})
	}
}
