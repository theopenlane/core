package events

import (
	"net/http"
	"time"

	"github.com/theopenlane/core/internal/ent/generated"
)

// ScanClient bundles the ent client with configuration for the scanner service.
type ScanClient struct {
	*generated.Client
	ScannerEndpoint string
	HTTPClient      *http.Client
}

// NewScanClient creates a ScanClient with reasonable defaults.
func NewScanClient(db *generated.Client, endpoint string) *ScanClient {
	return &ScanClient{
		Client:          db,
		ScannerEndpoint: endpoint,
		HTTPClient:      &http.Client{Timeout: 60 * time.Second},
	}
}
