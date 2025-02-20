package spider

import (
	"regexp"
	"time"

	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/katana/pkg/output"
)

// Options are the functional parameters for katana
type Options struct {
	// URLs contains a list of URLs for crawling
	URLs goflags.StringSlice
	// Resume the scan from the state stored in the resume config file
	Resume string
	// Exclude host matching specified filter ('cdn', 'private-ips', cidr, ip, regex)
	Exclude goflags.StringSlice
	// Scope contains a list of regexes for in-scope URLS
	Scope goflags.StringSlice
	// OutOfScope contains a list of regexes for out-scope URLS
	OutOfScope goflags.StringSlice
	// NoScope disables host based default scope
	NoScope bool
	// DisplayOutScope displays out of scope items in results
	DisplayOutScope bool
	// ExtensionsMatch contains extensions to match explicitly
	ExtensionsMatch goflags.StringSlice
	// ExtensionFilter contains additional items for filter list
	ExtensionFilter goflags.StringSlice
	// OutputMatchCondition is the condition to match output
	OutputMatchCondition string
	// OutputFilterCondition is the condition to filter output
	OutputFilterCondition string
	// MaxDepth is the maximum depth to crawl
	MaxDepth int
	// BodyReadSize is the maximum size of response body to read
	BodyReadSize int
	// Timeout is the time to wait for request in seconds
	Timeout int
	// CrawlDuration is the duration in seconds to crawl target from
	CrawlDuration time.Duration
	// Delay is the delay between each crawl requests in seconds
	Delay int
	// RateLimit is the maximum number of requests to send per second
	RateLimit int
	// Retries is the number of retries to do for request
	Retries int
	// RateLimitMinute is the maximum number of requests to send per minute
	RateLimitMinute int
	// Concurrency is the number of concurrent crawling goroutines
	Concurrency int
	// Parallelism is the number of urls processing goroutines
	Parallelism int
	// FormConfig is the path to the form configuration file
	FormConfig string
	// Proxy is the URL for the proxy server
	Proxy string
	// Strategy is the crawling strategy. depth-first or breadth-first
	Strategy string
	// FieldScope is the scope field for default DNS scope
	FieldScope string
	// OutputFile is the file to write output to
	OutputFile string
	// KnownFiles enables crawling of knows files like robots.txt, sitemap.xml, etc
	KnownFiles string
	// Fields is the fields to format in output
	Fields string
	// StoreFields is the fields to store in separate per-host files
	StoreFields string
	// FieldConfig is the path to the custom field configuration file
	FieldConfig string
	// NoColors disables coloring of response output
	NoColors bool
	// JSON enables writing output in JSON format
	JSON bool
	// Silent shows only output
	Silent bool
	// Verbose specifies showing verbose output
	Verbose bool
	// TechDetect enables technology detection
	TechDetect bool
	// Version enables showing of crawler version
	Version bool
	// ScrapeJSResponses enables scraping of relative endpoints from javascript
	ScrapeJSResponses bool
	// ScrapeJSLuiceResponses enables scraping of endpoints from javascript using jsluice
	ScrapeJSLuiceResponses bool
	// CustomHeaders is a list of custom headers to add to request
	CustomHeaders goflags.StringSlice
	// Headless enables headless scraping
	Headless bool
	// AutomaticFormFill enables optional automatic form filling and submission
	AutomaticFormFill bool
	// FormExtraction enables extraction of form, input, textarea & select elements
	FormExtraction bool
	// UseInstalledChrome skips chrome install and use local instance
	UseInstalledChrome bool
	// ShowBrowser specifies whether the show the browser in headless mode
	ShowBrowser bool
	// HeadlessOptionalArguments specifies optional arguments to pass to Chrome
	HeadlessOptionalArguments goflags.StringSlice
	// HeadlessNoSandbox specifies if chrome should be start in --no-sandbox mode
	HeadlessNoSandbox bool
	// SystemChromePath : Specify the chrome binary path for headless crawling
	SystemChromePath string
	// ChromeWSUrl : Specify the Chrome debugger websocket url for a running Chrome instance to attach to
	ChromeWSUrl string
	// OnResult allows callback function on a result
	OnResult OnResultCallback
	// StoreResponse specifies if katana should store http requests/responses
	StoreResponse bool
	// StoreResponseDir specifies if katana should use a custom directory to store http requests/responses
	StoreResponseDir string
	// NoClobber specifies if katana should overwrite existing output files
	NoClobber bool
	// StoreFieldDir specifies if katana should use a custom directory to store fields
	StoreFieldDir string
	// OmitRaw omits raw requests/responses from the output
	OmitRaw bool
	// OmitBody omits the response body from the output
	OmitBody bool
	// ChromeDataDir : 	Specify the --user-data-dir to chrome binary to preserve sessions
	ChromeDataDir string
	// HeadlessNoIncognito specifies if chrome should be started without incognito mode
	HeadlessNoIncognito bool
	// XhrExtraction extract xhr requests
	XhrExtraction bool
	// HealthCheck determines if a self-healthcheck should be performed
	HealthCheck bool
	// PprofServer enables pprof server
	PprofServer bool
	// ErrorLogFile specifies a file to write with the errors of all requests
	ErrorLogFile string
	// Resolvers contains custom resolvers
	Resolvers goflags.StringSlice
	// OutputMatchRegex is the regex to match output url
	OutputMatchRegex goflags.StringSlice
	// OutputFilterRegex is the regex to filter output url
	OutputFilterRegex goflags.StringSlice
	// FilterRegex is the slice regex to filter url
	FilterRegex []*regexp.Regexp
	// MatchRegex is the slice regex to match url
	MatchRegex []*regexp.Regexp
	// DisableUpdateCheck disables automatic update check
	DisableUpdateCheck bool
	// IgnoreQueryParams ignore crawling same path with different query-param values
	IgnoreQueryParams bool
	// Debug
	Debug bool
	// TlsImpersonate enables experimental tls ClientHello randomization for standard crawler
	TLSImpersonate bool
	// DisableRedirects disables the following of redirects
	DisableRedirects bool
}

