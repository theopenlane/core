package certx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// GetDomainCerts queries crt.sh for all certificates for a given domain
func GetDomainCerts(ctx context.Context, domain string) (CertsReport, error) {
	errors := []string{}

	baseURL := "https://crt.sh/?q=%s&output=json"
	escapedDomain := url.QueryEscape(domain) // Properly escape the domain in the URL
	apiURL := fmt.Sprintf(baseURL, escapedDomain)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		errors = append(errors, err.Error())
		return CertsReport{Domain: domain, Errors: errors}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		errors = append(errors, err.Error())
		return CertsReport{Domain: domain, Errors: errors}, err
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			errors = append(errors, cerr.Error())
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errors = append(errors, err.Error())
	}

	var records []CertificateRecord

	err = json.Unmarshal(body, &records)
	if err != nil {
		errors = append(errors, err.Error())
	}

	// 4. Create the CertReport struct
	report := CertsReport{
		Domain:       domain,
		Certificates: records,
		Errors:       errors,
	}

	return report, nil
}
