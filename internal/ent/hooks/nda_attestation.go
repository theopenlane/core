package hooks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/font"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/internal/objects/upload"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage/proxy"
)

// signedNDADocumentData captures the expected structure of the document data for a trust center NDA submission
type signedNDADocumentData struct {
	PDFFileID         string               `json:"pdf_file_id" jsonschema:"required"`
	Acknowledgment    bool                 `json:"acknowledgment" jsonschema:"required"`
	SignatoryInfo     signatoryInformation `json:"signatory_info" jsonschema:"required"`
	TrustCenterID     string               `json:"trust_center_id" jsonschema:"required"`
	SignatureMetadata signatureMetadata    `json:"signature_metadata" jsonschema:"required"`
}

// signatoryInformation captures the key details of the NDA signer for attestation purposes
type signatoryInformation struct {
	Email       string `json:"email" jsonschema:"required,format=email"`
	LastName    string `json:"last_name" jsonschema:"required,minLength=1"`
	FirstName   string `json:"first_name" jsonschema:"required,minLength=1"`
	CompanyName string `json:"company_name" jsonschema:"required,minLength=1"`
}

// signatureMetadata captures the contextual details of the NDA signing event for attestation purposes
type signatureMetadata struct {
	UserID    string `json:"user_id" jsonschema:"required"`
	PDFHash   string `json:"pdf_hash" jsonschema:"required"`
	Timestamp string `json:"timestamp" jsonschema:"required,format=date-time"`
	IPAddress string `json:"ip_address" jsonschema:"required"`
	UserAgent string `json:"user_agent"`
}

// ndaAttestationResult holds the output of the NDA attestation process
type ndaAttestationResult struct {
	AttestedPDF     []byte
	AttestedPDFHash string
	TrustCenterURL  string
	OrgName         string
	TemplateFileID  string
}

// attestNDADocument performs the full NDA attestation flow: downloads the original PDF,
// appends an attestation certificate, uploads the result, and resolves trust center metadata
func attestNDADocument(ctx context.Context, client *generated.Client, docData *generated.DocumentData, templateID, trustCenterID string) (*ndaAttestationResult, error) {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	templateFile, err := fetchNDATemplateFile(allowCtx, client, templateID)
	if err != nil {
		return nil, err
	}

	var ndaMetadata signedNDADocumentData
	if err := jsonx.RoundTrip(docData.Data, &ndaMetadata); err != nil {
		return nil, ErrFailedToUnmarshalNDAMetadata
	}

	storageFile := interceptors.StorageFileFromEnt(templateFile)

	getFileCtx := objects.WithModuleHint(ctx, models.CatalogTrustCenterModule)

	downloaded, err := client.ObjectManager.Download(getFileCtx, nil, storageFile, nil)
	if err != nil {
		logx.FromContext(ctx).Error().Str("file", storageFile.ID).Str("provider", templateFile.StorageProvider).Err(err).Msg("failed to download original PDF")
		return nil, ErrFailedToDownloadNDAPDF
	}

	attestedPDF, err := appendAttestationPage(bytes.NewReader(downloaded.File), &ndaMetadata)
	if err != nil {
		return nil, ErrFailedToCreateAttestedPDF
	}

	pdfHash := sha256.Sum256(attestedPDF)
	attestedPDFHash := hex.EncodeToString(pdfHash[:])

	if err := uploadAttestedPDF(allowCtx, client, attestedPDF, docData, attestedPDFHash, templateFile.ID); err != nil {
		return nil, err
	}

	tc, err := client.TrustCenter.Query().
		Where(trustcenter.IDEQ(trustCenterID)).
		WithSetting().
		WithCustomDomain().
		Only(allowCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", trustCenterID).Msg("failed to fetch trust center")

		return nil, ErrFailedToFetchTrustCenter
	}

	orgName := resolveOrgName(tc)

	var customDomain string
	if tc.Edges.CustomDomain != nil {
		customDomain = tc.Edges.CustomDomain.CnameRecord
	}

	tcURL := buildTrustCenterURL(customDomain, tc.Slug)

	return &ndaAttestationResult{
		AttestedPDF:     attestedPDF,
		AttestedPDFHash: attestedPDFHash,
		TrustCenterURL:  tcURL,
		OrgName:         orgName,
		TemplateFileID:  templateFile.ID,
	}, nil
}

