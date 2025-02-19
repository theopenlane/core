package static

import (
	"bufio"
	"embed"
	"io/fs"
	"path"

	"github.com/rs/zerolog/log"
)

//go:embed lists/*.txt
var sectlistsFolder embed.FS

// SecList represents a list of security-related items
// the typical use case for this would be things like subdomain lists, wordlists, domain take over fingerprints, email black lists, etc
// there may be opportunities to refactor this to create jsonschema object for the lists and then load them from json files instead of text files
type SecList struct {
	Name  string
	Items []string
}

// fileNameFromURL extracts the file name from a URL (e.g., https://example.com/list.txt -> list.txt)
func fileNameFromURL(url string) string {
	return path.Base(url)
}

// hasSecList checks if a SecList exists in the embedded filesystem
func hasSecList(name string) bool {
	_, err := sectlistsFolder.Open("lists/" + name)

	return err == nil
}

// NewSecList creates a new SecList with the given name
func NewSecList(name string) *SecList {
	return &SecList{
		Name:  name,
		Items: []string{},
	}
}

// NewSecListFromURL creates a new SecList from a URL
// this is currently structured like this so that we can eventually add the ability to fetch the list from a URL if it's not present within the embedded filesystem
func NewSecListFromURL(name, url string) (*SecList, error) {
	filename := fileNameFromURL(url)
	if hasSecList(filename) {
		return NewSecListFromEmbeddedFile(name, filename)
	}

	s := NewSecList(name)

	// this intentionally does nothing for now - TODO(MKA): add httpsling client, fetch through files
	err := s.DownloadFromURL(url)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// loadFile loads a SecList from a fs.File which is the embedded filesystem
func (s *SecList) loadFile(file fs.File) error {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		s.Items = append(s.Items, line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// loadFromEmbeddedFile loads a SecList from an embedded file
func (s *SecList) loadFromEmbeddedFile(filepath string) error {
	file, err := sectlistsFolder.Open("lists/" + filepath)
	if err != nil {
		return err
	}

	return s.loadFile(file)
}

// NewSecListFromEmbeddedFile creates a new SecList from an embedded file
func NewSecListFromEmbeddedFile(name, filename string) (*SecList, error) {
	new := NewSecList(name)

	err := new.loadFromEmbeddedFile(filename)
	if err != nil {
		return nil, err
	}

	return new, nil
}

// DownloadFromURL downloads a SecList from a URL
func (s *SecList) DownloadFromURL(url string) error {
	log.Debug().Msgf("Eventually this would download %s", url)
	return nil
}
