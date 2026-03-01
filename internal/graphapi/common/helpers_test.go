package common //nolint:revive

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gocarina/gocsv"
	"github.com/samber/lo"
	"github.com/vektah/gqlparser/v2/ast"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/objects"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
)

// TestStripOperation validates operation prefix stripping behavior.
func TestStripOperation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Create operation",
			input:    "createUser",
			expected: "User",
		},
		{
			name:     "Update operation",
			input:    "updateUser",
			expected: "User",
		},
		{
			name:     "Delete operation",
			input:    "deleteUser",
			expected: "User",
		},
		{
			name:     "Get operation",
			input:    "getUser",
			expected: "User",
		},
		{
			name:     "No operation",
			input:    "User",
			expected: "User",
		},
		{
			name:     "Non-matching prefix",
			input:    "fetchUser",
			expected: "fetchUser",
		},
		{
			name:     "Create Upload operation",
			input:    "createUploadProcedure",
			expected: "Procedure",
		},
		{
			name:     "Create Upload Document",
			input:    "createUploadDocument",
			expected: "Document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripOperation(tt.input)
			assert.Check(t, is.Equal(tt.expected, result))
		})
	}
}

// TestRetrieveObjectDetails verifies upload object metadata extraction.
func TestRetrieveObjectDetails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fieldName   string
		key         string
		arguments   ast.ArgumentList
		expected    *pkgobjects.File
		expectedErr error
	}{
		{
			name:      "Matching upload argument",
			fieldName: "createUser",
			key:       "file",
			arguments: ast.ArgumentList{
				&ast.Argument{
					Name: "file",
					Value: &ast.Value{
						ExpectedType: &ast.Type{
							NamedType: "Upload",
						},
					},
				},
			},
			expected: &pkgobjects.File{
				CorrelatedObjectType: "User",
				FileMetadata: pkgobjects.FileMetadata{
					Key: "file",
				},
			},
			expectedErr: nil,
		},
		{
			name:      "Non-matching upload argument",
			fieldName: "createUser",
			key:       "image",
			arguments: ast.ArgumentList{
				&ast.Argument{
					Name: "file",
					Value: &ast.Value{
						ExpectedType: &ast.Type{
							NamedType: "Upload",
						},
					},
				},
			},
			expected:    &pkgobjects.File{},
			expectedErr: ErrUnableToDetermineObjectType,
		},
		{
			name:        "No upload argument",
			fieldName:   "createUser",
			key:         "file",
			arguments:   ast.ArgumentList{},
			expected:    &pkgobjects.File{},
			expectedErr: ErrUnableToDetermineObjectType,
		},
		{
			name:      "Non-upload argument",
			fieldName: "createUser",
			key:       "file",
			arguments: ast.ArgumentList{
				&ast.Argument{
					Name: "file",
					Value: &ast.Value{
						ExpectedType: &ast.Type{
							NamedType: "String",
						},
					},
				},
			},
			expected:    &pkgobjects.File{},
			expectedErr: ErrUnableToDetermineObjectType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rctx := &graphql.FieldContext{
				Field: graphql.CollectedField{
					Field: &ast.Field{
						Name:      tt.fieldName,
						Arguments: tt.arguments,
					},
				},
			}

			upload := &pkgobjects.File{
				OriginalName: "meow.txt",
			}

			result, err := retrieveObjectDetails(rctx, nil, "", tt.key, upload)
			if tt.expectedErr != nil {

				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.Equal(tt.expected.CorrelatedObjectType, result.CorrelatedObjectType))
			assert.Check(t, is.Equal(tt.expected.Key, result.Key))
		})
	}
}

// TestTemplateKindFromVariables ensures template kind is parsed from variables.
func TestTemplateKindFromVariables(t *testing.T) {
	t.Parallel()

	variables := map[string]any{
		"input": map[string]any{
			"kind": enums.TemplateKindTrustCenterNda.String(),
		},
	}

	kind := templateKindFromVariables(variables, "input")
	assert.Check(t, kind != nil)
	assert.Check(t, is.Equal(enums.TemplateKindTrustCenterNda, *kind))
}

// TestRetrieveObjectDetailsTemplateKindFromInput ensures template kind metadata is applied from input.
func TestRetrieveObjectDetailsTemplateKindFromInput(t *testing.T) {
	t.Parallel()

	rctx := &graphql.FieldContext{
		Field: graphql.CollectedField{
			Field: &ast.Field{
				Name: "createTemplate",
				Arguments: ast.ArgumentList{
					&ast.Argument{
						Name: "templateFiles",
						Value: &ast.Value{
							ExpectedType: &ast.Type{
								NamedType: "Upload",
							},
						},
					},
				},
			},
		},
	}

	variables := map[string]any{
		"input": map[string]any{
			"kind": enums.TemplateKindTrustCenterNda.String(),
		},
	}

	upload := &pkgobjects.File{
		OriginalName: "nda.pdf",
	}

	result, err := retrieveObjectDetails(rctx, variables, "input", "templateFiles", upload)
	assert.NilError(t, err)
	assert.Check(t, result.ProviderHints != nil)
	assert.Check(t, is.Equal(enums.TemplateKindTrustCenterNda.String(), result.ProviderHints.Metadata[objects.TemplateKindMetadataKey]))
}

