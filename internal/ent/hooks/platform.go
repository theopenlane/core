package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/logx"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

var platformFileKeys = []string{"architectureDiagrams", "dataFlowDiagrams", "trustBoundaryDiagrams"}

// HookPlatformFiles runs on platform mutations to check for uploaded files
func HookPlatformFiles() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.PlatformFunc(func(ctx context.Context, m *generated.PlatformMutation) (generated.Value, error) {
			// nothing to do if this is a delete operation
			if isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			fileIDs := pkgobjects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				for _, key := range platformFileKeys {
					ctx, err = pkgobjects.ProcessFilesForMutation(ctx, m, key)
					if err != nil {
						return nil, err
					}

					switch key {
					case "architectureDiagrams":
						files, _ := pkgobjects.FilesFromContextWithKey(ctx, key)
						archFiles := make([]string, len(files))
						for i, f := range files {
							archFiles[i] = f.ID
						}

						m.AddArchitectureDiagramIDs(archFiles...)
						if err := m.Client().File.Update().Where(file.IDIn(archFiles...)).SetCategoryType("architecture_diagram").Exec(ctx); err != nil {
							logx.FromContext(ctx).Error().Err(err).Msg("unable to set file category type for architecture diagram files")
							return nil, err
						}
					case "dataFlowDiagrams":
						files, _ := pkgobjects.FilesFromContextWithKey(ctx, key)
						dataFlowFiles := make([]string, len(files))
						for i, f := range files {
							dataFlowFiles[i] = f.ID
						}

						m.AddDataFlowDiagramIDs(dataFlowFiles...)
						if err := m.Client().File.Update().Where(file.IDIn(dataFlowFiles...)).SetCategoryType("data_flow_diagram").Exec(ctx); err != nil {
							logx.FromContext(ctx).Error().Err(err).Msg("unable to set file category type for data flow diagram files")
							return nil, err
						}
					case "trustBoundaryDiagrams":
						files, _ := pkgobjects.FilesFromContextWithKey(ctx, key)
						trustBoundaryFiles := make([]string, len(files))
						for i, f := range files {
							trustBoundaryFiles[i] = f.ID
						}

						m.AddTrustBoundaryDiagramIDs(trustBoundaryFiles...)
						if err := m.Client().File.Update().Where(file.IDIn(trustBoundaryFiles...)).SetCategoryType("trust_boundary_diagram").Exec(ctx); err != nil {
							logx.FromContext(ctx).Error().Err(err).Msg("unable to set file category type for trust boundary diagram files")
							return nil, err
						}
					}

				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}
