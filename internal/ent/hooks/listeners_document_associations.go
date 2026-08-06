package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	// docDetailsField is the rich-text field scanned for control references on association-eligible documents
	docDetailsField = "details"
	// docRevisionField is the revision field re-asserted on link updates so associations do not bump revisions
	docRevisionField = "revision"
)

// RegisterGalaDocumentAssociationListeners registers listeners that link
// referenced controls to documents asynchronously after document creation.
func RegisterGalaDocumentAssociationListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	return eventqueue.RegisterMutationListeners(g,
		documentAssociationListener(generated.TypeActionPlan),
		documentAssociationListener(generated.TypeInternalPolicy),
		documentAssociationListener(generated.TypeProcedure),
	)
}

// documentAssociationListener builds the document association listener for one schema type
func documentAssociationListener(schemaType string) eventqueue.MutationListener {
	return eventqueue.MutationListener{
		Schema:     schemaType,
		Name:       "document.associations." + schemaType,
		Operations: []string{ent.OpCreate.String()},
		Handle:     handleDocumentAssociationCreated,
	}
}

// handleDocumentAssociationCreated links controls referenced in a new document's details
// through the schema catalog: the row is loaded generically, matched control references
// become add-edge keys, and a single catalog update re-asserts the loaded revision so the
// links do not bump the document's revision
func handleDocumentAssociationCreated(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
	schema, ok := entityops.LookupSchema(payload.MutationType)
	if !ok {
		return nil
	}

	row, err := schema.Load(inv.Context, inv.Client, inv.EntityID)
	switch {
	case generated.IsNotFound(err):
		return nil
	case err != nil:
		return err
	}

	fields, err := jsonx.Decode[map[string]any](row)
	if err != nil {
		return err
	}

	details, _ := fields[docDetailsField].(string)

	links := getDocumentAssociationsForDetails(inv.Context, inv.Client, details)
	if !links.hasAssociations() {
		return nil
	}

	update := map[string]any{docRevisionField: fields[docRevisionField]}

	for edgeName, ids := range map[string][]string{
		"controls":    links.controlIDs,
		"subcontrols": links.subcontrolIDs,
	} {
		edge, ok := schema.EdgeByName(edgeName)
		if !ok || edge.AddField == "" || len(ids) == 0 {
			continue
		}

		update[edge.AddField] = lo.Uniq(ids)
	}

	if len(update) == 1 {
		return nil
	}

	updatePayload, err := jsonx.ToRawMessage(update)
	if err != nil {
		return err
	}

	return schema.Update(workflows.AllowContext(inv.Context), inv.Client, inv.EntityID, updatePayload)
}

func getDocumentAssociationsForDetails(ctx context.Context, client *generated.Client, details string) *edgeLinks {
	if details == "" {
		return nil
	}

	orgControls := getOrganizationControlsFromClient(ctx, client)
	if orgControls == nil {
		return nil
	}

	return findControlMatches(details, orgControls)
}

func (e *edgeLinks) hasAssociations() bool {
	return e != nil && (len(e.controlIDs) > 0 || len(e.subcontrolIDs) > 0)
}
