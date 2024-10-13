package objects

import (
	"bytes"
	"crypto/md5" //nolint:gosec // MD5 is used for checksums, not for hashing passwords
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/gabriel-vasile/mimetype"
	"github.com/rs/zerolog/log"
)

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
	hash := md5.New() //nolint:gosec // MD5 is used for checksums, not for hashing passwords
	if _, err := io.Copy(hash, data); err != nil {
		return "", fmt.Errorf("could not read file: %w", err)
	}

	if _, err := data.Seek(0, io.SeekStart); err != nil { // seek back to beginning of file
		return "", fmt.Errorf("could not seek to beginning of file: %w", err)
	}

	return base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}

// DetectContentType leverages http.DetectContentType to identify the content type
// of the provided data. It expects that the passed io object will be seeked to its
// beginning and will seek back to the beginning after reading its content.
func DetectContentType(data io.ReadSeeker) (string, error) {
	if _, err := data.Seek(0, io.SeekStart); err != nil { // seek back to beginning of file
		return "", fmt.Errorf("could not seek to beginning of file: %w", err)
	}

	// the default return value will default to application/octet-stream if unable to detect the MIME type
	contentType, readErr := mimetype.DetectReader(data)
	if readErr != nil {
		return "", fmt.Errorf("encountered error reading file content type: %w", readErr)
	}

	if _, err := data.Seek(0, io.SeekStart); err != nil { // seek back to beginning of file
		return "", fmt.Errorf("could not seek to beginning of file: %w", err)
	}

	return contentType.String(), nil
}