// OnResultCallback is a callback function that is called when a result is found
type OnResultCallback func(result output.Result) error

// Option is a functional option for the katana crawler
type Option func(*Options)

// WithURLs sets the URLs to crawl
func WithURLs(urls []string) Option {
	return func(options *Options) {
		options.URLs = goflags.StringSlice(urls)
	}
}

// WithResume sets the resume file to use
func WithResume(resume string) Option {
	return func(options *Options) {
		options.Resume = resume
	}
}

// WithExclude sets the exclude filter to use
func WithExclude(exclude []string) Option {
	return func(options *Options) {
		options.Exclude = goflags.StringSlice(exclude)
	}
}

// WithScope sets the scope regexes to use
func WithScope(scope []string) Option {
	return func(options *Options) {
		options.Scope = goflags.StringSlice(scope)
	}
}

// WithOutOfScope sets the out of scope regexes to use
func WithOutOfScope(outOfScope []string) Option {
	return func(options *Options) {
		options.OutOfScope = goflags.StringSlice(outOfScope)
	}
}

// WithNoScope sets the no scope flag
func WithNoScope(noScope bool) Option {
	return func(options *Options) {
		options.NoScope = noScope
	}
}

// WithDisplayOutScope sets the display out of scope flag
func WithDisplayOutScope(displayOutScope bool) Option {
	return func(options *Options) {
		options.DisplayOutScope = displayOutScope
	}
}

// WithExtensionsMatch sets the extensions to match
func WithExtensionsMatch(extensionsMatch []string) Option {
	return func(options *Options) {
		options.ExtensionsMatch = goflags.StringSlice(extensionsMatch)
	}
}

// WithExtensionFilter sets the extension filter
func WithExtensionFilter(extensionFilter []string) Option {
	return func(options *Options) {
		options.ExtensionFilter = goflags.StringSlice(extensionFilter)
	}
}

// WithOutputMatchCondition sets the output match condition
func WithOutputMatchCondition(outputMatchCondition string) Option {
	return func(options *Options) {
		options.OutputMatchCondition = outputMatchCondition
	}
}

