package graphapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestCreateReviewUpdatesEntityReviewFields(t *testing.T) {
	frequency := enums.FrequencyMonthly

	tt := []struct {
		name         string
		input        testclient.CreateReviewInput
		client       *testclient.TestClient
		ctx          context.Context
		expectedErr  string
		expectFields bool
		setup        func(t *testing.T) ([]string, []string)
	}{
		{
			name: "happy path, review with entity",
			input: testclient.CreateReviewInput{
				Title:   "Test Review",
				Summary: lo.ToPtr("Test summary"),
			},
			client:       suite.client.api,
			ctx:          sharedTestUser1.UserCtx,
			expectFields: true,
			setup: func(t *testing.T) ([]string, []string) {
				entity := (&EntityBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
				ctx := setContext(sharedTestUser1.UserCtx, suite.client.db)
				err := suite.client.db.Entity.UpdateOneID(entity.ID).
					SetReviewFrequency(frequency).
					Exec(ctx)
				assert.NilError(t, err)
				return []string{entity.ID}, []string{entity.EntityTypeID}
			},
		},
		{
			name: "review with review frequency updates next review date",
			input: testclient.CreateReviewInput{
				Title:   "Review with Frequency",
				Summary: lo.ToPtr("With frequency"),
			},
			client:       suite.client.api,
			ctx:          sharedTestUser1.UserCtx,
			expectFields: true,
			setup: func(t *testing.T) ([]string, []string) {
				entity := (&EntityBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
				_, err := suite.client.api.UpdateEntity(sharedTestUser1.UserCtx, entity.ID, testclient.UpdateEntityInput{
					ReviewFrequency: lo.ToPtr(frequency),
				}, nil, nil, nil, nil)
				assert.NilError(t, err)
				return []string{entity.ID}, []string{entity.EntityTypeID}
			},
		},
		{
			name: "review without entities does not update review fields",
			input: testclient.CreateReviewInput{
				Title:   "Review Without Entities",
				Summary: lo.ToPtr("This review has no entities linked"),
			},
			client:       suite.client.api,
			ctx:          sharedTestUser1.UserCtx,
			expectFields: false,
			setup: func(t *testing.T) ([]string, []string) {
				entity := (&EntityBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
				return []string{entity.ID}, []string{entity.EntityTypeID}
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			entityIDs, entityTypeIDs := tc.setup(t)
			if tc.expectFields {
				tc.input.EntityIDs = entityIDs
			}

			resp, err := tc.client.CreateReview(tc.ctx, tc.input)

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			review := resp.GetCreateReview().GetReview()
			assert.Assert(t, review != nil)

			if tc.expectFields {
				for _, entityID := range entityIDs {
					resp, err := suite.client.api.GetEntityByID(sharedTestUser1.UserCtx, entityID)
					assert.NilError(t, err)

					updatedEntity := resp.Entity

					assert.Check(t, lo.FromPtr(updatedEntity.ReviewedBy) != "", "reviewed_by should be set")
					assert.Check(t, updatedEntity.LastReviewedAt != nil, "last_reviewed_at should be set")

					if tc.input.Title == "Review with Frequency" {
						assert.Check(t, updatedEntity.NextReviewAt != nil, "next_review_at should be set when entity has review frequency")
						if updatedEntity.NextReviewAt != nil && updatedEntity.LastReviewedAt != nil {
							lastReviewedTime := time.Time(*updatedEntity.LastReviewedAt)
							expectedNextReview := lastReviewedTime.AddDate(0, 1, 0)
							nextReviewTime := time.Time(*updatedEntity.NextReviewAt).UTC()
							assert.Check(t, is.DeepEqual(
								expectedNextReview,
								nextReviewTime,
							), "next_review_at should be one month after last_reviewed_at")
						}
					}
				}
			} else {
				for _, entityID := range entityIDs {
					resp, err := suite.client.api.GetEntityByID(sharedTestUser1.UserCtx, entityID)
					assert.NilError(t, err)

					entity := resp.Entity

					assert.Check(t, is.Equal("", lo.FromPtr(entity.ReviewedBy)), "reviewed_by should not be set when no entities linked")
					assert.Check(t, is.Nil(entity.LastReviewedAt), "last_reviewed_at should not be set when no entities linked")
					assert.Check(t, is.Nil(entity.NextReviewAt), "next_review_at should not be set when no entities linked")
				}
			}

			_, err = suite.client.api.DeleteReview(sharedTestUser1.UserCtx, review.ID)
			assert.NilError(t, err)

			(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, IDs: entityIDs}).MustDelete(sharedTestUser1.UserCtx, t)
			(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, IDs: entityTypeIDs}).MustDelete(sharedTestUser1.UserCtx, t)
		})
	}
}

func TestCreateReview(t *testing.T) {
	entity1 := (&EntityBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	program := (&ProgramBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	program2 := (&ProgramBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	control := (&ControlBuilder{client: suite.client, ProgramID: program.ID}).MustNew(sharedTestUser1.UserCtx, t)

	programInOrg2 := (&ProgramBuilder{client: suite.client}).MustNew(sharedTestUser2.UserCtx, t)
	controlInOrg2 := (&ControlBuilder{client: suite.client, ProgramID: programInOrg2.ID}).MustNew(sharedTestUser2.UserCtx, t)

	firstProgramReviewResp, err := suite.client.api.CreateReview(sharedTestUser1.UserCtx, testclient.CreateReviewInput{
		Title:      "First Program Review",
		ProgramIDs: []string{program.ID},
	})
	assert.NilError(t, err)
	assert.Check(t, firstProgramReviewResp != nil)

	entitiesToCleanup := []string{entity1.ID}
	entityTypesToCleanup := []string{entity1.EntityTypeID}
	controlsToCleanup := []string{control.ID}
	programsToCleanup := []string{program.ID, program2.ID}
	controlsNoAccessToCleanup := []string{controlInOrg2.ID}
	programsNoAccessToCleanup := []string{programInOrg2.ID}

	testCases := []struct {
		name        string
		reviewInput testclient.CreateReviewInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			reviewInput: testclient.CreateReviewInput{
				Title: "Minimal Review",
			},
			client: suite.client.api,
			ctx:    sharedTestUser1.UserCtx,
		},
		{
			name: "happy path, full input",
			reviewInput: testclient.CreateReviewInput{
				Title:     "Full Review",
				Summary:   lo.ToPtr("Test summary"),
				Details:   lo.ToPtr("Test details"),
				State:     lo.ToPtr("completed"),
				Category:  lo.ToPtr("security"),
				EntityIDs: []string{entity1.ID},
				Source:    lo.ToPtr("manual"),
				Tags:      []string{"test", "review"},
			},
			client: suite.client.api,
			ctx:    sharedTestUser1.UserCtx,
		},
		{
			name: "happy path, auditor creating review linked to program",
			reviewInput: testclient.CreateReviewInput{
				Title:      "Program Review",
				ProgramIDs: []string{program.ID},
			},
			client: suite.client.api,
			ctx:    sharedAuditorUser.UserCtx,
		},
		{
			name: "happy path, second review linked to same program",
			reviewInput: testclient.CreateReviewInput{
				Title:      "Second Program Review",
				ProgramIDs: []string{program.ID},
			},
			client: suite.client.api,
			ctx:    sharedTestUser1.UserCtx,
		},
		{
			name: "happy path, auditor creating review linked to program with reviewer",
			reviewInput: testclient.CreateReviewInput{
				Title:      "Program Review with Reviewer",
				ProgramIDs: []string{program2.ID},
				Reporter:   lo.ToPtr("Reporter 1"),
				ReviewerID: lo.ToPtr(sharedAuditorUser.ID),
			},
			client: suite.client.api,
			ctx:    sharedAuditorUser.UserCtx,
		},
		{
			name: "happy path, auditor creating review linked to control",
			reviewInput: testclient.CreateReviewInput{
				Title:      "Control Review",
				ControlIDs: []string{control.ID},
			},
			client: suite.client.api,
			ctx:    sharedAuditorUser.UserCtx,
		},
		{
			name: "auditor unable to access program",
			reviewInput: testclient.CreateReviewInput{
				Title:      "Unauthorized Program Review",
				ProgramIDs: []string{programInOrg2.ID},
			},
			client:      suite.client.api,
			ctx:         sharedAuditorUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "auditor unable to access control",
			reviewInput: testclient.CreateReviewInput{
				Title:      "Unauthorized Control Review",
				ControlIDs: []string{controlInOrg2.ID},
			},
			client:      suite.client.api,
			ctx:         sharedAuditorUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "happy path, using PAT",
			reviewInput: testclient.CreateReviewInput{
				Title: "PAT Review",
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, auditor",
			reviewInput: testclient.CreateReviewInput{
				Title: "Auditor Review",
			},
			client: suite.client.api,
			ctx:    sharedAuditorUser.UserCtx,
		},
		{
			name: "not authorized to create review",
			reviewInput: testclient.CreateReviewInput{
				Title: "Unauthorized Review",
			},
			client:      suite.client.api,
			ctx:         sharedViewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateReview(tc.ctx, tc.reviewInput)

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			review := resp.GetCreateReview().GetReview()
			assert.Assert(t, review != nil)
			assert.Check(t, is.Equal(tc.reviewInput.Title, review.Title))

			if tc.reviewInput.Summary != nil {
				assert.Check(t, is.Equal(*tc.reviewInput.Summary, *review.Summary))
			}

			if tc.reviewInput.Reporter != nil {
				assert.Check(t, is.Equal(*tc.reviewInput.Reporter, lo.FromPtr(review.Reporter)))
			}

			if tc.reviewInput.ReviewerID != nil {
				assert.Check(t, is.Equal(*tc.reviewInput.ReviewerID, lo.FromPtr(review.ReviewerID)))
			}

			if len(tc.reviewInput.ProgramIDs) > 0 {
				ctx := setContext(sharedTestUser1.UserCtx, suite.client.db)
				entReview, err := suite.client.db.Review.Get(ctx, review.ID)
				assert.NilError(t, err)

				programs, err := entReview.QueryPrograms().All(ctx)
				assert.NilError(t, err)
				assert.Check(t, is.Len(programs, len(tc.reviewInput.ProgramIDs)))
				if len(programs) > 0 {
					assert.Check(t, is.Equal(tc.reviewInput.ProgramIDs[0], programs[0].ID))
				}

				if tc.reviewInput.ProgramIDs[0] == program.ID {
					entProgram, err := suite.client.db.Program.Get(ctx, program.ID)
					assert.NilError(t, err)

					linkedReviews, err := entProgram.QueryReviews().All(ctx)
					assert.NilError(t, err)
					assert.Check(t, is.Len(linkedReviews, 2))
				}
			}

			if len(tc.reviewInput.ControlIDs) > 0 {
				ctx := setContext(sharedTestUser1.UserCtx, suite.client.db)
				entReview, err := suite.client.db.Review.Get(ctx, review.ID)
				assert.NilError(t, err)

				controls, err := entReview.QueryControls().All(ctx)
				assert.NilError(t, err)
				assert.Check(t, is.Len(controls, len(tc.reviewInput.ControlIDs)))
				if len(controls) > 0 {
					assert.Check(t, is.Equal(tc.reviewInput.ControlIDs[0], controls[0].ID))
				}
			}

			if len(tc.reviewInput.EntityIDs) > 0 || len(tc.reviewInput.ProgramIDs) > 0 || len(tc.reviewInput.ControlIDs) > 0 {
				_, err = suite.client.api.DeleteReview(sharedTestUser1.UserCtx, review.ID)
				assert.NilError(t, err)
			}
		})
	}

	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, IDs: entitiesToCleanup}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, IDs: entityTypesToCleanup}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: controlsToCleanup}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.ReviewDeleteOne]{client: suite.client.db.Review, ID: firstProgramReviewResp.CreateReview.Review.ID}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, IDs: programsToCleanup}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: controlsNoAccessToCleanup}).MustDelete(sharedTestUser2.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, IDs: programsNoAccessToCleanup}).MustDelete(sharedTestUser2.UserCtx, t)
}

