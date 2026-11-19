package objects

import (
	"context"
	"sync"
)

var (
	// uploadWaitGroup tracks in-flight uploads for graceful shutdown
	uploadWaitGroup sync.WaitGroup
)

// WaitForUploads waits for in-flight uploads to complete or ctx to end
func WaitForUploads(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		defer close(done)

		uploadWaitGroup.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// AddUpload increments the upload wait group
func AddUpload() {
	uploadWaitGroup.Add(1)
}

// DoneUpload decrements the upload wait group
func DoneUpload() {
	uploadWaitGroup.Done()
}
