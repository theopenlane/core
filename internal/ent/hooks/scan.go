package hooks

import (
	"context"
	"errors"
	"slices"
	"time"

	"entgo.io/ent"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/internal/ent/generated/scan"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/gemini"
	pkgobjects "github.com/theopenlane/core/v2/pkg/objects"
)

var (
	// ErrReportScanFileRequired is returned when a report scan is created without a PDF attached
	ErrReportScanFileRequired = errors.New("scan must include a PDF report")
	// ErrScanOriginUnresolved is returned when a scan is created without a caller to derive its origin from
	ErrScanOriginUnresolved = errors.New("scan origin could not be determined from the caller")
	// ErrReportScanSingleFile is returned when more than one PDF report is attached to a report scan
	ErrReportScanSingleFile = errors.New("a report scan may only have one PDF report attached")
	// ErrReportScanFileImmutable is returned when the report is changed after parsing has started
	ErrReportScanFileImmutable = errors.New("the report cannot be changed once the scan has left pending")
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

// HookScanFiles runs on scan mutations to attach uploaded files, holding a report scan to a single
// report that cannot be swapped once parsing has started
func HookScanFiles() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ScanFunc(func(ctx context.Context, m *generated.ScanMutation) (generated.Value, error) {
			if isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			fileIDs := pkgobjects.GetFileIDsFromContext(ctx)

			if err := checkReportScanFiles(ctx, m, fileIDs); err != nil {
				return nil, err
			}

			if len(fileIDs) > 0 {
				var err error

				ctx, err = pkgobjects.ProcessFilesForMutation(ctx, m, "scanFiles")
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

// checkReportScanFiles holds a report scan to one attached report and rejects a file change once
// the scan has left pending; fileIDs are the uploads carried on the request context
func checkReportScanFiles(ctx context.Context, m *generated.ScanMutation, fileIDs []string) error {
	attached := lo.Union(fileIDs, m.FilesIDs())

	if m.Op().Is(ent.OpCreate) {
		if isReportScanMutation(m) && len(attached) > 1 {
			return ErrReportScanSingleFile
		}

		return nil
	}

	if len(attached) == 0 && len(m.RemovedFilesIDs()) == 0 && !m.FilesCleared() {
		return nil
	}

	ids, err := getMutationIDs(ctx, m)
	if err != nil {
		return err
	}

	reportScans, err := m.Client().Scan.Query().
		Where(
			scan.IDIn(ids...),
			scan.ScanTypeEQ(enums.ScanTypeReport),
			scan.PerformedBy(gemini.PerformedBy),
		).
		Select(scan.FieldID, scan.FieldStatus).
		All(ctx)
	if err != nil {
		return err
	}

	for _, reportScan := range reportScans {
		if reportScan.Status != enums.ScanStatusPending {
			return ErrReportScanFileImmutable
		}

		if err := checkReportScanFileCount(ctx, m, reportScan, len(attached)); err != nil {
			return err
		}
	}

	return nil
}

// checkReportScanFileCount rejects an update that would leave a report scan holding more than one report
func checkReportScanFileCount(ctx context.Context, m *generated.ScanMutation, reportScan *generated.Scan, attached int) error {
	existing, err := m.Client().Scan.QueryFiles(reportScan).Count(ctx)
	if err != nil {
		return err
	}

	if exceedsSingleReportFile(existing, len(m.RemovedFilesIDs()), attached, m.FilesCleared()) {
		return ErrReportScanSingleFile
	}

	return nil
}

// exceedsSingleReportFile reports whether a files edge change leaves a scan holding more than one report
func exceedsSingleReportFile(existing, removed, attached int, cleared bool) bool {
	remaining := existing - removed
	if cleared {
		remaining = 0
	}

	return remaining+attached > 1
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