func TestQueryReview(t *testing.T) {
	entity1 := (&EntityBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)

	entityIDs := []string{entity1.ID}
	entityTypeIDs := []string{entity1.EntityTypeID}

	createResp, err := suite.client.api.CreateReview(sharedTestUser1.UserCtx, testclient.CreateReviewInput{
		Title:     "Query Test Review",
		EntityIDs: entityIDs,
		Summary:   lo.ToPtr("Test summary for query"),
	})
	assert.NilError(t, err)
	review := createResp.GetCreateReview().GetReview()
	assert.Assert(t, review != nil)

	reviewID := review.ID

	t.Run("get review by ID", func(t *testing.T) {
		resp, err := suite.client.api.GetReviewByID(sharedTestUser1.UserCtx, reviewID)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		fetchedReview := resp.GetReview()
		assert.Assert(t, fetchedReview != nil)
		assert.Check(t, is.Equal(reviewID, fetchedReview.ID))
		assert.Check(t, is.Equal("Query Test Review", fetchedReview.Title))
	})

	t.Run("review not found", func(t *testing.T) {
		_, err := suite.client.api.GetReviewByID(sharedTestUser1.UserCtx, "invalid-id")
		assert.ErrorContains(t, err, "review not found")
	})

	_, err = suite.client.api.DeleteReview(sharedTestUser1.UserCtx, reviewID)
	assert.NilError(t, err)

	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, IDs: entityIDs}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, IDs: entityTypeIDs}).MustDelete(sharedTestUser1.UserCtx, t)
}

