package hooks

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	storagetypes "github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/interceptors"
)

var (
	conf = model.NewDefaultConfiguration()
)

func TestCreateAttestationCertificate(t *testing.T) {
	data := &signedNDADocumentData{
		SignatoryInfo: signatoryInformation{
			FirstName:   "Jane",
			LastName:    "Doe",
			Email:       "jane@example.com",
			CompanyName: "Acme Corp",
		},
		SignatureMetadata: signatureMetadata{
			Timestamp: "2025-06-15T14:30:00Z",
			IPAddress: "192.168.1.1",
			UserAgent: "Mozilla/5.0",
		},
	}

	pdfBytes, err := createAttestationCertificate(data)
	require.NoError(t, err)
	require.NotEmpty(t, pdfBytes)

	assert.Equal(t, "%PDF", string(pdfBytes[:4]))

	pageCount, err := api.PageCount(bytes.NewReader(pdfBytes), conf)
	require.NoError(t, err)
	assert.Equal(t, 1, pageCount)

	err = api.Validate(bytes.NewReader(pdfBytes), nil)
	assert.NoError(t, err)
}

func TestCreateAttestationCertificate_EmptyFields(t *testing.T) {
	data := &signedNDADocumentData{}

	pdfBytes, err := createAttestationCertificate(data)
	require.NoError(t, err)
	require.NotEmpty(t, pdfBytes)

	err = api.Validate(bytes.NewReader(pdfBytes), conf)
	assert.NoError(t, err)
}

func TestAppendAttestationPage(t *testing.T) {
	originalPDF := generateMinimalPDF(t)

	data := &signedNDADocumentData{
		SignatoryInfo: signatoryInformation{
			FirstName:   "John",
			LastName:    "Smith",
			Email:       "john@example.com",
			CompanyName: "Test Inc",
		},
		SignatureMetadata: signatureMetadata{
			Timestamp: "2025-06-15T10:00:00Z",
			IPAddress: "10.0.0.1",
			UserAgent: "TestAgent/1.0",
		},
	}

	merged, err := appendAttestationPage(bytes.NewReader(originalPDF), data)
	require.NoError(t, err)
	require.NotEmpty(t, merged)

	assert.Equal(t, "%PDF", string(merged[:4]))

	pageCount, err := api.PageCount(bytes.NewReader(merged), conf)
	require.NoError(t, err)
	assert.Equal(t, 2, pageCount)

	err = api.Validate(bytes.NewReader(merged), conf)
	assert.NoError(t, err)
}

func TestFormatAttestTimestamp(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "valid RFC3339",
			input:  "2025-06-15T14:30:00Z",
			expect: "June 15, 2025 2:30 PM UTC",
		},
		{
			name:   "valid with offset",
			input:  "2025-01-02T03:04:05+00:00",
			expect: "January 2, 2025 3:04 AM UTC",
		},
		{
			name:   "invalid falls back to raw string",
			input:  "not-a-date",
			expect: "not-a-date",
		},
		{
			name:   "empty string",
			input:  "",
			expect: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, formatAttestTimestamp(tc.input))
		})
	}
}

func TestResolveOrgName(t *testing.T) {
	t.Run("from setting", func(t *testing.T) {
		tc := &generated.TrustCenter{}
		tc.Edges.Setting = &generated.TrustCenterSetting{CompanyName: "Acme"}

		assert.Equal(t, "Acme", resolveOrgName(tc))
	})

	t.Run("empty when no setting", func(t *testing.T) {
		tc := &generated.TrustCenter{}

		assert.Equal(t, "", resolveOrgName(tc))
	})

	t.Run("empty when company name blank", func(t *testing.T) {
		tc := &generated.TrustCenter{}
		tc.Edges.Setting = &generated.TrustCenterSetting{}

		assert.Equal(t, "", resolveOrgName(tc))
	})
}

func TestStorageFileFromEnt(t *testing.T) {
	entFile := &generated.File{
		ID:                  "file-123",
		ProvidedFileName:    "nda_template.pdf",
		StoragePath:         "org/files/nda_template.pdf",
		StorageVolume:       "documents",
		StorageRegion:       "us-east-1",
		DetectedContentType: "application/pdf",
		PersistedFileSize:   42000,
		StorageProvider:     "s3",
		URI:                 "s3://documents/org/files/nda_template.pdf",
	}

	result := interceptors.StorageFileFromEnt(entFile)

	assert.Equal(t, entFile.ID, result.ID)
	assert.Equal(t, entFile.ProvidedFileName, result.OriginalName)
	assert.Equal(t, entFile.StoragePath, result.FileMetadata.Key)
	assert.Equal(t, entFile.StorageVolume, result.FileMetadata.Bucket)
	assert.Equal(t, entFile.StorageRegion, result.FileMetadata.Region)
	assert.Equal(t, entFile.DetectedContentType, result.FileMetadata.ContentType)
	assert.Equal(t, entFile.PersistedFileSize, result.FileMetadata.Size)
	assert.Equal(t, entFile.URI, result.FileMetadata.FullURI)
	assert.Equal(t, storagetypes.ProviderType(entFile.StorageProvider), result.FileMetadata.ProviderType)
}

