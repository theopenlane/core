package validators

import (
	"errors"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/objects"
	pkgobjects "github.com/theopenlane/core/v2/pkg/objects"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

// NDAValidator requires trust center NDAs to be PDFs that open without a password, since the attestation merge cannot decrypt them
var NDAValidator storage.ValidationFunc = ndaValidator

func ndaValidator(f storage.File) error {
	if !isNDAUpload(f) {
		return nil
	}

	if err := pkgobjects.MimeTypeValidator(pdfMimeType)(f); err != nil {
		return err
	}

	if f.RawFile == nil {
		return nil
	}

	_, err := api.ReadContext(f.RawFile, model.NewDefaultConfiguration())
	if errors.Is(err, pdfcpu.ErrWrongPassword) {
		log.Warn().Str("file", f.OriginalName).Msg("rejected password protected nda upload")

		return ErrPasswordProtectedNDA
	}

	return nil
}

// isNDAUpload reports whether the upload carries the trust center NDA template kind hint
func isNDAUpload(f storage.File) bool {
	if f.ProviderHints == nil || f.ProviderHints.Metadata == nil {
		return false
	}

	return f.ProviderHints.Metadata[objects.TemplateKindMetadataKey] == enums.TemplateKindTrustCenterNda.String()
}