func TestQueryReviews(t *testing.T) {
	entity1 := (&EntityBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)

	entityIDs := []string{entity1.ID}
	entityTypeIDs := []string{entity1.EntityTypeID}
	reviewsToCleanup := []string{}

	beforeResp, err := suite.client.api.GetAllReviews(sharedTestUser1.UserCtx)
	initialCount := 0
	if err == nil && beforeResp.GetReviews() != nil {
		initialCount = len(beforeResp.GetReviews().Edges)
	}

	review1, err := suite.client.api.CreateReview(sharedTestUser1.UserCtx, testclient.CreateReviewInput{
		Title:   "First Review",
		Summary: lo.ToPtr("First summary"),
	})
	assert.NilError(t, err)
	reviewsToCleanup = append(reviewsToCleanup, review1.GetCreateReview().GetReview().ID)

	review2, err := suite.client.api.CreateReview(sharedTestUser1.UserCtx, testclient.CreateReviewInput{
		Title:   "Second Review",
		Summary: lo.ToPtr("Second summary"),
	})
	assert.NilError(t, err)
	reviewsToCleanup = append(reviewsToCleanup, review2.GetCreateReview().GetReview().ID)

	review3, err := suite.client.api.CreateReview(sharedTestUser1.UserCtx, testclient.CreateReviewInput{
		Title:     "Entity Review",
		EntityIDs: entityIDs,
	})
	assert.NilError(t, err)
	reviewsToCleanup = append(reviewsToCleanup, review3.GetCreateReview().GetReview().ID)

	testCases := []struct {
		name         string
		client       *testclient.TestClient
		ctx          context.Context
		expectedDiff int
	}{
		{
			name:         "list all reviews",
			client:       suite.client.api,
			ctx:          sharedTestUser1.UserCtx,
			expectedDiff: 3,
		},
		{
			name:         "list reviews using PAT",
			client:       suite.client.apiWithPAT,
			ctx:          context.Background(),
			expectedDiff: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllReviews(tc.ctx)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			reviews := resp.GetReviews()
			assert.Check(t, is.Len(reviews.Edges, initialCount+tc.expectedDiff),
				"expected %d reviews, got %d (initial: %d)", initialCount+tc.expectedDiff, len(reviews.Edges), initialCount)
		})
	}

	for _, reviewID := range reviewsToCleanup {
		_, err = suite.client.api.DeleteReview(sharedTestUser1.UserCtx, reviewID)
		if err != nil {
			t.Logf("failed to delete review %s: %v", reviewID, err)
		}
	}

	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, IDs: entityIDs}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, IDs: entityTypeIDs}).MustDelete(sharedTestUser1.UserCtx, t)
}

