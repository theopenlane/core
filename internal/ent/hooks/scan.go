package hooks

import (
	"context"
	"errors"
	"slices"
	"time"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/gemini"
	pkgobjects "github.com/theopenlane/core/v2/pkg/objects"
)

var (
	// ErrReportScanFileRequired is returned when a report scan is created without a PDF attached
	ErrReportScanFileRequired = errors.New("scan must include a PDF report")
	// ErrScanOriginUnresolved is returned when a scan is created without a caller to derive its origin from
	ErrScanOriginUnresolved = errors.New("scan origin could not be determined from the caller")
)

// scanDateTypes are the scan types that run once when they are submitted, so the submit time is
// the time they executed; scheduled scans record theirs per run instead
var scanDateTypes = []enums.ScanType{enums.ScanTypeDomain, enums.ScanTypeReport}

// HookScanDefaults stamps the values derived when a scan is submitted
func HookScanDefaults() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ScanFunc(func(ctx context.Context, m *generated.ScanMutation) (generated.Value, error) {
			origin, err := scanOriginFromContext(ctx)
			if err != nil {
				return nil, err
			}

			m.SetOrigin(origin)

			if _, ok := m.ScanDate(); !ok && slices.Contains(scanDateTypes, mutationScanType(m)) {
				m.SetScanDate(models.DateTime(time.Now()))
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

// HookScanFiles runs on scan mutations to attach uploaded files
func HookScanFiles() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ScanFunc(func(ctx context.Context, m *generated.ScanMutation) (generated.Value, error) {
			if isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			fileIDs := pkgobjects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = pkgobjects.ProcessFilesForMutation(ctx, m, gemini.FileUploadKey)
				if err != nil {
					return nil, err
				}

				m.AddFileIDs(fileIDs...)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// HookScanReportSubmit rejects a report scan created without a file to parse
func HookScanReportSubmit() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ScanFunc(func(ctx context.Context, m *generated.ScanMutation) (generated.Value, error) {
			if !isReportScanMutation(m) {
				return next.Mutate(ctx, m)
			}

			if len(pkgobjects.GetFileIDsFromContext(ctx)) == 0 && len(m.FilesIDs()) == 0 {
				return nil, ErrReportScanFileRequired
			}

			metadata, _ := m.Metadata()
			if _, err := gemini.RequestedParts(metadata); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

// isReportScanMutation reports whether the mutation creates a system-parsed report scan
func isReportScanMutation(m *generated.ScanMutation) bool {
	scanType, _ := m.ScanType()
	performedBy, _ := m.PerformedBy()

	return scanType == enums.ScanTypeReport && performedBy == gemini.PerformedBy
}

// mutationScanType is the type the scan will be stored with, falling back to the schema default
// when the create did not set one
func mutationScanType(m *generated.ScanMutation) enums.ScanType {
	if scanType, ok := m.ScanType(); ok {
		return scanType
	}

	return enums.ScanTypeDomain
}

// scanOriginFromContext derives the origin in a fixed order: an integration actor, then an
// internal operation, then the user's authentication type
func scanOriginFromContext(ctx context.Context) (enums.ScanOrigin, error) {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		return "", ErrScanOriginUnresolved
	}

	switch {
	case caller.Has(auth.CapIntegrationActor):
		return enums.ScanOriginIntegration, nil
	case caller.Has(auth.CapInternalOperation):
		return enums.ScanOriginSystem, nil
	}

	switch caller.AuthenticationType {
	case auth.JWTAuthentication:
		return enums.ScanOriginUser, nil
	case auth.PATAuthentication, auth.APITokenAuthentication:
		return enums.ScanOriginAPI, nil
	default:
		return "", ErrScanOriginUnresolved
	}
}
