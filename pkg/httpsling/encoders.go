package httpsling

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"

	"github.com/goccy/go-yaml"
	"github.com/valyala/bytebufferpool"
)

// Encoder is the interface that wraps the Encode method
type Encoder interface {
	// Encode encodes the provided value into a reader
	Encode(v any) (io.Reader, error)
	// ContentType returns the content type of the encoded data
	ContentType() string
}

// Decoder is the interface that wraps the Decode method
type Decoder interface {
	// Decode decodes the data from the reader into the provided value
	Decode(r io.Reader, v any) error
}

// StreamCallback is a callback function that is called when data is received
type StreamCallback func([]byte) error

// StreamErrCallback is a callback function that is called when an error occurs
type StreamErrCallback func(error)

// StreamDoneCallback is a callback function that is called when the stream is done
type StreamDoneCallback func()

// JSONEncoder handles encoding of JSON data
type JSONEncoder struct {
	// MarshalFunc is the custom marshal function to use
	MarshalFunc func(v any) ([]byte, error)
}

// JSONDecoder handles decoding of JSON data
type JSONDecoder struct {
	UnmarshalFunc func(data []byte, v any) error
}

// DefaultJSONEncoder instance using the standard json.Marshal function
var DefaultJSONEncoder = &JSONEncoder{
	MarshalFunc: json.Marshal,
}

// DefaultJSONDecoder instance using the standard json.Unmarshal function
var DefaultJSONDecoder = &JSONDecoder{
	UnmarshalFunc: json.Unmarshal,
}

var bufferPool bytebufferpool.Pool

// poolReader wraps bytes.Reader to return the buffer to the pool when closed
type poolReader struct {
	// *bytes.Reader is an io.Reader
	*bytes.Reader
	// poolBuf is a bytebufferpool.ByteBuffer
	poolBuf *bytebufferpool.ByteBuffer
}

// ContentType returns the content type for JSON data
func (e *JSONEncoder) ContentType() string {
	return ContentTypeJSONUTF8
}

// Encode marshals the provided value into JSON format
func (e *JSONEncoder) Encode(v any) (io.Reader, error) {
	var err error

	var data []byte

	if e.MarshalFunc == nil {
		data, err = json.Marshal(v) // Fallback to standard JSON marshal if no custom function is provided
	} else {
		data, err = e.MarshalFunc(v)
	}

	if err != nil {
		return nil, err
	}

	buf := GetBuffer()

	_, err = buf.Write(data)

	if err != nil {
		PutBuffer(buf) // Ensure the buffer is returned to the pool in case of an error
		return nil, err
	}

	// we need to ensure the buffer will be returned to the pool after being read
	reader := &poolReader{Reader: bytes.NewReader(buf.B), poolBuf: buf}

	return reader, nil
}

// Decode reads the data from the reader and unmarshals it into the provided value
func (d *JSONDecoder) Decode(r io.Reader, v any) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	if d.UnmarshalFunc != nil {
		return d.UnmarshalFunc(data, v)
	}

	return json.Unmarshal(data, v)
}

// GetBuffer retrieves a buffer from the pool
func GetBuffer() *bytebufferpool.ByteBuffer {
	return bufferPool.Get()
}

// PutBuffer returns a buffer to the pool
func PutBuffer(b *bytebufferpool.ByteBuffer) {
	bufferPool.Put(b)
}

func (r *poolReader) Close() error {
	PutBuffer(r.poolBuf)

	return nil
}

// XMLEncoder handles encoding of XML data
type XMLEncoder struct {
	MarshalFunc func(v any) ([]byte, error)
}

// DefaultXMLEncoder instance using the standard xml.Marshal function
var DefaultXMLEncoder = &XMLEncoder{
	MarshalFunc: xml.Marshal,
}

// XMLDecoder handles decoding of XML data
type XMLDecoder struct {
	UnmarshalFunc func(data []byte, v any) error
}

// DefaultXMLDecoder instance using the standard xml.Unmarshal function
var DefaultXMLDecoder = &XMLDecoder{
	UnmarshalFunc: xml.Unmarshal,
}

// Encode marshals the provided value into XML format
func (e *XMLEncoder) Encode(v any) (io.Reader, error) {
	var err error

	var data []byte

	if e.MarshalFunc != nil {
		data, err = e.MarshalFunc(v)
	} else {
		data, err = xml.Marshal(v)
	}

	if err != nil {
		return nil, err
	}

	buf := GetBuffer()
	_, err = buf.Write(data)

	if err != nil {
		PutBuffer(buf)
		return nil, err
	}

	return &poolReader{Reader: bytes.NewReader(buf.B), poolBuf: buf}, nil
}

// ContentType returns the content type for XML data
func (e *XMLEncoder) ContentType() string {
	return ContentTypeXMLUTF8
}

// Decode unmarshals the XML data from the reader into the provided value
func (d *XMLDecoder) Decode(r io.Reader, v any) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	if d.UnmarshalFunc != nil {
		return d.UnmarshalFunc(data, v)
	}

	return xml.Unmarshal(data, v)
}

// YAMLEncoder handles encoding of YAML data
type YAMLEncoder struct {
	MarshalFunc func(v any) ([]byte, error)
}

// DefaultYAMLEncoder instance using the goccy/go-yaml Marshal function
var DefaultYAMLEncoder = &YAMLEncoder{
	MarshalFunc: yaml.Marshal,
}

// YAMLDecoder handles decoding of YAML data
type YAMLDecoder struct {
	UnmarshalFunc func(data []byte, v any) error
}

// DefaultYAMLDecoder instance using the goccy/go-yaml Unmarshal function
var DefaultYAMLDecoder = &YAMLDecoder{
	UnmarshalFunc: yaml.Unmarshal,
}

// Encode marshals the provided value into YAML format
func (e *YAMLEncoder) Encode(v any) (io.Reader, error) {
	var err error

	var data []byte

	if e.MarshalFunc != nil {
		data, err = e.MarshalFunc(v)
	} else {
		data, err = yaml.Marshal(v)
	}

	if err != nil {
		return nil, err
	}

	buf := GetBuffer()
	_, err = buf.Write(data)

	if err != nil {
		PutBuffer(buf)
		return nil, err
	}

	return &poolReader{Reader: bytes.NewReader(buf.B), poolBuf: buf}, nil
}

// ContentType returns the content type for YAML data
func (e *YAMLEncoder) ContentType() string {
	return ContentTypeYAMLUTF8
}

// Decode reads the data from the reader and unmarshals it into the provided value
func (d *YAMLDecoder) Decode(r io.Reader, v any) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	if d.UnmarshalFunc != nil {
		return d.UnmarshalFunc(data, v)
	}

	// Fallback to standard YAML unmarshal using goccy/go-yaml
	return yaml.Unmarshal(data, v)
}