// uploadAttestedPDF uploads the attested PDF to storage, associates it with the document data,
// and stores the SHA-256 hash of the final PDF on the document data record
func uploadAttestedPDF(ctx context.Context, client *generated.Client, attestedPDF []byte, docData *generated.DocumentData, pdfHash, templateFileID string) error {
	fileName := fmt.Sprintf("attested_%s", templateFileID)

	file := pkgobjects.File{
		RawFile:              bytes.NewReader(attestedPDF),
		OriginalName:         fileName,
		FieldName:            "documentDataFile",
		CorrelatedObjectID:   docData.ID,
		CorrelatedObjectType: "DocumentData",
		FileMetadata: pkgobjects.FileMetadata{
			ContentType: "application/pdf",
			Size:        int64(len(attestedPDF)),
			Key:         "documentDataFile",
		},
	}

	_, uploadedFiles, err := upload.HandleUploads(ctx, client.ObjectManager, []pkgobjects.File{file})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to upload pdf")
		return ErrFailedToUploadAttestedPDF
	}

	if len(uploadedFiles) == 0 {
		logx.FromContext(ctx).Error().Msg("no files were uploaded")
		return ErrNoUploadedFiles
	}

	data := docData.Data
	data["attested_pdf_hash"] = pdfHash

	if err := client.DocumentData.UpdateOneID(docData.ID).
		AddFileIDs(uploadedFiles[0].ID).
		SetData(data).
		Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to associate files")
		return ErrFailedToAssociateFile
	}

	return nil
}

// resolveOrgName extracts the organization name from a trust center
func resolveOrgName(tc *generated.TrustCenter) string {
	if tc.Edges.Setting != nil && tc.Edges.Setting.CompanyName != "" {
		return tc.Edges.Setting.CompanyName
	}

	return ""
}

// fetchNDATemplateFile loads the NDA template and returns its first associated file
func fetchNDATemplateFile(ctx context.Context, client *generated.Client, templateID string) (*generated.File, error) {
	docTemplate, err := client.Template.Query().Where(template.ID(templateID)).Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to fetch nda template")
		return nil, ErrFailedToFetchNDATemplate
	}

	fileCtx := proxy.WithPresignInterceptorBypass(ctx)

	files, err := docTemplate.QueryFiles().All(fileCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to fetch nda template files")
		return nil, ErrFailedToFetchNDATemplateFiles
	}

	if len(files) == 0 {
		return nil, ErrMissingNDATemplateFile
	}

	return files[0], nil
}

// appendAttestationPage merges the original PDF with a generated attestation certificate page
func appendAttestationPage(originalPDF io.ReadSeeker, data *signedNDADocumentData) ([]byte, error) {
	certPage, err := createAttestationCertificate(data)
	if err != nil {
		return nil, ErrFailedToCreateAttestationCert
	}

	var buf bytes.Buffer

	readers := []io.ReadSeeker{originalPDF, bytes.NewReader(certPage)}

	if err = api.MergeRaw(readers, &buf, false, nil); err != nil {
		return nil, ErrFailedToMergeAttestationPage
	}

	return buf.Bytes(), nil
}

const (
	attestTitleSize      = 16
	attestDescSize       = 9
	attestFieldSize      = 10
	attestMargin         = 60.0
	attestContentWidth   = 475.0
	attestLabelColWidth  = 115.0
	attestCellPadX       = 8.0
	attestCellPadY       = 6.0
	attestCellPadSides   = 2 * attestCellPadX
	attestCellPadEnds    = 2 * attestCellPadY
	attestTextLineHeight = 14.0
	attestGridLineWidth  = 1.0
	attestTitleY         = 70.0
	attestTopDividerY    = 98.0
	attestDescY          = 118.0
	attestFieldStartY    = 170.0

	attestFontRegular  = "Helvetica"
	attestFontBold     = "Helvetica-Bold"
	attestGridColor    = "#aaaaaa"
	attestLabelBgColor = "#f5f5f5"
	attestTitleColor   = "#1a1a1a"
	attestDescColor    = "#666666"
	attestLabelColor   = "#333333"
)

var (
	fontRefTitle = attestationFontRef{Name: "$title"}
	fontRefDesc  = attestationFontRef{Name: "$desc"}
	fontRefLabel = attestationFontRef{Name: "$label"}
	fontRefValue = attestationFontRef{Name: "$value"}
)

// attestationField represents a label-value pair rendered on the attestation certificate
type attestationField struct {
	Label string
	Value string
}

// attestationPage is the JSON structure fed to pdfcpu's Create API
type attestationPage struct {
	Paper  string                        `json:"paper"`
	Origin string                        `json:"origin"`
	Fonts  map[string]attestationFont    `json:"fonts"`
	Pages  map[string]attestationContent `json:"pages"`
}

// attestationFont describes a named font for the pdfcpu JSON schema
type attestationFont struct {
	Name string `json:"name"`
	Size int    `json:"size"`
	Col  string `json:"col,omitempty"`
}

