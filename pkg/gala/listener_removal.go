package gala

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/riverqueue/river/rivertype"
)

// RemoveListeners detaches listeners and purges jobs no remaining listener can handle
// Failed cleanup remains retryable with the same IDs
func (g *Gala) RemoveListeners(ctx context.Context, ids ...ListenerID) error {
	if g == nil {
		return ErrGalaRequired
	}

	topics := g.registry.detachListeners(ids)
	if len(topics) == 0 {
		return nil
	}

	if g.jobController != nil {
		for topic := range g.listenerRemovalTopics(topics) {
			fragment, err := listenerTopicMetadataFragment(topic)
			if err != nil {
				return errors.Join(ErrRiverListenerCleanupFailed, err)
			}

			_, err = g.purgeActiveJobsWithMetadata(ctx, fragment, func(job *rivertype.JobRow) (bool, error) {
				envelope, err := decodeRiverJobEnvelope(job)
				if err != nil {
					return false, err
				}

				// Verify the encoded topic before destructive cleanup
				if envelope.Topic != topic {
					return false, nil
				}

				needed, err := g.envelopeHasMatchingListener(envelope)
				if err != nil {
					return false, err
				}

				return !needed, nil
			})
			if err != nil {
				return errors.Join(ErrRiverListenerCleanupFailed, err)
			}
		}
	}

	g.registry.completeListenerRemoval(ids)

	return nil
}

// listenerRemovalTopics expands affected topics to include retired aliases
func (g *Gala) listenerRemovalTopics(topics []TopicName) map[TopicName]struct{} {
	affected := make(map[TopicName]struct{}, len(topics))
	for _, topic := range topics {
		affected[topic] = struct{}{}
	}

	for retired, replacement := range g.topicRenames {
		if _, ok := affected[replacement]; ok {
			affected[retired] = struct{}{}
		}
	}

	return affected
}

func listenerTopicMetadataFragment(topic TopicName) (string, error) {
	fragment, err := json.Marshal(map[string]string{"topic": string(topic)})

	return string(fragment), err
}

func decodeRiverJobEnvelope(job *rivertype.JobRow) (Envelope, error) {
	if job == nil {
		return Envelope{}, ErrRiverDispatchJobEnvelopeRequired
	}

	var args EnvelopeArgs
	if err := json.Unmarshal(job.EncodedArgs, &args); err != nil {
		return Envelope{}, ErrRiverEnvelopeDecodeFailed
	}

	return decodeDispatchEnvelope(args.EnvelopePayload())
}

func (g *Gala) envelopeHasMatchingListener(envelope Envelope) (bool, error) {
	registration, err := g.registry.topicRegistration(envelope.Topic)
	if err != nil {
		renamed, ok := g.topicRenames[envelope.Topic]
		if !ok {
			return false, nil
		}

		envelope.Topic = renamed
		registration, err = g.registry.topicRegistration(renamed)
		if err != nil {
			return false, nil
		}
	}

	listeners := g.registry.registeredListeners(envelope.Topic)
	if len(listeners) == 0 {
		return false, nil
	}

	payload, err := registration.decode(envelope.Payload)
	if err != nil {
		return false, err
	}

	operation := payloadOperation(payload)
	if renamed, ok := g.operationRenames[operation]; ok {
		operation = renamed
	}

	for _, listener := range listeners {
		if listenerMatches(listener, operation) {
			return true, nil
		}
	}

	return false, nil
}
