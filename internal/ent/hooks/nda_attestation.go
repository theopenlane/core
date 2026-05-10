package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/pkg/logx"
)

type signedNDADocumentData struct {
	PDFFileID         string               `json:"pdf_file_id"`
	Acknowledgment    bool                 `json:"acknowledgment"`
	SignatoryInfo     signatoryInformation `json:"signatory_info"`
	TrustCenterID     string               `json:"trust_center_id"`
	SignatureMetadata signatureMetadata    `json:"signature_metadata"`
}

type signatoryInformation struct {
	Email       string `json:"email"`
	LastName    string `json:"last_name"`
	FirstName   string `json:"first_name"`
	CompanyName string `json:"company_name"`
}

type signatureMetadata struct {
	UserID    string `json:"user_id"`
	PDFHash   string `json:"pdf_hash"`
	Timestamp string `json:"timestamp"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}

// ndaAttestationResult holds the output of the NDA attestation process
type ndaAttestationResult struct {
	AttestedPDF    []byte
	TrustCenterURL string
	OrgName        string
}

// attestNDADocument performs the full NDA attestation flow: downloads the original PDF,
// appends an attestation certificate, uploads the result, and resolves trust center metadata
func attestNDADocument(ctx context.Context, client *generated.Client, docData *generated.DocumentData, templateID, trustCenterID string) (*ndaAttestationResult, error) {
	logger := logx.FromContext(ctx)
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	docTemplate, err := client.Template.Query().
		Where(template.ID(templateID)).
		WithFiles().
		Only(allowCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch NDA template: %w", err)
	}

	if len(docTemplate.Edges.Files) == 0 {
		return nil, ErrMissingNDATemplateFile
	}

	templateFile := docTemplate.Edges.Files[0]
	if templateFile.PresignedURL == "" {
		return nil, ErrMissingNDATemplateFile
	}

	dataBytes, err := json.Marshal(docData.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal document data: %w", err)
	}

	var ndaMetadata signedNDADocumentData
	if err := json.Unmarshal(dataBytes, &ndaMetadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal NDA metadata: %w", err)
	}

	originalPDF, err := downloadFileContent(ctx, templateFile.PresignedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download original NDA PDF: %w", err)
	}

	attestedPDF, err := appendAttestationPage(bytes.NewReader(originalPDF), &ndaMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create attested PDF: %w", err)
	}

	tc, err := client.TrustCenter.Query().
		Where(trustcenter.IDEQ(trustCenterID)).
		WithSetting().
		WithCustomDomain().
		Only(allowCtx)
	if err != nil {
		logger.Error().Err(err).Str("trust_center_id", trustCenterID).Msg("failed to fetch trust center")
		return nil, fmt.Errorf("failed to fetch trust center: %w", err)
	}

	orgName := resolveOrgName(tc)

	var customDomain string
	if tc.Edges.CustomDomain != nil {
		customDomain = tc.Edges.CustomDomain.CnameRecord
	}

	tcURL := buildTrustCenterURL(customDomain, tc.Slug)

	return &ndaAttestationResult{
		AttestedPDF:    attestedPDF,
		TrustCenterURL: tcURL,
		OrgName:        orgName,
	}, nil
}

// resolveOrgName extracts the organization name from a trust center
func resolveOrgName(tc *generated.TrustCenter) string {
	if tc.Edges.Setting != nil && tc.Edges.Setting.CompanyName != "" {
		return tc.Edges.Setting.CompanyName
	}

	return ""
}

// appendAttestationPage merges the original PDF with a generated attestation certificate page
func appendAttestationPage(originalPDF io.ReadSeeker, data *signedNDADocumentData) ([]byte, error) {
	certPage, err := createAttestationCertificate(data)
	if err != nil {
		return nil, fmt.Errorf("failed to create attestation certificate: %w", err)
	}

	var buf bytes.Buffer

	readers := []io.ReadSeeker{originalPDF, bytes.NewReader(certPage)}

	if err = api.MergeRaw(readers, &buf, false, nil); err != nil {
		return nil, fmt.Errorf("failed to merge attestation page: %w", err)
	}

	return buf.Bytes(), nil
}

const (
	attestFontSize     = 18
	attestFieldSize    = 11
	attestLabelWidth   = 40
	attestFieldHeight  = 8
	attestLineSpacing  = 10
	attestTitleSpacing = 20
)

// createAttestationCertificate generates a single-page PDF with the signature certification details
func createAttestationCertificate(data *signedNDADocumentData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", attestFontSize)
	pdf.Cell(0, attestFieldHeight+2, "Signature Certification")
	pdf.Ln(attestTitleSpacing)

	fields := []struct{ label, value string }{
		{"Name", data.SignatoryInfo.FirstName + " " + data.SignatoryInfo.LastName},
		{"Email", data.SignatoryInfo.Email},
		{"Company", data.SignatoryInfo.CompanyName},
		{"Timestamp", formatAttestTimestamp(data.SignatureMetadata.Timestamp)},
		{"IP Address", data.SignatureMetadata.IPAddress},
		{"Browser", data.SignatureMetadata.UserAgent},
	}

	for _, f := range fields {
		pdf.SetFont("Helvetica", "B", attestFieldSize)
		pdf.Cell(attestLabelWidth, attestFieldHeight, f.label+":")
		pdf.SetFont("Helvetica", "", attestFieldSize)
		pdf.Cell(0, attestFieldHeight, f.value)
		pdf.Ln(attestLineSpacing)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate attestation PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// formatAttestTimestamp formats an RFC3339 timestamp into a human-readable form
func formatAttestTimestamp(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}

	return t.Format("January 2, 2006 3:04 PM UTC")
}

// downloadFileContent downloads a file from the given URL and returns its bytes
func downloadFileContent(ctx context.Context, fileURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