// attestationContent wraps the content block within a page
type attestationContent struct {
	Content attestationPageContent `json:"content"`
}

// attestationPageContent holds the elements rendered on a page
type attestationPageContent struct {
	Text []attestationTextBox   `json:"text,omitempty"`
	Box  []attestationSimpleBox `json:"box,omitempty"`
}

// attestationTextBox describes a positioned text element
type attestationTextBox struct {
	Value string             `json:"value"`
	Pos   [2]float64         `json:"pos"`
	Font  attestationFontRef `json:"font"`
	Width float64            `json:"width,omitempty"`
}

// attestationFontRef references a named font
type attestationFontRef struct {
	Name string `json:"name"`
}

// attestationSimpleBox describes a positioned filled rectangle
type attestationSimpleBox struct {
	Pos     [2]float64 `json:"pos"`
	Width   float64    `json:"width"`
	Height  float64    `json:"height"`
	FillCol string     `json:"fillCol"`
}

// attestationRowLayout captures the computed position and wrapped text for a single table row
type attestationRowLayout struct {
	field   attestationField
	wrapped string
	height  float64
	y       float64
}

// createAttestationCertificate generates a single-page PDF with the signature certification details
// rendered as a visible table with labeled rows, cell backgrounds, and grid lines
func createAttestationCertificate(data *signedNDADocumentData) ([]byte, error) {
	rows, tableHeight := layoutAttestationRows(attestationFieldsFrom(data))
	textBoxes := buildAttestationText(rows)
	boxes := buildAttestationBoxes(rows, tableHeight)

	return renderAttestationPage(textBoxes, boxes)
}

// attestationFieldsFrom extracts the label-value pairs from signed NDA metadata
func attestationFieldsFrom(data *signedNDADocumentData) []attestationField {
	return []attestationField{
		{"Name", data.SignatoryInfo.FirstName + " " + data.SignatoryInfo.LastName},
		{"Email", data.SignatoryInfo.Email},
		{"Company", data.SignatoryInfo.CompanyName},
		{"Date Signed", formatAttestTimestamp(data.SignatureMetadata.Timestamp)},
		{"IP Address", data.SignatureMetadata.IPAddress},
		{"User Agent", data.SignatureMetadata.UserAgent},
		{"Document Hash", data.SignatureMetadata.PDFHash},
	}
}

// layoutAttestationRows wraps field values and computes row heights and y positions
func layoutAttestationRows(fields []attestationField) ([]attestationRowLayout, float64) {
	valueTextWidth := attestContentWidth - attestLabelColWidth - attestCellPadSides

	var rows []attestationRowLayout

	y := attestFieldStartY

	for _, f := range fields {
		wrapped := wrapText(f.Value, attestFontRegular, attestFieldSize, valueTextWidth)
		lineCount := strings.Count(wrapped, "\n") + 1
		height := float64(lineCount)*attestTextLineHeight + attestCellPadEnds
		rows = append(rows, attestationRowLayout{field: f, wrapped: wrapped, height: height, y: y})
		y += height
	}

	return rows, y - attestFieldStartY
}

// buildAttestationText creates the header and per-row text elements for the attestation table
func buildAttestationText(rows []attestationRowLayout) []attestationTextBox {
	baselinePad := attestCellPadY + font.Ascent(attestFontRegular, attestFieldSize)

	textBoxes := []attestationTextBox{
		{
			Value: "SIGNATURE CERTIFICATION",
			Pos:   [2]float64{attestMargin, attestTitleY},
			Font:  fontRefTitle,
		},
		{
			Value: "This document certifies that the individual identified below has electronically signed the attached Non-Disclosure Agreement.",
			Pos:   [2]float64{attestMargin, attestDescY},
			Font:  fontRefDesc,
			Width: attestContentWidth,
		},
	}

	for _, r := range rows {
		textBoxes = append(textBoxes, attestationTextBox{
			Value: r.field.Label,
			Pos:   [2]float64{attestMargin + attestCellPadX, r.y + baselinePad},
			Font:  fontRefLabel,
		})

		for i, line := range strings.Split(r.wrapped, "\n") {
			textBoxes = append(textBoxes, attestationTextBox{
				Value: line,
				Pos:   [2]float64{attestMargin + attestLabelColWidth + attestCellPadX, r.y + baselinePad + float64(i)*attestTextLineHeight},
				Font:  fontRefValue,
			})
		}
	}

	return textBoxes
}

