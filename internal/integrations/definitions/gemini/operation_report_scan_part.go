package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	"github.com/theopenlane/core/v2/internal/ent/generated/scan"
	"github.com/theopenlane/core/v2/internal/ent/interceptors"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/docextract"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/objects"
	"github.com/theopenlane/core/v2/pkg/pdftext"
)

// Handle adapts ReportScanPartRequest to the generic operation registration boundary
func (r ReportScanPartRequest) Handle() types.OperationHandler {
	return providerkit.WithClientRequestConfig(geminiClient, ReportScanPartOp, ErrOperationConfigInvalid, func(ctx context.Context, request types.OperationRequest, client *Client, cfg ReportScanPartRequest) (json.RawMessage, error) {
		result, err := r.Run(ctx, request, client, cfg)
		if err != nil {
			return nil, err
		}

		return providerkit.EncodeResult(result, ErrResultEncode)
	})
}

// Run downloads the scan's pdf and extracts the single requested section
func (ReportScanPartRequest) Run(ctx context.Context, request types.OperationRequest, client *Client, cfg ReportScanPartRequest) (ReportScanPartResult, error) {
	ctx = logx.WithFields(ctx, logx.LogFields{"organization_id": cfg.OrganizationID, "scan_id": cfg.ScanID, "part": cfg.Part})
	systemCtx := scanSystemContext(ctx, cfg.OrganizationID)

	scanRecord, err := request.DB.Scan.Query().Where(
		scan.ID(cfg.ScanID),
		scan.OwnerID(cfg.OrganizationID),
		scan.ScanTypeEQ(enums.ScanTypeReport),
	).Only(systemCtx)
	if err != nil {
		return ReportScanPartResult{}, err
	}

	req := docextract.Request{Section: cfg.Part, Scope: cfg.RefCodes, Cache: cfg.Cache}

	// with a shared cache the part references it, otherwise the document is sent with the request
	if req.Cache != "" {
		sections, err := client.Extract(ctx, nil, client.SOC2, req)
		if err == nil {
			return ReportScanPartResult{Sections: sections}, nil
		}

		if !errors.Is(err, docextract.ErrCacheMissing) {
			return ReportScanPartResult{}, err
		}

		logx.FromContext(ctx).Warn().Err(err).Str("cache", docextract.CacheID(req.Cache)).Msg("report scan: document cache missing, rebuilding")
	}

	pdf, err := downloadReport(systemCtx, request.DB, scanRecord)
	if err != nil {
		return ReportScanPartResult{}, err
	}

	if req.Cache != "" {
		req.Cache, err = client.CacheDocument(ctx, bytes.NewReader(pdf), DocumentCacheTTL)
		if err != nil {
			return ReportScanPartResult{}, err
		}
	}

	sections, err := client.Extract(ctx, bytes.NewReader(pdf), client.SOC2, req)
	if err != nil {
		return ReportScanPartResult{}, err
	}

	return ReportScanPartResult{Sections: sections}, nil
}

// downloadReport fetches the scan's attached pdf from object storage
func downloadReport(ctx context.Context, db *generated.Client, scanRecord *generated.Scan) ([]byte, error) {
	if db.ObjectManager == nil {
		return nil, ErrObjectManagerRequired
	}

	reportFile, err := scanRecord.QueryFiles().Where(file.DetectedContentTypeEQ(pdftext.ContentType)).First(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, ErrReportFileMissing
		}

		return nil, err
	}

	storageFile := interceptors.StorageFileFromEnt(reportFile)

	downloaded, err := db.ObjectManager.Download(ctx, nil, storageFile, &objects.DownloadOptions{
		FileName:     reportFile.ProvidedFileName,
		ContentType:  reportFile.DetectedContentType,
		FileMetadata: storageFile.FileMetadata,
	})
	if err != nil {
		return nil, err
	}

	return downloaded.File, nil
}