func TestDeleteReview(t *testing.T) {
	entity1 := (&EntityBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)

	entityIDs := []string{entity1.ID}
	entityTypeIDs := []string{entity1.EntityTypeID}

	createResp, err := suite.client.api.CreateReview(sharedTestUser1.UserCtx, testclient.CreateReviewInput{
		Title:     "Delete Test Review",
		EntityIDs: entityIDs,
	})
	assert.NilError(t, err)

	reviewID := createResp.GetCreateReview().GetReview().ID

	t.Run("delete review", func(t *testing.T) {
		resp, err := suite.client.api.DeleteReview(sharedTestUser1.UserCtx, reviewID)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		deletedReview := resp.GetDeleteReview()
		assert.Check(t, is.Equal(reviewID, deletedReview.DeletedID))
	})

	t.Run("review not found after delete", func(t *testing.T) {
		_, err := suite.client.api.GetReviewByID(sharedTestUser1.UserCtx, reviewID)
		assert.ErrorContains(t, err, "review not found")
	})

	t.Run("not authorized to delete", func(t *testing.T) {
		anotherReview, err := suite.client.api.CreateReview(sharedTestUser1.UserCtx, testclient.CreateReviewInput{
			Title: "Another Review",
		})
		assert.NilError(t, err)
		anotherReviewID := anotherReview.GetCreateReview().GetReview().ID

		_, err = suite.client.api.DeleteReview(sharedViewOnlyUser.UserCtx, anotherReviewID)
		assert.ErrorContains(t, err, notAuthorizedErrorMsg)

		_, err = suite.client.api.DeleteReview(sharedTestUser1.UserCtx, anotherReviewID)
		assert.NilError(t, err)
	})

	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, IDs: entityIDs}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, IDs: entityTypeIDs}).MustDelete(sharedTestUser1.UserCtx, t)
}

