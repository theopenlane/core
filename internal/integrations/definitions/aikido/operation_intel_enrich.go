package aikido

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/logx"
	"github.com/theopenlane/utils/mapx"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/vulnerability"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/celx"
	"github.com/theopenlane/core/v2/pkg/gala"
)

const (
	// aikidoIDPrefix marks Aikido's own advisories, the only ids enriched from Aikido Intel since vulnintel enriches CVE
	// and GHSA ids from OSV for every provider
	aikidoIDPrefix = "AIKIDO-"
	// intelEnrichMaxPerRun caps the Intel lookups one enrichment run makes, the rest wait for the next ingest
	intelEnrichMaxPerRun = 200
)

// intelEnrichResult reports the work one Intel enrichment run did
type intelEnrichResult struct {
	// Considered counts the vulnerabilities still missing Intel details
	Considered int `json:"considered"`
	// Updated counts the vulnerabilities that gained Intel details
	Updated int `json:"updated"`
}

// runIntelEnrich fills the Intel details of the installation's open Aikido advisories that have none yet, the published
// date marks an advisory as enriched so each run only looks up new ones and ones an earlier run could not fetch
func runIntelEnrich(ctx context.Context, request types.OperationRequest, client *APIClient, _ IntelEnrichSync) (json.RawMessage, error) {
	vulns, err := request.DB.Vulnerability.Query().
		Where(
			vulnerability.HasIntegrationsWith(integration.ID(request.Integration.ID)),
			vulnerability.OpenEQ(true),
			vulnerability.CveIDHasPrefix(aikidoIDPrefix),
			vulnerability.PublishedAtIsNil(),
		).
		Limit(intelEnrichMaxPerRun).
		All(ctx)
	if err != nil {
		return nil, err
	}

	ids := lo.Uniq(lo.Map(vulns, vulnerabilityCveID))

	details, err := client.fetchEach(ctx, ids, pathCVEFormat)
	if err != nil {
		return nil, err
	}

	result := intelEnrichResult{Considered: len(vulns)}

	for _, vuln := range vulns {
		update, changed := intelUpdate(request.DB.Vulnerability.UpdateOneID(vuln.ID), vuln, details[vuln.CveID])
		if !changed {
			continue
		}

		if err := update.Exec(ctx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("vulnerability_id", vuln.ID).Msg("aikido: failed to apply intel details")

			continue
		}

		result.Updated++
	}

	return providerkit.EncodeResult(result, ErrPayloadEncode)
}

// vulnerabilityCveID returns the advisory id of a vulnerability
func vulnerabilityCveID(vuln *generated.Vulnerability, _ int) string {
	return vuln.CveID
}

// intelUpdate applies Intel details to a vulnerability, its tldr replaces the description the issue sync leaves empty for
// Aikido advisories and the short description fills an empty summary, reporting whether anything changed
func intelUpdate(update *generated.VulnerabilityUpdateOne, vuln *generated.Vulnerability, details map[string]any) (*generated.VulnerabilityUpdateOne, bool) {
	if details == nil {
		return update, false
	}

	changed := false

	if tldr := strings.TrimSpace(mapx.Value[string](details, "tldr")); tldr != "" && tldr != vuln.Description {
		update.SetDescription(celx.Paragraphs(tldr))
		changed = true
	}

	if short := mapx.Value[string](details, "short_description"); vuln.Summary == "" && short != "" {
		update.SetSummary(short)
		changed = true
	}

	if published, ok := intelPublished(mapx.Value[string](details, "published")); ok {
		update.SetPublishedAt(models.DateTime(published))
		changed = true
	}

	return update, changed
}

// intelPublished parses the Intel publish date, which Aikido sends as either a date or a full RFC 3339 timestamp
func intelPublished(value string) (time.Time, bool) {
	for _, layout := range []string{time.RFC3339, time.DateOnly} {
		if published, err := time.Parse(layout, value); err == nil {
			return published, true
		}
	}

	return time.Time{}, false
}

// intelEnrichListener dispatches the Intel enrichment after an Aikido installation's run persists vulnerabilities, so the
// paced Intel lookups run as their own job instead of holding up the issue sync
func intelEnrichListener() types.GalaListenerRegistration {
	return types.GalaListenerRegistration{
		Name: "aikido.intel.enrich",
		Register: func(g *gala.Gala, services types.RuntimeServices) ([]gala.ListenerID, error) {
			dispatcher := intelEnrichDispatcher{services: services}

			return gala.Register(g, gala.Definition[operations.VulnerabilitiesIngestedEnvelope]{
				Topic:  operations.VulnerabilitiesIngestedTopic,
				Handle: dispatcher.handle,
			})
		},
	}
}

// intelEnrichDispatcher queues the Intel enrichment operation for Aikido installations
type intelEnrichDispatcher struct {
	// services dispatches operations through the runtime
	services types.RuntimeServices
}

// handle dispatches the Intel enrichment when the ingest came from an Aikido installation
func (d intelEnrichDispatcher) handle(hc gala.HandlerContext, envelope operations.VulnerabilitiesIngestedEnvelope) error {
	if envelope.DefinitionID != definitionID.ID() {
		return nil
	}

	ctx := auth.EnsureIntegrationCaller(hc.Context, envelope.OwnerID)

	_, err := d.services.Dispatch(ctx, types.DispatchRequest{
		IntegrationID: envelope.IntegrationID,
		Operation:     intelEnrichOp.Name(),
		RunType:       enums.IntegrationRunTypeEvent,
	})

	return err
}
