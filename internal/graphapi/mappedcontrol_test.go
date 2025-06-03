package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/ulids"
)

func TestQueryMappedControl(t *testing.T) {
	// create an mappedControl to be queried using testUser1
	toControl := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	fromControl := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	mappedControl := (&MappedControlBuilder{client: suite.client, ToControlIDs: []string{toControl.ID}, FromControlIDs: []string{fromControl.ID}}).MustNew(testUser1.UserCtx, t)

	toControls := mappedControl.Edges.ToControls
	fromControls := mappedControl.Edges.FromControls

	// add test cases for querying the mappedControl
	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: mappedControl.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, read only user, should have read access",
			queryID: mappedControl.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: mappedControl.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "mappedControl not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "mappedControl not found, using not authorized user",
			queryID:  mappedControl.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetMappedControlByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.MappedControl.ID))

			fromControls := resp.MappedControl.FromControls.Edges
			assert.Check(t, is.Len(fromControls, 1), "expected exactly one from control")
			toControls := resp.MappedControl.ToControls.Edges
			assert.Check(t, is.Len(toControls, 1), "expected exactly one to control")
		})
	}

	(&Cleanup[*generated.MappedControlDeleteOne]{client: suite.client.db.MappedControl, ID: mappedControl.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{toControls[0].ID, fromControls[0].ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryMappedControls(t *testing.T) {
	// create multiple objects to be queried using testUser1
	controlsToDelete := []*generated.Control{}

	mappedControl1 := (&MappedControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	controlsToDelete = mappedControl1.Edges.ToControls
	controlsToDelete = append(controlsToDelete, mappedControl1.Edges.FromControls...)

	mappedControl2 := (&MappedControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	controlsToDelete = append(controlsToDelete, mappedControl2.Edges.ToControls...)
	controlsToDelete = append(controlsToDelete, mappedControl2.Edges.FromControls...)

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using api token",
			client:          suite.client.apiWithToken,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "happy path, using pat",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "another user, no mappedControls should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllMappedControls(tc.ctx)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.MappedControls.Edges, tc.expectedResults))
		})
	}

	(&Cleanup[*generated.MappedControlDeleteOne]{client: suite.client.db.MappedControl, IDs: []string{mappedControl1.ID, mappedControl2.ID}}).MustDelete(testUser1.UserCtx, t)

	for _, control := range controlsToDelete {
		(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control.ID}}).MustDelete(testUser1.UserCtx, t)
	}
}

