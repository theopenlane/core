package graphapi

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/mappedcontrol"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/internal/ent/generated/subcontrol"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/rout"
	"golang.org/x/mod/semver"
)

// calculateFrameworkUpgradeDiff compares controls in an organization against a new standard revision
// and returns a diff of changes
func (r *Resolver) calculateFrameworkUpgradeDiff(ctx context.Context, standardID string, targetRevision *string) (*model.FrameworkUpgradeDiff, error) {
	logger := logx.FromContext(ctx)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil || orgID == "" {
		return nil, rout.NewMissingRequiredFieldError("owner_id")
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	// Get the current standard by ID
	currentStandard, err := withTransactionalMutation(ctx).Standard.Query().
		Where(standard.ID(standardID)).
		WithControls(func(cq *generated.ControlQuery) {
			cq.Where(control.OwnerID(orgID), control.DeletedAtIsNil())
		}).
		Only(allowCtx)
	if err != nil {
		logger.Error().Err(err).Str("standard_id", standardID).Msg("error retrieving current standard")
		return nil, err
	}

	// Find the target revision of the same standard
	targetStandardQuery := withTransactionalMutation(ctx).Standard.Query().
		Where(
			standard.ShortName(currentStandard.ShortName),
			standard.DeletedAtIsNil(),
		)

	var targetStandard *generated.Standard
	if targetRevision != nil {
		targetStandardQuery = targetStandardQuery.Where(standard.Revision(*targetRevision))
		targetStandard, err = targetStandardQuery.
			WithControls(func(cq *generated.ControlQuery) {
				if currentStandard.IsPublic {
					cq.Where(control.SystemOwned(true), control.DeletedAtIsNil())
				} else {
					cq.Where(control.OwnerID(currentStandard.OwnerID), control.DeletedAtIsNil())
				}
			}).
			First(allowCtx)
	} else {
		var candidates []*generated.Standard
		candidates, err = targetStandardQuery.All(allowCtx)
		if err != nil {
			logger.Error().Err(err).Str("short_name", currentStandard.ShortName).Msg("error retrieving target standard revisions")
			return nil, err
		}

		var latest *generated.Standard
		latest, err = latestStandardByRevision(candidates)
		if err != nil {
			logger.Error().Err(err).Str("short_name", currentStandard.ShortName).Msg("error selecting latest standard revision")
			return nil, err
		}

		targetStandard, err = withTransactionalMutation(ctx).Standard.Query().
			Where(standard.ID(latest.ID)).
			WithControls(func(cq *generated.ControlQuery) {
				if currentStandard.IsPublic {
					cq.Where(control.SystemOwned(true), control.DeletedAtIsNil())
				} else {
					cq.Where(control.OwnerID(currentStandard.OwnerID), control.DeletedAtIsNil())
				}
			}).
			Only(allowCtx)
	}
	if err != nil {
		logger.Error().Err(err).Str("short_name", currentStandard.ShortName).Msg("error retrieving target standard revision")
		return nil, err
	}

	// Check if trying to upgrade to the same revision
	if currentStandard.Revision == targetStandard.Revision {
		return nil, fmt.Errorf("%w: current and target revisions are the same", common.ErrInvalidInput)
	}

	// Get controls currently in the organization for this standard
	orgControls := currentStandard.Edges.Controls

	// Get controls from the target standard revision
	targetControls := targetStandard.Edges.Controls

	// Build lookup maps by refCode for efficient comparison
	orgControlMap := make(map[string]*generated.Control)
	for _, c := range orgControls {
		orgControlMap[c.RefCode] = c
	}

	targetControlMap := make(map[string]*generated.Control)
	for _, c := range targetControls {
		targetControlMap[c.RefCode] = c
	}

	diff := &model.FrameworkUpgradeDiff{
		CurrentStandard:   currentStandard,
		TargetStandard:    targetStandard,
		AddedControls:     []*generated.Control{},
		UpdatedControls:   []*model.ControlUpdateChange{},
		RemovedControls:   []*generated.Control{},
		UnchangedControls: []*generated.Control{},
	}

	// Find added and updated controls
	for refCode, targetControl := range targetControlMap {
		if orgControl, exists := orgControlMap[refCode]; exists {
			// Control exists in both - check if it has changed
			if changes := compareControls(orgControl, targetControl); len(changes) > 0 {
				diff.UpdatedControls = append(diff.UpdatedControls, &model.ControlUpdateChange{
					CurrentControl: orgControl,
					TargetControl:  targetControl,
					Changes:        changes,
				})
			} else {
				diff.UnchangedControls = append(diff.UnchangedControls, orgControl)
			}
		} else {
			// Control exists in target but not in org - it is new
			diff.AddedControls = append(diff.AddedControls, targetControl)
		}
	}

	// Find removed controls
	for refCode, orgControl := range orgControlMap {
		if _, exists := targetControlMap[refCode]; !exists {
			// Control exists in org but not in target - it was removed
			diff.RemovedControls = append(diff.RemovedControls, orgControl)
		}
	}

	// Calculate summary
	addedCount := len(diff.AddedControls)
	updatedCount := len(diff.UpdatedControls)
	removedCount := len(diff.RemovedControls)
	unchangedCount := len(diff.UnchangedControls)

	diff.Summary = &model.FrameworkUpgradeSummary{
		AddedCount:     addedCount,
		UpdatedCount:   updatedCount,
		RemovedCount:   removedCount,
		UnchangedCount: unchangedCount,
		TotalAffected:  addedCount + updatedCount + removedCount,
	}

	return diff, nil
}

// compareControls compares two controls and returns a list of changed fields
func compareControls(current, target *generated.Control) []string {
	changes := []string{}

	if current.Title != target.Title {
		changes = append(changes, "title")
	}

	if current.Description != target.Description {
		changes = append(changes, "description")
	}

	if !descriptionJSONEqual(current.DescriptionJSON, target.DescriptionJSON) {
		changes = append(changes, "description_json")
	}

	if current.Source != target.Source {
		changes = append(changes, "source")
	}

	if current.ControlType != target.ControlType {
		changes = append(changes, "control_type")
	}

	if current.Category != target.Category {
		changes = append(changes, "category")
	}

	if current.CategoryID != target.CategoryID {
		changes = append(changes, "category_id")
	}

	if current.Subcategory != target.Subcategory {
		changes = append(changes, "subcategory")
	}

	if !stringSlicesEqual(current.Tags, target.Tags) {
		changes = append(changes, "tags")
	}

	if !stringSlicesEqual(current.Aliases, target.Aliases) {
		changes = append(changes, "aliases")
	}

	if !stringSlicesEqual(current.MappedCategories, target.MappedCategories) {
		changes = append(changes, "mapped_categories")
	}

	if !assessmentObjectivesEqual(current.AssessmentObjectives, target.AssessmentObjectives) {
		changes = append(changes, "assessment_objectives")
	}

	if !assessmentMethodsEqual(current.AssessmentMethods, target.AssessmentMethods) {
		changes = append(changes, "assessment_methods")
	}

	if !stringSlicesEqual(current.ControlQuestions, target.ControlQuestions) {
		changes = append(changes, "control_questions")
	}

	if !implementationGuidanceEqual(current.ImplementationGuidance, target.ImplementationGuidance) {
		changes = append(changes, "implementation_guidance")
	}

	if !exampleEvidenceEqual(current.ExampleEvidence, target.ExampleEvidence) {
		changes = append(changes, "example_evidence")
	}

	if !referencesEqual(current.References, target.References) {
		changes = append(changes, "references")
	}

	if !testingProceduresEqual(current.TestingProcedures, target.TestingProcedures) {
		changes = append(changes, "testing_procedures")
	}

	if !evidenceRequestsEqual(current.EvidenceRequests, target.EvidenceRequests) {
		changes = append(changes, "evidence_requests")
	}

	return changes
}

// compareSubcontrols compares two subcontrols and returns a list of changed fields
func compareSubcontrols(current, target *generated.Subcontrol) []string {
	changes := []string{}

	if current.Title != target.Title {
		changes = append(changes, "title")
	}

	if current.Description != target.Description {
		changes = append(changes, "description")
	}

	if !descriptionJSONEqual(current.DescriptionJSON, target.DescriptionJSON) {
		changes = append(changes, "description_json")
	}

	if current.Source != target.Source {
		changes = append(changes, "source")
	}

	if current.ControlType != target.ControlType {
		changes = append(changes, "control_type")
	}

	if current.Category != target.Category {
		changes = append(changes, "category")
	}

	if current.CategoryID != target.CategoryID {
		changes = append(changes, "category_id")
	}

	if current.Subcategory != target.Subcategory {
		changes = append(changes, "subcategory")
	}

	if !stringSlicesEqual(current.Tags, target.Tags) {
		changes = append(changes, "tags")
	}

	if !stringSlicesEqual(current.Aliases, target.Aliases) {
		changes = append(changes, "aliases")
	}

	if !stringSlicesEqual(current.MappedCategories, target.MappedCategories) {
		changes = append(changes, "mapped_categories")
	}

	if !assessmentObjectivesEqual(current.AssessmentObjectives, target.AssessmentObjectives) {
		changes = append(changes, "assessment_objectives")
	}

	if !assessmentMethodsEqual(current.AssessmentMethods, target.AssessmentMethods) {
		changes = append(changes, "assessment_methods")
	}

	if !stringSlicesEqual(current.ControlQuestions, target.ControlQuestions) {
		changes = append(changes, "control_questions")
	}

	if !implementationGuidanceEqual(current.ImplementationGuidance, target.ImplementationGuidance) {
		changes = append(changes, "implementation_guidance")
	}

	if !exampleEvidenceEqual(current.ExampleEvidence, target.ExampleEvidence) {
		changes = append(changes, "example_evidence")
	}

	if !referencesEqual(current.References, target.References) {
		changes = append(changes, "references")
	}

	if !testingProceduresEqual(current.TestingProcedures, target.TestingProcedures) {
		changes = append(changes, "testing_procedures")
	}

	if !evidenceRequestsEqual(current.EvidenceRequests, target.EvidenceRequests) {
		changes = append(changes, "evidence_requests")
	}

	return changes
}

// stringSlicesEqual compares two string slices for equality
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aSet := make(map[string]bool)
	for _, s := range a {
		aSet[s] = true
	}

	for _, s := range b {
		if !aSet[s] {
			return false
		}
	}

	return true
}

