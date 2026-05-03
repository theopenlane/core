//go:build test

package graphapi

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ent "github.com/theopenlane/core/internal/ent/generated"
	slackdef "github.com/theopenlane/core/internal/integrations/definitions/slack"
	"github.com/theopenlane/core/internal/integrations/registry"
	intr "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
)

// newTestRuntime builds a lightweight runtime with the Slack mock registered,
// suitable for unit tests that only need catalog/definition lookups
func newTestRuntime(t *testing.T) (*intr.Runtime, *slackdef.MockSlackRuntime) {
	t.Helper()

	mock := slackdef.NewMockSlackRuntime()
	t.Cleanup(mock.Close)

	def, err := mock.Builder()()
	require.NoError(t, err)

	reg := registry.New()
	require.NoError(t, reg.Register(def))

	return intr.NewForTesting(reg), mock
}

func TestBuildOperationEntries(t *testing.T) {
	tests := []struct {
		name     string
		ops      []types.OperationRegistration
		wantLen  int
		wantName string
	}{
		{
			name:    "empty operations",
			ops:     nil,
			wantLen: 0,
		},
		{
			name: "single operation with schema",
			ops: []types.OperationRegistration{
				{
					Name:         "message.send",
					Description:  "Send a message",
					ConfigSchema: json.RawMessage(`{"type":"object"}`),
				},
			},
			wantLen:  1,
			wantName: "message.send",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := buildOperationEntries(tt.ops)
			assert.Len(t, entries, tt.wantLen)

			if tt.wantName != "" && len(entries) > 0 {
				assert.Equal(t, tt.wantName, entries[0].Name)
				assert.Equal(t, tt.ops[0].Description, entries[0].Description)
				assert.JSONEq(t, string(tt.ops[0].ConfigSchema), string(entries[0].ConfigSchema))
			}
		})
	}
}

func TestWorkflowMetadataExtensions_NilRuntime(t *testing.T) {
	result := workflowMetadataExtensions(context.Background(), nil, nil)
	require.NotNil(t, result)

	integrations, ok := result["integrations"].(map[string]any)
	require.True(t, ok)

	contract, ok := integrations["action_contract"].(map[string]any)
	require.True(t, ok)

	assert.NotEmpty(t, contract["target_selector"])
	assert.NotEmpty(t, contract["operation_selector"])
	assert.NotEmpty(t, contract["scope_fields"])
	assert.NotEmpty(t, contract["scope_variables"])
	assert.NotEmpty(t, contract["run_types"])

	providers, ok := integrations["providers"].([]integrationProviderExtensions)
	require.True(t, ok)
	assert.Empty(t, providers, "nil runtime should produce no providers")
}

func TestResolveOrgIntegrationAvailability_NilDB(t *testing.T) {
	result := resolveOrgIntegrationAvailability(context.Background(), nil)
	assert.Nil(t, result, "nil db should return nil availability")
}

func TestResolveOrgIntegrationAvailability_NoCallerContext(t *testing.T) {
	result := resolveOrgIntegrationAvailability(context.Background(), &ent.Client{})
	assert.Nil(t, result, "unauthenticated context should return nil availability")
}

func TestIntegrationWorkflowProviders_CatalogOnly(t *testing.T) {
	rt, _ := newTestRuntime(t)

	providers := integrationWorkflowProviders(context.Background(), rt, nil)
	require.Len(t, providers, 1)

	slack := providers[0]
	assert.Equal(t, slackdef.DefinitionID.ID(), slack.Provider)
	assert.NotEmpty(t, slack.DisplayName)
	assert.NotEmpty(t, slack.Operations)
	assert.False(t, slack.Available, "no DB means no availability data, Available should be false")
	assert.Empty(t, slack.Installations)

	opNames := lo.Map(slack.Operations, func(op integrationOperationEntry, _ int) string {
		return op.Name
	})
	assert.Contains(t, opNames, slackdef.MessageSendOp.Name())
	assert.True(t, slices.IsSorted(opNames), "operations should be sorted alphabetically")
}

