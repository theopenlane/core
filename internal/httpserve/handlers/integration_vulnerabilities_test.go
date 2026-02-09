package handlers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/integrations/types"
)

// TestExtractAlertEnvelopes verifies typed alert envelope extraction.
func TestExtractAlertEnvelopes(t *testing.T) {
	payload, err := json.Marshal(map[string]any{"id": 1})
	require.NoError(t, err)

	alerts := []types.AlertEnvelope{
		{AlertType: "dependabot", Resource: "octo/repo", Payload: payload},
	}

	result := types.OperationResult{
		Details: map[string]any{"alerts": alerts},
	}

	got := extractAlertEnvelopes(result)
	require.Len(t, got, 1)
	require.Equal(t, "dependabot", got[0].AlertType)
	require.Equal(t, "octo/repo", got[0].Resource)
	require.Equal(t, json.RawMessage(payload), got[0].Payload)
}

// TestExtractAlertEnvelopes_MapInput verifies map-based alert extraction.
func TestExtractAlertEnvelopes_MapInput(t *testing.T) {
	payload, err := json.Marshal(map[string]any{"id": 2})
	require.NoError(t, err)

	result := types.OperationResult{
		Details: map[string]any{
			"alerts": []any{
				map[string]any{
					"alertType": "code_scanning",
					"resource":  "octo/repo",
					"payload":   payload,
				},
			},
		},
	}

	got := extractAlertEnvelopes(result)
	require.Len(t, got, 1)
	require.Equal(t, "code_scanning", got[0].AlertType)
	require.Equal(t, "octo/repo", got[0].Resource)
	require.Equal(t, json.RawMessage(payload), got[0].Payload)
}

// TestBuildGitHubVulnerability_Dependabot verifies dependabot mapping.
func TestBuildGitHubVulnerability_Dependabot(t *testing.T) {
	payload := map[string]any{
		"id":         42,
		"state":      "open",
		"html_url":   "https://github.com/octo/repo/security/dependabot/42",
		"created_at": "2025-01-02T03:04:05Z",
		"updated_at": "2025-01-03T03:04:05Z",
		"security_advisory": map[string]any{
			"summary":      "Test advisory",
			"description":  "Details",
			"severity":     "high",
			"published_at": "2025-01-01T00:00:00Z",
		},
		"security_vulnerability": map[string]any{
			"severity": "high",
		},
		"dependency": map[string]any{
			"package": map[string]any{"name": "lodash"},
		},
	}

	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	alert := types.AlertEnvelope{
		AlertType: "dependabot",
		Resource:  "octo/repo",
		Payload:   raw,
	}

	vuln, ok := buildGitHubVulnerability(alert, "github_app")
	require.True(t, ok)
	require.Equal(t, "dependabot:octo/repo:42", vuln.externalID)
	require.Equal(t, "octo", vuln.externalOwnerID)
	require.Equal(t, "lodash", vuln.displayName)
	require.Equal(t, "Test advisory", vuln.summary)
	require.Equal(t, "high", vuln.severity)
	require.NotNil(t, vuln.open)
	require.True(t, *vuln.open)
	require.Equal(t, payload["html_url"], vuln.externalURI)
	require.NotNil(t, vuln.discoveredAt)
	require.NotNil(t, vuln.sourceUpdatedAt)
}