func TestMutationCreateMappedControl(t *testing.T) {
	toControl := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	fromControl := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	toSubcontrol := (&SubcontrolBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	fromSubcontrol := (&SubcontrolBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.CreateMappedControlInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateMappedControlInput{
				MappingType: &enums.MappingTypeEqual,
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateMappedControlInput{
				MappingType:       &enums.MappingTypeEqual,
				ToControlIDs:      []string{toControl.ID},
				FromControlIDs:    []string{fromControl.ID},
				ToSubcontrolIDs:   []string{toSubcontrol.ID},
				FromSubcontrolIDs: []string{fromSubcontrol.ID},
				Relation:          lo.ToPtr("Controls are equal"),
				Confidence:        lo.ToPtr(int64(87)),
				Tags:              []string{"tag1", "tag2"},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: openlaneclient.CreateMappedControlInput{
				OwnerID:           &testUser1.OrganizationID,
				MappingType:       &enums.MappingTypeSubset,
				ToControlIDs:      []string{toControl.ID},
				FromSubcontrolIDs: []string{fromSubcontrol.ID},
				Relation:          lo.ToPtr("Controls are a subset"),
				Confidence:        lo.ToPtr(int64(21)),
				Tags:              []string{"tag1", "tag2"},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "using pat, missing owner ID",
			request: openlaneclient.CreateMappedControlInput{
				MappingType:       &enums.MappingTypeSubset,
				ToControlIDs:      []string{toControl.ID},
				FromSubcontrolIDs: []string{fromSubcontrol.ID},
				Relation:          lo.ToPtr("Controls are a subset"),
				Confidence:        lo.ToPtr(int64(21)),
				Tags:              []string{"tag1", "tag2"},
			},
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			expectedErr: "owner_id is required",
		},
		{
			name: "happy path, using api token",
			request: openlaneclient.CreateMappedControlInput{
				MappingType:       &enums.MappingTypeSubset,
				ToControlIDs:      []string{toControl.ID},
				FromSubcontrolIDs: []string{fromSubcontrol.ID},
				Relation:          lo.ToPtr("Controls are a subset"),
				Confidence:        lo.ToPtr(int64(21)),
				Tags:              []string{"tag1", "tag2"},
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: openlaneclient.CreateMappedControlInput{
				MappingType:    &enums.MappingTypeEqual,
				ToControlIDs:   []string{toControl.ID},
				FromControlIDs: []string{fromControl.ID},
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "invalid confidence, should be between 0 and 100",
			request: openlaneclient.CreateMappedControlInput{
				MappingType:    &enums.MappingTypeEqual,
				ToControlIDs:   []string{toControl.ID},
				FromControlIDs: []string{fromControl.ID},
				Confidence:     lo.ToPtr(int64(101)),
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx,
			expectedErr: "value out of range",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateMappedControl(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.request.MappingType != nil {
				assert.Check(t, is.Equal(*tc.request.MappingType, resp.CreateMappedControl.MappedControl.MappingType))
			} else {
				assert.Check(t, is.Equal(enums.MappingTypeEqual, resp.CreateMappedControl.MappedControl.MappingType))
			}

			if tc.request.Relation != nil {
				assert.Check(t, is.Equal(*tc.request.Relation, *resp.CreateMappedControl.MappedControl.Relation))
			} else {
				assert.Check(t, is.Equal(*resp.CreateMappedControl.MappedControl.Relation, ""))
			}

			if tc.request.Confidence != nil {
				assert.Check(t, is.Equal(*tc.request.Confidence, *resp.CreateMappedControl.MappedControl.Confidence))
			} else {
				assert.Check(t, resp.CreateMappedControl.MappedControl.Confidence == nil)
			}

			if tc.request.ToControlIDs != nil {
				assert.Check(t, is.Len(resp.CreateMappedControl.MappedControl.ToControls.Edges, len(tc.request.ToControlIDs)))
				for _, toControlID := range tc.request.ToControlIDs {
					assert.Check(t, lo.ContainsBy(resp.CreateMappedControl.MappedControl.ToControls.Edges, func(edge *openlaneclient.CreateMappedControl_CreateMappedControl_MappedControl_ToControls_Edges) bool {
						return edge.Node.ID == toControlID
					}), "expected toControl with ID %s to be present in the response", toControlID)
				}
			} else {
				assert.Check(t, is.Len(resp.CreateMappedControl.MappedControl.ToControls.Edges, 0), "expected no toControls in the response")
			}

			if tc.request.FromControlIDs != nil {
				assert.Check(t, is.Len(resp.CreateMappedControl.MappedControl.FromControls.Edges, len(tc.request.FromControlIDs)))
				for _, fromControlID := range tc.request.FromControlIDs {
					assert.Check(t, lo.ContainsBy(resp.CreateMappedControl.MappedControl.FromControls.Edges, func(edge *openlaneclient.CreateMappedControl_CreateMappedControl_MappedControl_FromControls_Edges) bool {
						return edge.Node.ID == fromControlID
					}), "expected fromControl with ID %s to be present in the response", fromControlID)
				}
			} else {
				assert.Check(t, is.Len(resp.CreateMappedControl.MappedControl.FromControls.Edges, 0), "expected no fromControls in the response")
			}

			if tc.request.ToSubcontrolIDs != nil {
				assert.Check(t, is.Len(resp.CreateMappedControl.MappedControl.ToSubcontrols.Edges, len(tc.request.ToSubcontrolIDs)))
				for _, toSubcontrolID := range tc.request.ToSubcontrolIDs {
					assert.Check(t, lo.ContainsBy(resp.CreateMappedControl.MappedControl.ToSubcontrols.Edges, func(edge *openlaneclient.CreateMappedControl_CreateMappedControl_MappedControl_ToSubcontrols_Edges) bool {
						return edge.Node.ID == toSubcontrolID
					}), "expected toSubcontrol with ID %s to be present in the response", toSubcontrolID)
				}

			} else {
				assert.Check(t, is.Len(resp.CreateMappedControl.MappedControl.ToSubcontrols.Edges, 0), "expected no toSubcontrols in the response")
			}

			if tc.request.FromSubcontrolIDs != nil {
				assert.Check(t, is.Len(resp.CreateMappedControl.MappedControl.FromSubcontrols.Edges, len(tc.request.FromSubcontrolIDs)))
				for _, fromSubcontrolID := range tc.request.FromSubcontrolIDs {
					assert.Check(t, lo.ContainsBy(resp.CreateMappedControl.MappedControl.FromSubcontrols.Edges, func(edge *openlaneclient.CreateMappedControl_CreateMappedControl_MappedControl_FromSubcontrols_Edges) bool {
						return edge.Node.ID == fromSubcontrolID
					}), "expected fromSubcontrol with ID %s to be present in the response", fromSubcontrolID)
				}
			} else {
				assert.Check(t, is.Len(resp.CreateMappedControl.MappedControl.FromSubcontrols.Edges, 0), "expected no fromSubcontrols in the response")
			}

			assert.Check(t, is.Len(resp.CreateMappedControl.MappedControl.Tags, len(tc.request.Tags)), "expected %d tags in the response", len(tc.request.Tags))

			// cleanup each object created
			(&Cleanup[*generated.MappedControlDeleteOne]{client: suite.client.db.MappedControl, ID: resp.CreateMappedControl.MappedControl.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}

	// cleanup the controls created for the mappedControl
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{toControl.ID, fromControl.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, IDs: []string{toSubcontrol.ID, fromSubcontrol.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateMappedControl(t *testing.T) {
	mappedControl := (&MappedControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	mappedControlAnotherOrg := (&MappedControlBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	controlA := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	controlB := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	subcontrolA := (&SubcontrolBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	subcontrolB := (&SubcontrolBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	controlAnotherOrg := (&ControlBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	resp, err := suite.client.api.GetMappedControlByID(testUser1.UserCtx, mappedControlAnotherOrg.ID)
	assert.ErrorContains(t, err, notFoundErrorMsg)
	assert.Assert(t, resp == nil)

	testCases := []struct {
		name            string
		requestID       string
		request         openlaneclient.UpdateMappedControlInput
		client          *openlaneclient.OpenlaneClient
		ctx             context.Context
		expectedErr     string
		controlNotAdded bool // used to indicate if the control was not added to the mapped control
	}{
		{
			name:      "happy path, update field",
			requestID: mappedControl.ID,
			request: openlaneclient.UpdateMappedControlInput{
				MappingType: lo.ToPtr(enums.MappingTypeSubset),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:      "happy path, update multiple fields",
			requestID: mappedControl.ID,
			request: openlaneclient.UpdateMappedControlInput{
				Relation:             lo.ToPtr("Updated relation"),
				Confidence:           lo.ToPtr(int64(75)),
				Tags:                 []string{"updated-tag1", "updated-tag2"},
				AddToControlIDs:      []string{controlA.ID, controlB.ID},
				AddFromSubcontrolIDs: []string{subcontrolA.ID, subcontrolB.ID},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:      "happy path, remove controls by org admin",
			requestID: mappedControl.ID,
			request: openlaneclient.UpdateMappedControlInput{
				RemoveToControlIDs:      []string{controlA.ID},
				RemoveFromSubcontrolIDs: []string{subcontrolA.ID},
				AddFromControlIDs:       []string{controlB.ID},
				AddToSubcontrolIDs:      []string{subcontrolB.ID},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:      "add controls from another org, not allowed",
			requestID: mappedControl.ID,
			request: openlaneclient.UpdateMappedControlInput{
				AddFromControlIDs: []string{controlAnotherOrg.ID},
			},
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			controlNotAdded: true, // this control should not be added
		},
		{
			name:      "update not allowed, not enough permissions",
			requestID: mappedControl.ID,
			request: openlaneclient.UpdateMappedControlInput{
				Relation: lo.ToPtr("Trying to update relation"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:      "update not allowed, owned by another org",
			requestID: mappedControlAnotherOrg.ID,
			request: openlaneclient.UpdateMappedControlInput{
				Relation: lo.ToPtr("Trying to update relation"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:      "update not allowed, no permissions",
			requestID: mappedControl.ID,
			request: openlaneclient.UpdateMappedControlInput{
				Relation: lo.ToPtr("Trying to update relation"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateMappedControl(tc.ctx, tc.requestID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.request.MappingType != nil {
				assert.Check(t, is.Equal(*tc.request.MappingType, resp.UpdateMappedControl.MappedControl.MappingType))
			}

			if tc.request.Relation != nil {
				assert.Check(t, is.Equal(*tc.request.Relation, *resp.UpdateMappedControl.MappedControl.Relation))
			}

			if tc.request.Confidence != nil {
				assert.Check(t, is.Equal(*tc.request.Confidence, *resp.UpdateMappedControl.MappedControl.Confidence))
			}

			if tc.request.AddToControlIDs != nil {
				for _, toControlID := range tc.request.AddToControlIDs {
					assert.Check(t, lo.ContainsBy(resp.UpdateMappedControl.MappedControl.ToControls.Edges, func(edge *openlaneclient.UpdateMappedControl_UpdateMappedControl_MappedControl_ToControls_Edges) bool {
						return edge.Node.ID == toControlID
					}), "expected toControl with ID %s to be present in the response", toControlID)
				}
			}

			if tc.request.AddFromControlIDs != nil {
				for _, fromControlID := range tc.request.AddFromControlIDs {
					if tc.controlNotAdded {
						assert.Check(t, !lo.ContainsBy(resp.UpdateMappedControl.MappedControl.FromControls.Edges, func(edge *openlaneclient.UpdateMappedControl_UpdateMappedControl_MappedControl_FromControls_Edges) bool {
							return edge.Node.ID == fromControlID
						}), "expected fromControl with ID %s to not be present in the response", fromControlID)
						continue
					} else {
						assert.Check(t, lo.ContainsBy(resp.UpdateMappedControl.MappedControl.FromControls.Edges, func(edge *openlaneclient.UpdateMappedControl_UpdateMappedControl_MappedControl_FromControls_Edges) bool {
							return edge.Node.ID == fromControlID
						}), "expected fromControl with ID %s to be present in the response", fromControlID)
					}
				}
			}

			if tc.request.AddToSubcontrolIDs != nil {
				for _, toSubcontrolID := range tc.request.AddToSubcontrolIDs {
					assert.Check(t, lo.ContainsBy(resp.UpdateMappedControl.MappedControl.ToSubcontrols.Edges, func(edge *openlaneclient.UpdateMappedControl_UpdateMappedControl_MappedControl_ToSubcontrols_Edges) bool {
						return edge.Node.ID == toSubcontrolID
					}), "expected toSubcontrol with ID %s to be present in the response", toSubcontrolID)
				}
			}

			if tc.request.AddFromSubcontrolIDs != nil {
				for _, fromSubcontrolID := range tc.request.AddFromSubcontrolIDs {
					assert.Check(t, lo.ContainsBy(resp.UpdateMappedControl.MappedControl.FromSubcontrols.Edges, func(edge *openlaneclient.UpdateMappedControl_UpdateMappedControl_MappedControl_FromSubcontrols_Edges) bool {
						return edge.Node.ID == fromSubcontrolID
					}), "expected fromSubcontrol with ID %s to be present in the response", fromSubcontrolID)
				}
			}

			if tc.request.RemoveToControlIDs != nil {
				for _, removedID := range tc.request.RemoveToControlIDs {
					assert.Check(t, !lo.ContainsBy(resp.UpdateMappedControl.MappedControl.ToControls.Edges, func(edge *openlaneclient.UpdateMappedControl_UpdateMappedControl_MappedControl_ToControls_Edges) bool {
						return edge.Node.ID == removedID
					}), "expected toControl with ID %s to be removed from the response", removedID)
				}
			}

			if tc.request.RemoveFromControlIDs != nil {
				for _, removedID := range tc.request.RemoveFromControlIDs {
					assert.Check(t, !lo.ContainsBy(resp.UpdateMappedControl.MappedControl.FromControls.Edges, func(edge *openlaneclient.UpdateMappedControl_UpdateMappedControl_MappedControl_FromControls_Edges) bool {
						return edge.Node.ID == removedID
					}), "expected fromControl with ID %s to be removed from the response", removedID)
				}
			}

			if tc.request.RemoveToSubcontrolIDs != nil {
				for _, removedID := range tc.request.RemoveToSubcontrolIDs {
					assert.Check(t, !lo.ContainsBy(resp.UpdateMappedControl.MappedControl.ToSubcontrols.Edges, func(edge *openlaneclient.UpdateMappedControl_UpdateMappedControl_MappedControl_ToSubcontrols_Edges) bool {
						return edge.Node.ID == removedID
					}), "expected toSubcontrol with ID %s to be removed from the response", removedID)
				}
			}

			if tc.request.RemoveFromSubcontrolIDs != nil {
				for _, removedID := range tc.request.RemoveFromSubcontrolIDs {
					assert.Check(t, !lo.ContainsBy(resp.UpdateMappedControl.MappedControl.FromSubcontrols.Edges, func(edge *openlaneclient.UpdateMappedControl_UpdateMappedControl_MappedControl_FromSubcontrols_Edges) bool {
						return edge.Node.ID == removedID
					}), "expected fromSubcontrol with ID %s to be removed from the response", removedID)
				}
			}

			if tc.request.Tags != nil {
				assert.Check(t, is.Len(resp.UpdateMappedControl.MappedControl.Tags, len(tc.request.Tags)), "expected %d tags in the response", len(tc.request.Tags))
				for _, tag := range tc.request.Tags {
					assert.Check(t, lo.Contains(resp.UpdateMappedControl.MappedControl.Tags, tag), "expected tag %s to be present in the response", tag)
				}
			}
		})
	}

	(&Cleanup[*generated.MappedControlDeleteOne]{client: suite.client.db.MappedControl, ID: mappedControl.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.MappedControlDeleteOne]{client: suite.client.db.MappedControl, ID: mappedControlAnotherOrg.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{controlA.ID, controlB.ID, mappedControl.Edges.FromControls[0].ID, mappedControl.Edges.ToControls[0].ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{mappedControlAnotherOrg.Edges.FromControls[0].ID, mappedControlAnotherOrg.Edges.ToControls[0].ID}}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, IDs: []string{subcontrolA.ID, subcontrolB.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteMappedControl(t *testing.T) {
	// create objects to be deleted
	mappedControl1 := (&MappedControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	mappedControl2 := (&MappedControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  mappedControl1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: mappedControl1.ID,
			client:     suite.client.api,
			ctx:        adminUser.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  mappedControl1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: mappedControl2.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteMappedControl(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteMappedControl.DeletedID))
		})
	}
}