func TestValidateTrustCenterNDAJSON(t *testing.T) {
	validDoc := map[string]any{
		"pdf_file_id":     "file-abc",
		"acknowledgment":  true,
		"trust_center_id": "tc-123",
		"signatory_info": map[string]any{
			"email":        "jane@example.com",
			"first_name":   "Jane",
			"last_name":    "Doe",
			"company_name": "Acme Corp",
		},
		"signature_metadata": map[string]any{
			"user_id":    "user-456",
			"pdf_hash":   "abc123hash",
			"timestamp":  "2025-06-15T14:30:00Z",
			"ip_address": "192.168.1.1",
			"user_agent": "Mozilla/5.0",
		},
	}

	t.Run("valid document passes", func(t *testing.T) {
		err := validateTrustCenterNDAJSON(validDoc, "tc-123", "jane@example.com", "user-456")
		assert.NoError(t, err)
	})

	t.Run("mismatched trust center id", func(t *testing.T) {
		err := validateTrustCenterNDAJSON(validDoc, "tc-wrong", "jane@example.com", "user-456")
		assert.ErrorIs(t, err, errDocInfoDoesNotMatchCaller)
	})

	t.Run("mismatched email", func(t *testing.T) {
		err := validateTrustCenterNDAJSON(validDoc, "tc-123", "wrong@example.com", "user-456")
		assert.ErrorIs(t, err, errDocInfoDoesNotMatchCaller)
	})

	t.Run("mismatched user id", func(t *testing.T) {
		err := validateTrustCenterNDAJSON(validDoc, "tc-123", "jane@example.com", "user-wrong")
		assert.ErrorIs(t, err, errDocInfoDoesNotMatchCaller)
	})

	t.Run("missing required field", func(t *testing.T) {
		incomplete := map[string]any{
			"pdf_file_id":     "file-abc",
			"acknowledgment":  true,
			"trust_center_id": "tc-123",
		}

		err := validateTrustCenterNDAJSON(incomplete, "tc-123", "jane@example.com", "user-456")
		assert.ErrorIs(t, err, errValidationFailed)
	})

	t.Run("empty signatory name fails schema", func(t *testing.T) {
		doc := map[string]any{
			"pdf_file_id":     "file-abc",
			"acknowledgment":  true,
			"trust_center_id": "tc-123",
			"signatory_info": map[string]any{
				"email":        "jane@example.com",
				"first_name":   "",
				"last_name":    "Doe",
				"company_name": "Acme Corp",
			},
			"signature_metadata": map[string]any{
				"user_id":    "user-456",
				"pdf_hash":   "abc123hash",
				"timestamp":  "2025-06-15T14:30:00Z",
				"ip_address": "192.168.1.1",
			},
		}

		err := validateTrustCenterNDAJSON(doc, "tc-123", "jane@example.com", "user-456")
		assert.ErrorIs(t, err, errValidationFailed)
	})

	t.Run("nil document", func(t *testing.T) {
		err := validateTrustCenterNDAJSON(nil, "tc-123", "jane@example.com", "user-456")
		assert.Error(t, err)
	})
}

func TestAttestationFieldsFrom_IncludesHash(t *testing.T) {
	data := &signedNDADocumentData{
		SignatoryInfo: signatoryInformation{
			FirstName:   "Jane",
			LastName:    "Doe",
			Email:       "jane@example.com",
			CompanyName: "Acme Corp",
		},
		SignatureMetadata: signatureMetadata{
			Timestamp: "2025-06-15T14:30:00Z",
			IPAddress: "192.168.1.1",
			UserAgent: "Mozilla/5.0",
			PDFHash:   "deadbeef1234",
		},
	}

	fields := attestationFieldsFrom(data)

	var hashField attestationField
	for _, f := range fields {
		if f.Label == "Document Hash" {
			hashField = f
			break
		}
	}

	assert.Equal(t, "deadbeef1234", hashField.Value)
}

func TestAppendAttestationPage_TwoPassHash(t *testing.T) {
	originalPDF := generateMinimalPDF(t)

	data := &signedNDADocumentData{
		SignatoryInfo: signatoryInformation{
			FirstName:   "Jane",
			LastName:    "Doe",
			Email:       "jane@example.com",
			CompanyName: "Acme Corp",
		},
		SignatureMetadata: signatureMetadata{
			Timestamp: "2025-06-15T14:30:00Z",
			IPAddress: "192.168.1.1",
			UserAgent: "Mozilla/5.0",
		},
	}

	// first pass: generate combined document to compute hash
	combined, err := appendAttestationPage(bytes.NewReader(originalPDF), data)
	require.NoError(t, err)

	pdfHash := sha256.Sum256(combined)
	computedHash := hex.EncodeToString(pdfHash[:])
	require.NotEmpty(t, computedHash)

	// second pass: assign hash and regenerate
	data.SignatureMetadata.PDFHash = computedHash

	attestedPDF, err := appendAttestationPage(bytes.NewReader(originalPDF), data)
	require.NoError(t, err)
	require.NotEmpty(t, attestedPDF)

	pageCount, err := api.PageCount(bytes.NewReader(attestedPDF), conf)
	require.NoError(t, err)
	assert.Equal(t, 2, pageCount)

	err = api.Validate(bytes.NewReader(attestedPDF), conf)
	assert.NoError(t, err)

	// attestation fields should reflect the computed hash
	fields := attestationFieldsFrom(data)
	var hashField attestationField
	for _, f := range fields {
		if f.Label == "Document Hash" {
			hashField = f
			break
		}
	}

	assert.Equal(t, computedHash, hashField.Value)
}

// generateMinimalPDF creates a valid single-page PDF for use as the original NDA document in merge tests
func generateMinimalPDF(t *testing.T) []byte {
	t.Helper()

	page := map[string]any{
		"paper":  "A4P",
		"origin": "UpperLeft",
		"fonts": map[string]any{
			"f": map[string]any{"name": "Helvetica", "size": 12},
		},
		"pages": map[string]any{
			"1": map[string]any{
				"content": map[string]any{
					"text": []map[string]any{
						{"value": "Original NDA", "pos": [2]float64{20, 20}, "font": map[string]any{"name": "$f"}},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(page)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, api.Create(nil, bytes.NewReader(jsonData), &buf, nil))

	return buf.Bytes()
}
