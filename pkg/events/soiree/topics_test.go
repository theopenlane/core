package soiree

import (
	"errors"
	"testing"
)

// mockListener simulates a listener function for testing.
func mockListener(id string, shouldError bool) Listener {
	return func(_ *EventContext) error {
		if shouldError {
			return errors.New("listener error " + id) // nolint: err113
		}

		return nil
	}
}

func TestNewTopic(t *testing.T) {
	topic := newTopic()
	if topic == nil {
		t.Error("newTopic() should not return nil")
	}
}

func TestAddRemoveListener(t *testing.T) {
	topic := newTopic()
	listener1 := mockListener("1", false)
	listener2 := mockListener("2", false)

	id1 := "1"
	topic.addListener(id1, listener1)

	if len(topic.listeners) != 1 {
		t.Error("addListener() failed to add listener 1")
	}

	id2 := "2"
	topic.addListener(id2, listener2)

	if len(topic.listeners) != 2 {
		t.Error("addListener() failed to add listener 2")
	}

	if err := topic.removeListener(id1); err != nil {
		t.Errorf("removeListener() failed to remove listener 1: %s", err.Error())
	}

	if len(topic.listeners) != 1 {
		t.Errorf("removeListener() failed to remove listener 1, remaining listeners: %d", len(topic.listeners))
	}

	if err := topic.removeListener(id2); err != nil {
		t.Errorf("removeListener() failed to remove listener 2: %s", err.Error())
	}

	if len(topic.listeners) != 0 {
		t.Errorf("removeListener() failed to remove listener 2, remaining listeners: %d", len(topic.listeners))
	}
}