// descriptionJSONEqual compares two JSON slices for equality
func descriptionJSONEqual(a, b []interface{}) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}

	return reflect.DeepEqual(a, b)
}

// assessmentObjectivesEqual compares two AssessmentObjective slices for equality
func assessmentObjectivesEqual(a, b []models.AssessmentObjective) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]models.AssessmentObjective)
	for _, obj := range a {
		aMap[obj.ID] = obj
	}

	for _, obj := range b {
		existing, exists := aMap[obj.ID]
		if !exists || existing.Class != obj.Class || existing.Objective != obj.Objective {
			return false
		}
	}

	return true
}

// assessmentMethodsEqual compares two AssessmentMethod slices for equality
func assessmentMethodsEqual(a, b []models.AssessmentMethod) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]models.AssessmentMethod)
	for _, method := range a {
		aMap[method.ID] = method
	}

	for _, method := range b {
		existing, exists := aMap[method.ID]
		if !exists || existing.Type != method.Type || existing.Method != method.Method {
			return false
		}
	}

	return true
}

// exampleEvidenceEqual compares two ExampleEvidence slices for equality
func exampleEvidenceEqual(a, b []models.ExampleEvidence) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]models.ExampleEvidence)
	for _, evidence := range a {
		key := evidence.DocumentationType + ":" + evidence.Description
		aMap[key] = evidence
	}

	for _, evidence := range b {
		key := evidence.DocumentationType + ":" + evidence.Description
		if _, exists := aMap[key]; !exists {
			return false
		}
	}

	return true
}

