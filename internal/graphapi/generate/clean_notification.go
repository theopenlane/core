//go:build ignore

package main

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rs/zerolog/log"
)

// Clean up notification references from ent.graphql since notifications
// are internal-only and should not be exposed via GraphQL
func main() {
	schemaPath, tmpPath := resolveSchemaPaths()

	input, err := os.Open(schemaPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open ent.graphql")
	}
	defer input.Close()

	output, err := os.Create(tmpPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create temp file")
	}
	defer output.Close()

	scanner := bufio.NewScanner(input)
	writer := bufio.NewWriter(output)
	defer writer.Flush()

	inNotificationBlock := false
	inDocString := false
	blockDepth := 0
	docStringBuffer := []string{}

	for scanner.Scan() {
		line := scanner.Text()

		// Track doc strings
		if strings.Contains(line, `"""`) {
			if inDocString {
				// End of doc string
				docStringBuffer = append(docStringBuffer, line)
				inDocString = false

				// Peek ahead to see if this is for a Notification type
				if !scanner.Scan() {
					break
				}
				nextLine := scanner.Text()

				// If next line is Notification type, skip the doc string
				if strings.HasPrefix(strings.TrimSpace(nextLine), "type NotificationConnection") ||
					strings.HasPrefix(strings.TrimSpace(nextLine), "type NotificationEdge") {
					docStringBuffer = []string{}
					inNotificationBlock = true
					blockDepth = 0
					continue
				}

				// Otherwise write the doc string
				for _, buffered := range docStringBuffer {
					writer.WriteString(buffered + "\n")
				}
				writer.WriteString(nextLine + "\n")
				docStringBuffer = []string{}
				continue
			} else {
				// Start of doc string
				inDocString = true
				docStringBuffer = []string{line}
				continue
			}
		}

		// Buffer doc string lines
		if inDocString {
			docStringBuffer = append(docStringBuffer, line)
			continue
		}

		// Skip notification-related input fields
		if strings.Contains(line, "addNotificationIDs") ||
			strings.Contains(line, "removeNotificationIDs") ||
			strings.Contains(line, "clearNotifications") {
			continue
		}

		// Detect start of Notification types (without docstring)
		if strings.HasPrefix(strings.TrimSpace(line), "type NotificationConnection") ||
			strings.HasPrefix(strings.TrimSpace(line), "type NotificationEdge") {
			inNotificationBlock = true
			blockDepth = 0
			continue
		}

		// Track block depth when in notification block
		if inNotificationBlock {
			if strings.Contains(line, "{") {
				blockDepth++
			}
			if strings.Contains(line, "}") {
				blockDepth--
				if blockDepth <= 0 {
					inNotificationBlock = false
				}
			}
			continue
		}

		// Skip notifications query field
		if strings.Contains(line, "notifications(") {
			// Skip until we find the closing parenthesis and return type
			for scanner.Scan() {
				line = scanner.Text()
				if strings.Contains(line, "): NotificationConnection!") {
					break
				}
			}
			continue
		}

		writer.WriteString(line + "\n")
	}

	if err := scanner.Err(); err != nil {
		log.Fatal().Err(err).Msg("error reading file")
	}

	writer.Flush()
	output.Close()
	input.Close()

	// Replace original with cleaned version
	if err := os.Rename(tmpPath, schemaPath); err != nil {
		log.Fatal().Err(err).Msg("failed to replace original file")
	}

	log.Info().Msg("Cleaned notification references from ent.graphql")
}

func resolveSchemaPaths() (string, string) {
	_, scriptFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal().Msg("failed to determine script path")
	}

	scriptDir := filepath.Dir(scriptFile)
	schemaPath := filepath.Join(scriptDir, "..", "schema", "ent.graphql")
	schemaPath = filepath.Clean(schemaPath)

	return schemaPath, schemaPath + ".tmp"
}