func TestUpdateReview(t *testing.T) {
	entity1 := (&EntityBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)

	entityIDs := []string{entity1.ID}
	entityTypeIDs := []string{entity1.EntityTypeID}

	createResp, err := suite.client.api.CreateReview(sharedTestUser1.UserCtx, testclient.CreateReviewInput{
		Title:   "Original Title",
		Summary: lo.ToPtr("Original summary"),
	})
	assert.NilError(t, err)

	reviewID := createResp.GetCreateReview().GetReview().ID

	t.Run("update review", func(t *testing.T) {
		resp, err := suite.client.api.UpdateReview(sharedTestUser1.UserCtx, reviewID, testclient.UpdateReviewInput{
			Title:   lo.ToPtr("Updated Title"),
			Summary: lo.ToPtr("Updated summary"),
		}, nil)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		updatedReview := resp.GetUpdateReview().GetReview()
		assert.Check(t, is.Equal("Updated Title", updatedReview.Title))
		assert.Check(t, is.Equal("Updated summary", *updatedReview.Summary))
	})

	t.Run("auditor can update review", func(t *testing.T) {
		createResp, err := suite.client.api.CreateReview(sharedAuditorUser.UserCtx, testclient.CreateReviewInput{
			Title: "Auditor Review",
		})
		assert.NilError(t, err)

		auditorReviewID := createResp.GetCreateReview().GetReview().ID

		resp, err := suite.client.api.UpdateReview(sharedAuditorUser.UserCtx, auditorReviewID, testclient.UpdateReviewInput{
			Title: lo.ToPtr("Auditor Updated Review"),
		}, nil)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal("Auditor Updated Review", resp.GetUpdateReview().GetReview().Title))

		_, err = suite.client.api.DeleteReview(sharedTestUser1.UserCtx, auditorReviewID)
		assert.NilError(t, err)
	})

	t.Run("update review not found", func(t *testing.T) {
		_, err := suite.client.api.UpdateReview(sharedTestUser1.UserCtx, "invalid-id", testclient.UpdateReviewInput{
			Title: lo.ToPtr("New Title"),
		}, nil)
		assert.ErrorContains(t, err, "review not found")
	})

	_, err = suite.client.api.DeleteReview(sharedTestUser1.UserCtx, reviewID)
	assert.NilError(t, err)

	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, IDs: entityIDs}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, IDs: entityTypeIDs}).MustDelete(sharedTestUser1.UserCtx, t)
}