// TestRetrieveObjectDetailsTemplateKindFromFieldName ensures template kind fallback by field name.
func TestRetrieveObjectDetailsTemplateKindFromFieldName(t *testing.T) {
	t.Parallel()

	rctx := &graphql.FieldContext{
		Field: graphql.CollectedField{
			Field: &ast.Field{
				Name: "createTrustCenterNDA",
				Arguments: ast.ArgumentList{
					&ast.Argument{
						Name: "templateFiles",
						Value: &ast.Value{
							ExpectedType: &ast.Type{
								NamedType: "Upload",
							},
						},
					},
				},
			},
		},
	}

	upload := &pkgobjects.File{
		OriginalName: "nda.pdf",
	}

	result, err := retrieveObjectDetails(rctx, map[string]any{}, "input", "templateFiles", upload)
	assert.NilError(t, err)
	assert.Check(t, result.ProviderHints != nil)
	assert.Check(t, is.Equal(enums.TemplateKindTrustCenterNda.String(), result.ProviderHints.Metadata[objects.TemplateKindMetadataKey]))
}

// TestGetOrgOwnerFromInput validates owner ID extraction from inputs.
func TestGetOrgOwnerFromInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       any
		expected    *string
		expectedErr error
	}{
		{
			name:        "Nil input",
			input:       nil,
			expected:    nil,
			expectedErr: nil,
		},
		{
			name: "Valid input with owner ID",
			input: generated.CreateProcedureInput{
				Name:    "Test Procedure",
				OwnerID: lo.ToPtr("owner123"),
			},
			expected:    lo.ToPtr("owner123"),
			expectedErr: nil,
		},
		{
			name:  "Valid input without owner ID",
			input: generated.CreateRiskInput{
				// No OwnerID field set
			},
			expected:    nil,
			expectedErr: nil,
		},
		{
			name: "Invalid input type will return nil",
			input: struct {
				Name string `json:"name"`
			}{
				Name: "test",
			},
			expected:    nil,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetOrgOwnerFromInput(&tt.input)
			if tt.expectedErr != nil {

				assert.Check(t, is.Nil(result))
				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.DeepEqual(tt.expected, result))
		})
	}
}

// TestGetBulkUploadOwnerInput validates owner ID resolution for bulk uploads.
func TestGetBulkUploadOwnerInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       []*generated.CreateProcedureInput // used as an example, should work with any type
		expected    *string
		expectedErr error
	}{
		{
			name:        "Nil input, nothing to do",
			input:       nil,
			expected:    nil,
			expectedErr: nil,
		},
		{
			name: "Valid input with consistent owner IDs",
			input: []*generated.CreateProcedureInput{
				{
					Name:    "Test Procedure 1",
					OwnerID: lo.ToPtr("owner123"),
				},
				{
					Name:    "Test Procedure 2",
					OwnerID: lo.ToPtr("owner123"),
				},
			},
			expected:    lo.ToPtr("owner123"),
			expectedErr: nil,
		},
		{
			name: "Valid input with inconsistent owner IDs",
			input: []*generated.CreateProcedureInput{
				{
					Name:    "Test Procedure 1",
					OwnerID: lo.ToPtr("owner123"),
				},
				{
					Name:    "Test Procedure 2",
					OwnerID: lo.ToPtr("owner456"),
				},
			},
			expected:    nil,
			expectedErr: ErrNoOrganizationID,
		},
		{
			name: "Valid input with missing owner ID",
			input: []*generated.CreateProcedureInput{
				{
					Name: "Test Procedure 1",
				},
			},
			expected:    nil,
			expectedErr: ErrNoOrganizationID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetBulkUploadOwnerInput(tt.input)
			if tt.expectedErr != nil {

				assert.Check(t, is.Nil(result))
				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.DeepEqual(tt.expected, result))
		})
	}
}

