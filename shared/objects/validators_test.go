package objects

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMimeTypeValidator(t *testing.T) {
	tests := []struct {
		name           string
		validMimeTypes []string
		file           File
		expectError    bool
		errorContains  string
	}{
		{
			name:           "valid mime type - exact match",
			validMimeTypes: []string{"image/png", "image/jpeg"},
			file: File{
				FileMetadata: FileMetadata{
					ContentType: "image/png",
				},
			},
			expectError: false,
		},
		{
			name:           "valid mime type - case insensitive",
			validMimeTypes: []string{"image/PNG", "image/JPEG"},
			file: File{
				FileMetadata: FileMetadata{
					ContentType: "image/png",
				},
			},
			expectError: false,
		},
		{
			name:           "valid mime type - multiple types",
			validMimeTypes: []string{"application/pdf", "image/png", "text/plain"},
			file: File{
				FileMetadata: FileMetadata{
					ContentType: "text/plain",
				},
			},
			expectError: false,
		},
		{
			name:           "invalid mime type",
			validMimeTypes: []string{"image/png", "image/jpeg"},
			file: File{
				FileMetadata: FileMetadata{
					ContentType: "application/pdf",
				},
			},
			expectError:   true,
			errorContains: "unsupported mime type",
		},
		{
			name:           "empty mime type list",
			validMimeTypes: []string{},
			file: File{
				FileMetadata: FileMetadata{
					ContentType: "image/png",
				},
			},
			expectError:   true,
			errorContains: "unsupported mime type",
		},
		{
			name:           "empty content type in file",
			validMimeTypes: []string{"image/png"},
			file: File{
				FileMetadata: FileMetadata{
					ContentType: "",
				},
			},
			expectError:   true,
			errorContains: "unsupported mime type",
		},
		{
			name:           "case insensitive validation",
			validMimeTypes: []string{"Image/PNG"},
			file: File{
				FileMetadata: FileMetadata{
					ContentType: "IMAGE/png",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := MimeTypeValidator(tt.validMimeTypes...)
			err := validator(tt.file)

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrUnsupportedMimeType)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChainValidators(t *testing.T) {
	t.Run("all validators pass", func(t *testing.T) {
		file := File{
			FileMetadata: FileMetadata{
				ContentType: "image/png",
				Size:        1024,
			},
		}

		validator1 := MimeTypeValidator("image/png", "image/jpeg")
		validator2 := func(f File) error {
			if f.Size > 2048 {
				return errors.New("file too large")
			}
			return nil
		}

		chainedValidator := ChainValidators(validator1, validator2)
		err := chainedValidator(file)

		assert.NoError(t, err)
	})

	t.Run("first validator fails", func(t *testing.T) {
		file := File{
			FileMetadata: FileMetadata{
				ContentType: "application/pdf",
				Size:        1024,
			},
		}

		validator1 := MimeTypeValidator("image/png", "image/jpeg")
		validator2 := func(f File) error {
			return nil
		}

		chainedValidator := ChainValidators(validator1, validator2)
		err := chainedValidator(file)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrUnsupportedMimeType)
	})

	t.Run("second validator fails", func(t *testing.T) {
		file := File{
			FileMetadata: FileMetadata{
				ContentType: "image/png",
				Size:        3000,
			},
		}

		validator1 := MimeTypeValidator("image/png", "image/jpeg")
		validator2 := func(f File) error {
			if f.Size > 2048 {
				return errors.New("file too large")
			}
			return nil
		}

		chainedValidator := ChainValidators(validator1, validator2)
		err := chainedValidator(file)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "file too large")
	})

	t.Run("empty validator chain", func(t *testing.T) {
		file := File{
			FileMetadata: FileMetadata{
				ContentType: "image/png",
			},
		}

		chainedValidator := ChainValidators()
		err := chainedValidator(file)

		assert.NoError(t, err)
	})

	t.Run("single validator", func(t *testing.T) {
		file := File{
			FileMetadata: FileMetadata{
				ContentType: "image/png",
			},
		}

		validator := MimeTypeValidator("image/png")
		chainedValidator := ChainValidators(validator)
		err := chainedValidator(file)

		assert.NoError(t, err)
	})

	t.Run("multiple validators with custom logic", func(t *testing.T) {
		file := File{
			FileMetadata: FileMetadata{
				ContentType: "image/png",
				Size:        1024,
			},
			OriginalName: "test.png",
		}

		validator1 := MimeTypeValidator("image/png")
		validator2 := func(f File) error {
			if f.Size == 0 {
				return errors.New("file size cannot be zero")
			}
			return nil
		}
		validator3 := func(f File) error {
			if f.OriginalName == "" {
				return errors.New("file name cannot be empty")
			}
			return nil
		}

		chainedValidator := ChainValidators(validator1, validator2, validator3)
		err := chainedValidator(file)

		assert.NoError(t, err)
	})

	t.Run("stops at first error", func(t *testing.T) {
		file := File{
			FileMetadata: FileMetadata{
				ContentType: "application/pdf",
				Size:        0,
			},
		}

		validationOrder := []string{}

		validator1 := func(f File) error {
			validationOrder = append(validationOrder, "validator1")
			return errors.New("first error")
		}
		validator2 := func(f File) error {
			validationOrder = append(validationOrder, "validator2")
			return errors.New("second error")
		}

		chainedValidator := ChainValidators(validator1, validator2)
		err := chainedValidator(file)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "first error")
		assert.Equal(t, []string{"validator1"}, validationOrder)
	})
}
