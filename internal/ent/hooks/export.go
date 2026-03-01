package hooks

import (
	"context"
	"encoding/json"
	"errors"

	"entgo.io/ent"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/logx"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/iam/auth"
)

var (
	errExportTypeNotProvided = errors.New("provide export type")
	errFieldsNotProvided     = errors.New("at least one field must be provided for the schema to export")
)

func HookExport() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ExportFunc(
			func(ctx context.Context, m *generated.ExportMutation) (generated.Value, error) {
				if m.Op().Is(ent.OpCreate) {
					return handleExportCreate(ctx, m, next)
				}

				return handleExportUpdate(ctx, m, next)
			})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}

func handleExportCreate(ctx context.Context, m *generated.ExportMutation, next ent.Mutator) (generated.Value, error) {
	exportType, ok := m.ExportType()
	if !ok {
		return nil, errExportTypeNotProvided
	}

	if err := ValidateExportType(exportType.String()); err != nil {
		return nil, err
	}

	fields := m.Fields()
	if len(fields) == 0 {
		return nil, errFieldsNotProvided
	}

	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		return nil, auth.ErrNoAuthUser
	}

	orgID, _ := caller.ActiveOrg()

	if orgID == "" || caller.SubjectID == "" {
		logx.FromContext(ctx).Error().Msg("authenticated user has no organization ID or user ID; unable to enqueue export job")

		return nil, ErrNoOrganizationID
	}

	// apply ownerID/systemOwned filters to the export filters
	originalFilters, _ := m.Filters()
	finalFilters, err := getFinalFilters(originalFilters, m, orgID)
	if err != nil {
		return nil, err
	}

	m.SetFilters(finalFilters)

	v, err := next.Mutate(ctx, m)
	if err != nil {
		return v, err
	}

	id, err := GetObjectIDFromEntValue(v)
	if err != nil {
		return v, err
	}

	args := jobspec.ExportContentArgs{
		ExportID:       id,
		UserID:         caller.SubjectID,
		OrganizationID: orgID,
	}

	if exportType == enums.ExportTypeEvidence {
		mode, _ := m.Mode()
		args.Mode = mode

		if metadata, ok := m.ExportMetadata(); ok {
			args.ExportMetadata = &metadata
		}
	}

	err = enqueueJob(ctx, m.Job, args, nil)

	return v, err
}

// getFinalFilters applies ownerID and systemOwned filters to the export filters if needed based on the Exportable annotation
func getFinalFilters(filters string, m *generated.ExportMutation, ownerID string) (string, error) {
	// check if it has owner field set
	exportType, _ := m.ExportType()
	addOwnerField := HasOwnerField(exportType.String())
	addSystemOwnedField := HasSystemOwnedField(exportType.String())

	if !addOwnerField && !addSystemOwnedField {
		return filters, nil
	}

	if filters == "" {
		filters = "{}"
	}

	var filterMap map[string]any

	if err := json.Unmarshal([]byte(filters), &filterMap); err != nil {
		return "", err
	}

	// filter by owner ID
	if addOwnerField {
		filterMap["ownerID"] = ownerID
	}

	// filter out system owned records
	if addSystemOwnedField {
		filterMap["systemOwned"] = false
	}

	finalFilters, err := json.Marshal(filterMap)
	if err != nil {
		return "", err
	}

	return string(finalFilters), nil
}

func handleExportUpdate(ctx context.Context, m *generated.ExportMutation, next ent.Mutator) (generated.Value, error) {
	fileIDs := pkgobjects.GetFileIDsFromContext(ctx)
	if len(fileIDs) > 0 {
		var err error

		ctx, err = checkExportFiles(ctx, m)
		if err != nil {
			return nil, err
		}

		m.AddFileIDs(fileIDs...)
	}

	return next.Mutate(ctx, m)
}

// checkExportFiles checks if export files are provided and sets the local file ID(s)
func checkExportFiles(ctx context.Context, m *generated.ExportMutation) (context.Context, error) {
	key := "exportFiles"

	files, _ := pkgobjects.FilesFromContextWithKey(ctx, key)
	if len(files) == 0 {
		return ctx, nil
	}

	return pkgobjects.ProcessFilesForMutation(ctx, m, key)
}
