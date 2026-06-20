package graphapi

import (
	"context"
	"maps"
	"slices"

	"entgo.io/contrib/entgql"
	"github.com/99designs/gqlgen/graphql"
	strcase "github.com/stoewer/go-strcase"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/evidence"
	"github.com/theopenlane/core/internal/ent/generated/internalpolicy"
	"github.com/theopenlane/core/internal/ent/generated/mappedcontrol"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/subcontrol"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/mapx"
	"github.com/theopenlane/iam/auth"
)

// GraphQL field name constants for ControlReport are used in field-presence checks and
// the non-column skip map to avoid scattering identical string literals.
const (
	fieldEdges           = "edges"
	fieldNode            = "node"
	fieldControls        = "controls"
	fieldEvidenceStatus  = "evidenceStatus"
	fieldLinkedPolicies  = "linkedPolicies"
	fieldSubcontrols     = "subcontrols"
	fieldRelatedControls = "relatedControls"
	fieldControlOwner    = "controlOwner"

	fieldSubcontrolRelated        = fieldSubcontrols + "." + fieldRelatedControls
	fieldSubcontrolEvidenceStatus = fieldSubcontrols + "." + fieldEvidenceStatus
	fieldSubcontrolLinkedPolicies = fieldSubcontrols + "." + fieldLinkedPolicies

	// path prefixes that scope field names to their position in each query shape
	pathPrefixPaginated = fieldEdges + "." + fieldNode + "."
	pathPrefixCategory  = fieldControls + "."
)

// nonColumnControlReportFields are ControlReport GraphQL fields that do not correspond to
// selectable ent columns — edges and fields computed after the initial query.
var nonColumnControlReportFields = map[string]struct{}{
	fieldEvidenceStatus:  {},
	fieldLinkedPolicies:  {},
	fieldSubcontrols:     {},
	fieldRelatedControls: {},
	fieldControlOwner:    {},
}

// collectControlReportEntFields walks the GraphQL selection set along path, then collects
// the ent column name for each selectable leaf field, skipping non-column fields.
func collectControlReportEntFields(ctx context.Context, path []string) []string {
	opCtx := graphql.GetOperationContext(ctx)
	selections := graphql.GetFieldContext(ctx).Field.Selections

	for _, segment := range path {
		found := false
		for _, f := range graphql.CollectFields(opCtx, selections, nil) {
			if f.Name == segment {
				selections = f.Selections
				found = true
				break
			}
		}
		if !found {
			return []string{}
		}
	}

	var entFields []string
	for _, f := range graphql.CollectFields(opCtx, selections, nil) {
		if _, skip := nonColumnControlReportFields[f.Name]; skip {
			continue
		}
		entFields = append(entFields, strcase.SnakeCase(f.Name))
	}
	return entFields
}

// getControlFields returns selectable ent column names for the paginated query shape (edges → node path)
func getControlFields(ctx context.Context) []string {
	return collectControlReportEntFields(ctx, []string{fieldEdges, fieldNode})
}

// getControlFieldsForCategory returns selectable ent column names for the by-category query shape (controls path)
func getControlFieldsForCategory(ctx context.Context) []string {
	return collectControlReportEntFields(ctx, []string{fieldControls})
}

// controlReportPathPrefixes are the known path prefixes for ControlReport fields,
// covering both the paginated and by-category query shapes.
var controlReportPathPrefixes = []string{pathPrefixPaginated, pathPrefixCategory}

// controlReportFieldRequested returns true if any of the given field paths (relative to a
// ControlReport node) are requested in the current operation, under any known path prefix.
func controlReportFieldRequested(ctx context.Context, fields ...string) bool {
	paths := make([]string, 0, len(fields)*len(controlReportPathPrefixes))
	for _, prefix := range controlReportPathPrefixes {
		for _, field := range fields {
			paths = append(paths, prefix+field)
		}
	}
	return graphql.AnyFieldRequested(ctx, paths...)
}

// hasEvidenceField reports whether evidenceStatus is requested in the current operation
func hasEvidenceField(ctx context.Context) bool {
	return controlReportFieldRequested(ctx, fieldEvidenceStatus)
}

// hasControlOwnerField reports whether controlOwner is requested in the current operation
func hasControlOwnerField(ctx context.Context) bool {
	return controlReportFieldRequested(ctx, fieldControlOwner)
}

// hasLinkedPoliciesField reports whether linkedPolicies is requested in the current operation
func hasLinkedPoliciesField(ctx context.Context) bool {
	return controlReportFieldRequested(ctx, fieldLinkedPolicies)
}

// hasSubcontrolsField reports whether subcontrols is requested in the current operation
func hasSubcontrolsField(ctx context.Context) bool {
	return controlReportFieldRequested(ctx, fieldSubcontrols)
}

