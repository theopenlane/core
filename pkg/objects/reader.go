package objects

import (
	"bytes"
	"io"
	"io/fs"
)

const (
	// MaxInMemorySize is the maximum file size we'll buffer in memory (10MB)
	MaxInMemorySize = 10 * 1024 * 1024
)

// BufferedReader wraps file data and provides both Reader and ReadSeeker interfaces
type BufferedReader struct {
	data   []byte
	reader *bytes.Reader
	size   int64
}

// NewBufferedReader creates a BufferedReader from raw data
func NewBufferedReader(data []byte) *BufferedReader {
	return &BufferedReader{
		data:   data,
		reader: bytes.NewReader(data),
		size:   int64(len(data)),
	}
}

// NewBufferedReaderFromReader creates a BufferedReader from an io.Reader
// This is the robust method for handling inbound file data that can work with all providers
// It buffers files up to MaxInMemorySize - if the file exceeds this, it returns an error
// indicating the caller should use disk-based buffering instead
func NewBufferedReaderFromReader(r io.Reader) (*BufferedReader, error) {
	return NewBufferedReaderFromReaderWithLimit(r, MaxInMemorySize)
}

// NewBufferedReaderFromReaderWithLimit creates a BufferedReader from an io.Reader with a size limit
func NewBufferedReaderFromReaderWithLimit(r io.Reader, maxSize int64) (*BufferedReader, error) {
	if r == nil {
		return nil, ErrReaderCannotBeNil
	}

	if maxSize <= 0 {
		maxSize = MaxInMemorySize
	}

	// Use LimitReader to prevent reading more than maxSize
	limitedReader := io.LimitReader(r, maxSize)

	// Read all data into memory buffer
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, ErrFailedToReadData
	}

	// Check if we hit the limit (incomplete read)
	if int64(len(data)) == maxSize {
		// Try to read one more byte to see if there's more data
		var extraByte [1]byte

		n, _ := r.Read(extraByte[:])
		if n > 0 {
			return nil, ErrFileSizeExceedsLimit
		}
	}

	return &BufferedReader{
		data:   data,
		reader: bytes.NewReader(data),
		size:   int64(len(data)),
	}, nil
}

// Read implements io.Reader
func (br *BufferedReader) Read(p []byte) (n int, err error) {
	return br.reader.Read(p)
}

// Seek implements io.Seeker
func (br *BufferedReader) Seek(offset int64, whence int) (int64, error) {
	return br.reader.Seek(offset, whence)
}

// Close implements io.Closer (no-op for memory buffer)
func (br *BufferedReader) Close() error {
	return nil
}

// Size returns the total size of the buffered data
func (br *BufferedReader) Size() int64 {
	return br.size
}

// Data returns a copy of the underlying data
func (br *BufferedReader) Data() []byte {
	result := make([]byte, len(br.data))
	copy(result, br.data)

	return result
}

// Reset resets the reader to the beginning
func (br *BufferedReader) Reset() {
	_, _ = br.reader.Seek(0, io.SeekStart)
}

// NewReader creates a new independent reader from the buffered data
func (br *BufferedReader) NewReader() io.Reader {
	return bytes.NewReader(br.data)
}

// NewReadSeeker creates a new independent ReadSeeker from the buffered data
func (br *BufferedReader) NewReadSeeker() io.ReadSeeker {
	return bytes.NewReader(br.data)
}

// SizedReader describes readers that can report their size without consuming the stream.
type SizedReader interface {
	Size() int64
}

// LenReader describes readers that expose remaining length semantics (e.g. *bytes.Reader).
type LenReader interface {
	Len() int
}

// StatReader describes readers backed by file descriptors that can return stat information.
type StatReader interface {
	Stat() (fs.FileInfo, error)
}

// InferReaderSize attempts to determine the total size of the provided reader without
// modifying its current position. It returns the reported size and true when available.
func InferReaderSize(r io.Reader) (int64, bool) {
	switch v := r.(type) {
	case SizedReader:
		return v.Size(), true
	case LenReader:
		return int64(v.Len()), true
	case StatReader:
		info, err := v.Stat()
		if err != nil {
			return 0, false
		}

		return info.Size(), true
	default:
		return 0, false
	}
}
