//go:build ignore

package main

import (
	"os"

	"github.com/rs/zerolog/log"

	"github.com/vektah/gqlparser/v2/formatter"

	"github.com/theopenlane/core/internal/genhelpers"
	gqlgenerated "github.com/theopenlane/core/internal/graphapi/generated"
)

// read in schema from internal package and save it to the schema file
func main() {
	genhelpers.SetupLogging()

	genhelpers.ChangeToRootDir("../../../")

	_, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get current directory")
	}

	log.Info().Msg("Generating schema for client")

	execSchema := gqlgenerated.NewExecutableSchema(gqlgenerated.Config{})
	schema := execSchema.Schema()

	// for i, t := range schema.Types {
	// 	log.Debug().Str("type", t.Name).Msg("type")

	// 	if t.Kind == ast.Object && !strings.Contains(t.Name, "History") && !strings.Contains(t.Name, "Connection") &&
	// 		!strings.Contains(t.Name, "Edge") && !strings.Contains(t.Name, "Payload") {
	// 		log.Debug().Str("type", t.Name).Msg("type")

	// 		if t.Fields.ForName("createdBy") != nil {
	// 			newField := &ast.FieldDefinition{
	// 				Name: "createdByMeow",
	// 				Type: &ast.Type{
	// 					NamedType: "String",
	// 				},
	// 			}

	// 			log.Debug().Msg("adding updatedByMeow field")

	// 			t.Fields = append(t.Fields, newField)
	// 		}

	// 		if t.Fields.ForName("updatedBy") != nil {
	// 			newField := &ast.FieldDefinition{
	// 				Name: "updatedByMeow",
	// 				Type: &ast.Type{
	// 					NamedType: "String",
	// 				},
	// 			}

	// 			log.Debug().Msg("adding updatedByMeow field")

	// 			t.Fields = append(t.Fields, newField)
	// 		}

	// 		schema.Types[i] = t
	// 	}
	// }

	f, err := os.Create("internal/graphapi/clientschema/schema.graphql")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create schema file")
	}

	defer f.Close()

	fmtr := formatter.NewFormatter(f)

	log.Info().Msg("writing schema.graphl to file")

	fmtr.FormatSchema(schema)
}