func TestReviewWithReviewFrequencyCalculation(t *testing.T) {
	frequencies := []struct {
		frequency enums.Frequency
		name      string
		addMonths int
		addYears  int
	}{
		{frequency: enums.FrequencyMonthly, name: "monthly", addMonths: 1},
		{frequency: enums.FrequencyQuarterly, name: "quarterly", addMonths: 3},
		{frequency: enums.FrequencyBiAnnually, name: "biannually", addMonths: 6},
		{frequency: enums.FrequencyYearly, name: "yearly", addYears: 1},
	}

	for _, freq := range frequencies {
		t.Run("next review date for "+freq.name+" frequency", func(t *testing.T) {
			testEntity := (&EntityBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
			entityIDs := []string{testEntity.ID}
			entityTypeIDs := []string{testEntity.EntityTypeID}

			defer func() {
				(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, IDs: entityIDs}).MustDelete(sharedTestUser1.UserCtx, t)
				(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, IDs: entityTypeIDs}).MustDelete(sharedTestUser1.UserCtx, t)
			}()

			_, err := suite.client.api.UpdateEntity(sharedTestUser1.UserCtx, testEntity.ID, testclient.UpdateEntityInput{
				ReviewFrequency: lo.ToPtr(freq.frequency),
			}, nil, nil, nil, nil)
			assert.NilError(t, err)

			_, err = suite.client.api.CreateReview(sharedTestUser1.UserCtx, testclient.CreateReviewInput{
				Title:     freq.name + " review",
				EntityIDs: []string{testEntity.ID},
				Summary:   lo.ToPtr("Testing " + freq.name + " frequency"),
			})
			assert.NilError(t, err)

			resp, err := suite.client.apiWithToken.GetEntityByID(sharedTestUser1.UserCtx, testEntity.ID)
			assert.NilError(t, err)

			updatedEntity := resp.Entity

			assert.Check(t, updatedEntity.LastReviewedAt != nil, "last_reviewed_at should be set")
			assert.Check(t, updatedEntity.NextReviewAt != nil, "next_review_at should be set for "+freq.name)

			lastReviewTime := time.Time(*updatedEntity.LastReviewedAt).UTC()

			var expectedReviewDate time.Time
			if freq.addYears > 0 {
				expectedReviewDate = lastReviewTime.AddDate(freq.addYears, 0, 0)
			} else {
				expectedReviewDate = lastReviewTime.AddDate(0, freq.addMonths, 0)
			}

			nextReviewTime := time.Time(*updatedEntity.NextReviewAt).UTC()

			assert.Check(t, is.DeepEqual(expectedReviewDate.Year(), nextReviewTime.Year()),
				"next_review_at year should match expected")

			assert.Check(t, is.DeepEqual(expectedReviewDate.Month(), nextReviewTime.Month()),
				"next_review_at month should match expected")

			assert.Check(t, is.DeepEqual(expectedReviewDate.Day(), nextReviewTime.Day()),
				"next_review_at day should match expected")
		})
	}
}

func TestReviewWithMultipleConnectedEntities(t *testing.T) {
	testEntity := (&EntityBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)

	ids := []string{testEntity.ID}
	typeIDs := []string{testEntity.EntityTypeID}

	_, err := suite.client.api.UpdateEntity(sharedTestUser1.UserCtx, testEntity.ID, testclient.UpdateEntityInput{
		ReviewFrequency: lo.ToPtr(enums.FrequencyMonthly),
	}, nil, nil, nil, nil)
	assert.NilError(t, err)

	_, err = suite.client.api.CreateReview(sharedTestUser1.UserCtx, testclient.CreateReviewInput{
		Title:     "Multi-Entity Review",
		EntityIDs: ids,
		Summary:   lo.ToPtr("we are reviewing multiple entities at once"),
	})
	assert.NilError(t, err)

	resp, err := suite.client.apiWithToken.GetEntityByID(sharedTestUser1.UserCtx, testEntity.ID)
	assert.NilError(t, err)

	newEntity := resp.Entity

	assert.Check(t, lo.FromPtr(newEntity.ReviewedBy) != "", "reviewed_by should be set")
	assert.Check(t, newEntity.LastReviewedAt != nil, "last_reviewed_at should be set")
	assert.Check(t, newEntity.NextReviewAt != nil, "next_review_at should be set")

	lastReviewedTime := time.Time(*newEntity.LastReviewedAt).UTC()
	nextReviewTime := time.Time(*newEntity.NextReviewAt).UTC()

	expectedReviewDate := lastReviewedTime.AddDate(0, 1, 0)

	assert.Check(t, is.DeepEqual(expectedReviewDate, nextReviewTime),
		"next_review_at should be one month after last_reviewed_at")

	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, IDs: ids}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, IDs: typeIDs}).MustDelete(sharedTestUser1.UserCtx, t)
}
