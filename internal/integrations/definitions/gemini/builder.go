package gemini

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// Builder returns the Gemini definition builder with the supplied runtime config applied.
// The definition is currently runtime-only
func Builder(runtime *RuntimeConfig) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		def := types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Family:      "Google",
				DisplayName: "Gemini",
				Description: "Google Gemini models for AI-assisted analysis, extraction, and content generation.",
				Category:    "ai",
				Tags:        []string{"ai", "reports"},
				Active:      true,
				Visible:     false,
			},
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: runtimeGeminiSchema,
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         geminiClient.ID(),
					Description: "Gemini API client, only provisioned through the runtime config",
					Build:       runtimeOnlyClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:           ReportScanRequestOp.Name(),
					Description:    "Schedule the per-section parse jobs for a pending report scan",
					Topic:          DefinitionID.OperationTopic(ReportScanRequestOp.Name()),
					ClientRef:      geminiClient.ID(),
					ConfigSchema:   reportScanRequestSchema,
					Policy:         types.ExecutionPolicy{SkipRunRecord: true},
					DisabledForAll: !runtime.Provisioned(),
					// RateLimit:          &types.RateLimitPolicy{Window: SubmitInterval},
					Handle:             ReportScanRequest{parts: runtime.Prompts.SOC2.ConfiguredParts()}.Handle(),
					CustomerSelectable: lo.ToPtr(false),
					Internal:           true,
				},
				{
					Name:               ReportScanPartOp.Name(),
					Description:        "Parse one section of the report attached to a report scan",
					Topic:              DefinitionID.OperationTopic(ReportScanPartOp.Name()),
					ClientRef:          geminiClient.ID(),
					ConfigSchema:       reportScanPartSchema,
					Policy:             types.ExecutionPolicy{SkipRunRecord: true},
					DisabledForAll:     !runtime.Provisioned(),
					Handle:             ReportScanPartRequest{}.Handle(),
					CustomerSelectable: lo.ToPtr(false),
					Internal:           true,
				},
				{
					Name:               ReportScanReleaseOp.Name(),
					Description:        "Release the cached report once a report scan finalizes",
					Topic:              DefinitionID.OperationTopic(ReportScanReleaseOp.Name()),
					ClientRef:          geminiClient.ID(),
					ConfigSchema:       reportScanReleaseSchema,
					Policy:             types.ExecutionPolicy{SkipRunRecord: true},
					DisabledForAll:     !runtime.Provisioned(),
					Handle:             ReportScanReleaseRequest{}.Handle(),
					CustomerSelectable: lo.ToPtr(false),
					Internal:           true,
				},
			},
			GalaListeners: []types.GalaListenerRegistration{
				reportScanListeners(),
			},
		}

		if runtime.Provisioned() {
			runtimeGeminiRef.SetConfig(runtime)

			marshaledConfig, err := runtimeGeminiRef.MarshalConfig()
			if err != nil {
				return types.Definition{}, fmt.Errorf("%w: %w", ErrRuntimeConfigDecode, err)
			}

			def.RuntimeIntegration = &types.RuntimeIntegrationRegistration{
				Ref:    runtimeGeminiRef.ID(),
				Schema: runtimeGeminiSchema,
				Config: marshaledConfig,
				Build:  runtimeClientBuilder(),
			}
		}

		return def, nil
	})
}
