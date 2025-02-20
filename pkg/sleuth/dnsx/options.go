package dnsx

import (
	"math"

	miekgdns "github.com/miekg/dns"
	pddnsx "github.com/projectdiscovery/dnsx/libs/dnsx"
)

// Options are the options for the DNSX client
type Options struct {
	// BaseResolvers are the base resolvers to use
	BaseResolvers []string
	// MaxRetries is the max number of retries to use
	MaxRetries int
	// QuestionTypes are the question types to use
	QuestionTypes []uint16
	// Trace is the trace option
	Trace bool
	// TraceMaxRecursion is the max recursion for the trace option
	TraceMaxRecursion int
	// Hostsfile is the hostsfile option
	Hostsfile bool
	// OutputCDN is the output CDN option
	OutputCDN bool
	// QueryAll is the query all option
	QueryAll bool
	// Proxy is the proxy option
	Proxy string
}

var questionTypes = []uint16{miekgdns.TypeA, miekgdns.TypeAAAA, miekgdns.TypeMX, miekgdns.TypeTXT, miekgdns.TypeNS, miekgdns.TypeCNAME, miekgdns.TypeSOA, miekgdns.TypeSPF}

var selectors = []string{"default", "selector1", "selector2", "google", "amazonses", "microsoft"}

// Option is a functional option for the DNSX client
type Option func(*Options)

// WithBaseResolvers sets the base resolvers for the DNSX client
func WithBaseResolvers(resolvers []string) Option {
	return func(o *Options) {
		o.BaseResolvers = resolvers
	}
}

// WithMaxRetries sets the max retries for the DNSX client
func WithMaxRetries(retries int) Option {
	return func(o *Options) {
		o.MaxRetries = retries
	}
}

// WithQuestionTypes sets the question types for the DNSX client
func WithQuestionTypes(types []uint16) Option {
	return func(o *Options) {
		o.QuestionTypes = types
	}
}

// WithTrace sets the trace option for the DNSX client
func WithTrace(trace bool) Option {
	return func(o *Options) {
		o.Trace = trace
	}
}

// WithTraceMaxRecursion sets the max recursion for the trace option
func WithTraceMaxRecursion(max int) Option {
	return func(o *Options) {
		o.TraceMaxRecursion = max
	}
}

// WithHostsfile sets the hostsfile option for the DNSX client
func WithHostsfile(hostsfile bool) Option {
	return func(o *Options) {
		o.Hostsfile = hostsfile
	}
}

// WithOutputCDN sets the output CDN option for the DNSX client
func WithOutputCDN(outputCDN bool) Option {
	return func(o *Options) {
		o.OutputCDN = outputCDN
	}
}

// WithQueryAll sets the query all option for the DNSX client
func WithQueryAll(queryAll bool) Option {
	return func(o *Options) {
		o.QueryAll = queryAll
	}
}

// NewOptions creates a new Options struct with default values and allows overrides
func NewOptions(opts ...Option) *Options {
	options := &Options{
		BaseResolvers:     pddnsx.DefaultResolvers,
		MaxRetries:        5, // nolint:mnd
		QuestionTypes:     []uint16{miekgdns.TypeA},
		TraceMaxRecursion: math.MaxUint16,
		Hostsfile:         true,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}
