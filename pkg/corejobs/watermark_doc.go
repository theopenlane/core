package corejobs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/riverqueue/river"
	"github.com/rs/zerolog"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
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
	// Initialize Openlane client if not already set
	if w.olClient == nil {
		olconfig := openlaneclient.NewDefaultConfig()

		baseURL, err := url.Parse(w.Config.OpenlaneAPIHost)
		if err != nil {
			return err
		}

		opts := []openlaneclient.ClientOption{openlaneclient.WithBaseURL(baseURL)}
		opts = append(opts, openlaneclient.WithCredentials(openlaneclient.Authorization{
			BearerToken: w.Config.OpenlaneAPIToken,
		}))

		cl, err := openlaneclient.New(olconfig, opts...)
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

	fmt.Printf("%+v\n", trustCenterDoc)

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

	fmt.Println(*trustCenterDoc.TrustCenterDoc.TrustCenterID)

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

	// Create temporary file for the original document
	originalTempFile, err := os.CreateTemp("", "original-*.pdf")
	if err != nil {
		logger.Error().Err(err).Msg("failed to create temporary file for original document")
		w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)
		return fmt.Errorf("failed to create temporary file for original document: %w", err)
	}
	defer func() {
		originalTempFile.Close()
		os.Remove(originalTempFile.Name())
	}()

	// Write original document to temporary file
	if _, err := originalTempFile.Write(originalDocBytes); err != nil {
		logger.Error().Err(err).Msg("failed to write original document to temporary file")
		w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)
		return fmt.Errorf("failed to write original document to temporary file: %w", err)
	}
	originalTempFile.Close() // Close file so pdfcpu can read it

	// Create temporary file for the watermarked document
	watermarkedTempFile, err := os.CreateTemp("", "watermarked-*.pdf")
	if err != nil {
		logger.Error().Err(err).Msg("failed to create temporary file for watermarked document")
		w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)
		return fmt.Errorf("failed to create temporary file for watermarked document: %w", err)
	}
	defer func() {
		watermarkedTempFile.Close()
		os.Remove(watermarkedTempFile.Name())
	}()
	watermarkedTempFile.Close() // Close file so pdfcpu can write to it

	// Convert client config to generated config for watermarking functions
	genConfig := w.convertWatermarkConfig(watermarkConfig)

	// Apply watermark based on config type
	switch {
	case watermarkConfig.Text != nil && *watermarkConfig.Text != "":
		// Text watermark
		err = watermarkPDFWithText(originalTempFile.Name(), watermarkedTempFile.Name(), genConfig)
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

		// Create temporary file for the watermark image with proper extension
		imageTempFile, err := os.CreateTemp("", "watermark-image-*.png")
		if err != nil {
			logger.Error().Err(err).Msg("failed to create temporary file for watermark image")
			w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)
			return fmt.Errorf("failed to create temporary file for watermark image: %w", err)
		}
		defer func() {
			imageTempFile.Close()
			os.Remove(imageTempFile.Name())
		}()

		// Write image to temporary file
		if _, err := imageTempFile.Write(imageBytes); err != nil {
			logger.Error().Err(err).Msg("failed to write watermark image to temporary file")
			w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)
			return fmt.Errorf("failed to write watermark image to temporary file: %w", err)
		}
		imageTempFile.Close() // Close file so pdfcpu can read it

		err = watermarkPDFWithImage(originalTempFile.Name(), watermarkedTempFile.Name(), imageTempFile.Name(), genConfig)
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

	// Read the watermarked document from the temporary file
	watermarkedDocBytes, err := os.ReadFile(watermarkedTempFile.Name())
	if err != nil {
		logger.Error().Err(err).Msg("failed to read watermarked document from temporary file")
		w.setWatermarkStatus(ctx, job.Args.TrustCenterDocumentID, enums.WatermarkStatusFailed)
		return fmt.Errorf("failed to read watermarked document from temporary file: %w", err)
	}

	uploadFile := &graphql.Upload{
		File:        bytes.NewReader(watermarkedDocBytes),
		Filename:    trustCenterDoc.TrustCenterDoc.OriginalFile.ProvidedFileName,
		Size:        int64(len(watermarkedDocBytes)),
		ContentType: http.DetectContentType(watermarkedDocBytes),
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

func watermarkPDFWithText(inputFile, outputFile string, config *generated.TrustCenterWatermarkConfig) error {
	// Create watermark description string
	// Format: "text:<text>, f:<font>, p:<points>, c:<color>, op:<opacity>, rot:<rotation>, pos:<position>"
	wmDesc := fmt.Sprintf("fontname:Helvetica, points:%d, fillcolor:%s, op:%.2f, rot:%.0f, pos:c",
		int(config.FontSize),
		config.Color,
		config.Opacity,
		config.Rotation,
	)

	fmt.Printf("Watermark description: %s\n", wmDesc)

	// Define watermark parameters
	selectedPages := []string{"1-"} // Apply to all pages. Use `nil` for all pages.
	onTop := true                   // true = watermark appears over content; false = under content

	// Apply text watermark to PDF using file-based operation
	if err := api.AddTextWatermarksFile(inputFile, outputFile, selectedPages, onTop, config.Text, wmDesc, nil); err != nil {
		return fmt.Errorf("failed to apply watermark: %v", err)
	}

	return nil
}

func watermarkPDFWithImage(inputFile, outputFile, imageFile string, config *generated.TrustCenterWatermarkConfig) error {
	// Create image watermark description
	wmDesc := fmt.Sprintf("op:%.2f, rot:%.0f, pos:c",
		config.Opacity,
		config.Rotation,
	)

	fmt.Printf("Image watermark description: %s\n", wmDesc)

	// Apply image watermark to PDF using file-based operation
	err := api.AddImageWatermarksFile(inputFile, outputFile, nil, true, imageFile, wmDesc, nil)
	if err != nil {
		return fmt.Errorf("failed to apply image watermark: %v", err)
	}

	return nil
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
