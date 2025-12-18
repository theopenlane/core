package corejobs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/riverqueue/river"
	"github.com/samber/lo"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

const (
	pdfMimeType = "application/pdf"
)

var (
	// ErrTrustCenterDocumentNoFile is returned when a trust center document has no file or presigned URL
	ErrTrustCenterDocumentNoFile = errors.New("trust center document has no file or presigned URL")
	// ErrTrustCenterDocumentNoTrustCenterID is returned when a trust center document has no trust center ID
	ErrTrustCenterDocumentNoTrustCenterID = errors.New("trust center document has no trust center ID")
	// ErrNoWatermarkConfigFound is returned when no watermark config is found for a trust center
	ErrNoWatermarkConfigFound = errors.New("no watermark config found for trust center")
	// ErrWatermarkConfigNoTextOrImage is returned when watermark config has neither text nor image
	ErrWatermarkConfigNoTextOrImage = errors.New("watermark config has neither text nor image")
	// ErrFileDownloadFailed is returned when file download fails with non-200 status code
	ErrFileDownloadFailed = errors.New("failed to download file")
)

// WatermarkDocArgs for the worker to process watermarking of a document
type WatermarkDocArgs struct {
	// TrustCenterDocumentID is the ID of the trust center document to watermark
	TrustCenterDocumentID string `json:"trust_center_document_id"`
}

// Kind satisfies the river.Job interface
func (WatermarkDocArgs) Kind() string { return "watermark_doc" }

type WatermarkWorkerConfig struct {
	OpenlaneConfig `koanf:",squash" jsonschema:"description=the openlane API configuration for watermarking"`

	Enabled bool `koanf:"enabled" json:"enabled" jsonschema:"required description=whether the watermark worker is enabled"`
}

// WatermarkDocWorker is the worker to process watermarking of a document
type WatermarkDocWorker struct {
	river.WorkerDefaults[WatermarkDocArgs]

	Config WatermarkWorkerConfig `koanf:"config" json:"config" jsonschema:"description=the configuration for watermarking"`

	olClient olclient.OpenlaneClient
}

// WithOpenlaneClient sets the openlane client to use for API requests
func (w *WatermarkDocWorker) WithOpenlaneClient(cl olclient.OpenlaneClient) *WatermarkDocWorker {
	w.olClient = cl
	return w
}