// referencesEqual compares two Reference slices for equality
func referencesEqual(a, b []models.Reference) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]models.Reference)
	for _, ref := range a {
		key := ref.Name + ":" + ref.URL
		aMap[key] = ref
	}

	for _, ref := range b {
		key := ref.Name + ":" + ref.URL
		if _, exists := aMap[key]; !exists {
			return false
		}
	}

	return true
}

// implementationGuidanceEqual compares two ImplementationGuidance slices for equality
func implementationGuidanceEqual(a, b []models.ImplementationGuidance) bool {
	if len(a) != len(b) {
		return false
	}

	aSorted := append([]models.ImplementationGuidance(nil), a...)
	bSorted := append([]models.ImplementationGuidance(nil), b...)

	models.Sort(aSorted)
	models.Sort(bSorted)

	for i := range aSorted {
		if aSorted[i].ReferenceID != bSorted[i].ReferenceID {
			return false
		}
		if !reflect.DeepEqual(aSorted[i].Guidance, bSorted[i].Guidance) {
			return false
		}
	}

	return true
}

// testingProceduresEqual compares two TestingProcedures slices for equality
func testingProceduresEqual(a, b []models.TestingProcedures) bool {
	if len(a) != len(b) {
		return false
	}

	aSorted := append([]models.TestingProcedures(nil), a...)
	bSorted := append([]models.TestingProcedures(nil), b...)

	models.Sort(aSorted)
	models.Sort(bSorted)

	for i := range aSorted {
		if aSorted[i].ReferenceID != bSorted[i].ReferenceID {
			return false
		}
		if !reflect.DeepEqual(aSorted[i].Procedures, bSorted[i].Procedures) {
			return false
		}
	}

	return true
}

// evidenceRequestsEqual compares two EvidenceRequests slices for equality
func evidenceRequestsEqual(a, b []models.EvidenceRequests) bool {
	if len(a) != len(b) {
		return false
	}

	aSorted := append([]models.EvidenceRequests(nil), a...)
	bSorted := append([]models.EvidenceRequests(nil), b...)

	models.Sort(aSorted)
	models.Sort(bSorted)

	for i := range aSorted {
		if aSorted[i].EvidenceRequestID != bSorted[i].EvidenceRequestID {
			return false
		}
		if aSorted[i].DocumentationArtifact != bSorted[i].DocumentationArtifact {
			return false
		}
		if aSorted[i].ArtifactDescription != bSorted[i].ArtifactDescription {
			return false
		}
		if aSorted[i].AreaOfFocus != bSorted[i].AreaOfFocus {
			return false
		}
	}

	return true
}

func latestStandardByRevision(standards []*generated.Standard) (*generated.Standard, error) {
	if len(standards) == 0 {
		return nil, ErrStandardNotFound
	}

	latest := standards[0]
	for _, candidate := range standards[1:] {
		if semver.Compare(candidate.Revision, latest.Revision) > 0 {
			latest = candidate
		}
	}

	return latest, nil
}

