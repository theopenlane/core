package sub

import (
	"io"
	"regexp"

	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/subfinder/v2/pkg/resolve"
	"github.com/projectdiscovery/subfinder/v2/pkg/runner"
)

// Options is a struct that contains the options for the subfinder tool
type Options struct {
	Verbose            bool                // Verbose flag indicates whether to show verbose output or not
	NoColor            bool                // NoColor disables the colored output
	JSON               bool                // JSON specifies whether to use json for output format or text file
	HostIP             bool                // HostIP specifies whether to write subdomains in host:ip format
	Silent             bool                // Silent suppresses any extra text and only writes subdomains to screen
	ListSources        bool                // ListSources specifies whether to list all available sources
	RemoveWildcard     bool                // RemoveWildcard specifies whether to remove potential wildcard or dead subdomains from the results.
	CaptureSources     bool                // CaptureSources specifies whether to save all sources that returned a specific domains or just the first source
	Stdin              bool                // Stdin specifies whether stdin input was given to the process
	Version            bool                // Version specifies if we should just show version and exit
	OnlyRecursive      bool                // Recursive specifies whether to use only recursive subdomain enumeration sources
	All                bool                // All specifies whether to use all (slow) sources.
	Statistics         bool                // Statistics specifies whether to report source statistics
	Threads            int                 // Threads controls the number of threads to use for active enumerations
	Timeout            int                 // Timeout is the seconds to wait for sources to respond
	MaxEnumerationTime int                 // MaxEnumerationTime is the maximum amount of time in minutes to wait for enumeration
	Domain             goflags.StringSlice // Domain is the domain to find subdomains for
	DomainsFile        string              // DomainsFile is the file containing list of domains to find subdomains for
	Output             io.Writer
	OutputFile         string               // Output is the file to write found subdomains to.
	OutputDirectory    string               // OutputDirectory is the directory to write results to in case list of domains is given
	Sources            goflags.StringSlice  `yaml:"sources,omitempty"`         // Sources contains a comma-separated list of sources to use for enumeration
	ExcludeSources     goflags.StringSlice  `yaml:"exclude-sources,omitempty"` // ExcludeSources contains the comma-separated sources to not include in the enumeration process
	Resolvers          goflags.StringSlice  `yaml:"resolvers,omitempty"`       // Resolvers is the comma-separated resolvers to use for enumeration
	ResolverList       string               // ResolverList is a text file containing list of resolvers to use for enumeration
	Config             string               // Config contains the location of the config file
	ProviderConfig     string               // ProviderConfig contains the location of the provider config file
	Proxy              string               // HTTP proxy
	RateLimit          int                  // Global maximum number of HTTP requests to send per second
	RateLimits         goflags.RateLimitMap // Maximum number of HTTP requests to send per second
	ExcludeIps         bool
	Match              goflags.StringSlice
	Filter             goflags.StringSlice
	matchRegexes       []*regexp.Regexp
	filterRegexes      []*regexp.Regexp
	ResultCallback     runner.OnResultCallback // OnResult callback
	DisableUpdateCheck bool                    // DisableUpdateCheck disable update checking
}

type ConfigOpts func(*Options)
type SubfinderOpts func(*Subfinder)

// WithThreads sets the number of threads for the subfinder tool
func WithThreads(threads int) ConfigOpts {
	return func(o *Options) {
		o.Threads = threads
	}
}