// TestNormalizeCSVEnumInputs ensures CSV enum normalization behavior.
func TestNormalizeCSVEnumInputs(t *testing.T) {
	t.Parallel()

	type csvEnumInput struct {
		Status *enums.TaskStatus
		Role   enums.Role
		State  enums.TaskStatus
	}

	statusMixed := enums.TaskStatus("In Review")
	statusEmpty := enums.TaskStatus("  ")
	stateMixed := enums.TaskStatus("in progress")
	stateEmpty := enums.TaskStatus(" ")
	roleLower := enums.Role("member")

	data := []*csvEnumInput{
		{
			Status: &statusMixed,
			Role:   roleLower,
			State:  stateMixed,
		},
		{
			Status: &statusEmpty,
			Role:   enums.Role("ADMIN"),
			State:  stateEmpty,
		},
		{
			Status: nil,
			Role:   enums.Role("OWNER"),
			State:  enums.TaskStatus("completed"),
		},
	}

	normalizeCSVEnumInputs(data)

	assert.Assert(t, data[0].Status != nil)
	assert.Check(t, is.Equal(enums.TaskStatusInReview, *data[0].Status))
	assert.Check(t, is.Equal(enums.RoleMember, data[0].Role))
	assert.Check(t, is.Equal(enums.TaskStatusInProgress, data[0].State))
	assert.Check(t, is.Nil(data[1].Status))
	assert.Check(t, is.Equal(enums.TaskStatusOpen, data[1].State))
	assert.Check(t, is.Nil(data[2].Status))
	assert.Check(t, is.Equal(enums.RoleOwner, data[2].Role))
	assert.Check(t, is.Equal(enums.TaskStatusCompleted, data[2].State))
}

// TestNormalizeCSVDateTimePointers ensures zero DateTime pointers are cleared.
func TestNormalizeCSVDateTimePointers(t *testing.T) {
	t.Parallel()

	type csvDateTimeInput struct {
		Due       *models.DateTime
		Completed *models.DateTime
	}

	zero := models.DateTime{}
	nonZero := models.DateTime(time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC))

	data := []*csvDateTimeInput{
		{
			Due:       &zero,
			Completed: &nonZero,
		},
		{
			Due:       nil,
			Completed: &zero,
		},
	}

	normalizeCSVEnumInputs(data)

	assert.Check(t, is.Nil(data[0].Due))
	assert.Assert(t, data[0].Completed != nil)
	assert.Check(t, is.Equal(false, data[0].Completed.IsZero()))
	assert.Check(t, is.Nil(data[1].Completed))
}

// TestWrapCSVUnmarshalErrorAddsHeader verifies CSV error wrapping adds headers.
func TestWrapCSVUnmarshalErrorAddsHeader(t *testing.T) {
	t.Parallel()

	type csvRow struct {
		Tags []string
	}

	csvData := []byte("Tags\nsecurity\n")
	var rows []*csvRow
	err := gocsv.UnmarshalBytes(csvData, &rows)
	assert.Assert(t, err != nil)

	wrapped := wrapCSVUnmarshalError(err, csvData)
	vErr, ok := wrapped.(*ValidationError)
	assert.Assert(t, ok)
	assert.Check(t, is.DeepEqual([]string{"Tags"}, vErr.Fields()))
	assert.Check(t, strings.Contains(vErr.Message(), "Tags"))
	assert.Check(t, strings.Contains(strings.ToLower(vErr.Message()), "json"))
}

// TestIsEmpty validates empty-value detection across types.
func TestIsEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "Nil value",
			input:    nil,
			expected: true,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "Non-empty string",
			input:    "hello",
			expected: false,
		},
		{
			name:     "Zero integer",
			input:    0,
			expected: true,
		},
		{
			name:     "Non-zero integer",
			input:    42,
			expected: false,
		},
		{
			name:     "Zero float",
			input:    0.0,
			expected: true,
		},
		{
			name:     "Non-zero float",
			input:    3.14,
			expected: false,
		},
		{
			name:     "Empty slice",
			input:    []int{},
			expected: true,
		},
		{
			name:     "Non-empty slice",
			input:    []int{1, 2, 3},
			expected: false,
		},
		{
			name:     "Empty map",
			input:    map[string]any{},
			expected: true,
		},
		{
			name:     "Non-empty map",
			input:    map[string]int{"a": 1},
			expected: false,
		},
		{
			name:     "Boolean false",
			input:    false,
			expected: true,
		},
		{
			name:     "Boolean true",
			input:    true,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmpty(tt.input)
			assert.Check(t, is.Equal(tt.expected, result))
		})
	}
}