// buildAttestationBoxes creates the label backgrounds, grid lines, and borders for the attestation table.
// pdfcpu SimpleBox in UpperLeft origin extends height upward from pos,
// so every box y is offset by +height to align with text positions
func buildAttestationBoxes(rows []attestationRowLayout, tableHeight float64) []attestationSimpleBox {
	boxes := []attestationSimpleBox{
		{
			Pos:     [2]float64{attestMargin, attestTopDividerY + attestGridLineWidth},
			Width:   attestContentWidth,
			Height:  attestGridLineWidth,
			FillCol: attestGridColor,
		},
	}

	for _, r := range rows {
		boxes = append(boxes, attestationSimpleBox{
			Pos:     [2]float64{attestMargin, r.y + r.height},
			Width:   attestLabelColWidth,
			Height:  r.height,
			FillCol: attestLabelBgColor,
		})
	}

	lineY := attestFieldStartY
	for i := 0; i <= len(rows); i++ {
		boxes = append(boxes, attestationSimpleBox{
			Pos:     [2]float64{attestMargin, lineY + attestGridLineWidth},
			Width:   attestContentWidth,
			Height:  attestGridLineWidth,
			FillCol: attestGridColor,
		})

		if i < len(rows) {
			lineY += rows[i].height
		}
	}

	tableBottom := attestFieldStartY + tableHeight
	for _, x := range []float64{attestMargin, attestMargin + attestLabelColWidth, attestMargin + attestContentWidth} {
		boxes = append(boxes, attestationSimpleBox{
			Pos:     [2]float64{x, tableBottom},
			Width:   attestGridLineWidth,
			Height:  tableHeight,
			FillCol: attestGridColor,
		})
	}

	return boxes
}

// renderAttestationPage assembles the page JSON structure and calls pdfcpu's Create API
func renderAttestationPage(textBoxes []attestationTextBox, boxes []attestationSimpleBox) ([]byte, error) {
	page := attestationPage{
		Paper:  "A4P",
		Origin: "UpperLeft",
		Fonts: map[string]attestationFont{
			"title": {Name: attestFontBold, Size: attestTitleSize, Col: attestTitleColor},
			"desc":  {Name: attestFontRegular, Size: attestDescSize, Col: attestDescColor},
			"label": {Name: attestFontBold, Size: attestFieldSize, Col: attestLabelColor},
			"value": {Name: attestFontRegular, Size: attestFieldSize},
		},
		Pages: map[string]attestationContent{
			"1": {Content: attestationPageContent{
				Text: textBoxes,
				Box:  boxes,
			}},
		},
	}

	jsonData, err := json.Marshal(page)
	if err != nil {
		return nil, ErrFailedToGenerateAttestationPDF
	}

	var buf bytes.Buffer
	if err := api.Create(nil, bytes.NewReader(jsonData), &buf, nil); err != nil {
		return nil, ErrFailedToGenerateAttestationPDF
	}

	return buf.Bytes(), nil
}

// wrapText breaks s into lines that fit within maxWidth points for the given core font and size.
// pdfcpu's WriteColumn scales fonts down instead of wrapping, so we pre-insert newlines.
// Words that exceed maxWidth on their own (e.g. hex hashes) are broken at character boundaries
func wrapText(s, fontName string, fontSize int, maxWidth float64) string {
	if font.TextWidth(s, fontName, fontSize) <= maxWidth {
		return s
	}

	var lines []string

	var line strings.Builder

	for word := range strings.FieldsSeq(s) {
		if line.Len() > 0 {
			if font.TextWidth(line.String()+" "+word, fontName, fontSize) <= maxWidth {
				line.WriteString(" " + word)

				continue
			}

			lines = append(lines, line.String())
			line.Reset()
		}

		if font.TextWidth(word, fontName, fontSize) <= maxWidth {
			line.WriteString(word)

			continue
		}

		broken := breakWord(word, fontName, fontSize, maxWidth)
		lines = append(lines, broken[:len(broken)-1]...)
		line.WriteString(broken[len(broken)-1])
	}

	if line.Len() > 0 {
		lines = append(lines, line.String())
	}

	return strings.Join(lines, "\n")
}

// breakWord splits a single word into lines at character boundaries when it exceeds maxWidth
func breakWord(word, fontName string, fontSize int, maxWidth float64) []string {
	var lines []string

	var chunk strings.Builder

	for _, r := range word {
		test := chunk.String() + string(r)
		if font.TextWidth(test, fontName, fontSize) > maxWidth && chunk.Len() > 0 {
			lines = append(lines, chunk.String())
			chunk.Reset()
		}

		chunk.WriteRune(r)
	}

	if chunk.Len() > 0 {
		lines = append(lines, chunk.String())
	}

	return lines
}

// formatAttestTimestamp formats an RFC3339 timestamp into a human-readable form
func formatAttestTimestamp(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}

	return t.Format("January 2, 2006 3:04 PM UTC")
}
