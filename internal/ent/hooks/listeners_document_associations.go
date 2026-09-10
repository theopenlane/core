package hooks

import (
	"context"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// DocumentAssociationListeners links controls referenced in a new document's details
func DocumentAssociationListeners() []gala.Registration {
	return entityops.ForSchemas([]*entityops.Schema{entityops.SchemaActionPlan, entityops.SchemaInternalPolicy, entityops.SchemaProcedure}, entityops.MutationListener{
		Operations: []string{entityops.OpCreate},
		Handle:     handleDocumentAssociationCreated,
	})
}

// handleDocumentAssociationCreated loads the created document generically and links
// referenced controls and subcontrols through the schema catalog
func handleDocumentAssociationCreated(inv entityops.Invocation, _ entityops.MutationPayload) error {
	row, err := inv.Schema.Load(inv.Context, inv.Client, inv.EntityID)
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

	details, _ := fields["details"].(string)

	links := getDocumentAssociationsForDetails(inv.Context, inv.Client, details)
	if !links.hasAssociations() {
		return nil
	}

	update := map[string]any{"revision": fields["revision"]}

	for edgeName, ids := range map[string][]string{
		"controls":    links.controlIDs,
		"subcontrols": links.subcontrolIDs,
	} {
		edge, ok := inv.Schema.EdgeByName(edgeName)
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

	return inv.Schema.Update(rule.WithInternalContext(inv.Context), inv.Client, inv.EntityID, updatePayload)
}

// getDocumentAssociationsForDetails returns the control and subcontrol IDs referenced in a document's details
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