// WithOutputFilterCondition sets the output filter condition
func WithOutputFilterCondition(outputFilterCondition string) Option {
	return func(options *Options) {
		options.OutputFilterCondition = outputFilterCondition
	}
}

// WithMaxDepth sets the maximum depth to crawl
func WithMaxDepth(maxDepth int) Option {
	return func(options *Options) {
		options.MaxDepth = maxDepth
	}
}

// WithBodyReadSize sets the maximum size of response body to read
func WithBodyReadSize(bodyReadSize int) Option {
	return func(options *Options) {
		options.BodyReadSize = bodyReadSize
	}
}

// WithTimeout sets the timeout for requests
func WithTimeout(timeout int) Option {
	return func(options *Options) {
		options.Timeout = timeout
	}
}

// WithCrawlDuration sets the crawl duration
func WithCrawlDuration(crawlDuration time.Duration) Option {
	return func(options *Options) {
		options.CrawlDuration = crawlDuration
	}
}

// WithDelay sets the delay between requests
func WithDelay(delay int) Option {
	return func(options *Options) {
		options.Delay = delay
	}
}

// WithRateLimit sets the rate limit for requests
func WithRateLimit(rateLimit int) Option {
	return func(options *Options) {
		options.RateLimit = rateLimit
	}
}

// WithRetries sets the number of retries for requests
func WithRetries(retries int) Option {
	return func(options *Options) {
		options.Retries = retries
	}
}

// WithRateLimitMinute sets the rate limit for requests per minute
func WithRateLimitMinute(rateLimitMinute int) Option {
	return func(options *Options) {
		options.RateLimitMinute = rateLimitMinute
	}
}

// WithConcurrency sets the number of concurrent crawling goroutines
func WithConcurrency(concurrency int) Option {
	return func(options *Options) {
		options.Concurrency = concurrency
	}
}

// WithParallelism sets the number of urls processing goroutines
func WithParallelism(parallelism int) Option {
	return func(options *Options) {
		options.Parallelism = parallelism
	}
}

// WithFormConfig sets the form configuration file
func WithFormConfig(formConfig string) Option {
	return func(options *Options) {
		options.FormConfig = formConfig
	}
}

// WithProxy sets the proxy URL
func WithProxy(proxy string) Option {
	return func(options *Options) {
		options.Proxy = proxy
	}
}

// WithStrategy sets the crawling strategy
func WithStrategy(strategy string) Option {
	return func(options *Options) {
		options.Strategy = strategy
	}
}

// WithFieldScope sets the field scope for default DNS scope
func WithFieldScope(fieldScope string) Option {
	return func(options *Options) {
		options.FieldScope = fieldScope
	}
}

// WithOutputFile sets the output file
func WithOutputFile(outputFile string) Option {
	return func(options *Options) {
		options.OutputFile = outputFile
	}
}

// WithKnownFiles sets the known files to crawl
func WithKnownFiles(knownFiles string) Option {
	return func(options *Options) {
		options.KnownFiles = knownFiles
	}
}

// WithFields sets the fields to format in output
func WithFields(fields string) Option {
	return func(options *Options) {
		options.Fields = fields
	}
}

// WithStoreFields sets the fields to store in separate per-host files
func WithStoreFields(storeFields string) Option {
	return func(options *Options) {
		options.StoreFields = storeFields
	}
}

// WithFieldConfig sets the custom field configuration file
func WithFieldConfig(fieldConfig string) Option {
	return func(options *Options) {
		options.FieldConfig = fieldConfig
	}
}

// WithNoColors disables coloring of response output
func WithNoColors(noColors bool) Option {
	return func(options *Options) {
		options.NoColors = noColors
	}
}

// WithJSON enables writing output in JSON format
func WithJSON(json bool) Option {
	return func(options *Options) {
		options.JSON = json
	}
}

// WithSilent shows only output
func WithSilent(silent bool) Option {
	return func(options *Options) {
		options.Silent = silent
	}
}

