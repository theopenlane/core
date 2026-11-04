//go:build test

package graphapi

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	slackdef "github.com/theopenlane/core/v2/internal/integrations/definitions/slack"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	intr "github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// newTestRuntime builds a lightweight runtime with the Slack mock registered,
// suitable for unit tests that only need catalog/definition lookups
func newTestRuntime(t *testing.T) (*intr.Runtime, *slackdef.MockSlackRuntime) {
	t.Helper()

	mock := slackdef.NewMockSlackRuntime()
	t.Cleanup(mock.Close)

	def, err := mock.Builder()()
	assert.NilError(t, err)

	reg := registry.New()
	assert.NilError(t, reg.Register(def))

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
					Name:        "message.send",
					Description: "Send a message",
					Input:       types.InputRegistration{Schema: json.RawMessage(`{"type":"object"}`)},
				},
			},
			wantLen:  1,
			wantName: "message.send",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := buildOperationEntries(tt.ops)
			assert.Check(t, is.Len(entries, tt.wantLen))

			if tt.wantName != "" && len(entries) > 0 {
				assert.Check(t, is.Equal(tt.wantName, entries[0].Name))
				assert.Check(t, is.Equal(tt.ops[0].Description, entries[0].Description))
				assert.Check(t, is.DeepEqual(unmarshalJSONAny(t, string(tt.ops[0].Input.Schema)), unmarshalJSONAny(t, string(entries[0].ConfigSchema))))
			}
		})
	}
}

func TestWorkflowMetadataExtensions_NilRuntime(t *testing.T) {
	result := workflowMetadataExtensions(context.Background(), nil, nil)
	assert.Assert(t, result != nil)

	integrations, ok := result["integrations"].(map[string]any)
	assert.Assert(t, ok)

	contract, ok := integrations["action_contract"].(map[string]any)
	assert.Assert(t, ok)

	assert.Check(t, len(contract["target_selector"].([]string)) > 0)
	assert.Check(t, len(contract["operation_selector"].([]string)) > 0)
	assert.Check(t, len(contract["scope_fields"].([]string)) > 0)
	assert.Check(t, len(contract["scope_variables"].([]string)) > 0)
	assert.Check(t, len(contract["run_types"].([]string)) > 0)

	providers, ok := integrations["providers"].([]integrationProviderExtensions)
	assert.Assert(t, ok)
	assert.Check(t, is.Len(providers, 0), "nil runtime should produce no providers")
}

func TestResolveOrgIntegrationAvailability_NilDB(t *testing.T) {
	result := resolveOrgIntegrationAvailability(context.Background(), nil)
	assert.Check(t, is.Nil(result), "nil db should return nil availability")
}

func TestResolveOrgIntegrationAvailability_NoCallerContext(t *testing.T) {
	result := resolveOrgIntegrationAvailability(context.Background(), &ent.Client{})
	assert.Check(t, is.Nil(result), "unauthenticated context should return nil availability")
}

func TestIntegrationWorkflowProviders_CatalogOnly(t *testing.T) {
	rt, _ := newTestRuntime(t)

	providers := integrationWorkflowProviders(context.Background(), rt, nil)
	assert.Assert(t, is.Len(providers, 1))

	slack := providers[0]
	assert.Check(t, is.Equal(slackdef.DefinitionID.ID(), slack.Provider))
	assert.Check(t, slack.DisplayName != "")
	assert.Check(t, len(slack.Operations) > 0)
	assert.Check(t, !slack.Available, "no DB means no availability data, Available should be false")
	assert.Check(t, is.Len(slack.Installations, 0))

	opNames := lo.Map(slack.Operations, func(op integrationOperationEntry, _ int) string {
		return op.Name
	})
	assert.Check(t, is.Contains(opNames, slackdef.MessageSendOp.Name()))
	assert.Check(t, slices.IsSorted(opNames), "operations should be sorted alphabetically")
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

			assert.Check(t, is.Equal(tt.wantAvailable, entry.Available))
			assert.Check(t, is.Len(entry.Installations, tt.wantInstallations))

			var totalTemplates int
			for _, op := range enriched {
				totalTemplates += len(op.Templates)
			}

			assert.Check(t, is.Equal(tt.wantTemplates, totalTemplates))
		})
	}
}

func unmarshalJSONAny(t *testing.T, raw string) any {
	t.Helper()

	var value any

	assert.NilError(t, json.Unmarshal([]byte(raw), &value))

	return value
}
