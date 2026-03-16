package operations

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestResolveInstallationFilterExprPreservesExactValue(t *testing.T) {
	t.Parallel()

	expr := " payload.labels.env == \"prod\" "
	raw, err := json.Marshal(installationFilterConfig{FilterExpr: expr})
	require.NoError(t, err)

	installation := &ent.Integration{
		Config: types.IntegrationConfig{
			ClientConfig: raw,
		},
	}

	got, err := resolveInstallationFilterExpr(installation)
	require.NoError(t, err)
	require.Equal(t, expr, got)
}

func TestProcessPayloadSetsWithOptionsAppliesInstallationFilterExpr(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	definition := types.Definition{
		DefinitionSpec: types.DefinitionSpec{
			ID:          "def_test_installation_filter",
			Slug:        "test-installation-filter",
			DisplayName: "Test Installation Filter",
			Active:      true,
			Visible:     true,
		},
		Mappings: []types.MappingRegistration{
			{
				Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
				Spec: types.MappingOverride{
					FilterExpr: "true",
					MapExpr:    "{",
				},
			},
		},
	}
	require.NoError(t, reg.Register(definition))

	raw, err := json.Marshal(installationFilterConfig{FilterExpr: "false"})
	require.NoError(t, err)

	installation := &ent.Integration{
		DefinitionID: definition.ID,
		Config: types.IntegrationConfig{
			ClientConfig: raw,
		},
	}

	payloadSets := []types.IngestPayloadSet{
		{
			Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
			Envelopes: []types.MappingEnvelope{
				{Payload: json.RawMessage(`{"labels":{"env":"prod"}}`)},
			},
		},
	}

	err = ProcessPayloadSetsWithOptions(
		context.Background(),
		reg,
		&ent.Client{},
		installation,
		[]types.IngestContract{{Schema: integrationgenerated.IntegrationMappingSchemaVulnerability}},
		payloadSets,
		IngestOptions{},
	)
	require.NoError(t, err)
}

func TestProcessPayloadSetsWithOptionsRejectsInvalidInstallationFilterConfig(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	definition := types.Definition{
		DefinitionSpec: types.DefinitionSpec{
			ID:          "def_test_invalid_installation_filter_config",
			Slug:        "test-invalid-installation-filter-config",
			DisplayName: "Test Invalid Installation Filter Config",
			Active:      true,
			Visible:     true,
		},
	}
	require.NoError(t, reg.Register(definition))

	installation := &ent.Integration{
		DefinitionID: definition.ID,
		Config: types.IntegrationConfig{
			ClientConfig: json.RawMessage(`{`),
		},
	}

	err := ProcessPayloadSetsWithOptions(context.Background(), reg, &ent.Client{}, installation, nil, nil, IngestOptions{})
	require.ErrorIs(t, err, ErrIngestInstallationFilterConfigInvalid)
}

func TestIngestRecordReturnsStaticFilterError(t *testing.T) {
	t.Parallel()

	definition := types.Definition{
		Mappings: []types.MappingRegistration{
			{
				Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
				Spec: types.MappingOverride{
					FilterExpr: "payload[",
				},
			},
		},
	}

	err := ingestRecord(
		context.Background(),
		nil,
		&ent.Integration{},
		definition,
		integrationgenerated.IntegrationMappingSchemaVulnerability,
		types.MappingEnvelope{Payload: json.RawMessage(`{"labels":{"env":"prod"}}`)},
		"",
		"",
	)
	require.ErrorIs(t, err, ErrIngestFilterFailed)
}
