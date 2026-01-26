package hooks

import (
	"context"
	"testing"

	"entgo.io/ent"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
)

func TestFilenameToTitle(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "with extension",
			filename: "test_file.txt",
			want:     "Test File",
		},
		{
			name:     "with hyphens",
			filename: "hello-world.pdf",
			want:     "Hello World",
		},
		{
			name:     "with underscores and hyphens",
			filename: "complex_file-name.docx",
			want:     "Complex File Name",
		},
		{
			name:     "no extension",
			filename: "simple_name",
			want:     "Simple Name",
		},
		{
			name:     "with extra spaces",
			filename: "  spaced_file.pdf  ",
			want:     "Spaced File",
		},
		{
			name:     "empty string",
			filename: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filenameToTitle(tt.filename)
			assert.Check(t, is.Equal(got, tt.want))
		})
	}
}
func TestUpdatePlaceholderText(t *testing.T) {
	tests := []struct {
		name     string
		details  string
		orgName  string
		expected string
	}{
		{
			name:     "replace with organization name",
			details:  "Welcome to {{company_name}}, we're glad to have you!",
			orgName:  "Openlane",
			expected: "Welcome to Openlane, we're glad to have you!",
		},
		{
			name:     "multiple replacements",
			details:  "{{company_name}} is great! Visit {{company_name}} today.",
			orgName:  "Openlane",
			expected: "Openlane is great! Visit Openlane today.",
		},
		{
			name:     "empty organization name",
			details:  "Welcome to {{company_name}}!",
			orgName:  "",
			expected: "Welcome to [Company Name]!",
		},
		{
			name:     "no placeholders",
			details:  "Just a regular string with no placeholders.",
			orgName:  "Openlane",
			expected: "Just a regular string with no placeholders.",
		},
		{
			name:     "empty details",
			details:  "",
			orgName:  "Openlane",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := updatePlaceholderText(tt.details, tt.orgName)
			assert.Check(t, is.Equal(result, tt.expected))
		})
	}
}

// mockStatusMutation is a mock implementation of statusMutation for testing
type mockStatusMutation struct {
	op         ent.Op
	status     enums.DocumentStatus
	approverID string
	delegateID string
	client     *generated.Client
}

func (m *mockStatusMutation) Op() ent.Op         { return m.op }
func (m *mockStatusMutation) Type() string       { return "MockDocument" }
func (m *mockStatusMutation) ID() (string, bool) { return "test-id", true }
func (m *mockStatusMutation) IDs(ctx context.Context) ([]string, error) {
	return []string{"test-id"}, nil
}
func (m *mockStatusMutation) Field(name string) (ent.Value, bool)  { return nil, false }
func (m *mockStatusMutation) Fields() []string                     { return []string{} }
func (m *mockStatusMutation) ClearedFields() []string              { return []string{} }
func (m *mockStatusMutation) Status() (enums.DocumentStatus, bool) { return m.status, true }
func (m *mockStatusMutation) ApproverID() (string, bool)           { return m.approverID, m.approverID != "" }
func (m *mockStatusMutation) DelegateID() (string, bool)           { return m.delegateID, m.delegateID != "" }
func (m *mockStatusMutation) OldApproverID(ctx context.Context) (string, error) {
	return m.approverID, nil
}
func (m *mockStatusMutation) OldDelegateID(ctx context.Context) (string, error) {
	return m.delegateID, nil
}
func (m *mockStatusMutation) ApprovalRequired() (bool, bool)                        { return true, true }
func (m *mockStatusMutation) OldApprovalRequired(ctx context.Context) (bool, error) { return true, nil }
func (m *mockStatusMutation) Client() *generated.Client                             { return m.client }

func TestGetApproverDelegateIDs(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name               string
		op                 ent.Op
		approverID         string
		delegateID         string
		expectedApproverID string
		expectedDelegateID string
	}{
		{
			name:               "create operation with both IDs",
			op:                 ent.OpCreate,
			approverID:         "approver-123",
			delegateID:         "delegate-456",
			expectedApproverID: "approver-123",
			expectedDelegateID: "delegate-456",
		},
		{
			name:               "create operation with only approver",
			op:                 ent.OpCreate,
			approverID:         "approver-123",
			delegateID:         "",
			expectedApproverID: "approver-123",
			expectedDelegateID: "",
		},
		{
			name:               "update operation with both IDs",
			op:                 ent.OpUpdate,
			approverID:         "approver-789",
			delegateID:         "delegate-012",
			expectedApproverID: "approver-789",
			expectedDelegateID: "delegate-012",
		},
		{
			name:               "updateone operation with only delegate",
			op:                 ent.OpUpdateOne,
			approverID:         "",
			delegateID:         "delegate-345",
			expectedApproverID: "",
			expectedDelegateID: "delegate-345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mut := &mockStatusMutation{
				op:         tt.op,
				approverID: tt.approverID,
				delegateID: tt.delegateID,
			}

			approverID, delegateID := getApproverDelegateIDs(ctx, mut)

			assert.Check(t, is.Equal(approverID, tt.expectedApproverID))
			assert.Check(t, is.Equal(delegateID, tt.expectedDelegateID))
		})
	}
}
