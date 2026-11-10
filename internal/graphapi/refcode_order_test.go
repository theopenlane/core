package graphapi_test

import (
	"context"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/controls"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	controlgen "github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/pkg/naturalsort"
)

// naturalRefCodeSuffixes are ref code suffixes listed in natural order
var naturalRefCodeSuffixes = []string{"1", "2", "2(1)", "2(2)", "2(10)", "3", "9", "10", "11", "12"}

const refCodePageSize = int64(3)

func refCodePrefix() string {
	return "NS" + strings.ToUpper(gofakeit.LetterN(8)) + "-"
}

func withPrefix(prefix string, suffixes []string) []string {
	return lo.Map(suffixes, func(s string, _ int) string { return prefix + s })
}

func shuffledCopy(values []string) []string {
	out := slices.Clone(values)
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })

	return out
}

func sortedCopy(values []string) []string {
	out := slices.Clone(values)
	slices.Sort(out)

	return out
}

func reversedCopy(values []string) []string {
	out := slices.Clone(values)
	slices.Reverse(out)

	return out
}

func controlRefCodes(resp *testclient.GetControls) []string {
	return lo.Map(resp.Controls.Edges, func(e *testclient.GetControls_Controls_Edges, _ int) string { return e.Node.RefCode })
}

func controlOrder(direction testclient.OrderDirection) []*testclient.ControlOrder {
	return []*testclient.ControlOrder{{Field: testclient.ControlOrderFieldRefCode, Direction: direction}}
}

// collectForward follows a connection page by page from the start and returns every ref code in the order served
func collectForward(fetch func(after *string) (refCodes []string, endCursor *string, hasNextPage bool)) []string {
	var got []string

	var after *string

	for {
		refCodes, endCursor, hasNextPage := fetch(after)
		got = append(got, refCodes...)

		if !hasNextPage {
			return got
		}

		after = endCursor
	}
}

func pagedControlRefCodes(t *testing.T, ctx context.Context, pageSize int64, direction testclient.OrderDirection, where *testclient.ControlWhereInput) []string {
	return collectForward(func(after *string) ([]string, *string, bool) {
		resp, err := suite.Client.API.GetControls(ctx, &pageSize, nil, after, nil, controlOrder(direction), where)
		assert.NilError(t, err)

		return controlRefCodes(resp), resp.Controls.PageInfo.EndCursor, resp.Controls.PageInfo.HasNextPage
	})
}

