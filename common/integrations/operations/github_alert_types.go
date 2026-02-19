package operations

import "strings"

const (
	// GitHubAlertTypeDependabot identifies Dependabot alerts
	GitHubAlertTypeDependabot = "dependabot"
	// GitHubAlertTypeCodeScanning identifies code scanning alerts
	GitHubAlertTypeCodeScanning = "code_scanning"
	// GitHubAlertTypeSecretScanning identifies secret scanning alerts
	GitHubAlertTypeSecretScanning = "secret_scanning"
)

// NormalizeGitHubAlertType standardizes GitHub alert type identifiers
func NormalizeGitHubAlertType(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")
	switch value {
	case "dependabot_alerts":
		return GitHubAlertTypeDependabot
	case "code_scanning_alerts":
		return GitHubAlertTypeCodeScanning
	case "secret_scanning_alerts":
		return GitHubAlertTypeSecretScanning
	}

	return value
}
