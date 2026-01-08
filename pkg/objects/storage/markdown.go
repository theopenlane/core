package storage

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter represents the front matter metadata in a markdown file
// only Title is supported for now, but will be extended in the future
type Frontmatter struct {
	// OpenlaneID is the unique identifier for the document in Openlane platform, used for syncing
	OpenlaneID string `yaml:"openlane_id"`

	// Title of the document
	Title string `yaml:"title"`
	// Status of the document
	Status string `yaml:"status"`
	// Tags associated with the document
	Tags []string `yaml:"tags"`
	// Revision of the document
	Revision string `yaml:"revision"`
	// Satisfies lists the standards or requirements that this document satisfies
	Satisfies map[string][]string `yaml:"satisfies"`
}

// ParseFrontmatter extracts YAML frontmatter and returns (metadata, content, error)
func ParseFrontmatter(input []byte) (*Frontmatter, []byte, error) {
	content := string(input)
	if !strings.HasPrefix(content, "---") {
		// no frontmatter present
		return nil, input, nil
	}

	// split into maximum 3 part by the `---` delimiter
	// first part is empty (before first ---)
	// second part is the frontmatter
	// third part is the content
	numParts := 3

	parts := strings.SplitN(content, "---", numParts)
	if len(parts) < numParts {
		return nil, input, nil
	}

	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
		return nil, nil, err
	}

	body := strings.TrimSpace(parts[2])
	return &fm, []byte(body), nil
}