func TestQueryControlsRefCodeNaturalOrder(t *testing.T) {
	prefix := refCodePrefix()
	expected := withPrefix(prefix, naturalRefCodeSuffixes)

	controlIDs := make([]string, 0, len(expected))
	idByRefCode := map[string]string{}

	for _, refCode := range shuffledCopy(expected) {
		c := (&th.ControlBuilder{Client: suite.Client, RefCode: refCode}).MustNew(th.SharedAdminUser.UserCtx, t)
		controlIDs = append(controlIDs, c.ID)
		idByRefCode[refCode] = c.ID
	}

	ctx := th.SharedTestUser1.UserCtx
	where := &testclient.ControlWhereInput{RefCodeHasPrefix: &prefix}
	first := int64(len(expected) + 1)

	authCases := []struct {
		name     string
		client   *testclient.TestClient
		ctx      context.Context
		expected []string
	}{
		{name: "org user", client: suite.Client.API, ctx: th.SharedTestUser1.UserCtx, expected: expected},
		{name: "view only user", client: suite.Client.API, ctx: th.SharedViewOnlyUser.UserCtx, expected: expected},
		{name: "api token", client: suite.Client.APIWithToken, ctx: context.Background(), expected: expected},
		{name: "personal access token", client: suite.Client.APIWithPAT, ctx: context.Background(), expected: expected},
		{name: "user from another org", client: suite.Client.API, ctx: th.SharedTestUser2.UserCtx, expected: []string{}},
	}

	for _, tc := range authCases {
		t.Run("ascending, "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetControls(tc.ctx, &first, nil, nil, nil, controlOrder(testclient.OrderDirectionAsc), where)
			assert.NilError(t, err)
			assert.DeepEqual(t, tc.expected, controlRefCodes(resp))
		})
	}

	t.Run("descending", func(t *testing.T) {
		resp, err := suite.Client.API.GetControls(ctx, &first, nil, nil, nil, controlOrder(testclient.OrderDirectionDesc), where)
		assert.NilError(t, err)
		assert.DeepEqual(t, reversedCopy(expected), controlRefCodes(resp))
	})

	t.Run("forward paging returns every row once", func(t *testing.T) {
		assert.DeepEqual(t, expected, pagedControlRefCodes(t, ctx, refCodePageSize, testclient.OrderDirectionAsc, where))
	})

	t.Run("backward paging returns every row once", func(t *testing.T) {
		var got []string

		var before *string

		for {
			resp, err := suite.Client.API.GetControls(ctx, nil, new(refCodePageSize), nil, before, controlOrder(testclient.OrderDirectionAsc), where)
			assert.NilError(t, err)

			got = append(controlRefCodes(resp), got...)

			if !resp.Controls.PageInfo.HasPreviousPage {
				break
			}

			before = resp.Controls.PageInfo.StartCursor
		}

		assert.DeepEqual(t, expected, got)
	})

	t.Run("multi order with mixed directions", func(t *testing.T) {
		approved := []string{expected[1], expected[4], expected[8]}
		for _, refCode := range approved {
			_, err := suite.Client.API.UpdateControl(th.SharedAdminUser.UserCtx, idByRefCode[refCode], testclient.UpdateControlInput{
				Status: &enums.ControlStatusApproved,
			})
			assert.NilError(t, err)
		}

		notApproved := lo.Without(expected, approved...)

		resp, err := suite.Client.API.GetControls(ctx, &first, nil, nil, nil, []*testclient.ControlOrder{
			{Field: testclient.ControlOrderFieldStatus, Direction: testclient.OrderDirectionDesc},
			{Field: testclient.ControlOrderFieldRefCode, Direction: testclient.OrderDirectionAsc},
		}, where)
		assert.NilError(t, err)
		assert.DeepEqual(t, append(notApproved, approved...), controlRefCodes(resp))
	})

	t.Run("control report follows the natural order", func(t *testing.T) {
		resp, err := suite.Client.API.GetControlReports(ctx, &first, nil, nil, nil, []*testclient.ControlReportOrder{
			{Field: testclient.ControlReportOrderFieldRefCode, Direction: testclient.OrderDirectionAsc},
		}, where)
		assert.NilError(t, err)

		got := lo.Map(resp.ControlReports.Edges, func(e *testclient.GetControlReports_ControlReports_Edges, _ int) string { return e.Node.RefCode })
		assert.DeepEqual(t, expected, got)
	})

	t.Run("renaming a ref code moves the row", func(t *testing.T) {
		renamed := prefix + "0"

		_, err := suite.Client.API.UpdateControl(th.SharedAdminUser.UserCtx, idByRefCode[expected[len(expected)-1]], testclient.UpdateControlInput{
			RefCode: &renamed,
		})
		assert.NilError(t, err)

		resp, err := suite.Client.API.GetControls(ctx, &first, nil, nil, nil, controlOrder(testclient.OrderDirectionAsc), where)
		assert.NilError(t, err)
		assert.DeepEqual(t, append([]string{renamed}, expected[:len(expected)-1]...), controlRefCodes(resp))
	})

	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, IDs: controlIDs}).MustDelete(th.SharedAdminUser.UserCtx, t)
}

func TestQueryControlsRefCodeEqualKeysPaging(t *testing.T) {
	prefix := refCodePrefix()
	refCodes := withPrefix(prefix, []string{"2", "02", "002", "1", "3", "03"})

	controlIDs := make([]string, 0, len(refCodes))
	for _, refCode := range refCodes {
		c := (&th.ControlBuilder{Client: suite.Client, RefCode: refCode}).MustNew(th.SharedAdminUser.UserCtx, t)
		controlIDs = append(controlIDs, c.ID)
	}

	ctx := th.SharedTestUser1.UserCtx
	where := &testclient.ControlWhereInput{RefCodeHasPrefix: &prefix}
	all := int64(len(refCodes) + 1)

	for _, direction := range []testclient.OrderDirection{testclient.OrderDirectionAsc, testclient.OrderDirectionDesc} {
		t.Run(direction.String(), func(t *testing.T) {
			resp, err := suite.Client.API.GetControls(ctx, &all, nil, nil, nil, controlOrder(direction), where)
			assert.NilError(t, err)

			unpaged := controlRefCodes(resp)
			assert.DeepEqual(t, unpaged, pagedControlRefCodes(t, ctx, 2, direction, where))

			groups := [][]string{withPrefix(prefix, []string{"1"}), withPrefix(prefix, []string{"2", "02", "002"}), withPrefix(prefix, []string{"3", "03"})}
			if direction == testclient.OrderDirectionDesc {
				slices.Reverse(groups)
			}

			offset := 0
			for _, group := range groups {
				assert.DeepEqual(t, sortedCopy(group), sortedCopy(unpaged[offset:offset+len(group)]))
				offset += len(group)
			}
		})
	}

	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, IDs: controlIDs}).MustDelete(th.SharedAdminUser.UserCtx, t)
}

