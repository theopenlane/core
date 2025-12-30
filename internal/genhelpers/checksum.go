package genhelpers

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

const (
	defaultFilePermissions = 0644
)

// HasSchemaChanges checks if there have been changes to schema or generated files
// by comparing checksums of the relevant directories
func HasSchemaChanges(checksumFile string, directories ...string) (bool, error) {
	// Calculate current checksum of schema and generated directories
	currentChecksum, err := calculateDirChecksum(directories...)
	if err != nil {
		return false, err
	}

	// Read previous checksum if it exists
	previousChecksum, err := os.ReadFile(checksumFile)
	if err != nil {
		if os.IsNotExist(err) {
			// First run, always generate
			_ = os.WriteFile(checksumFile, []byte(currentChecksum), defaultFilePermissions)
			return true, nil
		}
		return false, err
	}

	// Compare checksums
	if string(previousChecksum) != currentChecksum {
		// Update checksum file
		_ = os.WriteFile(checksumFile, []byte(currentChecksum), defaultFilePermissions)
		return true, nil
	}

	return false, nil
}

// SetSchemaChecksum updates the checksum file with the current checksum
func SetSchemaChecksum(checksumFile string, paths ...string) error {
	// Calculate current checksum of schema and generated directories
	currentChecksum, err := calculateDirChecksum(paths...)
	if err != nil {
		return err
	}

	// Update checksum file
	_ = os.WriteFile(checksumFile, []byte(currentChecksum), defaultFilePermissions)
	return nil
}

// calculateDirChecksum calculates a checksum for multiple directories
func calculateDirChecksum(paths ...string) (string, error) {
	h := sha256.New()

	for _, path := range paths {
		if err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			// Only process .go files and .graphql files
			if filepath.Ext(filePath) != ".go" && filepath.Ext(filePath) != ".graphql" {
				return nil
			}

			f, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(h, f); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
