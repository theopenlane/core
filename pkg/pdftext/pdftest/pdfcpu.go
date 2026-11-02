package pdftest

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/stretchr/testify/require"
)

// minimalFontSize is the font size of the text MinimalPDF writes
const minimalFontSize = 12

// MinimalPDF returns a single page pdf with one line of text
func MinimalPDF(t *testing.T) []byte {
	t.Helper()

	page := map[string]any{
		"paper":  "A4P",
		"origin": "UpperLeft",
		"fonts":  map[string]any{"f": map[string]any{"name": "Helvetica", "size": minimalFontSize}},
		"pages": map[string]any{
			"1": map[string]any{"content": map[string]any{
				"text": []map[string]any{{"value": "Original NDA", "pos": [2]float64{20, 20}, "font": map[string]any{"name": "$f"}}},
			}},
		},
	}

	jsonData, err := json.Marshal(page)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, api.Create(t.Context(), nil, bytes.NewReader(jsonData), &buf, nil))

	return buf.Bytes()
}

// EncryptPDF returns pdf encrypted with AES-256, an owner password, and the given user password
func EncryptPDF(t *testing.T, pdf []byte, userPW string) []byte {
	t.Helper()

	conf := model.NewDefaultConfiguration()
	conf.OwnerPW = "owner-only"
	conf.UserPW = userPW
	conf.EncryptUsingAES = true
	conf.EncryptKeyLength = 256
	conf.Permissions = model.PermissionsNone

	var buf bytes.Buffer
	require.NoError(t, api.Encrypt(t.Context(), bytes.NewReader(pdf), &buf, conf))

	return buf.Bytes()
}
