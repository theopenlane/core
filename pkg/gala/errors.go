package gala

import (
	"errors"
	"fmt"
)

var (
	// ErrGalaRequired is returned when a nil gala runtime is used
	ErrGalaRequired = errors.New("gala: gala is required")
	// ErrRegistryRequired is returned when a nil topic registry is used
	ErrRegistryRequired = errors.New("gala: registry is required")
	// ErrTopicNameRequired is returned when a topic name is empty
	ErrTopicNameRequired = errors.New("gala: topic name is required")
	// ErrTopicAlreadyRegistered is returned when a topic is registered more than once
	ErrTopicAlreadyRegistered = errors.New("gala: topic already registered")
	// ErrTopicNotRegistered is returned when topic metadata cannot be found
	ErrTopicNotRegistered = errors.New("gala: topic not registered")
	// ErrCodecRequired is returned when a registration is missing a codec
	ErrCodecRequired = errors.New("gala: codec is required")
	// ErrListenerNameRequired is returned when a listener name is empty
	ErrListenerNameRequired = errors.New("gala: listener name is required")
	// ErrListenerHandlerRequired is returned when a listener callback is missing
	ErrListenerHandlerRequired = errors.New("gala: listener handler is required")
	// ErrListenerTopicNotRegistered is returned when a listener is attached before topic registration
	ErrListenerTopicNotRegistered = errors.New("gala: listener topic not registered")
	// ErrPayloadTypeMismatch is returned when payload casting fails for a topic or listener
	ErrPayloadTypeMismatch = errors.New("gala: payload type mismatch")
	// ErrPayloadEncodeFailed is returned when payload serialization fails
	ErrPayloadEncodeFailed = errors.New("gala: payload encode failed")
	// ErrPayloadDecodeFailed is returned when payload deserialization fails
	ErrPayloadDecodeFailed = errors.New("gala: payload decode failed")
	// ErrEnvelopePayloadRequired is returned when an envelope has an empty payload
	ErrEnvelopePayloadRequired = errors.New("gala: envelope payload is required")
	// ErrDispatcherRequired is returned when emit is attempted without a dispatcher
	ErrDispatcherRequired = errors.New("gala: dispatcher is required")
	// ErrDispatchFailed is returned when dispatch fails
	ErrDispatchFailed = errors.New("gala: dispatch failed")
	// ErrContextCodecRequired is returned when context codec registration receives nil
	ErrContextCodecRequired = errors.New("gala: context codec is required")
	// ErrContextCodecKeyRequired is returned when a context codec key is empty
	ErrContextCodecKeyRequired = errors.New("gala: context codec key is required")
	// ErrContextCodecAlreadyRegistered is returned when a context codec key is duplicated
	ErrContextCodecAlreadyRegistered = errors.New("gala: context codec already registered")
	// ErrContextSnapshotCaptureFailed is returned when snapshot capture fails
	ErrContextSnapshotCaptureFailed = errors.New("gala: context snapshot capture failed")
	// ErrContextSnapshotRestoreFailed is returned when snapshot restore fails
	ErrContextSnapshotRestoreFailed = errors.New("gala: context snapshot restore failed")
	// ErrRiverJobClientRequired is returned when a river dispatcher is built without a job client
	ErrRiverJobClientRequired = errors.New("gala: river job client is required")
	// ErrRiverGalaProviderRequired is returned when a river worker is built without a gala provider
	ErrRiverGalaProviderRequired = errors.New("gala: river gala provider is required")
	// ErrRiverDispatchJobEnvelopeRequired is returned when a river dispatch job has no envelope payload
	ErrRiverDispatchJobEnvelopeRequired = errors.New("gala: river dispatch job envelope is required")
	// ErrRiverEnvelopeEncodeFailed is returned when encoding a river envelope payload fails
	ErrRiverEnvelopeEncodeFailed = errors.New("gala: river envelope encode failed")
	// ErrRiverEnvelopeDecodeFailed is returned when decoding a river envelope payload fails
	ErrRiverEnvelopeDecodeFailed = errors.New("gala: river envelope decode failed")
	// ErrRiverDispatchInsertFailed is returned when inserting a durable river dispatch job fails
	ErrRiverDispatchInsertFailed = errors.New("gala: river dispatch insert failed")
	// ErrRiverConnectionURIRequired is returned when river runtime setup is missing a connection URI
	ErrRiverConnectionURIRequired = errors.New("gala: river connection URI is required")
	// ErrRiverClientInitializationFailed is returned when building the river queue client fails
	ErrRiverClientInitializationFailed = errors.New("gala: river client initialization failed")
	// ErrRiverWorkerStartFailed is returned when starting gala river workers fails
	ErrRiverWorkerStartFailed = errors.New("gala: river worker start failed")
	// ErrRiverWorkerStopFailed is returned when stopping gala river workers fails
	ErrRiverWorkerStopFailed = errors.New("gala: river worker stop failed")
	// ErrRiverClientCloseFailed is returned when closing the gala river queue client fails
	ErrRiverClientCloseFailed = errors.New("gala: river client close failed")
	// ErrDispatchModeInvalid is returned when an unknown gala dispatch mode is configured.
	ErrDispatchModeInvalid = errors.New("gala: dispatch mode is invalid")
	// ErrListenerPanicked is returned when a listener panics during execution
	ErrListenerPanicked = errors.New("gala: listener panicked")
)

// ListenerError captures a listener execution failure with context
type ListenerError struct {
	// ListenerName is the name of the listener that failed
	ListenerName string
	// Cause is the underlying error from the listener
	Cause error
	// Panicked indicates whether the listener panicked
	Panicked bool
}

// Error returns an error message for listener execution failures
func (e ListenerError) Error() string {
	if e.Panicked {
		if e.Cause != nil {
			return fmt.Sprintf("gala: listener %q panicked: %v", e.ListenerName, e.Cause)
		}

		return "gala: listener panicked"
	}

	if e.Cause != nil {
		return fmt.Sprintf("gala: listener %q execution failed: %v", e.ListenerName, e.Cause)
	}

	return "gala: listener execution failed"
}

// Unwrap returns the underlying cause for use with errors.Is and errors.As
func (e ListenerError) Unwrap() error {
	return e.Cause
}
