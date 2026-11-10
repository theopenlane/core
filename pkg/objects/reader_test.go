package objects

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestNewBufferedReader(t *testing.T) {
	data := []byte("test data")
	br := NewBufferedReader(data)

	assert.Assert(t, br != nil)
	assert.Check(t, is.Equal(int64(len(data)), br.Size()))
	assert.Check(t, is.DeepEqual(data, br.Data()))
}

func TestNewBufferedReaderFromReader(t *testing.T) {
	tests := []struct {
		name        string
		input       io.Reader
		expectError error
		expectSize  int64
	}{
		{
			name:        "nil reader",
			input:       nil,
			expectError: ErrReaderCannotBeNil,
		},
		{
			name:        "empty reader",
			input:       strings.NewReader(""),
			expectError: nil,
			expectSize:  0,
		},
		{
			name:        "small data",
			input:       strings.NewReader("hello world"),
			expectError: nil,
			expectSize:  11,
		},
		{
			name:        "exactly max size",
			input:       strings.NewReader(strings.Repeat("a", MaxInMemorySize)),
			expectError: nil,
			expectSize:  MaxInMemorySize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br, err := NewBufferedReaderFromReader(tt.input)

			if tt.expectError != nil {
				assert.Check(t, is.ErrorIs(err, tt.expectError))
				assert.Check(t, is.Nil(br))
			} else {
				assert.NilError(t, err)
				assert.Assert(t, br != nil)
				assert.Check(t, is.Equal(tt.expectSize, br.Size()))
			}
		})
	}
}

func TestNewBufferedReaderFromReaderWithLimit(t *testing.T) {
	tests := []struct {
		name        string
		input       io.Reader
		maxSize     int64
		expectError error
		expectSize  int64
	}{
		{
			name:        "exceeds limit",
			input:       strings.NewReader("hello world"),
			maxSize:     5,
			expectError: ErrFileSizeExceedsLimit,
		},
		{
			name:        "within limit",
			input:       strings.NewReader("hello"),
			maxSize:     10,
			expectError: nil,
			expectSize:  5,
		},
		{
			name:        "exactly at limit",
			input:       strings.NewReader("hello"),
			maxSize:     5,
			expectError: nil,
			expectSize:  5,
		},
		{
			name:        "negative limit uses default",
			input:       strings.NewReader("test"),
			maxSize:     -1,
			expectError: nil,
			expectSize:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br, err := NewBufferedReaderFromReaderWithLimit(tt.input, tt.maxSize)

			if tt.expectError != nil {
				assert.Check(t, is.ErrorIs(err, tt.expectError))
				assert.Check(t, is.Nil(br))
			} else {
				assert.NilError(t, err)
				assert.Assert(t, br != nil)
				assert.Check(t, is.Equal(tt.expectSize, br.Size()))
			}
		})
	}
}

func TestBufferedReaderRead(t *testing.T) {
	data := []byte("hello world")
	br := NewBufferedReader(data)

	buf := make([]byte, 5)
	n, err := br.Read(buf)

	assert.NilError(t, err)
	assert.Check(t, is.Equal(5, n))
	assert.Check(t, is.DeepEqual([]byte("hello"), buf))
}

func TestBufferedReaderSeek(t *testing.T) {
	data := []byte("hello world")
	br := NewBufferedReader(data)

	// Seek to position 6
	pos, err := br.Seek(6, io.SeekStart)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(int64(6), pos))

	// Read from new position
	buf := make([]byte, 5)
	n, err := br.Read(buf)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(5, n))
	assert.Check(t, is.DeepEqual([]byte("world"), buf))
}

func TestBufferedReaderClose(t *testing.T) {
	data := []byte("test")
	br := NewBufferedReader(data)

	err := br.Close()
	assert.Check(t, err)
}

func TestBufferedReaderReset(t *testing.T) {
	data := []byte("hello world")
	br := NewBufferedReader(data)

	// Read some data
	buf := make([]byte, 5)
	_, _ = br.Read(buf)

	// Reset
	br.Reset()

	// Read again from start
	buf2 := make([]byte, 5)
	n, err := br.Read(buf2)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(5, n))
	assert.Check(t, is.DeepEqual([]byte("hello"), buf2))
}

func TestBufferedReaderData(t *testing.T) {
	data := []byte("test data")
	br := NewBufferedReader(data)

	// Get data copy
	dataCopy := br.Data()

	// Verify it's a copy, not the same slice
	assert.Check(t, is.DeepEqual(data, dataCopy))
	assert.Check(t, &data[0] != &dataCopy[0])

	// Modify copy shouldn't affect original
	dataCopy[0] = 'X'
	assert.Check(t, string(data) != string(dataCopy))
}

func TestBufferedReaderNewReader(t *testing.T) {
	data := []byte("hello")
	br := NewBufferedReader(data)

	// Read from original
	buf1 := make([]byte, 2)
	_, _ = br.Read(buf1)

	// Create new reader
	newReader := br.NewReader()

	// New reader should start from beginning
	buf2 := make([]byte, 5)
	n, err := newReader.Read(buf2)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(5, n))
	assert.Check(t, is.DeepEqual(data, buf2))
}

func TestBufferedReaderNewReadSeeker(t *testing.T) {
	data := []byte("hello")
	br := NewBufferedReader(data)

	rs := br.NewReadSeeker()
	assert.Assert(t, rs != nil)

	// Test Read
	buf := make([]byte, 5)
	n, err := rs.Read(buf)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(5, n))
	assert.Check(t, is.DeepEqual(data, buf))

	// Test Seek
	pos, err := rs.Seek(0, io.SeekStart)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(int64(0), pos))
}

type mockSizedReader struct {
	io.Reader
	size int64
}

func (m *mockSizedReader) Size() int64 {
	return m.size
}

type mockLenReader struct {
	*bytes.Reader
}

func (m *mockLenReader) Len() int {
	return m.Reader.Len()
}

type mockStatReader struct {
	io.Reader
	info fs.FileInfo
	err  error
}

func (m *mockStatReader) Stat() (fs.FileInfo, error) {
	return m.info, m.err
}

type mockFileInfo struct {
	size int64
}

func (m *mockFileInfo) Name() string       { return "test" }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() fs.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() any           { return nil }

func TestInferReaderSize(t *testing.T) {
	tests := []struct {
		name        string
		reader      io.Reader
		expectSize  int64
		expectFound bool
	}{
		{
			name:        "SizedReader",
			reader:      &mockSizedReader{size: 1024},
			expectSize:  1024,
			expectFound: true,
		},
		{
			name:        "LenReader",
			reader:      &mockLenReader{Reader: bytes.NewReader(make([]byte, 512))},
			expectSize:  512,
			expectFound: true,
		},
		{
			name:        "StatReader with valid stat",
			reader:      &mockStatReader{info: &mockFileInfo{size: 2048}},
			expectSize:  2048,
			expectFound: true,
		},
		{
			name:        "StatReader with error",
			reader:      &mockStatReader{err: errors.New("stat failed")},
			expectSize:  0,
			expectFound: false,
		},
		{
			name:        "plain io.Reader",
			reader:      io.NopCloser(bytes.NewBuffer([]byte("test"))),
			expectSize:  0,
			expectFound: false,
		},
		{
			name:        "BufferedReader",
			reader:      NewBufferedReader([]byte("test data")),
			expectSize:  9,
			expectFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, found := InferReaderSize(tt.reader)
			assert.Check(t, is.Equal(tt.expectFound, found))
			if found {
				assert.Check(t, is.Equal(tt.expectSize, size))
			}
		})
	}
}
