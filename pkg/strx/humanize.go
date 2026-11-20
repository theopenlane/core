// Package strx provides string formatting helpers
package strx

import (
	"strings"
	"unicode"

	"github.com/samber/lo"
)

// initialismDisplayForms are the words kept in their written form when a snake_case identifier is converted to sentence case
var initialismDisplayForms = []string{
	// vulnerability identifiers, scoring, and classes
	"CVE", "CVSS", "CWE", "EPSS", "KEV", "NVD", "GHSA", "OSV", "CPE", "SBOM",
	"XSS", "CSRF", "SSRF", "RCE", "XXE", "IDOR", "LFI", "RFI", "MITM", "DDoS",
	"SCA", "SAST", "DAST",
	// cloud providers and services
	"AWS", "EC2", "S3", "RDS", "EKS", "ECS", "ECR", "EBS", "EFS", "ELB", "AMI", "ARN", "IMDS", "IMDSv2",
	"SNS", "SQS", "ACM", "WAF", "AKS", "ACR", "NSG",
	"GCP", "GCE", "GCS", "GKE", "COS", "CMEK", "CSEK", "KMS", "IAM", "IAP", "VPC", "VPN", "VM", "OS",
	"CDN", "NAT", "CIDR", "K8S", "RBAC", "ABAC", "ACL", "GitHub", "macOS",
	// protocols, auth, and networking
	"API", "DNS", "DNSSEC", "HTTP", "HTTPS", "TCP", "UDP", "FTP", "SFTP", "SMTP", "SMB", "SNMP", "LDAP",
	"NTP", "ICMP", "RPC", "RDP", "SSH", "VNC", "SSL", "TLS", "IP", "IPv4", "IPv6", "RSA", "X509",
	"JWT", "SAML", "OIDC", "OAuth", "SSO", "MFA", "TOTP", "PKI", "HSTS", "CORS", "CSP",
	// compliance
	"CIS", "PII", "PHI", "PCI", "DSS", "GDPR", "HIPAA", "NIST", "CSF", "ISO", "SOC", "SOC1", "SOC2", "SOC3",
	"SOX", "FedRAMP", "StateRAMP", "CMMC", "FISMA", "HITRUST", "CCPA", "CPRA", "GLBA", "FERPA", "COPPA",
	"TISAX", "CSA",
	// formats and general
	"ID", "URL", "URI", "UUID", "JSON", "YAML", "XML", "HTML", "SQL", "CPU", "UI", "CI",
}

// initialisms maps the upper case form of each initialism to its display form
var initialisms = lo.KeyBy(initialismDisplayForms, strings.ToUpper)

// HumanizeSnake converts a snake_case or UPPER_SNAKE identifier to sentence case, keeping known initialisms in their written form
// (WEAK_SSL_POLICY -> "Weak SSL policy"), values containing spaces or mixed case are returned unchanged
func HumanizeSnake(s string) string {
	if strings.Contains(s, " ") || isMixedCase(s) {
		return s
	}

	words := lo.Compact(strings.Split(strings.ToLower(s), "_"))

	for i, word := range words {
		upper := strings.ToUpper(word)

		if display, ok := initialisms[upper]; ok {
			words[i] = display

			continue
		}

		if i == 0 {
			words[i] = lo.Capitalize(word)
		}
	}

	return strings.Join(words, " ")
}

// isMixedCase reports whether s contains both upper and lower case letters
func isMixedCase(s string) bool {
	return strings.IndexFunc(s, unicode.IsUpper) >= 0 && strings.IndexFunc(s, unicode.IsLower) >= 0
}