func TestApplyProviderAvailability(t *testing.T) {
	defID := slackdef.DefinitionID.ID()
	opName := slackdef.MessageSendOp.Name()

	tests := []struct {
		name              string
		availability      *orgIntegrationAvailability
		wantAvailable     bool
		wantTemplates     int
		wantInstallations int
	}{
		{
			name: "installations only, no templates",
			availability: &orgIntegrationAvailability{
				installationsByDefinition: map[string][]*ent.Integration{
					defID: {{ID: "inst-1", Name: "Test Slack"}},
				},
				templatesByTopicPattern: map[string][]*ent.NotificationTemplate{},
			},
			wantAvailable:     false,
			wantInstallations: 1,
			wantTemplates:     0,
		},
		{
			name: "templates only, no installations",
			availability: &orgIntegrationAvailability{
				installationsByDefinition: map[string][]*ent.Integration{},
				templatesByTopicPattern: map[string][]*ent.NotificationTemplate{
					opName: {{ID: "tpl-1", Key: "slack-notify", Name: "Slack Notify"}},
				},
			},
			wantAvailable:     false,
			wantInstallations: 0,
			wantTemplates:     1,
		},
		{
			name: "both installations and templates present",
			availability: &orgIntegrationAvailability{
				installationsByDefinition: map[string][]*ent.Integration{
					defID: {{ID: "inst-1", Name: "Test Slack"}},
				},
				templatesByTopicPattern: map[string][]*ent.NotificationTemplate{
					opName: {{ID: "tpl-1", Key: "slack-notify", Name: "Slack Notify"}},
				},
			},
			wantAvailable:     true,
			wantInstallations: 1,
			wantTemplates:     1,
		},
		{
			name: "installation for wrong definition",
			availability: &orgIntegrationAvailability{
				installationsByDefinition: map[string][]*ent.Integration{
					"def_some_other_provider": {{ID: "inst-1", Name: "Other"}},
				},
				templatesByTopicPattern: map[string][]*ent.NotificationTemplate{
					opName: {{ID: "tpl-1", Key: "slack-notify", Name: "Slack Notify"}},
				},
			},
			wantAvailable:     false,
			wantInstallations: 0,
			wantTemplates:     1,
		},
		{
			name: "template for non-matching operation",
			availability: &orgIntegrationAvailability{
				installationsByDefinition: map[string][]*ent.Integration{
					defID: {{ID: "inst-1", Name: "Test Slack"}},
				},
				templatesByTopicPattern: map[string][]*ent.NotificationTemplate{
					"NonexistentOperation": {{ID: "tpl-1", Key: "notify", Name: "Notify"}},
				},
			},
			wantAvailable:     false,
			wantInstallations: 1,
			wantTemplates:     0,
		},
		{
			name: "multiple installations and templates",
			availability: &orgIntegrationAvailability{
				installationsByDefinition: map[string][]*ent.Integration{
					defID: {
						{ID: "inst-1", Name: "Slack Workspace A"},
						{ID: "inst-2", Name: "Slack Workspace B"},
					},
				},
				templatesByTopicPattern: map[string][]*ent.NotificationTemplate{
					opName: {
						{ID: "tpl-1", Key: "alert", Name: "Alert"},
						{ID: "tpl-2", Key: "digest", Name: "Daily Digest"},
					},
				},
			},
			wantAvailable:     true,
			wantInstallations: 2,
			wantTemplates:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operations := []integrationOperationEntry{
				{Name: opName, Description: "Send a message"},
				{Name: "HealthCheck", Description: "Health check"},
			}

			entry := integrationProviderExtensions{
				Provider: defID,
			}

			enriched := applyProviderAvailability(&entry, defID, operations, tt.availability)

			assert.Equal(t, tt.wantAvailable, entry.Available)
			assert.Len(t, entry.Installations, tt.wantInstallations)

			var totalTemplates int
			for _, op := range enriched {
				totalTemplates += len(op.Templates)
			}

			assert.Equal(t, tt.wantTemplates, totalTemplates)
		})
	}
}

