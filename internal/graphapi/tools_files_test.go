package graphapi_test

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/pkg/objects/storage"
)

const (
	logoFilePath = "testdata/uploads/logo.png"
	pdfFilePath  = "testdata/uploads/hello.pdf"
	txtFilePath  = "testdata/uploads/hello.txt"
)

// logoFileFunc creates a graphql.Upload func for the logo.png test file
func logoFileFunc(t *testing.T) func() *graphql.Upload {
	return func() *graphql.Upload {
		return uploadFile(t, logoFilePath)
	}
}

// uploadFileFunc creates a graphql.Upload func for the specified file path
func uploadFileFunc(t *testing.T, path string) func() *graphql.Upload {
	return func() *graphql.Upload {
		return uploadFile(t, path)
	}
}

// uploadFile creates a graphql.Upload for the specified file path
func uploadFile(t *testing.T, path string) *graphql.Upload {
	pdfFile, err := storage.NewUploadFile(path)
	assert.NilError(t, err)
	return &graphql.Upload{
		File:        pdfFile.RawFile,
		Filename:    pdfFile.OriginalName,
		Size:        pdfFile.Size,
		ContentType: pdfFile.ContentType,
	}
}

func getMD5Hash(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	assert.NilError(t, err)

	sum := md5.Sum(data)

	return hex.EncodeToString(sum[:])
}
