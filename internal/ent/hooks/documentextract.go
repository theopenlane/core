package hooks

import (
	"context"
	"strings"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/subcontrol"
	"github.com/theopenlane/core/pkg/logx"
)

type addControlsMutation interface {
	AddControlIDs(ids ...string)
}

type addSubcontrolsMutation interface {
	AddSubcontrolIDs(ids ...string)
}

// HookParseAssociations is an ent hook that parses associations from a document
// such as referenced controls and adds the necessary edges
func HookParseAssociations() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mut := m.(detailsMutation)

			details, ok := mut.Details()
			if !ok || details == "" {
				return next.Mutate(ctx, m)
			}

			edgeLinks := getDocumentAssociations(ctx, mut)

			if edgeLinks == nil {
				return next.Mutate(ctx, m)
			}

			if len(edgeLinks.controlIDs) > 0 {
				conMut, ok := mut.(addControlsMutation)
				if ok {
					conMut.AddControlIDs(edgeLinks.controlIDs...)
				}

				subconMut, ok := mut.(addSubcontrolsMutation)
				if ok {
					subconMut.AddSubcontrolIDs(edgeLinks.controlIDs...)
				}
			}

			return next.Mutate(ctx, m)
		})
	}, hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne))
}

// edgeLinks is a struct that holds the IDs of associated entities that should be linked to the document being created or updated, such as controlIDs
type edgeLinks struct {
	controlIDs []string
}

// getDocumentAssociations will read text details and try to extract any associations to other entities in the system, such as referenced control IDs, asset IDs, and identity holder IDs. This is a placeholder implementation and should be replaced with actual parsing logic based on the expected format of the details text.
func getDocumentAssociations(ctx context.Context, m detailsMutation) *edgeLinks {
	// example text SO/IEC 27001:2022 Clause 4.2; A.5.31; A.5.32; A.5.36; SOC 2 CC2.2
	orgControls := getOrganizationControls(ctx, m)

	if orgControls == nil {
		return nil
	}

	edges := &edgeLinks{}

	details, ok := m.Details()
	if ok || details != "" {
		for refCode, info := range orgControls {
			if strings.Contains(strings.ToLower(details), refCode) {
				edges.controlIDs = append(edges.controlIDs, info.ID)
			}
		}
	}

	return edges
}

// controlInfo is a struct that holds relevant information about a control that may be referenced in the document details, such as its ID, reference code, and framework
type controlInfo struct {
	ID           string
	RefCode      string
	Framework    string
	IsSubControl bool
}

// controlMapping contains the refCode with additional info
type controlMapping map[string]controlInfo

func getOrganizationControls(ctx context.Context, m detailsMutation) controlMapping {
	result := controlMapping{}
	controls, err := m.Client().Control.Query().Where(
		control.IsTrustCenterControl(false),
		control.SystemOwned(false),
	).All(ctx)
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("failed to query controls for association parsing, no control associations will be parsed from the document details")

		// do not return the error as we want to continue processing the document even if we fail to query controls, just without any control associations parsed
		return nil
	}

	for _, c := range controls {
		result[c.RefCode] = controlInfo{
			ID:           c.ID,
			RefCode:      c.RefCode,
			Framework:    *c.ReferenceFramework,
			IsSubControl: false,
		}
	}

	subcontrols, err := m.Client().Subcontrol.Query().Where(
		subcontrol.SystemOwned(false),
	).All(ctx)
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("failed to query controls for association parsing, no control associations will be parsed from the document details")

		// do not return the error as we want to continue processing the document even if we fail to query controls, just without any control associations parsed
		return nil
	}

	for _, c := range subcontrols {
		result[c.RefCode] = controlInfo{
			ID:           c.ID,
			RefCode:      c.RefCode,
			Framework:    *c.ReferenceFramework,
			IsSubControl: true,
		}
	}

	return result
}