func TestQuerySubcontrolsRefCodeNaturalOrder(t *testing.T) {
	parent := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)

	prefix := refCodePrefix()
	expected := withPrefix(prefix, naturalRefCodeSuffixes)

	subcontrolIDs := make([]string, 0, len(expected))
	for _, refCode := range shuffledCopy(expected) {
		sc := (&th.SubcontrolBuilder{Client: suite.Client, Name: refCode, ControlID: parent.ID}).MustNew(th.SharedAdminUser.UserCtx, t)
		subcontrolIDs = append(subcontrolIDs, sc.ID)
	}

	ctx := th.SharedTestUser1.UserCtx
	where := &testclient.SubcontrolWhereInput{ControlID: &parent.ID}
	order := func(direction testclient.OrderDirection) []*testclient.SubcontrolOrder {
		return []*testclient.SubcontrolOrder{{Field: testclient.SubcontrolOrderFieldRefCode, Direction: direction}}
	}
	refCodes := func(resp *testclient.GetSubcontrolRefCodes) []string {
		return lo.Map(resp.Subcontrols.Edges, func(e *testclient.GetSubcontrolRefCodes_Subcontrols_Edges, _ int) string { return e.Node.RefCode })
	}
	first := int64(len(expected) + 1)

	t.Run("ascending and descending", func(t *testing.T) {
		resp, err := suite.Client.API.GetSubcontrolRefCodes(ctx, &first, nil, order(testclient.OrderDirectionAsc), where)
		assert.NilError(t, err)
		assert.DeepEqual(t, expected, refCodes(resp))

		resp, err = suite.Client.API.GetSubcontrolRefCodes(ctx, &first, nil, order(testclient.OrderDirectionDesc), where)
		assert.NilError(t, err)
		assert.DeepEqual(t, reversedCopy(expected), refCodes(resp))
	})

	t.Run("forward paging returns every row once", func(t *testing.T) {
		got := collectForward(func(after *string) ([]string, *string, bool) {
			resp, err := suite.Client.API.GetSubcontrolRefCodes(ctx, new(refCodePageSize), after, order(testclient.OrderDirectionAsc), where)
			assert.NilError(t, err)

			return refCodes(resp), resp.Subcontrols.PageInfo.EndCursor, resp.Subcontrols.PageInfo.HasNextPage
		})

		assert.DeepEqual(t, expected, got)
	})

	t.Run("nested connection on the parent control", func(t *testing.T) {
		got := collectForward(func(after *string) ([]string, *string, bool) {
			resp, err := suite.Client.API.GetControlSubcontrolRefCodes(ctx, parent.ID, new(refCodePageSize), after, order(testclient.OrderDirectionAsc))
			assert.NilError(t, err)

			page := resp.Control.Subcontrols

			return lo.Map(page.Edges, func(e *testclient.GetControlSubcontrolRefCodes_Control_Subcontrols_Edges, _ int) string {
				return e.Node.RefCode
			}), page.PageInfo.EndCursor, page.PageInfo.HasNextPage
		})

		assert.DeepEqual(t, expected, got)
	})

	t.Run("control report lists subcontrols in natural order", func(t *testing.T) {
		resp, err := suite.Client.API.GetControlReports(ctx, nil, nil, nil, nil, nil, &testclient.ControlWhereInput{ID: &parent.ID})
		assert.NilError(t, err)
		assert.Equal(t, len(resp.ControlReports.Edges), 1)

		got := lo.Map(resp.ControlReports.Edges[0].Node.Subcontrols, func(sc *testclient.GetControlReports_ControlReports_Edges_Node_Subcontrols, _ int) string {
			return sc.RefCode
		})
		assert.DeepEqual(t, expected, got)
	})

	(&th.Cleanup[*generated.SubcontrolDeleteOne]{Client: suite.Client.DB.Subcontrol, IDs: subcontrolIDs}).MustDelete(th.SharedAdminUser.UserCtx, t)
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, ID: parent.ID}).MustDelete(th.SharedAdminUser.UserCtx, t)
}

