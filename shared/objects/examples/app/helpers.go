//go:build examples

package app

import (
	"os"
	"path/filepath"
	"runtime"
)

var (
	_, callerFile, _, ok = runtime.Caller(0)
	examplesRoot         = func() string {
		if ok {
			return filepath.Dir(filepath.Dir(callerFile))
		}
		return "."
	}()
)

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func resolvePath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(examplesRoot, p)
}

func composeFilePath() string {
	return filepath.Join(examplesRoot, "docker-compose.yml")
}
