package corejobs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/riverqueue/river"
	"github.com/rs/zerolog"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// WatermarkDocArgs for the worker to process watermarking of a document
type WatermarkDocArgs struct {
	// ID of the trust center document
	TrustCenterDocumentID string `json:"trust_center_document_id"`
}

// Kind satisfies the river.Job interface
func (WatermarkDocArgs) Kind() string { return "watermark_doc" }

type WatermarkWorkerConfig struct {
	Enabled bool `koanf:"enabled" json:"enabled" jsonschema:"required description=whether the watermark worker is enabled"`

	OpenlaneAPIHost  string `koanf:"openlaneAPIHost" json:"openlaneAPIHost" jsonschema:"required description=the openlane api host"`
	OpenlaneAPIToken string `koanf:"openlaneAPIToken" json:"openlaneAPIToken" jsonschema:"required description=the openlane api token"`
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
	logger := zerolog.Ctx(ctx).With().Str("trust_center_document_id", job.Args.TrustCenterDocumentID).Logger()
	logger.Info().Msg("starting document watermarking")

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

	if trustCenterDoc.TrustCenterDoc.OriginalFile == nil || trustCenterDoc.TrustCenterDoc.OriginalFile.PresignedURL == nil || *trustCenterDoc.TrustCenterDoc.File.PresignedURL == "" {
		logger.Error().Msg("trust center document has no file or presigned URL")
		w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)
		return fmt.Errorf("trust center document has no file or presigned URL")
	}

	if trustCenterDoc.TrustCenterDoc.TrustCenterID == nil {
		logger.Error().Msg("trust center document has no trust center ID")
		w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)
		return fmt.Errorf("trust center document has no trust center ID")
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
		return fmt.Errorf("no watermark config found for trust center")
	}

	watermarkConfig := watermarkConfigs.TrustCenterWatermarkConfigs.Edges[0].Node

	// Download the original document
	originalDocBytes, err := w.downloadFile(*trustCenterDoc.TrustCenterDoc.OriginalFile.PresignedURL)
	if err != nil {
		logger.Error().Err(err).Msg("failed to download original document")
		w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)
		return fmt.Errorf("failed to download original document: %w", err)
	}

	// Create a buffer for the watermarked document
	var watermarkedDoc bytes.Buffer
	originalReader := bytes.NewReader(originalDocBytes)

	// Convert client config to generated config for watermarking functions
	genConfig := w.convertWatermarkConfig(watermarkConfig)

	// Apply watermark based on config type
	if watermarkConfig.Text != nil && *watermarkConfig.Text != "" {
		// Text watermark
		err = watermarkPDFWithText(originalReader, &watermarkedDoc, genConfig)
		if err != nil {
			logger.Error().Err(err).Msg("failed to apply text watermark")
			w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)
			return fmt.Errorf("failed to apply text watermark: %w", err)
		}
	} else if watermarkConfig.File != nil && watermarkConfig.File.PresignedURL != nil && *watermarkConfig.File.PresignedURL != "" {
		// Image watermark
		imageBytes, err := w.downloadFile(*watermarkConfig.File.PresignedURL)
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
	} else {
		logger.Error().Msg("watermark config has neither text nor image")
		w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)
		return fmt.Errorf("watermark config has neither text nor image")
	}

	// For now, we'll just log the successful watermarking and update the status
	// In a production environment, you would upload the watermarked document to storage
	// and update the document with the new file ID
	logger.Info().Int("watermarked_bytes", len(watermarkedDoc.Bytes())).Msg("document watermarked successfully")

	uploadFile := &graphql.Upload{
		File:        bytes.NewReader(watermarkedDoc.Bytes()),
		Filename:    trustCenterDoc.TrustCenterDoc.OriginalFile.ProvidedFileName,
		Size:        int64(watermarkedDoc.Len()),
		ContentType: http.DetectContentType(watermarkedDoc.Bytes()),
	}
	// Update the trust center document status to success
	successStatus := enums.WatermarkStatusSuccess
	originalFileID := trustCenterDoc.TrustCenterDoc.FileID
	_, err = w.olClient.UpdateTrustCenterDoc(ctx, job.Args.TrustCenterDocumentID, openlaneclient.UpdateTrustCenterDocInput{
		WatermarkStatus: &successStatus,
		OriginalFileID:  originalFileID,
	}, nil, uploadFile)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update trust center document status")
		w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)
		return fmt.Errorf("failed to update trust center document status: %w", err)
	}

	logger.Info().Msg("document watermarking completed successfully")
	return nil
}

func watermarkPDFWithText(rs io.ReadSeeker, w io.Writer, config *generated.TrustCenterWatermarkConfig) error {
	// Create watermark description string
	// Format: "text:<text>, f:<font>, p:<points>, c:<color>, op:<opacity>, rot:<rotation>, pos:<position>"
	wmDesc := fmt.Sprintf("fontname:Helvetica, points:%.2f, fillcolor:%s, op:%.2f, rot:%.0f",
		config.FontSize,
		config.Color,
		config.Opacity,
		config.Rotation,
	)
	// Define watermark parameters
	selectedPages := []string{"1-"} // Apply to all pages. Use `nil` for all pages.
	onTop := true                   // true = watermark appears over content; false = under content
	wm, err := api.TextWatermark(config.Text, wmDesc, onTop, false, types.POINTS)
	if err != nil {
		return err
	}

	fmt.Printf("Watermark description: %s\n", wmDesc)

	return api.AddWatermarks(rs, w, selectedPages, wm, nil)
}

func watermarkPDFWithImage(rs io.ReadSeeker, w io.Writer, imgReader io.Reader, config *generated.TrustCenterWatermarkConfig) error {
	// Create image watermark description
	wmDesc := fmt.Sprintf("op:%.2f, rot:%.0f",
		config.Opacity,
		config.Rotation,
	)

	fmt.Printf("Image watermark description: %s\n", wmDesc)

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
		zerolog.Ctx(ctx).Error().Err(err).Str("doc_id", docID).Str("status", status.String()).Msg("failed to set watermark status")
	}
}

// downloadFile downloads a file from the given URL and returns its bytes
func (w *WatermarkDocWorker) downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// convertWatermarkConfig converts the client watermark config to the generated type
func (w *WatermarkDocWorker) convertWatermarkConfig(clientConfig *openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges_Node) *generated.TrustCenterWatermarkConfig {
	config := &generated.TrustCenterWatermarkConfig{
		ID: clientConfig.ID,
	}

	if clientConfig.TrustCenterID != nil {
		config.TrustCenterID = *clientConfig.TrustCenterID
	}

	if clientConfig.Text != nil {
		config.Text = *clientConfig.Text
	}

	if clientConfig.FontSize != nil {
		config.FontSize = *clientConfig.FontSize
	}

	if clientConfig.Opacity != nil {
		config.Opacity = *clientConfig.Opacity
	}

	if clientConfig.Rotation != nil {
		config.Rotation = *clientConfig.Rotation
	}

	if clientConfig.Color != nil {
		config.Color = *clientConfig.Color
	}

	if clientConfig.Font != nil {
		config.Font = enums.Font(*clientConfig.Font)
	}

	return config
}
