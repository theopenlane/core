package graphapi_test

import (
	"context"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/utils/ulids"
)

func TestQueryTemplate(t *testing.T) {
	// create an template to be queried using testUser1
	template := (&TemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// create a system admin root template
	templateRoot := (&TemplateBuilder{client: suite.client, TemplateType: enums.RootTemplate}).MustNew(systemAdminUser.UserCtx, t)

	// add test cases for querying the Template
	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: template.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, root template",
			queryID: templateRoot.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, root template, view only user",
			queryID: templateRoot.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: template.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: template.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "template not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "template not found, using not authorized user",
			queryID:  template.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:    "anonymous user can access root template",
			queryID: templateRoot.ID,
			client:  suite.client.api,
			ctx:     createAnonymousTrustCenterContext(ulids.New().String(), testUser1.OrganizationID),
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTemplateByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.Template.ID))
			assert.Check(t, is.Equal(enums.TemplateKindQuestionnaire, *resp.Template.Kind))
		})
	}

	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: template.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: templateRoot.ID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationCreateTemplate(t *testing.T) {
	// Helper function to create fresh file uploads for each test case
	createPDFUpload := func() *graphql.Upload {
		pdfFile, err := storage.NewUploadFile("testdata/uploads/hello.pdf")
		assert.NilError(t, err)
		return &graphql.Upload{
			File:        pdfFile.RawFile,
			Filename:    pdfFile.OriginalName,
			Size:        pdfFile.Size,
			ContentType: pdfFile.ContentType,
		}
	}

	createPNGUpload := func() *graphql.Upload {
		pngFile, err := storage.NewUploadFile("testdata/uploads/logo.png")
		assert.NilError(t, err)
		return &graphql.Upload{
			File:        pngFile.RawFile,
			Filename:    pngFile.OriginalName,
			Size:        pngFile.Size,
			ContentType: pngFile.ContentType,
		}
	}

	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name          string
		input         testclient.CreateTemplateInput
		templateFiles []*graphql.Upload
		client        *testclient.TestClient
		ctx           context.Context
		expectedErr   string
	}{
		{
			name: "happy path, minimal input without files",
			input: testclient.CreateTemplateInput{
				Name: "Test Template",
				Jsonconfig: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{
							"type": "string",
						},
					},
				},
			},
			templateFiles: nil,
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
		},
		{
			name: "happy path, full input without files",
			input: testclient.CreateTemplateInput{
				Name:         "Full Test Template",
				Description:  lo.ToPtr("A comprehensive test template"),
				Kind:         &enums.TemplateKindQuestionnaire,
				TemplateType: &enums.Document,
				Tags:         []string{"test", "template", "questionnaire"},
				Jsonconfig: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{
							"type":  "string",
							"title": "Name",
						},
						"email": map[string]any{
							"type":   "string",
							"format": "email",
							"title":  "Email",
						},
					},
					"required": []string{"name", "email"},
				},
				Uischema: map[string]any{
					"type": "VerticalLayout",
					"elements": []map[string]any{
						{
							"type":  "Control",
							"scope": "#/properties/name",
						},
						{
							"type":  "Control",
							"scope": "#/properties/email",
						},
					},
				},
			},
			templateFiles: nil,
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
		},
		{
			name: "happy path, with single PDF file",
			input: testclient.CreateTemplateInput{
				Name: "Template with PDF",
				Jsonconfig: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"document": map[string]any{
							"type":  "string",
							"title": "Document",
						},
					},
				},
			},
			templateFiles: []*graphql.Upload{createPDFUpload()},
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
		},
		{
			name: "happy path, with multiple files",
			input: testclient.CreateTemplateInput{
				Name:        "Template with Multiple Files",
				Description: lo.ToPtr("Template with various file types"),
				Jsonconfig: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"files": map[string]any{
							"type":  "array",
							"title": "Files",
						},
					},
				},
			},
			templateFiles: []*graphql.Upload{createPDFUpload(), createPNGUpload()},
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
		},
		{
			name: "missing required name field",
			input: testclient.CreateTemplateInput{
				Jsonconfig: map[string]any{
					"type": "object",
				},
			},
			templateFiles: nil,
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
			expectedErr:   "value is less than the required length",
		},
		{
			name: "missing required jsonconfig field",
			input: testclient.CreateTemplateInput{
				Name: "Template without JSON config",
			},
			templateFiles: nil,
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
			expectedErr:   "cannot be null",
		},
		{
			name: "happy path, system admin creating root template",
			input: testclient.CreateTemplateInput{
				Name:         "Root Template",
				TemplateType: &enums.RootTemplate,
				Jsonconfig: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"field": map[string]any{
							"type": "string",
						},
					},
				},
			},
			templateFiles: nil,
			client:        suite.client.api,
			ctx:           systemAdminUser.UserCtx,
		},
		{
			name: "trust center NDA with no trust center",
			input: testclient.CreateTemplateInput{
				Name: "Test Template",
				Kind: &enums.TemplateKindTrustCenterNda,
				Jsonconfig: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{
							"type": "string",
						},
					},
				},
			},
			templateFiles: nil,
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
			expectedErr:   "generated: constraint failed: pq: new row for relation",
		},
		{
			name: "trust center NDA with trust center",
			input: testclient.CreateTemplateInput{
				Name: "Test Template",
				Kind: &enums.TemplateKindTrustCenterNda,
				Jsonconfig: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{
							"type": "string",
						},
					},
				},
				TrustCenterID: lo.ToPtr(trustCenter.ID),
			},
			templateFiles: nil,
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			// Set up mock expectations for file uploads if files are provided
			if len(tc.templateFiles) > 0 && tc.expectedErr == "" {
				uploads := make([]graphql.Upload, len(tc.templateFiles))
				for i, file := range tc.templateFiles {
					uploads[i] = *file
				}
				expectUpload(t, suite.client.mockProvider, uploads)
			} else if len(tc.templateFiles) > 0 {
				// For error cases with files, we still need to set up the mock but expect it not to be called
				expectUploadCheckOnly(t, suite.client.mockProvider)
			}

			resp, err := tc.client.CreateTemplate(tc.ctx, tc.input, tc.templateFiles)

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			require.Nil(t, err)
			require.NotNil(t, resp)

			// Verify basic template fields
			template := resp.CreateTemplate.Template
			assert.Check(t, is.Equal(tc.input.Name, template.Name))
			assert.Check(t, template.Jsonconfig != nil)

			// Verify optional fields if provided
			if tc.input.Description != nil {
				assert.Check(t, is.Equal(*tc.input.Description, *template.Description))
			}

			if tc.input.Kind != nil {
				assert.Check(t, is.Equal(*tc.input.Kind, *template.Kind))
			}

			if tc.input.Uischema != nil {
				assert.Check(t, template.Uischema != nil)
			}

			// Verify file uploads if files were provided
			if len(tc.templateFiles) > 0 {
				assert.Check(t, is.Len(template.Files.Edges, len(tc.templateFiles)))

				// Verify each uploaded file has an ID and presigned URL
				for _, edge := range template.Files.Edges {
					assert.Check(t, edge.Node != nil)
					assert.Check(t, edge.Node.ID != "")
					assert.Check(t, edge.Node.PresignedURL != nil)
				}
			}

			// Cleanup the created template
			(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: template.ID}).MustDelete(tc.ctx, t)
		})
	}
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateTemplate(t *testing.T) {
	// Helper function to create fresh file uploads for each test case
	createPDFUpload := func() *graphql.Upload {
		pdfFile, err := storage.NewUploadFile("testdata/uploads/hello.pdf")
		assert.NilError(t, err)
		return &graphql.Upload{
			File:        pdfFile.RawFile,
			Filename:    pdfFile.OriginalName,
			Size:        pdfFile.Size,
			ContentType: pdfFile.ContentType,
		}
	}

	createPNGUpload := func() *graphql.Upload {
		pngFile, err := storage.NewUploadFile("testdata/uploads/logo.png")
		assert.NilError(t, err)
		return &graphql.Upload{
			File:        pngFile.RawFile,
			Filename:    pngFile.OriginalName,
			Size:        pngFile.Size,
			ContentType: pngFile.ContentType,
		}
	}

	// Create a template to be updated
	template := (&TemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	template2 := (&TemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// Create a root template for system admin tests
	rootTemplate := (&TemplateBuilder{
		client:       suite.client,
		TemplateType: enums.RootTemplate,
		Name:         "Root Template for Update",
	}).MustNew(systemAdminUser.UserCtx, t)

	testCases := []struct {
		name          string
		templateID    string
		input         testclient.UpdateTemplateInput
		templateFiles []*graphql.Upload
		client        *testclient.TestClient
		ctx           context.Context
		expectedErr   string
	}{
		{
			name:       "happy path, update name only",
			templateID: template.ID,
			input: testclient.UpdateTemplateInput{
				Name: lo.ToPtr("Updated Template Name"),
			},
			templateFiles: nil,
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
		},
		{
			name:       "happy path, update description",
			templateID: template.ID,
			input: testclient.UpdateTemplateInput{
				Description: lo.ToPtr("Updated template description"),
			},
			templateFiles: nil,
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
		},
		{
			name:       "happy path, update multiple fields",
			templateID: template.ID,
			input: testclient.UpdateTemplateInput{
				Name:        lo.ToPtr("Multi-field Updated Template"),
				Description: lo.ToPtr("Updated description with multiple changes"),
				Kind:        &enums.TemplateKindQuestionnaire,
				Jsonconfig: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"updatedField": map[string]any{
							"type":  "string",
							"title": "Updated Field",
						},
						"newField": map[string]any{
							"type":  "number",
							"title": "New Field",
						},
					},
					"required": []string{"updatedField"},
				},
				Uischema: map[string]any{
					"type": "VerticalLayout",
					"elements": []map[string]any{
						{
							"type":  "Control",
							"scope": "#/properties/updatedField",
						},
						{
							"type":  "Control",
							"scope": "#/properties/newField",
						},
					},
				},
			},
			templateFiles: nil,
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
		},
		{
			name:       "happy path, update with single file",
			templateID: template.ID,
			input: testclient.UpdateTemplateInput{
				Name: lo.ToPtr("Template with Updated File"),
				Jsonconfig: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"document": map[string]any{
							"type":  "string",
							"title": "Updated Document",
						},
					},
				},
			},
			templateFiles: []*graphql.Upload{createPDFUpload()},
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
		},
		{
			name:       "happy path, update with multiple files",
			templateID: template2.ID,
			input: testclient.UpdateTemplateInput{
				Name:        lo.ToPtr("Template with Multiple Updated Files"),
				Description: lo.ToPtr("Template updated with various file types"),
				Jsonconfig: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"files": map[string]any{
							"type":  "array",
							"title": "Updated Files",
						},
					},
				},
			},
			templateFiles: []*graphql.Upload{createPDFUpload(), createPNGUpload()},
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
		},
		{
			name:       "happy path, update tags",
			templateID: template.ID,
			input: testclient.UpdateTemplateInput{
				Tags: []string{"updated", "template", "test"},
			},
			templateFiles: nil,
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
		},
		{
			name:       "happy path, append tags",
			templateID: template.ID,
			input: testclient.UpdateTemplateInput{
				AppendTags: []string{"additional", "tag"},
			},
			templateFiles: nil,
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
		},
		{
			name:       "happy path, using personal access token",
			templateID: template.ID,
			input: testclient.UpdateTemplateInput{
				Name: lo.ToPtr("PAT Updated Template"),
			},
			templateFiles: nil,
			client:        suite.client.apiWithPAT,
			ctx:           context.Background(),
		},
		{
			name:       "happy path, system admin updating root template",
			templateID: rootTemplate.ID,
			input: testclient.UpdateTemplateInput{
				Name:        lo.ToPtr("Updated Root Template"),
				Description: lo.ToPtr("Updated root template description"),
			},
			templateFiles: nil,
			client:        suite.client.api,
			ctx:           systemAdminUser.UserCtx,
		},
		{
			name:       "template not found, invalid ID",
			templateID: "invalid-id",
			input: testclient.UpdateTemplateInput{
				Name: lo.ToPtr("Should Not Work"),
			},
			templateFiles: nil,
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
			expectedErr:   notFoundErrorMsg,
		},
		{
			name:       "template not found, unauthorized user",
			templateID: template.ID,
			input: testclient.UpdateTemplateInput{
				Name: lo.ToPtr("Unauthorized Update"),
			},
			templateFiles: nil,
			client:        suite.client.api,
			ctx:           testUser2.UserCtx,
			expectedErr:   notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			// Set up mock expectations for file uploads if files are provided
			if len(tc.templateFiles) > 0 && tc.expectedErr == "" {
				uploads := make([]graphql.Upload, len(tc.templateFiles))
				for i, file := range tc.templateFiles {
					uploads[i] = *file
				}
				expectUpload(t, suite.client.mockProvider, uploads)
			} else if len(tc.templateFiles) > 0 {
				// For error cases with files, we still need to set up the mock but expect it not to be called
				expectUploadCheckOnly(t, suite.client.mockProvider)
			}

			resp, err := tc.client.UpdateTemplate(tc.ctx, tc.templateID, tc.input, tc.templateFiles)

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Verify basic template fields
			updatedTemplate := resp.UpdateTemplate.Template
			assert.Check(t, is.Equal(tc.templateID, updatedTemplate.ID))

			// Verify updated fields if provided
			if tc.input.Name != nil {
				assert.Check(t, is.Equal(*tc.input.Name, updatedTemplate.Name))
			}

			if tc.input.Description != nil {
				assert.Check(t, is.Equal(*tc.input.Description, *updatedTemplate.Description))
			}

			if tc.input.ClearDescription != nil && *tc.input.ClearDescription {
				assert.Check(t, updatedTemplate.Description == nil)
			}

			if tc.input.Kind != nil {
				assert.Check(t, is.Equal(*tc.input.Kind, *updatedTemplate.Kind))
			}

			if tc.input.Jsonconfig != nil {
				assert.Check(t, updatedTemplate.Jsonconfig != nil)
				// Verify specific fields in jsonconfig if needed
				if properties, ok := tc.input.Jsonconfig["properties"]; ok {
					templateProperties, exists := updatedTemplate.Jsonconfig["properties"]
					assert.Check(t, exists)
					assert.Check(t, templateProperties != nil)

					// Check for specific properties that were updated
					if updatedField, hasUpdated := properties.(map[string]any)["updatedField"]; hasUpdated {
						templateProps := templateProperties.(map[string]any)
						assert.Check(t, templateProps["updatedField"] != nil)
						updatedFieldMap := updatedField.(map[string]any)
						templateUpdatedField := templateProps["updatedField"].(map[string]any)
						assert.Check(t, is.Equal(updatedFieldMap["title"], templateUpdatedField["title"]))
					}
				}
			}

			if tc.input.Uischema != nil {
				assert.Check(t, updatedTemplate.Uischema != nil)
			}

			if tc.input.ClearUischema != nil && *tc.input.ClearUischema && tc.input.Uischema == nil {
				assert.Check(t, updatedTemplate.Uischema == nil)
			}

			if tc.input.Tags != nil {
				// Note: Tags verification would depend on how tags are returned in the response
				// This might need adjustment based on the actual response structure
			}

			// Verify file uploads if files were provided
			if len(tc.templateFiles) > 0 {
				t.Logf("Number of files: %+v", updatedTemplate.Files.Edges)
				assert.Check(t, is.Len(updatedTemplate.Files.Edges, len(tc.templateFiles)))

				// Verify each uploaded file has an ID and presigned URL
				for _, edge := range updatedTemplate.Files.Edges {
					assert.Check(t, edge.Node != nil)
					assert.Check(t, edge.Node.ID != "")
					assert.Check(t, edge.Node.PresignedURL != nil)
				}
			}

			// Verify that updatedAt and updatedBy fields are set
			assert.Check(t, updatedTemplate.UpdatedAt != nil)
			assert.Check(t, updatedTemplate.UpdatedBy != nil)
		})
	}

	// Cleanup the created templates
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: template.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: template2.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: rootTemplate.ID}).MustDelete(systemAdminUser.UserCtx, t)
}
