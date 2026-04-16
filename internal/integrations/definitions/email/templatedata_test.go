package email

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTemplateData_ConfigFieldsSerialized(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "TestCo",
		FromEmail:   "noreply@test.com",
		ProductURL:  "https://app.test.com",
	}

	data, err := buildTemplateData(cfg, nil)
	require.NoError(t, err)

	assert.Equal(t, "TestCo", data["companyName"])
	assert.Equal(t, "noreply@test.com", data["fromEmail"])
	assert.Equal(t, "https://app.test.com", data["productURL"])
}

func TestBuildTemplateData_YearInjected(t *testing.T) {
	data, err := buildTemplateData(RuntimeEmailConfig{}, nil)
	require.NoError(t, err)

	assert.Equal(t, time.Now().Year(), data["year"])
}

func TestBuildTemplateData_VarsOverrideConfig(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "Default",
	}

	vars := map[string]any{
		"companyName": "Override",
	}

	data, err := buildTemplateData(cfg, vars)
	require.NoError(t, err)

	assert.Equal(t, "Override", data["companyName"])
}

func TestBuildTemplateData_VarsMergedWithConfig(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "TestCo",
	}

	vars := map[string]any{
		"recipientFirstName": "Alice",
		"customField":        "custom-value",
	}

	data, err := buildTemplateData(cfg, vars)
	require.NoError(t, err)

	assert.Equal(t, "TestCo", data["companyName"])
	assert.Equal(t, "Alice", data["recipientFirstName"])
	assert.Equal(t, "custom-value", data["customField"])
}

func TestBuildTemplateData_EmptyVars(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "TestCo",
	}

	data, err := buildTemplateData(cfg, map[string]any{})
	require.NoError(t, err)

	assert.Equal(t, "TestCo", data["companyName"])
}

func TestBuildTemplateData_NilVars(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "TestCo",
	}

	data, err := buildTemplateData(cfg, nil)
	require.NoError(t, err)

	assert.Equal(t, "TestCo", data["companyName"])
}
