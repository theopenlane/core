package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/vulnerability"
)

// githubVulnerability models normalized GitHub alert fields for persistence.
type githubVulnerability struct {
	// externalID is the unique identifier used for deduplication.
	externalID string
	// externalOwnerID is the owning org/user identifier.
	externalOwnerID string
	// displayName is the user-friendly alert title.
	displayName string
	// summary is the short description of the alert.
	summary string
	// description is the long-form description of the alert.
	description string
	// severity is the alert severity label.
	severity string
	// status is the GitHub alert state.
	status string
	// category identifies the alert category.
	category string
	// externalURI is the URL to the alert in GitHub.
	externalURI string
	// open indicates whether the alert is open.
	open *bool
	// publishedAt is the advisory publish time.
	publishedAt *time.Time
	// discoveredAt is when the alert was created.
	discoveredAt *time.Time
	// sourceUpdatedAt is when the alert was last updated.
	sourceUpdatedAt *time.Time
	// metadata carries additional alert metadata.
	metadata map[string]any
	// rawPayload stores the raw alert payload for auditing.
	rawPayload map[string]any
}

// persistIntegrationVulnerabilities upserts vulnerability records from operation results.
func (h *Handler) persistIntegrationVulnerabilities(ctx context.Context, orgID string, provider types.ProviderType, result types.OperationResult) (map[string]any, error) {
	summary := map[string]any{
		"created": 0,
		"updated": 0,
		"skipped": 0,
	}

	if h == nil || h.DBClient == nil {
		return summary, errDBClientNotConfigured
	}

	if strings.TrimSpace(orgID) == "" {
		return summary, ErrMissingOrganizationContext
	}

	alerts := extractAlertEnvelopes(result)
	summary["total"] = len(alerts)
	if len(alerts) == 0 {
		return summary, nil
	}

	integrationRecord, err := h.DBClient.Integration.Query().
		Where(
			integration.OwnerIDEQ(orgID),
			integration.KindEQ(string(provider)),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return summary, ErrIntegrationNotFound
		}
		return summary, err
	}

	for _, alert := range alerts {
		vuln, ok := buildGitHubVulnerability(alert, string(provider))
		if !ok {
			summary["skipped"] = summary["skipped"].(int) + 1
			continue
		}

		existing, err := h.DBClient.Vulnerability.Query().
			Where(
				vulnerability.OwnerIDEQ(orgID),
				vulnerability.ExternalIDEQ(vuln.externalID),
			).
			Only(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return summary, err
			}

			create := h.DBClient.Vulnerability.Create().
				SetOwnerID(orgID).
				SetExternalID(vuln.externalID).
				AddIntegrationIDs(integrationRecord.ID)
			applyVulnerabilityCreate(create, vuln, provider)
			if err := create.Exec(ctx); err != nil {
				return summary, err
			}

			summary["created"] = summary["created"].(int) + 1
			continue
		}

		update := existing.Update().AddIntegrationIDs(integrationRecord.ID)
		applyVulnerabilityUpdate(update, vuln, provider)
		if err := update.Exec(ctx); err != nil {
			return summary, err
		}

		summary["updated"] = summary["updated"].(int) + 1
	}

	return summary, nil
}

// extractAlertEnvelopes normalizes alert envelopes from operation result details.
func extractAlertEnvelopes(result types.OperationResult) []types.AlertEnvelope {
	if result.Details == nil {
		return nil
	}

	raw, ok := result.Details["alerts"]
	if !ok || raw == nil {
		return nil
	}

	switch v := raw.(type) {
	case []types.AlertEnvelope:
		return v
	case []any:
		alerts := make([]types.AlertEnvelope, 0, len(v))
		for _, item := range v {
			switch alert := item.(type) {
			case types.AlertEnvelope:
				alerts = append(alerts, alert)
			case map[string]any:
				parsed := types.AlertEnvelope{
					AlertType: strings.TrimSpace(stringifyValue(alert["alertType"])),
					Resource:  strings.TrimSpace(stringifyValue(alert["resource"])),
					Action:    strings.TrimSpace(stringifyValue(alert["action"])),
				}
				if payload, ok := alert["payload"]; ok {
					parsed.Payload = rawMessageFromAny(payload)
				}
				alerts = append(alerts, parsed)
			}
		}
		return alerts
	default:
		return nil
	}
}

