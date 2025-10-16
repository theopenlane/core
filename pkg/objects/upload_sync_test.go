package objects

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddUpload(t *testing.T) {
	uploadWaitGroup = sync.WaitGroup{}

	AddUpload()

	done := make(chan bool)
	go func() {
		uploadWaitGroup.Wait()
		done <- true
	}()

	select {
	case <-done:
		t.Fatal("WaitGroup should not be done yet")
	case <-time.After(50 * time.Millisecond):
	}

	uploadWaitGroup.Done()
}

func TestDoneUpload(t *testing.T) {
	uploadWaitGroup = sync.WaitGroup{}

	uploadWaitGroup.Add(1)
	DoneUpload()

	done := make(chan bool)
	go func() {
		uploadWaitGroup.Wait()
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("WaitGroup should be done")
	}
}

func TestWaitForUploads(t *testing.T) {
	uploadWaitGroup = sync.WaitGroup{}

	uploadWaitGroup.Add(1)

	done := make(chan bool)
	go func() {
		WaitForUploads()
		done <- true
	}()

	select {
	case <-done:
		t.Fatal("WaitForUploads should still be waiting")
	case <-time.After(50 * time.Millisecond):
	}

	uploadWaitGroup.Done()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("WaitForUploads should have completed")
	}
}

func TestUploadSync_Concurrent(t *testing.T) {
	uploadWaitGroup = sync.WaitGroup{}

	const numGoroutines = 10

	for i := 0; i < numGoroutines; i++ {
		AddUpload()
	}

	done := make(chan bool)
	go func() {
		WaitForUploads()
		done <- true
	}()

	select {
	case <-done:
		t.Fatal("WaitForUploads should still be waiting")
	case <-time.After(50 * time.Millisecond):
	}

	for i := 0; i < numGoroutines; i++ {
		DoneUpload()
	}

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("WaitForUploads should have completed")
	}
}

func TestUploadSync_MultipleWaiters(t *testing.T) {
	uploadWaitGroup = sync.WaitGroup{}

	AddUpload()

	waiter1Done := make(chan bool)
	waiter2Done := make(chan bool)

	go func() {
		WaitForUploads()
		waiter1Done <- true
	}()

	go func() {
		WaitForUploads()
		waiter2Done <- true
	}()

	select {
	case <-waiter1Done:
		t.Fatal("waiter1 should still be waiting")
	case <-waiter2Done:
		t.Fatal("waiter2 should still be waiting")
	case <-time.After(50 * time.Millisecond):
	}

	DoneUpload()

	timeout := time.After(100 * time.Millisecond)
	waiter1Received := false
	waiter2Received := false

	for !waiter1Received || !waiter2Received {
		select {
		case <-waiter1Done:
			waiter1Received = true
		case <-waiter2Done:
			waiter2Received = true
		case <-timeout:
			t.Fatal("both waiters should have completed")
		}
	}

	assert.True(t, waiter1Received)
	assert.True(t, waiter2Received)
}

func TestUploadSync_NoUploads(t *testing.T) {
	uploadWaitGroup = sync.WaitGroup{}

	done := make(chan bool)
	go func() {
		WaitForUploads()
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("WaitForUploads should complete immediately when counter is zero")
	}
}

func TestUploadSync_SequentialOperations(t *testing.T) {
	uploadWaitGroup = sync.WaitGroup{}

	AddUpload()
	DoneUpload()

	done := make(chan bool)
	go func() {
		WaitForUploads()
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("WaitForUploads should complete immediately after all uploads are done")
	}

	AddUpload()
	AddUpload()

	go func() {
		WaitForUploads()
		done <- true
	}()

	select {
	case <-done:
		t.Fatal("WaitForUploads should be waiting for new uploads")
	case <-time.After(50 * time.Millisecond):
	}

	DoneUpload()
	DoneUpload()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("WaitForUploads should complete after all new uploads are done")
	}
}