// applyFrameworkUpgrade applies the calculated diff to upgrade controls in an organization
// to a new standard revision
func (r *Resolver) applyFrameworkUpgrade(ctx context.Context, standardID string, targetRevision *string) error {
	logger := logx.FromContext(ctx)

	if err := r.ensureFrameworkUpgradeAccess(ctx); err != nil {
		return err
	}

	// Calculate the diff first
	diff, err := r.calculateFrameworkUpgradeDiff(ctx, standardID, targetRevision)
	if err != nil {
		return err
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil || orgID == "" {
		return rout.NewMissingRequiredFieldError("owner_id")
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	// Process added controls - clone them into the organization
	if len(diff.AddedControls) > 0 {
		logger.Info().Int("count", len(diff.AddedControls)).Msg("cloning new controls from target revision")

		mr := &mutationResolver{r}
		_, err = mr.cloneControls(ctx, diff.AddedControls, nil)
		if err != nil {
			logger.Error().Err(err).Msg("error cloning new controls")
			return err
		}
	}

	// Process updated controls - update them in place
	for _, updateDiff := range diff.UpdatedControls {
		logger.Info().Str("ref_code", updateDiff.CurrentControl.RefCode).Strs("changes", updateDiff.Changes).Msg("updating control")

		updateInput := buildControlUpdateInput(updateDiff.CurrentControl, updateDiff.TargetControl)

		err = withTransactionalMutation(ctx).Control.UpdateOneID(updateDiff.CurrentControl.ID).
			SetInput(updateInput).
			Exec(allowCtx)
		if err != nil {
			logger.Error().Err(err).Str("control_id", updateDiff.CurrentControl.ID).Msg("error updating control")
			return err
		}
	}

	if err := r.syncFrameworkSubcontrols(ctx, diff, orgID); err != nil {
		logger.Error().Err(err).Msg("error syncing subcontrols for framework upgrade")
		return err
	}

	if err := r.syncFrameworkMappings(ctx, diff, orgID, diff.RemovedControls); err != nil {
		logger.Error().Err(err).Msg("error syncing mapped controls for framework upgrade")
		return err
	}

	// Process removed controls - delete them
	if len(diff.RemovedControls) > 0 {
		logger.Info().Int("count", len(diff.RemovedControls)).Msg("removing controls no longer in target revision")

		controlIDs := lo.Map(diff.RemovedControls, func(c *generated.Control, _ int) string {
			return c.ID
		})

		_, err = withTransactionalMutation(ctx).Control.Delete().
			Where(control.IDIn(controlIDs...)).
			Exec(allowCtx)
		if err != nil {
			logger.Error().Err(err).Msg("error deleting removed controls")
			return err
		}
	}

	if err := withTransactionalMutation(ctx).Control.Update().
		Where(
			control.OwnerID(orgID),
			control.DeletedAtIsNil(),
			control.HasStandardWith(standard.ShortName(diff.CurrentStandard.ShortName)),
		).
		SetReferenceFrameworkRevision(diff.TargetStandard.Revision).
		Exec(allowCtx); err != nil {
		return err
	}

	if err := withTransactionalMutation(ctx).Subcontrol.Update().
		Where(
			subcontrol.OwnerID(orgID),
			subcontrol.DeletedAtIsNil(),
			subcontrol.HasControlWith(control.HasStandardWith(standard.ShortName(diff.CurrentStandard.ShortName))),
		).
		SetReferenceFrameworkRevision(diff.TargetStandard.Revision).
		Exec(allowCtx); err != nil {
		return err
	}

	return nil
}

func (r *Resolver) ensureFrameworkUpgradeAccess(ctx context.Context) error {
	if auth.IsSystemAdminFromContext(ctx) {
		return nil
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil || orgID == "" {
		return rout.NewMissingRequiredFieldError("owner_id")
	}

	userID, err := auth.GetSubjectIDFromContext(ctx)
	if err != nil || userID == "" {
		return err
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	membership, err := withTransactionalMutation(ctx).OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(orgID),
			orgmembership.UserID(userID),
		).
		Only(allowCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			return generated.ErrPermissionDenied
		}
		return err
	}

	if membership.Role != enums.RoleOwner && membership.Role != enums.RoleAdmin {
		return generated.ErrPermissionDenied
	}

	return nil
}

func (r *Resolver) syncFrameworkSubcontrols(ctx context.Context, diff *model.FrameworkUpgradeDiff, orgID string) error {
	logger := logx.FromContext(ctx)
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	orgControls, err := withTransactionalMutation(ctx).Control.Query().
		Where(
			control.OwnerID(orgID),
			control.DeletedAtIsNil(),
			control.StandardID(diff.CurrentStandard.ID),
		).
		WithSubcontrols(func(sq *generated.SubcontrolQuery) {
			sq.Where(subcontrol.OwnerID(orgID), subcontrol.DeletedAtIsNil())
		}).
		All(allowCtx)
	if err != nil {
		logger.Error().Err(err).Msg("error retrieving org controls for subcontrol sync")
		return err
	}

	targetControlsQuery := withTransactionalMutation(ctx).Control.Query().
		Where(
			control.StandardID(diff.TargetStandard.ID),
			control.DeletedAtIsNil(),
		)
	if diff.TargetStandard.IsPublic {
		targetControlsQuery = targetControlsQuery.Where(control.SystemOwned(true))
	} else {
		targetControlsQuery = targetControlsQuery.Where(control.OwnerID(diff.TargetStandard.OwnerID))
	}

	targetControls, err := targetControlsQuery.WithSubcontrols(func(sq *generated.SubcontrolQuery) {
		sq.Where(subcontrol.DeletedAtIsNil())
		if diff.TargetStandard.IsPublic {
			sq.Where(subcontrol.SystemOwned(true))
		} else {
			sq.Where(subcontrol.OwnerID(diff.TargetStandard.OwnerID))
		}
	}).All(allowCtx)
	if err != nil {
		logger.Error().Err(err).Msg("error retrieving target controls for subcontrol sync")
		return err
	}

	orgControlMap := make(map[string]*generated.Control)
	for _, c := range orgControls {
		orgControlMap[c.RefCode] = c
	}

	targetControlMap := make(map[string]*generated.Control)
	for _, c := range targetControls {
		targetControlMap[c.RefCode] = c
	}

	for refCode, targetControl := range targetControlMap {
		orgControl, exists := orgControlMap[refCode]
		if !exists {
			continue
		}

		if err := syncControlSubcontrols(ctx, orgControl, targetControl, diff.TargetStandard, orgID); err != nil {
			logger.Error().Err(err).Str("control_ref_code", refCode).Msg("error syncing subcontrols")
			return err
		}
	}

	return nil
}

func syncControlSubcontrols(ctx context.Context, orgControl, targetControl *generated.Control, targetStandard *generated.Standard, orgID string) error {
	logger := logx.FromContext(ctx)
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	orgSubcontrolMap := make(map[string]*generated.Subcontrol)
	for _, sc := range orgControl.Edges.Subcontrols {
		orgSubcontrolMap[sc.RefCode] = sc
	}

	targetSubcontrolMap := make(map[string]*generated.Subcontrol)
	for _, sc := range targetControl.Edges.Subcontrols {
		targetSubcontrolMap[sc.RefCode] = sc
	}

	for refCode, targetSubcontrol := range targetSubcontrolMap {
		if orgSubcontrol, exists := orgSubcontrolMap[refCode]; exists {
			if changes := compareSubcontrols(orgSubcontrol, targetSubcontrol); len(changes) > 0 {
				updateInput := buildSubcontrolUpdateInput(orgSubcontrol, targetSubcontrol)
				if err := withTransactionalMutation(ctx).Subcontrol.UpdateOneID(orgSubcontrol.ID).
					SetInput(updateInput).
					Exec(allowCtx); err != nil {
					logger.Error().Err(err).Str("subcontrol_id", orgSubcontrol.ID).Msg("error updating subcontrol")
					return err
				}
			}
			continue
		}

		createInput := buildSubcontrolCreateInput(targetSubcontrol, orgControl.ID, orgID, targetStandard)
		if err := withTransactionalMutation(ctx).Subcontrol.Create().SetInput(*createInput).Exec(allowCtx); err != nil {
			logger.Error().Err(err).Str("control_id", orgControl.ID).Msg("error creating subcontrol")
			return err
		}
	}

	removedIDs := []string{}
	for refCode, orgSubcontrol := range orgSubcontrolMap {
		if _, exists := targetSubcontrolMap[refCode]; !exists {
			removedIDs = append(removedIDs, orgSubcontrol.ID)
		}
	}

	if len(removedIDs) > 0 {
		if _, err := withTransactionalMutation(ctx).Subcontrol.Delete().
			Where(subcontrol.IDIn(removedIDs...)).
			Exec(allowCtx); err != nil {
			logger.Error().Err(err).Str("control_id", orgControl.ID).Msg("error deleting removed subcontrols")
			return err
		}
	}

	return nil
}

func buildSubcontrolCreateInput(target *generated.Subcontrol, controlID, orgID string, standard *generated.Standard) *generated.CreateSubcontrolInput {
	input := &generated.CreateSubcontrolInput{
		Tags:                   target.Tags,
		RefCode:                target.RefCode,
		Title:                  &target.Title,
		Aliases:                target.Aliases,
		Description:            &target.Description,
		DescriptionJSON:        target.DescriptionJSON,
		Source:                 &target.Source,
		ControlID:              controlID,
		ControlType:            &target.ControlType,
		Category:               &target.Category,
		CategoryID:             &target.CategoryID,
		Subcategory:            &target.Subcategory,
		MappedCategories:       target.MappedCategories,
		AssessmentObjectives:   target.AssessmentObjectives,
		AssessmentMethods:      target.AssessmentMethods,
		ControlQuestions:       target.ControlQuestions,
		ImplementationGuidance: target.ImplementationGuidance,
		ExampleEvidence:        target.ExampleEvidence,
		References:             target.References,
		TestingProcedures:      target.TestingProcedures,
		EvidenceRequests:       target.EvidenceRequests,
		Status:                 lo.ToPtr(enums.ControlStatusNotImplemented),
		OwnerID:                &orgID,
	}

	if standard != nil {
		input.ReferenceFramework = &standard.ShortName
		input.ReferenceFrameworkRevision = &standard.Revision
	} else if target.ReferenceFramework != nil {
		input.ReferenceFramework = target.ReferenceFramework
	}

	if input.ReferenceFrameworkRevision == nil && target.ReferenceFrameworkRevision != nil {
		input.ReferenceFrameworkRevision = target.ReferenceFrameworkRevision
	}

	return input
}

func (r *Resolver) syncFrameworkMappings(ctx context.Context, diff *model.FrameworkUpgradeDiff, orgID string, removedControls []*generated.Control) error {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	allowCtx = contextx.With(allowCtx, hooks.SuggestedMappingContextKey{})

	if err := deleteMappingsForRemovedControls(ctx, orgID, removedControls); err != nil {
		return err
	}

	nonManualIDs, err := withTransactionalMutation(ctx).MappedControl.Query().
		Where(
			mappedcontrol.OwnerID(orgID),
			mappedcontrol.DeletedAtIsNil(),
			mappedcontrol.SourceNEQ(enums.MappingSourceManual),
			mappedcontrol.Or(
				mappedcontrol.HasFromControlsWith(control.HasStandardWith(standard.ShortName(diff.CurrentStandard.ShortName))),
				mappedcontrol.HasToControlsWith(control.HasStandardWith(standard.ShortName(diff.CurrentStandard.ShortName))),
				mappedcontrol.HasFromSubcontrolsWith(subcontrol.HasControlWith(control.HasStandardWith(standard.ShortName(diff.CurrentStandard.ShortName)))),
				mappedcontrol.HasToSubcontrolsWith(subcontrol.HasControlWith(control.HasStandardWith(standard.ShortName(diff.CurrentStandard.ShortName)))),
			),
		).
		IDs(allowCtx)
	if err != nil {
		return err
	}

	if err := deleteMappedControlIDs(ctx, nonManualIDs); err != nil {
		return err
	}

	frameworkMappings, err := queryFrameworkMappings(ctx, diff)
	if err != nil {
		return err
	}

	manualKeys, err := loadManualMappingKeys(ctx, diff, orgID)
	if err != nil {
		return err
	}

	inputs, err := buildFrameworkMappingInputs(ctx, frameworkMappings, orgID, manualKeys)
	if err != nil {
		return err
	}

	if len(inputs) == 0 {
		return nil
	}

	mr := &mutationResolver{r}
	_, err = mr.bulkCreateMappedControl(allowCtx, inputs)
	return err
}

func deleteMappingsForRemovedControls(ctx context.Context, orgID string, removedControls []*generated.Control) error {
	if len(removedControls) == 0 {
		return nil
	}

	removedControlIDs := lo.Map(removedControls, func(c *generated.Control, _ int) string {
		return c.ID
	})

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	mappingIDs, err := withTransactionalMutation(ctx).MappedControl.Query().
		Where(
			mappedcontrol.OwnerID(orgID),
			mappedcontrol.DeletedAtIsNil(),
			mappedcontrol.Or(
				mappedcontrol.HasFromControlsWith(control.IDIn(removedControlIDs...)),
				mappedcontrol.HasToControlsWith(control.IDIn(removedControlIDs...)),
				mappedcontrol.HasFromSubcontrolsWith(subcontrol.HasControlWith(control.IDIn(removedControlIDs...))),
				mappedcontrol.HasToSubcontrolsWith(subcontrol.HasControlWith(control.IDIn(removedControlIDs...))),
			),
		).
		IDs(allowCtx)
	if err != nil {
		return err
	}

	return deleteMappedControlIDs(ctx, mappingIDs)
}

func deleteMappedControlIDs(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	uniqueIDs := lo.Uniq(ids)

	for _, id := range uniqueIDs {
		if err := withTransactionalMutation(ctx).MappedControl.DeleteOneID(id).Exec(allowCtx); err != nil {
			return err
		}
		if err := generated.MappedControlEdgeCleanup(ctx, id); err != nil {
			return err
		}
	}

	return nil
}

func queryFrameworkMappings(ctx context.Context, diff *model.FrameworkUpgradeDiff) ([]*generated.MappedControl, error) {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	query := withTransactionalMutation(ctx).MappedControl.Query().
		Where(
			mappedcontrol.DeletedAtIsNil(),
			mappedcontrol.SourceNEQ(enums.MappingSourceManual),
			mappedcontrol.Or(
				mappedcontrol.HasFromControlsWith(control.StandardID(diff.TargetStandard.ID)),
				mappedcontrol.HasToControlsWith(control.StandardID(diff.TargetStandard.ID)),
				mappedcontrol.HasFromSubcontrolsWith(subcontrol.HasControlWith(control.StandardID(diff.TargetStandard.ID))),
				mappedcontrol.HasToSubcontrolsWith(subcontrol.HasControlWith(control.StandardID(diff.TargetStandard.ID))),
			),
		).
		WithFromControls(func(cq *generated.ControlQuery) {
			cq.WithStandard()
		}).
		WithToControls(func(cq *generated.ControlQuery) {
			cq.WithStandard()
		}).
		WithFromSubcontrols(func(sq *generated.SubcontrolQuery) {
			sq.WithControl(func(cq *generated.ControlQuery) {
				cq.WithStandard()
			})
		}).
		WithToSubcontrols(func(sq *generated.SubcontrolQuery) {
			sq.WithControl(func(cq *generated.ControlQuery) {
				cq.WithStandard()
			})
		})

	if diff.TargetStandard.IsPublic {
		query = query.Where(mappedcontrol.SystemOwned(true))
	} else {
		query = query.Where(mappedcontrol.OwnerID(diff.TargetStandard.OwnerID))
	}

	return query.All(allowCtx)
}

func loadManualMappingKeys(ctx context.Context, diff *model.FrameworkUpgradeDiff, orgID string) (map[string]struct{}, error) {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	manualMappings, err := withTransactionalMutation(ctx).MappedControl.Query().
		Where(
			mappedcontrol.OwnerID(orgID),
			mappedcontrol.DeletedAtIsNil(),
			mappedcontrol.SourceEQ(enums.MappingSourceManual),
			mappedcontrol.Or(
				mappedcontrol.HasFromControlsWith(control.HasStandardWith(standard.ShortName(diff.CurrentStandard.ShortName))),
				mappedcontrol.HasToControlsWith(control.HasStandardWith(standard.ShortName(diff.CurrentStandard.ShortName))),
				mappedcontrol.HasFromSubcontrolsWith(subcontrol.HasControlWith(control.HasStandardWith(standard.ShortName(diff.CurrentStandard.ShortName)))),
				mappedcontrol.HasToSubcontrolsWith(subcontrol.HasControlWith(control.HasStandardWith(standard.ShortName(diff.CurrentStandard.ShortName)))),
			),
		).
		WithFromControls().
		WithToControls().
		WithFromSubcontrols().
		WithToSubcontrols().
		All(allowCtx)
	if err != nil {
		return nil, err
	}

	keys := make(map[string]struct{}, len(manualMappings))
	for _, mapping := range manualMappings {
		keys[mappedControlKey(
			collectControlIDs(mapping.Edges.FromControls),
			collectSubcontrolIDs(mapping.Edges.FromSubcontrols),
			collectControlIDs(mapping.Edges.ToControls),
			collectSubcontrolIDs(mapping.Edges.ToSubcontrols),
		)] = struct{}{}
	}

	return keys, nil
}

func buildFrameworkMappingInputs(ctx context.Context, mappings []*generated.MappedControl, orgID string, manualKeys map[string]struct{}) ([]*generated.CreateMappedControlInput, error) {
	standardShortNames := collectMappingStandardShortNames(mappings)
	controlMap, subcontrolMap, err := buildOrgControlMaps(ctx, orgID, standardShortNames)
	if err != nil {
		return nil, err
	}

	inputs := []*generated.CreateMappedControlInput{}
	seen := map[string]struct{}{}

	for _, mapping := range mappings {
		fromControlIDs := mapControlRefs(mapping.Edges.FromControls, controlMap)
		fromSubcontrolIDs := mapSubcontrolRefs(mapping.Edges.FromSubcontrols, subcontrolMap)
		toControlIDs := mapControlRefs(mapping.Edges.ToControls, controlMap)
		toSubcontrolIDs := mapSubcontrolRefs(mapping.Edges.ToSubcontrols, subcontrolMap)

		if len(fromControlIDs) == 0 && len(fromSubcontrolIDs) == 0 {
			continue
		}
		if len(toControlIDs) == 0 && len(toSubcontrolIDs) == 0 {
			continue
		}

		key := mappedControlKey(fromControlIDs, fromSubcontrolIDs, toControlIDs, toSubcontrolIDs)
		if _, exists := manualKeys[key]; exists {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		input := &generated.CreateMappedControlInput{
			OwnerID:           &orgID,
			MappingType:       lo.ToPtr(mapping.MappingType),
			Source:            &mapping.Source,
			FromControlIDs:    fromControlIDs,
			ToControlIDs:      toControlIDs,
			FromSubcontrolIDs: fromSubcontrolIDs,
			ToSubcontrolIDs:   toSubcontrolIDs,
		}

		if mapping.Relation != "" {
			input.Relation = &mapping.Relation
		}

		if mapping.Confidence != nil {
			input.Confidence = mapping.Confidence
		}

		inputs = append(inputs, input)
	}

	return inputs, nil
}

func collectMappingStandardShortNames(mappings []*generated.MappedControl) []string {
	names := map[string]struct{}{}

	for _, mapping := range mappings {
		for _, control := range mapping.Edges.FromControls {
			if name := controlStandardShortName(control); name != "" {
				names[name] = struct{}{}
			}
		}
		for _, control := range mapping.Edges.ToControls {
			if name := controlStandardShortName(control); name != "" {
				names[name] = struct{}{}
			}
		}
		for _, sc := range mapping.Edges.FromSubcontrols {
			if name, _ := subcontrolStandardShortName(sc); name != "" {
				names[name] = struct{}{}
			}
		}
		for _, sc := range mapping.Edges.ToSubcontrols {
			if name, _ := subcontrolStandardShortName(sc); name != "" {
				names[name] = struct{}{}
			}
		}
	}

	return lo.Keys(names)
}

func buildOrgControlMaps(ctx context.Context, orgID string, standardShortNames []string) (map[string]string, map[string]string, error) {
	if len(standardShortNames) == 0 {
		return map[string]string{}, map[string]string{}, nil
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	orgControls, err := withTransactionalMutation(ctx).Control.Query().
		Where(
			control.OwnerID(orgID),
			control.DeletedAtIsNil(),
			control.HasStandardWith(standard.ShortNameIn(standardShortNames...)),
		).
		WithStandard().
		All(allowCtx)
	if err != nil {
		return nil, nil, err
	}

	controlMap := make(map[string]string, len(orgControls))
	controlIDs := make([]string, 0, len(orgControls))

	for _, c := range orgControls {
		shortName := controlStandardShortName(c)
		if shortName == "" {
			continue
		}
		controlMap[controlRefKey(shortName, c.RefCode)] = c.ID
		controlIDs = append(controlIDs, c.ID)
	}

	if len(controlIDs) == 0 {
		return controlMap, map[string]string{}, nil
	}

	orgSubcontrols, err := withTransactionalMutation(ctx).Subcontrol.Query().
		Where(
			subcontrol.OwnerID(orgID),
			subcontrol.DeletedAtIsNil(),
			subcontrol.HasControlWith(control.IDIn(controlIDs...)),
		).
		WithControl(func(cq *generated.ControlQuery) {
			cq.WithStandard()
		}).
		All(allowCtx)
	if err != nil {
		return nil, nil, err
	}

	subcontrolMap := make(map[string]string, len(orgSubcontrols))
	for _, sc := range orgSubcontrols {
		shortName, controlRef := subcontrolStandardShortName(sc)
		if shortName == "" || controlRef == "" {
			continue
		}
		subcontrolMap[subcontrolRefKey(shortName, controlRef, sc.RefCode)] = sc.ID
	}

	return controlMap, subcontrolMap, nil
}

func mapControlRefs(controls []*generated.Control, controlMap map[string]string) []string {
	ids := []string{}
	seen := map[string]struct{}{}

	for _, c := range controls {
		shortName := controlStandardShortName(c)
		if shortName == "" {
			continue
		}

		key := controlRefKey(shortName, c.RefCode)
		if id, ok := controlMap[key]; ok {
			if _, exists := seen[id]; exists {
				continue
			}
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}

	return ids
}

func mapSubcontrolRefs(subcontrols []*generated.Subcontrol, subcontrolMap map[string]string) []string {
	ids := []string{}
	seen := map[string]struct{}{}

	for _, sc := range subcontrols {
		shortName, controlRef := subcontrolStandardShortName(sc)
		if shortName == "" || controlRef == "" {
			continue
		}

		key := subcontrolRefKey(shortName, controlRef, sc.RefCode)
		if id, ok := subcontrolMap[key]; ok {
			if _, exists := seen[id]; exists {
				continue
			}
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}

	return ids
}

func controlStandardShortName(c *generated.Control) string {
	if c.Edges.Standard != nil {
		return c.Edges.Standard.ShortName
	}
	if c.ReferenceFramework != nil {
		return *c.ReferenceFramework
	}
	return ""
}

func subcontrolStandardShortName(sc *generated.Subcontrol) (string, string) {
	if sc.Edges.Control == nil {
		return "", ""
	}

	controlRef := sc.Edges.Control.RefCode
	if sc.Edges.Control.Edges.Standard != nil {
		return sc.Edges.Control.Edges.Standard.ShortName, controlRef
	}
	if sc.Edges.Control.ReferenceFramework != nil {
		return *sc.Edges.Control.ReferenceFramework, controlRef
	}

	return "", controlRef
}

func controlRefKey(shortName, refCode string) string {
	return shortName + "::" + refCode
}

func subcontrolRefKey(shortName, controlRefCode, refCode string) string {
	return shortName + "::" + controlRefCode + "::" + refCode
}

func mappedControlKey(fromControls, fromSubcontrols, toControls, toSubcontrols []string) string {
	fromControls = append([]string(nil), fromControls...)
	fromSubcontrols = append([]string(nil), fromSubcontrols...)
	toControls = append([]string(nil), toControls...)
	toSubcontrols = append([]string(nil), toSubcontrols...)

	sort.Strings(fromControls)
	sort.Strings(fromSubcontrols)
	sort.Strings(toControls)
	sort.Strings(toSubcontrols)

	return strings.Join([]string{
		strings.Join(fromControls, ","),
		strings.Join(fromSubcontrols, ","),
		strings.Join(toControls, ","),
		strings.Join(toSubcontrols, ","),
	}, "|")
}

func collectControlIDs(controls []*generated.Control) []string {
	return lo.Map(controls, func(c *generated.Control, _ int) string {
		return c.ID
	})
}

func collectSubcontrolIDs(subcontrols []*generated.Subcontrol) []string {
	return lo.Map(subcontrols, func(sc *generated.Subcontrol, _ int) string {
		return sc.ID
	})
}

// buildControlUpdateInput creates an UpdateControlInput from the current and target control
// preserving only the fields that should be updated from the standard
func buildControlUpdateInput(current, target *generated.Control) generated.UpdateControlInput {
	input := generated.UpdateControlInput{
		Title:       &target.Title,
		Description: &target.Description,
		Source:      &target.Source,
		ControlType: &target.ControlType,
		Category:    &target.Category,
		CategoryID:  &target.CategoryID,
		Subcategory: &target.Subcategory,
	}

	setSliceUpdate(current.Tags, target.Tags, &input.ClearTags, &input.Tags)
	setSliceUpdate(current.DescriptionJSON, target.DescriptionJSON, &input.ClearDescriptionJSON, &input.DescriptionJSON)
	setSliceUpdate(current.Aliases, target.Aliases, &input.ClearAliases, &input.Aliases)
	setSliceUpdate(current.MappedCategories, target.MappedCategories, &input.ClearMappedCategories, &input.MappedCategories)
	setSliceUpdate(current.AssessmentObjectives, target.AssessmentObjectives, &input.ClearAssessmentObjectives, &input.AssessmentObjectives)
	setSliceUpdate(current.AssessmentMethods, target.AssessmentMethods, &input.ClearAssessmentMethods, &input.AssessmentMethods)
	setSliceUpdate(current.ControlQuestions, target.ControlQuestions, &input.ClearControlQuestions, &input.ControlQuestions)
	setSliceUpdate(current.ImplementationGuidance, target.ImplementationGuidance, &input.ClearImplementationGuidance, &input.ImplementationGuidance)
	setSliceUpdate(current.ExampleEvidence, target.ExampleEvidence, &input.ClearExampleEvidence, &input.ExampleEvidence)
	setSliceUpdate(current.References, target.References, &input.ClearReferences, &input.References)
	setSliceUpdate(current.TestingProcedures, target.TestingProcedures, &input.ClearTestingProcedures, &input.TestingProcedures)
	setSliceUpdate(current.EvidenceRequests, target.EvidenceRequests, &input.ClearEvidenceRequests, &input.EvidenceRequests)

	return input
}

// buildSubcontrolUpdateInput creates an UpdateSubcontrolInput from the current and target subcontrol
// preserving only the fields that should be updated from the standard
func buildSubcontrolUpdateInput(current, target *generated.Subcontrol) generated.UpdateSubcontrolInput {
	input := generated.UpdateSubcontrolInput{
		Title:       &target.Title,
		Description: &target.Description,
		Source:      &target.Source,
		ControlType: &target.ControlType,
		Category:    &target.Category,
		CategoryID:  &target.CategoryID,
		Subcategory: &target.Subcategory,
	}

	setSliceUpdate(current.Tags, target.Tags, &input.ClearTags, &input.Tags)
	setSliceUpdate(current.DescriptionJSON, target.DescriptionJSON, &input.ClearDescriptionJSON, &input.DescriptionJSON)
	setSliceUpdate(current.Aliases, target.Aliases, &input.ClearAliases, &input.Aliases)
	setSliceUpdate(current.MappedCategories, target.MappedCategories, &input.ClearMappedCategories, &input.MappedCategories)
	setSliceUpdate(current.AssessmentObjectives, target.AssessmentObjectives, &input.ClearAssessmentObjectives, &input.AssessmentObjectives)
	setSliceUpdate(current.AssessmentMethods, target.AssessmentMethods, &input.ClearAssessmentMethods, &input.AssessmentMethods)
	setSliceUpdate(current.ControlQuestions, target.ControlQuestions, &input.ClearControlQuestions, &input.ControlQuestions)
	setSliceUpdate(current.ImplementationGuidance, target.ImplementationGuidance, &input.ClearImplementationGuidance, &input.ImplementationGuidance)
	setSliceUpdate(current.ExampleEvidence, target.ExampleEvidence, &input.ClearExampleEvidence, &input.ExampleEvidence)
	setSliceUpdate(current.References, target.References, &input.ClearReferences, &input.References)
	setSliceUpdate(current.TestingProcedures, target.TestingProcedures, &input.ClearTestingProcedures, &input.TestingProcedures)
	setSliceUpdate(current.EvidenceRequests, target.EvidenceRequests, &input.ClearEvidenceRequests, &input.EvidenceRequests)

	return input
}

func setSliceUpdate[T any](current, target []T, clear *bool, set *[]T) {
	if target == nil {
		if len(current) > 0 {
			*clear = true
		}
		return
	}

	*set = target
}
