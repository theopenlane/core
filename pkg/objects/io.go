package objects

import (
	"crypto/md5"
	"encoding/base64"
	"io"
	"os"

	"github.com/gabriel-vasile/mimetype"
	"github.com/pkg/errors"
)

// ReaderToSeeker function takes an io.Reader as input and returns an io.ReadSeeker which can be used to upload files to the object storage
func ReaderToSeeker(r io.Reader) (io.ReadSeeker, error) {
	tmpfile, err := os.CreateTemp("", "upload-")
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(tmpfile, r)
	if err != nil {
		_ = tmpfile.Close()
		_ = os.Remove(tmpfile.Name())

		return nil, err
	}

	_, err = tmpfile.Seek(0, 0)
	if err != nil {
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
	hash := md5.New()
	if _, err := io.Copy(hash, data); err != nil {
		return "", errors.Wrap(err, "could not read file")
	}

	if _, err := data.Seek(0, io.SeekStart); err != nil { // seek back to beginning of file
		return "", errors.Wrap(err, "could not seek to beginning of file")
	}

	return base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}

// DetectContentType leverages http.DetectContentType to identify the content type
// of the provided data. It expects that the passed io object will be seeked to its
// beginning and will seek back to the beginning after reading its content.
func DetectContentType(data io.ReadSeeker) (string, error) {
	if _, err := data.Seek(0, io.SeekStart); err != nil { // seek back to beginning of file
		return "", errors.Wrap(err, "could not seek to beginning of file")
	}

	// the default return value will default to application/octet-stream if unable to detect the MIME type
	contentType, readErr := mimetype.DetectReader(data)
	if readErr != nil {
		return "", errors.Wrap(readErr, "encountered error reading file content type")
	}

	if _, err := data.Seek(0, io.SeekStart); err != nil { // seek back to beginning of file
		return "", errors.Wrap(err, "could not seek to beginning of file")
	}

	return contentType.String(), nil
}
