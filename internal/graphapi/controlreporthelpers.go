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
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
)

// nonColumnControlReportFields are ControlReport GraphQL fields that do not correspond to
// selectable ent columns — edges and fields computed after the initial query.
var nonColumnControlReportFields = map[string]struct{}{
	"evidenceStatus":  {},
	"linkedPolicies":  {},
	"subcontrols":     {},
	"relatedControls": {},
	"controlOwner":    {},
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

func getControlFields(ctx context.Context) []string {
	return collectControlReportEntFields(ctx, []string{"edges", "node"})
}

func getControlFieldsForCategory(ctx context.Context) []string {
	return collectControlReportEntFields(ctx, []string{"controls"})
}

// controlReportPathPrefixes are the known path prefixes for ControlReport fields,
// covering both the paginated and by-category query shapes.
var controlReportPathPrefixes = []string{"edges.node.", "controls."}

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

func hasEvidenceField(ctx context.Context) bool {
	return controlReportFieldRequested(ctx, "evidenceStatus")
}

func hasLinkedPoliciesField(ctx context.Context) bool {
	return controlReportFieldRequested(ctx, "linkedPolicies")
}

func hasSubcontrolsField(ctx context.Context) bool {
	return controlReportFieldRequested(ctx, "subcontrols")
}

func hasSubcontrolRelatedControlsField(ctx context.Context) bool {
	return controlReportFieldRequested(ctx, "subcontrols.relatedControls")
}

func hasSubcontrolEvidenceField(ctx context.Context) bool {
	return controlReportFieldRequested(ctx, "subcontrols.evidenceStatus")
}

func hasSubcontrolLinkedPoliciesField(ctx context.Context) bool {
	return controlReportFieldRequested(ctx, "subcontrols.linkedPolicies")
}

func hasAnySubcontrolAdditionalField(ctx context.Context) bool {
	return controlReportFieldRequested(ctx, "subcontrols.relatedControls", "subcontrols.evidenceStatus", "subcontrols.linkedPolicies")
}

func getSubcontrolRelatedControlInfo(ctx context.Context, sc *model.ControlReport, frameworksInOrg []string) ([]*model.ControlInfo, error) {
	result, err := getMappedControlsBySubcontrolID(ctx, sc.ID)
	if err != nil {
		return nil, err
	}

	return processMappedControlResults(ctx, result, sc.ID, sc.RefCode, sc.ReferenceFramework, frameworksInOrg)
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
			if c.ID == selfID {
				continue
			}
			controlInfo := controlEdgeToControlInfo(c)
			if !shouldCheckForControl(controlInfo, frameworksInOrg) {
				continue
			}
			key := generateMapControlKey(c.RefCode, c.ReferenceFramework)
			logx.FromContext(ctx).Warn().Str("key", key).Msg("adding control found in from controls")
			if c.SystemOwned {
				systemOwnedMappedControls[key] = controlInfo
			} else {
				allInOrgControlMappings[key] = controlInfo
			}
		}

		for _, c := range r.Edges.ToControls {
			if c.ID == selfID {
				continue
			}
			controlInfo := controlEdgeToControlInfo(c)
			if !shouldCheckForControl(controlInfo, frameworksInOrg) {
				continue
			}
			key := generateMapControlKey(c.RefCode, c.ReferenceFramework)
			logx.FromContext(ctx).Warn().Str("key", key).Msg("adding control found in to controls")
			if c.SystemOwned {
				systemOwnedMappedControls[key] = controlInfo
			} else {
				allInOrgControlMappings[key] = controlInfo
			}
		}

		for _, c := range r.Edges.FromSubcontrols {
			if c.ID == selfID {
				continue
			}
			controlInfo := subcontrolEdgeToControlInfo(c)
			if !shouldCheckForControl(controlInfo, frameworksInOrg) {
				continue
			}
			key := generateMapControlKey(c.RefCode, c.ReferenceFramework)
			logx.FromContext(ctx).Warn().Str("key", key).Msg("adding control found in from subcontrols")
			if c.SystemOwned {
				systemOwnedMappedControls[key] = controlInfo
			} else {
				allInOrgControlMappings[key] = controlInfo
			}
		}

		for _, c := range r.Edges.ToSubcontrols {
			if c.ID == selfID {
				continue
			}
			controlInfo := subcontrolEdgeToControlInfo(c)
			if !shouldCheckForControl(controlInfo, frameworksInOrg) {
				continue
			}
			key := generateMapControlKey(c.RefCode, c.ReferenceFramework)
			logx.FromContext(ctx).Warn().Str("key", key).Msg("adding control found in to subcontrols")
			if c.SystemOwned {
				systemOwnedMappedControls[key] = controlInfo
			} else {
				allInOrgControlMappings[key] = controlInfo
			}
		}

		logx.FromContext(ctx).Warn().Int("count_system_owned", len(systemOwnedMappedControls)).Int("count_org_owned", len(allInOrgControlMappings)).Msg("found mapped controls, getting org info")
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

func getEvidenceStatus(ctx context.Context, c *model.ControlReport, isSubcontrol bool) error {
	orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
	if err != nil {
		return err
	}

	controlIDs := []string{}
	subcontrolIDs := []string{}

	if isSubcontrol {
		subcontrolIDs = append(subcontrolIDs, c.ID)
	} else {
		controlIDs = append(controlIDs, c.ID)
	}

	for _, rc := range c.RelatedControls {
		if rc.IsSubcontrol {
			subcontrolIDs = append(subcontrolIDs, rc.ID)
		} else {
			controlIDs = append(controlIDs, rc.ID)
		}
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	edgePredicates := []predicate.Evidence{}
	if len(controlIDs) > 0 {
		edgePredicates = append(edgePredicates, evidence.HasControlsWith(control.IDIn(controlIDs...)))
	}

	if len(subcontrolIDs) > 0 {
		edgePredicates = append(edgePredicates, evidence.HasSubcontrolsWith(subcontrol.IDIn(subcontrolIDs...)))
	}

	evidenceList, err := withTransactionalMutation(ctx).Evidence.Query().Where(
		evidence.OwnerIDIn(orgIDs...),
		evidence.Or(edgePredicates...),
	).All(allowCtx)
	if err != nil {
		return err
	}

	c.EvidenceStatus = &model.ControlEvidence{
		TotalCount:  len(evidenceList),
		WorstStatus: worstEvidenceStatus(evidenceList),
	}

	return nil
}

func getLinkedPolicies(ctx context.Context, c *model.ControlReport, isSubcontrol bool) error {
	orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
	if err != nil {
		return err
	}

	controlIDs := []string{}
	subcontrolIDs := []string{}

	if isSubcontrol {
		subcontrolIDs = append(subcontrolIDs, c.ID)
	} else {
		controlIDs = append(controlIDs, c.ID)
	}

	for _, rc := range c.RelatedControls {
		if rc.IsSubcontrol {
			subcontrolIDs = append(subcontrolIDs, rc.ID)
		} else {
			controlIDs = append(controlIDs, rc.ID)
		}
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	edgePredicates := []predicate.InternalPolicy{}
	if len(controlIDs) > 0 {
		edgePredicates = append(edgePredicates, internalpolicy.HasControlsWith(control.IDIn(controlIDs...)))
	}

	if len(subcontrolIDs) > 0 {
		edgePredicates = append(edgePredicates, internalpolicy.HasSubcontrolsWith(subcontrol.IDIn(subcontrolIDs...)))
	}

	policies, err := withTransactionalMutation(ctx).InternalPolicy.Query().Where(
		internalpolicy.OwnerIDIn(orgIDs...),
		internalpolicy.Or(edgePredicates...),
	).Select("id", "name", "status").All(allowCtx)
	if err != nil {
		return err
	}

	policySummaries := []*model.PolicySummary{}
	for _, p := range policies {
		policySummaries = append(policySummaries, &model.PolicySummary{
			ID:     p.ID,
			Name:   p.Name,
			Status: p.Status,
		})
	}

	c.LinkedPolicies = &model.ControlPolicies{
		TotalCount:       len(policies),
		InternalPolicies: policySummaries,
	}

	return nil
}

func convertControlToControlReportEdge(controls *generated.ControlConnection) *model.ControlReportConnection {
	edges := make([]*model.ControlReportEdge, len(controls.Edges))

	for i, c := range controls.Edges {
		edges[i] = &model.ControlReportEdge{
			Node: &model.ControlReport{
				ID:                 c.Node.ID,
				RefCode:            c.Node.RefCode,
				Description:        &c.Node.Description,
				Title:              &c.Node.Title,
				Status:             &c.Node.Status,
				ControlOwner:       c.Node.Edges.ControlOwner,
				ReferenceFramework: c.Node.ReferenceFramework,
				Category:           &c.Node.Category,
				Subcategory:        &c.Node.Subcategory,
				RelatedControls:    []*model.ControlInfo{},
				EvidenceStatus:     &model.ControlEvidence{},
				LinkedPolicies:     &model.ControlPolicies{},
				Subcontrols:        convertSubcontrolToControlReportEdge(c.Node.Edges.Subcontrols),
			},
		}
	}

	return &model.ControlReportConnection{
		Edges:      edges,
		TotalCount: controls.TotalCount,
	}
}

func convertSubcontrolToControlReportEdge(controls []*generated.Subcontrol) []*model.ControlReport {
	edges := make([]*model.ControlReport, len(controls))

	for i, c := range controls {
		edges[i] = &model.ControlReport{
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

	return edges
}

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
		"relatedControls",
		"evidenceStatus",
		"linkedPolicies",
		"subcontrols.relatedControls",
		"subcontrols.evidenceStatus",
		"subcontrols.linkedPolicies",
	)
}

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

	for i, c := range reports {
		if needsRelated {
			reports[i].RelatedControls, err = getRelatedControlInfoFromReport(ctx, c, frameworksInOrg)
			if err != nil {
				return err
			}
		}

		if wantsEvidence {
			if err := getEvidenceStatus(ctx, reports[i], false); err != nil {
				return err
			}
		}

		if wantsPolicies {
			if err := getLinkedPolicies(ctx, reports[i], false); err != nil {
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

			if wantsScEvidence {
				if err := getEvidenceStatus(ctx, reports[i].Subcontrols[j], true); err != nil {
					return err
				}
			}

			if wantsScPolicies {
				if err := getLinkedPolicies(ctx, reports[i].Subcontrols[j], true); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func getRelatedControlInfoFromReport(ctx context.Context, c *model.ControlReport, frameworksInOrg []string) ([]*model.ControlInfo, error) {
	result, err := getControlMappings(ctx, c.RefCode, c.ReferenceFramework, nil)
	if err != nil {
		return nil, err
	}

	return processMappedControlResults(ctx, result, c.ID, c.RefCode, c.ReferenceFramework, frameworksInOrg)
}

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
