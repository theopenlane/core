package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
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
	ctx = logx.WithFields(ctx, logx.LogFields{"scan_id": cfg.ScanID, "part": cfg.Part})

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

		logx.FromContext(ctx).Warn().Err(err).Str("cache", docextract.CacheID(req.Cache)).Msg("report scan: document cache missing, sending the document")

		req.Cache = ""
	}

	pdf, err := downloadReportFile(ctx, request.DB, cfg.ReportFileID)
	if err != nil {
		return ReportScanPartResult{}, err
	}

	sections, err := client.Extract(ctx, bytes.NewReader(pdf), client.SOC2, req)
	if err != nil {
		return ReportScanPartResult{}, err
	}

	return ReportScanPartResult{Sections: sections}, nil
}

// downloadReport resolves the scan's attached pdf and fetches it, returning the id of the file it
// read so the submit can record it and hand it to every part job
func downloadReport(ctx context.Context, db *generated.Client, scanRecord *generated.Scan) ([]byte, string, error) {
	if db.ObjectManager == nil {
		return nil, "", ErrObjectManagerRequired
	}

	reportFile, err := resolveReportFile(ctx, scanRecord)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, "", ErrReportFileMissing
		}

		return nil, "", err
	}

	pdf, err := download(ctx, db, reportFile)
	if err != nil {
		return nil, "", err
	}

	return pdf, reportFile.ID, nil
}

// downloadReportFile fetches the report the submit already resolved, so a part job reads the file
// directly instead of walking the scan to find it again
func downloadReportFile(ctx context.Context, db *generated.Client, fileID string) ([]byte, error) {
	if db.ObjectManager == nil {
		return nil, ErrObjectManagerRequired
	}

	reportFile, err := db.File.Get(ctx, fileID)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, ErrReportFileMissing
		}

		return nil, err
	}

	return download(ctx, db, reportFile)
}

// download pulls the file's bytes out of object storage
func download(ctx context.Context, db *generated.Client, reportFile *generated.File) ([]byte, error) {
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

// resolveReportFile picks the scan's report by the file id recorded at submit, falling back to the
// oldest attached pdf for scans submitted before the id was recorded
func resolveReportFile(ctx context.Context, scanRecord *generated.Scan) (*generated.File, error) {
	query := scanRecord.QueryFiles().Where(file.DetectedContentTypeEQ(pdftext.ContentType))

	if fileID, ok := ReportFileID(scanRecord.Metadata); ok {
		query = query.Where(file.ID(fileID))
	}

	return query.Order(file.ByCreatedAt()).First(ctx)
}
