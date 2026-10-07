package graphapi_test

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
)

func TestMutationCreateEvidenceFileProvenance(t *testing.T) {
	artifactSHA256 := th.GetSHA256Hash(t, th.LogoFilePath)

	testCases := []struct {
		name          string
		claims        map[string]any
		client        *testclient.TestClient
		ctx           context.Context
		ownerID       *string
		wantCreatedBy string
		expectedErr   string
	}{
		{
			name:          "matching hash is stored and attributed to the user",
			claims:        map[string]any{"schema_version": 1, "artifact_sha256": artifactSHA256},
			client:        suite.Client.API,
			ctx:           th.SharedTestUser1.UserCtx,
			wantCreatedBy: th.SharedTestUser1.ID,
		},
		{
			name:          "personal access token upload is attributed to the token owner",
			claims:        map[string]any{"schema_version": 1, "artifact_sha256": artifactSHA256},
			client:        suite.Client.APIWithPAT,
			ctx:           context.Background(),
			ownerID:       &th.SharedTestUser1.OrganizationID,
			wantCreatedBy: th.SharedTestUser1.ID,
		},
		{
			name:   "api token upload without a claimed hash is stored",
			claims: map[string]any{"schema_version": 1},
			client: suite.Client.APIWithToken,
			ctx:    context.Background(),
		},
		{
			name:        "mismatched hash fails the upload",
			claims:      map[string]any{"artifact_sha256": strings.Repeat("0", len(artifactSHA256))},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "does not match the uploaded file",
		},
		{
			name:        "malformed hash fails the upload",
			claims:      map[string]any{"artifact_sha256": "not-a-hash"},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "must be a 64 character hex-encoded SHA-256",
		},
		{
			name:          "upload without provenance leaves it null",
			client:        suite.Client.API,
			ctx:           th.SharedTestUser1.UserCtx,
			wantCreatedBy: th.SharedTestUser1.ID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			upload := th.UploadFile(t, th.LogoFilePath)

			if tc.expectedErr == "" {
				th.ExpectUploadNillable(t, suite.Client.MockProvider, []*graphql.Upload{upload})
			}

			var metadata []*testclient.FileMetadataInput
			if tc.claims != nil {
				metadata = []*testclient.FileMetadataInput{{Provenance: tc.claims}}
			}

			resp, err := tc.client.CreateEvidenceWithFileProvenance(tc.ctx,
				testclient.CreateEvidenceInput{Name: "Provenance Evidence", OwnerID: tc.ownerID},
				[]*graphql.Upload{upload}, metadata)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			defer (&th.Cleanup[*generated.EvidenceDeleteOne]{Client: suite.Client.DB.Evidence, ID: resp.CreateEvidence.Evidence.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)

			assert.Assert(t, is.Len(resp.CreateEvidence.Evidence.Files.Edges, 1))

			node := resp.CreateEvidence.Evidence.Files.Edges[0].Node
			assert.Check(t, is.Equal(th.GetMD5Hash(t, th.LogoFilePath), lo.FromPtr(node.Md5Hash)))
			assert.Check(t, is.Equal(artifactSHA256, lo.FromPtr(node.Sha256Hash)))
			assert.Check(t, lo.FromPtr(node.CreatedBy) != "")

			if tc.wantCreatedBy != "" {
				assert.Check(t, is.Equal(tc.wantCreatedBy, lo.FromPtr(node.CreatedBy)))
			}

			if tc.claims == nil {
				assert.Check(t, node.Provenance == nil)

				return
			}

			assert.Check(t, is.Equal(float64(1), node.Provenance["schema_version"]))

			if claimed, ok := tc.claims["artifact_sha256"]; ok {
				assert.Check(t, is.Equal(claimed, node.Provenance["artifact_sha256"]))
			}
		})
	}
}

func TestFileProvenanceIsImmutable(t *testing.T) {
	for _, builder := range []reflect.Type{
		reflect.TypeFor[*generated.FileUpdate](),
		reflect.TypeFor[*generated.FileUpdateOne](),
	} {
		for _, method := range []string{"SetProvenance", "ClearProvenance", "SetSha256Hash", "ClearSha256Hash"} {
			_, ok := builder.MethodByName(method)
			assert.Check(t, !ok, "%s must not expose %s", builder, method)
		}
	}

	for _, field := range []string{"Provenance", "ClearProvenance", "Sha256Hash", "ClearSha256Hash"} {
		_, ok := reflect.TypeFor[generated.UpdateFileInput]().FieldByName(field)
		assert.Check(t, !ok, "UpdateFileInput must not carry %s", field)
	}

	_, ok := reflect.TypeFor[generated.CreateFileInput]().FieldByName("Provenance")
	assert.Check(t, !ok, "CreateFileInput must not carry Provenance")
}
