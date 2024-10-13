package objects

import "github.com/rs/zerolog/log"

// Progress is used to track the progress of a file upload
// It implements the io.Writer interface so it can be passed to an io.TeeReader()
type Progress struct {
	// TotalSize is the total size of the file being uploaded
	TotalSize int64
	// BytesRead is the number of bytes that have been read so far
	BytesRead int64
}

// Write is used to satisfy the io.Writer interface Instead of writing somewhere, it simply aggregates the total bytes on each read
func (pr *Progress) Write(p []byte) (n int, err error) {
	n, err = len(p), nil

	pr.BytesRead += int64(n)

	pr.Print()

	return
}

// Print displays the current progress of the file upload
func (pr *Progress) Print() {
	if pr.BytesRead == pr.TotalSize {
		log.Debug().Msg("file upload complete")

		return
	}

	log.Debug().Int64("bytes_read", pr.BytesRead).Msg("file upload in progress")
}
