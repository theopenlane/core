//go:build test

package graphapi_test

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
)

const paginatedEvidencesQuery = `query GetEvidences($orderBy: [EvidenceOrder!], $first: Int, $after: Cursor) {
  evidences(orderBy: $orderBy, first: $first, after: $after) {
    pageInfo { endCursor hasNextPage }
    edges { node { id } }
  }
}`

type paginatedEvidencesResponse struct {
	Evidences struct {
		PageInfo testclient.PageInfo                        `json:"pageInfo" graphql:"pageInfo"`
		Edges    []*testclient.GetEvidences_Evidences_Edges `json:"edges" graphql:"edges"`
	} `json:"evidences" graphql:"evidences"`
}

func TestQueryEvidencesPaginationFillsPageAfterFGAFilter(t *testing.T) {
	users := suite.SeedFreshMinimalOrgUsers(t, true)
	ownerCtx := users.Owner.UserCtx

	program := (&th.ProgramBuilder{Client: suite.Client}).MustNew(ownerCtx, t)
	(&th.ProgramMemberBuilder{Client: suite.Client, ProgramID: program.ID, UserID: users.Member.ID, Role: enums.RoleMember.String()}).MustNew(ownerCtx, t)

	const hiddenPerGap = 12

	older := (&th.EvidenceBuilder{Client: suite.Client, ProgramID: program.ID}).MustNew(ownerCtx, t)
	for range hiddenPerGap {
		(&th.EvidenceBuilder{Client: suite.Client}).MustNew(ownerCtx, t)
	}

	newer := (&th.EvidenceBuilder{Client: suite.Client, ProgramID: program.ID}).MustNew(ownerCtx, t)
	for range hiddenPerGap {
		(&th.EvidenceBuilder{Client: suite.Client}).MustNew(ownerCtx, t)
	}

	concreteClient, ok := suite.Client.API.TestGraphClient.(*testclient.Client)
	assert.Assert(t, ok)

	vars := map[string]any{
		"orderBy": []map[string]any{{"field": "created_at", "direction": "DESC"}},
		"first":   1,
	}

	var firstPage paginatedEvidencesResponse
	err := concreteClient.Client.Post(users.Member.UserCtx, "GetEvidences", paginatedEvidencesQuery, &firstPage, vars)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(firstPage.Evidences.Edges, 1))
	assert.Check(t, is.Equal(firstPage.Evidences.Edges[0].Node.ID, newer.ID))
	assert.Check(t, firstPage.Evidences.PageInfo.HasNextPage)
	assert.Assert(t, firstPage.Evidences.PageInfo.EndCursor != nil)

	vars["after"] = *firstPage.Evidences.PageInfo.EndCursor

	var secondPage paginatedEvidencesResponse
	err = concreteClient.Client.Post(users.Member.UserCtx, "GetEvidences", paginatedEvidencesQuery, &secondPage, vars)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(secondPage.Evidences.Edges, 1))
	assert.Check(t, is.Equal(secondPage.Evidences.Edges[0].Node.ID, older.ID))
	assert.Check(t, !secondPage.Evidences.PageInfo.HasNextPage)

	th.CleanupOrganizationDataWithContext(ownerCtx, t)
}