// Work satisfies the river.Worker interface for the watermark doc worker
func (w *WatermarkDocWorker) Work(ctx context.Context, job *river.Job[WatermarkDocArgs]) error {
	logger := logx.FromContext(ctx).With().Str("trust_center_document_id", job.Args.TrustCenterDocumentID).Logger()
	logger.Info().Msg("starting document watermarking")

	if w.olClient == nil {
		cl, err := w.Config.getOpenlaneClient()
		if err != nil {
			return err
		}

		w.olClient = cl
	}

	// Set status to in progress
	inProgressStatus := enums.WatermarkStatusInProgress
	_, err := w.olClient.UpdateTrustCenterDoc(ctx, job.Args.TrustCenterDocumentID, openlaneclient.UpdateTrustCenterDocInput{
		WatermarkStatus: &inProgressStatus,
	}, nil, nil)
	if err != nil {
		logger.Error().Err(err).Msg("failed to set watermark status to in progress")
		return fmt.Errorf("failed to set watermark status to in progress: %w", err)
	}

	// Fetch the trust center document
	trustCenterDoc, err := w.olClient.GetTrustCenterDocByID(ctx, job.Args.TrustCenterDocumentID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch trust center document")
		w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)
		return fmt.Errorf("failed to fetch trust center document: %w", err)
	}

	if trustCenterDoc.TrustCenterDoc.OriginalFile == nil || trustCenterDoc.TrustCenterDoc.OriginalFile.PresignedURL == nil || *trustCenterDoc.TrustCenterDoc.OriginalFile.PresignedURL == "" {
		logger.Error().Msg("trust center document has no file or presigned URL")
		w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)
		return ErrTrustCenterDocumentNoFile
	}

	if trustCenterDoc.TrustCenterDoc.TrustCenterID == nil {
		logger.Error().Msg("trust center document has no trust center ID")
		w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)
		return ErrTrustCenterDocumentNoTrustCenterID
	}

	// Fetch the trust center watermark config
	watermarkConfigs, err := w.olClient.GetTrustCenterWatermarkConfigs(ctx, nil, nil, &openlaneclient.TrustCenterWatermarkConfigWhereInput{
		TrustCenterID: trustCenterDoc.TrustCenterDoc.TrustCenterID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch trust center watermark config")
		w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)
		return fmt.Errorf("failed to fetch trust center watermark config: %w", err)
	}

	if len(watermarkConfigs.TrustCenterWatermarkConfigs.Edges) == 0 {
		logger.Error().Msg("no watermark config found for trust center")
		w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)
		return ErrNoWatermarkConfigFound
	}

	watermarkConfig := watermarkConfigs.TrustCenterWatermarkConfigs.Edges[0].Node

	// Download the original document
	originalDocBytes, err := w.downloadFile(ctx, *trustCenterDoc.TrustCenterDoc.OriginalFile.PresignedURL)
	if err != nil {
		logger.Error().Err(err).Msg("failed to download original document")
		w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)
		return fmt.Errorf("failed to download original document: %w", err)
	}

	contentType := http.DetectContentType(originalDocBytes)
	if contentType != pdfMimeType {
		_, err = w.olClient.UpdateTrustCenterDoc(ctx, job.Args.TrustCenterDocumentID, openlaneclient.UpdateTrustCenterDocInput{
			WatermarkStatus: lo.ToPtr(enums.WatermarkStatusSuccess),
		}, nil, nil)
		if err != nil {
			logger.Error().Err(err).Str("mimetype", contentType).
				Msg("failed to update status for document")
			return fmt.Errorf("failed to update status for non-PDF document: %w", err)
		}
		return nil
	}

	// Create a buffer for the watermarked document
	var watermarkedDoc bytes.Buffer
	originalReader := bytes.NewReader(originalDocBytes)

	// Convert client config to generated config for watermarking functions
	genConfig := w.convertWatermarkConfig(watermarkConfig)

	// Apply watermark based on config type
	switch {
	case watermarkConfig.Text != nil && *watermarkConfig.Text != "":
		// Text watermark
		err = watermarkPDFWithText(originalReader, &watermarkedDoc, genConfig)
		if err != nil {
			logger.Error().Err(err).Msg("failed to apply text watermark")
			w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)

			return fmt.Errorf("failed to apply text watermark: %w", err)
		}
	case watermarkConfig.File != nil && watermarkConfig.File.PresignedURL != nil && *watermarkConfig.File.PresignedURL != "":
		// Image watermark
		imageBytes, err := w.downloadFile(ctx, *watermarkConfig.File.PresignedURL)
		if err != nil {
			logger.Error().Err(err).Msg("failed to download watermark image")
			w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)

			return fmt.Errorf("failed to download watermark image: %w", err)
		}

		imageReader := bytes.NewReader(imageBytes)

		err = watermarkPDFWithImage(originalReader, &watermarkedDoc, imageReader, genConfig)
		if err != nil {
			logger.Error().Err(err).Msg("failed to apply image watermark")
			w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)

			return fmt.Errorf("failed to apply image watermark: %w", err)
		}
	default:
		logger.Error().Msg("watermark config has neither text nor image")
		w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)

		return ErrWatermarkConfigNoTextOrImage
	}

	uploadFile := &graphql.Upload{
		File:        bytes.NewReader(watermarkedDoc.Bytes()),
		Filename:    fmt.Sprintf("watermarked_%s", trustCenterDoc.TrustCenterDoc.OriginalFile.ProvidedFileName),
		Size:        int64(watermarkedDoc.Len()),
		ContentType: http.DetectContentType(watermarkedDoc.Bytes()),
	}
	// Update the trust center document status to success
	successStatus := enums.WatermarkStatusSuccess

	_, err = w.olClient.UpdateTrustCenterDoc(ctx, job.Args.TrustCenterDocumentID, openlaneclient.UpdateTrustCenterDocInput{
		WatermarkStatus: &successStatus,
	}, nil, uploadFile)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update trust center document status")
		w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)

		return fmt.Errorf("failed to update trust center document status: %w", err)
	}

	logger.Info().Msg("document watermarking completed successfully")

	return nil
}