func TestRefCodeSortKeyOnBulkUpdate(t *testing.T) {
	prefix := refCodePrefix()
	c := (&th.ControlBuilder{Client: suite.Client, RefCode: prefix + "5"}).MustNew(th.SharedAdminUser.UserCtx, t)

	internalCtx := th.SetInternalContext(th.SharedAdminUser.UserCtx, suite.Client.DB)

	renamed := prefix + "50"
	_, err := suite.Client.DB.Control.Update().
		Where(controlgen.ID(c.ID)).
		SetRefCode(renamed).
		Save(internalCtx)
	assert.NilError(t, err)

	got, err := suite.Client.DB.Control.Get(internalCtx, c.ID)
	assert.NilError(t, err)
	assert.Equal(t, naturalsort.Key(renamed), got.RefCodeSortKey)

	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, ID: c.ID}).MustDelete(th.SharedAdminUser.UserCtx, t)
}

func TestBackfillRefCodeSortKeys(t *testing.T) {
	prefix := refCodePrefix()
	parent := (&th.ControlBuilder{Client: suite.Client, RefCode: prefix + "10"}).MustNew(th.SharedAdminUser.UserCtx, t)
	sub := (&th.SubcontrolBuilder{Client: suite.Client, Name: prefix + "10(2)", ControlID: parent.ID}).MustNew(th.SharedAdminUser.UserCtx, t)
	deleted := (&th.ControlBuilder{Client: suite.Client, RefCode: prefix + "11"}).MustNew(th.SharedAdminUser.UserCtx, t)

	internalCtx := th.SetInternalContext(th.SharedAdminUser.UserCtx, suite.Client.DB)
	allCtx := entx.SkipSoftDelete(internalCtx)

	for _, id := range []string{parent.ID, deleted.ID} {
		assert.NilError(t, suite.Client.DB.Control.UpdateOneID(id).SetRefCodeSortKey("").Exec(internalCtx))
	}

	assert.NilError(t, suite.Client.DB.Subcontrol.UpdateOneID(sub.ID).SetRefCodeSortKey("").Exec(internalCtx))
	assert.NilError(t, suite.Client.DB.Control.DeleteOneID(deleted.ID).Exec(internalCtx))

	before, err := suite.Client.DB.Control.Get(internalCtx, parent.ID)
	assert.NilError(t, err)

	sweepCtx := auth.WithCallerCapabilities(auth.WithSystemSweepContext(context.Background()), auth.CapBypassManagedGroup)

	for range 2 {
		assert.NilError(t, controls.BackfillRefCodeSortKeys(sweepCtx, suite.Client.DB))

		gotControl, err := suite.Client.DB.Control.Get(internalCtx, parent.ID)
		assert.NilError(t, err)
		assert.Equal(t, naturalsort.Key(prefix+"10"), gotControl.RefCodeSortKey)
		assert.Equal(t, before.UpdatedAt, gotControl.UpdatedAt)
		assert.Equal(t, before.UpdatedBy, gotControl.UpdatedBy)

		gotSub, err := suite.Client.DB.Subcontrol.Get(internalCtx, sub.ID)
		assert.NilError(t, err)
		assert.Equal(t, naturalsort.Key(prefix+"10(2)"), gotSub.RefCodeSortKey)

		gotDeleted, err := suite.Client.DB.Control.Get(allCtx, deleted.ID)
		assert.NilError(t, err)
		assert.Equal(t, naturalsort.Key(prefix+"11"), gotDeleted.RefCodeSortKey)
	}

	(&th.Cleanup[*generated.SubcontrolDeleteOne]{Client: suite.Client.DB.Subcontrol, ID: sub.ID}).MustDelete(th.SharedAdminUser.UserCtx, t)
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, ID: parent.ID}).MustDelete(th.SharedAdminUser.UserCtx, t)
}
