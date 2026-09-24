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

// cloneMappings fetches the system owned control mappings that are connected to the provided ids,
// validates the controls exists in the org, skips any one that is already mapped and creates the rest
func (r *mutationResolver) cloneMappings(ctx context.Context, ids []string, orgID string) error {
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

	//check mappings from either the to side or from of controls or subcontrols
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

	orgControls, err := client.Control.Query().
		Where(
			control.IDIn(controlIDs...)).
		WithStandard().
		All(allowCtx)
	if err != nil {
		return err
	}

	if len(orgControls) == 0 {
		return nil
	}

	sourceKeys := lo.Associate(orgControls, func(c *generated.Control) (string, controlKey) {

		input, _ := controls.CreateCloneControlInput(c, nil, orgID)

		return c.ID, controlKey{
			refCode:    input.RefCode,
			standardID: lo.FromPtr(input.StandardID),
		}
	})

	refCodesToRetrieve := lo.Map(orgControls, func(c *generated.Control, _ int) string {
		return c.RefCode
	})

	controlsToMap, err := client.Control.Query().
		Select(control.FieldID, control.FieldRefCode, control.FieldStandardID).
		Where(
			control.SystemOwned(false),
			control.RefCodeIn(refCodesToRetrieve...)).
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

	controlLookup, subcontrolLookup := mapControls(sourceKeys, sourceSubcontrols, controlsToMap)

	type cloneTarget struct {
		template *generated.MappedControl
		controls controlMappings
	}

	itemsToClone := lo.FilterMap(mappings, func(m *generated.MappedControl, _ int) (cloneTarget, bool) {
		mappedControls, hasBothSides := buildMappings(m, controlLookup, subcontrolLookup)
		return cloneTarget{template: m, controls: mappedControls}, hasBothSides
	})

	if len(itemsToClone) == 0 {
		return nil
	}

	destinationIDs := lo.Map(controlsToMap, func(c *generated.Control, _ int) string {
		return c.ID
	})

	existingMappedControls, err := client.MappedControl.Query().Where(
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

	mapped := lo.FilterMap(itemsToClone, func(mapping cloneTarget, _ int) (*generated.CreateMappedControlInput, bool) {
		m, controls := mapping.template, mapping.controls

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

type controlKey struct {
	refCode    string
	standardID string
}

func mapControls(keys map[string]controlKey, subcontrols map[string]*generated.Subcontrol, controls []*generated.Control) (map[string][]string, map[string][]string) {

	type subcontrolKey struct {
		controlID string
		refCode   string
	}

	controlsByKey := map[controlKey][]string{}
	subcontrolsByKey := map[subcontrolKey][]string{}

	for _, c := range controls {
		key := controlKey{c.RefCode, c.StandardID}
		controlsByKey[key] = append(controlsByKey[key], c.ID)

		for _, sc := range c.Edges.Subcontrols {
			key := subcontrolKey{c.ID, sc.RefCode}
			subcontrolsByKey[key] = append(subcontrolsByKey[key], sc.ID)
		}
	}

	controlLookup := lo.MapValues(keys, func(key controlKey, _ string) []string {
		return controlsByKey[key]
	})

	subcontrolLookup := lo.MapValues(subcontrols, func(sc *generated.Subcontrol, _ string) []string {
		var ids []string
		for _, controlID := range controlLookup[sc.ControlID] {
			key := subcontrolKey{controlID, sc.RefCode}
			ids = append(ids, subcontrolsByKey[key]...)
		}
		return ids
	})

	return controlLookup, subcontrolLookup
}

func buildMappings(m *generated.MappedControl, controlLookup, subcontrolLookup map[string][]string) (controlMappings, bool) {
	mappings := controlMappings{}

	fn := func(ids []string, id string, lookup map[string][]string) []string {
		if lookup == nil {
			return append(ids, id)
		}

		return append(ids, lookup[id]...)
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
