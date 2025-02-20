package sub

import (
	"errors"

	"github.com/projectdiscovery/subfinder/v2/pkg/resolve"
	"github.com/projectdiscovery/subfinder/v2/pkg/runner"
)

// Subfinder is a struct that contains the configuration for the subfinder tool
type Subfinder struct {
	Runner *runner.Runner
	// Resolver is the resolver for the subfinder tool
	Resolver *resolve.Resolver
	// Config is the configuration for the subfinder tool
	Options *Options
}

// NewOptions creates a new Options struct with default values and allows overrides
func NewOptions(opt ...ConfigOpts) *Options {
	options := &Options{
		Verbose: false,
		// intentionally hard coding this for now because it's just empty but the package requires it be loaded to function
		ProviderConfig:     "pkg/sleuth/sub/config/provider-config.yaml",
		Threads:            10, // nolint:mnd
		MaxEnumerationTime: 10, // nolint:mnd
	}

	for _, opt := range opt {
		opt(options)
	}

	return options
}

var (
	ErrOptionsNil = errors.New("options cannot be nil")
)

// NewSubfinder creates a new Subfinder instance with the given options
func NewSubfinder(opt *Options) (*Subfinder, error) {
	if opt == nil {
		return nil, ErrOptionsNil
	}

	runnerOptions := MapToTypesOptions(opt)

	runner, err := runner.NewRunner(runnerOptions)
	if err != nil {
		return nil, err
	}

	// TODO(MKA): I didn't have enough time to test / confirm functionally what a regular resolver is vs the pool
	resolver := resolve.New()
	resolverPool := resolver.NewResolutionPool(opt.Threads, opt.RemoveWildcard)

	return &Subfinder{
		Runner:   runner,
		Resolver: resolverPool.Resolver,
		Options:  opt,
	}, nil
}

// MapToTypesOptions maps Options to types.Options
func MapToTypesOptions(options *Options) *runner.Options {
	return &runner.Options{
		Verbose:            options.Verbose,
		NoColor:            options.NoColor,
		JSON:               options.JSON,
		HostIP:             options.HostIP,
		Silent:             options.Silent,
		ListSources:        options.ListSources,
		RemoveWildcard:     options.RemoveWildcard,
		CaptureSources:     options.CaptureSources,
		Stdin:              options.Stdin,
		Version:            options.Version,
		OnlyRecursive:      options.OnlyRecursive,
		All:                options.All,
		Statistics:         options.Statistics,
		Threads:            options.Threads,
		Timeout:            options.Timeout,
		MaxEnumerationTime: options.MaxEnumerationTime,
		Domain:             options.Domain,
		DomainsFile:        options.DomainsFile,
		Output:             options.Output,
		OutputFile:         options.OutputFile,
		OutputDirectory:    options.OutputDirectory,
		Sources:            options.Sources,
		ExcludeSources:     options.ExcludeSources,
		Resolvers:          options.Resolvers,
		ResolverList:       options.ResolverList,
		Config:             options.Config,
		ProviderConfig:     options.ProviderConfig,
		Proxy:              options.Proxy,
		RateLimit:          options.RateLimit,
		RateLimits:         options.RateLimits,
		ExcludeIps:         options.ExcludeIps,
		Match:              options.Match,
		Filter:             options.Filter,
		DisableUpdateCheck: options.DisableUpdateCheck,
	}
}
