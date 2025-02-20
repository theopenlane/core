package sub

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/rs/zerolog/log"
)

var (
	ErrNilRunner  = errors.New("runner is nil")
	ErrNilOptions = errors.New("options are nil")
)

// Enumerate is a method that runs the subfinder tool to enumerate subdomains for a given domain
func (s *Subfinder) Enumerate() ([]string, error) {
	if s.Runner == nil {
		return nil, ErrNilRunner
	}

	if s.Options == nil {
		return nil, ErrNilOptions
	}

	output := &bytes.Buffer{}

	sourceMap, err := s.Runner.EnumerateSingleDomain("theopenlane.io", []io.Writer{output})
	if err != nil {
		log.Error().Err(err).Msg("Error running subfinder")
		return nil, err
	}

	subdomains := make([]string, 0, len(sourceMap))
	for subdomain := range sourceMap {
		subdomains = append(subdomains, subdomain)
	}

	return subdomains, nil
}

// ConvertBufferToSlice converts a bytes.Buffer to a slice of strings
func ConvertBufferToSlice(buffer *bytes.Buffer) []string {
	outputLines := strings.Split(buffer.String(), "\n")

	for i, line := range outputLines {
		outputLines[i] = strings.TrimSpace(line)
	}

	return outputLines
}
