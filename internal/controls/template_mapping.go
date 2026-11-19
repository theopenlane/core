package controls

import (
	"context"
	"slices"
	"strings"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/ent/generated/mappedcontrol"
	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
	programquery "github.com/theopenlane/core/v2/internal/ent/generated/program"
	"github.com/theopenlane/core/v2/internal/ent/generated/standard"
	"github.com/theopenlane/core/v2/internal/ent/generated/subcontrol"
)

const (
	// batch writes to prevent large allocation for the batch builders of mappings.
	// We could end up holding a large slice of pointers, so instead flush the writes
	// every time we get to 100
	mappingWriteBatchSize = 100
)

type controlMappings struct {
	fromControls    []string
	toControls      []string
	fromSubcontrols []string
	toSubcontrols   []string
}

type dedupSetKey struct {
	fromControls    string
	toControls      string
	fromSubcontrols string
	toSubcontrols   string
}

// uniqueKey maps the existing mappings from the db to make sure subsequent imports do
// not recreate them again
func (c controlMappings) uniqueKey() dedupSetKey {
	fn := func(ids []string) string {
		slices.Sort(ids)
		return strings.Join(ids, ",")
	}

	return dedupSetKey{
		fromControls:    fn(c.fromControls),
		toControls:      fn(c.toControls),
		fromSubcontrols: fn(c.fromSubcontrols),
		toSubcontrols:   fn(c.toSubcontrols),
	}
}

// CloneTemplateMappings imports template mappings between organization copies.
func CloneTemplateMappings(ctx context.Context, ids []string, program *generated.Program) error {
	if len(ids) == 0 {
		return nil
	}

	mappings, err := fetchMappings(ctx, ids)
	if err != nil {
		return err
	}

	if len(mappings) == 0 {
		return nil
	}

	controlIDs, sourceSubcontrols := collateReferencesFromMappings(mappings)

	if len(controlIDs) == 0 {
		return nil
	}

	sourceControls, err := loadMappingControls(ctx, controlIDs)
	if err != nil {
		return err
	}

	requestedControlsToClone := lo.Associate(ids, func(id string) (string, bool) {
		return id, true
	})

	templateControls := lo.Filter(sourceControls, func(c *generated.Control, _ int) bool {
		return IsOpenlaneBaseControl(c) && (program == nil || requestedControlsToClone[c.ID])
	})

	if len(templateControls) == 0 {
		return nil
	}

	sourceRefCodes := lo.Associate(templateControls, func(c *generated.Control) (string, string) {
		return c.ID, c.RefCode
	})

	controlsToMap, err := fetchTemplateControls(ctx, sourceControls, templateControls, program)
	if err != nil {
		return err
	}

	templateCopies := lo.Filter(controlsToMap, func(c *generated.Control, _ int) bool {
		return c.StandardID == ""
	})

	if len(templateCopies) == 0 {
		return nil
	}

	controlLookup, subcontrolLookup := mapControls(sourceControls, sourceRefCodes, sourceSubcontrols, controlsToMap)

	existingMappings, err := computeExistingMappings(ctx, templateCopies)
	if err != nil {
		return err
	}

	return createMappings(ctx, mappings, sourceRefCodes, controlLookup, subcontrolLookup, existingMappings)
}

