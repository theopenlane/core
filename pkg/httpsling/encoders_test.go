package httpsling_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/httpsling"
)

func TestJSONEncoderEncode(t *testing.T) {
	encoder := &httpsling.JSONEncoder{}

	// Test encoding a struct
	data := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{
		Name: "John Doe",
		Age:  30,
	}

	reader, err := encoder.Encode(data)
	require.NoError(t, err)

	encodedData, err := io.ReadAll(reader)
	require.NoError(t, err)

	expectedData := `{"name":"John Doe","age":30}`
	require.Equal(t, expectedData, string(encodedData))
}

func TestJSONDecoderDecode(t *testing.T) {
	decoder := &httpsling.JSONDecoder{}

	// Test decoding JSON data into a struct
	jsonData := `{"name":"John Snow","age":30}`

	var data struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	err := decoder.Decode(bytes.NewReader([]byte(jsonData)), &data)
	require.NoError(t, err)

	expectedData := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{
		Name: "John Snow",
		Age:  30,
	}
	require.Equal(t, expectedData, data)
}

func TestXMLDecoderDecode(t *testing.T) {
	decoder := &httpsling.XMLDecoder{}

	// Test decoding XML data into a struct
	xmlData := `<root><name>John Meow</name><age>30</age></root>`

	var data struct {
		Name string `xml:"name"`
		Age  int    `xml:"age"`
	}

	err := decoder.Decode(bytes.NewReader([]byte(xmlData)), &data)
	require.NoError(t, err)

	expectedData := struct {
		Name string `xml:"name"`
		Age  int    `xml:"age"`
	}{
		Name: "John Meow",
		Age:  30,
	}
	require.Equal(t, expectedData, data)
}

func TestYAMLEncoderEncode(t *testing.T) {
	encoder := &httpsling.YAMLEncoder{}

	// Test encoding a struct
	data := struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}{
		Name: "John Flow",
		Age:  30,
	}

	reader, err := encoder.Encode(data)
	require.NoError(t, err)

	encodedData, err := io.ReadAll(reader)
	require.NoError(t, err)

	expectedData := "name: John Flow\nage: 30\n"
	require.Equal(t, expectedData, string(encodedData))
}

func TestYAMLDecoderDecode(t *testing.T) {
	decoder := &httpsling.YAMLDecoder{}

	// Test decoding YAML data into a struct
	yamlData := "name: John Show\nage: 30\n"

	var data struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}

	err := decoder.Decode(bytes.NewReader([]byte(yamlData)), &data)
	require.NoError(t, err)

	expectedData := struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}{
		Name: "John Show",
		Age:  30,
	}
	require.Equal(t, expectedData, data)
}