// WithVerbose specifies showing verbose output
func WithVerbose(verbose bool) Option {
	return func(options *Options) {
		options.Verbose = verbose
	}
}

// WithTechDetect enables technology detection
func WithTechDetect(techDetect bool) Option {
	return func(options *Options) {
		options.TechDetect = techDetect
	}
}

// WithVersion enables showing of crawler version
func WithVersion(version bool) Option {
	return func(options *Options) {
		options.Version = version
	}
}

// WithScrapeJSResponses enables scraping of relative endpoints from javascript
func WithScrapeJSResponses(scrapeJSResponses bool) Option {
	return func(options *Options) {
		options.ScrapeJSResponses = scrapeJSResponses
	}
}

// WithScrapeJSLuiceResponses enables scraping of endpoints from javascript using jsluice
func WithScrapeJSLuiceResponses(scrapeJSLuiceResponses bool) Option {
	return func(options *Options) {
		options.ScrapeJSLuiceResponses = scrapeJSLuiceResponses
	}
}

// WithCustomHeaders sets the custom headers to add to request
func WithCustomHeaders(customHeaders []string) Option {
	return func(options *Options) {
		options.CustomHeaders = goflags.StringSlice(customHeaders)
	}
}

// WithHeadless enables headless scraping
func WithHeadless(headless bool) Option {
	return func(options *Options) {
		options.Headless = headless
	}
}

// WithAutomaticFormFill enables optional automatic form filling and submission
func WithAutomaticFormFill(automaticFormFill bool) Option {
	return func(options *Options) {
		options.AutomaticFormFill = automaticFormFill
	}
}

// WithFormExtraction enables extraction of form, input, textarea & select elements
func WithFormExtraction(formExtraction bool) Option {
	return func(options *Options) {
		options.FormExtraction = formExtraction
	}
}

// WithUseInstalledChrome skips chrome install and use local instance
func WithUseInstalledChrome(useInstalledChrome bool) Option {
	return func(options *Options) {
		options.UseInstalledChrome = useInstalledChrome
	}
}

// WithShowBrowser specifies whether the show the browser in headless mode
func WithShowBrowser(showBrowser bool) Option {
	return func(options *Options) {
		options.ShowBrowser = showBrowser
	}
}

// WithHeadlessOptionalArguments specifies optional arguments to pass to Chrome
func WithHeadlessOptionalArguments(headlessOptionalArguments []string) Option {
	return func(options *Options) {
		options.HeadlessOptionalArguments = goflags.StringSlice(headlessOptionalArguments)
	}
}

// WithHeadlessNoSandbox specifies if chrome should be start in --no-sandbox mode
func WithHeadlessNoSandbox(headlessNoSandbox bool) Option {
	return func(options *Options) {
		options.HeadlessNoSandbox = headlessNoSandbox
	}
}

// WithSystemChromePath specifies the chrome binary path for headless crawling
func WithSystemChromePath(systemChromePath string) Option {
	return func(options *Options) {
		options.SystemChromePath = systemChromePath
	}
}

// WithChromeWSUrl specifies the Chrome debugger websocket url for a running Chrome instance to attach to
func WithChromeWSUrl(chromeWSUrl string) Option {
	return func(options *Options) {
		options.ChromeWSUrl = chromeWSUrl
	}
}

// WithOnResult allows callback function on a result
func WithOnResult(onResult OnResultCallback) Option {
	return func(options *Options) {
		options.OnResult = onResult
	}
}

// WithStoreResponse specifies if katana should store http requests/responses
func WithStoreResponse(storeResponse bool) Option {
	return func(options *Options) {
		options.StoreResponse = storeResponse
	}
}

// WithStoreResponseDir specifies if katana should use a custom directory to store http requests/responses
func WithStoreResponseDir(storeResponseDir string) Option {
	return func(options *Options) {
		options.StoreResponseDir = storeResponseDir
	}
}

// WithNoClobber specifies if katana should overwrite existing output files
func WithNoClobber(noClobber bool) Option {
	return func(options *Options) {
		options.NoClobber = noClobber
	}
}

