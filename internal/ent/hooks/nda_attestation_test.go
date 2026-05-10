package hooks

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-pdf/fpdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/generated"
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

	pageCount, err := api.PageCount(bytes.NewReader(pdfBytes), nil)
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

	err = api.Validate(bytes.NewReader(pdfBytes), nil)
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

	pageCount, err := api.PageCount(bytes.NewReader(merged), nil)
	require.NoError(t, err)
	assert.Equal(t, 2, pageCount)

	err = api.Validate(bytes.NewReader(merged), nil)
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

func TestDownloadFileContent(t *testing.T) {
	expected := []byte("fake-pdf-content")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(expected)
	}))
	defer srv.Close()

	data, err := downloadFileContent(context.Background(), srv.URL)
	require.NoError(t, err)
	assert.Equal(t, expected, data)
}

func TestDownloadFileContent_NonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := downloadFileContent(context.Background(), srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
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

// generateMinimalPDF creates a valid single-page PDF for use as the original NDA document in merge tests
func generateMinimalPDF(t *testing.T) []byte {
	t.Helper()

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 12) //nolint:mnd
	pdf.Cell(0, 10, "Original NDA")  //nolint:mnd

	var buf bytes.Buffer
	require.NoError(t, pdf.Output(&buf))

	return buf.Bytes()
}
