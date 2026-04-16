package hooks

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/logx"
)

// versionMutation allows for setting the revision field in a mutation
type versionMutation interface {
	SetRevision(revision string)
}

// HookDetailsVersion is an ent hook that parses the versions from the details of a document
// creation
func HookDetailsVersion() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mut := m.(detailsMutation)

			details, ok := mut.Details()
			if !ok || details == "" {
				return next.Mutate(ctx, m)
			}

			if version := findVersion(details); version != "" {
				verMut, ok := mut.(versionMutation)
				if ok {
					verMut.SetRevision(version)
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

// findVersion attempts to find the version of the document in the details by
// looking for a line that starts with "Version:" and extracting the version number from it. It returns the version in semver format if found, or an empty string if not found.
func findVersion(details string) string {
	lines := strings.Split(details, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(trimmed), "version:") {
			parts := strings.SplitN(trimmed, ":", 2) //nolint:mnd
			if len(parts) == 2 {                     //nolint:mnd
				// convert to semver format if possible, e.g. "Version: 1.0" -> "1.0.0"
				version := strings.TrimSpace(parts[1])
				if version == "" {
					return ""
				}

				parts := strings.Split(version, ".")
				for len(parts) < 3 {
					parts = append(parts, "0")
				}
				// Validate all parts are numbers
				for _, p := range parts[:3] {
					if _, err := strconv.Atoi(p); err != nil {
						return ""
					}
				}

				version = strings.Join(parts, ".")

				if !strings.HasPrefix(version, "v") {
					version = "v" + version
				}
				return version
			}
		}
	}

	return ""
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
				if info.isSubControl {
					edges.subcontrolIDs = append(edges.subcontrolIDs, info.id)
				} else {
					edges.controlIDs = append(edges.controlIDs, info.id)
				}
			}
		}
	}
	return edges
}

// controlInfo is a struct that holds relevant information about a control that may be referenced in the document details, such as its ID, reference code, and framework
type controlInfo struct {
	id           string
	fefCode      string
	framework    string
	isSubControl bool
}

// controlMapping contains the refCode with additional info
type controlMapping map[string]controlInfo

func getOrganizationControlsFromClient(ctx context.Context, client *generated.Client) controlMapping {
	result := controlMapping{}
	controls, err := client.Control.Query().Where(
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
			id:           c.ID,
			fefCode:      c.RefCode,
			framework:    refFramework,
			isSubControl: false,
		}

		for _, sc := range c.Edges.Subcontrols {
			subRefFramework := ""
			if sc.ReferenceFramework != nil {
				subRefFramework = *sc.ReferenceFramework
			}

			result[sc.RefCode] = controlInfo{
				id:           sc.ID,
				fefCode:      sc.RefCode,
				framework:    subRefFramework,
				isSubControl: true,
			}
		}

	}

	return result
}
