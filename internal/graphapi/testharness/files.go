//go:build test

package testharness

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

// test upload fixtures, resolved from the repo root so they do not depend on the
// working directory of whichever suite is running
var (
	LogoFilePath = filepath.Join(repoRoot(), "internal", "graphapi", "testdata", "uploads", "logo.png")
	PdfFilePath  = filepath.Join(repoRoot(), "internal", "graphapi", "testdata", "uploads", "hello.pdf")
	TxtFilePath  = filepath.Join(repoRoot(), "internal", "graphapi", "testdata", "uploads", "hello.txt")
)

// LogoFileFunc creates a graphql.Upload func for the logo.png test file
func LogoFileFunc(t *testing.T) func() *graphql.Upload {
	return func() *graphql.Upload {
		return UploadFile(t, LogoFilePath)
	}
}

// UploadFileFunc creates a graphql.Upload func for the specified file path
func UploadFileFunc(t *testing.T, path string) func() *graphql.Upload {
	return func() *graphql.Upload {
		return UploadFile(t, path)
	}
}

// UploadFile creates a graphql.Upload for the specified file path
func UploadFile(t *testing.T, path string) *graphql.Upload {
	pdfFile, err := storage.NewUploadFile(path)
	assert.NilError(t, err)
	return &graphql.Upload{
		File:        pdfFile.RawFile,
		Filename:    pdfFile.OriginalName,
		Size:        pdfFile.Size,
		ContentType: pdfFile.ContentType,
	}
}

func GetMD5Hash(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	assert.NilError(t, err)

	sum := md5.Sum(data)

	return hex.EncodeToString(sum[:])
}
