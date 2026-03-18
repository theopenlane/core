package graphapi_test

import (
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/pkg/objects/storage"
	"gotest.tools/v3/assert"
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
