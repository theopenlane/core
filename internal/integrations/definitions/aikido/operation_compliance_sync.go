package aikido

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// pathReportPage is the app path of a framework's compliance report page, followed by the framework key
const pathReportPage = "/reports/"

// complianceFramework is one Aikido compliance report collected as check results
type complianceFramework struct {
	// key is the report path segment
	key string
	// standard is the framework name tagged on its check results
	standard string
}

// complianceFrameworks are the Aikido compliance reports sharing the overview shape
var complianceFrameworks = []complianceFramework{
	{key: "soc2", standard: "SOC 2"},
	{key: "iso", standard: "ISO 27001"},
	{key: "pci", standard: "PCI DSS"},
	{key: "gdpr", standard: "GDPR"},
	{key: "nis2", standard: "NIS2"},
	{key: "dora", standard: "DORA"},
	{key: "cis", standard: "CIS"},
}

// complianceOverview is one framework report, grouping rule results by control under each section
type complianceOverview struct {
	// Overview maps each section key to its controls
	Overview map[string][]complianceControl `json:"overview"`
}

// complianceControl is one framework control and the Aikido rules evaluated for it
type complianceControl struct {
	// Code is the Aikido control code, such as cc3.3 or 1.2_network_security_controls_are_configured
	Code string `json:"code"`
	// Title is the control title, which carries the ISO 27001 reference such as A.8.2
	Title string `json:"title"`
	// Results are the rules evaluated for the control
	Results []complianceResult `json:"results"`
}

// complianceResult is one Aikido rule evaluated for a control
type complianceResult struct {
	// Measure is the measure key the rule belongs to
	Measure string `json:"measure"`
	// Title is the rule title
	Title string `json:"title"`
	// MeasureTitle is the measure description
	MeasureTitle string `json:"measure_title"`
	// Status is the rule status, such as complying, failing or disabled
	Status string `json:"status"`
}

// complianceCheck is the payload of one rule result ingested as a check result
type complianceCheck struct {
	// Framework is the report the rule was evaluated in
	Framework string `json:"framework"`
	// Standard is the framework name
	Standard string `json:"standard"`
	// Section is the overview section key
	Section string `json:"section"`
	// ControlCode is the Aikido control code
	ControlCode string `json:"control_code"`
	// ControlTitle is the control title
	ControlTitle string `json:"control_title"`
	// RefCode is the control reference matched against Openlane control ref codes
	RefCode string `json:"ref_code"`
	// ReportURL is the framework's report page in Aikido, which has no anchor per rule
	ReportURL string `json:"report_url"`
	complianceResult
}

// runComplianceSync collects the Aikido compliance reports as check results
func runComplianceSync(ctx context.Context, _ types.OperationRequest, client *APIClient, _ ComplianceSync) ([]types.IngestPayloadSet, error) {
	return CollectCompliance(ctx, client)
}

// CollectCompliance emits a check result per rule in every compliance report, skipping a report the workspace cannot return
func CollectCompliance(ctx context.Context, c *APIClient) ([]types.IngestPayloadSet, error) {
	envelopes := make([]types.MappingEnvelope, 0)

	for _, framework := range complianceFrameworks {
		overview, err := c.complianceOverview(ctx, framework.key)
		if err != nil {
			if errors.Is(err, ErrUnauthorized) || ctx.Err() != nil {
				return nil, fmt.Errorf("%w: %s: %w", ErrComplianceFetchFailed, framework.key, err)
			}

			logx.FromContext(ctx).Warn().Err(err).Str("framework", framework.key).Msg("aikido: unable to fetch compliance report, skipping it")

			continue
		}

		for section, controls := range overview.Overview {
			for _, control := range controls {
				for _, result := range control.Results {
					check := complianceCheck{
						Framework:        framework.key,
						Standard:         framework.standard,
						Section:          section,
						ControlCode:      control.Code,
						ControlTitle:     control.Title,
						RefCode:          complianceRefCode(control),
						ReportURL:        c.appURL + pathReportPage + framework.key,
						complianceResult: result,
					}

					envelope, err := providerkit.MarshalEnvelope(control.Title, check, ErrPayloadEncode)
					if err != nil {
						logx.FromContext(ctx).Error().Err(err).Str("framework", framework.key).Str("control", control.Code).Msg("aikido: failed to marshal compliance check envelope")

						return nil, err
					}

					envelopes = append(envelopes, envelope)
				}
			}
		}
	}

	return []types.IngestPayloadSet{
		{
			Schema:    entityops.SchemaCheckResult.Name,
			Envelopes: envelopes,
		},
	}, nil
}

// complianceRefCode derives the framework control reference, taken from an ISO style "A.8.2 - title"
// control title, else the code up to its first underscore, upper cased so cc3.3 becomes CC3.3
func complianceRefCode(control complianceControl) string {
	if ref, _, found := strings.Cut(control.Title, " - "); found {
		return strings.TrimSpace(ref)
	}

	code, _, _ := strings.Cut(control.Code, "_")

	return strings.ToUpper(code)
}

// complianceOverview fetches one framework report, which the API returns either bare or as the only element of an array
func (c *APIClient) complianceOverview(ctx context.Context, framework string) (complianceOverview, error) {
	var raw json.RawMessage

	if _, err := c.get(ctx, fmt.Sprintf(pathComplianceOverviewFormat, framework), nil, &raw); err != nil {
		return complianceOverview{}, err
	}

	var reports []complianceOverview
	if err := json.Unmarshal(raw, &reports); err == nil {
		if len(reports) == 0 {
			return complianceOverview{}, nil
		}

		return reports[0], nil
	}

	var report complianceOverview
	if err := json.Unmarshal(raw, &report); err != nil {
		return complianceOverview{}, fmt.Errorf("%w: %w", ErrAPIRequest, err)
	}

	return report, nil
}