func watermarkPDFWithText(rs io.ReadSeeker, w io.Writer, config *openlaneclient.TrustCenterWatermarkConfig) error {
	// Create watermark description string
	// You can find the allowed configuration docs here: https://pdfcpu.io/core/watermark.html
	wmDesc := fmt.Sprintf("fontname:%s, points:%.0f, fillcolor:%s, op:%.2f, rot:%.0f",
		config.Font.ToFontStr(),
		*config.FontSize,
		*config.Color,
		*config.Opacity,
		*config.Rotation,
	)

	// Define watermark parameters
	selectedPages := []string{"1-"} // Apply to all pages. Use `nil` for all pages.
	onTop := true                   // true = watermark appears over content; false = under content

	wm, err := api.TextWatermark(*config.Text, wmDesc, onTop, false, types.POINTS)
	if err != nil {
		return err
	}

	return api.AddWatermarks(rs, w, selectedPages, wm, nil)
}

func watermarkPDFWithImage(rs io.ReadSeeker, w io.Writer, imgReader io.Reader, config *openlaneclient.TrustCenterWatermarkConfig) error {
	// Create image watermark description
	// You can find the allowed configuration docs here: https://pdfcpu.io/core/watermark.html
	wmDesc := fmt.Sprintf("op:%.2f, rot:%.0f",
		*config.Opacity,
		*config.Rotation,
	)

	selectedPages := []string{"1-"} // Apply to all pages. Use `nil` for all pages.
	onTop := true                   // true = watermark appears over content; false = under content
	unit := types.POINTS

	wm, err := api.ImageWatermarkForReader(imgReader, wmDesc, onTop, false, unit)
	if err != nil {
		return err
	}

	return api.AddWatermarks(rs, w, selectedPages, wm, nil)
}

// setWatermarkStatus is a helper function to set the watermark status
func (w *WatermarkDocWorker) setWatermarkStatus(ctx context.Context, docID string, status enums.WatermarkStatus) {
	_, err := w.olClient.UpdateTrustCenterDoc(ctx, docID, openlaneclient.UpdateTrustCenterDocInput{
		WatermarkStatus: &status,
	}, nil, nil)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("doc_id", docID).Str("status", status.String()).Msg("failed to set watermark status")
	}
}

// downloadFile downloads a file from the given URL and returns its bytes
func (w *WatermarkDocWorker) downloadFile(ctx context.Context, url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrFileDownloadFailed, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// convertWatermarkConfig converts the client watermark config to the generated type
func (w *WatermarkDocWorker) convertWatermarkConfig(clientConfig *openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges_Node) *openlaneclient.TrustCenterWatermarkConfig {
	config := &openlaneclient.TrustCenterWatermarkConfig{
		ID: clientConfig.ID,
	}

	if clientConfig.TrustCenterID != nil {
		config.TrustCenterID = clientConfig.TrustCenterID
	}

	if clientConfig.Text != nil {
		config.Text = clientConfig.Text
	}

	if clientConfig.FontSize != nil {
		config.FontSize = clientConfig.FontSize
	}

	if clientConfig.Opacity != nil {
		config.Opacity = clientConfig.Opacity
	}

	if clientConfig.Rotation != nil {
		config.Rotation = clientConfig.Rotation
	}

	if clientConfig.Color != nil {
		config.Color = clientConfig.Color
	}

	if clientConfig.Font != nil {
		config.Font = clientConfig.Font
	} else {
		// Set default font if none is provided
		config.Font = &enums.FontHelvetica
	}

	return config
}
