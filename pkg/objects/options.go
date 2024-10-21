package objects

import (
	"encoding/json"
	"io"

	"github.com/spf13/cast"
	"gopkg.in/cheggaaa/pb.v2"
)

// Option is a function that configures the Objects
type Option func(*Objects)

// WithStorage allows you to provide a storage backend to the Objects
func WithStorage(store Storage) Option {
	return func(o *Objects) {
		o.Storage = store
	}
}

// WithMaxFileSize allows you limit the size of file uploads to accept
func WithMaxFileSize(i int64) Option {
	return func(o *Objects) {
		o.MaxSize = i
	}
}

// WithMaxMemory allows you limit the amount of memory to use when parsing a multipart form
func WithMaxMemory(i int64) Option {
	return func(o *Objects) {
		o.MaxMemory = i
	}
}

// WithIgnoreNonExistentKey allows you to configure the handler to skip multipart form key values which do not match the configured
func WithIgnoreNonExistentKey(ignore bool) Option {
	return func(o *Objects) {
		o.IgnoreNonExistentKeys = ignore
	}
}

// WithValidationFunc allows you to provide a custom validation function
func WithValidationFunc(validationFunc ValidationFunc) Option {
	return func(o *Objects) {
		o.ValidationFunc = validationFunc
	}
}

// WithKeys allows you to configure the keys to look for in the multipart form on the REST request
func WithKeys(keys []string) Option {
	return func(o *Objects) {
		o.Keys = keys
	}
}

// WithNameFuncGenerator allows you configure how you'd like to rename your uploaded files
func WithNameFuncGenerator(nameFunc NameGeneratorFunc) Option {
	return func(o *Objects) {
		o.NameFuncGenerator = nameFunc
	}
}

// WithUploaderFunc allows you to provide a custom uploader function
func WithUploaderFunc(uploader UploaderFunc) Option {
	return func(o *Objects) {
		o.Uploader = uploader
	}
}

// WithSkipper allows you to provide a custom skipper function
func WithSkipper(skipper SkipperFunc) Option {
	return func(o *Objects) {
		o.Skipper = skipper
	}
}

// WithErrorResponseHandler allows you to provide a custom error response handler
func WithErrorResponseHandler(errHandler ErrResponseHandler) Option {
	return func(o *Objects) {
		o.ErrorResponseHandler = errHandler
	}
}

// WithUploadFileOptions allows you to provide options for uploading a file
func WithUploadFileOptions(opts *UploadFileOptions) Option {
	return func(o *Objects) {
		o.UploadFileOptions = opts
	}
}

// WithDownloadFileOptions allows you to provide options for downloading a file
func WithDownloadFileOptions(opts *DownloadFileOptions) Option {
	return func(o *Objects) {
		o.DownloadFileOptions = opts
	}
}

func WithMetdata(mp map[string]interface{}) Option {
	return func(o *Objects) {
		if o.UploadFileOptions.Metadata == nil {
			o.UploadFileOptions.Metadata = map[string]string{}
		}

		for k, v := range mp {
			s, err := cast.ToStringE(v)
			if err == nil {
				o.UploadFileOptions.Metadata[k] = s
				continue
			}

			bts, err := json.Marshal(v)
			if err == nil {
				o.UploadFileOptions.Metadata[k] = string(bts)
				continue
			}

			o.UploadFileOptions.Metadata[k] = "<<INVALID_METADATA>>"
		}
	}
}

// UploadOption is a function that configures the UploadFileOptions
type UploadOption func(*UploadFileOptions)

// UploadProgress allows you to provide a progress bar for the upload
func UploadProgress(p *pb.ProgressBar) UploadOption {
	return func(opts *UploadFileOptions) {
		opts.Progress = p
	}
}

// UploadProgressOutput allows you to provide a writer for the progress bar
func UploadProgressOutput(out io.Writer) UploadOption {
	return func(opts *UploadFileOptions) {
		opts.ProgressOutput = out
	}
}

// UploadProgressFinishMessage allows you to provide a message to display when the upload is complete
func UploadProgressFinishMessage(s string) UploadOption {
	return func(opts *UploadFileOptions) {
		opts.ProgressFinishMessage = s
	}
}

// UploadMetadata allows you to provide metadata for the upload
func UploadMetadata(mp map[string]interface{}) UploadOption {
	return func(opts *UploadFileOptions) {
		if opts.Metadata == nil {
			opts.Metadata = map[string]string{}
		}

		for k, v := range mp {
			s, err := cast.ToStringE(v)
			if err == nil {
				opts.Metadata[k] = s
				continue
			}

			bts, err := json.Marshal(v)
			if err == nil {
				opts.Metadata[k] = string(bts)
				continue
			}

			opts.Metadata[k] = "<<INVALID_METADATA>>"
		}
	}
}
