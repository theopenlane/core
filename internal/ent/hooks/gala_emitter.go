package hooks

import (
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/gala"
)

// GalaEmitter configures independent Gala mutation dispatch from ent hooks.
type GalaEmitter struct {
	runtimeProvider    func() *gala.Runtime
	dualEmitEnabled    bool
	failOnEnqueueError bool
	topics             map[string]struct{}
	topicModes         map[string]workflows.GalaTopicMode
}

// GalaEmitterOpts configures a GalaEmitter.
type GalaEmitterOpts func(*GalaEmitter)

// NewGalaEmitter constructs a GalaEmitter and applies the provided option set.
func NewGalaEmitter(opts ...GalaEmitterOpts) *GalaEmitter {
	g := &GalaEmitter{}

	lo.ForEach(opts, func(opt GalaEmitterOpts, _ int) { opt(g) })

	return g
}

// WithGalaRuntimeProvider injects a runtime resolver used at emit time.
func WithGalaRuntimeProvider(provider func() *gala.Runtime) GalaEmitterOpts {
	return func(g *GalaEmitter) {
		g.runtimeProvider = provider
	}
}

// WithGalaDualEmitEnabled toggles mutation emission into Gala.
func WithGalaDualEmitEnabled(enabled bool) GalaEmitterOpts {
	return func(g *GalaEmitter) {
		g.dualEmitEnabled = enabled
	}
}

// WithGalaFailOnEnqueueError controls strict-mode logging for Gala enqueue failures.
func WithGalaFailOnEnqueueError(enabled bool) GalaEmitterOpts {
	return func(g *GalaEmitter) {
		g.failOnEnqueueError = enabled
	}
}

// WithGalaTopics scopes Gala emit to specific mutation topics. Empty means all topics.
func WithGalaTopics(topics []string) GalaEmitterOpts {
	return func(g *GalaEmitter) {
		if len(topics) == 0 {
			g.topics = nil
			return
		}

		g.topics = lo.SliceToMap(topics, func(topic string) (string, struct{}) {
			return strings.TrimSpace(topic), struct{}{}
		})
		g.topics = lo.PickBy(g.topics, func(topic string, _ struct{}) bool {
			return topic != ""
		})
		if len(g.topics) == 0 {
			g.topics = nil
		}
	}
}

// WithGalaTopicModes overrides migration behavior by mutation topic.
func WithGalaTopicModes(topicModes map[string]workflows.GalaTopicMode) GalaEmitterOpts {
	return func(g *GalaEmitter) {
		validModes := lo.PickBy(topicModes, func(topic string, mode workflows.GalaTopicMode) bool {
			return strings.TrimSpace(topic) != "" && mode.IsValid()
		})
		if len(validModes) == 0 {
			g.topicModes = nil
			return
		}

		g.topicModes = lo.Assign(map[string]workflows.GalaTopicMode{}, validModes)
	}
}

// runtime resolves the configured gala runtime if one is available.
func (g *GalaEmitter) runtime() *gala.Runtime {
	if g == nil || g.runtimeProvider == nil {
		return nil
	}

	return g.runtimeProvider()
}

// shouldDispatch reports whether the topic should be emitted into Gala.
func (g *GalaEmitter) shouldDispatch(topic string) bool {
	if g == nil {
		return false
	}

	if mode, ok := g.topicMode(topic); ok {
		return mode == workflows.GalaTopicModeDualEmit || mode == workflows.GalaTopicModeV2Only
	}

	if !g.dualEmitEnabled {
		return false
	}

	if len(g.topics) == 0 {
		return true
	}

	if _, ok := g.topics["*"]; ok {
		return true
	}

	_, ok := g.topics[topic]

	return ok
}

// topicMode resolves explicit per-topic mode overrides using exact-match, then wildcard.
func (g *GalaEmitter) topicMode(topic string) (workflows.GalaTopicMode, bool) {
	if g == nil || len(g.topicModes) == 0 {
		return "", false
	}

	if mode, ok := g.topicModes[topic]; ok {
		return mode, true
	}

	if mode, ok := g.topicModes["*"]; ok {
		return mode, true
	}

	return "", false
}
