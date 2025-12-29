package objects

import "sync"

var (
	// uploadWaitGroup tracks in-flight uploads for graceful shutdown
	uploadWaitGroup sync.WaitGroup
	mu              sync.Mutex
)

// WaitForUploads waits for all in-flight uploads to complete
func WaitForUploads() {
	uploadWaitGroup.Wait()
}

// AddUpload increments the upload wait group
func AddUpload() {
	mu.Lock()
	defer mu.Unlock()
	uploadWaitGroup.Add(1)
}

// DoneUpload decrements the upload wait group
func DoneUpload() {
	uploadWaitGroup.Done()
}
