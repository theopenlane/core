package hooks

import (
	"context"
	"regexp"
	"strings"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/logx"
)

// addControlsMutation allows for adding to the controls edge in a mutation
type addControlsMutation interface {
	AddControlIDs(ids ...string)
}

// addSubcontrolsMutation allows for adding to the subcontrols edge in a mutation
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
			}

			if len(edgeLinks.subcontrolIDs) > 0 {
				subconMut, ok := mut.(addSubcontrolsMutation)
				if ok {
					subconMut.AddSubcontrolIDs(edgeLinks.subcontrolIDs...)
				}
			}

			return next.Mutate(ctx, m)
		})
		// only do on create for now to avoid updating associations that a user may have manually removed
	}, hook.HasOp(ent.OpCreate))
}

// edgeLinks is a struct that holds the IDs of associated entities that should be linked to the document being created or updated, such as controlIDs
type edgeLinks struct {
	controlIDs    []string
	subcontrolIDs []string
}

// getDocumentAssociations will read text details and try to extract any associations to other entities in the system, such as referenced control IDs, asset IDs, and identity holder IDs. This is a placeholder implementation and should be replaced with actual parsing logic based on the expected format of the details text.
func getDocumentAssociations(ctx context.Context, m detailsMutation) *edgeLinks {
	orgControls := getOrganizationControls(ctx, m)

	if orgControls == nil {
		return nil
	}

	details, ok := m.Details()
	if !ok || details == "" {
		return nil
	}

	return findControlMatches(details, orgControls)
}

// findControlMatches searches the text for matches to a control, this is a simple implementation that will be
// expanded upon in the future to handle more complex matching based on more context in the document such as presence of a framework
func findControlMatches(details string, controls controlMapping) *edgeLinks {
	edges := &edgeLinks{}
	lines := strings.Split(details, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip markdown headers
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		lineLower := strings.ToLower(trimmed)
		for refCode, info := range controls {
			pattern := `(?i)(^|[\s,;:\(\)\[\]\{\}])` + regexp.QuoteMeta(refCode) + `([\s,;:\(\)\[\]\{\}\.]|$)`
			re := regexp.MustCompile(pattern)
			if re.MatchString(lineLower) {
				if info.IsSubControl {
					edges.subcontrolIDs = append(edges.subcontrolIDs, info.ID)
				} else {
					edges.controlIDs = append(edges.controlIDs, info.ID)
				}
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
	).WithSubcontrols().All(ctx)
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("failed to query controls for association parsing, no control associations will be parsed from the document details")

		// do not return the error as we want to continue processing the document even if we fail to query controls, just without any control associations parsed
		return nil
	}

	for _, c := range controls {
		refFramework := ""
		if c.ReferenceFramework != nil {
			refFramework = *c.ReferenceFramework
		}

		result[c.RefCode] = controlInfo{
			ID:           c.ID,
			RefCode:      c.RefCode,
			Framework:    refFramework,
			IsSubControl: false,
		}

		for _, sc := range c.Edges.Subcontrols {
			subRefFramework := ""
			if sc.ReferenceFramework != nil {
				subRefFramework = *sc.ReferenceFramework
			}

			result[sc.RefCode] = controlInfo{
				ID:           sc.ID,
				RefCode:      sc.RefCode,
				Framework:    subRefFramework,
				IsSubControl: true,
			}
		}

	}

	return result
}
