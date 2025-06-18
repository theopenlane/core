package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/sleuth/scan"
)

const TopicDomainScan = "domain.scan"

// DomainScanEvent is emitted when a new domain scan should be executed.
type DomainScanEvent struct {
	ScanID string
	Domain string
	OrgID  string
}

// RegisterDomainScanListener registers the asynchronous domain scan listener.
func RegisterDomainScanListener(pool *soiree.EventPool) error {
	if pool == nil {
		return fmt.Errorf("nil event pool")
	}

	_, err := pool.On(TopicDomainScan, handleDomainScan)
	return err
}

func handleDomainScan(evt soiree.Event) error {
	payload, ok := evt.Payload().(DomainScanEvent)
	if !ok {
		log.Error().Msg("invalid domain scan payload")
		return nil
	}

	sc := evt.Client().(*ScanClient)
	client := sc.Client
	ctx := evt.Context()

	report, err := requestDomainScan(ctx, sc, payload.Domain)
	status := enums.ScanStatusCompleted
	if err != nil {
		log.Error().Err(err).Msg("domain scan failed")
		status = enums.ScanStatusFailed
	}

	techIDs := []string{}
	for name, info := range report.Technologies {
		description := info.Description
		if description == "" && client.Summarizer != nil {
			sum, serr := client.Summarizer.Summarize(ctx, "Asset info: "+name)
			if serr == nil {
				description = sum
			}
		}
	}

	log.Info().Str("status", status.String()).Str("domain", payload.Domain).Str("scan_id", payload.ScanID).Msg("domain scan status")

	log.Info().Str("techIDs", fmt.Sprintf("%v", techIDs))

	return nil
}

func requestDomainScan(ctx context.Context, sc *ScanClient, domain string) (scan.DomainScanReport, error) {
	if sc.ScannerEndpoint == "" {
		return scan.ScanDomain(ctx, domain)
	}

	var report scan.DomainScanReport
	body, _ := json.Marshal(map[string]string{"domain": domain})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sc.ScannerEndpoint+"/scan/domain", bytes.NewReader(body))
	if err != nil {
		return report, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := sc.HTTPClient.Do(req)
	if err != nil {
		return report, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return report, fmt.Errorf("scanner response %s", resp.Status)
	}
	err = json.NewDecoder(resp.Body).Decode(&report)
	return report, err
}
