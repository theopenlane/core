package spider

import (
	"math"

	"github.com/projectdiscovery/katana/pkg/engine/hybrid"
	"github.com/projectdiscovery/katana/pkg/types"
)

// Spider is a wrapper struct for the katana crawler
type Spider struct {
	//	Client  *standard.Crawler
	Options *Options
	Client  *hybrid.Crawler
}

// NewOptions creates a new Options struct with default values and allows overrides
func NewOptions(opts ...Option) *Options {
	options := &Options{
		MaxDepth: 3,  // nolint:mnd
		Timeout:  30, // nolint:mnd
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// LinkDetails provides the details of a single link found during a web spider operation
type LinkDetails struct {
	Link         string   `json:"link" yaml:"link"`
	Status       int      `json:"status" yaml:"status"`
	Technologies []string `json:"technologies" yaml:"technologies"`
}

// A WebSpiderReport represents a holistic report of all the links that were found during a web spider operation, including
// non-fatal errors that occurred during the operation
type WebSpiderReport struct {
	Targets []string      `json:"targets" yaml:"targets"`
	Links   []LinkDetails `json:"links" yaml:"links"`
	Errors  []string      `json:"errors" yaml:"errors"`
}

// performWebSpider creates the options for the katana spider operation
func performWebSpider(targets []string, opts ...Option) ([]LinkDetails, []string, error) {
	errors := []string{}
	links := []LinkDetails{}

	// TODO(MKA): evaluate some of these defaults
	options := &Options{
		MaxDepth:               3, // nolint:mnd
		ScrapeJSResponses:      false,
		ScrapeJSLuiceResponses: false,
		FieldScope:             "rdn",
		BodyReadSize:           math.MaxInt,
		Timeout:                10, // nolint:mnd
		Concurrency:            10, // nolint:mnd
		Parallelism:            10, // nolint:mnd
		Retries:                1,
		Delay:                  0,             // nolint:mnd
		RateLimit:              150,           // nolint:mnd
		Strategy:               "depth-first", // Visit strategy (depth-first, breadth-first)
	}

	for _, opt := range opts {
		opt(options)
	}

	crawlerOptions, err := types.NewCrawlerOptions(MapOptionsToTypesOptions(options))
	if err != nil {
		return links, errors, err
	}

	crawler, err := hybrid.New(crawlerOptions)
	if err != nil {
		return links, errors, err
	}

	for _, target := range targets {
		err := crawler.Crawl(target)
		if err != nil {
			errors = append(errors, err.Error())
		}
	}

	return links, errors, nil
}

// PerformWebSpider performs a web spider operation against the provided targets, returning a WebSpiderReport with the
// results of the spider
func PerformWebSpider(targets []string) WebSpiderReport {
	links, errors, err := performWebSpider(targets)
	if err != nil {
		errors = append(errors, err.Error())
	}

	report := WebSpiderReport{
		Targets: targets,
		Links:   links,
		Errors:  errors,
	}

	return report
}

// MapOptionsToTypesOptions maps Options parameters to types.Options
func MapOptionsToTypesOptions(options *Options) *types.Options {
	return &types.Options{
		MaxDepth:               options.MaxDepth,
		Timeout:                options.Timeout,
		Proxy:                  options.Proxy,
		ScrapeJSResponses:      options.ScrapeJSResponses,
		ScrapeJSLuiceResponses: options.ScrapeJSLuiceResponses,
		FieldScope:             options.FieldScope,
		BodyReadSize:           options.BodyReadSize,
		Concurrency:            options.Concurrency,
		Parallelism:            options.Parallelism,
		Retries:                options.Retries,
		Delay:                  options.Delay,
		RateLimit:              options.RateLimit,
		Strategy:               options.Strategy,
		TechDetect:             options.TechDetect,
		OmitRaw:                options.OmitRaw,
		OmitBody:               options.OmitBody,
		TlsImpersonate:         options.TLSImpersonate,
		DisableRedirects:       options.DisableRedirects,
	}
}
