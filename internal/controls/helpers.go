package controls

import (
	"context"
	"strings"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
	"github.com/theopenlane/core/v2/internal/ent/generated/standard"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"
)

// CloneFilterOptions holds the filter options for cloning controls
type CloneFilterOptions struct {
	// StandardID of the standard to filter the controls by
	StandardID *string
	// StandardShortName is the short name of the standard to filter the controls by
	StandardShortName *string
	// StandardFrameworkName is the name of the framework the standard belongs to, used in conjunction with StandardShortName to filter controls by standard when cloning
	StandardFrameworkName *string
	// StandardVersion is the revision of the standard to filter the controls by, this is optional, if more than one revision exists and this is not provided, the most recent revision will be used
	StandardVersion *string
	// RefCodes is a list of RefCodes to filter the controls by
	RefCodes []string
	// Categories is a list of Categories to filter the controls by
	Categories []string
}

// StandardFilter returns the predicate to filter by standard based on the filter options
func StandardFilter(opts CloneFilterOptions) []predicate.Standard {
	// safety check to make sure at least the ID or shortName is set
	if !FilterByStandard(opts) {
		return nil
	}

	stdWhereFilter := []predicate.Standard{}

	if opts.StandardID != nil {
		stdWhereFilter = append(stdWhereFilter, standard.ID(*opts.StandardID))
	} else {
		stdWhereFilter = append(stdWhereFilter, standard.ShortNameEqualFold(*opts.StandardShortName))
		if opts.StandardFrameworkName != nil {
			stdWhereFilter = append(stdWhereFilter, standard.FrameworkEqualFold(*opts.StandardFrameworkName))
		}

		if opts.StandardVersion != nil {
			stdWhereFilter = append(stdWhereFilter, standard.VersionEqualFold(*opts.StandardVersion))
		}
	}

	return stdWhereFilter
}

// controlRefCodeFilter returns the predicate to filter by control refCode or alias based on the filter options
func controlRefCodeFilter(opts CloneFilterOptions) []predicate.Control {
	if len(opts.RefCodes) > 0 {
		refOpts := []predicate.Control{control.RefCodeIn(opts.RefCodes...)}
		for _, refCode := range opts.RefCodes {
			refOpts = append(refOpts, func(s *sql.Selector) {
				s.Where(sqljson.ValueContains(control.FieldAliases, refCode))
			})
		}

		return []predicate.Control{
			control.Or(
				refOpts...,
			),
		}
	}

	return nil
}

// ControlFilterByStandard returns the predicate to filter controls by standard and other filter options
// to return the correct controls to clone
func ControlFilterByStandard(ctx context.Context, opts CloneFilterOptions, std *generated.Standard) ([]predicate.Control, error) {
	if std == nil {
		return nil, ErrStandardNotFound
	}

	where := []predicate.Control{
		control.DeletedAtIsNil(),
		control.StandardID(std.ID),
	}

	if std.IsPublic {
		where = append(where, control.SystemOwned(true))
	} else {
		orgID, err := auth.GetOrganizationIDFromContext(ctx)
		if err != nil {
			return nil, rout.NewMissingRequiredFieldError("owner_id")
		}

		where = append(where, control.OwnerID(orgID))
	}

	if len(opts.Categories) > 0 {
		where = append(where, control.CategoryIn(opts.Categories...))
	}

	refCodeFilter := controlRefCodeFilter(opts)
	if refCodeFilter != nil {
		where = append(where, refCodeFilter...)
	}

	return where, nil
}

// FilterByStandard returns true if the filter options contain a standardID or standardShortName
func FilterByStandard(opts CloneFilterOptions) bool {
	return opts.StandardID != nil || opts.StandardShortName != nil
}

// GetControlIDFromRefCode searches for a control ID by ref code or alias in the provided controls
// returns the control ID and a boolean indicating if it is a subcontrol or not
func GetControlIDFromRefCode(refCode string, controls []*generated.Control) (*string, bool) {
	for _, c := range controls {
		if c.RefCode == refCode {
			return &c.ID, false
		}
	}

	for _, c := range controls {
		for _, alias := range c.Aliases {
			if alias == refCode {
				return &c.ID, false
			}
		}
	}

	for _, c := range controls {
		sc := c.Edges.Subcontrols
		if sc == nil {
			continue
		}

		for _, s := range sc {
			if s.RefCode == refCode {
				return &c.ID, true
			}

			for _, alias := range s.Aliases {
				if alias == refCode {
					return &c.ID, true
				}
			}
		}
	}

	return nil, false
}

// CleanImplementationGuidance cleans and formats the implementation guidance input
// by splitting on new lines and trimming spaces
func CleanImplementationGuidance(input *string) *models.ImplementationGuidance {
	if input == nil || *input == "" {
		return nil
	}

	parts := strings.Split(*input, "\n")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)

	}

	guide := models.ImplementationGuidance{
		Guidance: parts,
	}

	return &guide
}