// WithTimeout sets the timeout for the subfinder tool
func WithTimeout(timeout int) ConfigOpts {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// WithMaxEnumerationTime sets the maximum enumeration time for the subfinder tool
func WithMaxEnumerationTime(maxEnumerationTime int) ConfigOpts {
	return func(o *Options) {
		o.MaxEnumerationTime = maxEnumerationTime
	}
}

// WithProxy sets the proxy for the subfinder tool
func WithProxy(proxy string) ConfigOpts {
	return func(o *Options) {
		o.Proxy = proxy
	}
}

// WithHostIP sets the host IP for the subfinder tool
func WithHostIP(hostIP bool) ConfigOpts {
	return func(o *Options) {
		o.HostIP = hostIP
	}
}

// WithSilent sets the silent mode for the subfinder tool
func WithSilent(silent bool) ConfigOpts {
	return func(o *Options) {
		o.Silent = silent
	}
}

// WithListSources sets the list sources flag for the subfinder tool
func WithListSources(listSources bool) ConfigOpts {
	return func(o *Options) {
		o.ListSources = listSources
	}
}

// WithRemoveWildcard sets the remove wildcard flag for the subfinder tool
func WithRemoveWildcard(removeWildcard bool) ConfigOpts {
	return func(o *Options) {
		o.RemoveWildcard = removeWildcard
	}
}

// WithCaptureSources sets the capture sources flag for the subfinder tool
func WithCaptureSources(captureSources bool) ConfigOpts {
	return func(o *Options) {
		o.CaptureSources = captureSources
	}
}

// WithStdin sets the stdin flag for the subfinder tool
func WithStdin(stdin bool) ConfigOpts {
	return func(o *Options) {
		o.Stdin = stdin
	}
}

// WithVersion sets the version flag for the subfinder tool
func WithVersion(version bool) ConfigOpts {
	return func(o *Options) {
		o.Version = version
	}
}

// WithOnlyRecursive sets the only recursive flag for the subfinder tool
func WithOnlyRecursive(onlyRecursive bool) ConfigOpts {
	return func(o *Options) {
		o.OnlyRecursive = onlyRecursive
	}
}

// WithAll sets the all flag for the subfinder tool
func WithAll(all bool) ConfigOpts {
	return func(o *Options) {
		o.All = all
	}
}

// WithStatistics sets the statistics flag for the subfinder tool
func WithStatistics(statistics bool) ConfigOpts {
	return func(o *Options) {
		o.Statistics = statistics
	}
}

// WithJSON sets the JSON flag for the subfinder tool
func WithJSON(json bool) ConfigOpts {
	return func(o *Options) {
		o.JSON = json
	}
}

// WithNoColor sets the no color flag for the subfinder tool
func WithNoColor(noColor bool) ConfigOpts {
	return func(o *Options) {
		o.NoColor = noColor
	}
}

// WithConfigFilePath sets the configuration file path for the subfinder tool
func WithConfigFilePath(configFilePath string) ConfigOpts {
	return func(o *Options) {
		o.Config = configFilePath
	}
}

// WithConfigFile sets the configuration file for the subfinder tool
func WithConfigFile(configFile string) ConfigOpts {
	return func(o *Options) {
		o.Config = configFile
	}
}

// WithProviderConfigFile sets the provider configuration file for the subfinder tool
func WithProviderConfigFile(providerConfigFile string) ConfigOpts {
	return func(o *Options) {
		o.ProviderConfig = providerConfigFile
	}
}

// WithOutput sets the output writer for the subfinder tool
func WithOutput(output io.Writer) ConfigOpts {
	return func(o *Options) {
		o.Output = output
	}
}

// WithOutputFile sets the output file for the subfinder tool
func WithOutputFile(outputFile string) ConfigOpts {
	return func(o *Options) {
		o.OutputFile = outputFile
	}
}

// WithOutputDirectory sets the output directory for the subfinder tool
func WithOutputDirectory(outputDirectory string) ConfigOpts {
	return func(o *Options) {
		o.OutputDirectory = outputDirectory
	}
}

// WithDomain sets the domain for the subfinder tool
func WithDomain(domain string) ConfigOpts {
	return func(o *Options) {
		o.Domain = goflags.StringSlice{domain}
	}
}

// WithDomainsFile sets the domains file for the subfinder tool
func WithDomainsFile(domainsFile string) ConfigOpts {
	return func(o *Options) {
		o.DomainsFile = domainsFile
	}
}

// WithSources sets the sources for the subfinder tool
func WithSources(sources goflags.StringSlice) ConfigOpts {
	return func(o *Options) {
		o.Sources = sources
	}
}

// WithExcludeSources sets the exclude sources for the subfinder tool
func WithExcludeSources(excludeSources goflags.StringSlice) ConfigOpts {
	return func(o *Options) {
		o.ExcludeSources = excludeSources
	}
}

// WithResolvers sets the resolvers for the subfinder tool
func WithResolvers(resolvers goflags.StringSlice) ConfigOpts {
	return func(o *Options) {
		o.Resolvers = resolvers
	}
}

// WithResolverList sets the resolver list for the subfinder tool
func WithResolverList(resolverList string) ConfigOpts {
	return func(o *Options) {
		o.ResolverList = resolverList
	}
}

// WithRateLimit sets the rate limit for the subfinder tool
func WithRateLimit(rateLimit int) ConfigOpts {
	return func(o *Options) {
		o.RateLimit = rateLimit
	}
}

// WithRateLimits sets the rate limits for the subfinder tool
func WithRateLimits(rateLimits goflags.RateLimitMap) ConfigOpts {
	return func(o *Options) {
		o.RateLimits = rateLimits
	}
}

// WithExcludeIps sets the exclude IPs for the subfinder tool
func WithExcludeIps(excludeIps bool) ConfigOpts {
	return func(o *Options) {
		o.ExcludeIps = excludeIps
	}
}

// WithMatch sets the match regexes for the subfinder tool
func WithMatch(match goflags.StringSlice) ConfigOpts {
	return func(o *Options) {
		o.Match = match
	}
}

// WithFilter sets the filter regexes for the subfinder tool
func WithFilter(filter goflags.StringSlice) ConfigOpts {
	return func(o *Options) {
		o.Filter = filter
	}
}

// WithMatchRegexes sets the match regexes for the subfinder tool
func WithMatchRegexes(regexes []*regexp.Regexp) ConfigOpts {
	return func(o *Options) {
		o.matchRegexes = regexes
	}
}

// WithFilterRegexes sets the filter regexes for the subfinder tool
func WithFilterRegexes(regexes []*regexp.Regexp) ConfigOpts {
	return func(o *Options) {
		o.filterRegexes = regexes
	}
}

// WithResultCallback sets the result callback for the subfinder tool
func WithResultCallback(callback runner.OnResultCallback) ConfigOpts {
	return func(o *Options) {
		o.ResultCallback = callback
	}
}

// WithDisableUpdateCheck sets the disable update check for the subfinder tool
func WithDisableUpdateCheck(disableUpdateCheck bool) ConfigOpts {
	return func(o *Options) {
		o.DisableUpdateCheck = disableUpdateCheck
	}
}

// WithVerbose sets the verbose flag for the subfinder tool
func WithVerbose(verbose bool) ConfigOpts {
	return func(o *Options) {
		o.Verbose = verbose
	}
}

// WithConfig sets the configuration for the subfinder tool
func WithConfig(config *Options) SubfinderOpts {
	return func(s *Subfinder) {
		s.Options = config
	}
}

// WithRunner sets the runner for the subfinder tool
func WithRunner(r *runner.Runner) SubfinderOpts {
	return func(s *Subfinder) {
		s.Runner = r
	}
}

// WithResolver sets the resolver for the subfinder tool
func WithResolver(r *resolve.Resolver) SubfinderOpts {
	return func(s *Subfinder) {
		s.Resolver = r
	}
}