// fetchMappings retrieves the system owned mappings that are mapped to the provided control ids
func fetchMappings(ctx context.Context, ids []string) ([]*generated.MappedControl, error) {
	client := generated.FromContext(ctx)

	controlIDsOnlyPredicate := func(q *generated.ControlQuery) {
		q.Select(control.FieldID)
	}

	subcontrolReferencesPredicate := func(q *generated.SubcontrolQuery) {
		q.Select(subcontrol.FieldID, subcontrol.FieldControlID, subcontrol.FieldRefCode)
	}

	// check mappings from either the to side or from of controls or subcontrols
	return client.MappedControl.Query().Where(
		mappedcontrol.SystemOwned(true),
		mappedcontrol.Or(
			mappedcontrol.HasFromControlsWith(control.IDIn(ids...)),
			mappedcontrol.HasToControlsWith(control.IDIn(ids...)),
			mappedcontrol.HasFromSubcontrolsWith(subcontrol.ControlIDIn(ids...)),
			mappedcontrol.HasToSubcontrolsWith(subcontrol.ControlIDIn(ids...)),
		),
	).
		Select(
			mappedcontrol.FieldID, mappedcontrol.FieldMappingType,
			mappedcontrol.FieldRelation, mappedcontrol.FieldConfidence,
		).
		WithFromControls(controlIDsOnlyPredicate).
		WithToControls(controlIDsOnlyPredicate).
		WithFromSubcontrols(subcontrolReferencesPredicate).
		WithToSubcontrols(subcontrolReferencesPredicate).
		All(ctx)
}

func collateReferencesFromMappings(mappings []*generated.MappedControl) ([]string, map[string]*generated.Subcontrol) {
	var ids []string

	sourceSubcontrols := map[string]*generated.Subcontrol{}

	for _, m := range mappings {
		// no need to allocate with appends, concat since mapping maybe large
		for _, edge := range [][]*generated.Control{m.Edges.FromControls, m.Edges.ToControls} {
			for _, c := range edge {
				ids = append(ids, c.ID)
			}
		}

		for _, edge := range [][]*generated.Subcontrol{m.Edges.FromSubcontrols, m.Edges.ToSubcontrols} {
			for _, sc := range edge {
				ids = append(ids, sc.ControlID)
				sourceSubcontrols[sc.ID] = sc
			}
		}
	}

	return lo.Uniq(ids), sourceSubcontrols
}

func loadMappingControls(ctx context.Context, controlIDs []string) ([]*generated.Control, error) {
	client := generated.FromContext(ctx)

	return client.Control.Query().Where(control.IDIn(controlIDs...)).
		Select(control.FieldID, control.FieldRefCode, control.FieldStandardID).
		WithStandard(func(q *generated.StandardQuery) {
			q.Select(standard.FieldID, standard.FieldSystemOwned)
		}).All(ctx)
}

// fetchTemplateControls finds organization copies with matching template or framework reference codes
// and loads their subcontrols for the given organization.
func fetchTemplateControls(ctx context.Context, sourceControls, templateControls []*generated.Control, program *generated.Program) ([]*generated.Control, error) {
	templateRefs := lo.Map(templateControls, func(c *generated.Control, _ int) string {
		return c.RefCode
	})

	frameworkRefs := map[string][]string{}

	for _, c := range sourceControls {
		if standardID := cloneStandardID(c); standardID != nil {
			frameworkRefs[*standardID] = append(frameworkRefs[*standardID], c.RefCode)
		}
	}

	matches := []predicate.Control{control.And(control.StandardIDIsNil(), control.RefCodeIn(templateRefs...))}
	frameworkControlMatches := make([]predicate.Control, 0, len(frameworkRefs))

	for standardID, refs := range frameworkRefs {
		frameworkControlMatches = append(frameworkControlMatches, control.And(control.StandardID(standardID), control.RefCodeIn(refs...)))
	}

	frameworkControlMatch := control.Or(frameworkControlMatches...)
	if program != nil {
		frameworkControlMatch = control.And(frameworkControlMatch, control.HasProgramsWith(programquery.IDEQ(program.ID)))
	}

	if len(frameworkControlMatches) > 0 {
		matches = append(matches, frameworkControlMatch)
	}

	return generated.FromContext(ctx).Control.Query().
		Select(control.FieldID, control.FieldRefCode, control.FieldStandardID).
		Where(
			control.SystemOwned(false),
			control.Or(matches...),
		).
		WithSubcontrols(func(q *generated.SubcontrolQuery) {
			q.Where(
				subcontrol.SystemOwned(false),
			).
				Select(subcontrol.FieldID, subcontrol.FieldControlID, subcontrol.FieldRefCode)
		}).
		All(ctx)
}

