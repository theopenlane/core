package scan

import (
	"github.com/theopenlane/core/pkg/sleuth/certx"
	"github.com/theopenlane/core/pkg/sleuth/dnsx"
	"github.com/theopenlane/core/pkg/sleuth/tech"
)

// DomainScanReport aggregates the results of multiple reconnaissance operations
// performed against a domain.
type DomainScanReport struct {
	DNS          dnsx.DNSRecordsReport   `json:"dns"`
	Certificates certx.CertsReport       `json:"certificates"`
	Technologies map[string]tech.AppInfo `json:"technologies"`
}
