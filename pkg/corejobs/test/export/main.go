package main

import (
	"context"
	"flag"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/riverboat/test/common"
)

// the main function here will insert a job into the river to export contents
// this will be picked up by the river server and processed
//
// You can create an export like this and retrieve the id
//
// mutation($input: CreateExportInput!) {
//   createExport(input: $input) {
//     export {
//       id
//       status
//       format
//       exportType
//     }
//   }
// }
//
// Variables:
// {
//   "input": {
//     "exportType": "INTERNAL_POLICY",
//     "format": "PDF",
//     "fields": ["id", "name", "details", "status", "createdAt", "approver"],
//     "filters": {
//       "statusEq": "published"
//     }
//   }
// }

// GraphQL Example: Create Export with Markdown Format
//
// mutation($input: CreateExportInput!) {
//   createExport(input: $input) {
//     export {
//       id
//       status
//       format
//       createdAt
//     }
//   }
// }
//
// Variables:
// {
//   "input": {
//     "exportType": "PROCEDURE",
//     "format": "MARKDOWN",
//     "fields": ["id", "name", "details", "status"],
//     "filters": {}
//   }
// }

// GraphQL Example: Create Export with DOCX Format
//
// mutation($input: CreateExportInput!) {
//   createExport(input: $input) {
//     export {
//       id
//       status
//       format
//       exportType
//     }
//   }
// }
//
// Variables:
// {
//   "input": {
//     "exportType": "ACTION_PLAN",
//     "format": "DOCX",
//     "fields": ["id", "title", "description", "dueDate", "priority", "status"]
//   }
// }

// query($id: ID!) {
//   export(id: $id) {
//     id
//     status
//     format
//     exportType
//     errorMessage
//     files {
//       edges {
//         node {
//           id
//           uri
//           providedFilename
//           size
//         }
//       }
//     }
//   }
// }

// GraphQL Example: Bulk Export (Multiple Policies)
//
// # Export multiple policies as separate PDFs
// # Create multiple exports with same format but different filters
//
// mutation($policyInput: CreateExportInput!, $procedureInput: CreateExportInput!) {
//   export1: createExport(input: $policyInput) {
//     export {
//       id
//       status
//       format
//     }
//   }
//   export2: createExport(input: $procedureInput) {
//     export {
//       id
//       status
//       format
//     }
//   }
// }
//
// Variables:
// {
//   "policyInput": {
//     "exportType": "INTERNAL_POLICY",
//     "format": "PDF",
//     "fields": ["id", "name", "details", "status"],
//     "filters": {
//       "statusEq": "published"
//     }
//   },
//   "procedureInput": {
//     "exportType": "PROCEDURE",
//     "format": "PDF",
//     "fields": ["id", "name", "details", "status"],
//     "filters": {
//       "statusEq": "published"
//     }
//   }
// }

// Example: Format-Specific Tips
//
// CSV Export:
// - Best for data analysis and imports
// - Use when you need to preserve exact field structure
// - Supports flattening of nested objects
//
// Markdown Export:
// - Best for documentation and version control
// - Single documents show full formatting
// - Great for GitHub/GitLab wikis
// - Use for creating shareable documentation
//
// DOCX Export:
// - Best for professional documents
// - Good for sharing with non-technical users
// - Supports further editing in Word
// - Includes title page and metadata
//
// PDF Export:
// - Best for final, read-only distribution
// - Ideal for printing and archival
// - Cannot be edited after export
// - Suitable for formal documentation

// Example: Handling Export Responses
//
// 1. Create export (returns ID)
// 2. Poll status endpoint until status = "READY" or "FAILED"
// 3. If READY:
//    - Download file from files.uri
//    - Open in appropriate application
// 4. If FAILED:
//    - Check errorMessage field
//    - Retry or adjust filters/fields
//
// Polling example:
// while (export.status === "PENDING") {
//   await delay(1000);
//   export = await fetchExport(exportId);
// }
// if (export.status === "READY") {
//   downloadFile(export.files[0].uri);
// }

func main() {
	client := common.NewInsertOnlyRiverClient()

	exportID := flag.String("id", "", "ID of the export to process")
	format := flag.String("format", "CSV", "Export format: CSV, PDF, MD, DOCX")

	flag.Parse()

	if *exportID == "" {
		log.Fatal().Msg("export id is required")
	}

	// Validate format
	validFormats := map[string]bool{
		"CSV":  true,
		"PDF":  true,
		"MD":   true,
		"DOCX": true,
	}

	if !validFormats[*format] {
		log.Fatal().Msgf("invalid format: %s (valid: CSV, PDF, MD, DOCX)", *format)
	}

	// Insert export job into queue
	_, err := client.Insert(context.Background(), corejobs.ExportContentArgs{
		ExportID: *exportID,
	}, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("error inserting export job")
	}

	log.Info().Msgf("export job successfully inserted for ID: %s with format: %s", *exportID, *format)
}