// WithStoreFieldDir specifies if katana should use a custom directory to store fields
func WithStoreFieldDir(storeFieldDir string) Option {
	return func(options *Options) {
		options.StoreFieldDir = storeFieldDir
	}
}

// WithOmitRaw omits raw requests/responses from the output
func WithOmitRaw(omitRaw bool) Option {
	return func(options *Options) {
		options.OmitRaw = omitRaw
	}
}

// WithOmitBody omits the response body from the output
func WithOmitBody(omitBody bool) Option {
	return func(options *Options) {
		options.OmitBody = omitBody
	}
}

// WithChromeDataDir specifies the --user-data-dir to chrome binary to preserve sessions
func WithChromeDataDir(chromeDataDir string) Option {
	return func(options *Options) {
		options.ChromeDataDir = chromeDataDir
	}
}

// WithHeadlessNoIncognito specifies if chrome should be started without incognito mode
func WithHeadlessNoIncognito(headlessNoIncognito bool) Option {
	return func(options *Options) {
		options.HeadlessNoIncognito = headlessNoIncognito
	}
}

// WithXhrExtraction enables extraction of xhr requests
func WithXhrExtraction(xhrExtraction bool) Option {
	return func(options *Options) {
		options.XhrExtraction = xhrExtraction
	}
}

// WithHealthCheck determines if a self-healthcheck should be performed
func WithHealthCheck(healthCheck bool) Option {
	return func(options *Options) {
		options.HealthCheck = healthCheck
	}
}

// WithPprofServer enables pprof server
func WithPprofServer(pprofServer bool) Option {
	return func(options *Options) {
		options.PprofServer = pprofServer
	}
}

// WithErrorLogFile specifies a file to write with the errors of all requests
func WithErrorLogFile(errorLogFile string) Option {
	return func(options *Options) {
		options.ErrorLogFile = errorLogFile
	}
}

// WithResolvers sets the custom resolvers
func WithResolvers(resolvers []string) Option {
	return func(options *Options) {
		options.Resolvers = goflags.StringSlice(resolvers)
	}
}

// WithOutputMatchRegex sets the regex to match output url
func WithOutputMatchRegex(outputMatchRegex []string) Option {
	return func(options *Options) {
		options.OutputMatchRegex = goflags.StringSlice(outputMatchRegex)
	}
}

// WithOutputFilterRegex sets the regex to filter output url
func WithOutputFilterRegex(outputFilterRegex []string) Option {
	return func(options *Options) {
		options.OutputFilterRegex = goflags.StringSlice(outputFilterRegex)
	}
}

// WithFilterRegex sets the slice regex to filter url
func WithFilterRegex(filterRegex []*regexp.Regexp) Option {
	return func(options *Options) {
		options.FilterRegex = filterRegex
	}
}

// WithMatchRegex sets the slice regex to match url
func WithMatchRegex(matchRegex []*regexp.Regexp) Option {
	return func(options *Options) {
		options.MatchRegex = matchRegex
	}
}

// WithDisableUpdateCheck disables automatic update check
func WithDisableUpdateCheck(disableUpdateCheck bool) Option {
	return func(options *Options) {
		options.DisableUpdateCheck = disableUpdateCheck
	}
}

// WithIgnoreQueryParams ignores crawling same path with different query-param values
func WithIgnoreQueryParams(ignoreQueryParams bool) Option {
	return func(options *Options) {
		options.IgnoreQueryParams = ignoreQueryParams
	}
}

// WithDebug enables debug mode
func WithDebug(debug bool) Option {
	return func(options *Options) {
		options.Debug = debug
	}
}

// WithTlsImpersonate enables experimental tls ClientHello randomization for standard crawler
func WithTLSImpersonate(tlsImpersonate bool) Option {
	return func(options *Options) {
		options.TLSImpersonate = tlsImpersonate
	}
}

// WithDisableRedirects disables the following of redirects
func WithDisableRedirects(disableRedirects bool) Option {
	return func(options *Options) {
		options.DisableRedirects = disableRedirects
	}
}
