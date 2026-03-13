package hooks

// severityLevelFromScore maps a CVSS v4.0 score to a normalized severity level
func severityLevelFromScore(score float64) string {
	switch {
	case score >= 9.0:
		return "CRITICAL"
	case score >= 7.0:
		return "HIGH"
	case score >= 4.0:
		return "MEDIUM"
	case score >= 0.1:
		return "LOW"
	default:
		return "NONE"
	}
}
