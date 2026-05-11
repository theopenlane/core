package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"

	storagetypes "github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/objects/upload"
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
	AttestedPDF    []byte
	TrustCenterURL string
	OrgName        string
	TemplateFileID string
}

// attestNDADocument performs the full NDA attestation flow: downloads the original PDF,
// appends an attestation certificate, uploads the result, and resolves trust center metadata
func attestNDADocument(ctx context.Context, client *generated.Client, docData *generated.DocumentData, templateID, trustCenterID string) (*ndaAttestationResult, error) {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	docTemplate, err := client.Template.Query().Where(template.ID(templateID)).Only(allowCtx)
	if err != nil {
		return nil, ErrFailedToFetchNDATemplate
	}

	fileCtx := proxy.WithPresignInterceptorBypass(allowCtx)

	files, err := docTemplate.QueryFiles().All(fileCtx)
	if err != nil {
		return nil, ErrFailedToFetchNDATemplateFiles
	}

	if len(files) == 0 {
		return nil, ErrMissingNDATemplateFile
	}

	templateFile := files[0]

	dataBytes, err := json.Marshal(docData.Data)
	if err != nil {
		return nil, ErrFailedToMarshalDocumentData
	}

	var ndaMetadata signedNDADocumentData
	if err := json.Unmarshal(dataBytes, &ndaMetadata); err != nil {
		return nil, ErrFailedToUnmarshalNDAMetadata
	}

	storageFile := storageFileFromEnt(templateFile)

	downloaded, err := client.ObjectManager.Download(ctx, nil, storageFile, nil)
	if err != nil {
		return nil, ErrFailedToDownloadNDAPDF
	}

	attestedPDF, err := appendAttestationPage(bytes.NewReader(downloaded.File), &ndaMetadata)
	if err != nil {
		return nil, ErrFailedToCreateAttestedPDF
	}

	if err := uploadAttestedPDF(allowCtx, client, attestedPDF, docData.ID, templateFile.ID); err != nil {
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
		AttestedPDF:    attestedPDF,
		TrustCenterURL: tcURL,
		OrgName:        orgName,
		TemplateFileID: templateFile.ID,
	}, nil
}

// uploadAttestedPDF uploads the attested PDF to storage and associates it with the document data
func uploadAttestedPDF(ctx context.Context, client *generated.Client, attestedPDF []byte, docDataID, templateFileID string) error {
	fileName := fmt.Sprintf("attested_%s", templateFileID)

	file := pkgobjects.File{
		RawFile:              bytes.NewReader(attestedPDF),
		OriginalName:         fileName,
		FieldName:            "documentDataFile",
		CorrelatedObjectID:   docDataID,
		CorrelatedObjectType: "DocumentData",
		FileMetadata: pkgobjects.FileMetadata{
			ContentType: "application/pdf",
			Size:        int64(len(attestedPDF)),
			Key:         "documentDataFile",
		},
	}

	_, uploadedFiles, err := upload.HandleUploads(ctx, client.ObjectManager, []pkgobjects.File{file})
	if err != nil {
		return ErrFailedToUploadAttestedPDF
	}

	if len(uploadedFiles) == 0 {
		return ErrNoUploadedFiles
	}

	if err := client.DocumentData.UpdateOneID(docDataID).
		AddFileIDs(uploadedFiles[0].ID).
		Exec(ctx); err != nil {

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
		return nil, ErrFailedToGenerateAttestationPDF
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

// storageFileFromEnt converts an ent File entity to a storage types File
func storageFileFromEnt(file *generated.File) *storagetypes.File {
	return &storagetypes.File{
		ID:           file.ID,
		OriginalName: file.ProvidedFileName,
		FileMetadata: storagetypes.FileMetadata{
			Key:          file.StoragePath,
			Bucket:       file.StorageVolume,
			Region:       file.StorageRegion,
			ContentType:  file.DetectedContentType,
			Size:         file.PersistedFileSize,
			ProviderType: storagetypes.ProviderType(file.StorageProvider),
			FullURI:      file.URI,
		},
	}
}
