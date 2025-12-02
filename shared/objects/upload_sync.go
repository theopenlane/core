package objects

import "sync"

var (
	// uploadWaitGroup tracks in-flight uploads for graceful shutdown
	uploadWaitGroup sync.WaitGroup
)

// WaitForUploads waits for all in-flight uploads to complete
func WaitForUploads() {
	uploadWaitGroup.Wait()
}

// AddUpload increments the upload wait group
func AddUpload() {
	uploadWaitGroup.Add(1)
}

// DoneUpload decrements the upload wait group
func DoneUpload() {
	uploadWaitGroup.Done()
}
