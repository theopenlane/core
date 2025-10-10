package validator

import (
	"fmt"
	"github.com/vektah/gqlparser/v2/parser"
	"os"
	"path/filepath"
	"strings"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/validator"
)

func ValidateFilter(filter string, exportType enums.ExportType) error {
	if filter == "" {
		return nil
	}

	// Directory containing schema files
	schemaDir := "../../graphapi/schema"

	// Load all .graphql files from schemaDir
	var sources []*ast.Source
	err := filepath.Walk(schemaDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".graphql") {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return fmt.Errorf("failed to read schema file %s: %w", path, readErr)
			}
			sources = append(sources, &ast.Source{
				Name:  info.Name(),
				Input: string(data),
			})
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to load schema files: %w", err)
	}

	if len(sources) == 0 {
		return fmt.Errorf("no .graphql schema files found in %s", schemaDir)
	}

	schema, err := gqlparser.LoadSchema(sources...)
	if err != nil {
		return fmt.Errorf("failed to load schema: %w", err)
	}

	// query with connection structure and where clause
	query := fmt.Sprintf(`
  query {
    %ss(where: {%s}) {
      edges {
        node {
          id
        }
      }
    }
  }
`, strings.ToLower(exportType.String()), filter)

	doc, err := parser.ParseQuery(&ast.Source{Input: query})
	if err != nil {
		return fmt.Errorf("invalid GraphQL syntax: %w", err)
	}

	errs := validator.ValidateWithRules(schema, doc, nil)
	if len(errs) > 0 {
		var errMsgs []string
		for _, e := range errs {
			errMsgs = append(errMsgs, e.Message)
		}
		return fmt.Errorf("invalid filter for %s: %s", exportType, strings.Join(errMsgs, "; "))
	}

	return nil
}