// hasSubcontrolEvidenceField reports whether subcontrols.evidenceStatus is requested in the current operation
func hasSubcontrolEvidenceField(ctx context.Context) bool {
	return controlReportFieldRequested(ctx, fieldSubcontrolEvidenceStatus)
}

// hasSubcontrolLinkedPoliciesField reports whether subcontrols.linkedPolicies is requested in the current operation
func hasSubcontrolLinkedPoliciesField(ctx context.Context) bool {
	return controlReportFieldRequested(ctx, fieldSubcontrolLinkedPolicies)
}

// hasAnySubcontrolAdditionalField reports whether any enrichment field under subcontrols is requested
func hasAnySubcontrolAdditionalField(ctx context.Context) bool {
	return controlReportFieldRequested(ctx, fieldSubcontrolRelated, fieldSubcontrolEvidenceStatus, fieldSubcontrolLinkedPolicies)
}

// classifyMappedControl routes a single ControlInfo into the system-owned or org-owned map,
// skipping the control that owns the mapping (selfID) and controls from frameworks not in the org.
func classifyMappedControl(info *model.ControlInfo, systemOwned bool, selfID string, frameworksInOrg []string, systemMap, orgMap map[string]*model.ControlInfo) {
	if info.ID == selfID || !shouldCheckForControl(info, frameworksInOrg) {
		return
	}

	key := generateMapControlKey(info.RefCode, info.ReferenceFramework)
	if systemOwned {
		systemMap[key] = info
	} else {
		orgMap[key] = info
	}
}

// processMappedControlResults converts a slice of MappedControl records into ControlInfo entries,
// excluding the entry identified by selfID and deduplicating via refCode+framework key
func processMappedControlResults(ctx context.Context, result []*generated.MappedControl, selfID string, selfRefCode string, selfFramework *string, frameworksInOrg []string) ([]*model.ControlInfo, error) {
	allControls := []*model.ControlInfo{}

	if len(result) == 0 {
		return allControls, nil
	}

	systemOwnedMappedControls := map[string]*model.ControlInfo{}
	allInOrgControlMappings := map[string]*model.ControlInfo{}

	for _, r := range result {
		for _, c := range r.Edges.FromControls {
			classifyMappedControl(controlEdgeToControlInfo(c), c.SystemOwned, selfID, frameworksInOrg, systemOwnedMappedControls, allInOrgControlMappings)
		}
		for _, c := range r.Edges.ToControls {
			classifyMappedControl(controlEdgeToControlInfo(c), c.SystemOwned, selfID, frameworksInOrg, systemOwnedMappedControls, allInOrgControlMappings)
		}
		for _, c := range r.Edges.FromSubcontrols {
			classifyMappedControl(subcontrolEdgeToControlInfo(c), c.SystemOwned, selfID, frameworksInOrg, systemOwnedMappedControls, allInOrgControlMappings)
		}
		for _, c := range r.Edges.ToSubcontrols {
			classifyMappedControl(subcontrolEdgeToControlInfo(c), c.SystemOwned, selfID, frameworksInOrg, systemOwnedMappedControls, allInOrgControlMappings)
		}
	}

	additional := getOrgMappedControlsInfo(ctx, systemOwnedMappedControls, selfRefCode, selfFramework)

	seenIDs := map[string]struct{}{}

	for _, c := range allInOrgControlMappings {
		if isSameControlInfo(selfRefCode, selfFramework, c) {
			continue
		}

		seenIDs[c.ID] = struct{}{}
		allControls = append(allControls, c)
	}

	for _, c := range additional {
		if _, ok := seenIDs[c.ID]; !ok {
			allControls = append(allControls, c)
		}
	}

	return allControls, nil
}

// collectAllEntityIDs returns deduplicated control and subcontrol IDs referenced by
// the reports, including IDs drawn from each report's RelatedControls.
func collectAllEntityIDs(reports []*model.ControlReport) (controlIDs, subcontrolIDs []string) {
	seenC := map[string]struct{}{}
	seenSC := map[string]struct{}{}

	addControl := func(id string) {
		if _, ok := seenC[id]; !ok {
			seenC[id] = struct{}{}
			controlIDs = append(controlIDs, id)
		}
	}
	addSubcontrol := func(id string) {
		if _, ok := seenSC[id]; !ok {
			seenSC[id] = struct{}{}
			subcontrolIDs = append(subcontrolIDs, id)
		}
	}

	for _, r := range reports {
		addControl(r.ID)
		for _, rc := range r.RelatedControls {
			if rc.IsSubcontrol {
				addSubcontrol(rc.ID)
			} else {
				addControl(rc.ID)
			}
		}
		for _, sc := range r.Subcontrols {
			addSubcontrol(sc.ID)
			for _, rc := range sc.RelatedControls {
				if rc.IsSubcontrol {
					addSubcontrol(rc.ID)
				} else {
					addControl(rc.ID)
				}
			}
		}
	}
	return
}

