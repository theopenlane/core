package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/gala"
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

func handleDocumentAssociationCreated(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
	switch payload.MutationType {
	case generated.TypeActionPlan:
		return parseActionPlanAssociations(inv.Context, inv.Client, inv.EntityID)
	case generated.TypeInternalPolicy:
		return parseInternalPolicyAssociations(inv.Context, inv.Client, inv.EntityID)
	case generated.TypeProcedure:
		return parseProcedureAssociations(inv.Context, inv.Client, inv.EntityID)
	default:
		return nil
	}
}

func parseActionPlanAssociations(ctx context.Context, client *generated.Client, documentID string) error {
	doc, err := client.ActionPlan.Get(ctx, documentID)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil
		}

		return err
	}

	links := getDocumentAssociationsForDetails(ctx, client, doc.Details)
	if links == nil || len(links.controlIDs) == 0 {
		return nil
	}

	return client.ActionPlan.UpdateOneID(doc.ID).
		SetRevision(doc.Revision).
		AddControlIDs(lo.Uniq(links.controlIDs)...).
		Exec(workflows.AllowContext(ctx))
}

func parseInternalPolicyAssociations(ctx context.Context, client *generated.Client, documentID string) error {
	doc, err := client.InternalPolicy.Get(ctx, documentID)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil
		}

		return err
	}

	links := getDocumentAssociationsForDetails(ctx, client, doc.Details)
	if links == nil || !links.hasAssociations() {
		return nil
	}

	update := client.InternalPolicy.UpdateOneID(doc.ID).SetRevision(doc.Revision)
	if len(links.controlIDs) > 0 {
		update.AddControlIDs(lo.Uniq(links.controlIDs)...)
	}

	if len(links.subcontrolIDs) > 0 {
		update.AddSubcontrolIDs(lo.Uniq(links.subcontrolIDs)...)
	}

	return update.Exec(workflows.AllowContext(ctx))
}

func parseProcedureAssociations(ctx context.Context, client *generated.Client, documentID string) error {
	doc, err := client.Procedure.Get(ctx, documentID)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil
		}

		return err
	}

	links := getDocumentAssociationsForDetails(ctx, client, doc.Details)
	if links == nil || !links.hasAssociations() {
		return nil
	}

	update := client.Procedure.UpdateOneID(doc.ID).SetRevision(doc.Revision)
	if len(links.controlIDs) > 0 {
		update.AddControlIDs(lo.Uniq(links.controlIDs)...)
	}

	if len(links.subcontrolIDs) > 0 {
		update.AddSubcontrolIDs(lo.Uniq(links.subcontrolIDs)...)
	}

	return update.Exec(workflows.AllowContext(ctx))
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
