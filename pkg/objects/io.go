package objects

import (
	"bytes"
	"crypto/md5" //nolint:gosec  #nosec G501 // MD5 is used for checksums, not for hashing passwords
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gabriel-vasile/mimetype"
	"github.com/rs/zerolog/log"
)

var (
	seekError = "could not seek to beginning of file"
)

// NewUploadFile function reads the content of the provided file path and returns a FileUpload object
func NewUploadFile(filePath string) (file FileUpload, err error) {
	f, err := os.Open(filePath) // #nosec G304
	if err != nil {
		return file, err
	}

	defer func() {
		closeErr := f.Close()
		if err == nil {
			err = closeErr
		}
	}()

	stat, err := f.Stat()
	if err != nil {
		return file, err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return file, err
	}

	file.File = bytes.NewReader(b)
	file.Filename = stat.Name()
	file.Size = stat.Size()
	file.ContentType = http.DetectContentType(b)

	return file, nil
}

// StreamToByte function reads the content of the provided io.Reader and returns it as a byte slice
func StreamToByte(stream io.ReadSeeker) ([]byte, error) {
	buf := new(bytes.Buffer)

	if _, err := buf.ReadFrom(stream); err != nil {
		return nil, err
	}

	// reset the file to the beginning
	if _, err := stream.Seek(0, io.SeekStart); err != nil {
		log.Error().Err(err).Msg("failed to reset file")

		return nil, err
	}

	return buf.Bytes(), nil
}

// ReaderToSeeker function takes an io.Reader as input and returns an io.ReadSeeker which can be used to upload files to the object storage
func ReaderToSeeker(r io.Reader) (io.ReadSeeker, error) {
	if r == nil {
		return nil, nil
	}

	tmpfile, err := os.CreateTemp("", "upload-")
	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(tmpfile, r); err != nil {
		_ = tmpfile.Close()
		_ = os.Remove(tmpfile.Name())

		return nil, err
	}

	if _, err = tmpfile.Seek(0, 0); err != nil {
		_ = tmpfile.Close()
		_ = os.Remove(tmpfile.Name())

		return nil, err
	}

	// Return the file, which implements io.ReadSeeker which you can now pass to the objects uploader
	return tmpfile, nil
}

// ComputeChecksum calculates the MD5 checksum for the provided data. It expects that
// the passed io object will be seeked to its beginning and will seek back to the
// beginning after reading its content.
func ComputeChecksum(data io.ReadSeeker) (string, error) {
	hash := md5.New() //nolint:gosec  #nosec G501 // MD5 is used for checksums, not for hashing passwords
	if _, err := io.Copy(hash, data); err != nil {
		return "", fmt.Errorf("could not read file: %w", err)
	}

	if _, err := data.Seek(0, io.SeekStart); err != nil { // seek back to beginning of file
		return "", fmt.Errorf("%s: %w", seekError, err)
	}

	return base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}

// DetectContentType leverages http.DetectContentType to identify the content type
// of the provided data. It expects that the passed io object will be seeked to its
// beginning and will seek back to the beginning after reading its content.
func DetectContentType(data io.ReadSeeker) (string, error) {
	if _, err := data.Seek(0, io.SeekStart); err != nil { // seek back to beginning of file
		return "", fmt.Errorf("%s: %w", seekError, err)
	}

	// the default return value will default to application/octet-stream if unable to detect the MIME type
	contentType, readErr := mimetype.DetectReader(data)
	if readErr != nil {
		return "", fmt.Errorf("encountered error reading file content type: %w", readErr)
	}

	if _, err := data.Seek(0, io.SeekStart); err != nil { // seek back to beginning of file
		return "", fmt.Errorf("%s: %w", seekError, err)
	}

	return contentType.String(), nil
}