// buildGitHubVulnerability maps an alert envelope into a normalized vulnerability.
func buildGitHubVulnerability(alert types.AlertEnvelope, source string) (githubVulnerability, bool) {
	repo := strings.TrimSpace(alert.Resource)
	if repo == "" {
		return githubVulnerability{}, false
	}

	payload := mapFromRaw(alert.Payload)
	if payload == nil {
		return githubVulnerability{}, false
	}

	alertType := strings.ToLower(strings.TrimSpace(alert.AlertType))
	if alertType == "" {
		return githubVulnerability{}, false
	}

	externalAlertID := firstNonEmptyString(
		getString(payload, "id"),
		getString(payload, "number"),
	)
	if externalAlertID == "" {
		return githubVulnerability{}, false
	}

	externalID := fmt.Sprintf("%s:%s:%s", alertType, repo, externalAlertID)

	owner := repo
	if idx := strings.Index(repo, "/"); idx > 0 {
		owner = repo[:idx]
	}

	state := strings.ToLower(getString(payload, "state"))
	var openPtr *bool
	if state != "" {
		open := isOpenState(state)
		openPtr = &open
	}

	vuln := githubVulnerability{
		externalID:      externalID,
		externalOwnerID: owner,
		status:          state,
		open:            openPtr,
		category:        alertType,
		externalURI:     getString(payload, "html_url"),
		discoveredAt:    getTime(payload, "created_at"),
		sourceUpdatedAt: getTime(payload, "updated_at"),
		metadata: map[string]any{
			"alert_type": alertType,
			"repository": repo,
			"action":     strings.TrimSpace(alert.Action),
		},
		rawPayload: payload,
	}

	switch alertType {
	case "dependabot":
		vuln.displayName = firstNonEmptyString(
			getString(payload, "dependency", "package", "name"),
			getString(payload, "security_advisory", "summary"),
		)
		vuln.summary = getString(payload, "security_advisory", "summary")
		vuln.description = getString(payload, "security_advisory", "description")
		vuln.severity = firstNonEmptyString(
			getString(payload, "security_vulnerability", "severity"),
			getString(payload, "security_advisory", "severity"),
		)
		vuln.publishedAt = getTime(payload, "security_advisory", "published_at")
		addMetadata(vuln.metadata, payload,
			"dependency",
			"security_advisory",
			"security_vulnerability",
		)
	case "code_scanning":
		vuln.displayName = firstNonEmptyString(
			getString(payload, "rule", "name"),
			getString(payload, "rule", "id"),
		)
		vuln.summary = firstNonEmptyString(
			getString(payload, "rule", "description"),
			vuln.displayName,
		)
		vuln.description = getString(payload, "rule", "description")
		vuln.severity = firstNonEmptyString(
			getString(payload, "rule", "security_severity_level"),
			getString(payload, "rule", "severity"),
		)
		addMetadata(vuln.metadata, payload,
			"rule",
			"tool",
			"most_recent_instance",
		)
	case "secret_scanning":
		vuln.displayName = firstNonEmptyString(
			getString(payload, "secret_type_display_name"),
			getString(payload, "secret_type"),
		)
		vuln.summary = vuln.displayName
		vuln.description = getString(payload, "resolution_comment")
		vuln.sourceUpdatedAt = coalesceTime(getTime(payload, "resolved_at"), vuln.sourceUpdatedAt)
		addMetadata(vuln.metadata, payload,
			"secret_type",
			"secret_type_display_name",
			"resolution",
		)
	default:
		return githubVulnerability{}, false
	}

	if strings.TrimSpace(source) != "" {
		vuln.metadata["source"] = source
	}

	return vuln, true
}

// applyVulnerabilityCreate applies normalized fields to a create builder.
func applyVulnerabilityCreate(create *ent.VulnerabilityCreate, vuln githubVulnerability, provider types.ProviderType) {
	if create == nil {
		return
	}

	create.SetExternalID(vuln.externalID)
	applyVulnerabilityFieldsForCreate(create, vuln, provider)
}

// applyVulnerabilityUpdate applies normalized fields to an update builder.
func applyVulnerabilityUpdate(update *ent.VulnerabilityUpdateOne, vuln githubVulnerability, provider types.ProviderType) {
	if update == nil {
		return
	}

	applyVulnerabilityFieldsForUpdate(update, vuln, provider)
}

// applyVulnerabilityFieldsForCreate sets vulnerability fields on a create builder.
func applyVulnerabilityFieldsForCreate(create *ent.VulnerabilityCreate, vuln githubVulnerability, provider types.ProviderType) {
	if strings.TrimSpace(vuln.externalOwnerID) != "" {
		create.SetExternalOwnerID(vuln.externalOwnerID)
	}
	if strings.TrimSpace(vuln.displayName) != "" {
		create.SetDisplayName(vuln.displayName)
	}
	if strings.TrimSpace(vuln.summary) != "" {
		create.SetSummary(vuln.summary)
	}
	if strings.TrimSpace(vuln.description) != "" {
		create.SetDescription(vuln.description)
	}
	if strings.TrimSpace(vuln.severity) != "" {
		create.SetSeverity(strings.ToLower(vuln.severity))
	}
	if strings.TrimSpace(vuln.status) != "" {
		create.SetStatus(vuln.status)
	}
	if strings.TrimSpace(vuln.category) != "" {
		create.SetCategory(vuln.category)
	}
	if strings.TrimSpace(vuln.externalURI) != "" {
		create.SetExternalURI(vuln.externalURI)
	}
	if provider != types.ProviderUnknown {
		create.SetSource(string(provider))
	}
	if vuln.open != nil {
		create.SetOpen(*vuln.open)
	}

	create.SetNillablePublishedAt(toDateTimePtr(vuln.publishedAt))
	create.SetNillableDiscoveredAt(toDateTimePtr(vuln.discoveredAt))
	create.SetNillableSourceUpdatedAt(toDateTimePtr(vuln.sourceUpdatedAt))

	if len(vuln.metadata) > 0 {
		create.SetMetadata(vuln.metadata)
	}
	if len(vuln.rawPayload) > 0 {
		create.SetRawPayload(vuln.rawPayload)
	}
}

