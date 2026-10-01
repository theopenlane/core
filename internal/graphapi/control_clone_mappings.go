package graphapi

import (
	"context"
	"slices"
	"strings"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/controls"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/ent/generated/mappedcontrol"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/standard"
	"github.com/theopenlane/core/v2/internal/ent/generated/subcontrol"
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

// cloneTemplateMappings replaces template IDs with organization IDs in their mappings.
// Framework controls keep their system IDs; reports resolve them to organization copies.
func (r *mutationResolver) cloneTemplateMappings(ctx context.Context, ids []string, orgID string, program *generated.Program) error {
	if len(ids) == 0 {
		return nil
	}

	client := withTransactionalMutation(ctx)

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	controlIDsOnlyPredicate := func(q *generated.ControlQuery) {
		q.Select(control.FieldID)
	}

	subcontrolReferencesPredicate := func(q *generated.SubcontrolQuery) {
		q.Select(subcontrol.FieldID, subcontrol.FieldControlID, subcontrol.FieldRefCode)
	}

	// check mappings from either the to side or from of controls or subcontrols
	mappings, err := client.MappedControl.Query().Where(
		mappedcontrol.SystemOwned(true),
		mappedcontrol.Or(
			mappedcontrol.HasFromControlsWith(control.IDIn(ids...)),
			mappedcontrol.HasToControlsWith(control.IDIn(ids...)),
			mappedcontrol.HasFromSubcontrolsWith(subcontrol.ControlIDIn(ids...)),
			mappedcontrol.HasToSubcontrolsWith(subcontrol.ControlIDIn(ids...)),
		),
	).
		WithFromControls(controlIDsOnlyPredicate).
		WithToControls(controlIDsOnlyPredicate).
		WithFromSubcontrols(subcontrolReferencesPredicate).
		WithToSubcontrols(subcontrolReferencesPredicate).
		All(allowCtx)
	if err != nil {
		return err
	}

	if len(mappings) == 0 {
		return nil
	}

	var controlIDs []string

	sourceSubcontrols := map[string]*generated.Subcontrol{}

	for _, m := range mappings {
		for _, c := range slices.Concat(m.Edges.FromControls, m.Edges.ToControls) {
			controlIDs = append(controlIDs, c.ID)
		}

		for _, sc := range slices.Concat(m.Edges.FromSubcontrols, m.Edges.ToSubcontrols) {
			controlIDs = append(controlIDs, sc.ControlID)
			sourceSubcontrols[sc.ID] = sc
		}
	}

	controlIDs = lo.Uniq(controlIDs)

	if len(controlIDs) == 0 {
		return nil
	}

	// fetch controls that match the provided ids but if program is provided,
	// make sure to include it as a filter too
	controlsQuery := client.Control.Query().Where(control.IDIn(controlIDs...))

	if program != nil {
		controlsIDsPredicate := control.IDIn(ids...)

		if program.FrameworkName == "" {

			controlsQuery.Where(controlsIDsPredicate)
		} else {

			controlsQuery.Where(
				control.Or(
					controlsIDsPredicate,
					control.ReferenceFramework(program.FrameworkName)),
			)
		}
	}

	matchedControls, err := controlsQuery.
		Select(control.FieldID, control.FieldRefCode, control.FieldStandardID).
		WithStandard(func(q *generated.StandardQuery) {
			q.Select(standard.FieldID, standard.FieldSystemOwned)
		}).All(allowCtx)
	if err != nil {
		return err
	}

	// filter out to match known system templates
	templateControls := lo.Filter(matchedControls, func(c *generated.Control, _ int) bool {
		return controls.IsOpenlaneBaseControl(c)
	})

	if len(templateControls) == 0 {
		return nil
	}

	sourceRefCodes := lo.Associate(templateControls, func(c *generated.Control) (string, string) {
		return c.ID, c.RefCode
	})

	refCodesToRetrieve := lo.Map(templateControls, func(c *generated.Control, _ int) string {
		return c.RefCode
	})

	controlsToMap, err := client.Control.Query().
		Select(control.FieldID, control.FieldRefCode).
		Where(
			control.StandardIDIsNil(),
			control.DeletedAtIsNil(),
			control.SystemOwned(false),
			control.RefCodeIn(refCodesToRetrieve...),
		).
		WithSubcontrols(func(q *generated.SubcontrolQuery) {
			q.Where(
				subcontrol.OwnerID(orgID),
				subcontrol.SystemOwned(false),
			).
				Select(subcontrol.FieldID, subcontrol.FieldControlID, subcontrol.FieldRefCode)
		}).
		All(allowCtx)
	if err != nil {
		return err
	}

	if len(controlsToMap) == 0 {
		return nil
	}

	controlLookup, subcontrolLookup := mapControls(sourceRefCodes, sourceSubcontrols, controlsToMap)

	for _, c := range matchedControls {
		if _, ok := sourceRefCodes[c.ID]; !ok {
			controlLookup[c.ID] = c.ID
		}
	}

	for id, sc := range sourceSubcontrols {
		if _, ok := sourceRefCodes[sc.ControlID]; !ok && controlLookup[sc.ControlID] != "" {
			subcontrolLookup[id] = id
		}
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
		All(allowCtx)
	if err != nil {
		return err
	}

	existingMappings := lo.Associate(existingMappedControls, func(m *generated.MappedControl) (dedupSetKey, struct{}) {
		mappingControls, _ := buildMappings(m, nil, nil)
		return mappingControls.uniqueKey(), struct{}{}
	})

	mapped := lo.FilterMap(mappings, func(m *generated.MappedControl, _ int) (*generated.CreateMappedControlInput, bool) {

		controls, hasBothSides := buildMappings(m, controlLookup, subcontrolLookup)
		if !hasBothSides {
			return nil, false
		}

		combinedControls := slices.Concat(m.Edges.FromControls, m.Edges.ToControls)

		// check the from and to side for a template control that has an already existing copy in the org
		controlExists := lo.SomeBy(combinedControls, func(c *generated.Control) bool {
			_, ok := sourceRefCodes[c.ID]
			return ok && controlLookup[c.ID] != ""
		})

		combinedSubcontrols := slices.Concat(m.Edges.FromSubcontrols, m.Edges.ToSubcontrols)

		// subcontrols can be mapped but parent control must exists ( which is already the template one)
		subcontrolsExists := lo.SomeBy(combinedSubcontrols, func(sc *generated.Subcontrol) bool {
			_, ok := sourceRefCodes[sc.ControlID]
			return ok && subcontrolLookup[sc.ID] != ""
		})

		if !controlExists && !subcontrolsExists {
			return nil, false
		}

		// if an existing mapping already exists, skip it
		key := controls.uniqueKey()
		if _, ok := existingMappings[key]; ok {
			return nil, false
		}

		existingMappings[key] = struct{}{}

		return &generated.CreateMappedControlInput{
			OwnerID:           lo.ToPtr(orgID),
			Source:            lo.ToPtr(enums.MappingSourceImported),
			MappingType:       &m.MappingType,
			Relation:          &m.Relation,
			Confidence:        m.Confidence,
			FromControlIDs:    controls.fromControls,
			ToControlIDs:      controls.toControls,
			FromSubcontrolIDs: controls.fromSubcontrols,
			ToSubcontrolIDs:   controls.toSubcontrols,
		}, true
	})

	if len(mapped) == 0 {
		return nil
	}

	_, err = r.bulkCreateMappedControl(allowCtx, mapped)
	return err
}

// mapControls tries to match the templates to the already retrieved org control copies
func mapControls(refCodes map[string]string, subcontrols map[string]*generated.Subcontrol, controls []*generated.Control) (map[string]string, map[string]string) {
	type subcontrolKey struct {
		controlID string
		refCode   string
	}

	controlsByRefCode := map[string]string{}
	subcontrolsByKey := map[subcontrolKey]string{}
	for _, c := range controls {
		controlsByRefCode[c.RefCode] = c.ID
		for _, sc := range c.Edges.Subcontrols {
			subcontrolsByKey[subcontrolKey{c.ID, sc.RefCode}] = sc.ID
		}
	}

	controlLookup := lo.MapValues(refCodes, func(refCode string, _ string) string {
		return controlsByRefCode[refCode]
	})

	subcontrolLookup := lo.MapValues(subcontrols, func(sc *generated.Subcontrol, _ string) string {
		return subcontrolsByKey[subcontrolKey{controlLookup[sc.ControlID], sc.RefCode}]
	})

	return controlLookup, subcontrolLookup
}

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
