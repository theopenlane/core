//go:build test

package graphapi_test

import (
	"testing"

	"entgo.io/contrib/entgql"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
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

const paginatedEvidencesProgramCountQuery = `query GetEvidences($orderBy: [EvidenceOrder!], $first: Int) {
  evidences(orderBy: $orderBy, first: $first) {
    edges { node { id programs(first: 5) { totalCount } } }
  }
}`

type paginatedEvidencesProgramCountResponse struct {
	Evidences struct {
		Edges []struct {
			Node struct {
				ID       string `json:"id"`
				Programs struct {
					TotalCount int64 `json:"totalCount"`
				} `json:"programs"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"evidences"`
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

func TestPaginateKeepsNamedEdgesAcrossFGABatches(t *testing.T) {
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

	first := 1
	order := []*generated.EvidenceOrder{{Field: generated.EvidenceOrderFieldCreatedAt, Direction: entgql.OrderDirectionDesc}}

	conn, err := suite.Client.DB.Evidence.Query().
		WithNamedPrograms("programs", withNamedProgramMembers).
		Paginate(users.Member.UserCtx, nil, &first, nil, nil, generated.WithEvidenceOrder(order))
	assert.NilError(t, err)
	assert.Assert(t, is.Len(conn.Edges, 1))
	assert.Check(t, is.Equal(conn.Edges[0].Node.ID, newer.ID))
	assert.Check(t, conn.PageInfo.HasNextPage)

	checkEagerLoadedProgramMember(t, conn.Edges[0].Node, program.ID, users.Member.ID)

	conn, err = suite.Client.DB.Evidence.Query().
		WithNamedPrograms("programs", withNamedProgramMembers).
		Paginate(users.Member.UserCtx, conn.PageInfo.EndCursor, &first, nil, nil, generated.WithEvidenceOrder(order))
	assert.NilError(t, err)
	assert.Assert(t, is.Len(conn.Edges, 1))
	assert.Check(t, is.Equal(conn.Edges[0].Node.ID, older.ID))

	checkEagerLoadedProgramMember(t, conn.Edges[0].Node, program.ID, users.Member.ID)

	th.CleanupOrganizationDataWithContext(ownerCtx, t)
}

func withNamedProgramMembers(q *generated.ProgramQuery) {
	q.WithNamedMembers("members")
}

func checkEagerLoadedProgramMember(t *testing.T, evidence *generated.Evidence, programID, userID string) {
	t.Helper()

	programs, err := evidence.NamedPrograms("programs")
	assert.NilError(t, err)
	assert.Assert(t, is.Len(programs, 1))
	assert.Check(t, is.Equal(programs[0].ID, programID))

	members, err := programs[0].NamedMembers("members")
	assert.NilError(t, err)

	found := false

	for _, m := range members {
		if m.UserID == userID {
			found = true
		}
	}

	assert.Check(t, found, "member %s not eager loaded on program %s", userID, programID)
}

func TestQueryEvidencesNestedTotalCountAcrossFGABatches(t *testing.T) {
	users := suite.SeedFreshMinimalOrgUsers(t, true)
	ownerCtx := users.Owner.UserCtx

	program := (&th.ProgramBuilder{Client: suite.Client}).MustNew(ownerCtx, t)
	(&th.ProgramMemberBuilder{Client: suite.Client, ProgramID: program.ID, UserID: users.Member.ID, Role: enums.RoleMember.String()}).MustNew(ownerCtx, t)

	const hiddenPerGap = 12

	visible := (&th.EvidenceBuilder{Client: suite.Client, ProgramID: program.ID}).MustNew(ownerCtx, t)
	for range hiddenPerGap {
		(&th.EvidenceBuilder{Client: suite.Client}).MustNew(ownerCtx, t)
	}

	concreteClient, ok := suite.Client.API.TestGraphClient.(*testclient.Client)
	assert.Assert(t, ok)

	vars := map[string]any{
		"orderBy": []map[string]any{{"field": "created_at", "direction": "DESC"}},
		"first":   1,
	}

	var resp paginatedEvidencesProgramCountResponse
	err := concreteClient.Client.Post(users.Member.UserCtx, "GetEvidences", paginatedEvidencesProgramCountQuery, &resp, vars)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(resp.Evidences.Edges, 1))
	assert.Check(t, is.Equal(resp.Evidences.Edges[0].Node.ID, visible.ID))
	assert.Check(t, is.Equal(resp.Evidences.Edges[0].Node.Programs.TotalCount, int64(1)))

	th.CleanupOrganizationDataWithContext(ownerCtx, t)
}
