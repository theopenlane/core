package tech

import (
	"io"

	unitutils "github.com/projectdiscovery/utils/unit"
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"github.com/theopenlane/httpsling"
)

// Tech is a struct that holds the Wappalyzer client and HTTP client
type Tech struct {
	TechClient *wappalyzer.Wappalyze
	HTTPClient *httpsling.Requester
	AppInfo    map[string]AppInfo
}

// AppInfo is a struct that holds information about an application
type AppInfo struct {
	Description string   `json:"description"`
	Website     string   `json:"website"`
	CPE         string   `json:"cpe"`
	Icon        string   `json:"icon"`
	Categories  []string `json:"categories"`
}

// NewTech creates a new tech and http client
func NewTech(target string) (*Tech, error) {
	techClient, err := wappalyzer.New()
	if err != nil {
		return nil, err
	}

	httpClient, err := httpsling.New(httpsling.URL(target))
	if err != nil {
		return nil, err
	}

	return &Tech{
		TechClient: techClient,
		HTTPClient: httpClient,
	}, nil
}

// for sanity
const (
	maxDefaultBody = 4 * unitutils.Mega
)

// GetTech fetches the technology information from the target URL
func (t *Tech) GetTech() (map[string]AppInfo, error) {
	appInformation := make(map[string]AppInfo)

	resp, err := t.HTTPClient.Send()
	if err != nil {
		return appInformation, err
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxDefaultBody))
	if err != nil {
		return appInformation, nil
	}

	// fingerprint headers and body
	appInfo := t.TechClient.FingerprintWithInfo(resp.Header, data)
	t.AppInfo = make(map[string]AppInfo)

	// TODO(MKA): consider adding the name of the technology to the AppInfo struct instead of this weirdness
	// it would be ideal to incorporate the results of the various subpackages into a "report"
	for name, app := range appInfo {
		(t.AppInfo)[name] = AppInfo{
			Description: app.Description,
			Website:     app.Website,
			CPE:         app.CPE,
			Icon:        app.Icon,
			Categories:  app.Categories,
		}
	}

	return t.AppInfo, nil
}
