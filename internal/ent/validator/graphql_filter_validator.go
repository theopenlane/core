package validator

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/common/enums"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
)

var ErrInvalidGraphQLFilter = errors.New("invalid graphql filter provided")

func ValidateFilter(filter string, exportType enums.ExportType) error {
	if filter == "" {
		log.Debug().Msg("No filter provided, skipping validation")
		return nil
	}

	schemaDir := "../../graphapi/schema"

	// Load all .graphql files in the directory
	var sources []*ast.Source
	err := filepath.Walk(schemaDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Failed accessing file path")
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".graphql") {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				log.Error().Err(readErr).Str("path", path).Msg("Failed to read schema file")
				return nil
			}
			sources = append(sources, &ast.Source{
				Name:  info.Name(),
				Input: string(data),
			})
		}
		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("Error walking through schema directory")
		return nil
	}

	if len(sources) == 0 {
		log.Error().Msg("No .graphql schema files found in " + schemaDir)
		return nil
	}

	schema, err := gqlparser.LoadSchema(sources...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load combined schema")
		return nil
	}

	// Construct query dynamically
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

	// Parse query
	doc, err := parser.ParseQuery(&ast.Source{Input: query})
	if err != nil {
		log.Error().Err(err).Str("query", query).Msg("GraphQL query syntax error")
		return nil
	}

	// Validate query against schema
	errs := validator.ValidateWithRules(schema, doc, nil)
	if len(errs) > 0 {
		var errMsgs []string
		for _, e := range errs {
			log.Error().Str("path", e.Path.String()).Str("rule", e.Rule).Msg(e.Message)
			errMsgs = append(errMsgs, fmt.Sprintf("[%s] %s", e.Rule, e.Message))
		}

		combinedErr := fmt.Errorf("%w for export type %s: %s",
			ErrInvalidGraphQLFilter, exportType, strings.Join(errMsgs, "; "))

		log.Error().Err(combinedErr).Msg("GraphQL filter validation failed")
		return combinedErr
	}

	return nil
}
