package gemini

import (
	"fmt"

	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/utils/jsonx"
)

// Builder returns the Gemini definition builder with the supplied runtime config applied.
// The definition is currently runtime-only
func Builder(runtime *RuntimeConfig) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		disabled := !runtime.Provisioned()
		request := ReportScanRequest{parts: runtime.Prompts.SOC2.ConfiguredParts()}

		def := types.Definition{
			ID:          DefinitionID.ID(),
			Family:      "Google",
			DisplayName: "Gemini",
			Description: "Google Gemini models for AI-assisted analysis, extraction, and content generation.",
			Category:    "ai",
			Tags:        []string{"ai", "reports"},
			Active:      true,
			Visible:     false,
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: jsonx.SchemaFrom[RuntimeConfig](),
			},
			Operations: []types.OperationRegistration{
				ReportScanRequestOp.Handles(request.runEncoded).
					// RateLimit(types.RateLimitPolicy{Window: submitInterval}).
					DisabledForAll(disabled).
					Description("Schedule the per-section parse jobs for a pending report scan").
					Registration(),
				VendorReportScanRequestOp.Handles(request.handleVendor).
					// RateLimit(types.RateLimitPolicy{Window: vendorReviewInterval}).
					DisabledForAll(disabled).
					Description("Schedule the parse jobs for a pending review of a vendor's report").
					Registration(),
				ReportScanPartOp.DisabledForAll(disabled).Description("Parse one section of the report attached to a report scan").Registration(),
				ReportScanReleaseOp.DisabledForAll(disabled).Description("Release the cached report once a report scan finalizes").Registration(),
				ReportScanImportOp.Description("Import the items a reviewer accepted from a parsed report").Registration(),
			},
			Mappings: reportScanImportMappings(),
			GalaListeners: []types.GalaListenerRegistration{
				reportScanListeners(),
			},
			RuntimeIntegration: &types.RuntimeIntegrationRegistration{
				Schema: jsonx.SchemaFrom[RuntimeConfig](),
				Build:  runtimeClientBuilder(),
			},
		}

		if runtime.Provisioned() {
			config, err := jsonx.ToRawMessage(runtime)
			if err != nil {
				return types.Definition{}, fmt.Errorf("%w: %w", ErrRuntimeConfigDecode, err)
			}

			def.RuntimeIntegration.Config = config
		}

		return def, nil
	})
}