// TestSetOrganizationForUploads verifies org selection rules for uploads.
func TestSetOrganizationForUploads(t *testing.T) {
	t.Parallel()

	primaryOrg := ulids.New().String()
	secondaryOrg := ulids.New().String()

	tests := []struct {
		name        string
		caller      *auth.Caller
		variables   map[string]any
		inputKey    string
		expectedOrg string
		expectedErr error
	}{
		{
			name: "Org already set in context",
			caller: &auth.Caller{
				OrganizationID:     primaryOrg,
				AuthenticationType: auth.PATAuthentication,
			},
			variables:   map[string]any{},
			inputKey:    "input",
			expectedOrg: primaryOrg,
		},
		{
			name: "PAT requires explicit owner",
			caller: &auth.Caller{
				OrganizationIDs:    []string{primaryOrg, secondaryOrg},
				AuthenticationType: auth.PATAuthentication,
			},
			variables: map[string]any{
				"input": map[string]any{
					"ownerID": primaryOrg,
				},
			},
			inputKey:    "input",
			expectedOrg: primaryOrg,
		},
		{
			name: "PAT missing owner errors",
			caller: &auth.Caller{
				OrganizationIDs:    []string{primaryOrg, secondaryOrg},
				AuthenticationType: auth.PATAuthentication,
			},
			variables:   nil,
			inputKey:    "input",
			expectedErr: ErrNoOrganizationID,
		},
		{
			name: "Non-PAT single authorized org fallback",
			caller: &auth.Caller{
				OrganizationIDs:    []string{primaryOrg},
				AuthenticationType: auth.APITokenAuthentication,
			},
			variables:   nil,
			inputKey:    "input",
			expectedOrg: primaryOrg,
		},
		{
			name: "Non-PAT multiple orgs require owner input",
			caller: &auth.Caller{
				OrganizationIDs:    []string{primaryOrg, secondaryOrg},
				AuthenticationType: auth.APITokenAuthentication,
			},
			variables:   nil,
			inputKey:    "input",
			expectedErr: ErrNoOrganizationID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := auth.WithCaller(context.Background(), tt.caller)

			err := setOrganizationForUploads(ctx, tt.variables, tt.inputKey)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				return
			}

			assert.NilError(t, err)
			helpersTestCaller, helpersTestCallerOk := auth.CallerFromContext(ctx)
			assert.Check(t, helpersTestCallerOk)
			assert.Check(t, is.Equal(tt.expectedOrg, helpersTestCaller.OrganizationID))
		})
	}
}

// csvListRow is a helper struct for list parsing tests.
type csvListRow struct {
	Tags  []string `csv:"Tags"`
	Names []string `csv:"Names"`
}

// TestUnmarshalBulkDataListParsingMixedDelimiters ensures mixed delimiters are handled.
func TestUnmarshalBulkDataListParsingMixedDelimiters(t *testing.T) {
	t.Parallel()

	csvData := "Tags,Names\n\"foo,bar\",alpha|beta\n"
	upload := graphql.Upload{
		File:        strings.NewReader(csvData),
		Filename:    "list.csv",
		Size:        int64(len(csvData)),
		ContentType: "text/csv",
	}

	rows, err := UnmarshalBulkData[csvListRow](upload)
	assert.NilError(t, err)
	assert.Check(t, is.Len(rows, 1))

	assert.Check(t, is.DeepEqual(rows[0].Tags, []string{"foo", "bar"}))
	assert.Check(t, is.DeepEqual(rows[0].Names, []string{"alpha", "beta"}))
}

// TestUnmarshalBulkDataListParsing validates basic list parsing.
func TestUnmarshalBulkDataListParsing(t *testing.T) {
	t.Parallel()

	type csvRow struct {
		Tags []string
	}

	csvData := "Tags\nsecurity;compliance\n"
	upload := graphql.Upload{
		File:        strings.NewReader(csvData),
		Filename:    "list.csv",
		Size:        int64(len(csvData)),
		ContentType: "text/csv",
	}

	rows, err := UnmarshalBulkData[csvRow](upload)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(rows, 1))
	assert.Check(t, is.DeepEqual([]string{"security", "compliance"}, rows[0].Tags))
}

// TestUnmarshalBulkDataInvalidJSONMap verifies JSON map validation errors.
func TestUnmarshalBulkDataInvalidJSONMap(t *testing.T) {
	t.Parallel()

	type csvRow struct {
		Metadata map[string]any
	}

	csvData := "Metadata\n{not-json}\n"
	upload := graphql.Upload{
		File:        strings.NewReader(csvData),
		Filename:    "map.csv",
		Size:        int64(len(csvData)),
		ContentType: "text/csv",
	}

	_, err := UnmarshalBulkData[csvRow](upload)
	assert.Assert(t, err != nil)
	assert.Check(t, strings.Contains(err.Error(), "list or object values must be valid JSON"))
}

// TestGetOrgOwnerFromInputWrapped verifies wrapped Input owner extraction.
func TestGetOrgOwnerFromInputWrapped(t *testing.T) {
	t.Parallel()

	type ownerInput struct {
		OwnerID *string `json:"ownerID"`
	}
	type wrappedInput struct {
		Input ownerInput
	}

	orgID := "org-123"
	input := &wrappedInput{
		Input: ownerInput{OwnerID: &orgID},
	}

	owner, err := GetOrgOwnerFromInput(input)
	assert.NilError(t, err)
	assert.Assert(t, owner != nil)
	assert.Check(t, is.Equal(*owner, orgID))
}
