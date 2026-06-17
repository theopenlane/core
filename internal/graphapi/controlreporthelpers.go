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
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/subcontrol"
	"github.com/theopenlane/core/internal/graphapi/model"
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

// hasLinkedPoliciesField reports whether linkedPolicies is requested in the current operation
func hasLinkedPoliciesField(ctx context.Context) bool {
	return controlReportFieldRequested(ctx, fieldLinkedPolicies)
}

// hasSubcontrolsField reports whether subcontrols is requested in the current operation
func hasSubcontrolsField(ctx context.Context) bool {
	return controlReportFieldRequested(ctx, fieldSubcontrols)
}

// hasSubcontrolRelatedControlsField reports whether subcontrols.relatedControls is requested in the current operation
func hasSubcontrolRelatedControlsField(ctx context.Context) bool {
	return controlReportFieldRequested(ctx, fieldSubcontrolRelated)
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

// getSubcontrolRelatedControlInfo fetches MappedControl records for a subcontrol and returns deduplicated related ControlInfo entries
func getSubcontrolRelatedControlInfo(ctx context.Context, sc *model.ControlReport, frameworksInOrg []string) ([]*model.ControlInfo, error) {
	result, err := getMappedControlsBySubcontrolID(ctx, sc.ID)
	if err != nil {
		return nil, err
	}

	return processMappedControlResults(ctx, result, sc.ID, sc.RefCode, sc.ReferenceFramework, frameworksInOrg)
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

	for _, c := range allInOrgControlMappings {
		if isSameControlInfo(selfRefCode, selfFramework, c) {
			continue
		}
		allControls = append(allControls, c)
	}

	allControls = append(allControls, additional...)

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
		q.Select(control.FieldID)
	}).WithSubcontrols(func(q *generated.SubcontrolQuery) {
		q.Select(subcontrol.FieldID)
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
		q.Select(control.FieldID)
	}).WithSubcontrols(func(q *generated.SubcontrolQuery) {
		q.Select(subcontrol.FieldID)
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
		out[i] = &model.ControlReport{
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
	return out
}

// needsRelatedControls returns true when any field that depends on the related-control
// mapping lookup is requested. relatedControls itself, plus evidenceStatus and
// linkedPolicies both expand their ID sets using the related controls list.
func needsRelatedControls(ctx context.Context) bool {
	return controlReportFieldRequested(ctx,
		fieldRelatedControls,
		fieldEvidenceStatus,
		fieldLinkedPolicies,
		fieldSubcontrolRelated,
		fieldSubcontrolEvidenceStatus,
		fieldSubcontrolLinkedPolicies,
	)
}

// enrichControlReports populates computed fields on reports in two passes: related controls first,
// then evidence and policies in batched queries across all entities
func enrichControlReports(ctx context.Context, reports []*model.ControlReport) error {
	var (
		err             error
		frameworksInOrg []string
	)

	needsRelated := needsRelatedControls(ctx)
	wantsEvidence := hasEvidenceField(ctx)
	wantsPolicies := hasLinkedPoliciesField(ctx)
	wantsScRelated := hasSubcontrolRelatedControlsField(ctx)
	wantsScEvidence := hasSubcontrolEvidenceField(ctx)
	wantsScPolicies := hasSubcontrolLinkedPoliciesField(ctx)

	if needsRelated {
		frameworksInOrg, err = getStandardsInOrg(ctx)
		if err != nil {
			return err
		}
	}

	// first pass: populate related controls (per-entity, must come first so IDs are available for pass 2)
	for i, c := range reports {
		if needsRelated {
			reports[i].RelatedControls, err = getRelatedControlInfoFromReport(ctx, c, frameworksInOrg)
			if err != nil {
				return err
			}
		}

		for j, sc := range reports[i].Subcontrols {
			if wantsScRelated {
				reports[i].Subcontrols[j].RelatedControls, err = getSubcontrolRelatedControlInfo(ctx, sc, frameworksInOrg)
				if err != nil {
					return err
				}
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
				if wantsEvidence {
					reports[i].EvidenceStatus = computeEvidenceStatus(evidenceMap, c.ID, c.RelatedControls)
				}
				for j, sc := range reports[i].Subcontrols {
					if wantsScEvidence {
						reports[i].Subcontrols[j].EvidenceStatus = computeEvidenceStatus(evidenceMap, sc.ID, sc.RelatedControls)
					}
				}
			}
		}

		if wantsPolicies || wantsScPolicies {
			policiesMap, err := buildPoliciesMap(ctx, controlIDs, subcontrolIDs)
			if err != nil {
				return err
			}
			for i, c := range reports {
				if wantsPolicies {
					reports[i].LinkedPolicies = computeLinkedPolicies(policiesMap, c.ID, c.RelatedControls)
				}
				for j, sc := range reports[i].Subcontrols {
					if wantsScPolicies {
						reports[i].Subcontrols[j].LinkedPolicies = computeLinkedPolicies(policiesMap, sc.ID, sc.RelatedControls)
					}
				}
			}
		}
	}

	return nil
}

// getRelatedControlInfoFromReport fetches all MappedControl records for a control and returns deduplicated related ControlInfo entries
func getRelatedControlInfoFromReport(ctx context.Context, c *model.ControlReport, frameworksInOrg []string) ([]*model.ControlInfo, error) {
	result, err := getControlMappings(ctx, c.RefCode, c.ReferenceFramework, nil)
	if err != nil {
		return nil, err
	}

	return processMappedControlResults(ctx, result, c.ID, c.RefCode, c.ReferenceFramework, frameworksInOrg)
}

// groupControlReportsByCategory groups a flat ControlReport slice into category buckets sorted by name; controls with no category use an empty string key
func groupControlReportsByCategory(controls []*model.ControlReport) []*model.ControlReportCategory {
	categoryMap := map[string]*model.ControlReportCategory{}

	for _, c := range controls {
		key := ""
		if c.Category != nil {
			key = *c.Category
		}
		if _, ok := categoryMap[key]; !ok {
			categoryMap[key] = &model.ControlReportCategory{
				Category: key,
				Controls: []*model.ControlReport{},
			}
		}
		categoryMap[key].Controls = append(categoryMap[key].Controls, c)
	}

	out := make([]*model.ControlReportCategory, 0, len(categoryMap))
	for _, k := range slices.Sorted(maps.Keys(categoryMap)) {
		cat := categoryMap[k]
		cat.TotalCount = len(cat.Controls)
		out = append(out, cat)
	}

	return out
}