// buildEvidenceMap fetches all evidence linked to the given entity IDs in one query
// and returns a map of entity ID to the evidence records linked to that entity.
func buildEvidenceMap(ctx context.Context, controlIDs, subcontrolIDs []string) (map[string][]*generated.Evidence, error) {
	if len(controlIDs) == 0 && len(subcontrolIDs) == 0 {
		return map[string][]*generated.Evidence{}, nil
	}

	orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// crate evidence edges for controls and subcontrols
	edgePredicates := make([]predicate.Evidence, 0, 2) //nolint:mnd
	if len(controlIDs) > 0 {
		edgePredicates = append(edgePredicates, evidence.HasControlsWith(control.IDIn(controlIDs...)))
	}
	if len(subcontrolIDs) > 0 {
		edgePredicates = append(edgePredicates, evidence.HasSubcontrolsWith(subcontrol.IDIn(subcontrolIDs...)))
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	evidenceList, err := withTransactionalMutation(ctx).Evidence.Query().Where(
		evidence.OwnerIDIn(orgIDs...),
		evidence.Or(edgePredicates...),
	).WithControls(func(q *generated.ControlQuery) {
		q.Where(control.IDIn(controlIDs...)).Select(control.FieldID)
	}).WithSubcontrols(func(q *generated.SubcontrolQuery) {
		q.Where(subcontrol.IDIn(subcontrolIDs...)).Select(subcontrol.FieldID)
	}).All(allowCtx)
	if err != nil {
		return nil, err
	}

	m := map[string][]*generated.Evidence{}
	for _, e := range evidenceList {
		for _, c := range e.Edges.Controls {
			m[c.ID] = append(m[c.ID], e)
		}
		for _, sc := range e.Edges.Subcontrols {
			m[sc.ID] = append(m[sc.ID], e)
		}
	}
	return m, nil
}

// buildPoliciesMap fetches all internal policies linked to the given entity IDs in one query
// and returns a map of entity ID to the policy records linked to that entity.
func buildPoliciesMap(ctx context.Context, controlIDs, subcontrolIDs []string) (map[string][]*generated.InternalPolicy, error) {
	if len(controlIDs) == 0 && len(subcontrolIDs) == 0 {
		return map[string][]*generated.InternalPolicy{}, nil
	}

	orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// create edges for controls and subcontrols
	edgePredicates := make([]predicate.InternalPolicy, 0, 2) //nolint:mnd
	if len(controlIDs) > 0 {
		edgePredicates = append(edgePredicates, internalpolicy.HasControlsWith(control.IDIn(controlIDs...)))
	}
	if len(subcontrolIDs) > 0 {
		edgePredicates = append(edgePredicates, internalpolicy.HasSubcontrolsWith(subcontrol.IDIn(subcontrolIDs...)))
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	policies, err := withTransactionalMutation(ctx).InternalPolicy.Query().Where(
		internalpolicy.OwnerIDIn(orgIDs...),
		internalpolicy.Or(edgePredicates...),
	).WithControls(func(q *generated.ControlQuery) {
		q.Where(control.IDIn(controlIDs...)).Select(control.FieldID)
	}).WithSubcontrols(func(q *generated.SubcontrolQuery) {
		q.Where(subcontrol.IDIn(subcontrolIDs...)).Select(subcontrol.FieldID)
	}).Select("id", "name", "status").All(allowCtx)
	if err != nil {
		return nil, err
	}

	m := map[string][]*generated.InternalPolicy{}
	for _, p := range policies {
		for _, c := range p.Edges.Controls {
			m[c.ID] = append(m[c.ID], p)
		}
		for _, sc := range p.Edges.Subcontrols {
			m[sc.ID] = append(m[sc.ID], p)
		}
	}
	return m, nil
}

// computeEvidenceStatus aggregates evidence for a single entity from the pre-fetched map,
// including evidence from related controls, deduplicating by evidence ID.
func computeEvidenceStatus(evidenceMap map[string][]*generated.Evidence, id string, relatedControls []*model.ControlInfo) *model.ControlEvidence {
	seen := map[string]struct{}{}
	var all []*generated.Evidence

	for _, e := range evidenceMap[id] {
		if _, ok := seen[e.ID]; !ok {
			seen[e.ID] = struct{}{}
			all = append(all, e)
		}
	}
	for _, rc := range relatedControls {
		for _, e := range evidenceMap[rc.ID] {
			if _, ok := seen[e.ID]; !ok {
				seen[e.ID] = struct{}{}
				all = append(all, e)
			}
		}
	}

	approvedCount := 0
	statusCounts := map[enums.EvidenceStatus]int{}
	for _, e := range all {
		statusCounts[e.Status]++
		if e.Status == enums.EvidenceStatusAuditorApproved {
			approvedCount++
		}
	}

	countByStatus := make([]*model.EvidenceCountByStatus, 0, len(statusCounts))
	for status, count := range statusCounts {
		countByStatus = append(countByStatus, &model.EvidenceCountByStatus{
			Status:     status,
			TotalCount: count,
		})
	}
	slices.SortFunc(countByStatus, func(a, b *model.EvidenceCountByStatus) int {
		return evidenceStatusRank(a.Status) - evidenceStatusRank(b.Status)
	})

	return &model.ControlEvidence{
		TotalCount:    len(all),
		WorstStatus:   worstEvidenceStatus(all),
		ApprovedCount: approvedCount,
		CountByStatus: countByStatus,
	}
}

// computeLinkedPolicies aggregates policies for a single entity from the pre-fetched map,
// including policies from related controls, deduplicating by policy ID.
func computeLinkedPolicies(policiesMap map[string][]*generated.InternalPolicy, id string, relatedControls []*model.ControlInfo) *model.ControlPolicies {
	seen := map[string]struct{}{}
	var summaries []*model.PolicySummary

	collect := func(p *generated.InternalPolicy) {
		if _, ok := seen[p.ID]; !ok {
			seen[p.ID] = struct{}{}
			summaries = append(summaries, &model.PolicySummary{
				ID:     p.ID,
				Name:   p.Name,
				Status: p.Status,
			})
		}
	}

	for _, p := range policiesMap[id] {
		collect(p)
	}

	for _, rc := range relatedControls {
		for _, p := range policiesMap[rc.ID] {
			collect(p)
		}
	}

	return &model.ControlPolicies{
		TotalCount:       len(summaries),
		InternalPolicies: summaries,
	}
}

// controlToReport maps a Control ent to a ControlReport model with empty enrichment fields populated later
func controlToReport(c *generated.Control) *model.ControlReport {
	return &model.ControlReport{
		ID:                 c.ID,
		RefCode:            c.RefCode,
		Description:        &c.Description,
		Title:              &c.Title,
		Status:             &c.Status,
		ControlOwner:       c.Edges.ControlOwner,
		ReferenceFramework: c.ReferenceFramework,
		Category:           &c.Category,
		Subcategory:        &c.Subcategory,
		RelatedControls:    []*model.ControlInfo{},
		EvidenceStatus:     &model.ControlEvidence{},
		LinkedPolicies:     &model.ControlPolicies{},
		Subcontrols:        convertSubcontrolToControlReportEdge(c.Edges.Subcontrols),
	}
}

// subcontrolToReport maps a Subcontrol ent to a ControlReport model with empty enrichment fields populated later
func subcontrolToReport(c *generated.Subcontrol) *model.ControlReport {
	return &model.ControlReport{
		ID:                 c.ID,
		RefCode:            c.RefCode,
		Description:        &c.Description,
		Title:              &c.Title,
		Status:             &c.Status,
		ControlOwner:       c.Edges.ControlOwner,
		ReferenceFramework: c.ReferenceFramework,
		Category:           &c.Category,
		Subcategory:        &c.Subcategory,
		RelatedControls:    []*model.ControlInfo{},
		EvidenceStatus:     &model.ControlEvidence{},
		LinkedPolicies:     &model.ControlPolicies{},
	}
}

// convertControlToControlReportEdge wraps a ControlConnection into a ControlReportConnection, preserving page info and total count
func convertControlToControlReportEdge(controls *generated.ControlConnection) *model.ControlReportConnection {
	edges := make([]*model.ControlReportEdge, len(controls.Edges))
	for i, c := range controls.Edges {
		edges[i] = &model.ControlReportEdge{Node: controlToReport(c.Node)}
	}
	return &model.ControlReportConnection{
		Edges:      edges,
		PageInfo:   &controls.PageInfo,
		TotalCount: controls.TotalCount,
	}
}

// convertSubcontrolToControlReportEdge converts a slice of Subcontrol records to ControlReport models
func convertSubcontrolToControlReportEdge(controls []*generated.Subcontrol) []*model.ControlReport {
	edges := make([]*model.ControlReport, len(controls))
	for i, c := range controls {
		edges[i] = subcontrolToReport(c)
	}
	return edges
}

// controlEdgeToControlInfo maps a Control record to ControlInfo with IsSubcontrol set to false
func controlEdgeToControlInfo(c *generated.Control) *model.ControlInfo {
	return &model.ControlInfo{
		ID:                 c.ID,
		RefCode:            c.RefCode,
		Description:        &c.Description,
		Title:              &c.Title,
		Status:             &c.Status,
		ControlOwner:       c.Edges.ControlOwner,
		ReferenceFramework: c.ReferenceFramework,
		IsSubcontrol:       false,
	}
}

// subcontrolEdgeToControlInfo maps a Subcontrol record to ControlInfo with IsSubcontrol set to true
func subcontrolEdgeToControlInfo(c *generated.Subcontrol) *model.ControlInfo {
	return &model.ControlInfo{
		ID:                 c.ID,
		RefCode:            c.RefCode,
		Description:        &c.Description,
		Title:              &c.Title,
		Status:             &c.Status,
		ControlOwner:       c.Edges.ControlOwner,
		ReferenceFramework: c.ReferenceFramework,
		IsSubcontrol:       true,
	}
}

// convertReportOrderToControlOrderBy translates GraphQL ControlReportOrder inputs to ent ControlOrder; defaults to created_at DESC when orderBy is nil
func convertReportOrderToControlOrderBy(orderBy []*model.ControlReportOrder) []*generated.ControlOrder {
	if orderBy == nil {
		return []*generated.ControlOrder{
			{
				Field:     generated.ControlOrderFieldCreatedAt,
				Direction: entgql.OrderDirectionDesc,
			},
		}
	}

	orderByOut := make([]*generated.ControlOrder, 0, len(orderBy))

	for _, ob := range orderBy {
		var field *generated.ControlOrderField
		switch ob.Field {
		case model.ControlReportOrderFieldCreatedAt:
			field = generated.ControlOrderFieldCreatedAt
		case model.ControlReportOrderFieldUpdatedAt:
			field = generated.ControlOrderFieldUpdatedAt
		case model.ControlReportOrderFieldRefCode:
			field = generated.ControlOrderFieldRefCode
		case model.ControlReportOrderFieldTitle:
			field = generated.ControlOrderFieldTitle
		case model.ControlReportOrderFieldReferenceFramework:
			field = generated.ControlOrderFieldReferenceFramework
		}

		if field != nil {
			orderByOut = append(orderByOut, &generated.ControlOrder{Field: field, Direction: ob.Direction})
		}
	}

	return orderByOut
}

// evidenceStatusSeverity orders statuses from worst (index 0) to best (last index)
var evidenceStatusSeverity = []enums.EvidenceStatus{
	enums.EvidenceStatusRejected,
	enums.EvidenceStatusMissingArtifact,
	enums.EvidenceStatusNeedsRenewal,
	enums.EvidenceStatusRequested,
	enums.EvidenceStatusDraft,
	enums.EvidenceStatusSubmitted,
	enums.EvidenceStatusInReview,
	enums.EvidenceStatusReadyForAuditor,
	enums.EvidenceStatusAuditorApproved,
}

// evidenceStatusRank returns the severity index of s; unknown statuses rank beyond all known values
func evidenceStatusRank(s enums.EvidenceStatus) int {
	for i, v := range evidenceStatusSeverity {
		if v == s {
			return i
		}
	}

	return len(evidenceStatusSeverity)
}

// worstEvidenceStatus returns the most severe status from the provided list.
// Returns nil if the list is empty.
func worstEvidenceStatus(evidences []*generated.Evidence) *enums.EvidenceStatus {
	if len(evidences) == 0 {
		return nil
	}

	worst := evidences[0].Status
	for _, e := range evidences[1:] {
		if evidenceStatusRank(e.Status) < evidenceStatusRank(worst) {
			worst = e.Status
		}
	}

	return &worst
}

// shouldCheckForControl returns true if there is no reference framework (custom control) or if the organization has controls for the specific framework
func shouldCheckForControl(c *model.ControlInfo, frameworksInOrg []string) bool {
	if c.ReferenceFramework == nil {
		return true
	}

	if *c.ReferenceFramework == "" {
		return true
	}

	return slices.Contains(frameworksInOrg, *c.ReferenceFramework)
}

// convertControlListToControlReports converts a flat Control slice to ControlReport models with empty enrichment fields
func convertControlListToControlReports(controls []*generated.Control) []*model.ControlReport {
	out := make([]*model.ControlReport, len(controls))
	for i, c := range controls {
		out[i] = controlToReport(c)
	}
	return out
}

// buildControlPredicates builds a deduplicated slice of control predicates for the given
// reports, scoped to system-owned or org-owned controls
func buildControlPredicates(controls []*model.ControlReport, orgIDs []string) []predicate.Control {
	seen := map[string]struct{}{}
	predicates := make([]predicate.Control, 0, len(controls))

	for _, c := range controls {
		key := generateMapControlKey(c.RefCode, c.ReferenceFramework)
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}

		p := []predicate.Control{
			control.RefCode(c.RefCode),
			control.Or(control.SystemOwned(true), control.OwnerIDIn(orgIDs...)),
		}

		if c.ReferenceFramework == nil {
			p = append(p, control.ReferenceFrameworkIsNil())
		} else {
			p = append(p, control.ReferenceFramework(*c.ReferenceFramework))
		}

		predicates = append(predicates, control.And(p...))
	}

	return predicates
}

// mcPart holds a single participant (control or subcontrol) collected from a MappedControl edge
type mcPart struct {
	key  string
	info *model.ControlInfo
}

// collectControlPart converts a Control edge into an mapped control part, records system-owned controls in sysControls
// it returns false if there are no results returned
func collectControlPart(c *generated.Control, frameworksInOrg []string, sysControls map[string]*model.ControlInfo) (mcPart, bool) {
	info := controlEdgeToControlInfo(c)
	if !shouldCheckForControl(info, frameworksInOrg) {
		return mcPart{}, false
	}

	key := generateMapControlKey(c.RefCode, c.ReferenceFramework)
	if c.SystemOwned {
		sysControls[key] = info
	}

	return mcPart{key, info}, true
}

// collectSubcontrolPart converts a Subcontrol edge into an mcPart, records system-owned controls
// in sysControls and returns the result. It will return false if the subcontrol's framework
// is not present in the org
func collectSubcontrolPart(sc *generated.Subcontrol, frameworksInOrg []string, sysControls map[string]*model.ControlInfo) (mcPart, bool) {
	info := subcontrolEdgeToControlInfo(sc)
	if !shouldCheckForControl(info, frameworksInOrg) {
		return mcPart{}, false
	}

	key := generateMapControlKey(sc.RefCode, sc.ReferenceFramework)
	if sc.SystemOwned {
		sysControls[key] = info
	}

	return mcPart{key, info}, true
}

// indexRelated adds all parts whose key differs from selfKey into raw[outerKey],
// deduplicating by refCode::framework via the inner map
func indexRelated(raw map[string]map[string]*model.ControlInfo, outerKey, selfKey string, parts []mcPart) {
	if raw[outerKey] == nil {
		raw[outerKey] = map[string]*model.ControlInfo{}
	}

	for _, p := range parts {
		if p.key != selfKey {
			raw[outerKey][p.key] = p.info
		}
	}
}

// indexEntry holds a single (outerKey, selfKey) pair produced by an outerKeys function
// outerKey is the map key to index under; selfKey is excluded from that key's related list
type indexEntry struct {
	outerKey string
	selfKey  string
}

// controlIndexKeys returns indexEntry values for control (non-subcontrol) parts,
// keyed and self-excluded by refCode::framework
func controlIndexKeys(parts []mcPart) []indexEntry {
	var entries []indexEntry

	for _, p := range parts {
		if !p.info.IsSubcontrol {
			entries = append(entries, indexEntry{p.key, p.key})
		}
	}

	return entries
}

// subcontrolIndexKeys returns indexEntry values for subcontrol parts,
// keyed by subcontrol ID with self-exclusion by refCode::framework
func subcontrolIndexKeys(parts []mcPart) []indexEntry {
	var entries []indexEntry
	for _, p := range parts {
		if p.info.IsSubcontrol {
			entries = append(entries, indexEntry{p.info.ID, p.key})
		}
	}
	return entries
}

// allIndexKeys returns indexEntry values for all parts — controls keyed by refCode::framework,
// subcontrols keyed by ID. Used when a single query serves both lookup types.
func allIndexKeys(parts []mcPart) []indexEntry {
	return append(controlIndexKeys(parts), subcontrolIndexKeys(parts)...)
}

// resolveRawRelated converts a raw related map (whose values may contain system-owned
// ControlInfos) to a final map where system-owned entries are replaced by their
// org-owned counterparts from orgLookup. sysControls identifies which keys are system-owned.
func resolveRawRelated(raw map[string]map[string]*model.ControlInfo, sysControls, orgLookup map[string]*model.ControlInfo) map[string][]*model.ControlInfo {
	m := make(map[string][]*model.ControlInfo, len(raw))

	for outerKey, inner := range raw {
		slice := make([]*model.ControlInfo, 0, len(inner))

		for refKey, info := range inner {
			if _, isSys := sysControls[refKey]; isSys {
				if orgInfo, ok := orgLookup[refKey]; ok {
					slice = append(slice, orgInfo)
				}
			} else {
				slice = append(slice, info)
			}
		}

		if len(slice) > 0 {
			m[outerKey] = slice
		}
	}

	return m
}

// buildRelatedFromMappedControls processes a pre-fetched MappedControl slice into a resolved related-control map
func buildRelatedFromMappedControls(ctx context.Context, mcs []*generated.MappedControl, frameworksInOrg []string, outerKeys func([]mcPart) []indexEntry) (map[string][]*model.ControlInfo, error) {
	sysControls := map[string]*model.ControlInfo{}
	raw := map[string]map[string]*model.ControlInfo{}

	for _, mc := range mcs {
		var parts []mcPart

		for _, c := range mc.Edges.FromControls {
			if p, ok := collectControlPart(c, frameworksInOrg, sysControls); ok {
				parts = append(parts, p)
			}
		}

		for _, c := range mc.Edges.ToControls {
			if p, ok := collectControlPart(c, frameworksInOrg, sysControls); ok {
				parts = append(parts, p)
			}
		}

		for _, sc := range mc.Edges.FromSubcontrols {
			if p, ok := collectSubcontrolPart(sc, frameworksInOrg, sysControls); ok {
				parts = append(parts, p)
			}
		}

		for _, sc := range mc.Edges.ToSubcontrols {
			if p, ok := collectSubcontrolPart(sc, frameworksInOrg, sysControls); ok {
				parts = append(parts, p)
			}
		}

		for _, e := range outerKeys(parts) {
			indexRelated(raw, e.outerKey, e.selfKey, parts)
		}
	}

	orgLookup, err := buildOrgControlLookupMap(ctx, sysControls)
	if err != nil {
		return nil, err
	}

	return resolveRawRelated(raw, sysControls, orgLookup), nil
}

// buildMappingsMap fetches MappedControl records in a single query for the given controls and/or
// subcontrol IDs and returns a unified map. Controls are keyed by refCode::framework; subcontrols
// by ID. Pass nil controls or empty scIDs to skip that half of the query.
func buildMappingsMap(ctx context.Context, controls []*model.ControlReport, scIDs []string, frameworksInOrg []string) (map[string][]*model.ControlInfo, error) {
	orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var where []predicate.MappedControl

	if len(controls) > 0 {
		ctrlPreds := buildControlPredicates(controls, orgIDs)
		where = append(where,
			mappedcontrol.HasFromControlsWith(control.Or(ctrlPreds...)),
			mappedcontrol.HasToControlsWith(control.Or(ctrlPreds...)),
		)
	}

	if len(scIDs) > 0 {
		ownership := subcontrol.Or(subcontrol.SystemOwned(true), subcontrol.OwnerIDIn(orgIDs...))
		where = append(where,
			mappedcontrol.HasFromSubcontrolsWith(subcontrol.IDIn(scIDs...), ownership),
			mappedcontrol.HasToSubcontrolsWith(subcontrol.IDIn(scIDs...), ownership),
		)
	}

	if len(where) == 0 {
		return map[string][]*model.ControlInfo{}, nil
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	mcs, err := withTransactionalMutation(ctx).MappedControl.Query().Where(
		mappedcontrol.Or(where...),
	).WithFromControls().WithToControls().WithFromSubcontrols().WithToSubcontrols().All(allowCtx)
	if err != nil {
		return nil, err
	}

	if len(mcs) == 0 {
		return map[string][]*model.ControlInfo{}, nil
	}

	return buildRelatedFromMappedControls(ctx, mcs, frameworksInOrg, allIndexKeys)
}

// buildOrgControlLookupMap collects all system-owned controls referenced across the
// pre-fetched MappedControl maps and resolves them to their org-owned counterparts in
// a single query, returning a map keyed by refCode::framework
func buildOrgControlLookupMap(ctx context.Context, sysControls map[string]*model.ControlInfo) (map[string]*model.ControlInfo, error) {
	if len(sysControls) == 0 {
		return map[string]*model.ControlInfo{}, nil
	}

	orgInfos, ok := findOrganizationControlInfoForMappings(ctx, sysControls)
	if !ok {
		return map[string]*model.ControlInfo{}, nil
	}

	lookup := make(map[string]*model.ControlInfo, len(orgInfos))

	for _, info := range orgInfos {
		lookup[generateMapControlKey(info.RefCode, info.ReferenceFramework)] = info
	}

	return lookup, nil
}

// needsControlMappings returns true when any control-level field requires the mapped
// control lookup
func needsControlMappings(ctx context.Context) bool {
	return controlReportFieldRequested(ctx,
		fieldRelatedControls,
		fieldEvidenceStatus,
		fieldLinkedPolicies,
	)
}

// needsSubcontrolMappings returns true when any subcontrol-level field requires the
// mapped control lookup for subcontrols.
func needsSubcontrolMappings(ctx context.Context) bool {
	return controlReportFieldRequested(ctx,
		fieldSubcontrolRelated,
		fieldSubcontrolEvidenceStatus,
		fieldSubcontrolLinkedPolicies,
	)
}

// enrichControlReports populates computed fields on reports in two passes: related controls first,
// then evidence and policies in batched queries across all entities.
// The first pass fetches all mapped controls in two queries (one for controls, one for
// subcontrols) and resolves system→org control pairs in one more query
func enrichControlReports(ctx context.Context, reports []*model.ControlReport) error {
	var (
		err             error
		frameworksInOrg []string
	)

	needsCtrlMappings := needsControlMappings(ctx)
	needsScMappings := needsSubcontrolMappings(ctx)
	wantsEvidence := hasEvidenceField(ctx)
	wantsPolicies := hasLinkedPoliciesField(ctx)
	wantsScEvidence := hasSubcontrolEvidenceField(ctx)
	wantsScPolicies := hasSubcontrolLinkedPoliciesField(ctx)

	if needsCtrlMappings || needsScMappings {
		frameworksInOrg, err = getStandardsInOrg(ctx)
		if err != nil {
			return err
		}
	}

	var mappingsMap map[string][]*model.ControlInfo

	if needsCtrlMappings || needsScMappings {
		var scIDs []string
		if needsScMappings {
			for _, r := range reports {
				for _, sc := range r.Subcontrols {
					scIDs = append(scIDs, sc.ID)
				}
			}
		}

		var ctrlReports []*model.ControlReport
		if needsCtrlMappings {
			ctrlReports = reports
		}

		mappingsMap, err = buildMappingsMap(ctx, ctrlReports, scIDs, frameworksInOrg)
		if err != nil {
			return err
		}
	}

	// first pass: populate related controls from pre-fetched map
	for i, c := range reports {
		key := generateMapControlKey(c.RefCode, c.ReferenceFramework)
		if rc := mappingsMap[key]; rc != nil {
			reports[i].RelatedControls = rc
		}

		for j, sc := range reports[i].Subcontrols {
			if rc := mappingsMap[sc.ID]; rc != nil {
				reports[i].Subcontrols[j].RelatedControls = rc
			}
		}
	}

	// second pass: batch evidence and policies across all entities in two queries total
	if wantsEvidence || wantsScEvidence || wantsPolicies || wantsScPolicies {
		controlIDs, subcontrolIDs := collectAllEntityIDs(reports)

		if wantsEvidence || wantsScEvidence {
			evidenceMap, err := buildEvidenceMap(ctx, controlIDs, subcontrolIDs)
			if err != nil {
				return err
			}

			for i, c := range reports {
				reports[i].EvidenceStatus = computeEvidenceStatus(evidenceMap, c.ID, c.RelatedControls)

				for j, sc := range reports[i].Subcontrols {
					reports[i].Subcontrols[j].EvidenceStatus = computeEvidenceStatus(evidenceMap, sc.ID, sc.RelatedControls)
				}
			}
		}

		if wantsPolicies || wantsScPolicies {
			policiesMap, err := buildPoliciesMap(ctx, controlIDs, subcontrolIDs)
			if err != nil {
				return err
			}

			for i, c := range reports {
				reports[i].LinkedPolicies = computeLinkedPolicies(policiesMap, c.ID, c.RelatedControls)

				for j, sc := range reports[i].Subcontrols {
					reports[i].Subcontrols[j].LinkedPolicies = computeLinkedPolicies(policiesMap, sc.ID, sc.RelatedControls)
				}
			}
		}
	}

	return nil
}

// groupControlReportsByCategory groups a flat ControlReport slice into category buckets sorted by name; controls with no category use an empty string key
func groupControlReportsByCategory(controls []*model.ControlReport) []*model.ControlReportCategory {
	categoryMap := map[string]*model.ControlReportCategory{}

	for _, c := range controls {
		key := ""
		if c.Category != nil {
			key = *c.Category
		}

		entry := mapx.GetOrInit(categoryMap, key, func() *model.ControlReportCategory {
			return &model.ControlReportCategory{Category: key}
		})

		entry.Controls = append(entry.Controls, c)
	}

	out := make([]*model.ControlReportCategory, 0, len(categoryMap))

	for _, k := range slices.Sorted(maps.Keys(categoryMap)) {
		cat := categoryMap[k]
		cat.TotalCount = len(cat.Controls)
		out = append(out, cat)
	}

	return out
}
