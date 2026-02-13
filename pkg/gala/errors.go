package gala

import "errors"

var (
	// ErrRuntimeRequired is returned when a nil runtime is used.
	ErrRuntimeRequired = errors.New("gala: runtime is required")
	// ErrTopicNameRequired is returned when a topic name is empty.
	ErrTopicNameRequired = errors.New("gala: topic name is required")
	// ErrTopicAlreadyRegistered is returned when a topic is registered more than once.
	ErrTopicAlreadyRegistered = errors.New("gala: topic already registered")
	// ErrTopicNotRegistered is returned when topic metadata cannot be found.
	ErrTopicNotRegistered = errors.New("gala: topic not registered")
	// ErrCodecRequired is returned when a registration is missing a codec.
	ErrCodecRequired = errors.New("gala: codec is required")
	// ErrListenerNameRequired is returned when a listener name is empty.
	ErrListenerNameRequired = errors.New("gala: listener name is required")
	// ErrListenerHandlerRequired is returned when a listener callback is missing.
	ErrListenerHandlerRequired = errors.New("gala: listener handler is required")
	// ErrListenerTopicNotRegistered is returned when a listener is attached before topic registration.
	ErrListenerTopicNotRegistered = errors.New("gala: listener topic not registered")
	// ErrPayloadTypeMismatch is returned when payload casting fails for a topic or listener.
	ErrPayloadTypeMismatch = errors.New("gala: payload type mismatch")
	// ErrPayloadEncodeFailed is returned when payload serialization fails.
	ErrPayloadEncodeFailed = errors.New("gala: payload encode failed")
	// ErrPayloadDecodeFailed is returned when payload deserialization fails.
	ErrPayloadDecodeFailed = errors.New("gala: payload decode failed")
	// ErrEnvelopeTopicRequired is returned when an envelope has no topic.
	ErrEnvelopeTopicRequired = errors.New("gala: envelope topic is required")
	// ErrEnvelopePayloadRequired is returned when an envelope has an empty payload.
	ErrEnvelopePayloadRequired = errors.New("gala: envelope payload is required")
	// ErrListenerExecutionFailed is returned when listener processing fails.
	ErrListenerExecutionFailed = errors.New("gala: listener execution failed")
	// ErrUnsupportedEmitMode is returned when a topic policy specifies an unknown emit mode.
	ErrUnsupportedEmitMode = errors.New("gala: unsupported emit mode")
	// ErrDurableDispatcherRequired is returned when durable emit mode is used without a durable dispatcher.
	ErrDurableDispatcherRequired = errors.New("gala: durable dispatcher is required")
	// ErrDurableDispatchFailed is returned when durable dispatch fails.
	ErrDurableDispatchFailed = errors.New("gala: durable dispatch failed")
	// ErrContextCodecRequired is returned when context codec registration receives nil.
	ErrContextCodecRequired = errors.New("gala: context codec is required")
	// ErrContextCodecKeyRequired is returned when a context codec key is empty.
	ErrContextCodecKeyRequired = errors.New("gala: context codec key is required")
	// ErrContextCodecAlreadyRegistered is returned when a context codec key is duplicated.
	ErrContextCodecAlreadyRegistered = errors.New("gala: context codec already registered")
	// ErrContextSnapshotCaptureFailed is returned when snapshot capture fails.
	ErrContextSnapshotCaptureFailed = errors.New("gala: context snapshot capture failed")
	// ErrContextSnapshotRestoreFailed is returned when snapshot restore fails.
	ErrContextSnapshotRestoreFailed = errors.New("gala: context snapshot restore failed")
	// ErrRiverJobClientRequired is returned when a river dispatcher is built without a job client.
	ErrRiverJobClientRequired = errors.New("gala: river job client is required")
	// ErrRiverRuntimeProviderRequired is returned when a river worker is built without a runtime provider.
	ErrRiverRuntimeProviderRequired = errors.New("gala: river runtime provider is required")
	// ErrRiverDispatchJobEnvelopeRequired is returned when a river dispatch job has no envelope payload.
	ErrRiverDispatchJobEnvelopeRequired = errors.New("gala: river dispatch job envelope is required")
	// ErrRiverEnvelopeEncodeFailed is returned when encoding a river envelope payload fails.
	ErrRiverEnvelopeEncodeFailed = errors.New("gala: river envelope encode failed")
	// ErrRiverEnvelopeDecodeFailed is returned when decoding a river envelope payload fails.
	ErrRiverEnvelopeDecodeFailed = errors.New("gala: river envelope decode failed")
)
