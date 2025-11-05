//go:build clistripe

package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/theopenlane/core/pkg/entitlements"
)

const defaultTagSubmatchCount = 2

var (
	outWriter io.Writer = os.Stdout

	// ErrUnknownMigrationStep is returned when an unknown migration step is requested.
	ErrUnknownMigrationStep = errors.New("unknown migration step")
	// ErrUnknownMigrationStage is returned when an unknown migration stage is encountered.
	ErrUnknownMigrationStage = errors.New("unknown migration stage")
	// ErrMigrationAborted indicates the user cancelled the interactive migration.
	ErrMigrationAborted = errors.New("migration aborted by user")
	// ErrAPIVersionEmpty indicates no API version was provided.
	ErrAPIVersionEmpty = errors.New("api version cannot be empty")
	// ErrAPIVersionFormatInvalid indicates the API version does not match the expected format.
	ErrAPIVersionFormatInvalid = errors.New("api version must match YYYY-MM-DD format")
	// ErrDefaultTagNotFound indicates a default struct tag could not be found.
	ErrDefaultTagNotFound = errors.New("default tag not found for field")

	apiVersionPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}(\.[a-z]+)?$`)
)

type migrationOptions struct {
	Events     []string
	NewVersion string
	RepoRoot   string
}

// getStripeClient is the factory for a Stripe client configured from CLI flags
func getStripeClient(c *cli.Command) (*entitlements.StripeClient, error) {
	apiKey := c.String("stripe-key")
	if apiKey == "" {
		return nil, entitlements.ErrMissingAPIKey
	}

	return entitlements.NewStripeClient(
		entitlements.WithAPIKey(apiKey),
	)
}

// getWebhookURL is the accessor for the CLI-provided webhook URL
func getWebhookURL(c *cli.Command) string {
	return c.String("webhook-url")
}

// getRepoRoot is the resolver for the repository root path
func getRepoRoot(c *cli.Command) string {
	if c == nil {
		return "."
	}

	root := c.String("config-root")
	if root == "" {
		root = "."
	}

	abs, err := filepath.Abs(root)
	if err != nil {
		return root
	}

	return abs
}

// readLine is the stdin helper that captures user input with an optional prompt
func readLine(prompt string) (string, error) {
	if prompt != "" {
		fmt.Fprint(outWriter, prompt)
	}

	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	return strings.TrimSpace(text), nil
}

// promptForNewAPIVersion is the interactive prompt for selecting a target API version
func promptForNewAPIVersion(currentVersion, suggested string) (string, error) {
	fmt.Fprintf(outWriter, "\nCurrent webhook API version: %s\n", displayValue(currentVersion))
	fmt.Fprintf(outWriter, "Latest Stripe SDK API version: %s\n", suggested)

	version, err := readLine("Enter the new Stripe API version to migrate to: ")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(version), nil
}

// confirmAction is the yes-or-no prompt helper used before making changes
func confirmAction(question string) (bool, error) {
	response, err := readLine(question + " [y/N]: ")
	if err != nil {
		return false, err
	}

	switch strings.ToLower(strings.TrimSpace(response)) {
	case "y", "yes":
		return true, nil
	case "":
		return false, nil
	default:
		return false, nil
	}
}

// validateAPIVersion is the guard that checks provided Stripe version strings
func validateAPIVersion(version string) error {
	if version == "" {
		return ErrAPIVersionEmpty
	}

	if !apiVersionPattern.MatchString(version) {
		return fmt.Errorf("api version %q must match YYYY-MM-DD format: %w", version, ErrAPIVersionFormatInvalid)
	}

	return nil
}

// displayValue is the formatter that returns a placeholder for empty strings
func displayValue(value string) string {
	if value == "" {
		return "<unset>"
	}

	return value
}

// sanitizeVersionForEnv is the normalizer that converts Stripe versions into env-safe tokens
func sanitizeVersionForEnv(version string) string {
	sanitized := strings.ToUpper(version)
	sanitized = strings.ReplaceAll(sanitized, "-", "_")
	sanitized = strings.ReplaceAll(sanitized, ".", "_")
	return sanitized
}

// secretEnvKeyForVersion is the builder for environment keys per API version
func secretEnvKeyForVersion(version string) string {
	return fmt.Sprintf("CORE_SUBSCRIPTION_STRIPEWEBHOOKSECRETS_%s", sanitizeVersionForEnv(version))
}

// readDefaultAPIVersions is the reader that loads default webhook versions from config
func readDefaultAPIVersions(repoRoot string) (string, string, error) {
	target := filepath.Join(repoRoot, "pkg", "entitlements", "config.go")

	content, err := os.ReadFile(target)
	if err != nil {
		return "", "", fmt.Errorf("failed to read %s: %w", target, err)
	}

	current := extractDefaultValue(content, "StripeWebhookAPIVersion")
	discard := extractDefaultValue(content, "StripeWebhookDiscardAPIVersion")

	return current, discard, nil
}

// extractDefaultValue is the parser that pulls default struct tag values for a field
func extractDefaultValue(content []byte, field string) string {
	pattern := fmt.Sprintf(`(?s)%s\s+.*?default:"([^"]*)"`, regexp.QuoteMeta(field))
	re := regexp.MustCompile(pattern)

	if match := re.FindSubmatch(content); len(match) == defaultTagSubmatchCount {
		return string(match[1])
	}

	return ""
}

// updateAPIVersionDefaults is the mutator that rewrites default Stripe versions in config
func updateAPIVersionDefaults(repoRoot, oldVersion, newVersion string) (bool, error) {
	target := filepath.Join(repoRoot, "pkg", "entitlements", "config.go")
	content, err := os.ReadFile(target)
	if err != nil {
		return false, fmt.Errorf("failed to read %s: %w", target, err)
	}

	info, err := os.Stat(target)
	if err != nil {
		return false, fmt.Errorf("failed to stat %s: %w", target, err)
	}

	updated := content
	var changed bool

	updateField := func(field, value string) error {
		pattern := fmt.Sprintf(`(?s)(%s\s+.*?default:")([^"]*)(")`, regexp.QuoteMeta(field))
		re := regexp.MustCompile(pattern)

		if !re.Match(updated) {
			return fmt.Errorf("could not locate default tag for %s: %w", field, ErrDefaultTagNotFound)
		}

		replacement := re.ReplaceAll(updated, []byte(fmt.Sprintf("${1}%s${3}", value)))

		if !bytes.Equal(updated, replacement) {
			changed = true
			updated = replacement
		}

		return nil
	}

	if err := updateField("StripeWebhookAPIVersion", newVersion); err != nil {
		return false, err
	}

	if err := updateField("StripeWebhookDiscardAPIVersion", oldVersion); err != nil {
		return false, err
	}

	if !changed {
		return false, nil
	}

	if err := os.WriteFile(target, updated, info.Mode().Perm()); err != nil {
		return false, fmt.Errorf("failed to write %s: %w", target, err)
	}

	return true, nil
}
