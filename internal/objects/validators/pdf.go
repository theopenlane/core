package validators

import (
	"errors"
	"slices"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

// passwordProtectedFields are the upload fields whose pdfs must open without a password, because
// the processing behind them cannot decrypt one
var passwordProtectedFields = []string{scanFilesField}

// PasswordProtectedValidator rejects password protected pdfs uploaded to the fields that require
// readable ones
var PasswordProtectedValidator storage.ValidationFunc = passwordProtectedValidator

func passwordProtectedValidator(f storage.File) error {
	if !slices.Contains(passwordProtectedFields, f.FieldName) {
		return nil
	}

	if !isPasswordProtected(f) {
		return nil
	}

	log.Warn().Str("file", f.OriginalName).Str("field", f.FieldName).Msg("rejected password protected upload")

	return ErrPasswordProtectedPDF
}

// isPasswordProtected reports whether the upload is a pdf that cannot be opened without a user
// password; a file with no reader, or one that fails to parse for any other reason, is left to
// the validator that cares about that failure
func isPasswordProtected(f storage.File) bool {
	if f.RawFile == nil {
		return false
	}

	_, err := api.ReadContext(f.RawFile, model.NewDefaultConfiguration())

	return errors.Is(err, pdfcpu.ErrWrongPassword)
}