func computeExistingMappings(ctx context.Context, controlsToMap []*generated.Control) (map[dedupSetKey]struct{}, error) {
	client := generated.FromContext(ctx)

	controlIDsOnlyPredicate := func(q *generated.ControlQuery) {
		q.Select(control.FieldID)
	}

	destinationIDs := lo.Map(controlsToMap, func(c *generated.Control, _ int) string {
		return c.ID
	})

	existingMappedControls, err := client.MappedControl.Query().
		Where(
			mappedcontrol.Or(
				mappedcontrol.HasFromControlsWith(control.IDIn(destinationIDs...)),
				mappedcontrol.HasToControlsWith(control.IDIn(destinationIDs...)),
				mappedcontrol.HasFromSubcontrolsWith(subcontrol.ControlIDIn(destinationIDs...)),
				mappedcontrol.HasToSubcontrolsWith(subcontrol.ControlIDIn(destinationIDs...)),
			),
		).
		Select(mappedcontrol.FieldID).
		WithFromControls(controlIDsOnlyPredicate).
		WithToControls(controlIDsOnlyPredicate).
		WithFromSubcontrols(func(q *generated.SubcontrolQuery) {
			q.Select(subcontrol.FieldID)
		}).
		WithToSubcontrols(func(q *generated.SubcontrolQuery) {
			q.Select(subcontrol.FieldID)
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return lo.Associate(existingMappedControls, func(m *generated.MappedControl) (dedupSetKey, struct{}) {
		mappingControls, _ := buildMappings(m, nil, nil)
		return mappingControls.uniqueKey(), struct{}{}
	}), nil
}

// createMappings imports mappings that retain both sides and include a matched template copy.
// It skips duplicate mappings and writes the accepted mappings in bounded batches.
func createMappings(ctx context.Context, mappings []*generated.MappedControl, sourceRefCodes, controlLookup, subcontrolLookup map[string]string, existingMappings map[dedupSetKey]struct{}) error {
	client := generated.FromContext(ctx)

	builders := make([]*generated.MappedControlCreate, 0, mappingWriteBatchSize)
	for _, m := range mappings {

		controls, hasBothSides := buildMappings(m, controlLookup, subcontrolLookup)
		if !hasBothSides {
			continue
		}

		// check the from and to side for a template control that has an already existing copy in the org
		hasControl := func(c *generated.Control) bool {
			_, ok := sourceRefCodes[c.ID]
			return ok && controlLookup[c.ID] != ""
		}
		controlExists := lo.SomeBy(m.Edges.FromControls, hasControl) || lo.SomeBy(m.Edges.ToControls, hasControl)

		// subcontrols can be mapped but parent control must exists ( which is already the template one)
		hasSubcontrol := func(sc *generated.Subcontrol) bool {
			_, ok := sourceRefCodes[sc.ControlID]
			return ok && subcontrolLookup[sc.ID] != ""
		}
		subcontrolsExists := lo.SomeBy(m.Edges.FromSubcontrols, hasSubcontrol) || lo.SomeBy(m.Edges.ToSubcontrols, hasSubcontrol)

		if !controlExists && !subcontrolsExists {
			continue
		}

		// if an existing mapping already exists, skip it
		key := controls.uniqueKey()
		if _, ok := existingMappings[key]; ok {
			continue
		}

		existingMappings[key] = struct{}{}

		input := generated.CreateMappedControlInput{
			Source:            lo.ToPtr(enums.MappingSourceImported),
			MappingType:       &m.MappingType,
			Relation:          &m.Relation,
			Confidence:        m.Confidence,
			FromControlIDs:    controls.fromControls,
			ToControlIDs:      controls.toControls,
			FromSubcontrolIDs: controls.fromSubcontrols,
			ToSubcontrolIDs:   controls.toSubcontrols,
		}
		builders = append(builders, client.MappedControl.Create().SetInput(input))

		if len(builders) == mappingWriteBatchSize {
			if err := client.MappedControl.CreateBulk(builders...).Exec(ctx); err != nil {
				return err
			}

			clear(builders)
			builders = builders[:0]
		}
	}

	if len(builders) > 0 {
		return client.MappedControl.CreateBulk(builders...).Exec(ctx)
	}

	return nil
}

// mapControls matches controls to organization copies by clone standard and reference code.
// Subcontrols match by their reference code and the matched parent control.
func mapControls(sources []*generated.Control, templateRefs map[string]string, subcontrols map[string]*generated.Subcontrol, controls []*generated.Control) (map[string]string, map[string]string) {
	type controlKey struct {
		standardID string
		refCode    string
	}

	type subcontrolKey struct {
		controlID string
		refCode   string
	}

	controlsByKey := map[controlKey]string{}
	subcontrolsByKey := map[subcontrolKey]string{}
	for _, c := range controls {
		controlsByKey[controlKey{c.StandardID, c.RefCode}] = c.ID
		for _, sc := range c.Edges.Subcontrols {
			subcontrolsByKey[subcontrolKey{c.ID, sc.RefCode}] = sc.ID
		}
	}

	controlLookup := lo.MapValues(templateRefs, func(refCode string, _ string) string {
		return controlsByKey[controlKey{refCode: refCode}]
	})

	for _, c := range sources {
		if standardID := cloneStandardID(c); standardID != nil {
			controlLookup[c.ID] = controlsByKey[controlKey{*standardID, c.RefCode}]
		}
	}

	subcontrolLookup := lo.MapValues(subcontrols, func(sc *generated.Subcontrol, _ string) string {
		return subcontrolsByKey[subcontrolKey{controlLookup[sc.ControlID], sc.RefCode}]
	})

	return controlLookup, subcontrolLookup
}

// buildMappings replaces control and subcontrol IDs using the lookups, drops unmatched IDs,
// and removes duplicates. A nil lookup preserves the original IDs.
// It also reports whether both sides still contain a control or subcontrol.
func buildMappings(m *generated.MappedControl, controlLookup, subcontrolLookup map[string]string) (controlMappings, bool) {
	mappings := controlMappings{}

	fn := func(ids []string, id string, lookup map[string]string) []string {
		if lookup == nil {
			return append(ids, id)
		}

		if mappedID := lookup[id]; mappedID != "" {
			return append(ids, mappedID)
		}
		return ids
	}

	for _, c := range m.Edges.FromControls {
		mappings.fromControls = fn(mappings.fromControls, c.ID, controlLookup)
	}

	for _, c := range m.Edges.ToControls {
		mappings.toControls = fn(mappings.toControls, c.ID, controlLookup)
	}

	for _, sc := range m.Edges.FromSubcontrols {
		mappings.fromSubcontrols = fn(mappings.fromSubcontrols, sc.ID, subcontrolLookup)
	}

	for _, sc := range m.Edges.ToSubcontrols {
		mappings.toSubcontrols = fn(mappings.toSubcontrols, sc.ID, subcontrolLookup)
	}

	mappings.fromControls = lo.Uniq(mappings.fromControls)
	mappings.toControls = lo.Uniq(mappings.toControls)
	mappings.fromSubcontrols = lo.Uniq(mappings.fromSubcontrols)
	mappings.toSubcontrols = lo.Uniq(mappings.toSubcontrols)

	existsOnBothSides := (len(mappings.fromControls)+len(mappings.fromSubcontrols)) > 0 && (len(mappings.toControls)+len(mappings.toSubcontrols)) > 0

	return mappings, existsOnBothSides
}
