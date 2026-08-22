package gala

import (
	"sync"

	"github.com/samber/lo"
)

// envelopeKindRegistry is the union of registered envelope job kinds across runtimes;
// River reads worker kind aliases from a zero-value args instance, so the list lives at
// package level and is populated from Config.Kinds before worker registration
var envelopeKindRegistry = struct {
	mu    sync.RWMutex
	kinds []string
}{}

// registerEnvelopeKinds merges kind names into the alias registry
func registerEnvelopeKinds(kinds []string) {
	envelopeKindRegistry.mu.Lock()
	defer envelopeKindRegistry.mu.Unlock()

	envelopeKindRegistry.kinds = lo.Uniq(append(envelopeKindRegistry.kinds, kinds...))
}

// EnvelopeArgs is the durable dispatch payload
type EnvelopeArgs struct {
	// Envelope is the encoded gala envelope payload
	Envelope []byte `json:"envelope"`
	// UniqueKey scopes the ByArgs uniqueness hash to this field alone via the river tag
	UniqueKey string `json:"unique_key,omitempty" river:"unique"`
}

// Kind satisfies river.JobArgs with the legacy dispatch kind
func (EnvelopeArgs) Kind() string {
	return riverDispatchJobKind
}

// KindAliases registers the shared envelope worker for every configured job kind
func (EnvelopeArgs) KindAliases() []string {
	envelopeKindRegistry.mu.RLock()
	defer envelopeKindRegistry.mu.RUnlock()

	return append([]string(nil), envelopeKindRegistry.kinds...)
}

// EnvelopePayload returns the encoded envelope bytes
func (a EnvelopeArgs) EnvelopePayload() []byte {
	return a.Envelope
}

// kindedEnvelopeArgs stamps a registered kind onto an envelope insert
type kindedEnvelopeArgs struct {
	EnvelopeArgs

	kind string
}

// Kind returns the stamped kind, falling back to the legacy dispatch kind
func (a kindedEnvelopeArgs) Kind() string {
	if a.kind == "" {
		return riverDispatchJobKind
	}

	return a.kind
}
