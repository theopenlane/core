package validators

import (
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

	if isPasswordProtected(f) {
		log.Warn().Str("file", f.OriginalName).Msg("rejected password protected nda upload")

		return ErrPasswordProtectedPDF
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