// applyVulnerabilityFieldsForUpdate sets vulnerability fields on an update builder.
func applyVulnerabilityFieldsForUpdate(update *ent.VulnerabilityUpdateOne, vuln githubVulnerability, provider types.ProviderType) {
	if strings.TrimSpace(vuln.externalOwnerID) != "" {
		update.SetExternalOwnerID(vuln.externalOwnerID)
	}
	if strings.TrimSpace(vuln.displayName) != "" {
		update.SetDisplayName(vuln.displayName)
	}
	if strings.TrimSpace(vuln.summary) != "" {
		update.SetSummary(vuln.summary)
	}
	if strings.TrimSpace(vuln.description) != "" {
		update.SetDescription(vuln.description)
	}
	if strings.TrimSpace(vuln.severity) != "" {
		update.SetSeverity(strings.ToLower(vuln.severity))
	}
	if strings.TrimSpace(vuln.status) != "" {
		update.SetStatus(vuln.status)
	}
	if strings.TrimSpace(vuln.category) != "" {
		update.SetCategory(vuln.category)
	}
	if strings.TrimSpace(vuln.externalURI) != "" {
		update.SetExternalURI(vuln.externalURI)
	}
	if provider != types.ProviderUnknown {
		update.SetSource(string(provider))
	}
	if vuln.open != nil {
		update.SetOpen(*vuln.open)
	}

	update.SetNillablePublishedAt(toDateTimePtr(vuln.publishedAt))
	update.SetNillableDiscoveredAt(toDateTimePtr(vuln.discoveredAt))
	update.SetNillableSourceUpdatedAt(toDateTimePtr(vuln.sourceUpdatedAt))

	if len(vuln.metadata) > 0 {
		update.SetMetadata(vuln.metadata)
	}
	if len(vuln.rawPayload) > 0 {
		update.SetRawPayload(vuln.rawPayload)
	}
}

// mapFromRaw decodes a JSON object payload into a map.
func mapFromRaw(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return nil
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil
	}

	return payload
}

// rawMessageFromAny marshals arbitrary values into a JSON raw message.
func rawMessageFromAny(value any) json.RawMessage {
	switch v := value.(type) {
	case json.RawMessage:
		return v
	case []byte:
		return json.RawMessage(v)
	case string:
		return json.RawMessage(v)
	default:
		encoded, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		return json.RawMessage(encoded)
	}
}

// getString extracts a nested string value from a map by path.
func getString(m map[string]any, path ...string) string {
	value, ok := getValue(m, path...)
	if !ok || value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%f", v)
	case int64, int, uint64, uint, float32:
		return strings.TrimSpace(fmt.Sprint(v))
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

// getTime parses a RFC3339 timestamp from a nested map value.
func getTime(m map[string]any, path ...string) *time.Time {
	value := getString(m, path...)
	if value == "" {
		return nil
	}

	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return &parsed
	}

	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return &parsed
	}

	return nil
}

// getValue walks a map using a path of keys and returns the value.
func getValue(m map[string]any, path ...string) (any, bool) {
	if len(path) == 0 {
		return nil, false
	}

	var current any = m
	for _, key := range path {
		next, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		value, ok := next[key]
		if !ok {
			return nil, false
		}
		current = value
	}

	return current, true
}

// toDateTimePtr converts a time pointer to a models.DateTime pointer.
func toDateTimePtr(value *time.Time) *models.DateTime {
	if value == nil {
		return nil
	}
	dt := models.DateTime(*value)
	return &dt
}

// isOpenState reports whether a GitHub alert state represents "open".
func isOpenState(state string) bool {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "open", "active", "new":
		return true
	default:
		return false
	}
}

// firstNonEmptyString returns the first non-empty string from the list.
func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

// coalesceTime returns the first non-nil time pointer.
func coalesceTime(values ...*time.Time) *time.Time {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

// addMetadata copies specified keys from the payload into the destination map.
func addMetadata(dest map[string]any, payload map[string]any, keys ...string) {
	if dest == nil {
		return
	}

	for _, key := range keys {
		if value, ok := payload[key]; ok {
			dest[key] = value
		}
	}
}
